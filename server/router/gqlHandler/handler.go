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

package gqlHandler

import (
	"context"
	"fmt"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/server/router/gqlHandler/middleware/errors"
	"github.com/dapperlabs/flow-playground-api/server/router/gqlHandler/resolver"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"net/http"
	"runtime/debug"
	"time"

	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
)

var flowResolver *resolver.Resolver = nil

func GraphQLHandler() http.HandlerFunc {
	flowResolver = resolver.NewResolver()
	srv := gqlHandler.NewDefaultServer(playground.NewExecutableSchema(playground.Config{Resolvers: flowResolver}))

	// Add middleware
	logger := logrus.StandardLogger()
	logger.Formatter = stackdriver.NewFormatter(stackdriver.WithService("flow-playground"))
	entry := logrus.NewEntry(logger)

	// Create a new hub for this subroutine and bind current client and handle to scope
	localHub := sentry.CurrentHub().Clone()
	localHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("query", "/query")
	})

	defer func() {
		err := recover()
		if err != nil {
			localHub.Recover(err)
			sentry.Flush(time.Second * 5)
		}
	}()
	srv.AroundResponses(errors.Middleware(entry, localHub))

	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) (userMessage error) {
		return fmt.Errorf("panic: %s\n\n%s", err, string(debug.Stack()))
	})

	return srv.ServeHTTP
}

func GetResolver() *resolver.Resolver {
	return flowResolver
}
