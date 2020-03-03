package playground_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	playground "github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/middleware"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/storage/datastore"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
	"github.com/dapperlabs/flow-playground-api/vm"
)

type Project struct {
	ID        string
	PrivateID string
	Persist   bool
	Accounts  []struct {
		ID        string
		Address   string
		DraftCode string
	}
	TransactionTemplates []TransactionTemplate
}

const MutationCreateProject = `
mutation($accounts: [String!], $transactionTemplates: [String!]) {
  createProject(input: { accounts: $accounts, transactionTemplates: $transactionTemplates }) {
    id
    persist
    accounts {
      id
      address
      draftCode
    }
    transactionTemplates {
      id
      script
      index
    }
  }
}
`

type CreateProjectResponse struct {
	CreateProject Project
}

const QueryGetProject = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
    accounts {
      id
      address
    }
  }
}
`

type GetProjectResponse struct {
	Project Project
}

const MutationUpdateProjectPersist = `
mutation($projectId: UUID!, $persist: Boolean!) {
  updateProject(input: { id: $projectId, persist: $persist }) {
    id
    persist
  }
}
`

type UpdateProjectResponse struct {
	UpdateProject struct {
		ID      string
		Persist bool
	}
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

const QueryGetProjectScriptTemplates = `
query($projectId: UUID!) {
  project(id: $projectId) {
    id
    scriptTemplates {
      id
      script 
      index
    }
  }
}
`

type GetProjectScriptTemplatesResponse struct {
	Project struct {
		ID              string
		ScriptTemplates []struct {
			ID     string
			Script string
			Index  int
		}
	}
}

const QueryGetAccount = `
query($accountId: UUID!) {
  account(id: $accountId) {
    id
    address
    draftCode
    deployedCode
  }
}
`

type GetAccountResponse struct {
	Account struct {
		ID           string
		Address      string
		DraftCode    string
		DeployedCode string
	}
}

const MutationUpdateAccountDraftCode = `
mutation($accountId: UUID!, $code: String!) {
  updateAccount(input: { id: $accountId, draftCode: $code }) {
    id
    address
    draftCode
    deployedCode
  }
}
`

const MutationUpdateAccountDeployedCode = `
mutation($accountId: UUID!, $code: String!) {
  updateAccount(input: { id: $accountId, deployedCode: $code }) {
    id
    address
    draftCode
    deployedCode
  }
}
`

type UpdateAccountResponse struct {
	UpdateAccount struct {
		ID           string
		Address      string
		DraftCode    string
		DeployedCode string
	}
}

type TransactionTemplate struct {
	ID     string
	Script string
	Index  int
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
	CreateTransactionTemplate TransactionTemplate
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
mutation($projectId: UUID!, $script: String!, $signers: [Address!]) {
  createTransactionExecution(input: {
    projectId: $projectId,
    script: $script,
    signers: $signers,
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

const MutationCreateScriptTemplate = `
mutation($projectId: UUID!, $script: String!) {
  createScriptTemplate(input: { projectId: $projectId, script: $script }) {
    id
    script
    index
  }
}
`

type CreateScriptTemplateResponse struct {
	CreateScriptTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const QueryGetScriptTemplate = `
query($templateId: UUID!) {
  scriptTemplate(id: $templateId) {
    id
    script
  }
}
`

type GetScriptTemplateResponse struct {
	ScriptTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const MutationUpdateScriptTemplateScript = `
