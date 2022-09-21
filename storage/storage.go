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
	InsertAccounts(accs []*model.Account) error
	GetAccount(id, pID uuid.UUID, acc *model.Account) error
	GetAccountsForProject(projectID uuid.UUID, accs *[]*model.Account) error
	DeleteAccount(id, pID uuid.UUID) error
	UpdateAccount(input model.UpdateAccount, acc *model.Account) error

	InsertTransactionTemplate(tpl *model.TransactionTemplate) error
	UpdateTransactionTemplate(input model.UpdateTransactionTemplate, tpl *model.TransactionTemplate) error
	GetTransactionTemplate(id, pID uuid.UUID, tpl *model.TransactionTemplate) error
	GetTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error
	DeleteTransactionTemplate(id, pID uuid.UUID) error

	InsertTransactionExecution(exe *model.TransactionExecution) error
	GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error

	InsertScriptTemplate(tpl *model.ScriptTemplate) error
	UpdateScriptTemplate(input model.UpdateScriptTemplate, tpl *model.ScriptTemplate) error
	GetScriptTemplate(id, pID uuid.UUID, tpl *model.ScriptTemplate) error
	GetScriptTemplatesForProject(projectID uuid.UUID, tpls *[]*model.ScriptTemplate) error
	DeleteScriptTemplate(id, pID uuid.UUID) error

	InsertScriptExecution(exe *model.ScriptExecution) error
	GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error
}

var ErrNotFound = errors.New("entity not found")
