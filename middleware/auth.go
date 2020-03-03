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

const (
	projectsSessionName = "flow-playground"
	sessionMaxAge       = 157680000 // 5 years in seconds
)

func ProjectInSession(ctx context.Context, proj *model.InternalProject) bool {
	session := getSession(ctx, projectsSessionName)

	privateID, ok := session.Values[proj.ID.String()]
	if !ok {
		return false
	}

	privateIDStr, ok := privateID.(string)
	if !ok {
		return false
	}

	return privateIDStr == proj.PrivateID.String()
}

func AddProjectToSession(ctx context.Context, proj *model.InternalProject) error {
	session := getSession(ctx, projectsSessionName)

	// Setting userID cookie value
	session.Values[proj.ID.String()] = proj.PrivateID.String()

	err := saveSession(ctx, session)
	if err != nil {
		return err
	}

	return nil
}

type httpContext struct {
	W *http.ResponseWriter
	R *http.Request
}

// getSession returns a cached session of the given name.
func getSession(ctx context.Context, name string) *sessions.Session {
	store := ctx.Value(sessionCtxKey).(*sessions.CookieStore)
	httpContext := ctx.Value(httpCtxKey).(httpContext)

	fmt.Println("COOKIES", httpContext.R.Cookies())

	// ignore error because a session is always returned even if one does not exist
	session, err := store.Get(httpContext.R, name)

	fmt.Println("GOT SESSION", session.Values)
	fmt.Println("GOT SESSION ERROR", err)

	return session
}

// saveSession saves a session by writing it to the HTTP response.
func saveSession(ctx context.Context, session *sessions.Session) error {
	httpContext := ctx.Value(httpCtxKey).(httpContext)

	session.Options = &sessions.Options{MaxAge: sessionMaxAge}

	err := session.Save(httpContext.R, *httpContext.W)
	if err != nil {
		return err
	}

	return nil
}

// ProjectSessions injects middleware for managing project sessions into an HTTP handler.
//
// Sessions will be stored using the provided sessions.CookieStore instance.
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

const mockSessionKey = "1bbcf50e2e5f3e2d1801db50742f6a97"

func mockCookieStore() *sessions.CookieStore {
	return sessions.NewCookieStore([]byte(mockSessionKey))
}

// MockProjectSessions returns project sessions middleware to be used for testing.
func MockProjectSessions() func(http.Handler) http.Handler {
	return ProjectSessions(mockCookieStore())
}

// MockProjectSessionCookie returns a session cookie that provides access to the given project.
func MockProjectSessionCookie(projectID, projectPrivateID string) *http.Cookie {
	store := mockCookieStore()

	r := &http.Request{}
	w := httptest.NewRecorder()

	session, _ := store.Get(r, projectsSessionName)

	session.Values[projectID] = projectPrivateID

	err := session.Save(r, w)
	if err != nil {
		panic(err)
	}

	return w.Result().Cookies()[0]
}
