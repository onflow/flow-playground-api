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
		emulatorCache:  newEmulatorCache(128),
		mutex:          newMutex(),
		accountsNumber: initAccountsNumber,
		emulatorPool:   newEmulatorPool(10),
	}
}

// Projects implements stateful operations on the blockchain, keeping records of transaction executions.
//
// Projects expose API to interact with the blockchain all in context of a project but also makes sure
// the state is persisted and implements state recreation with caching and resource locking.
type Projects struct {
	store          storage.Store
	emulatorCache  *emulatorCache
	emulatorPool   *emulatorPool
	mutex          *mutex
	accountsNumber int
}

// Reset the blockchain state and return the new account models
func (p *Projects) Reset(projectID uuid.UUID) ([]*model.Account, error) {
	var project model.Project
	err := p.store.GetProject(projectID, &project)
	if err != nil {
		return nil, err
	}

	p.emulatorCache.reset(projectID)

	err = p.store.ResetProjectState(&project)
	if err != nil {
		return nil, err
	}

	accounts, err := p.createInitialAccounts(projectID)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// ExecuteTransaction executes a transaction from the new transaction execution model and persists the execution.
func (p *Projects) ExecuteTransaction(execution model.NewTransactionExecution) (*model.TransactionExecution, error) {
	projID := execution.ProjectID
	p.mutex.load(projID).Lock()
	defer p.mutex.remove(projID).Unlock()
	em, err := p.load(projID)
	if err != nil {
		return nil, err
	}

	signers := make([]flowsdk.Address, len(execution.Signers))
	for i, sig := range execution.Signers {
		signers[i] = sig.ToFlowAddress()
	}

	result, tx, err := em.executeTransaction(
		execution.Script,
		execution.Arguments,
		execution.SignersToFlow(),
	)
	if err != nil {
		return nil, err
	}

	exe := model.TransactionExecutionFromFlow(execution.ProjectID, result, tx)
	err = p.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, err
	}

	return exe, nil
}

