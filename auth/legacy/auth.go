package legacy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	gorillasessions "github.com/gorilla/sessions"

	"github.com/dapperlabs/flow-playground-api/middleware/sessions"
	"github.com/dapperlabs/flow-playground-api/model"
)

const projectSecretKeyName = "project-secret"

// ProjectInSession returns true if the given project is authorized in the current session.
//
// A project is authorized in a session if the session contains a reference to the
// project's secret.
func ProjectInSession(ctx context.Context, proj *model.InternalProject) bool {
	session := sessions.Get(ctx, getProjectSessionName(proj))

	secret, ok := session.Values[projectSecretKeyName]
	if !ok {
		return false
	}

	secretStr, ok := secret.(string)
	if !ok {
		return false
	}

	return secretStr == proj.Secret.String()
}

func getProjectSessionName(proj *model.InternalProject) string {
	return getProjectSessionNameFromString(proj.ID.String())
}

func getProjectSessionNameFromString(projectID string) string {
	return fmt.Sprintf("flow-%s", projectID)
}

const mockSessionAuthenticationKey = "1bbcf50e2e5f3e2d1801db50742f6a97"

func mockCookieStore() *gorillasessions.CookieStore {
	return gorillasessions.NewCookieStore([]byte(mockSessionAuthenticationKey))
}

// MockProjectSessions returns project sessions middleware to be used for testing.
func MockProjectSessions() func(http.Handler) http.Handler {
	return sessions.Middleware(mockCookieStore())
}

// MockProjectSessionCookie returns a session cookie that provides access to the given project.
func MockProjectSessionCookie(projectID, secret string) *http.Cookie {
	store := mockCookieStore()

	r := &http.Request{}
	w := httptest.NewRecorder()

	session, _ := store.Get(r, getProjectSessionNameFromString(projectID))

	session.Values[projectSecretKeyName] = secret

	err := session.Save(r, w)
	if err != nil {
		panic(err)
	}

	return w.Result().Cookies()[0]
}
