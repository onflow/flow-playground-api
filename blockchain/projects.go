/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/telemetry"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
)

// improvement: create instance pool as a possible optimization. We can pre-instantiate empty
// instances of emulators waiting around to be assigned to a project if init time will be proved to be an issue

// NewProjects creates an instance of the projects with provided storage access and caching.
func NewProjects(store storage.Store, initAccountsNumber int) *Projects {
	return &Projects{
		store:          store,
		cache:          newCache(128),
		mutex:          newMutex(),
		accountsNumber: initAccountsNumber,
	}
}

// Projects implements stateful operations on the blockchain, keeping records of transaction executions.
//
// Projects expose API to interact with the blockchain all in context of a project but also makes sure
// the state is persisted and implements state recreation with caching and resource locking.
type Projects struct {
	store          storage.Store
	cache          *cache
	mutex          *mutex
	accountsNumber int
}

// Reset the blockchain state.
func (p *Projects) Reset(project *model.InternalProject) ([]*model.InternalAccount, error) {
	p.cache.reset(project.ID)
	telemetry.DebugLog("[projects] reset - start")

	err := p.store.ResetProjectState(project)
	if err != nil {
		return nil, err
	}

	telemetry.DebugLog("[projects] reset - project state reset")
	accounts, err := p.CreateInitialAccounts(project.ID)
	if err != nil {
		return nil, err
	}

	telemetry.DebugLog("[projects] reset - accounts created")
	return accounts, nil
}

// ExecuteTransaction executes a transaction from the new transaction execution model and persists the execution.
func (p *Projects) ExecuteTransaction(execution model.NewTransactionExecution) (*model.TransactionExecution, error) {
	telemetry.StartRuntimeCalculation()
	defer telemetry.EndRuntimeCalculation()
	telemetry.DebugLog("[projects] execute transaction - start")

	projID := execution.ProjectID
	p.mutex.load(projID).Lock()
	defer p.mutex.remove(projID).Unlock()
	emulator, err := p.load(projID)
	if err != nil {
		return nil, err
	}

	telemetry.DebugLog("[projects] execute transaction - emulator loaded")

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

	telemetry.DebugLog("[projects] execute transaction - emulator executed")

	exe := model.TransactionExecutionFromFlow(execution.ProjectID, result, tx)
	err = p.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, err
	}

	telemetry.DebugLog("[projects] execute transaction - execution inserted")

	return exe, nil
}

// ExecuteScript executes the script.
func (p *Projects) ExecuteScript(execution model.NewScriptExecution) (*model.ScriptExecution, error) {
	projID := execution.ProjectID
	p.mutex.load(projID).RLock()
	defer p.mutex.remove(projID).RUnlock()
	emulator, err := p.load(projID)
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
	err = p.store.InsertScriptExecution(exe)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert script execution record")
	}

	return exe, nil
}

// GetAccount by the address along with its storage information.
func (p *Projects) GetAccount(projectID uuid.UUID, address model.Address) (*model.Account, error) {
	p.mutex.load(projectID).RLock()
	account, err := p.getAccount(projectID, address)
	p.mutex.remove(projectID).RUnlock()
	return account, err
}

func (p *Projects) CreateInitialAccounts(projectID uuid.UUID) ([]*model.Account, error) {
	telemetry.StartRuntimeCalculation()
	defer telemetry.EndRuntimeCalculation()
	telemetry.DebugLog("[projects] create initial accounts - start")
	accounts := make([]*model.InternalAccount, p.accountsNumber)
	for i := 0; i < p.accountsNumber; i++ {
		account, err := p.CreateAccount(projectID)
		if err != nil {
			return nil, err
		}

		accounts[i] = &model.InternalAccount{
			ProjectChildID: model.NewProjectChildID(uuid.New(), projectID),
			Address:        account.Address,
			Index:          i,
		}
	}
	telemetry.DebugLog("[projects] create initial accounts - end")
	return accounts, nil
}

