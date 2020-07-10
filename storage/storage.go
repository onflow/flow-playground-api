package storage

import (
	"errors"

	"github.com/dapperlabs/flow-go/engine/execution/state/delta"
	"github.com/google/uuid"

	"github.com/dapperlabs/flow-playground-api/model"
)

type Store interface {
	InsertUser(user *model.User) error
	GetUserBySessionID(sessionID string, user *model.User) error

	CreateProject(
		proj *model.InternalProject,
		registerDeltas []delta.Delta,
		accounts []*model.InternalAccount,
		ttpl []*model.TransactionTemplate,
		stpl []*model.ScriptTemplate,
	) error
	UpdateProject(input model.UpdateProject, proj *model.InternalProject) error
	UpdateProjectOwner(id, userID uuid.UUID) error
	GetProject(id uuid.UUID, proj *model.InternalProject) error

	InsertAccount(acc *model.InternalAccount) error
	GetAccount(id model.ProjectChildID, acc *model.InternalAccount) error
	UpdateAccount(input model.UpdateAccount, acc *model.InternalAccount) error
	UpdateAccountAfterDeployment(
		input model.UpdateAccount,
		states map[uuid.UUID]map[string][]byte,
		delta delta.Delta,
		acc *model.InternalAccount,
	) error
	GetAccountsForProject(projectID uuid.UUID, accs *[]*model.InternalAccount) error
	DeleteAccount(id model.ProjectChildID) error

	InsertTransactionTemplate(tpl *model.TransactionTemplate) error
	UpdateTransactionTemplate(input model.UpdateTransactionTemplate, tpl *model.TransactionTemplate) error
	GetTransactionTemplate(id model.ProjectChildID, tpl *model.TransactionTemplate) error
	GetTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error
	DeleteTransactionTemplate(id model.ProjectChildID) error

	InsertTransactionExecution(
		exe *model.TransactionExecution,
		states map[uuid.UUID]map[string][]byte,
		delta delta.Delta,
	) error
	GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error

	InsertScriptTemplate(tpl *model.ScriptTemplate) error
	UpdateScriptTemplate(input model.UpdateScriptTemplate, tpl *model.ScriptTemplate) error
	GetScriptTemplate(id model.ProjectChildID, tpl *model.ScriptTemplate) error
	GetScriptTemplatesForProject(projectID uuid.UUID, tpls *[]*model.ScriptTemplate) error
	DeleteScriptTemplate(id model.ProjectChildID) error

	InsertScriptExecution(exe *model.ScriptExecution) error
	GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error

	InsertRegisterDelta(projectID uuid.UUID, delta delta.Delta, isAccountCreation bool) error
	GetRegisterDeltasForProject(projectID uuid.UUID, deltas *[]*model.RegisterDelta) error
	ClearProjectState(projectID uuid.UUID) (int, error)
}

var ErrNotFound = errors.New("entity not found")
