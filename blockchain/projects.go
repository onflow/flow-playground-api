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
		flowKitCache:   newFlowKitCache(128),
		mutex:          newMutex(),
		accountsNumber: initAccountsNumber,
		flowKitPool:    newFlowKitPool(10),
	}
}

// Projects implements stateful operations on the blockchain, keeping records of transaction executions.
//
// Projects expose API to interact with the blockchain all in context of a project but also makes sure
// the state is persisted and implements state recreation with caching and resource locking.
type Projects struct {
	store          storage.Store
	flowKitCache   *flowKitCache
	flowKitPool    *flowKitPool
	mutex          *mutex
	accountsNumber int
}

// Reset the blockchain state and return the new account models
func (p *Projects) Reset(projectID uuid.UUID) error {
	var project model.Project
	err := p.store.GetProject(projectID, &project)
	if err != nil {
		return err
	}

	p.flowKitCache.reset(projectID)

	err = p.store.ResetProjectState(&project)
	if err != nil {
		return err
	}

	return nil
}

// ExecuteTransaction executes a transaction from the new transaction execution model and persists the execution.
func (p *Projects) ExecuteTransaction(execution model.NewTransactionExecution) (*model.TransactionExecution, error) {
	projID := execution.ProjectID
	p.mutex.load(projID).Lock()
	defer p.mutex.remove(projID).Unlock()
	fk, err := p.load(projID)
	if err != nil {
		return nil, err
	}

	signers := make([]flowsdk.Address, len(execution.Signers))
	for i, sig := range execution.Signers {
		signers[i] = sig.ToFlowAddress()
	}

	tx, result, err := fk.executeTransaction(
		execution.Script,
		execution.Arguments,
		execution.SignersToFlow(),
	)
	if err != nil {
		return nil, err
	}

	blockHeight, err := fk.getLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	exe := model.TransactionExecutionFromFlow(execution.ProjectID, result, tx, blockHeight)
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
	fk, err := p.load(projID)
	if err != nil {
		return nil, err
	}

	result, err := fk.executeScript(execution.Script, execution.Arguments)
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

// CreateAccount creates a new account and return the account model as well as record the execution.
func (p *Projects) CreateAccount(projectID uuid.UUID) (*model.Account, error) {
	// TODO: Delete this?
	p.mutex.load(projectID).Lock()
	defer p.mutex.remove(projectID).Unlock()
	fk, err := p.load(projectID)
	if err != nil {
		return nil, err
	}

	flowAccount, err := fk.createAccount()
	if err != nil {
		return nil, err
	}

	address := model.NewAddressFromBytes(flowAccount.Address.Bytes())

	return p.getAccount(projectID, address)
}

// DeployContract deploys a new contract to the provided address and return the updated account as well as record the execution.
// If a contract with the same name is already deployed to this address, then it will be updated.
func (p *Projects) DeployContract(
	projectID uuid.UUID,
	address model.Address,
	script string,
	arguments []string,
) (*model.ContractDeployment, error) {
	p.mutex.load(projectID).Lock()
	defer p.mutex.remove(projectID).Unlock()
	fk, err := p.load(projectID)
	if err != nil {
		return nil, err
	}

	contractName, err := parseContractName(script)
	if err != nil {
		return nil, err
	}

	flowAccount, _, err := fk.getAccount(address.ToFlowAddress())
	if err != nil {
		return nil, err
	}

	if _, ok := flowAccount.Contracts[contractName]; ok {
		// A contract with this name has already been deployed to this account
		// Rollback to block height before this contract was initially deployed
		var deployment model.ContractDeployment
		err := p.store.GetContractDeploymentOnAddress(projectID, contractName, address, &deployment)
		if err != nil {
			return nil, err
		}

		blockHeight := deployment.BlockHeight

		// Delete all contract deployments + transaction_executions >= blockHeight
		err = p.store.TruncateDeploymentsAndExecutionsAtBlockHeight(projectID, blockHeight)
		if err != nil {
			return nil, err
		}

		// Reload emulator after block height rollback
		p.flowKitCache.reset(projectID)
		fk, err = p.load(projectID)
		if err != nil {
			return nil, err
		}
	}

	tx, result, err := fk.deployContract(address.ToFlowAddress(), script, arguments)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}

	blockHeight, err := fk.getLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	//exe := model.TransactionExecutionFromFlow(projectID, result, tx, blockHeight) TODO: Remove?
	deploy := model.ContractDeploymentFromFlow(projectID, contractName, script, arguments, result, tx, blockHeight)

	err = p.store.InsertContractDeployment(deploy)
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
	fk, err := p.load(projectID)
	if err != nil {
		return nil, err
	}

	flowAccount, store, err := fk.getAccount(address.ToFlowAddress())
	if err != nil {
		return nil, err
	}

	// TODO: Account storage
	_ = store
	/*
		jsonStorage, err := json.Marshal(store)
		if err != nil {
			return nil, errors.Wrap(err, "error marshaling account storage")
		}
	*/

	account := model.AccountFromFlow(flowAccount, projectID)
	account.ProjectID = projectID
	account.State = "" // TODO: Add account storage

	return account, nil
}