// CreateAccount creates a new account and return the account model as well as record the execution.
func (p *Projects) CreateAccount(projectID uuid.UUID) (*model.Account, error) {
	telemetry.StartRuntimeCalculation()
	defer telemetry.EndRuntimeCalculation()
	telemetry.DebugLog("[projects] create account")

	p.mutex.load(projectID).Lock()
	defer p.mutex.remove(projectID).Unlock()
	emulator, err := p.load(projectID)
	if err != nil {
		return nil, err
	}

	account, tx, result, err := emulator.createAccount()
	if err != nil {
		return nil, err
	}

	exe := model.TransactionExecutionFromFlow(projectID, result, tx)
	telemetry.DebugLog("[projects] create account - insert executions in store")
	err = p.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, err
	}
	telemetry.DebugLog("[projects] create account - end")
	return model.AccountFromFlow(account, projectID), nil
}

// DeployContract deploys a new contract to the provided address and return the updated account as well as record the execution.
func (p *Projects) DeployContract(
	projectID uuid.UUID,
	address model.Address,
	script string,
) (*model.Account, error) {
	telemetry.StartRuntimeCalculation()
	defer telemetry.EndRuntimeCalculation()

	telemetry.DebugLog("[projects] deploy contract - start")

	p.mutex.load(projectID).Lock()
	defer p.mutex.remove(projectID).Unlock()
	emulator, err := p.load(projectID)
	if err != nil {
		return nil, err
	}

	telemetry.DebugLog("[projects] deploy contract - emulator loaded")

	result, tx, err := emulator.deployContract(address.ToFlowAddress(), script)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}

	telemetry.DebugLog("[projects] deploy contract - contract deployed")

	exe := model.TransactionExecutionFromFlow(projectID, result, tx)
	err = p.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, err
	}

	telemetry.DebugLog("[projects] deploy contract - execution inserted")

	return p.getAccount(projectID, address)
}

func (p *Projects) getAccount(projectID uuid.UUID, address model.Address) (*model.Account, error) {
	telemetry.StartRuntimeCalculation()
	defer telemetry.EndRuntimeCalculation()
	emulator, err := p.load(projectID)
	if err != nil {
		return nil, err
	}

	telemetry.DebugLog("[projects] get account - emulator loaded")

	flowAccount, store, err := emulator.getAccount(address.ToFlowAddress())
	if err != nil {
		return nil, err
	}

	telemetry.DebugLog("[projects] get account - account retrieved from emualator")

	jsonStorage, err := json.Marshal(store)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling account storage")
	}

	account := model.AccountFromFlow(flowAccount, projectID)
	account.ProjectID = projectID
	account.State = string(jsonStorage)

	return account, nil
}

// load initializes an emulator and run transactions previously executed in the project to establish a state.
//
// Do not call this method directly, it is not concurrency safe.
func (p *Projects) load(projectID uuid.UUID) (blockchain, error) {
	telemetry.StartRuntimeCalculation()
	defer telemetry.EndRuntimeCalculation()
	telemetry.DebugLog("[projects] load - start")

	var executions []*model.TransactionExecution
	err := p.store.GetTransactionExecutionsForProject(projectID, &executions)
	if err != nil {
		return nil, err
	}

	telemetry.DebugLog("[projects] load - retrieve executions")

	emulator, executions, err := p.cache.get(projectID, executions)
	if emulator == nil || err != nil {
		emulator, err = newEmulator()
		if err != nil {
			return nil, err
		}
	}

	telemetry.DebugLog("[projects] load - resolve cache")

	for _, execution := range executions {
		result, _, err := emulator.executeTransaction(
			execution.Script,
			execution.Arguments,
			execution.SignersToFlow(),
		)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf(
				"execution error: not able to recreate the project state %s with execution ID %s",
				projectID,
				execution.ID.String(),
			))
		}
		if result.Error != nil && len(execution.Errors) == 0 {
			sentry.CaptureMessage(fmt.Sprintf(
				"project %s state recreation failure: execution ID %s failed with result: %s, debug: %v",
				projectID.String(),
				execution.ID.String(),
				result.Error.Error(),
				result.Debug,
			))
			return nil, errors.Wrap(err, fmt.Sprintf(
				"result error: not able to recreate the project state %s with execution ID %s",
				projectID,
				execution.ID.String(),
			))
		}
	}

	telemetry.DebugLog("[projects] load - executions completed")

	p.cache.add(projectID, emulator)

	return emulator, nil
}
