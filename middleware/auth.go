package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"

	"github.com/dapperlabs/flow-playground-api/model"
)

type ctxKey string

var (
	projectsCtxKey = ctxKey("projects")
	httpCtxKey     = ctxKey("http")
	sessionCtxKey  = ctxKey("session")
)

type projects struct {
	cookies []*http.Cookie
}

func (p *projects) hasPermission(project *model.InternalProject) bool {
	expectedCookieName := projectCookieKey(project.ID.String())

	for _, cookie := range p.cookies {
		if cookie.Name == expectedCookieName {
			return project.PrivateID.String() == cookie.Value
		}
	}

	return false
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			projectsContext := &projects{
				cookies: r.Cookies(),
			}

			httpContext := HTTPContext{
				W: &w,
				R: r,
			}

			ctx := context.WithValue(r.Context(), httpCtxKey, httpContext)
			ctx = context.WithValue(r.Context(), projectsCtxKey, projectsContext)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func HasProjectPermission(ctx context.Context, project *model.InternalProject) bool {
	session := GetSession(ctx, "FLOW_PLAYGROUND")

	privateID, ok := session.Values[project.ID.String()]
	if !ok {
		return false
	}

	privateIDStr, ok := privateID.(string)
	if !ok {
		return false
	}

	return privateIDStr == project.PrivateID.String()
}

func ProjectCookie(projectID, projectPrivateID string) *http.Cookie {
	return &http.Cookie{
		Name:  projectCookieKey(projectID),
		Value: projectPrivateID,
	}
}

func projectCookieKey(projectID string) string {
	return fmt.Sprintf("proj-%s", projectID)
}

type HTTPContext struct {
	W *http.ResponseWriter
	R *http.Request
}

// GetSession returns a cached session of the given name
func GetSession(ctx context.Context, name string) *sessions.Session {
	store := ctx.Value(sessionCtxKey).(*sessions.CookieStore)
	httpContext := ctx.Value(httpCtxKey).(HTTPContext)

	// Ignore err because a session is always returned even if one doesn't exist
	session, _ := store.Get(httpContext.R, name)

	return session
}

// SaveSession saves the session by writing it to the response
func SaveSession(ctx context.Context, session *sessions.Session) error {
	httpContext := ctx.Value(httpCtxKey).(HTTPContext)

	err := session.Save(httpContext.R, *httpContext.W)

	return err
}

// InjectHTTPMiddleware handles injecting the ResponseWriter and Request structs
// into context so that resolver methods can use these to set and read cookies. It also passes a // CookieStore initialized in `server.go` into context for facilitated cookie handling.
func InjectHTTPMiddleware(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpContext := HTTPContext{
				W: &w,
				R: r,
			}

			ctx := context.WithValue(r.Context(), httpCtxKey, httpContext)
			ctx = context.WithValue(ctx, sessionCtxKey, store)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
