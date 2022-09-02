package blockchain

import (
	"encoding/json"
	"fmt"

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
	return &State{store, cache}
}

// State implements stateful operations on the blockchain, keeping records of transaction executions.
//
// State exposes API to interact with the blockchain but also makes sure the state is persisted and
// implements state recreation with caching and resource locking.
type State struct {
	store storage.Store
	cache *lru.Cache
}

// bootstrap initializes an emulator and run transactions previously executed in the project to establish a state.
func (s *State) bootstrap(projectID uuid.UUID) (*emulator, error) {
	// todo add locking of resources
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

// ExecuteTransaction executes a transaction from the new transaction execution model and persists the execution.
func (s *State) ExecuteTransaction(execution model.NewTransactionExecution) (*model.TransactionExecution, error) {
	emulator, err := s.bootstrap(execution.ProjectID)
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

	return exe, err
}

// ExecuteScript executes the script.
func (s *State) ExecuteScript(
	projectID uuid.UUID,
	execution model.NewScriptExecution,
) (*model.ScriptExecution, error) {
	emulator, err := s.bootstrap(projectID)
	if err != nil {
		return nil, err
	}

	result, err := emulator.executeScript(execution.Script, execution.Arguments)
	if err != nil {
		return nil, err
	}

	return model.ScriptExecutionFromFlow(result, projectID, execution.Script, execution.Arguments)
}

// GetAccount by the address along with its storage information.
func (s *State) GetAccount(projectID uuid.UUID, address model.Address) (*model.Account, error) {
	emulator, err := s.bootstrap(projectID)
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
	emulator, err := s.bootstrap(projectID)
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
	emulator, err := s.bootstrap(projectID)
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

	account, err := s.GetAccount(projectID, address)
	if err != nil {
		return nil, err
	}

	return account, nil
}
