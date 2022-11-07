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

// TODO: Create new migrator tests from stable playground v1 to playground v2
// TODO: Create a projects with accounts and templates and test migration

/*
import (
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/migrate"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
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

		assert.Equal(t, migrate.V0.String(), proj.Version.String())

		assertAllAccountsExist(t, c.scripts, proj)

		migrated, err := c.migrator.MigrateProject(proj.ID, proj.Version, migrate.V0_1_0)
		require.NoError(t, err)
		assert.True(t, migrated)

		proj, err = c.projects.Get(proj.ID)
		require.NoError(t, err)

		assert.Equal(t, migrate.V0_1_0.String(), proj.Version.String())

		assertAllAccountsExist(t, c.scripts, proj)
	})(t)
}

func TestMigrateV0_1_0ToV0_2_0(t *testing.T) {
	migrateTest(migrate.V0_1_0, func(t *testing.T, c migrateTestCase) {
		v0_2_0 := semver.MustParse("v0.2.0")

		proj, err := c.projects.Create(c.user, model.NewProject{})
		require.NoError(t, err)

		assert.Equal(t, migrate.V0_1_0.String(), proj.Version.String())

		migrated, err := c.migrator.MigrateProject(proj.ID, proj.Version, v0_2_0)
		require.NoError(t, err)
		assert.True(t, migrated)

		proj, err = c.projects.Get(proj.ID)
		require.NoError(t, err)

		assert.Equal(t, v0_2_0.String(), proj.Version.String())
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
		store := storage.NewInMemory()
		chain := blockchain.NewProjects(store, 5)
		scripts := controller.NewScripts(store, chain)
		projects := controller.NewProjects(startVersion, store, chain)

		migrator := migrate.NewMigrator(store, projects)

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

func assertAllAccountsExist(t *testing.T, scripts *controller.Scripts, proj *model.Project) {
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

func Test_MigrationV0_12_0(t *testing.T) {
	store := storage.NewInMemory()

	chain := blockchain.NewProjects(store, 5)
	projects := controller.NewProjects(semver.MustParse("v0.5.0"), store, chain)

	migrator := migrate.NewMigrator(store, projects)

	user := model.User{
		ID: uuid.New(),
	}

	err := store.InsertUser(&user)
	require.NoError(t, err)

	projID := uuid.New()
	err = store.CreateProject(&model.Project{
		ID:                        projID,
		UserID:                    user.ID,
		Secret:                    uuid.New(),
		PublicID:                  uuid.New(),
		ParentID:                  nil,
		Title:                     "e2eTest project",
		Description:               "e2eTest description",
		Readme:                    "",
		Seed:                      1,
		TransactionExecutionCount: 5,
		Persist:                   true,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
		Version:                   semver.MustParse("v0.10.0"),
	}, []*model.TransactionTemplate{{
		ID:        uuid.New(),
		ProjectID: projID,
		Title:     "tx template",
		Script: `
			import A from 0x01
			transaction {}
		`,
		Index: 0,
	}}, []*model.ScriptTemplate{{
		ID:        uuid.New(),
		ProjectID: projID,
		Title:     "script template",
		Script: `
			import Foo from 0x01
			transaction {}
		`,
	}})
	require.NoError(t, err)

	accTmpl := `
		import A%d from 0x0%d
		pub contract B {}`

	for i := 0; i < 5; i++ {
		err = store.InsertAccount(&model.Account{
			ID:        uuid.New(),
			ProjectID: projID,
			Address:   model.NewAddressFromString(fmt.Sprintf("0x0%d", i+1)),
			DraftCode: fmt.Sprintf(accTmpl, i, i+1),
			Index:     i,
		})
		require.NoError(t, err)
	}

	err = store.InsertTransactionExecution(&model.TransactionExecution{
		ID:        uuid.New(),
		ProjectID: projID,
		Index:     0,
		Script: `
			import Bar from 0x01
			transaction {}
		`,
		Arguments: []string{`{ "type": "Address", "value": "0x01" }`},
		Signers:   []model.Address{model.NewAddressFromString("0x01")},
		Errors:    nil,
		Events:    nil,
		Logs:      nil,
	})
	require.NoError(t, err)

	newVer := semver.MustParse("0.12.0")
	migrated, err := migrator.MigrateProject(projID, semver.MustParse("v0.10.0"), newVer)
	require.NoError(t, err)
	assert.True(t, migrated)

	var accs []*model.Account
	err = store.GetAccountsForProject(projID, &accs)
	require.NoError(t, err)
	assert.Len(t, accs, 5)
	for i, a := range accs {
		assert.Equal(t, model.NewAddressFromString(fmt.Sprintf("0x0%d", i+5)), a.Address) // assert address was shifted
		assert.Equal(t, fmt.Sprintf(accTmpl, i, i+5), a.DraftCode)                        // assert code script was shifted
	}

	var exes []*model.TransactionExecution
	err = store.GetTransactionExecutionsForProject(projID, &exes)
	require.NoError(t, err)
	assert.Len(t, exes, 5)
	for i, exe := range exes {
		assert.Equal(t, "flow.AccountCreated", exe.Events[5].Type)
		assert.Equal(t, i, exes[i].Index)
	}

	var scriptExes []*model.ScriptExecution
	err = store.GetScriptExecutionsForProject(projID, &scriptExes)
	require.NoError(t, err)
	assert.Len(t, scriptExes, 0)

	var project model.Project
	err = store.GetProject(projID, &project)
	require.NoError(t, err)

	assert.Equal(t, newVer, project.Version)
	assert.Equal(t, 5, project.TransactionExecutionCount)
}
*/