mutation($templateId: UUID!, $script: String!) {
  updateScriptTemplate(input: { id: $templateId, script: $script }) {
    id
    script
    index
  }
}
`

const MutationUpdateScriptTemplateIndex = `
mutation($templateId: UUID!, $index: Int!) {
  updateScriptTemplate(input: { id: $templateId, index: $index }) {
    id
    script
    index
  }
}
`

type UpdateScriptTemplateResponse struct {
	UpdateScriptTemplate struct {
		ID     string
		Script string
		Index  int
	}
}

const MutationDeleteScriptTemplate = `
mutation($templateId: UUID!) {
  deleteScriptTemplate(id: $templateId)
}
`

type DeleteScriptTemplateResponse struct {
	DeleteScriptTemplate string
}

func TestProjects(t *testing.T) {
	t.Run("Create empty project", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		c.MustPost(
			MutationCreateProject,
			&resp,
		)

		assert.NotEmpty(t, resp.CreateProject.ID)

		// project should be created with 4 default accounts
		assert.Len(t, resp.CreateProject.Accounts, playground.MaxAccounts)

		// project should not be persisted
		assert.False(t, resp.CreateProject.Persist)
	})

	t.Run("Create project with 2 accounts", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		accounts := []string{
			"pub contract Foo {}",
			"pub contract Bar {}",
		}

		c.MustPost(
			MutationCreateProject,
			&resp,
			client.Var("accounts", accounts),
		)

		// project should still be created with 4 default accounts
		assert.Len(t, resp.CreateProject.Accounts, playground.MaxAccounts)

		assert.Equal(t, accounts[0], resp.CreateProject.Accounts[0].DraftCode)
		assert.Equal(t, accounts[1], resp.CreateProject.Accounts[1].DraftCode)
		assert.Equal(t, "", resp.CreateProject.Accounts[2].DraftCode)
	})

	t.Run("Create project with 4 accounts", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		accounts := []string{
			"pub contract Foo {}",
			"pub contract Bar {}",
			"pub contract Dog {}",
			"pub contract Cat {}",
		}

		c.MustPost(
			MutationCreateProject,
			&resp,
			client.Var("accounts", accounts),
		)

		// project should still be created with 4 default accounts
		assert.Len(t, resp.CreateProject.Accounts, playground.MaxAccounts)

		assert.Equal(t, accounts[0], resp.CreateProject.Accounts[0].DraftCode)
		assert.Equal(t, accounts[1], resp.CreateProject.Accounts[1].DraftCode)
		assert.Equal(t, accounts[2], resp.CreateProject.Accounts[2].DraftCode)
	})

	t.Run("Create project with transaction templates", func(t *testing.T) {
		c := newClient()

		var resp CreateProjectResponse

		templates := []string{
			"transaction { execute { log(\"foo\") } }",
			"transaction { execute { log(\"bar\") } }",
		}

		c.MustPost(
			MutationCreateProject,
			&resp,
			client.Var("transactionTemplates", templates),
		)

		assert.Len(t, resp.CreateProject.TransactionTemplates, 2)
		assert.Equal(t, templates[0], resp.CreateProject.TransactionTemplates[0].Script)
		assert.Equal(t, templates[1], resp.CreateProject.TransactionTemplates[1].Script)
	})

	t.Run("Get project", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp GetProjectResponse

		c.MustPost(
			QueryGetProject,
			&resp,
			client.Var("projectId", project.ID),
		)

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

		project := createProject(c)

		var resp UpdateProjectResponse

		err := c.Post(
			MutationUpdateProjectPersist,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("persist", true),
		)

		assert.Error(t, err)
	})

	t.Run("Persist project", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp UpdateProjectResponse

		c.MustPost(
			MutationUpdateProjectPersist,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("persist", true),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.Equal(t, project.ID, resp.UpdateProject.ID)
		assert.True(t, resp.UpdateProject.Persist)
	})
}

func TestTransactionTemplates(t *testing.T) {
	t.Run("Create transaction template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateTransactionTemplateResponse

		err := c.Post(
			MutationCreateTransactionTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", "foo"),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.CreateTransactionTemplate.ID)
	})

	t.Run("Create transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateTransactionTemplateResponse

		c.MustPost(
			MutationCreateTransactionTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", "foo"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.NotEmpty(t, resp.CreateTransactionTemplate.ID)
		assert.Equal(t, "foo", resp.CreateTransactionTemplate.Script)
	})

	t.Run("Get transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateTransactionTemplateResponse

		c.MustPost(
			MutationCreateTransactionTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", "foo"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
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

	t.Run("Update transaction template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateTransactionTemplateResponse

		c.MustPost(
			MutationCreateTransactionTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", "foo"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		templateID := respA.CreateTransactionTemplate.ID

		var respB UpdateTransactionTemplateResponse

		err := c.Post(
			MutationUpdateTransactionTemplateScript,
			&respB,
			client.Var("templateId", templateID),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Update transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateTransactionTemplateResponse

		c.MustPost(
			MutationCreateTransactionTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", "foo"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		templateID := respA.CreateTransactionTemplate.ID

		var respB UpdateTransactionTemplateResponse

		c.MustPost(
			MutationUpdateTransactionTemplateScript,
			&respB,
			client.Var("templateId", templateID),
			client.Var("script", "bar"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
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
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
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

		project := createProject(c)

		templateA := createTransactionTemplate(c, project)
		templateB := createTransactionTemplate(c, project)
		templateC := createTransactionTemplate(c, project)

		var resp GetProjectTransactionTemplatesResponse

		c.MustPost(
			QueryGetProjectTransactionTemplates,
			&resp,
			client.Var("projectId", project.ID),
		)

		assert.Len(t, resp.Project.TransactionTemplates, 3)
		assert.Equal(t, templateA.ID, resp.Project.TransactionTemplates[0].ID)
		assert.Equal(t, templateB.ID, resp.Project.TransactionTemplates[1].ID)
		assert.Equal(t, templateC.ID, resp.Project.TransactionTemplates[2].ID)

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

	t.Run("Delete transaction template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		template := createTransactionTemplate(c, project)

		var resp DeleteTransactionTemplateResponse

		err := c.Post(
			MutationDeleteTransactionTemplate,
			&resp,
			client.Var("templateId", template.ID),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.DeleteTransactionTemplate)
	})

	t.Run("Delete transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		template := createTransactionTemplate(c, project)

		var resp DeleteTransactionTemplateResponse

		c.MustPost(
			MutationDeleteTransactionTemplate,
			&resp,
			client.Var("templateId", template.ID),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.Equal(t, template.ID, resp.DeleteTransactionTemplate)
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

	t.Run("Create execution without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateTransactionExecutionResponse

		const script = "transaction { execute { log(\"Hello, World!\") } }"

		err := c.Post(
			MutationCreateTransactionExecution,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", script),
		)

		assert.Error(t, err)
	})

	t.Run("Create execution", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateTransactionExecutionResponse

		const script = "transaction { execute { log(\"Hello, World!\") } }"

		c.MustPost(
			MutationCreateTransactionExecution,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", script),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.Empty(t, resp.CreateTransactionExecution.Error)
		assert.Contains(t, resp.CreateTransactionExecution.Logs, "\"Hello, World!\"")
		assert.Equal(t, script, resp.CreateTransactionExecution.Script)
	})

	t.Run("Multiple executions", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateTransactionExecutionResponse

		const script = "transaction { execute { Account([], []) } }"

		c.MustPost(
			MutationCreateTransactionExecution,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", script),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.Empty(t, respA.CreateTransactionExecution.Error)
		require.Len(t, respA.CreateTransactionExecution.Events, 1)

		eventA := respA.CreateTransactionExecution.Events[0]

		// first account should have address 0x05
		assert.Equal(t, "flow.AccountCreated", eventA.Type)
		assert.Equal(t, "0000000000000000000000000000000000000005", eventA.Values[0].Value)

		var respB CreateTransactionExecutionResponse

		c.MustPost(
			MutationCreateTransactionExecution,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("script", script),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		require.Len(t, respB.CreateTransactionExecution.Events, 1)

		eventB := respB.CreateTransactionExecution.Events[0]

		// second account should have address 0x06
		assert.Equal(t, "flow.AccountCreated", eventB.Type)
		assert.Equal(t, "0000000000000000000000000000000000000006", eventB.Values[0].Value)
	})

	t.Run("Multiple executions with cache reset", func(t *testing.T) {
		// manually construct resolver
		store := memory.NewStore()
		computer := vm.NewComputer(store)
		resolver := playground.NewResolver(store, computer)

		c := newClientWithResolver(resolver)

		project := createProject(c)

		var respA CreateTransactionExecutionResponse

		const script = "transaction { execute { Account([], []) } }"

		c.MustPost(
			MutationCreateTransactionExecution,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", script),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.Empty(t, respA.CreateTransactionExecution.Error)
		require.Len(t, respA.CreateTransactionExecution.Events, 1)

		eventA := respA.CreateTransactionExecution.Events[0]

		// first account should have address 0x05
		assert.Equal(t, "flow.AccountCreated", eventA.Type)
		assert.Equal(t, "0000000000000000000000000000000000000005", eventA.Values[0].Value)

		// clear ledger cache
		computer.ClearCache()

		var respB CreateTransactionExecutionResponse

		c.MustPost(
			MutationCreateTransactionExecution,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("script", script),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		require.Len(t, respB.CreateTransactionExecution.Events, 1)

		eventB := respB.CreateTransactionExecution.Events[0]

		// second account should have address 0x06
		assert.Equal(t, "flow.AccountCreated", eventB.Type)
		assert.Equal(t, "0000000000000000000000000000000000000006", eventB.Values[0].Value)
	})
}

func TestScriptTemplates(t *testing.T) {
	t.Run("Create script template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateScriptTemplateResponse

		err := c.Post(
			MutationCreateScriptTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", "foo"),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.CreateScriptTemplate.ID)
	})

	t.Run("Create script template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var resp CreateScriptTemplateResponse

		c.MustPost(
			MutationCreateScriptTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", "foo"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.NotEmpty(t, resp.CreateScriptTemplate.ID)
		assert.Equal(t, "foo", resp.CreateScriptTemplate.Script)
	})

	t.Run("Get script template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateScriptTemplateResponse

		c.MustPost(
			MutationCreateScriptTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", "foo"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		var respB GetScriptTemplateResponse

		c.MustPost(
			QueryGetScriptTemplate,
			&respB,
			client.Var("templateId", respA.CreateScriptTemplate.ID),
		)

		assert.Equal(t, respA.CreateScriptTemplate.ID, respB.ScriptTemplate.ID)
		assert.Equal(t, respA.CreateScriptTemplate.Script, respB.ScriptTemplate.Script)
	})

	t.Run("Get non-existent script template", func(t *testing.T) {
		c := newClient()

		var resp GetScriptTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetScriptTemplate,
			&resp,
			client.Var("templateId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Update script template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateScriptTemplateResponse

		c.MustPost(
			MutationCreateScriptTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", "foo"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		templateID := respA.CreateScriptTemplate.ID

		var respB UpdateScriptTemplateResponse

		err := c.Post(
			MutationUpdateScriptTemplateScript,
			&respB,
			client.Var("templateId", templateID),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Update script template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		var respA CreateScriptTemplateResponse

		c.MustPost(
			MutationCreateScriptTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", "foo"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		templateID := respA.CreateScriptTemplate.ID

		var respB UpdateScriptTemplateResponse

		c.MustPost(
			MutationUpdateScriptTemplateScript,
			&respB,
			client.Var("templateId", templateID),
			client.Var("script", "bar"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.Equal(t, respA.CreateScriptTemplate.ID, respB.UpdateScriptTemplate.ID)
		assert.Equal(t, respA.CreateScriptTemplate.Index, respB.UpdateScriptTemplate.Index)
		assert.Equal(t, "bar", respB.UpdateScriptTemplate.Script)

		var respC UpdateScriptTemplateResponse

		c.MustPost(
			MutationUpdateScriptTemplateIndex,
			&respC,
			client.Var("templateId", templateID),
			client.Var("index", 1),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.Equal(t, respA.CreateScriptTemplate.ID, respC.UpdateScriptTemplate.ID)
		assert.Equal(t, 1, respC.UpdateScriptTemplate.Index)
		assert.Equal(t, respB.UpdateScriptTemplate.Script, respC.UpdateScriptTemplate.Script)
	})

	t.Run("Update non-existent script template", func(t *testing.T) {
		c := newClient()

		var resp UpdateScriptTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			MutationUpdateScriptTemplateScript,
			&resp,
			client.Var("templateId", badID),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Get script templates for project", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		templateIDA := createScriptTemplate(c, project)
		templateIDB := createScriptTemplate(c, project)
		templateIDC := createScriptTemplate(c, project)

		var resp GetProjectScriptTemplatesResponse

		c.MustPost(
			QueryGetProjectScriptTemplates,
			&resp,
			client.Var("projectId", project.ID),
		)

		assert.Len(t, resp.Project.ScriptTemplates, 3)
		assert.Equal(t, templateIDA, resp.Project.ScriptTemplates[0].ID)
		assert.Equal(t, templateIDB, resp.Project.ScriptTemplates[1].ID)
		assert.Equal(t, templateIDC, resp.Project.ScriptTemplates[2].ID)

		assert.Equal(t, 0, resp.Project.ScriptTemplates[0].Index)
		assert.Equal(t, 1, resp.Project.ScriptTemplates[1].Index)
		assert.Equal(t, 2, resp.Project.ScriptTemplates[2].Index)
	})

	t.Run("Get script templates for non-existent project", func(t *testing.T) {
		c := newClient()

		var resp GetProjectScriptTemplatesResponse

		badID := uuid.New().String()

		err := c.Post(

			QueryGetProjectScriptTemplates,
			&resp,
			client.Var("projectId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Delete script template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		templateID := createScriptTemplate(c, project)

		var resp DeleteScriptTemplateResponse

		err := c.Post(
			MutationDeleteScriptTemplate,
			&resp,
			client.Var("templateId", templateID),
		)

		assert.Error(t, err)
	})

	t.Run("Delete script template", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		templateID := createScriptTemplate(c, project)

		var resp DeleteScriptTemplateResponse

		c.MustPost(
			MutationDeleteScriptTemplate,
			&resp,
			client.Var("templateId", templateID),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.Equal(t, templateID, resp.DeleteScriptTemplate)
	})
}

func TestAccounts(t *testing.T) {
	t.Run("Get account", func(t *testing.T) {
		c := newClient()

		project := createProject(c)
		account := project.Accounts[0]

		var resp GetAccountResponse

		c.MustPost(
			QueryGetAccount,
			&resp,
			client.Var("accountId", account.ID),
		)

		assert.Equal(t, account.ID, resp.Account.ID)
	})

	t.Run("Get non-existent account", func(t *testing.T) {
		c := newClient()

		var resp GetAccountResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetAccount,
			&resp,
			client.Var("accountId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Update account draft code without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)
		account := project.Accounts[0]

		var respA GetAccountResponse

		c.MustPost(
			QueryGetAccount,
			&respA,
			client.Var("accountId", account.ID),
		)

		assert.Equal(t, "", respA.Account.DraftCode)

		var respB UpdateAccountResponse

		err := c.Post(
			MutationUpdateAccountDraftCode,
			&respB,
			client.Var("accountId", account.ID),
			client.Var("code", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Update account draft code", func(t *testing.T) {
		c := newClient()

		project := createProject(c)
		account := project.Accounts[0]

		var respA GetAccountResponse

		c.MustPost(
			QueryGetAccount,
			&respA,
			client.Var("accountId", account.ID),
		)

		assert.Equal(t, "", respA.Account.DraftCode)

		var respB UpdateAccountResponse

		c.MustPost(
			MutationUpdateAccountDraftCode,
			&respB,
			client.Var("accountId", account.ID),
			client.Var("code", "bar"),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.Equal(t, "bar", respB.UpdateAccount.DraftCode)
	})

	t.Run("Update account invalid deployed code", func(t *testing.T) {
		c := newClient()

		project := createProject(c)
		account := project.Accounts[0]

		var respA GetAccountResponse

		c.MustPost(
			QueryGetAccount,
			&respA,
			client.Var("accountId", account.ID),
		)

		assert.Equal(t, "", respA.Account.DeployedCode)

		var respB UpdateAccountResponse

		err := c.Post(
			MutationUpdateAccountDeployedCode,
			&respB,
			client.Var("accountId", account.ID),
			client.Var("code", "INVALID CADENCE"),
		)

		assert.Error(t, err)
		assert.Equal(t, "", respB.UpdateAccount.DeployedCode)
	})

	t.Run("Update account deployed code without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		account := project.Accounts[0]

		var resp UpdateAccountResponse

		const contract = "pub contract Foo {}"

		err := c.Post(
			MutationUpdateAccountDeployedCode,
			&resp,
			client.Var("accountId", account.ID),
			client.Var("code", contract),
		)

		assert.Error(t, err)
	})

	t.Run("Update account deployed code", func(t *testing.T) {
		c := newClient()

		project := createProject(c)

		account := project.Accounts[0]

		var respA GetAccountResponse

		c.MustPost(
			QueryGetAccount,
			&respA,
			client.Var("accountId", account.ID),
		)

		assert.Equal(t, "", respA.Account.DeployedCode)

		var respB UpdateAccountResponse

		const contract = "pub contract Foo {}"

		c.MustPost(
			MutationUpdateAccountDeployedCode,
			&respB,
			client.Var("accountId", account.ID),
			client.Var("code", contract),
			client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
		)

		assert.Equal(t, contract, respB.UpdateAccount.DeployedCode)
	})

	t.Run("Update non-existent account", func(t *testing.T) {
		c := newClient()

		var resp UpdateAccountResponse

		badID := uuid.New().String()

		err := c.Post(
			MutationUpdateAccountDraftCode,
			&resp,
			client.Var("accountId", badID),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
	})
}

const counterContract = `
  pub contract Counting {

      pub event CountIncremented(count: Int)

      pub resource Counter {
          pub var count: Int

          init() {
              self.count = 0
          }

          pub fun add(_ count: Int) {
              self.count = self.count + count
              emit CountIncremented(count: self.count)
          }
      }

      pub fun createCounter(): @Counter {
          return <-create Counter()
      }
  }
