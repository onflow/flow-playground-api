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
	"github.com/kelseyhightower/envconfig"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"

	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createStore() storage.Store {
	var store storage.Store

	if strings.EqualFold(os.Getenv("FLOW_STORAGEBACKEND"), storage.PostgreSQL) {
		var datastoreConf storage.DatabaseConfig
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

func createControllers() (storage.Store, *model.User, *blockchain.Projects, *Projects, *Files) {
	store := createStore()
	user := createUser(store)
	chain := blockchain.NewProjects(store, 5)
	projects := NewProjects(version, store, chain)
	files := NewFiles(store, chain)

	return store, user, chain, projects, files
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
		assert.Equal(t, 5, dbProj.TransactionExecutionCount)
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

		accounts, err := projects.Reset(proj)
		assert.NoError(t, err)
		require.Equal(t, len(accounts), 5) // Initial accounts

		var dbProj model.Project
		err = store.GetProject(proj.ID, &dbProj)
		require.NoError(t, err)

		assert.Equal(t, 5, dbProj.TransactionExecutionCount)
	})
}

func Test_StateRecreation(t *testing.T) {
	_, user, _, projects, files := createControllers()

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
	println("CONTRACT TITLE: ", contractDeployment.Title)
	require.NoError(t, err)

	assert.Equal(t, "HelloWorld", contractDeployment.File.Title)

	// Re-deploy contract
	contractDeployment, err = files.DeployContract(deploy)
	require.NoError(t, err)

	_ = contractDeployment // TODO: add verifications

	/*
		// check what deployed on accounts
		allAccs, err := accounts.AllForProjectID(newProj.ID)
		require.NoError(t, err)
		for i, rAcc := range allAccs {
			assert.Equal(t, // asserting that account addresses are ordered
				flow.HexToAddress(fmt.Sprintf("0x0%d", i+5)).String(),
				rAcc.Address.ToFlowAddress().String(),
			)
			if rAcc.ID == redeployAcc.ID {
				// only one redeploy account has deployed code due to clear state
				assert.Equal(t, contract1, rAcc.DeployedCode)
			} else {
				assert.Equal(t, "", rAcc.DeployedCode)
			}
		}

		tx2 := `import HelloWorld from 0x05
			transaction {
				prepare(auth: AuthAccount) {}
				execute {}
			}`

		for i := 0; i < 5; i++ {
			txExe, err := transactions.CreateTransactionExecution(model.NewTransactionExecution{
				ProjectID: newProj.ID,
				Script:    tx2,
				Signers:   []model.Address{redeployAcc.Address},
			})
			require.NoError(t, err)
			assert.Len(t, txExe.Errors, 0)
		}

		exes, err := transactions.AllExecutionsForProjectID(newProj.ID)
		require.NoError(t, err)
		for i, exe := range exes {
			assert.Equal(t, exe.Index, i)
		}

	*/

}
