package memory

import (
	"sort"
	"sync"

	"github.com/google/uuid"

	"github.com/dapperlabs/flow-go/engine/execution/state"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Store struct {
	mut                   sync.RWMutex
	projects              map[uuid.UUID]model.InternalProject
	accounts              map[uuid.UUID]model.InternalAccount
	transactionTemplates  map[uuid.UUID]model.TransactionTemplate
	transactionExecutions map[uuid.UUID]model.TransactionExecution
	scriptTemplates       map[uuid.UUID]model.ScriptTemplate
	scriptExecutions      map[uuid.UUID]model.ScriptExecution
	registerDeltas        map[uuid.UUID][]model.RegisterDelta
}

func NewStore() storage.Store {
	return &Store{
		mut:                   sync.RWMutex{},
		projects:              make(map[uuid.UUID]model.InternalProject),
		accounts:              make(map[uuid.UUID]model.InternalAccount),
		transactionTemplates:  make(map[uuid.UUID]model.TransactionTemplate),
		transactionExecutions: make(map[uuid.UUID]model.TransactionExecution),
		scriptTemplates:       make(map[uuid.UUID]model.ScriptTemplate),
		scriptExecutions:      make(map[uuid.UUID]model.ScriptExecution),
		registerDeltas:        make(map[uuid.UUID][]model.RegisterDelta),
	}
}

func (s *Store) CreateProject(
	proj *model.InternalProject,
	deltas []state.Delta,
	accounts []*model.InternalAccount,
	ttpls []*model.TransactionTemplate,
	stpls []*model.ScriptTemplate) error {
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

	if input.Persist != nil {
		p.Persist = *input.Persist
	}

	s.projects[input.ID] = p

	*proj = p

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

func (s *Store) InsertAccount(acc *model.InternalAccount) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	return s.insertAccount(acc)
}

func (s *Store) insertAccount(acc *model.InternalAccount) error {
	s.accounts[acc.ID] = *acc
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

	return nil
}

func (s *Store) UpdateAccountAfterDeployment(
	input model.UpdateAccount,
	states map[uuid.UUID]map[string][]byte,
	delta state.Delta,
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

func (s *Store) updateAccountState(id uuid.UUID, state map[string][]byte) error {
	account := s.accounts[id]
	account.State = state

	s.accounts[id] = account

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

	return nil
}

func (s *Store) InsertTransactionExecution(
	exe *model.TransactionExecution,
	states map[uuid.UUID]map[string][]byte,
	delta state.Delta,
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

func (s *Store) InsertRegisterDelta(projectID uuid.UUID, delta state.Delta, isAccountCreation bool) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	return s.insertRegisterDelta(projectID, delta, isAccountCreation)
}

func (s *Store) insertRegisterDelta(projectID uuid.UUID, delta state.Delta, isAccountCreation bool) error {
	p, ok := s.projects[projectID]
	if !ok {
		return storage.ErrNotFound
	}

	index := p.TransactionCount + 1

	regDelta := model.RegisterDelta{
		ProjectID:         projectID,
		Index:             index,
		Delta:             delta,
		IsAccountCreation: isAccountCreation,
	}

	s.registerDeltas[projectID] = append(s.registerDeltas[projectID], regDelta)

	p.TransactionCount = index

	s.projects[projectID] = p

	return nil
}

func (s *Store) GetRegisterDeltasForProject(projectID uuid.UUID, deltas *[]state.Delta) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	res := make([]state.Delta, 0)

	for _, delta := range s.registerDeltas[projectID] {
		res = append(res, delta.Delta)
	}

	*deltas = res

	return nil
}

func (s *Store) ClearProjectState(projectID uuid.UUID) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	for accountID, account := range s.accounts {
		if account.ProjectID != projectID {
			continue
		}

		account.DeployedCode = ""
		account.DeployedContracts = nil

		s.accounts[accountID] = account
	}

	currentRegisterDeltas := s.registerDeltas[projectID]
	newRegisterDeltas := make([]model.RegisterDelta, 0)

	for _, registerDelta := range currentRegisterDeltas {
		// only keep account deltas
		if !registerDelta.IsAccountCreation {
			continue
		}

		newRegisterDeltas = append(newRegisterDeltas, registerDelta)
	}

	s.registerDeltas[projectID] = newRegisterDeltas

	newRegisterDeltaCount := len(newRegisterDeltas)

	project := s.projects[projectID]
	project.TransactionCount = newRegisterDeltaCount
	s.projects[projectID] = project

	for txExecutionID, txExecution := range s.transactionExecutions {
		if txExecution.ProjectID != projectID {
			continue
		}

		delete(s.transactionExecutions, txExecutionID)
	}

	for scriptExecutionID, scriptExecution := range s.scriptExecutions {
		if scriptExecution.ProjectID != projectID {
			continue
		}

		delete(s.scriptExecutions, scriptExecutionID)
	}

	return nil
}
