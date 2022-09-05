package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"

	"github.com/golang/groupcache/lru"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// improvement: create instance pool as a possible optimization. We can pre-instantiate empty
// instances of emulators waiting around to be assigned to a project if init time will be proved to be an issue

// NewState creates an instance of the state with provided storage access and caching.
func NewState(store storage.Store, cache *lru.Cache) *State {
	return &State{
		store: store,
		cache: cache,
		mu:    map[uuid.UUID]*sync.RWMutex{},
	}
}

// State implements stateful operations on the blockchain, keeping records of transaction executions.
//
// State exposes API to interact with the blockchain but also makes sure the state is persisted and
// implements state recreation with caching and resource locking.
type State struct {
	store storage.Store
	cache *lru.Cache
	mu    map[uuid.UUID]*sync.RWMutex
}

// load initializes an emulator and run transactions previously executed in the project to establish a state.
func (s *State) load(projectID uuid.UUID) (*emulator, error) {
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
		result, _, err := emulator.executeTransaction(execution.Script, execution.Arguments, execution.Signers)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("not able to recreate the project state %s", projectID))
		}
		if result.Error != nil && len(execution.Errors) == 0 {
			return nil, errors.Wrap(result.Error, fmt.Sprintf("not able to recreate the project state %s", projectID))
		}
	}

	s.cache.Add(projectID, emulator)

	return emulator, nil
}

func (s *State) lock(uuid uuid.UUID) {
	m, ok := s.mu[uuid]
	if !ok {
		m = &sync.RWMutex{}
	}

	m.Lock()
}

func (s *State) unlock(uuid uuid.UUID) {
	m := s.mu[uuid]
	m.Unlock()
	delete(s.mu, uuid)
}

func (s *State) readLock(uuid uuid.UUID) {
	m, ok := s.mu[uuid]
	if !ok {
		m = &sync.RWMutex{}
	}

	m.RLock()
}

func (s *State) readUnlock(uuid uuid.UUID) {
	m := s.mu[uuid]
	m.RUnlock()
	delete(s.mu, uuid)
}

// ExecuteTransaction executes a transaction from the new transaction execution model and persists the execution.
func (s *State) ExecuteTransaction(execution model.NewTransactionExecution) (*model.TransactionExecution, error) {
	projID := execution.ProjectID
	s.lock(projID)
	defer s.unlock(projID)
	emulator, err := s.load(projID)
	if err != nil {
		return nil, err
	}

	result, tx, err := emulator.executeTransaction(execution.Script, execution.Arguments, execution.Signers)
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
func (s *State) ExecuteScript(execution model.NewScriptExecution) (*model.ScriptExecution, error) {
	projID := execution.ProjectID
	s.readLock(projID)
	defer s.readUnlock(projID)
	emulator, err := s.load(projID)
	if err != nil {
		return nil, err
	}

	result, err := emulator.executeScript(execution.Script, execution.Arguments)
	if err != nil {
		return nil, err
	}

	exe := model.ScriptExecutionFromFlow(result, projID, execution.Script, execution.Arguments)
	err = s.store.InsertScriptExecution(exe)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert script execution record")
	}

	return exe, nil
}

// GetAccount by the address along with its storage information.
func (s *State) GetAccount(projectID uuid.UUID, address model.Address) (*model.Account, error) {
	s.readLock(projectID)
	account, err := s.getAccount(projectID, address)
	s.readUnlock(projectID)
	return account, err
}

func (s *State) getAccount(projectID uuid.UUID, address model.Address) (*model.Account, error) {
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

// CreateAccount creates a new account and return the account model as well as record the execution.
func (s *State) CreateAccount(projectID uuid.UUID) (*model.Account, error) {
	s.lock(projectID)
	defer s.unlock(projectID)
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
func (s *State) DeployContract(projectID uuid.UUID, address model.Address, script string) (*model.Account, error) {
	s.lock(projectID)
	defer s.unlock(projectID)
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
