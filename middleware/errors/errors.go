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

	"github.com/sirupsen/logrus"

	"github.com/99designs/gqlgen/graphql"
)

type gqlErrCtxKeyType string

var (
	errLoggerFieldsCtxKey = gqlErrCtxKeyType("error-logger-fields")
)

// Middleware is a catch-all middleware for GQL request errors.
func Middleware(entry *logrus.Entry) graphql.RequestMiddleware {
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
