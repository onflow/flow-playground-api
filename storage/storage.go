package storage

import (
	"errors"

	"github.com/google/uuid"

	"github.com/dapperlabs/flow-playground-api/model"
)

type Store interface {
	InsertProject(proj *model.Project) error
	GetProject(id uuid.UUID, proj *model.Project) error
	InsertTransactionTemplate(tpl *model.TransactionTemplate) error
	UpdateTransactionTemplate(tpl *model.TransactionTemplate) error
	GetTransactionTemplate(id uuid.UUID, tpl *model.TransactionTemplate) error
	GetTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error
}

var ErrNotFound = errors.New("entity not found")
