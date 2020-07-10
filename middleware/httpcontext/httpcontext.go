package httpcontext

import (
	"context"
	"net/http"
)

const httpCtxKeyWriter = "http_writer"
const httpCtxKeyRequest = "http_request"

func Request(ctx context.Context) *http.Request {
	return ctx.Value(httpCtxKeyRequest).(*http.Request)
}

func Writer(ctx context.Context) http.ResponseWriter {
	return ctx.Value(httpCtxKeyWriter).(http.ResponseWriter)
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httpCtxKeyWriter, w)
			ctx = context.WithValue(ctx, httpCtxKeyRequest, r)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
