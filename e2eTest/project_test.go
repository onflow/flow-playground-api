package e2eTest

import (
	"github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

}
