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
	client2 "github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestScriptTemplates(t *testing.T) {
	t.Run("Create script template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateScriptTemplateResponse

		err := c.Post(
			MutationCreateScriptTemplate,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("title", "foo"),
			client2.Var("script", "bar"),
		)

		assert.Error(t, err)
		assert.Empty(t, resp.CreateScriptTemplate.ID)
	})

	t.Run("Create script template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp CreateScriptTemplateResponse

		err := c.Post(
			MutationCreateScriptTemplate,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("title", "foo"),
			client2.Var("script", "bar"),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.CreateScriptTemplate.ID)
		assert.Equal(t, "foo", resp.CreateScriptTemplate.Title)
		assert.Equal(t, "bar", resp.CreateScriptTemplate.Script)
	})

	t.Run("Get script template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA CreateScriptTemplateResponse

		err := c.Post(
			MutationCreateScriptTemplate,
			&respA,
			client2.Var("projectId", project.ID),
			client2.Var("title", "foo"),
			client2.Var("script", "bar"),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		var respB GetScriptTemplateResponse

		err = c.Post(
			QueryGetScriptTemplate,
			&respB,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", respA.CreateScriptTemplate.ID),
		)
		require.NoError(t, err)

		assert.Equal(t, respA.CreateScriptTemplate.ID, respB.ScriptTemplate.ID)
		assert.Equal(t, respA.CreateScriptTemplate.Script, respB.ScriptTemplate.Script)
	})

	t.Run("Get non-existent script template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp GetScriptTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			QueryGetScriptTemplate,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Update script template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA CreateScriptTemplateResponse

		err := c.Post(
			MutationCreateScriptTemplate,
			&respA,
			client2.Var("projectId", project.ID),
			client2.Var("title", "foo"),
			client2.Var("script", "apple"),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		templateID := respA.CreateScriptTemplate.ID

		var respB UpdateScriptTemplateResponse

		err = c.Post(
			MutationUpdateScriptTemplateScript,
			&respB,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", templateID),
			client2.Var("script", "orange"),
		)
		assert.Error(t, err)
	})

	t.Run("Update script template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA CreateScriptTemplateResponse

		err := c.Post(
			MutationCreateScriptTemplate,
			&respA,
			client2.Var("projectId", project.ID),
			client2.Var("title", "foo"),
			client2.Var("script", "apple"),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		templateID := respA.CreateScriptTemplate.ID

		var respB UpdateScriptTemplateResponse

		err = c.Post(
			MutationUpdateScriptTemplateScript,
			&respB,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", templateID),
			client2.Var("script", "orange"),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, respA.CreateScriptTemplate.ID, respB.UpdateScriptTemplate.ID)
		assert.Equal(t, respA.CreateScriptTemplate.Index, respB.UpdateScriptTemplate.Index)
		assert.Equal(t, "orange", respB.UpdateScriptTemplate.Script)

		var respC UpdateScriptTemplateResponse

		err = c.Post(
			MutationUpdateScriptTemplateIndex,
			&respC,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", templateID),
			client2.Var("index", 1),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, respA.CreateScriptTemplate.ID, respC.UpdateScriptTemplate.ID)
		assert.Equal(t, 1, respC.UpdateScriptTemplate.Index)
		assert.Equal(t, respB.UpdateScriptTemplate.Script, respC.UpdateScriptTemplate.Script)
	})

	t.Run("Update non-existent script template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp UpdateScriptTemplateResponse

		badID := uuid.New().String()

		err := c.Post(
			MutationUpdateScriptTemplateScript,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", badID),
			client2.Var("script", "bar"),
		)

		assert.Error(t, err)
	})

	t.Run("Get script templates for project", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		templateIDA := createScriptTemplate(t, c, project)
		templateIDB := createScriptTemplate(t, c, project)
		templateIDC := createScriptTemplate(t, c, project)

		var resp GetProjectScriptTemplatesResponse

		err := c.Post(
			QueryGetProjectScriptTemplates,
			&resp,
			client2.Var("projectId", project.ID),
		)
		require.NoError(t, err)

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
			client2.Var("projectId", badID),
		)

		assert.Error(t, err)
	})

	t.Run("Delete script template without permission", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		templateID := createScriptTemplate(t, c, project)

		var resp DeleteScriptTemplateResponse

		err := c.Post(
			MutationDeleteScriptTemplate,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", templateID),
		)

		assert.Error(t, err)
	})

	t.Run("Delete script template", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		templateID := createScriptTemplate(t, c, project)

		var resp DeleteScriptTemplateResponse

		err := c.Post(
			MutationDeleteScriptTemplate,
			&resp,
			client2.Var("projectId", project.ID),
			client2.Var("templateId", templateID),
			client2.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		assert.Equal(t, templateID, resp.DeleteScriptTemplate)
	})
}

func createScriptTemplate(t *testing.T, c *Client, project Project) string {
	var resp CreateScriptTemplateResponse

	err := c.Post(
		MutationCreateScriptTemplate,
		&resp,
		client2.Var("projectId", project.ID),
		client2.Var("title", "foo"),
		client2.Var("script", "bar"),
		client2.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	return resp.CreateScriptTemplate.ID
}
