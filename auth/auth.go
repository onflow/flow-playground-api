package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dapperlabs/flow-playground-api/model"
)

var projectsCtxKey = &contextKey{"projects"}

type contextKey struct {
	name string
}

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
			projects := &projects{
				cookies: r.Cookies(),
			}

			ctx := context.WithValue(r.Context(), projectsCtxKey, projects)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func HasProjectPermission(ctx context.Context, project *model.InternalProject) bool {
	projects, _ := ctx.Value(projectsCtxKey).(*projects)
	return projects.hasPermission(project)
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
