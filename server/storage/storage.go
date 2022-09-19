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
	"context"
	"github.com/dapperlabs/flow-playground-api/server/config"
	"github.com/dapperlabs/flow-playground-api/server/model"
	"github.com/dapperlabs/flow-playground-api/server/storage/datastore"
	"github.com/dapperlabs/flow-playground-api/server/storage/memory"
	"github.com/kelseyhightower/envconfig"
	"log"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/uuid"
)

// Global storage
var store Store = nil

// GetStorage returns global storage based on global configuration
func GetStorage() Store {
	if store == nil {
		if strings.EqualFold(config.GetConfig().StorageBackend, "datastore") {
			var datastoreConf DatastoreConfig

			if err := envconfig.Process("FLOW_DATASTORE", &datastoreConf); err != nil {
				log.Fatal(err)
			}

			var err error
			store, err = datastore.NewDatastore(
				context.Background(),
				&datastore.Config{
					DatastoreProjectID: datastoreConf.GCPProjectID,
					DatastoreTimeout:   datastoreConf.Timeout,
				},
			)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			store = memory.NewStore()
		}
	}
	return store
}

// SetStorage sets global storage to a specified storage
func SetStorage(newStore Store) {
	store = newStore
}

type DatastoreConfig struct {
	GCPProjectID string        `required:"true"`
	Timeout      time.Duration `default:"5s"`
}

type Store interface {
	InsertUser(user *model.User) error
	GetUser(id uuid.UUID, user *model.User) error

	CreateProject(
		proj *model.InternalProject,
		ttpl []*model.TransactionTemplate,
		stpl []*model.ScriptTemplate,
	) error
	UpdateProject(input model.UpdateProject, proj *model.InternalProject) error
	UpdateProjectOwner(id, userID uuid.UUID) error
	UpdateProjectVersion(id uuid.UUID, version *semver.Version) error
	ResetProjectState(proj *model.InternalProject) error
	GetProject(id uuid.UUID, proj *model.InternalProject) error

	InsertAccount(acc *model.InternalAccount) error
	GetAccount(id model.ProjectChildID, acc *model.InternalAccount) error
	GetAccountsForProject(projectID uuid.UUID, accs *[]*model.InternalAccount) error
	DeleteAccount(id model.ProjectChildID) error
	UpdateAccount(input model.UpdateAccount, acc *model.InternalAccount) error

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
