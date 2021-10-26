/*
 * Flow Playground
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package memory

import (
	"sort"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/uuid"
	"github.com/onflow/flow-go/engine/execution/state/delta"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Store struct {
	mut                   sync.RWMutex
	users                 map[uuid.UUID]model.User
	projects              map[uuid.UUID]model.InternalProject
	accounts              map[uuid.UUID]model.InternalAccount
	transactionTemplates  map[uuid.UUID]model.TransactionTemplate
	transactionExecutions map[uuid.UUID]model.TransactionExecution
	scriptTemplates       map[uuid.UUID]model.ScriptTemplate
	scriptExecutions      map[uuid.UUID]model.ScriptExecution
	registerDeltas        map[uuid.UUID][]model.RegisterDelta
}

func NewStore() *Store {
	return &Store{
		mut:                   sync.RWMutex{},
		users:                 make(map[uuid.UUID]model.User),
		projects:              make(map[uuid.UUID]model.InternalProject),
		accounts:              make(map[uuid.UUID]model.InternalAccount),
		transactionTemplates:  make(map[uuid.UUID]model.TransactionTemplate),
		transactionExecutions: make(map[uuid.UUID]model.TransactionExecution),
		scriptTemplates:       make(map[uuid.UUID]model.ScriptTemplate),
		scriptExecutions:      make(map[uuid.UUID]model.ScriptExecution),
		registerDeltas:        make(map[uuid.UUID][]model.RegisterDelta),
	}
}

func (s *Store) InsertUser(user *model.User) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.users[user.ID] = *user
	return nil
}

func (s *Store) GetUser(id uuid.UUID, user *model.User) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return storage.ErrNotFound
	}

	*user = u

	return nil
}

func (s *Store) CreateProject(
	proj *model.InternalProject,
	deltas []delta.Delta,
	accounts []*model.InternalAccount,
	ttpls []*model.TransactionTemplate,
	stpls []*model.ScriptTemplate,
) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	if err := s.insertProject(proj); err != nil {
		return err
	}

	for _, delta := range deltas {
		if err := s.insertRegisterDelta(proj.ID, delta, true); err != nil {
			return err
		}
	}

	for _, account := range accounts {
		if err := s.insertAccount(account); err != nil {
			return err
		}
	}

	for _, ttpl := range ttpls {
		if err := s.insertTransactionTemplate(ttpl); err != nil {
			return err
		}
	}

	for _, stpl := range stpls {
		if err := s.insertScriptTemplate(stpl); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) insertProject(proj *model.InternalProject) error {
	proj.CreatedAt = time.Now()
	proj.UpdatedAt = proj.CreatedAt

	s.projects[proj.ID] = *proj
	return nil
}

func (s *Store) UpdateProject(input model.UpdateProject, proj *model.InternalProject) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	p, ok := s.projects[input.ID]
	if !ok {
		return storage.ErrNotFound
	}

	if input.Title != nil {
		p.Title = *input.Title
	}

	if input.Description != nil {
		p.Description = *input.Description
	}

	if input.Readme != nil {
		p.Readme = *input.Readme
	}

	if input.Persist != nil {
		p.Persist = *input.Persist
	}

	s.projects[input.ID] = p

	*proj = p

	err := s.markProjectUpdatedAt(input.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateProjectOwner(id, userID uuid.UUID) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	p, ok := s.projects[id]
	if !ok {
		return storage.ErrNotFound
	}

	p.UserID = userID

	s.projects[id] = p

	return nil
}

func (s *Store) UpdateProjectVersion(id uuid.UUID, version *semver.Version) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	p, ok := s.projects[id]
	if !ok {
		return storage.ErrNotFound
	}

	p.Version = version

	s.projects[id] = p

	return nil
}

func (s *Store) GetProject(id uuid.UUID, proj *model.InternalProject) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	p, ok := s.projects[id]
	if !ok {
		return storage.ErrNotFound
	}

	*proj = p

	return nil
}

func (s *Store) markProjectUpdatedAt(id uuid.UUID) error {
	p, ok := s.projects[id]
	if !ok {
		return storage.ErrNotFound
	}

	p.UpdatedAt = time.Now()

	s.projects[id] = p

	return nil
}

func (s *Store) InsertAccount(acc *model.InternalAccount) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	return s.insertAccount(acc)
}

func (s *Store) insertAccount(acc *model.InternalAccount) error {
	s.accounts[acc.ID] = *acc

	err := s.markProjectUpdatedAt(acc.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAccount(id model.ProjectChildID, acc *model.InternalAccount) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	p, ok := s.accounts[id.ID]
	if !ok {
		return storage.ErrNotFound
	}

	*acc = p

	return nil
}

func (s *Store) UpdateAccount(input model.UpdateAccount, acc *model.InternalAccount) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	return s.updateAccount(input, acc)
}

func (s *Store) updateAccount(input model.UpdateAccount, acc *model.InternalAccount) error {
	a, ok := s.accounts[input.ID]
	if !ok {
		return storage.ErrNotFound
	}

	if input.DraftCode != nil {
		a.DraftCode = *input.DraftCode
	}

	if input.DeployedCode != nil {
		a.DeployedCode = *input.DeployedCode
	}

	if input.DeployedContracts != nil {
		a.DeployedContracts = *input.DeployedContracts
	}

	s.accounts[input.ID] = a

	*acc = a

	err := s.markProjectUpdatedAt(a.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateAccountAfterDeployment(
	input model.UpdateAccount,
	states map[uuid.UUID]model.AccountState,
	delta delta.Delta,
	acc *model.InternalAccount,
) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	err := s.updateAccount(input, acc)
	if err != nil {
		return err
	}

	for accountID, state := range states {
		err = s.updateAccountState(accountID, state)
		if err != nil {
			return err
		}
	}

	err = s.insertRegisterDelta(input.ProjectID, delta, false)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) updateAccountState(id uuid.UUID, state model.AccountState) error {
	a := s.accounts[id]

	a.SetState(state)

	s.accounts[id] = a

	err := s.markProjectUpdatedAt(a.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAccountsForProject(projectID uuid.UUID, accs *[]*model.InternalAccount) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.getAccountsForProject(projectID, accs)
}

func (s *Store) getAccountsForProject(projectID uuid.UUID, accs *[]*model.InternalAccount) error {
	res := make([]*model.InternalAccount, 0)

	for _, acc := range s.accounts {
		if acc.ProjectID == projectID {
			a := acc
			res = append(res, &a)
		}
	}

	// sort results by index
	sort.Slice(res, func(i, j int) bool { return res[i].Index < res[j].Index })

	*accs = res

	return nil
}

func (s *Store) DeleteAccount(id model.ProjectChildID) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	_, ok := s.accounts[id.ID]
	if !ok {
		return storage.ErrNotFound
	}

	delete(s.accounts, id.ID)

	err := s.markProjectUpdatedAt(id.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) InsertTransactionTemplate(tpl *model.TransactionTemplate) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	return s.insertTransactionTemplate(tpl)
}

func (s *Store) insertTransactionTemplate(tpl *model.TransactionTemplate) error {
	var tpls []*model.TransactionTemplate
	err := s.getTransactionTemplatesForProject(tpl.ProjectID, &tpls)
	if err != nil {
		return err
	}

	count := len(tpls)

	// set index to one after last
	tpl.Index = count

	s.transactionTemplates[tpl.ID] = *tpl

	err = s.markProjectUpdatedAt(tpl.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateTransactionTemplate(
	input model.UpdateTransactionTemplate,
	tpl *model.TransactionTemplate,
) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	t, ok := s.transactionTemplates[input.ID]
	if !ok {
		return storage.ErrNotFound
	}

	if input.Title != nil {
		t.Title = *input.Title
	}

	if input.Index != nil {
		t.Index = *input.Index
	}

	if input.Script != nil {
		t.Script = *input.Script
	}

	s.transactionTemplates[input.ID] = t

	*tpl = t

	err := s.markProjectUpdatedAt(t.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetTransactionTemplate(id model.ProjectChildID, tpl *model.TransactionTemplate) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	t, ok := s.transactionTemplates[id.ID]
	if !ok {
		return storage.ErrNotFound
	}

	*tpl = t

	return nil
}

func (s *Store) GetTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.getTransactionTemplatesForProject(projectID, tpls)
}

func (s *Store) getTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error {
	res := make([]*model.TransactionTemplate, 0)

	for _, tpl := range s.transactionTemplates {
		if tpl.ProjectID == projectID {
			t := tpl
			res = append(res, &t)
		}
	}

	// sort results by index
	sort.Slice(res, func(i, j int) bool { return res[i].Index < res[j].Index })

	*tpls = res

	return nil
}

func (s *Store) DeleteTransactionTemplate(id model.ProjectChildID) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	_, ok := s.transactionTemplates[id.ID]
	if !ok {
		return storage.ErrNotFound
	}

	delete(s.transactionTemplates, id.ID)

	err := s.markProjectUpdatedAt(id.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) InsertTransactionExecution(
	exe *model.TransactionExecution,
	states map[uuid.UUID]model.AccountState,
	delta delta.Delta,
) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	var exes []*model.TransactionExecution
	err := s.getTransactionExecutionsForProject(exe.ProjectID, &exes)
	if err != nil {
		return err
	}

	count := len(exes)

	// set index to one after last
	exe.Index = count

	s.transactionExecutions[exe.ID] = *exe

	for accountID, state := range states {
		err = s.updateAccountState(accountID, state)
		if err != nil {
			return err
		}
	}

	err = s.insertRegisterDelta(exe.ProjectID, delta, false)
	if err != nil {
		return err
	}

	err = s.markProjectUpdatedAt(exe.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.getTransactionExecutionsForProject(projectID, exes)
}

func (s *Store) getTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error {
	res := make([]*model.TransactionExecution, 0)

	for _, exe := range s.transactionExecutions {
		if exe.ProjectID == projectID {
			e := exe
			res = append(res, &e)
		}
	}

	// sort results by index
	sort.Slice(res, func(i, j int) bool { return res[i].Index < res[j].Index })

	*exes = res

	return nil
}

func (s *Store) InsertScriptTemplate(tpl *model.ScriptTemplate) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	return s.insertScriptTemplate(tpl)
}

func (s *Store) insertScriptTemplate(tpl *model.ScriptTemplate) error {
	var tpls []*model.ScriptTemplate
	err := s.getScriptTemplatesForProject(tpl.ProjectID, &tpls)
	if err != nil {
		return err
	}

	count := len(tpls)

	// set index to one after last
	tpl.Index = count

	s.scriptTemplates[tpl.ID] = *tpl

	err = s.markProjectUpdatedAt(tpl.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateScriptTemplate(
	input model.UpdateScriptTemplate,
	tpl *model.ScriptTemplate,
) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	t, ok := s.scriptTemplates[input.ID]
	if !ok {
		return storage.ErrNotFound
	}

	if input.Title != nil {
		t.Title = *input.Title
	}

	if input.Index != nil {
		t.Index = *input.Index
	}

	if input.Script != nil {
		t.Script = *input.Script
	}

	s.scriptTemplates[input.ID] = t

	*tpl = t

	err := s.markProjectUpdatedAt(t.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetScriptTemplate(id model.ProjectChildID, tpl *model.ScriptTemplate) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	t, ok := s.scriptTemplates[id.ID]
	if !ok {
		return storage.ErrNotFound
	}

	*tpl = t

	return nil
}

func (s *Store) GetScriptTemplatesForProject(projectID uuid.UUID, tpls *[]*model.ScriptTemplate) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.getScriptTemplatesForProject(projectID, tpls)
}

func (s *Store) getScriptTemplatesForProject(projectID uuid.UUID, tpls *[]*model.ScriptTemplate) error {
	res := make([]*model.ScriptTemplate, 0)

	for _, tpl := range s.scriptTemplates {
		if tpl.ProjectID == projectID {
			t := tpl
			res = append(res, &t)
		}
	}

	// sort results by index
	sort.Slice(res, func(i, j int) bool { return res[i].Index < res[j].Index })

	*tpls = res

	return nil
}

func (s *Store) DeleteScriptTemplate(id model.ProjectChildID) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	_, ok := s.scriptTemplates[id.ID]
	if !ok {
		return storage.ErrNotFound
	}

	delete(s.scriptTemplates, id.ID)

	err := s.markProjectUpdatedAt(id.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) InsertScriptExecution(exe *model.ScriptExecution) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	var exes []*model.ScriptExecution
	err := s.getScriptExecutionsForProject(exe.ProjectID, &exes)
	if err != nil {
		return err
	}

	count := len(exes)

	// set index to one after last
	exe.Index = count

	s.scriptExecutions[exe.ID] = *exe

	err = s.markProjectUpdatedAt(exe.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.getScriptExecutionsForProject(projectID, exes)
}

func (s *Store) getScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error {
	res := make([]*model.ScriptExecution, 0)

	for _, exe := range s.scriptExecutions {
		if exe.ProjectID == projectID {
			e := exe
			res = append(res, &e)
		}
	}

	// sort results by index
	sort.Slice(res, func(i, j int) bool { return res[i].Index < res[j].Index })

	*exes = res

	return nil
}

func (s *Store) insertRegisterDelta(projectID uuid.UUID, delta delta.Delta, isAccountCreation bool) error {
	p, ok := s.projects[projectID]
	if !ok {
		return storage.ErrNotFound
	}

	index := p.TransactionCount + 1

	regDelta := model.RegisterDelta{
		ProjectID: projectID,
		Index:     index,
		Delta:     delta,
	}

	s.registerDeltas[projectID] = append(s.registerDeltas[projectID], regDelta)

	p.TransactionCount = index

	s.projects[projectID] = p

	return nil
}

func (s *Store) GetRegisterDeltasForProject(projectID uuid.UUID, deltas *[]*model.RegisterDelta) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	res := make([]*model.RegisterDelta, 0)

	for _, d := range s.registerDeltas[projectID] {
		dCopy := d
		res = append(res, &dCopy)
	}

	*deltas = res

	return nil
}

func (s *Store) ResetProjectState(newDeltas []delta.Delta, proj *model.InternalProject) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	// clear deployed code from accounts

	for accountID, account := range s.accounts {
		if account.ProjectID != proj.ID {
			continue
		}

		account.DeployedCode = ""
		account.DeployedContracts = nil

		s.accounts[accountID] = account
	}

	// replace existing register deltas with new deltas

	newRegisterDeltas := make([]model.RegisterDelta, len(newDeltas))

	for i, delta := range newDeltas {
		registerDelta := model.RegisterDelta{
			ProjectID: proj.ID,
			Index:     i,
			Delta:     delta,
		}

		newRegisterDeltas = append(newRegisterDeltas, registerDelta)
	}

	s.registerDeltas[proj.ID] = newRegisterDeltas

	// update transaction count

	project := s.projects[proj.ID]
	project.TransactionCount = len(newRegisterDeltas)
	s.projects[proj.ID] = project

	*proj = project

	// delete all transaction executions

	for txExecutionID, txExecution := range s.transactionExecutions {
		if txExecution.ProjectID != proj.ID {
			continue
		}

		delete(s.transactionExecutions, txExecutionID)
	}

	// delete all scripts executions

	for scriptExecutionID, scriptExecution := range s.scriptExecutions {
		if scriptExecution.ProjectID != proj.ID {
			continue
		}

		delete(s.scriptExecutions, scriptExecutionID)
	}

	err := s.markProjectUpdatedAt(proj.ID)
	if err != nil {
		return err
	}

	return nil
}
