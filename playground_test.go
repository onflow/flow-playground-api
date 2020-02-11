package playground_test

import (
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/handler"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
)

const MutationCreateProject = `
mutation {
  createProject {
    id
  }
}
`

const QueryGetProject = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
  }
}
`

const QueryGetProjectTransactionTemplates = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
    transactionTemplates {
      id
      script 
      index
    }
  }
}
`

const MutationCreateTransactionTemplate = `
mutation($projectId: UUID!, $script: String!) {
  createTransactionTemplate(input: { projectId: $projectId, script: $script }) {
    id
	script
    index
  }
}
`

const QueryGetTransactionTemplate = `
query($templateId: UUID!) {
  transactionTemplate(id: $templateId) {
    id
    script
  }
}
`

const MutationUpdateTransactionTemplateScript = `
mutation($templateId: UUID!, $script: String!) {
  updateTransactionTemplate(input: { id: $templateId, script: $script }) {
    id
	script
    index
  }
}
`

const MutationUpdateTransactionTemplateIndex = `
mutation($templateId: UUID!, $index: Int!) {
  updateTransactionTemplate(input: { id: $templateId, index: $index }) {
    id
	script
    index
  }
}
`

const MutationDeleteTransactionTemplate = `
mutation($templateId: UUID!) {
  deleteTransactionTemplate(id: $templateId)
}
`

func TestProjects(t *testing.T) {
	t.Run("Create project", func(t *testing.T) {
		c := newClient()

		var resp struct {
			CreateProject struct{ ID string }
		}

		c.MustPost(MutationCreateProject, &resp)

		assert.NotEmpty(t, resp.CreateProject.ID)
	})

	t.Run("Get project", func(t *testing.T) {
		c := newClient()

		var respA struct {
			CreateProject struct{ ID string }
		}

		c.MustPost(MutationCreateProject, &respA)

		var respB struct {
			Project struct{ ID string }
		}

		c.MustPost(
			QueryGetProject,
			&respB,
			client.Var("projectId", respA.CreateProject.ID),
		)

		assert.Equal(t, respA.CreateProject.ID, respB.Project.ID)
	})

	t.Run("Get non-existent project", func(t *testing.T) {
		c := newClient()

		var resp struct {
			Project struct{ ID string }
		}

		badID := uuid.New().String()

		err := c.Post(
			QueryGetProject,
			&resp,
			client.Var("projectId", badID),
		)

		assert.Error(t, err)
	})
}

