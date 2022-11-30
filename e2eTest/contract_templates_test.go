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

func TestContractTemplates(t *testing.T) {

	t.Run("Create contract template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateContractTemplateResponse

		err := c.Post(
			MutationCreateContractTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.CreateContractTemplate.ID)
	})

	t.Run("Create contract template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateContractTemplateResponse

		err := c.Post(
			MutationCreateContractTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.CreateContractTemplate.ID)
		assert.Equal(t, "foo", resp.CreateContractTemplate.Title)
		assert.Equal(t, "bar", resp.CreateContractTemplate.Script)
	})

	t.Run("Get contract template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA CreateContractTemplateResponse

		err := c.Post(
			MutationCreateContractTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "bar"),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		var respB GetContractTemplateResponse

		err = c.Post(
			QueryGetContractTemplate,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", respA.CreateContractTemplate.ID),
		)
		require.NoError(t, err)

		assert.Equal(t, respA.CreateContractTemplate.ID, respB.ContractTemplate.ID)
		assert.Equal(t, respA.CreateContractTemplate.Script, respB.ContractTemplate.Script)
	})

	t.Run("Get non-existent contract template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp GetContractTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetContractTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Update contract template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA CreateContractTemplateResponse

		err := c.Post(
			MutationCreateContractTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "apple"),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		templateID := respA.CreateContractTemplate.ID

		var respB UpdateContractTemplateResponse

		err = c.Post(
			MutationUpdateContractTemplateScript,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("script", "orange"),
		)
		assert.Error(t, err)
	})

	t.Run("Update contract template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA CreateContractTemplateResponse

		err := c.Post(
			MutationCreateContractTemplate,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("title", "foo"),
			client.Var("script", "apple"),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		templateID := respA.CreateContractTemplate.ID

		var respB UpdateContractTemplateResponse

		err = c.Post(
			MutationUpdateContractTemplateScript,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("script", "orange"),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, respA.CreateContractTemplate.ID, respB.UpdateContractTemplate.ID)
		assert.Equal(t, respA.CreateContractTemplate.Index, respB.UpdateContractTemplate.Index)
		assert.Equal(t, "orange", respB.UpdateContractTemplate.Script)

		var respC struct {
			UpdateContractTemplate struct {
				ID     string
				Script string
				Index  int
			}
		}

		err = c.Post(
			MutationUpdateContractTemplateIndex,
			&respC,
			client.Var("projectId", project.ID),
			client.Var("templateId", templateID),
			client.Var("index", 1),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, respA.CreateContractTemplate.ID, respC.UpdateContractTemplate.ID)
		assert.Equal(t, 1, respC.UpdateContractTemplate.Index)
		assert.Equal(t, respB.UpdateContractTemplate.Script, respC.UpdateContractTemplate.Script)
	})

	t.Run("Update non-existent contract template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp UpdateContractTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			MutationUpdateContractTemplateScript,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", badID),
			client.Var("script", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Get contract templates for project", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		templateA := createContractTemplate(t, c, project)
		templateB := createContractTemplate(t, c, project)
		templateC := createContractTemplate(t, c, project)

		var resp GetProjectContractTemplatesResponse

		err := c.Post(
			QueryGetProjectContractTemplates,
			&resp,
			client.Var("projectId", project.ID),
		)
		require.NoError(t, err)

		assert.Len(t, resp.Project.ContractTemplates, 3)
		assert.Equal(t, templateA.ID, resp.Project.ContractTemplates[0].ID)
		assert.Equal(t, templateB.ID, resp.Project.ContractTemplates[1].ID)
		assert.Equal(t, templateC.ID, resp.Project.ContractTemplates[2].ID)

		assert.Equal(t, 0, resp.Project.ContractTemplates[0].Index)
		assert.Equal(t, 1, resp.Project.ContractTemplates[1].Index)
		assert.Equal(t, 2, resp.Project.ContractTemplates[2].Index)
	})

	t.Run("Get contract templates for non-existent project", func(t *testing.T) {
		c := newClient()

		var resp GetProjectContractTemplatesResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetProjectContractTemplates,
			&resp,
			client.Var("projectId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Delete contract template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		template := createContractTemplate(t, c, project)

		var resp DeleteContractTemplateResponse

		err := c.Post(
			MutationDeleteContractTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", template.ID),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.DeleteContractTemplate)
	})

	t.Run("Delete contract template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		template := createContractTemplate(t, c, project)

		var resp DeleteContractTemplateResponse

		err := c.Post(
			MutationDeleteContractTemplate,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("templateId", template.ID),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, template.ID, resp.DeleteContractTemplate)
	})

}

func createContractTemplate(t *testing.T, c *Client, project Project) ContractTemplate {
	var resp CreateContractTemplateResponse

	err := c.Post(
		MutationCreateContractTemplate,
		&resp,
		client.Var("projectId", project.ID),
		client.Var("title", "foo"),
		client.Var("script", "bar"),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	return resp.CreateContractTemplate
}
