package storage

import (
	"errors"

	"github.com/google/uuid"

	"github.com/dapperlabs/flow-go/engine/execution/state"

	"github.com/dapperlabs/flow-playground-api/model"
)

type Store interface {
	InsertProject(proj *model.InternalProject) error
	UpdateProject(input model.UpdateProject, proj *model.InternalProject) error
	GetProject(id uuid.UUID, proj *model.InternalProject) error

	InsertAccount(acc *model.InternalAccount) error
	GetAccount(id uuid.UUID, acc *model.InternalAccount) error
	UpdateAccount(input model.UpdateAccount, acc *model.InternalAccount) error
	UpdateAccountState(accountID uuid.UUID, state map[string][]byte) error
	GetAccountsForProject(projectID uuid.UUID, accs *[]*model.InternalAccount) error
	DeleteAccount(id uuid.UUID) error

	InsertTransactionTemplate(tpl *model.TransactionTemplate) error
	UpdateTransactionTemplate(input model.UpdateTransactionTemplate, tpl *model.TransactionTemplate) error
	GetTransactionTemplate(id uuid.UUID, tpl *model.TransactionTemplate) error
	GetTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error
	DeleteTransactionTemplate(id uuid.UUID) error

	InsertTransactionExecution(exe *model.TransactionExecution, delta state.Delta) error
	GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error

	InsertScriptTemplate(tpl *model.ScriptTemplate) error
	UpdateScriptTemplate(input model.UpdateScriptTemplate, tpl *model.ScriptTemplate) error
	GetScriptTemplate(id uuid.UUID, tpl *model.ScriptTemplate) error
	GetScriptTemplatesForProject(projectID uuid.UUID, tpls *[]*model.ScriptTemplate) error
	DeleteScriptTemplate(id uuid.UUID) error

	InsertScriptExecution(exe *model.ScriptExecution) error
	GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error

	InsertRegisterDelta(projectID uuid.UUID, delta state.Delta) error
	GetRegisterDeltasForProject(projectID uuid.UUID, deltas *[]state.Delta) error
}

var ErrNotFound = errors.New("entity not found")