func TestTransactionTemplates(t *testing.T) {
	t.Run("Create transaction template", func(t *testing.T) {
		c := newClient()

		projectID := createProject(c)

		var resp struct {
			CreateTransactionTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		c.MustPost(
			MutationCreateTransactionTemplate,
			&resp,
			client.Var("projectId", projectID),
			client.Var("script", "foo"),
		)

		assert.NotEmpty(t, resp.CreateTransactionTemplate.ID)
		assert.Equal(t, "foo", resp.CreateTransactionTemplate.Script)
	})

	t.Run("Get transaction template", func(t *testing.T) {
		c := newClient()

		projectID := createProject(c)

		var respA struct {
			CreateTransactionTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		c.MustPost(
			MutationCreateTransactionTemplate,
			&respA,
			client.Var("projectId", projectID),
			client.Var("script", "foo"),
		)

		var respB struct {
			TransactionTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		c.MustPost(
			QueryGetTransactionTemplate,
			&respB,
			client.Var("templateId", respA.CreateTransactionTemplate.ID),
		)

		assert.Equal(t, respA.CreateTransactionTemplate.ID, respB.TransactionTemplate.ID)
		assert.Equal(t, respA.CreateTransactionTemplate.Script, respB.TransactionTemplate.Script)
	})

	t.Run("Get non-existent transaction template", func(t *testing.T) {
		c := newClient()

		var resp struct {
			TransactionTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		badID := uuid.New().String()

		err := c.Post(
			QueryGetTransactionTemplate,
			&resp,
			client.Var("templateId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Update transaction template", func(t *testing.T) {
		c := newClient()

		projectID := createProject(c)

		var respA struct {
			CreateTransactionTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		c.MustPost(
			MutationCreateTransactionTemplate,
			&respA,
			client.Var("projectId", projectID),
			client.Var("script", "foo"),
		)

		templateID := respA.CreateTransactionTemplate.ID

		var respB struct {
			UpdateTransactionTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		c.MustPost(
			MutationUpdateTransactionTemplateScript,
			&respB,
			client.Var("templateId", templateID),
			client.Var("script", "bar"),
		)

		assert.Equal(t, respA.CreateTransactionTemplate.ID, respB.UpdateTransactionTemplate.ID)
		assert.Equal(t, respA.CreateTransactionTemplate.Index, respB.UpdateTransactionTemplate.Index)
		assert.Equal(t, "bar", respB.UpdateTransactionTemplate.Script)

		var respC struct {
			UpdateTransactionTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		c.MustPost(
			MutationUpdateTransactionTemplateIndex,
			&respC,
			client.Var("templateId", templateID),
			client.Var("index", 1),
		)

		assert.Equal(t, respA.CreateTransactionTemplate.ID, respC.UpdateTransactionTemplate.ID)
		assert.Equal(t, 1, respC.UpdateTransactionTemplate.Index)
		assert.Equal(t, respB.UpdateTransactionTemplate.Script, respC.UpdateTransactionTemplate.Script)
	})

	t.Run("Update non-existent transaction template", func(t *testing.T) {
		c := newClient()

		var resp struct {
			TransactionTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		badID := uuid.New().String()

		err := c.Post(
			MutationUpdateTransactionTemplateScript,
			&resp,
			client.Var("templateId", badID),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Get transaction templates for project", func(t *testing.T) {
		c := newClient()

		projectID := createProject(c)

		templateIDA := createTransactionTemplate(c, projectID)
		templateIDB := createTransactionTemplate(c, projectID)
		templateIDC := createTransactionTemplate(c, projectID)

		var resp struct {
			Project struct {
				ID                   string
				TransactionTemplates []struct {
					ID     string
					Script string
					Index  int
				}
			}
		}

		c.MustPost(
			QueryGetProjectTransactionTemplates,
			&resp,
			client.Var("projectId", projectID),
		)

		assert.Len(t, resp.Project.TransactionTemplates, 3)
		assert.Equal(t, templateIDA, resp.Project.TransactionTemplates[0].ID)
		assert.Equal(t, templateIDB, resp.Project.TransactionTemplates[1].ID)
		assert.Equal(t, templateIDC, resp.Project.TransactionTemplates[2].ID)

		assert.Equal(t, 0, resp.Project.TransactionTemplates[0].Index)
		assert.Equal(t, 1, resp.Project.TransactionTemplates[1].Index)
		assert.Equal(t, 2, resp.Project.TransactionTemplates[2].Index)
	})

	t.Run("Get transaction templates for non-existent project", func(t *testing.T) {
		c := newClient()

		var resp struct {
			Project struct {
				TransactionTemplates []struct {
					ID     string
					Script string
				}
			}
		}

		badID := uuid.New().String()

		err := c.Post(
			QueryGetProjectTransactionTemplates,
			&resp,
			client.Var("projectId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Delete transaction template", func(t *testing.T) {
		c := newClient()

		projectID := createProject(c)

		templateID := createTransactionTemplate(c, projectID)

		var resp struct {
			DeleteTransactionTemplate string
		}

		c.MustPost(MutationDeleteTransactionTemplate, &resp, client.Var("templateId", templateID))

		assert.Equal(t, templateID, resp.DeleteTransactionTemplate)
	})
}

func newClient() *client.Client {
	resolver := playground.NewResolver(memory.NewStore())

	return client.New(
		handler.GraphQL(
			playground.NewExecutableSchema(playground.Config{Resolvers: resolver}),
		),
	)
}

func createProject(c *client.Client) string {
	var resp struct {
		CreateProject struct{ ID string }
	}

	c.MustPost(MutationCreateProject, &resp)

	return resp.CreateProject.ID
}

func createTransactionTemplate(c *client.Client, projectID string) string {
	var resp struct {
		CreateTransactionTemplate struct {
			ID     string
			Script string
			Index  int
		}
	}

	c.MustPost(
		MutationCreateTransactionTemplate,
		&resp,
		client.Var("projectId", projectID),
		client.Var("script", "foo"),
	)

	return resp.CreateTransactionTemplate.ID
}
