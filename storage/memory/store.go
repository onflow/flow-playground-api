package memory

import (
	"sort"
	"sync"

	"github.com/dapperlabs/flow-go/engine/execution/execution/state"
	"github.com/google/uuid"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Store struct {
	mut                   sync.RWMutex
	projects              map[uuid.UUID]model.Project
	accounts              map[uuid.UUID]model.Account
	transactionTemplates  map[uuid.UUID]model.TransactionTemplate
	transactionExecutions map[uuid.UUID]model.TransactionExecution
	registerDeltas        []model.RegisterDelta
}

func NewStore() *Store {
	return &Store{
		mut:                   sync.RWMutex{},
		projects:              make(map[uuid.UUID]model.Project),
		accounts:              make(map[uuid.UUID]model.Account),
		transactionTemplates:  make(map[uuid.UUID]model.TransactionTemplate),
		transactionExecutions: make(map[uuid.UUID]model.TransactionExecution),
		registerDeltas:        make([]model.RegisterDelta, 0),
	}
}

func (s *Store) InsertProject(proj *model.Project) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.projects[proj.ID] = *proj

	return nil
}

func (s *Store) GetProject(id uuid.UUID, proj *model.Project) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	p, ok := s.projects[id]
	if !ok {
		return storage.ErrNotFound
	}

	*proj = p

	return nil
}

func (s *Store) InsertAccount(acc *model.Account) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.accounts[acc.ID] = *acc

	return nil
}

func (s *Store) GetAccount(id uuid.UUID, acc *model.Account) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	p, ok := s.accounts[id]
	if !ok {
		return storage.ErrNotFound
	}

	*acc = p

	return nil
}

func (s *Store) UpdateAccount(input model.UpdateAccount, acc *model.Account) error {
	s.mut.Lock()
	defer s.mut.Unlock()

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

	s.accounts[input.ID] = a

	*acc = a

	return nil
}

func (s *Store) GetAccountsForProject(projectID uuid.UUID, accs *[]*model.Account) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.getAccountsForProject(projectID, accs)
}

func (s *Store) getAccountsForProject(projectID uuid.UUID, accs *[]*model.Account) error {
	res := make([]*model.Account, 0)

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

func (s *Store) DeleteAccount(id uuid.UUID) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	_, ok := s.accounts[id]
	if !ok {
		return storage.ErrNotFound
	}

	delete(s.accounts, id)

	return nil
}

func (s *Store) InsertTransactionTemplate(tpl *model.TransactionTemplate) error {
	s.mut.Lock()
	defer s.mut.Unlock()

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

func (s *Store) GetTransactionTemplate(id uuid.UUID, tpl *model.TransactionTemplate) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	t, ok := s.transactionTemplates[id]
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

func (s *Store) DeleteTransactionTemplate(id uuid.UUID) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	_, ok := s.transactionTemplates[id]
	if !ok {
		return storage.ErrNotFound
	}

	delete(s.transactionTemplates, id)

	return nil
}

func (s *Store) InsertTransactionExecution(exe *model.TransactionExecution, delta state.Delta) error {
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

	err = s.insertRegisterDelta(exe.ProjectID, delta)
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

func (s *Store) InsertRegisterDelta(projectID uuid.UUID, delta state.Delta) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	return s.insertRegisterDelta(projectID, delta)
}

func (s *Store) insertRegisterDelta(projectID uuid.UUID, delta state.Delta) error {
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

	s.registerDeltas = append(s.registerDeltas, regDelta)

	p.TransactionCount = index

	s.projects[projectID] = p

	return nil
}

func (s *Store) GetRegisterDeltasForProject(projectID uuid.UUID, deltas *[]state.Delta) error {
	s.mut.RLock()
	defer s.mut.RUnlock()

	res := make([]state.Delta, 0)

	for _, delta := range s.registerDeltas {
		if delta.ProjectID == projectID {
			res = append(res, delta.Delta)
		}
	}
	*deltas = res

	return nil
}
