package middleware

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/99designs/gqlgen/graphql"
)

type gqlErrCtxKeyType string

var (
	errLoggerFieldsCtxKey = gqlErrCtxKeyType("error-logger-fields")
)

// ErrorMiddleware the catch all for GLQ request errors
func ErrorMiddleware(entry *logrus.Entry) graphql.RequestMiddleware {
	return func(ctx context.Context, next func(ctx context.Context) []byte) []byte {
		debugFields := logrus.Fields{}
		ctx = context.WithValue(ctx, errLoggerFieldsCtxKey, debugFields)
		res := next(ctx)
		reqCtx := graphql.GetRequestContext(ctx)

		for _, err := range reqCtx.Errors {
			contextEntry := entry.
				WithFields(debugFields)

			if cause := err.Extensions["cause"]; cause != nil {
				contextEntry.
					WithError(cause.(error)).
					Error("GQL Request Server Error")
			} else if err != nil {
				contextEntry.WithError(err).Warnf("GQL Request Client Error: %v err = %+v", err.Extensions["general_error"], err)
			}
		}

		return res
	}
}