`

// generateAddTwoToCounterScript generates a script that increments a counter.
// If no counter exists, it is created.
func generateAddTwoToCounterScript(counterAddress string) string {
	return fmt.Sprintf(
		`
            import 0x%s

            transaction {

                prepare(signer: Account) {
                    if signer.storage[Counting.Counter] == nil {
                        let existing <- signer.storage[Counting.Counter] <- Counting.createCounter()
                        destroy existing

                        signer.published[&Counting.Counter] = &signer.storage[Counting.Counter] as &Counting.Counter
                    }

                    signer.published[&Counting.Counter]?.add(2)
                }
            }
        `,
		counterAddress,
	)
}

func TestContractInteraction(t *testing.T) {
	c := newClient()

	project := createProject(c)

	accountA := project.Accounts[0]
	accountB := project.Accounts[1]

	var respA UpdateAccountResponse

	c.MustPost(
		MutationUpdateAccountDeployedCode,
		&respA,
		client.Var("accountId", accountA.ID),
		client.Var("code", counterContract),
		client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
	)

	assert.Equal(t, counterContract, respA.UpdateAccount.DeployedCode)

	addScript := generateAddTwoToCounterScript(accountA.Address)

	var respB CreateTransactionExecutionResponse

	c.MustPost(
		MutationCreateTransactionExecution,
		&respB,
		client.Var("projectId", project.ID),
		client.Var("script", addScript),
		client.Var("signers", []string{accountB.Address}),
		client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
	)

	assert.Empty(t, respB.CreateTransactionExecution.Error)

}

type Client struct {
	client   *client.Client
	resolver *playground.Resolver
}

func (c *Client) Post(query string, response interface{}, options ...client.Option) error {
	return c.client.Post(query, response, options...)
}

func (c *Client) MustPost(query string, response interface{}, options ...client.Option) {
	c.client.MustPost(query, response, options...)
}

func newClient() *Client {
	var store storage.Store

	// TODO: Should eventually start up the emulator and run all tests with datastore backend
	if strings.EqualFold(os.Getenv("STORE_BACKEND"), "datastore") {
		var err error
		store, err = datastore.NewDatastore(context.Background(), &datastore.Config{
			DatastoreProjectID: "dl-flow",
			DatastoreTimeout:   time.Second * 5,
		})
		if err != nil {
			// If datastore is expected, panic when we can't init
			panic(err)
		}
	} else {
		store = memory.NewStore()
	}

	computer := vm.NewComputer(store)

	resolver := playground.NewResolver(store, computer)

	return newClientWithResolver(resolver)
}

func newClientWithResolver(resolver *playground.Resolver) *Client {
	router := chi.NewRouter()
	router.Use(middleware.MockProjectSessions())

	router.Handle(
		"/",
		handler.GraphQL(
			playground.NewExecutableSchema(playground.Config{Resolvers: resolver}),
		),
	)

	return &Client{
		client:   client.New(router),
		resolver: resolver,
	}
}

func createProject(c *Client) Project {
	var resp CreateProjectResponse

	c.MustPost(
		MutationCreateProject,
		&resp,
		client.Var("accounts", []string{}),
		client.Var("transactionTemplates", []string{}),
	)

	proj := resp.CreateProject
	internalProj := c.resolver.LastCreatedProject()

	proj.PrivateID = internalProj.PrivateID.String()

	return proj
}

func createTransactionTemplate(c *Client, project Project) TransactionTemplate {
	var resp CreateTransactionTemplateResponse

	c.MustPost(
		MutationCreateTransactionTemplate,
		&resp,
		client.Var("projectId", project.ID),
		client.Var("script", "foo"),
		client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
	)

	return resp.CreateTransactionTemplate
}

func createScriptTemplate(c *Client, project Project) string {
	var resp struct {
		CreateScriptTemplate struct {
			ID     string
			Script string
			Index  int
		}
	}

	c.MustPost(
		MutationCreateScriptTemplate,
		&resp,
		client.Var("projectId", project.ID),
		client.Var("script", "foo"),
		client.AddCookie(middleware.MockProjectSessionCookie(project.ID, project.PrivateID)),
	)

	return resp.CreateScriptTemplate.ID
}
