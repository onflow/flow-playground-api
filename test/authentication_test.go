package test

import (
	legacyauth "github.com/dapperlabs/flow-playground-api/auth/legacy"
	client2 "github.com/dapperlabs/flow-playground-api/test/client"
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
			client2.Var("projectId", project.ID),
			client2.Var("title", project.Title),
			client2.Var("description", project.Description),
			client2.Var("readme", project.Readme),
			client2.Var("persist", true),
			client2.AddCookie(legacyauth.MockProjectSessionCookie(project.ID, project.Secret)),
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
			client2.Var("projectId", project.ID),
			client2.Var("title", project.Title),
			client2.Var("description", project.Description),
			client2.Var("readme", project.Readme),
			client2.Var("persist", false),
			client2.AddCookie(c.SessionCookie()),
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
			client2.Var("title", "foo"),
			client2.Var("description", "desc"),
			client2.Var("readme", "rtfm"),
			client2.Var("seed", 42),
			client2.Var("numberOfAccounts", initAccounts),
			client2.AddCookie(&malformedCookie),
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
			client2.Var("projectId", projectID),
			client2.Var("title", projectTitle),
			client2.Var("description", projectDescription),
			client2.Var("readme", projectReadme),
			client2.Var("persist", true),
			client2.AddCookie(c.SessionCookie()),
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
			client2.Var("projectId", project.ID),
			client2.Var("persist", true),
			client2.AddCookie(&malformedCookie),
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
			client2.Var("projectId", projectA.ID),
			client2.Var("persist", true),
			client2.AddCookie(cookieB),
		)

		// should not be able to update project A with cookie B
		assert.Error(t, err)
	})
}