// ExecuteScript executes the script.
func (p *Projects) ExecuteScript(execution model.NewScriptExecution) (*model.ScriptExecution, error) {
	projID := execution.ProjectID
	p.mutex.load(projID).RLock()
	defer p.mutex.remove(projID).RUnlock()
	em, err := p.load(projID)
	if err != nil {
		return nil, err
	}

	result, err := em.executeScript(execution.Script, execution.Arguments)
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

// CreateInitialAccounts returns the number of accounts that were created
func (p *Projects) CreateInitialAccounts(projectID uuid.UUID) ([]*model.Account, error) {
	p.mutex.load(projectID).Lock()
	defer p.mutex.remove(projectID).Unlock()
	return p.createInitialAccounts(projectID)
}

func (p *Projects) createInitialAccounts(projectID uuid.UUID) ([]*model.Account, error) {
	em, err := p.load(projectID)
	if err != nil {
		return nil, err
	}

	var accounts []*model.Account
	for i := 0; i < p.accountsNumber; i++ {
		_, tx, result, err := em.createAccount()
		if err != nil {
			return nil, err
		}

		exe := model.TransactionExecutionFromFlow(projectID, result, tx)
		err = p.store.InsertTransactionExecution(exe)
		if err != nil {
			return nil, err
		}

		account, err := p.getAccount(projectID, model.NewAddressFromIndex(i))
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// CreateAccount creates a new account and return the account model as well as record the execution.
func (p *Projects) CreateAccount(projectID uuid.UUID) (*model.Account, error) {
	p.mutex.load(projectID).Lock()
	defer p.mutex.remove(projectID).Unlock()
	em, err := p.load(projectID)
	if err != nil {
		return nil, err
	}

	account, tx, result, err := em.createAccount()
	if err != nil {
		return nil, err
	}

	exe := model.TransactionExecutionFromFlow(projectID, result, tx)
	err = p.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, err
	}

	address := model.NewAddressFromBytes(account.Address.Bytes())

	return p.getAccount(projectID, address)
}

// DeployContract deploys a new contract to the provided address and return the updated account as well as record the execution.
// If a contract with the same name is already deployed to this address, then it will be updated.
func (p *Projects) DeployContract(
	projectID uuid.UUID,
	address model.Address,
	script string,
) (*model.ContractDeployment, error) {
	p.mutex.load(projectID).Lock()
	defer p.mutex.remove(projectID).Unlock()
	em, err := p.load(projectID)
	if err != nil {
		return nil, err
	}

	contractName, err := parseContractName(script)
	if err != nil {
		return nil, err
	}

	flowAccount, _, err := em.getAccount(address.ToFlowAddress())
	if err != nil {
		return nil, err
	}

	if _, ok := flowAccount.Contracts[contractName]; ok {
		// A contract with this name has already been deployed to this account
		// Remove previous contract from the account
		result, tx, err := em.removeContract(address.ToFlowAddress(), contractName)
		if err != nil {
			return nil, err
		}
		if result.Error != nil {
			return nil, result.Error
		}

		// Insert transaction execution of contract removal into db
		exe := model.TransactionExecutionFromFlow(projectID, result, tx)
		err = p.store.InsertTransactionExecution(exe)
		if err != nil {
			return nil, err
		}

		// Remove contract deployment from db
		err = p.store.DeleteContractDeploymentByName(projectID, address, contractName)
		if err != nil {
			return nil, err
		}
	}

	result, tx, err := em.deployContract(address.ToFlowAddress(), script)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}

	exe := model.TransactionExecutionFromFlow(projectID, result, tx)
	deploy := model.ContractDeploymentFromFlow(projectID, contractName, script, result, tx)

	err = p.store.InsertContractDeploymentWithExecution(deploy, exe)
	if err != nil {
		return nil, err
	}

	return deploy, nil
}

// GetAccount by the address along with its storage information.
func (p *Projects) GetAccount(projectID uuid.UUID, address model.Address) (*model.Account, error) {
	p.mutex.load(projectID).RLock()
	defer p.mutex.remove(projectID).RUnlock()
	return p.getAccount(projectID, address)
}

func (p *Projects) GetAccounts(projectID uuid.UUID, addresses []model.Address) ([]*model.Account, error) {
	p.mutex.load(projectID).RLock()
	defer p.mutex.remove(projectID).RUnlock()

	accounts := make([]*model.Account, len(addresses))
	for i, address := range addresses {
		account, err := p.getAccount(projectID, address)
		if err != nil {
			return nil, err
		}

		accounts[i] = account
	}

	return accounts, nil
}

func (p *Projects) getAccount(projectID uuid.UUID, address model.Address) (*model.Account, error) {
	em, err := p.load(projectID)
	if err != nil {
		return nil, err
	}

	flowAccount, store, err := em.getAccount(address.ToFlowAddress())
	if err != nil {
		return nil, err
	}

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
	var executions []*model.TransactionExecution

	err := p.store.GetTransactionExecutionsForProject(projectID, &executions)
	if err != nil {
		return nil, err
	}

	em := p.emulatorCache.get(projectID)
	if em == nil { // if cache miss create new emulator
		em, err = p.emulatorPool.new()
		if err != nil {
			return nil, err
		}
	}

	height, err := em.getLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	// this can happen if project was cleared but on another replica, this replica gets the request after
	// and will get cleared 0 executions from database but has a stale emulator in its own cache
	if height > len(executions) {
		p.emulatorCache.reset(projectID)
		em, err = newEmulator()
		if err != nil {
			return nil, err
		}
		height = 0
	}

	executions, err = p.filterMissingExecutions(executions, height)
	if err != nil {
		return nil, err
	}

	em, err = p.runMissingExecutions(projectID, em, executions)
	if err != nil {
		return nil, err
	}

	p.emulatorCache.add(projectID, em)

	return em, nil
}

func (p *Projects) runMissingExecutions(
	projectID uuid.UUID,
	em *emulator,
	executions []*model.TransactionExecution) (*emulator, error) {

	for _, execution := range executions {
		result, _, err := em.executeTransaction(
			execution.Script,
			execution.Arguments,
			execution.SignersToFlow(),
		)
		if err != nil {
			err := errors.Wrap(err, fmt.Sprintf(
				"execution error: not able to recreate the project state %s with execution ID %s",
				projectID,
				execution.ID.String(),
			))

			sentry.CaptureException(err)
			return nil, err
		}
		if result.Error != nil && len(execution.Errors) == 0 {
			err := fmt.Errorf(
				"project %s state recreation failure: execution ID %s failed with result: %s, debug: %v",
				projectID.String(),
				execution.ID.String(),
				result.Error.Error(),
				result.Debug,
			)

			sentry.CaptureException(err)
			return nil, err
		}
	}

	return em, nil
}

// filterMissingExecutions gets all missed executions in current replica cache
//
// based on the executions the function receives it compares that to the emulator block height, since
// one execution is always one block it can compare the heights to the length. If it finds some executions
// that are not part of emulator it returns that subset, so they can be applied on top.
func (p *Projects) filterMissingExecutions(
	executions []*model.TransactionExecution,
	height int,
) ([]*model.TransactionExecution, error) {
	// this will only set executions that are missing from the emulator
	executions = executions[height:]

	return executions, nil
}
