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

package migrate

// TODO: Remove old migrators and create migrator from stable playground to playground v2

import (
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type Migrator struct {
	store    storage.Store
	projects *controller.Projects
}

var V1_0_0 = semver.MustParse("v1.0.0")
var V2_0_0 = semver.MustParse("v2.0.0")

func NewMigrator(store storage.Store, projects *controller.Projects) *Migrator {
	return &Migrator{
		store:    store,
		projects: projects,
	}
}

// MigrateProject migrates a project from one API version to another.
//
// This function only support projects upgrades (from lower to higher version). Project
// downgrades are not yet supported.
//
// This function returns a boolean indicating whether or not the project was migrated.
func (m *Migrator) MigrateProject(id uuid.UUID, from, to *semver.Version) (bool, error) {
	from = sanitizeVersion(from)
	to = sanitizeVersion(to)

	if !from.LessThan(to) {
		// No migration work to do
		return false, nil
	}

	if from.LessThan(V1_0_0) {
		// Current version is too old to migrate to v2
		return false, errors.New("Current version " + from.String() + " cannot be migrated to " + to.String())
	}

	if from.LessThan(V2_0_0) {
		err := m.migrateToV2_0_0(id)
		if err != nil {
			return false, errors.Wrapf(err, "failed to migrate project from %s to %s", from.String(), to.String())
		}
	}

	return true, nil
}

// migrateToV2_0_0 migrates a project to the version v2.0.0
//
// Steps:
// - 1. todo
// - 2. todo
func (m *Migrator) migrateToV2_0_0(projectID uuid.UUID) error {
	// 1. reset project state
	// TODO: Need to use the old project model? And then create a new project model to store in db!
	createdAccounts, err := m.projects.Reset(&model.Project{ID: projectID})
	if err != nil {
		return errors.Wrap(err, "migration failed to reset project state")
	}

	// 2. TODO: Add back GetAccountsForProject in order to retrieve the contracts + number of accounts
	//    TODO: Need the v1.0.0 account model to do migration
	var oldAccounts []*v1_0_0Account
	err = v1_0_0GetAccountsForProject(projectID, &oldAccounts)
	if err != nil {
		return errors.Wrap(err, "migration failed to get accounts")
	}

	// 3. Create contract files from old account draft codes
	var contractFiles []*model.File

	for i, account := range oldAccounts {
		contractFiles = append(contractFiles, &model.File{
			ID:        uuid.New(),
			ProjectID: projectID,
			Title:     "", // TODO: do we need to get the title here? Probably not?
			Type:      model.ContractFile,
			Index:     i,
			Script:    account.DraftCode,
		})
	}

	// TODO: implement GetV1Project for the old model
	oldProject, err := m.projects.GetV1Project(projectID)
	if err != nil {
		return errors.Wrap(err, "migration failed to get project")
	}

	newProject = model.Project{
		ID:                        projectID,
		UserID:                    uuid.UUID{},
		Secret:                    uuid.UUID{},
		PublicID:                  uuid.UUID{},
		ParentID:                  oldProject.ParentID,
		Title:                     oldProject.Title,
		Description:               oldProject.Description,
		Readme:                    oldProject.Readme,
		Seed:                      oldProject.Seed,
		NumberOfAccounts:          len(oldAccounts),
		TransactionExecutionCount: 0, // TODO: just reset these?
		Persist:                   oldProject.Persist,
		CreatedAt:                 oldProject.CreatedAt,
		UpdatedAt:                 time.Now(),
		Version:                   V2_0_0,
		Mutable:                   false,
	}

	// 4. TODO: Convert transaction templates and script templates to files and add to DB
	//    TODO: Get old templates from database

	err = m.projects.UpdateVersion(projectID, V2_0_0)
	if err != nil {
		return errors.Wrap(err, "failed to update project version")
	}

	return nil
}

func sanitizeVersion(version *semver.Version) *semver.Version {
	if version == nil {
		return V0
	}

	return version
}
