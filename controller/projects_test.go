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

func createProjects(t *testing.T) (*Projects, storage.Store, *model.User) {
	var store storage.Store

	if strings.EqualFold(os.Getenv("FLOW_STORAGEBACKEND"), storage.PostgreSQL) {
		store = storage.NewPostgreSQL()
	} else {
		store = storage.NewInMemory()
	}

	user := &model.User{
		ID: uuid.New(),
	}

	err := store.InsertUser(user)
	require.NoError(t, err)

	chain := blockchain.NewProjects(store, 5)
	return NewProjects(version, store, chain), store, user
}

func seedProject(projects *Projects, user *model.User) *model.Project {
	project, _ := projects.Create(user, model.NewProject{
		Title:                "test title",
		Description:          "test description",
		Readme:               "test readme",
		Seed:                 1,
		Accounts:             []string{"a"},
		TransactionTemplates: nil,
		ScriptTemplates:      nil,
	})
	return project
}

func Test_CreateProject(t *testing.T) {
	projects, store, user := createProjects(t)

	t.Run("successful creation", func(t *testing.T) {
		title := "test title"
		desc := "test desc"
		readme := "test readme"

		project, err := projects.Create(user, model.NewProject{
			Title:                title,
			Description:          desc,
			Readme:               readme,
			Seed:                 1,
			Accounts:             []string{"a"},
			TransactionTemplates: nil,
			ScriptTemplates:      nil,
		})
		require.NoError(t, err)
		assert.Equal(t, title, project.Title)
		assert.Equal(t, desc, project.Description)
		assert.Equal(t, readme, project.Readme)
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
		projects, store, user := createProjects(t)
		proj := seedProject(projects, user)

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
		projects, store, user := createProjects(t)
		proj := seedProject(projects, user)

		err := store.InsertTransactionExecution(&model.TransactionExecution{
			ID:        uuid.New(),
			ProjectID: proj.ID,
			Index:     6,
			Script:    "test",
		})
		require.NoError(t, err)

		accounts, err := projects.Reset(proj)
		require.Len(t, accounts, 5)

		var dbProj model.Project
		err = store.GetProject(proj.ID, &dbProj)
		require.NoError(t, err)

		assert.Equal(t, 5, dbProj.TransactionExecutionCount)
	})
}
