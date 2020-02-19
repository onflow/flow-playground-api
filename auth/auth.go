package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

var projectsCtxKey = &contextKey{"projects"}

type contextKey struct {
	name string
}

type projects struct {
	store   storage.Store
	cookies []*http.Cookie
}

func (p *projects) hasPermission(projectID uuid.UUID) bool {
	expectedCookieName := fmt.Sprintf("proj-%s", projectID.String())
	for _, cookie := range p.cookies {
		if cookie.Name == expectedCookieName {
			var proj model.InternalProject

			err := p.store.GetProject(projectID, &proj)
			if err != nil {
				// TODO: handle this differently?
				return false
			}

			// TODO: okay to compare strings here?
			return proj.PrivateID.String() == cookie.Value
		}
	}

	return false
}

func Middleware(store storage.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			projects := &projects{
				store:   store,
				cookies: r.Cookies(),
			}

			ctx := context.WithValue(r.Context(), projectsCtxKey, projects)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func HasProjectPermission(ctx context.Context, projectID uuid.UUID) bool {
	projects, _ := ctx.Value(projectsCtxKey).(*projects)
	return projects.hasPermission(projectID)
}

func ProjectCookie(projectID, projectPrivateID string) *http.Cookie {
	return &http.Cookie{
		Name:  fmt.Sprintf("proj-%s", projectID),
		Value: projectPrivateID,
	}
}
