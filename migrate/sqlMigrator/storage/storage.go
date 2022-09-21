package storage

import (
	"errors"

	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
)

type Store interface {
	InsertUser(user *model.User) error
	GetUser(id uuid.UUID, user *model.User) error

	CreateProject(
		proj *model.Project,
		ttpl []*model.TransactionTemplate,
		stpl []*model.ScriptTemplate,
	) error
	UpdateProject(input model.UpdateProject, proj *model.Project) error
	UpdateProjectOwner(id, userID uuid.UUID) error
	UpdateProjectVersion(id uuid.UUID, version *semver.Version) error
	ResetProjectState(proj *model.Project) error
	GetProject(id uuid.UUID, proj *model.Project) error

	InsertAccount(acc *model.Account) error
	GetAccount(id model.ProjectChildID, acc *model.Account) error
	GetAccountsForProject(projectID uuid.UUID, accs *[]*model.Account) error
	DeleteAccount(id model.ProjectChildID) error
	UpdateAccount(input model.UpdateAccount, acc *model.Account) error

	InsertTransactionTemplate(tpl *model.TransactionTemplate) error
	UpdateTransactionTemplate(input model.UpdateTransactionTemplate, tpl *model.TransactionTemplate) error
	GetTransactionTemplate(id model.ProjectChildID, tpl *model.TransactionTemplate) error
	GetTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error
	DeleteTransactionTemplate(id model.ProjectChildID) error

	InsertTransactionExecution(exe *model.TransactionExecution) error
	GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error

	InsertScriptTemplate(tpl *model.ScriptTemplate) error
	UpdateScriptTemplate(input model.UpdateScriptTemplate, tpl *model.ScriptTemplate) error
	GetScriptTemplate(id model.ProjectChildID, tpl *model.ScriptTemplate) error
	GetScriptTemplatesForProject(projectID uuid.UUID, tpls *[]*model.ScriptTemplate) error
	DeleteScriptTemplate(id model.ProjectChildID) error

	InsertScriptExecution(exe *model.ScriptExecution) error
	GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error
}

var ErrNotFound = errors.New("entity not found")
