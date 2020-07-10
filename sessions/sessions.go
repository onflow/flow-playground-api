package sessions

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"

	"github.com/dapperlabs/flow-playground-api/middleware/httpcontext"
)

const sessionCtxKey = "session"

// Middleware injects middleware for managing project sessions into an HTTP handler.
//
// Sessions are stored using the provided sessions.CookieStore instance.
func Middleware(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), sessionCtxKey, store)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// Get returns the session with the given name, or creates one if it does not exist.
func Get(ctx context.Context, name string) *sessions.Session {
	store := ctx.Value(sessionCtxKey).(*sessions.CookieStore)

	// ignore error because a session is always returned even if one does not exist
	session, _ := store.Get(httpcontext.Request(ctx), name)

	return session
}

// Save saves a session by writing it to the HTTP response.
func Save(ctx context.Context, session *sessions.Session) error {
	err := session.Save(
		httpcontext.Request(ctx),
		httpcontext.Writer(ctx),
	)
	if err != nil {
		return err
	}

	return nil
}
