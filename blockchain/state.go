package blockchain

import (
	"encoding/json"
	"fmt"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
)

// todo create instance pool as a possible optimization: we can pre-instantiate empty instances of emulators waiting around to be assigned to a project if init time will be proved to be an issue

func NewState(store storage.Store) *State {
	return &State{store}
}

type State struct {
	store storage.Store
	// cache
}

// bootstrap initializes an emulator and run transactions previously executed in the project to establish a state.
func (s *State) bootstrap(projectID uuid.UUID) (*emulator, error) {
	// todo cache

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
		// todo BE CAREFUL: there will be transactions recorded in transaction execution that failed, so they will fail again - treat that with care
		result, err := emulator.executeTransaction(execution.Script, execution.Arguments, execution.Signers)
		if err != nil || (!result.Succeeded() && len(execution.Errors) == 0) {
			// todo refactor - handle a case where an existing project is not able to be recreated - track this in sentry
			return nil, fmt.Errorf(fmt.Sprintf("not able to recreate a project %s", projectID))
		}
	}

	return emulator, nil
}

func (s *State) ExecuteTransaction(
	projectID uuid.UUID,
	script string,
	arguments []string,
	authorizers []model.Address,
) (*model.TransactionExecution, error) {
	emulator, err := s.bootstrap(projectID)
	if err != nil {
		return nil, err
	}

	result, err := emulator.executeTransaction(script, arguments, authorizers)
	if err != nil {
		return nil, err
	}

	exe, err := model.TransactionExecutionFromFlow(result, projectID, script, arguments, authorizers)
	if err != nil {
		return nil, err
	}

	err = s.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, err
	}

	return exe, err
}

func (s *State) ExecuteScript(projectID uuid.UUID, script string, arguments []string) (*model.ScriptExecution, error) {
	emulator, err := s.bootstrap(projectID)
	if err != nil {
		return nil, err
	}

	result, err := emulator.executeScript(script, arguments)
	if err != nil {
		return nil, err
	}

	return model.ScriptExecutionFromFlow(result, projectID, script, arguments)
}

func (s *State) GetAccount(projectID uuid.UUID, address model.Address) (*model.Account, error) {
	emulator, err := s.bootstrap(projectID)
	if err != nil {
		return nil, err
	}

	account, store, err := emulator.getAccount(address)
	if err != nil {
		return nil, err
	}

	jsonStorage, _ := json.Marshal(store)

	var addr model.Address
	copy(address[:], account.Address[:])

	contractNames := make([]string, 0)
	contractCode := ""
	for name, code := range account.Contracts {
		contractNames = append(contractNames, name)
		contractCode = string(code)
		break // we only allow one deployed contract on account so only get the first if present
	}

	// todo refactor think about defining a different account model, blockchain account or similar
	return &model.Account{
		ProjectID:         projectID,
		Address:           addr,
		DeployedCode:      contractCode,
		DeployedContracts: contractNames,
		State:             string(jsonStorage),
	}, nil
}

func (s *State) CreateAccount(projectID uuid.UUID) (*flowsdk.Account, error) {
	emulator, err := s.bootstrap(projectID)
	if err != nil {
		return nil, err
	}

	account, tx, result, err := emulator.createAccount()
	if err != nil {
		return nil, err
	}

	exe, err := model.TransactionExecutionFromFlowSDK(projectID, result, tx)
	if err != nil {
		return nil, err
	}

	err = s.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// todo check return types what is needed and should we convert to model types
func (s *State) DeployContract(projectID uuid.UUID, address model.Address, script string) (*model.Account, error) {
	emulator, err := s.bootstrap(projectID)
	if err != nil {
		return nil, err
	}

	result, tx, err := emulator.deployContract(address, script)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}

	exe, err := model.TransactionExecutionFromFlowSDK(projectID, result, tx)
	if err != nil {
		return nil, err
	}

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
