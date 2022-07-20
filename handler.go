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

package playground

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/handler"
	"github.com/getsentry/sentry-go"
	"net/http"
	"runtime/debug"
	"time"
)

func GraphQLHandler(resolver *Resolver, options ...handler.Option) http.HandlerFunc {
	// init crash reporting
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()

	options = append(
		options,
		handler.RecoverFunc(func(ctx context.Context, err interface{}) (userMessage error) {
			return fmt.Errorf("panic: %s\n\n%s", err, string(debug.Stack()))
		}),
	)

	panic("testing panic in handler.go")

	return handler.GraphQL(
		NewExecutableSchema(Config{Resolvers: resolver}),
		options...,
	)
}
