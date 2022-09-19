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

	"github.com/dapperlabs/flow-playground-api/storage/datastore"

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
	fmt.Println("migrate v0.12")
	if from.LessThan(V0_12_0) {
		fmt.Println("migrating v0.12")
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
	err := m.projects.Reset(&proj)
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
// - 3. Get all transaction executions and update with shifted addresses in script, arguments and signers
// - 4. Get all script executions and update with shifted addresses in script and arguments
func (m *Migrator) migrateToV0_12_0(projectID uuid.UUID) error {
	fmt.Println("MIGRATION start")

	store, ok := m.store.(*datastore.Datastore)
	if !ok {
		return nil // only migrate datastore
	}

	var project model.InternalProject
	err := store.GetProject(projectID, &project)
	if err != nil {
		return errors.Wrap(err, "migration failed to get project")
	}

	// update to migrated version
	project.Version = V0_12_0

	// 1. reset project state
	err = m.projects.Reset(&project)
	if err != nil {
		return errors.Wrap(err, "migration failed to reset project state")
	}

	var accounts []*model.InternalAccount
	err = m.store.GetAccountsForProject(projectID, &accounts)
	if err != nil {
		return errors.Wrap(err, "migration failed to get accounts")
	}

	// 2. migrate accounts
	for i, acc := range accounts {
		accounts[i].Address = adapter.AddressFromAPI(acc.Address)
		accounts[i].DraftCode = adapter.ContentAddressFromAPI(acc.DraftCode)
	}

	var exes []*model.TransactionExecution
	err = m.store.GetTransactionExecutionsForProject(projectID, &exes)
	if err != nil {
		return errors.Wrap(err, "migration failed to get executions")
	}

	// 3. migrate transaction executions
	for i, exe := range exes {
		exes[i].Script = adapter.ContentAddressFromAPI(exe.Script)
		for j, sig := range exe.Signers {
			exes[i].Signers[j] = adapter.AddressFromAPI(sig)
		}
		for j, arg := range exe.Arguments {
			exes[i].Arguments[j] = adapter.ContentAddressFromAPI(arg)
		}
	}

	var scripts []*model.ScriptExecution
	err = m.store.GetScriptExecutionsForProject(projectID, &scripts)
	if err != nil {
		return errors.Wrap(err, "migration failed to get scripts")
	}

	// 4. migrate scripts
	for i, s := range scripts {
		scripts[i].Script = adapter.ContentAddressFromAPI(s.Script)
		for j, arg := range s.Arguments {
			scripts[i].Arguments[j] = adapter.ContentAddressFromAPI(arg)
		}
	}

	return store.MigrateToV0_12_0(project, accounts, exes, scripts)
}

func sanitizeVersion(version *semver.Version) *semver.Version {
	if version == nil {
		return V0
	}

	return version
}
