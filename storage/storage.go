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
	UpdateTransactionTemplate(input model.UpdateTransactionTemplate, tpl *model.TransactionTemplate) error
	GetTransactionTemplate(id uuid.UUID, tpl *model.TransactionTemplate) error
	GetTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error
	DeleteTransactionTemplate(id uuid.UUID) error

	InsertScriptTemplate(tpl *model.ScriptTemplate) error
	UpdateScriptTemplate(input model.UpdateScriptTemplate, tpl *model.ScriptTemplate) error
	GetScriptTemplate(id uuid.UUID, tpl *model.ScriptTemplate) error
	GetScriptTemplatesForProject(projectID uuid.UUID, tpls *[]*model.ScriptTemplate) error
}

var ErrNotFound = errors.New("entity not found")
