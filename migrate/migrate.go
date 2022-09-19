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

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/adapter"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Migrator struct {
	store    storage.Store
	projects *controller.Projects
}

var V0 = semver.MustParse("v0.0.0")
var V0_1_0 = semver.MustParse("v0.1.0")
var V0_12_0 = semver.MustParse("v0.12.0")

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
		return false, nil
	}

	if from.LessThan(V0_1_0) {
		err := m.migrateToV0_1_0(id)
		if err != nil {
			return false, errors.Wrapf(err, "failed to migrate project from %s to %s", V0, V0_1_0)
		}
	}
	if from.LessThan(V0_12_0) {
		err := m.migrateToV0_12_0(id)
		if err != nil {
			return false, errors.Wrapf(err, "failed to migrate project from %s to %s", V0, V0_12_0)
		}
	}

	// If no migration steps are left, set project version to latest.
	err := m.projects.UpdateVersion(id, to)
	if err != nil {
		return false, errors.Wrap(err, "failed to update project version")
	}

	return true, nil
}

// migrateToV0_1_0 migrates a project from v0.0.0 to v0.1.0.
//
// Steps:
// - 1. Reset project state and recreate initial accounts
// - 2. Update project version tag
func (m *Migrator) migrateToV0_1_0(id uuid.UUID) error {
	proj := model.InternalProject{
		ID: id,
	}

	// TODO:
	//  Update storage interface to allow atomic transactions.
	//  Ideally the project state should be wiped and the version incremented in the
	//  same transaction.

	// Step 1/2
	_, err := m.projects.Reset(&proj)
	if err != nil {
		return errors.Wrap(err, "failed to reset project state")
	}

	// Step 2/2
	err = m.projects.UpdateVersion(id, V0_1_0)
	if err != nil {
		return errors.Wrap(err, "failed to update project version")
	}

	return nil
}

// migrateToV0_12_0 migrates a project to the version v0.12.0
//
// Steps:
// - 1. Reset project state recreate initial accounts
// - 2. Get all accounts for project and update with shifted addresses and removed unused fields
func (m *Migrator) migrateToV0_12_0(projectID uuid.UUID) error {
	// 1. reset project state
	createdAccounts, err := m.projects.Reset(&model.InternalProject{
		ID: projectID,
	})
	if err != nil {
		return errors.Wrap(err, "migration failed to reset project state")
	}

	var accounts []*model.InternalAccount
	err = m.store.GetAccountsForProject(projectID, &accounts)
	if err != nil {
		return errors.Wrap(err, "migration failed to get accounts")
	}

	if len(accounts) != len(createdAccounts) {
		return fmt.Errorf("migration failture, created accounts length doesn't match existing accounts")
	}

	// 2. migrate accounts
	for i, acc := range accounts {
		acc.Address = createdAccounts[i].Address
		acc.DraftCode = adapter.ContentAddressFromAPI(acc.DraftCode)

		err = m.store.DeleteAccount(acc.ProjectChildID)
		if err != nil {
			return errors.Wrap(err, "migration failed to migrate accounts")
		}

		err = m.store.InsertAccount(acc)
		if err != nil {
			return errors.Wrap(err, "migration failed to migrate accounts")
		}
	}

	err = m.projects.UpdateVersion(projectID, V0_12_0)
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
