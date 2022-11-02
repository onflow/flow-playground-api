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
	legacyauth "github.com/dapperlabs/flow-playground-api/auth/legacy"
	"github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestAuthentication(t *testing.T) {
	t.Run("Migrate legacy auth", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var respA UpdateProjectResponse

		oldSessionCookie := c.SessionCookie()

		// clear session cookie before making request
		c.ClearSessionCookie()

		err := c.Post(
			MutationUpdateProjectPersist,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("title", project.Title),
			client.Var("description", project.Description),
			client.Var("readme", project.Readme),
			client.Var("persist", true),
			client.AddCookie(legacyauth.MockProjectSessionCookie(project.ID, project.Secret)),
		)
		require.NoError(t, err)

		assert.Equal(t, project.ID, respA.UpdateProject.ID)
		assert.Equal(t, project.Title, respA.UpdateProject.Title)
		assert.Equal(t, project.Description, respA.UpdateProject.Description)
		assert.Equal(t, project.Readme, respA.UpdateProject.Readme)
		assert.True(t, respA.UpdateProject.Persist)

		// a new session cookie should be set
		require.NotNil(t, c.SessionCookie())
		assert.NotEqual(t, oldSessionCookie.Value, c.SessionCookie().Value)

		var respB UpdateProjectResponse

		err = c.Post(
			MutationUpdateProjectPersist,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("title", project.Title),
			client.Var("description", project.Description),
			client.Var("readme", project.Readme),
			client.Var("persist", false),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		// should be able to perform update using new session cookie
		assert.Equal(t, project.ID, respB.UpdateProject.ID)
		assert.Equal(t, project.Title, respB.UpdateProject.Title)
		assert.Equal(t, project.Description, respB.UpdateProject.Description)
		assert.Equal(t, project.Readme, respB.UpdateProject.Readme)
		assert.False(t, respB.UpdateProject.Persist)
	})

	t.Run("Create project with malformed session cookie", func(t *testing.T) {
		c := newClient()

		var respA CreateProjectResponse

		malformedCookie := http.Cookie{
			Name:  sessionName,
			Value: "foo",
		}

		err := c.Post(
			MutationCreateProject,
			&respA,
			client.Var("title", "foo"),
			client.Var("description", "desc"),
			client.Var("readme", "rtfm"),
			client.Var("seed", 42),
			client.Var("numberOfAccounts", initAccounts),
			client.AddCookie(&malformedCookie),
		)
		require.NoError(t, err)

		projectID := respA.CreateProject.ID
		projectTitle := respA.CreateProject.Title
		projectDescription := respA.CreateProject.Description
		projectReadme := respA.CreateProject.Readme

		assert.NotEmpty(t, projectID)
		assert.Equal(t, 42, respA.CreateProject.Seed)

		// session cookie should be overwritten with new value
		assert.NotNil(t, c.SessionCookie())

		var respB UpdateProjectResponse

		err = c.Post(
			MutationUpdateProjectPersist,
			&respB,
			client.Var("projectId", projectID),
			client.Var("title", projectTitle),
			client.Var("description", projectDescription),
			client.Var("readme", projectReadme),
			client.Var("persist", true),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		// should be able to perform update using new session cookie
		assert.Equal(t, projectID, respB.UpdateProject.ID)
		assert.Equal(t, projectTitle, respB.UpdateProject.Title)
		assert.Equal(t, projectDescription, respB.UpdateProject.Description)
		assert.Equal(t, projectReadme, respB.UpdateProject.Readme)
		assert.True(t, respB.UpdateProject.Persist)
	})

	t.Run("Update project with malformed session cookie", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		var resp UpdateProjectResponse

		malformedCookie := http.Cookie{
			Name:  sessionName,
			Value: "foo",
		}

		c.ClearSessionCookie()

		err := c.Post(
			MutationUpdateProjectPersist,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("persist", true),
			client.AddCookie(&malformedCookie),
		)

		assert.Error(t, err)

		// session cookie should not be set
		assert.Nil(t, c.SessionCookie())
	})

	t.Run("Update project with invalid session cookie", func(t *testing.T) {
		c := newClient()

		projectA := createProject(t, c)
		_ = createProject(t, c)

		cookieB := c.SessionCookie()

		var resp UpdateProjectResponse

		err := c.Post(
			MutationUpdateProjectPersist,
			&resp,
			client.Var("projectId", projectA.ID),
			client.Var("persist", true),
			client.AddCookie(cookieB),
		)

		// should not be able to update project A with cookie B
		assert.Error(t, err)
	})
}
