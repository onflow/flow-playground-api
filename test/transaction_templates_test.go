package test

import (
	client2 "github.com/dapperlabs/flow-playground-api/test/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransactionTemplates(t *testing.T) {
	t.Run("Create transaction template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateTransactionTemplateResponse

		err := c.Post(
			MutationCreateTransactionTemplate,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("title", "foo"),
			client2.Var("script", "bar"),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.CreateTransactionTemplate.ID)
	})

	t.Run("Create transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateTransactionTemplateResponse

		err := c.Post(
			MutationCreateTransactionTemplate,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("title", "foo"),
			client2.Var("script", "bar"),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.CreateTransactionTemplate.ID)
		assert.Equal(t, "foo", resp.CreateTransactionTemplate.Title)
		assert.Equal(t, "bar", resp.CreateTransactionTemplate.Script)
	})

	t.Run("Get transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA CreateTransactionTemplateResponse

		err := c.Post(
			MutationCreateTransactionTemplate,
			&respA,
			client2.Var("projectId", project.ID),
			client2.Var("title", "foo"),
			client2.Var("script", "bar"),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		var respB GetTransactionTemplateResponse

		err = c.Post(
			QueryGetTransactionTemplate,
			&respB,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", respA.CreateTransactionTemplate.ID),
		)
		require.NoError(t, err)

		assert.Equal(t, respA.CreateTransactionTemplate.ID, respB.TransactionTemplate.ID)
		assert.Equal(t, respA.CreateTransactionTemplate.Script, respB.TransactionTemplate.Script)
	})

	t.Run("Get non-existent transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp GetTransactionTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetTransactionTemplate,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Update transaction template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA CreateTransactionTemplateResponse

		err := c.Post(
			MutationCreateTransactionTemplate,
			&respA,
			client2.Var("projectId", project.ID),
			client2.Var("title", "foo"),
			client2.Var("script", "apple"),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		templateID := respA.CreateTransactionTemplate.ID

		var respB UpdateTransactionTemplateResponse

		err = c.Post(
			MutationUpdateTransactionTemplateScript,
			&respB,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", templateID),
			client2.Var("script", "orange"),
		)
		assert.Error(t, err)
	})

	t.Run("Update transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA CreateTransactionTemplateResponse

		err := c.Post(
			MutationCreateTransactionTemplate,
			&respA,
			client2.Var("projectId", project.ID),
			client2.Var("title", "foo"),
			client2.Var("script", "apple"),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		templateID := respA.CreateTransactionTemplate.ID

		var respB UpdateTransactionTemplateResponse

		err = c.Post(
			MutationUpdateTransactionTemplateScript,
			&respB,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", templateID),
			client2.Var("script", "orange"),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, respA.CreateTransactionTemplate.ID, respB.UpdateTransactionTemplate.ID)
		assert.Equal(t, respA.CreateTransactionTemplate.Index, respB.UpdateTransactionTemplate.Index)
		assert.Equal(t, "orange", respB.UpdateTransactionTemplate.Script)

		var respC struct {
			UpdateTransactionTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		err = c.Post(
			MutationUpdateTransactionTemplateIndex,
			&respC,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", templateID),
			client2.Var("index", 1),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, respA.CreateTransactionTemplate.ID, respC.UpdateTransactionTemplate.ID)
		assert.Equal(t, 1, respC.UpdateTransactionTemplate.Index)
		assert.Equal(t, respB.UpdateTransactionTemplate.Script, respC.UpdateTransactionTemplate.Script)
	})

	t.Run("Update non-existent transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp UpdateTransactionTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			MutationUpdateTransactionTemplateScript,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", badID),
			client2.Var("script", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Get transaction templates for project", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		templateA := createTransactionTemplate(t, c, project)
		templateB := createTransactionTemplate(t, c, project)
		templateC := createTransactionTemplate(t, c, project)

		var resp GetProjectTransactionTemplatesResponse

		err := c.Post(
			QueryGetProjectTransactionTemplates,
			&resp,
			client2.Var("projectId", project.ID),
		)
		require.NoError(t, err)

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
			client2.Var("projectId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Delete transaction template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		template := createTransactionTemplate(t, c, project)

		var resp DeleteTransactionTemplateResponse

		err := c.Post(
			MutationDeleteTransactionTemplate,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", template.ID),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.DeleteTransactionTemplate)
	})

	t.Run("Delete transaction template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		template := createTransactionTemplate(t, c, project)

		var resp DeleteTransactionTemplateResponse

		err := c.Post(
			MutationDeleteTransactionTemplate,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", template.ID),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, template.ID, resp.DeleteTransactionTemplate)
	})
}

func createTransactionTemplate(t *testing.T, c *Client, project Project) TransactionTemplate {
	var resp CreateTransactionTemplateResponse

	err := c.Post(
		MutationCreateTransactionTemplate,
		&resp,
		client2.Var("projectId", project.ID),
		client2.Var("title", "foo"),
		client2.Var("script", "bar"),
		client2.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	return resp.CreateTransactionTemplate
}
