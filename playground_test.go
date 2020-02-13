package playground_test

import (
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/handler"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
	"github.com/dapperlabs/flow-playground-api/vm"
)

const MutationCreateProject = `
mutation {
  createProject {
    id
  }
}
`

type CreateProjectResponse struct {
	CreateProject struct{ ID string }
}

const QueryGetProject = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
  }
}
`

type GetProjectResponse struct {
	Project struct{ ID string }
}

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

type GetProjectTransactionTemplatesResponse struct {
	Project struct {
		ID                   string
		TransactionTemplates []struct {
			ID     string
			Script string
			Index  int
		}
	}
}

const MutationCreateTransactionTemplate = `
mutation($projectId: UUID!, $script: String!) {
  createTransactionTemplate(input: { projectId: $projectId, script: $script }) {
    id
	script
    index
  }
}
`

type CreateTransactionTemplateResponse struct {
	CreateTransactionTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const QueryGetTransactionTemplate = `
query($templateId: UUID!) {
  transactionTemplate(id: $templateId) {
    id
    script
	index
  }
}
`

type GetTransactionTemplateResponse struct {
	TransactionTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

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

type UpdateTransactionTemplateResponse struct {
	UpdateTransactionTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const MutationDeleteTransactionTemplate = `
mutation($templateId: UUID!) {
  deleteTransactionTemplate(id: $templateId)
}
`

type DeleteTransactionTemplateResponse struct {
	DeleteTransactionTemplate string
}

const MutationCreateTransactionExecution = `
mutation($projectId: UUID!, $script: String!) {
  createTransactionExecution(input: {
    projectId: $projectId,
    script: $script,
  }) {
    id
    script
    error
	logs
    events {
      type
      values {
        type
        value
      }
    }
  }
}
`

type CreateTransactionExecutionResponse struct {
	CreateTransactionExecution struct {
		ID     string
		Script string
		Error  string
		Logs   []string
		Events []struct {
			Type   string
			Values []struct {
				Type  string
				Value string
			}
		}
	}
}

func TestProjects(t *testing.T) {
	t.Run("Create project", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		c.MustPost(MutationCreateProject, &resp)

		assert.NotEmpty(t, resp.CreateProject.ID)
	})

	t.Run("Get project", func(t *testing.T) {
		c := newClient()

		var respA CreateProjectResponse

		c.MustPost(MutationCreateProject, &respA)

		var respB GetProjectResponse

		c.MustPost(
			QueryGetProject,
			&respB,
			client.Var("projectId", respA.CreateProject.ID),
		)

		assert.Equal(t, respA.CreateProject.ID, respB.Project.ID)
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
}

func TestTransactionTemplates(t *testing.T) {
	t.Run("Create transaction template", func(t *testing.T) {
		c := newClient()

		projectID := createProject(c)

		var resp CreateTransactionTemplateResponse

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

		var respA CreateTransactionTemplateResponse

		c.MustPost(
			MutationCreateTransactionTemplate,
			&respA,
			client.Var("projectId", projectID),
			client.Var("script", "foo"),
		)

		var respB GetTransactionTemplateResponse

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

		var resp GetTransactionTemplateResponse

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

		var respA CreateTransactionTemplateResponse

		c.MustPost(
			MutationCreateTransactionTemplate,
			&respA,
			client.Var("projectId", projectID),
			client.Var("script", "foo"),
		)

		templateID := respA.CreateTransactionTemplate.ID

		var respB UpdateTransactionTemplateResponse

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

		var resp UpdateTransactionTemplateResponse

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

		var resp GetProjectTransactionTemplatesResponse

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

		var resp GetProjectTransactionTemplatesResponse

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

func TestTransactionExecutions(t *testing.T) {
	t.Run("Create execution for non-existent project", func(t *testing.T) {
		c := newClient()

		badID := uuid.New().String()

		var resp CreateTransactionExecutionResponse

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client.Var("projectId", badID),
			client.Var("script", "transaction { execute { log(\"Hello, World!\") } }"),
		)

		assert.Error(t, err)
	})

	t.Run("Create simple execution", func(t *testing.T) {
		c := newClient()

		projectID := createProject(c)

		var resp CreateTransactionExecutionResponse

		const script = "transaction { execute { log(\"Hello, World!\") } }"

		c.MustPost(
			MutationCreateTransactionExecution,
			&resp,
			client.Var("projectId", projectID),
			client.Var("script", script),
		)

		assert.Empty(t, resp.CreateTransactionExecution.Error)
		assert.Contains(t, resp.CreateTransactionExecution.Logs, "\"Hello, World!\"")
		assert.Equal(t, script, resp.CreateTransactionExecution.Script)
	})

	t.Run("Multiple executions", func(t *testing.T) {
		c := newClient()

		projectID := createProject(c)

		var respA CreateTransactionExecutionResponse

		const script = "transaction { execute { Account([], []) } }"

		c.MustPost(
			MutationCreateTransactionExecution,
			&respA,
			client.Var("projectId", projectID),
			client.Var("script", script),
		)

		assert.Empty(t, respA.CreateTransactionExecution.Error)
		require.Len(t, respA.CreateTransactionExecution.Events, 1)

		eventA := respA.CreateTransactionExecution.Events[0]

		// first account should have address 0x01
		assert.Equal(t, "flow.AccountCreated", eventA.Type)
		assert.Equal(t, "0000000000000000000000000000000000000001", eventA.Values[0].Value)

		var respB CreateTransactionExecutionResponse

		c.MustPost(
			MutationCreateTransactionExecution,
			&respB,
			client.Var("projectId", projectID),
			client.Var("script", script),
		)

		require.Len(t, respB.CreateTransactionExecution.Events, 1)

		eventB := respB.CreateTransactionExecution.Events[0]

		// second account should have address 0x02
		assert.Equal(t, "flow.AccountCreated", eventB.Type)
		assert.Equal(t, "0000000000000000000000000000000000000002", eventB.Values[0].Value)
	})

	t.Run("Multiple executions with cache reset", func(t *testing.T) {
		// manually construct resolver
		store := memory.NewStore()
		computer := vm.NewComputer(store)
		resolver := playground.NewResolver(store, computer)

		c := newClientWithResolve(resolver)

		projectID := createProject(c)

		var respA CreateTransactionExecutionResponse

		const script = "transaction { execute { Account([], []) } }"

		c.MustPost(
			MutationCreateTransactionExecution,
			&respA,
			client.Var("projectId", projectID),
			client.Var("script", script),
		)

		assert.Empty(t, respA.CreateTransactionExecution.Error)
		require.Len(t, respA.CreateTransactionExecution.Events, 1)

		eventA := respA.CreateTransactionExecution.Events[0]

		// first account should have address 0x01
		assert.Equal(t, "flow.AccountCreated", eventA.Type)
		assert.Equal(t, "0000000000000000000000000000000000000001", eventA.Values[0].Value)

		// clear ledger cache
		computer.ClearCache()

		var respB CreateTransactionExecutionResponse

		c.MustPost(
			MutationCreateTransactionExecution,
			&respB,
			client.Var("projectId", projectID),
			client.Var("script", script),
		)

		require.Len(t, respB.CreateTransactionExecution.Events, 1)

		eventB := respB.CreateTransactionExecution.Events[0]

		// second account should have address 0x02
		assert.Equal(t, "flow.AccountCreated", eventB.Type)
		assert.Equal(t, "0000000000000000000000000000000000000002", eventB.Values[0].Value)
	})
}

func newClient() *client.Client {
	store := memory.NewStore()
	computer := vm.NewComputer(store)

	resolver := playground.NewResolver(store, computer)

	return newClientWithResolve(resolver)
}

func newClientWithResolve(resolver *playground.Resolver) *client.Client {
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
