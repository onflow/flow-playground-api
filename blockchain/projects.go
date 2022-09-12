package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/getsentry/sentry-go"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"

	"github.com/golang/groupcache/lru"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// improvement: create instance pool as a possible optimization. We can pre-instantiate empty
// instances of emulators waiting around to be assigned to a project if init time will be proved to be an issue

// NewProjects creates an instance of the projects with provided storage access and caching.
func NewProjects(store storage.Store, cache *lru.Cache) *Projects {
	return &Projects{
		store: store,
		cache: cache,
		mu:    map[uuid.UUID]*sync.RWMutex{},
	}
}

// Projects implements stateful operations on the blockchain, keeping records of transaction executions.
//
// Projects expose API to interact with the blockchain all in context of a project but also makes sure
// the state is persisted and implements state recreation with caching and resource locking.
type Projects struct {
	store     storage.Store
	cache     *lru.Cache
	mu        map[uuid.UUID]*sync.RWMutex
	muCounter sync.Map
}

// load initializes an emulator and run transactions previously executed in the project to establish a state.
func (s *Projects) load(projectID uuid.UUID) (*emulator, error) {
	val, ok := s.cache.Get(projectID)
	if ok {
		return val.(*emulator), nil
	}

	emulator, err := newEmulator()
	if err != nil {
		return nil, err
	}

	var executions []*model.TransactionExecution
	err = s.store.GetTransactionExecutionsForProject(projectID, &executions)
	if err != nil {
		return nil, err
	}

	for _, execution := range executions {
		result, _, err := emulator.executeTransaction(execution.Script, execution.Arguments, execution.SignersToFlow())
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("not able to recreate the project state %s", projectID))
		}
		if result.Error != nil && len(execution.Errors) == 0 {
			sentry.CaptureMessage(fmt.Sprintf(
				"project %s state recreation failure: execution %s failed with result: %s, debug: %s",
				projectID.String(),
				execution.ID.String(),
				result.Error.Error(),
				result.Debug.Message,
			))
			return nil, errors.Wrap(result.Error, fmt.Sprintf("not able to recreate the project state %s", projectID))
		}
	}

	s.cache.Add(projectID, emulator)

	return emulator, nil
}

// loadLock retrieves the mutex lock by the project ID and increase the usage counter.
func (s *Projects) loadLock(uuid uuid.UUID) *sync.RWMutex {
	counter, ok := s.muCounter.LoadOrStore(uuid, 0)
	s.muCounter.Store(uuid, counter.(int)+1)

	_, ok = s.mu[uuid]
	if !ok {
		s.mu[uuid] = &sync.RWMutex{}
	}

	return s.mu[uuid]
}

// removeLock returns the mutex lock by the project ID and decreases usage counter, deleting the map entry if at 0.
func (s *Projects) removeLock(uuid uuid.UUID) *sync.RWMutex {
	m := s.mu[uuid]

	counter, _ := s.muCounter.Load(uuid)
	if counter == 0 {
		delete(s.mu, uuid)
		s.muCounter.Delete(uuid)
	} else {
		s.muCounter.Store(uuid, counter.(int)-1)
	}

	return m
}

// Reset the blockchain state.
func (s *Projects) Reset(project *model.InternalProject) error {
	s.cache.Remove(project.ID)

	err := s.store.ResetProjectState(project)
	if err != nil {
		return err
	}

	_, err = s.CreateInitialAccounts(project.ID, 5)
	if err != nil {
		return err
	}

	return nil
}

// ExecuteTransaction executes a transaction from the new transaction execution model and persists the execution.
func (s *Projects) ExecuteTransaction(execution model.NewTransactionExecution) (*model.TransactionExecution, error) {
	projID := execution.ProjectID
	s.loadLock(projID).Lock()
	defer s.removeLock(projID).Unlock()
	emulator, err := s.load(projID)
	if err != nil {
		return nil, err
	}

	signers := make([]flowsdk.Address, len(execution.Signers))
	for i, sig := range execution.Signers {
		signers[i] = sig.ToFlowAddress()
	}

	result, tx, err := emulator.executeTransaction(
		execution.Script,
		execution.Arguments,
		execution.SignersToFlow(),
	)
	if err != nil {
		return nil, err
	}

	exe := model.TransactionExecutionFromFlow(execution.ProjectID, result, tx)
	err = s.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, err
	}

	return exe, nil
}

// ExecuteScript executes the script.
func (s *Projects) ExecuteScript(execution model.NewScriptExecution) (*model.ScriptExecution, error) {
	projID := execution.ProjectID
	s.loadLock(projID).RLock()
	defer s.removeLock(projID).RUnlock()
	emulator, err := s.load(projID)
	if err != nil {
		return nil, err
	}

	result, err := emulator.executeScript(execution.Script, execution.Arguments)
	if err != nil {
		return nil, err
	}

	exe := model.ScriptExecutionFromFlow(
		result,
		projID,
		execution.Script,
		execution.Arguments,
	)
	err = s.store.InsertScriptExecution(exe)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert script execution record")
	}

	return exe, nil
}

// GetAccount by the address along with its storage information.
func (s *Projects) GetAccount(projectID uuid.UUID, address model.Address) (*model.Account, error) {
	s.loadLock(projectID).RLock()
	account, err := s.getAccount(projectID, address)
	s.removeLock(projectID).RUnlock()
	return account, err
}

func (s *Projects) getAccount(projectID uuid.UUID, address model.Address) (*model.Account, error) {
	emulator, err := s.load(projectID)
	if err != nil {
		return nil, err
	}

	flowAccount, store, err := emulator.getAccount(address.ToFlowAddress())
	if err != nil {
		return nil, err
	}

	jsonStorage, err := json.Marshal(store)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling account storage")
	}

	account := model.AccountFromFlow(flowAccount)
	account.ProjectID = projectID
	account.State = string(jsonStorage)

	return account, nil
}

func (s *Projects) CreateInitialAccounts(projectID uuid.UUID, numAccounts int) ([]*model.InternalAccount, error) {
	accounts := make([]*model.InternalAccount, numAccounts)
	for i := 0; i < numAccounts; i++ {
		account, err := s.CreateAccount(projectID)
		if err != nil {
			return nil, err
		}

		accounts[i] = &model.InternalAccount{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: projectID,
			},
			Address: account.Address,
		}
	}

	return accounts, nil
}

// CreateAccount creates a new account and return the account model as well as record the execution.
func (s *Projects) CreateAccount(projectID uuid.UUID) (*model.Account, error) {
	s.loadLock(projectID).Lock()
	defer s.removeLock(projectID).Unlock()
	emulator, err := s.load(projectID)
	if err != nil {
		return nil, err
	}

	account, tx, result, err := emulator.createAccount()
	if err != nil {
		return nil, err
	}

	exe := model.TransactionExecutionFromFlow(projectID, result, tx)
	err = s.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, err
	}

	return model.AccountFromFlow(account), nil
}

// DeployContract deploys a new contract to the provided address and return the updated account as well as record the execution.
func (s *Projects) DeployContract(
	projectID uuid.UUID,
	address model.Address,
	script string,
) (*model.Account, error) {
	s.loadLock(projectID).Lock()
	defer s.removeLock(projectID).Unlock()
	emulator, err := s.load(projectID)
	if err != nil {
		return nil, err
	}

	result, tx, err := emulator.deployContract(address.ToFlowAddress(), script)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}

	exe := model.TransactionExecutionFromFlow(projectID, result, tx)
	err = s.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, err
	}

	return s.getAccount(projectID, address)
}
