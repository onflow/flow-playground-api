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
	"sync"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"

	"github.com/getsentry/sentry-go"
	"github.com/golang/groupcache/lru"
	"github.com/google/uuid"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
)

// improvement: create instance pool as a possible optimization. We can pre-instantiate empty
// instances of emulators waiting around to be assigned to a project if init time will be proved to be an issue

// NewProjects creates an instance of the projects with provided storage access and caching.
func NewProjects(store storage.Store, cache *lru.Cache) *Projects {
	return &Projects{
		store: store,
		cache: cache,
	}
}

// Projects implements stateful operations on the blockchain, keeping records of transaction executions.
//
// Projects expose API to interact with the blockchain all in context of a project but also makes sure
// the state is persisted and implements state recreation with caching and resource locking.
type Projects struct {
	store     storage.Store
	cache     *lru.Cache
	mu        sync.Map
	muCounter sync.Map
}

// load initializes an emulator and run transactions previously executed in the project to establish a state.
//
// Do not call this method directly, it is not concurrency safe.
func (s *Projects) load(projectID uuid.UUID) (blockchain, error) {
	val, ok := s.cache.Get(projectID)
	if ok {
		return val.(blockchain), nil
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
		result, _, err := emulator.executeTransaction(
			execution.Script,
			execution.Arguments,
			execution.SignersToFlowWithoutTranslation(),
		)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("not able to recreate the project state %s", projectID))
		}
		if result.Error != nil && len(execution.Errors) == 0 {
			sentry.CaptureMessage(fmt.Sprintf(
				"project %s state recreation failure: execution %s failed with result: %s, debug: %v",
				projectID.String(),
				execution.ID.String(),
				result.Error.Error(),
				result.Debug,
			))
			return nil, errors.Wrap(result.Error, fmt.Sprintf("not able to recreate the project state %s", projectID))
		}
	}

	s.cache.Add(projectID, emulator)

	return emulator, nil
}

// loadLock retrieves the mutex lock by the project ID and increase the usage counter.
func (s *Projects) loadLock(uuid uuid.UUID) *sync.RWMutex {
	counter, _ := s.muCounter.LoadOrStore(uuid, 0)
	s.muCounter.Store(uuid, counter.(int)+1)

	mu, _ := s.mu.LoadOrStore(uuid, &sync.RWMutex{})
	return mu.(*sync.RWMutex)
}

// removeLock returns the mutex lock by the project ID and decreases usage counter, deleting the map entry if at 0.
func (s *Projects) removeLock(uuid uuid.UUID) *sync.RWMutex {
	m, ok := s.mu.Load(uuid)
	if !ok {
		sentry.CaptureMessage("trying to access non-existing mutex")
	}

	counter, ok := s.muCounter.Load(uuid)
	if !ok {
		sentry.CaptureMessage("trying to access non-existing mutex counter")
	}

	if counter == 0 {
		s.mu.Delete(uuid)
		s.muCounter.Delete(uuid)
	} else {
		s.muCounter.Store(uuid, counter.(int)-1)
	}

	return m.(*sync.RWMutex)
}

// Reset the blockchain state.
func (s *Projects) Reset(project *model.InternalProject) error {
	s.cache.Remove(project.ID)

	err := s.store.ResetProjectState(project)
	if err != nil {
		return err
	}

	_, err = s.CreateInitialAccounts(project.ID, 5) // todo don't pass number literal
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

	account := model.AccountFromFlow(flowAccount, projectID)
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

	return model.AccountFromFlow(account, projectID), nil
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
