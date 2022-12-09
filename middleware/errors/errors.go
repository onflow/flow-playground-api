/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/dapperlabs/flow-playground-api/telemetry"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

var ServerErr = errors.New("something went wrong, we are looking into the issue")
var GraphqlErr = errors.New("invalid graphql request")

type UserError struct {
	msg string
}

func NewUserError(msg string) *UserError {
	return &UserError{msg}
}

func (i *UserError) Error() string {
	return fmt.Sprintf("user error: %s", i.msg)
}

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
func Middleware(entry *logrus.Entry, localHub *sentry.Hub) graphql.ResponseMiddleware {
	return func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		debugFields := logrus.Fields{}
		ctx = context.WithValue(ctx, errLoggerFieldsCtxKey, debugFields)
		res := next(ctx)
		errList := graphql.GetErrors(ctx)

		for i, err := range errList {
			contextEntry := entry.
				WithFields(debugFields)

			if code := err.Extensions["code"]; code != nil {
				res.Errors[i].Message = GraphqlErr.Error()
			} else if err != nil {
				var userErr *UserError
				if errors.As(err, &userErr) {
					telemetry.UserErrorCounter.Inc()
					res.Extensions["code"] = "BAD_REQUEST"
				} else {
					sentry.CaptureException(err)
					telemetry.ServerErrorCounter.Inc()
					res.Errors[i].Message = ServerErr.Error()
					res.Extensions["code"] = "INTERNAL_SERVER_ERROR"
				}

				contextEntry.
					WithError(err).
					Warnf("GQL Request Client Error: %v err = %+v", err.Extensions["general_error"], err)
			}
		}

		return res
	}
}
