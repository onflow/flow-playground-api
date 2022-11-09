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
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
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
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
			client.AddCookie(c.SessionCookie()),
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
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		var respB GetTransactionTemplateResponse

		err = c.Post(
			QueryGetTransactionTemplate,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", respA.CreateTransactionTemplate.ID),
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
			client.Var("projectId", project.ID),
			client.Var("templateId", badID),
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
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "apple"),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		templateID := respA.CreateTransactionTemplate.ID

		var respB UpdateTransactionTemplateResponse

		err = c.Post(
			MutationUpdateTransactionTemplateScript,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("script", "orange"),
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
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "apple"),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		templateID := respA.CreateTransactionTemplate.ID

		var respB UpdateTransactionTemplateResponse

		err = c.Post(
			MutationUpdateTransactionTemplateScript,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("script", "orange"),
			client.AddCookie(c.SessionCookie()),
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
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("index", 1),
			client.AddCookie(c.SessionCookie()),
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
			client.Var("projectId", project.ID),
			client.Var("templateId", badID),
			client.Var("script", "bar"),
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
			client.Var("projectId", project.ID),
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
			client.Var("projectId", badID),
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
			client.Var("projectId", project.ID),
			client.Var("templateId", template.ID),
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
			client.Var("projectId", project.ID),
			client.Var("templateId", template.ID),
			client.AddCookie(c.SessionCookie()),
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
		client.Var("projectId", project.ID),
		client.Var("title", "foo"),
		client.Var("script", "bar"),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	return resp.CreateTransactionTemplate
}
