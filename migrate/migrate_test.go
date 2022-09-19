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

package migrate_test

import (
	"fmt"
	"github.com/dapperlabs/flow-playground-api/server/storage"
	"github.com/dapperlabs/flow-playground-api/server/storage/memory"
	"testing"

	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/golang/groupcache/lru"

	"github.com/Masterminds/semver"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/migrate"
	"github.com/dapperlabs/flow-playground-api/model"
)

const numAccounts = 4

func TestMigrateNilToV0(t *testing.T) {
	migrateTest(migrate.V0, func(t *testing.T, c migrateTestCase) {
		projID := uuid.New()

		migrated, err := c.migrator.MigrateProject(projID, nil, migrate.V0)
		require.NoError(t, err)
		assert.False(t, migrated)
	})(t)
}

func TestMigrateV0ToV0(t *testing.T) {
	migrateTest(migrate.V0, func(t *testing.T, c migrateTestCase) {
		projID := uuid.New()

		migrated, err := c.migrator.MigrateProject(projID, migrate.V0, migrate.V0)
		require.NoError(t, err)
		assert.False(t, migrated)
	})(t)
}

func TestMigrateV0ToV0_1_0(t *testing.T) {
	migrateTest(migrate.V0, func(t *testing.T, c migrateTestCase) {
		proj, err := c.projects.Create(c.user, model.NewProject{})
		require.NoError(t, err)

		assert.Equal(t, migrate.V0, proj.Version)

		assertAllAccountsExist(t, c.scripts, proj)

		migrated, err := c.migrator.MigrateProject(proj.ID, proj.Version, migrate.V0_1_0)
		require.NoError(t, err)
		assert.True(t, migrated)

		proj, err = c.projects.Get(proj.ID)
		require.NoError(t, err)

		assert.Equal(t, migrate.V0_1_0, proj.Version)

		assertAllAccountsExist(t, c.scripts, proj)
	})(t)
}

func TestMigrateV0_1_0ToV0_2_0(t *testing.T) {
	migrateTest(migrate.V0_1_0, func(t *testing.T, c migrateTestCase) {
		v0_2_0 := semver.MustParse("v0.2.0")

		proj, err := c.projects.Create(c.user, model.NewProject{})
		require.NoError(t, err)

		assert.Equal(t, migrate.V0_1_0, proj.Version)

		migrated, err := c.migrator.MigrateProject(proj.ID, proj.Version, v0_2_0)
		require.NoError(t, err)
		assert.True(t, migrated)

		proj, err = c.projects.Get(proj.ID)
		require.NoError(t, err)

		assert.Equal(t, v0_2_0, proj.Version)
	})(t)
}

type migrateTestCase struct {
	store      storage.Store
	blockchain *blockchain.Projects
	scripts    *controller.Scripts
	projects   *controller.Projects
	migrator   *migrate.Migrator
	user       *model.User
}

func migrateTest(startVersion *semver.Version, f func(t *testing.T, c migrateTestCase)) func(t *testing.T) {
	return func(t *testing.T) {
		store := memory.NewStore()
		chain := blockchain.NewProjects(store, lru.New(128), 5)
		scripts := controller.NewScripts(store, chain)
		projects := controller.NewProjects(startVersion, store, chain)

		migrator := migrate.NewMigrator(projects)

		user := model.User{
			ID: uuid.New(),
		}

		err := store.InsertUser(&user)
		require.NoError(t, err)

		f(t, migrateTestCase{
			store:      store,
			blockchain: chain,
			scripts:    scripts,
			projects:   projects,
			migrator:   migrator,
			user:       &user,
		})
	}
}

func assertAllAccountsExist(t *testing.T, scripts *controller.Scripts, proj *model.InternalProject) {
	for i := 1; i <= numAccounts; i++ {
		script := fmt.Sprintf(`pub fun main() { getAccount(0x%x) }`, i)

		result, err := scripts.CreateExecution(model.NewScriptExecution{
			ProjectID: proj.ID,
			Script:    script,
			Arguments: nil,
		})
		require.NoError(t, err)

		assert.Empty(t, result.Errors)
	}
}
