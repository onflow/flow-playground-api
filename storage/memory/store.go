package memory

import (
	"sync"

	"github.com/google/uuid"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Store struct {
	mut                  sync.RWMutex
	projects             map[uuid.UUID]model.Project
	transactionTemplates map[uuid.UUID]model.TransactionTemplate
}

func NewStore() *Store {
	return &Store{
		mut:                  sync.RWMutex{},
		projects:             make(map[uuid.UUID]model.Project),
		transactionTemplates: make(map[uuid.UUID]model.TransactionTemplate),
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

	t, ok := s.projects[id]
	if !ok {
		return storage.ErrNotFound
	}

	*proj = t

	return nil
}

func (s *Store) InsertTransactionTemplate(tpl *model.TransactionTemplate) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	count := len(s.transactionTemplates)

	// set index to one after last
	tpl.Index = count

	s.transactionTemplates[tpl.ID] = *tpl

	return nil
}

func (s *Store) UpdateTransactionTemplate(tpl *model.TransactionTemplate) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	_, ok := s.transactionTemplates[tpl.ID]
	if !ok {
		return storage.ErrNotFound
	}

	s.transactionTemplates[tpl.ID] = *tpl

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
	res := make([]*model.TransactionTemplate, 0)

	for _, tpl := range s.transactionTemplates {
		if tpl.ProjectID == projectID {
			t := tpl
			res = append(res, &t)
		}
	}

	*tpls = res

	return nil
}
