/*
 * Flow Playground
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package errors

import (
	"context"
	"github.com/getsentry/sentry-go"

	"github.com/sirupsen/logrus"

	"github.com/99designs/gqlgen/graphql"
)

type errCtxKeyType string

var (
	errLoggerFieldsCtxKey = errCtxKeyType("error-logger-fields")
	sentryLevelCtxKey     = errCtxKeyType("sentry-level")
)

// SentryLogLevel is a helper method that gets the log level from the context.
func SentryLogLevel(ctx context.Context) (sentry.Level, bool) {
	sentryLevel, ok := ctx.Value(sentryLevelCtxKey).(sentry.Level)
	return sentryLevel, ok
}

// Middleware is a catch-all middleware for GQL request errors.
func Middleware(entry *logrus.Entry, localHub *sentry.Hub) graphql.RequestMiddleware {
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
				sentryCtx := context.WithValue(ctx, sentryLevelCtxKey, sentry.LevelError)
				localHub.RecoverWithContext(sentryCtx, err)
			} else if err != nil {
				contextEntry.WithError(err).Warnf("GQL Request Client Error: %v err = %+v", err.Extensions["general_error"], err)
				sentryCtx := context.WithValue(ctx, sentryLevelCtxKey, sentry.LevelWarning)
				localHub.RecoverWithContext(sentryCtx, err)
			}
		}

		return res
	}
}
