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

package e2eTest

import (
	"github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestProjects(t *testing.T) {

	t.Run("Create empty project", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		err := c.Post(
			MutationCreateProject,
			&resp,
			client.Var("title", "foo"),
			client.Var("description", "bar"),
			client.Var("readme", "bah"),
			client.Var("seed", 42),
			client.Var("numberOfAccounts", initAccounts),
		)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.CreateProject.ID)
		assert.Equal(t, 42, resp.CreateProject.Seed)
		//assert.Equal(t, version.String(), resp.CreateProject.Version)

		// project should be created with 5 default accounts
		assert.Equal(t, initAccounts, resp.CreateProject.NumberOfAccounts)

		// project should not be persisted
		assert.False(t, resp.CreateProject.Persist)
	})

	t.Run("Create project with 2 contract templates", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		contractTemplates := []*model.NewProjectContractTemplate{
			{Title: "Foo", Script: "pub contract Foo {}"},
			{Title: "Bar", Script: "pub contract Bar {}"},
		}

		err := c.Post(
			MutationCreateProject,
			&resp,
			client.Var("title", "foo"),
			client.Var("description", "desc"),
			client.Var("readme", "rtfm"),
			client.Var("seed", 42),
			client.Var("numberOfAccounts", initAccounts),
			client.Var("contractTemplates", contractTemplates),
		)
		require.NoError(t, err)

		// Verify contract templates
		assert.Equal(t, "Foo", resp.CreateProject.ContractTemplates[0].Title)
		assert.Equal(t, "pub contract Foo {}", resp.CreateProject.ContractTemplates[0].Script)
		assert.Equal(t, "Bar", resp.CreateProject.ContractTemplates[1].Title)
		assert.Equal(t, "pub contract Bar {}", resp.CreateProject.ContractTemplates[1].Script)

	})

	t.Run("Create project with transaction templates", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		templates := []struct {
			Title  string `json:"title"`
			Script string `json:"script"`
		}{
			{
				"foo", "transaction { execute { log(\"foo\") } }",
			},
			{
				"bar", "transaction { execute { log(\"bar\") } }",
			},
		}

		err := c.Post(
			MutationCreateProject,
			&resp,
			client.Var("title", "foo"),
			client.Var("seed", 42),
			client.Var("description", "desc"),
			client.Var("readme", "rtfm"),
			client.Var("transactionTemplates", templates),
			client.Var("numberOfAccounts", initAccounts),
		)
		require.NoError(t, err)

		assert.Len(t, resp.CreateProject.TransactionTemplates, 2)
		assert.Equal(t, templates[0].Title, resp.CreateProject.TransactionTemplates[0].Title)
		assert.Equal(t, templates[0].Script, resp.CreateProject.TransactionTemplates[0].Script)
		assert.Equal(t, templates[1].Title, resp.CreateProject.TransactionTemplates[1].Title)
		assert.Equal(t, templates[1].Script, resp.CreateProject.TransactionTemplates[1].Script)
	})

	t.Run("Get project", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp GetProjectResponse

		err := c.Post(
			QueryGetProject,
			&resp,
			client.Var("projectId", project.ID),
		)
		require.NoError(t, err)

		assert.Equal(t, project.ID, resp.Project.ID)
	})

	t.Run("Get non-existent project", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetProject,
			&resp,
			client.Var("projectId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Persist project without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp UpdateProjectResponse

		err := c.Post(
			MutationUpdateProjectPersist,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("title", project.Title),
			client.Var("description", project.Description),
			client.Var("readme", project.Readme),
			client.Var("persist", true),
		)

		assert.Error(t, err)
	})

	t.Run("Persist project", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp UpdateProjectResponse

		err := c.Post(
			MutationUpdateProjectPersist,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("title", project.Title),
			client.Var("description", project.Description),
			client.Var("readme", project.Readme),
			client.Var("persist", true),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, project.ID, resp.UpdateProject.ID)
		assert.Equal(t, project.Title, resp.UpdateProject.Title)
		assert.Equal(t, project.Description, resp.UpdateProject.Description)
		assert.Equal(t, project.Readme, resp.UpdateProject.Readme)
		assert.True(t, resp.UpdateProject.Persist)
	})

	t.Run("Delete project", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse
		err := c.Post(
			MutationCreateProject,
			&resp,
			client.Var("title", "foo1"),
			client.Var("description", "bar"),
			client.Var("readme", "bah"),
			client.Var("seed", 42),
			client.Var("numberOfAccounts", initAccounts),
		)
		require.NoError(t, err)

		err = c.Post(
			MutationCreateProject,
			&resp,
			client.Var("title", "foo2"),
			client.Var("description", "bar"),
			client.Var("readme", "bah"),
			client.Var("seed", 42),
			client.Var("numberOfAccounts", initAccounts),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		var deleteResp DeleteProjectResponse
		err = c.Post(
			MutationDeleteProject,
			&deleteResp,
			client.Var("projectId", resp.CreateProject.ID),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		assert.Equal(t, deleteResp.DeleteProject, resp.CreateProject.ID)
	})

	t.Run("Maximum projects limit", func(t *testing.T) {
		const MaxProjectsLimit = 10
		const additionalAttempts = 5 // Try to create projects over the limit

		c := newClient()

		var firstProjResp CreateProjectResponse
		var resp CreateProjectResponse

		var err error = nil
		for projNum := 1; projNum <= MaxProjectsLimit+additionalAttempts; projNum++ {
			if projNum == 1 {
				err = c.Post(
					MutationCreateProject,
					&firstProjResp,
					client.Var("title", "foo"+strconv.Itoa(projNum)),
					client.Var("description", "bar"),
					client.Var("readme", "bah"),
					client.Var("seed", 42),
					client.Var("numberOfAccounts", initAccounts),
				)
				resp = firstProjResp
			} else {
				// Post with session cookie to keep the same userID
				err = c.Post(
					MutationCreateProject,
					&resp,
					client.Var("title", "foo"+strconv.Itoa(projNum)),
					client.Var("description", "bar"),
					client.Var("readme", "bah"),
					client.Var("seed", 42),
					client.Var("numberOfAccounts", initAccounts),
					client.AddCookie(c.SessionCookie()),
				)
			}
			if projNum <= MaxProjectsLimit {
				require.NoError(t, err)
				assert.NotEmpty(t, resp.CreateProject.ID)
				assert.Equal(t, 42, resp.CreateProject.Seed)
				assert.Equal(t, initAccounts, resp.CreateProject.NumberOfAccounts)
			} else {
				require.Error(t, err)
			}
		}

		// Delete a project and make sure we can create a new one
		var deleteResp DeleteProjectResponse
		err = c.Post(
			MutationDeleteProject,
			&deleteResp,
			client.Var("projectId", firstProjResp.CreateProject.ID),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		assert.Equal(t, firstProjResp.CreateProject.ID, deleteResp.DeleteProject)

		err = c.Post(
			MutationCreateProject,
			&resp,
			client.Var("title", "fooNew"),
			client.Var("description", "bar"),
			client.Var("readme", "bah"),
			client.Var("seed", 42),
			client.Var("numberOfAccounts", initAccounts),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.CreateProject.ID)
		assert.Equal(t, 42, resp.CreateProject.Seed)
		assert.Equal(t, initAccounts, resp.CreateProject.NumberOfAccounts)
	})

}

func TestGetProjectList(t *testing.T) {
	t.Run("get project list", func(t *testing.T) {
		c := newClient()

		var projResp1 CreateProjectResponse
		err := c.Post(
			MutationCreateProject,
			&projResp1,
			client.Var("title", "foo1"),
			client.Var("description", "bar"),
			client.Var("readme", "bah"),
			client.Var("seed", 42),
			client.Var("numberOfAccounts", initAccounts),
		)
		require.NoError(t, err)

		var projResp2 CreateProjectResponse
		err = c.Post(
			MutationCreateProject,
			&projResp2,
			client.Var("title", "foo2"),
			client.Var("description", "bar"),
			client.Var("readme", "bah"),
			client.Var("seed", 42),
			client.Var("numberOfAccounts", initAccounts),
			client.AddCookie(c.SessionCookie()), // Use the same cookie for the same userID
		)
		require.NoError(t, err)

		var listResp GetProjectListResponse

		err = c.Post(
			QueryGetProjectList,
			&listResp,
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, projResp1.CreateProject.ID, listResp.ProjectList.Projects[1].ID)
		assert.Equal(t, "foo1", listResp.ProjectList.Projects[1].Title)
		assert.Equal(t, projResp2.CreateProject.ID, listResp.ProjectList.Projects[0].ID)
		assert.Equal(t, "foo2", listResp.ProjectList.Projects[0].Title)
	})

	t.Run("validate 2 users project lists", func(t *testing.T) {
		c := newClient()

		var projResp1 CreateProjectResponse
		err := c.Post(
			MutationCreateProject,
			&projResp1,
			client.Var("title", "foo1"),
			client.Var("description", "bar"),
			client.Var("readme", "bah"),
			client.Var("seed", 42),
			client.Var("numberOfAccounts", initAccounts),
		)
		require.NoError(t, err)

		user1 := c.SessionCookie()

		var projResp2 CreateProjectResponse
		err = c.Post(
			MutationCreateProject,
			&projResp2,
			client.Var("title", "foo2"),
			client.Var("description", "bar"),
			client.Var("readme", "bah"),
			client.Var("seed", 42),
			client.Var("numberOfAccounts", initAccounts),
		)
		require.NoError(t, err)

		user2 := c.SessionCookie()

		var user1ListResp GetProjectListResponse
		err = c.Post(
			QueryGetProjectList,
			&user1ListResp,
			client.AddCookie(user1),
		)
		require.NoError(t, err)

		assert.Equal(t, 1, len(user1ListResp.ProjectList.Projects))
		assert.Equal(t, "foo1", user1ListResp.ProjectList.Projects[0].Title)

		var user2ListResp GetProjectListResponse
		err = c.Post(
			QueryGetProjectList,
			&user2ListResp,
			client.AddCookie(user2),
		)
		require.NoError(t, err)

		assert.Equal(t, 1, len(user2ListResp.ProjectList.Projects))
		assert.Equal(t, "foo2", user2ListResp.ProjectList.Projects[0].Title)
	})
}
