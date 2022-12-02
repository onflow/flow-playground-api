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

package playground

import (
	"context"
	"fmt"
	"github.com/dapperlabs/flow-playground-api/server/telemetry"
	"github.com/getsentry/sentry-go"
	"github.com/ravilushqa/otelgqlgen"
	"net/http"
	"runtime/debug"

	"github.com/99designs/gqlgen/graphql"
	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
)

func GraphQLHandler(resolver *Resolver, middlewares ...graphql.ResponseMiddleware) http.HandlerFunc {
	srv := gqlHandler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver}))

	srv.Use(telemetry.MetricTracer{})
	srv.Use(otelgqlgen.Middleware())

	for _, middleware := range middlewares {
		srv.AroundResponses(middleware)
	}

	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) (userMessage error) {
		sentry.CaptureException(fmt.Errorf("panic: %v, stack: %s", err, string(debug.Stack())))
		return fmt.Errorf("panic: %s\n\n%s", err, string(debug.Stack()))
	})

	return srv.ServeHTTP
}
