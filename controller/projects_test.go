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

package controller

import (
	"github.com/dapperlabs/flow-playground-api/server/config"
	"github.com/kelseyhightower/envconfig"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createStore() storage.Store {
	var store storage.Store

	if strings.EqualFold(os.Getenv("FLOW_STORAGEBACKEND"), storage.PostgreSQL) {
		var datastoreConf config.DatabaseConfig
		if err := envconfig.Process("FLOW_DB", &datastoreConf); err != nil {
			panic(err)
		}

		store = storage.NewPostgreSQL(&datastoreConf)
	} else {
		store = storage.NewInMemory()
	}

	return store
}

func createUser(store storage.Store) *model.User {
	user := &model.User{ID: uuid.New()}

	err := store.InsertUser(user)
	if err != nil {
		panic(err)
	}
	return user
}

func createProjects() (*Projects, storage.Store, *model.User) {
	store := createStore()
	user := createUser(store)
	chain := blockchain.NewProjects(store, 5)
	return NewProjects(version, store, chain), store, user
}

func createControllers() (storage.Store, *model.User, *blockchain.Projects, *Projects, *Files, *Accounts) {
	store := createStore()
	user := createUser(store)
	chain := blockchain.NewProjects(store, 5)
	projects := NewProjects(version, store, chain)
	files := NewFiles(store, chain)
	accounts := NewAccounts(store, chain)

	return store, user, chain, projects, files, accounts
}

const seedTitle = "e2eTest title"
const seedDesc = "e2eTest desc"
const seedReadme = "e2eTest readme"

func seedProject(projects *Projects, user *model.User) (*model.Project, error) {
	contract := model.NewProjectContractTemplate{
		Title:  "contract template 1",
		Script: "a",
	}

	project, err := projects.Create(user, model.NewProject{
		Title:                seedTitle,
		Description:          seedDesc,
		Readme:               seedReadme,
		Seed:                 1,
		ContractTemplates:    []*model.NewProjectContractTemplate{&contract},
		TransactionTemplates: nil,
		ScriptTemplates:      nil,
	})
	return project, err
}

func Test_CreateProject(t *testing.T) {
	projects, store, user := createProjects()

	t.Run("successful creation", func(t *testing.T) {
		project, err := seedProject(projects, user)
		require.NoError(t, err)

		assert.Equal(t, seedTitle, project.Title)
		assert.Equal(t, seedDesc, project.Description)
		assert.Equal(t, seedReadme, project.Readme)
		assert.Equal(t, 1, project.Seed)
		assert.False(t, project.Persist)
		assert.Equal(t, user.ID, project.UserID)

		var dbProj model.Project
		err = store.GetProject(project.ID, &dbProj)
		require.NoError(t, err)

		assert.Equal(t, project.Title, dbProj.Title)
		assert.Equal(t, project.Description, dbProj.Description)
		assert.Equal(t, 0, dbProj.TransactionExecutionCount)
	})

	t.Run("successful update", func(t *testing.T) {
		projects, store, user := createProjects()
		proj, err := seedProject(projects, user)
		require.NoError(t, err)

		title := "update title"
		desc := "update desc"
		readme := "readme"
		persist := true

		updated, err := projects.Update(model.UpdateProject{
			ID:          proj.ID,
			Title:       &title,
			Description: &desc,
			Readme:      &readme,
			Persist:     &persist,
		})
		require.NoError(t, err)
		assert.Equal(t, desc, updated.Description)
		assert.Equal(t, title, updated.Title)
		assert.Equal(t, readme, updated.Readme)
		assert.Equal(t, persist, updated.Persist)

		var dbProj model.Project
		err = store.GetProject(proj.ID, &dbProj)
		require.NoError(t, err)
		assert.Equal(t, dbProj.ID, updated.ID)
		assert.Equal(t, dbProj.Description, updated.Description)
		assert.Equal(t, dbProj.Persist, updated.Persist)
	})

	t.Run("reset state", func(t *testing.T) {
		projects, store, user := createProjects()
		proj, err := seedProject(projects, user)
		require.NoError(t, err)

		err = store.InsertTransactionExecution(&model.TransactionExecution{
			File: model.File{
				ID:        uuid.New(),
				ProjectID: proj.ID,
				Index:     6,
				Script:    "e2eTest",
			},
		})
		require.NoError(t, err)

		err = projects.Reset(proj.ID)
		assert.NoError(t, err)

		// TODO: Get accounts
		//require.Equal(t, len(accounts), 5) // Initial accounts

		var dbProj model.Project
		err = store.GetProject(proj.ID, &dbProj)
		require.NoError(t, err)

		assert.Equal(t, 0, dbProj.TransactionExecutionCount)
	})
}

func Test_AccessedTime(t *testing.T) {
	t.Run("update accessed time", func(t *testing.T) {
		projects, store, user := createProjects()

		project, err := seedProject(projects, user)
		require.NoError(t, err)
		require.NotEmpty(t, project.AccessedAt)

		var dbProj model.Project
		err = store.GetProject(project.ID, &dbProj)
		require.NoError(t, err)

		require.Equal(t, project.AccessedAt.UnixMilli(), dbProj.AccessedAt.UnixMilli())

		time.Sleep(2 * time.Second)

		getProj, err := projects.Get(project.ID)
		require.NoError(t, err)

		require.True(t, project.AccessedAt.Before(getProj.AccessedAt))
	})
}

func Test_StaleProjects(t *testing.T) {
	t.Run("get stale projects", func(t *testing.T) {
		projects, store, user := createProjects()

		for i := 0; i < 5; i++ {
			project, err := seedProject(projects, user)
			require.NoError(t, err)
			require.NotEmpty(t, project.AccessedAt)
		}

		stale := time.Second * 1

		// Make all projects stale
		time.Sleep(time.Second * 2)

		var staleProjects []*model.Project
		err := store.GetStaleProjects(stale, &staleProjects)
		require.NoError(t, err)
		require.Equal(t, 5, len(staleProjects))
	})

	t.Run("no stale projects", func(t *testing.T) {
		projects, store, user := createProjects()

		for i := 0; i < 5; i++ {
			project, err := seedProject(projects, user)
			require.NoError(t, err)
			require.NotEmpty(t, project.AccessedAt)
		}

		stale := time.Hour * 1

		time.Sleep(time.Second * 2)

		var staleProjects []*model.Project
		err := store.GetStaleProjects(stale, &staleProjects)
		require.NoError(t, err)
		require.Empty(t, staleProjects)
	})

	t.Run("delete stale projects", func(t *testing.T) {
		projects, store, user := createProjects()

		for i := 0; i < 5; i++ {
			project, err := seedProject(projects, user)
			require.NoError(t, err)
			require.NotEmpty(t, project.AccessedAt)

			err = store.InsertScriptExecution(&model.ScriptExecution{
				File: model.File{
					ID:        uuid.New(),
					ProjectID: project.ID,
					Title:     "title",
					Type:      0,
					Index:     0,
					Script:    "script",
				},
			})
			require.NoError(t, err)
		}

		// This execution should not be deleted
		newProjID := uuid.New()
		err := store.InsertScriptExecution(&model.ScriptExecution{
			File: model.File{
				ID:        uuid.New(),
				ProjectID: newProjID,
				Title:     "title",
				Type:      0,
				Index:     0,
				Script:    "script",
			},
		})
		require.NoError(t, err)

		stale := time.Second * 1

		// Make all projects stale
		time.Sleep(time.Second * 2)

		var staleProjects []*model.Project
		err = store.GetStaleProjects(stale, &staleProjects)
		require.NoError(t, err)
		require.NotEmpty(t, staleProjects)

		var exes []*model.ScriptExecution
		err = store.GetScriptExecutionsForProject(staleProjects[0].ID, &exes)
		require.NoError(t, err)
		require.NotEmpty(t, exes)

		testProjID := staleProjects[0].ID

		err = store.DeleteStaleProjects(stale)
		require.NoError(t, err)

		err = store.GetStaleProjects(stale, &staleProjects)
		require.NoError(t, err)
		require.Empty(t, staleProjects)

		err = store.GetScriptExecutionsForProject(testProjID, &exes)
		require.NoError(t, err)
		require.Empty(t, exes)

		err = store.GetScriptExecutionsForProject(newProjID, &exes)
		require.NoError(t, err)
		require.NotEmpty(t, exes)
	})
}

func Test_StateRecreation(t *testing.T) {
	_, user, _, projects, files, accounts := createControllers()

	contract1 := `pub contract HelloWorld { 
		init() {
			log("hello")
		} 
	}`

	tx1 := `transaction {
		prepare(auth: AuthAccount) {}
		execute {
			log("hello tx")		
		}
	}`

	script1 := `pub fun main(): Int {
		return 42;
	}`

	txTpls := []*model.NewProjectTransactionTemplate{{
		Title:  "tx template 1",
		Script: tx1,
	}}

	scTpls := []*model.NewProjectScriptTemplate{{
		Title:  "script template 1",
		Script: script1,
	}}

	ctTpls := []*model.NewProjectContractTemplate{
		{
			Title:  "contract template 1",
			Script: contract1,
		},
		{
			Title:  "contract template 2",
			Script: contract1,
		},
		{
			Title:  "contract template 3",
			Script: contract1,
		},
	}

	newProject := model.NewProject{
		ParentID:             nil,
		Title:                "Test Title",
		Description:          "Test Desc",
		Readme:               "Test Readme",
		Seed:                 1,
		NumberOfAccounts:     5,
		ContractTemplates:    ctTpls,
		TransactionTemplates: txTpls,
		ScriptTemplates:      scTpls,
	}

	p, err := projects.Create(user, newProject)
	require.NoError(t, err)

	newProj, err := projects.Get(p.ID)
	require.NoError(t, err)

	// Deploy a contract
	contractFiles, err := files.GetFilesForProject(newProj.ID, model.ContractFile)
	require.NoError(t, err)

	deploy := model.NewContractDeployment{
		ProjectID: contractFiles[0].ProjectID,
		Script:    contractFiles[0].Script,
		Address:   model.Address(flow.HexToAddress("0x01")),
	}

	contractDeployment, err := files.DeployContract(deploy)
	require.NoError(t, err)

	assert.Equal(t, "HelloWorld", contractDeployment.File.Title)

	// Re-deploy contract
	_, err = files.DeployContract(deploy)
	require.NoError(t, err)

	acc, err := accounts.GetByAddress(model.Address(flow.HexToAddress("0x01")), newProj.ID)
	require.NoError(t, err)

	assert.Contains(t, acc.DeployedContracts, "HelloWorld")
}
