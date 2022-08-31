package blockchain

import (
	"fmt"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/onflow/flow-emulator/types"
)

// todo create instance pool as a possible optimization: we can pre-instantiate empty instances of emulators waiting around to be assigned to a project if init time will be proved to be an issue

// access cache and add to cache

// bootstrap emulator with transactions

type State struct {
	store storage.Store
	// cache
}

// bootstrap initializes an emulator and run transactions previously executed in the project to establish a state.
func (s *State) bootstrap(ID uuid.UUID) (*Emulator, error) {
	// todo cache

	emulator, err := NewEmulator()
	if err != nil {
		return nil, err
	}

	var executions []*model.TransactionExecution
	err = s.store.GetTransactionExecutionsForProject(ID, &executions)
	if err != nil {
		return nil, err
	}

	for _, execution := range executions {
		// todo BE CAREFUL: there will be transactions recorded in transaction execution that failed, so they will fail again - treat that with care
		result, err := emulator.ExecuteTransaction(execution.Script, execution.Arguments, nil)
		if err != nil || (!result.Succeeded() && len(execution.Errors) == 0) {
			// todo handle a case where an existing project is not able to be recreated - track this in sentry
			return nil, fmt.Errorf(fmt.Sprintf("not able to recreate a project %s", ID))
		}
	}

	return emulator, nil
}

func (s *State) new(ID uuid.UUID) (*Emulator, error) {
	return NewEmulator()
}

func (s *State) ExecuteTransaction(
	ID uuid.UUID,
	script string,
	arguments []string,
	authorizers []model.Address,
) (*types.TransactionResult, error) {
	emulator, err := s.bootstrap(ID)
	if err != nil {
		return nil, err
	}

	result, err := emulator.ExecuteTransaction(script, arguments, authorizers)
	if err != nil {
		return nil, err
	}

	err = s.store.InsertTransactionExecution(&model.TransactionExecution{
		ProjectChildID:   ID,
		Index:            0,
		Script:           script,
		Arguments:        arguments,
		SignerAccountIDs: signers,
		Errors:           result.Error,
		Events:           result.Events,
		Logs:             result.Logs,
	})
	if err != nil {
		return nil, err
	}

	return result, err
}
