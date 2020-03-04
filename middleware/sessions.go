package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/sessions"

	"github.com/dapperlabs/flow-playground-api/model"
)

type ctxKey string

var (
	httpCtxKey    = ctxKey("http")
	sessionCtxKey = ctxKey("session")
)

const projectSecretKeyName = "project-secret"

type httpContext struct {
	W *http.ResponseWriter
	R *http.Request
}

// ProjectSessions injects middleware for managing project sessions into an HTTP handler.
//
// Sessions are stored using the provided sessions.CookieStore instance.
func ProjectSessions(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpContext := httpContext{W: &w, R: r}

			ctx := context.WithValue(r.Context(), httpCtxKey, httpContext)
			ctx = context.WithValue(ctx, sessionCtxKey, store)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// ProjectInSession returns true if the given project is authorized in the current session.
//
// A project is authorized in a session if the session contains a reference to the
// project's secret.
func ProjectInSession(ctx context.Context, proj *model.InternalProject) bool {
	session := getSession(ctx, getProjectSessionName(proj))

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

// AddProjectToSession adds the given project's secret to the current session.
//
// This function re-saves the session and updates the session cookie with a new max age.
func AddProjectToSession(ctx context.Context, proj *model.InternalProject) error {
	session := getSession(ctx, getProjectSessionName(proj))

	session.Values[projectSecretKeyName] = proj.Secret.String()

	err := saveSession(ctx, session)
	if err != nil {
		return err
	}

	return nil
}

// getSession returns the session with the given name, or creates one if it does not exist.
func getSession(ctx context.Context, name string) *sessions.Session {
	store := ctx.Value(sessionCtxKey).(*sessions.CookieStore)
	httpContext := ctx.Value(httpCtxKey).(httpContext)

	// ignore error because a session is always returned even if one does not exist
	session, _ := store.Get(httpContext.R, name)

	return session
}

// saveSession saves a session by writing it to the HTTP response.
func saveSession(ctx context.Context, session *sessions.Session) error {
	httpContext := ctx.Value(httpCtxKey).(httpContext)

	err := session.Save(httpContext.R, *httpContext.W)
	if err != nil {
		return err
	}

	return nil
}

func getProjectSessionName(proj *model.InternalProject) string {
	return getProjectSessionNameFromString(proj.ID.String())
}

func getProjectSessionNameFromString(projectID string) string {
	return fmt.Sprintf("flow-%s", projectID)
}

const mockSessionAuthenticationKey = "1bbcf50e2e5f3e2d1801db50742f6a97"

func mockCookieStore() *sessions.CookieStore {
	return sessions.NewCookieStore([]byte(mockSessionAuthenticationKey))
}

// MockProjectSessions returns project sessions middleware to be used for testing.
func MockProjectSessions() func(http.Handler) http.Handler {
	return ProjectSessions(mockCookieStore())
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