func (p *Projects) GetFlowJson(projectID uuid.UUID) (string, error) {
	fk, err := p.load(projectID)
	if err != nil {
		return "", err
	}

	return fk.getFlowJson()
}

// load initializes an emulator and run transactions previously executed in the project to establish a state.
//
// Do not call this method directly, it is not concurrency safe.
func (p *Projects) load(projectID uuid.UUID) (blockchain, error) {
	fk, err := p.rebuildState(projectID)
	if err != nil {
		err = p.Reset(projectID)
		if err != nil {
			return nil, err
		}

		fk, err = p.rebuildState(projectID)
		if err != nil {
			return nil, err
		}
	}

	p.flowKitCache.add(projectID, fk)

	return fk, nil
}

func (p *Projects) rebuildState(projectID uuid.UUID) (*flowKit, error) {
	var executions []*model.TransactionExecution
	err := p.store.GetTransactionExecutionsForProject(projectID, &executions)
	if err != nil {
		return nil, err
	}

	var deployments []*model.ContractDeployment
	err = p.store.GetContractDeploymentsForProject(projectID, &deployments)
	if err != nil {
		return nil, err
	}

	fk := p.flowKitCache.get(projectID)
	if fk == nil { // if cache miss create new flowKit
		fk, err = p.flowKitPool.new()
		if err != nil {
			return nil, err
		}
	}

	height, err := fk.getLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	// This can happen if project was cleared but on another replica, this replica gets the request after
	// and will get cleared 0 executions from database but has a stale emulator in its own cache
	// This also occurs when a rollback is required due to contract redeployment
	if height > len(executions) {
		p.flowKitCache.reset(projectID)
		fk, err = p.flowKitPool.new()
		if err != nil {
			return nil, err
		}
		height = fk.initBlockHeight()
	}

	fk, err = p.runMissingBlocks(projectID, fk, height, executions, deployments)
	if err != nil {
		return nil, err
	}

	return fk, nil
}

func (p *Projects) getExecutionOrDeploymentAtHeight(
	height int,
	exes []*model.TransactionExecution,
	deploys []*model.ContractDeployment,
) (*model.TransactionExecution, *model.ContractDeployment, error) {
	for _, exec := range exes {
		if exec.BlockHeight == height {
			return exec, nil, nil
		}
	}

	for _, deploy := range deploys {
		if deploy.BlockHeight == height {
			return nil, deploy, nil
		}
	}

	return nil, nil, errors.Errorf("no execution or deployment found at height %d", height)
}

// runMissingBlocks executes missing transactions and deploys missing contracts
// occurring after the specified height
func (p *Projects) runMissingBlocks(
	projectID uuid.UUID,
	fk *flowKit,
	height int,
	exes []*model.TransactionExecution,
	deploys []*model.ContractDeployment,
) (*flowKit, error) {
	totalBlockHeight := fk.initBlockHeight() + len(exes) + len(deploys)

	for height < totalBlockHeight {
		// Add next missing block
		txExec, deploy, err := p.getExecutionOrDeploymentAtHeight(height+1, exes, deploys)
		if err != nil {
			return nil, err
		}
		if txExec != nil {
			_, result, err := fk.executeTransaction(
				txExec.Script,
				txExec.Arguments,
				txExec.SignersToFlow(),
			)
			if err != nil {
				return nil, stateRecreationError(projectID, txExec.ID, err)
			}
			if result.Error != nil && len(txExec.Errors) == 0 {
				return nil, transactionResultError(projectID, txExec.ID, result.Error)
			}
		} else if deploy != nil {
			_, result, err := fk.deployContract(deploy.Address.ToFlowAddress(), deploy.Script, deploy.Arguments)
			if err != nil {
				return nil, stateRecreationError(projectID, deploy.ID, err)
			}
			if result.Error != nil && len(deploys[0].Errors) == 0 {
				return nil, transactionResultError(projectID, deploy.ID, result.Error)
			}
		} else {
			// This should never happen
			err := fmt.Errorf("no execution or deployment found for block height %d", height+1)
			sentry.CaptureException(err)
			return nil, err
		}
		height++
	}

	return fk, nil
}

func stateRecreationError(
	projectID uuid.UUID, exeID uuid.UUID, err error) error {
	err = errors.Wrap(err, fmt.Sprintf(
		"execution error: not able to recreate the project state %s with execution ID %s",
		projectID,
		exeID.String(),
	))
	sentry.CaptureException(err)
	return err
}

func transactionResultError(projectID uuid.UUID, exeID uuid.UUID, resultError error) error {
	err := fmt.Errorf(
		"project %s state recreation failure: execution ID %s failed with result: %s",
		projectID.String(),
		exeID.String(),
		resultError.Error(),
	)
	sentry.CaptureException(err)
	return err
}
