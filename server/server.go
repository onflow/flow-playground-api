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

package main

import (
	"fmt"
	"github.com/dapperlabs/flow-playground-api/server/config"
	"github.com/dapperlabs/flow-playground-api/server/storage"
	sentryWrapper "github.com/dapperlabs/flow-playground-api/server/telemetry/sentry"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/httplog"

	playground "github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/auth"
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/build"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/middleware/errors"
	"github.com/dapperlabs/flow-playground-api/middleware/httpcontext"
	"github.com/dapperlabs/flow-playground-api/middleware/monitoring"
	"github.com/dapperlabs/flow-playground-api/middleware/sessions"

	gqlPlayground "github.com/99designs/gqlgen/graphql/playground"
	"github.com/Masterminds/semver"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/golang/groupcache/lru"
	gsessions "github.com/gorilla/sessions"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

const sessionName = "flow-playground"

func main() {
	sentryWrapper.InitializeSentry()
	defer sentryWrapper.Cleanup()

	const initAccountsNumber = 5

	sessionAuthKey := []byte(config.GetConfig().SessionAuthKey)
	authenticator := auth.NewAuthenticator(storage.GetStorage(), sessionName)
	chain := blockchain.NewProjects(storage.GetStorage(), lru.New(128), initAccountsNumber)
	resolver := playground.NewResolver(build.Version(), storage.GetStorage(), authenticator, chain)

	router := chi.NewRouter()
	router.Use(monitoring.Middleware())

	if config.GetConfig().Debug {
		logger := httplog.NewLogger("playground-api", httplog.Options{Concise: true})
		router.Use(httplog.RequestLogger(logger))
		router.Handle("/", gqlPlayground.Handler("GraphQL playground", "/query"))
	}

	logger := logrus.StandardLogger()
	logger.Formatter = stackdriver.NewFormatter(stackdriver.WithService("flow-playground"))
	entry := logrus.NewEntry(logger)

	router.Route("/query", func(r chi.Router) {
		// Add CORS middleware around every request
		// See https://github.com/rs/cors for full option listing
		r.Use(cors.New(cors.Options{
			AllowedOrigins:   config.GetConfig().AllowedOrigins,
			AllowCredentials: true,
		}).Handler)

		cookieStore := gsessions.NewCookieStore(sessionAuthKey)
		cookieStore.MaxAge(int(config.GetConfig().SessionMaxAge.Seconds()))

		cookieStore.Options.Secure = config.GetConfig().SessionCookiesSecure
		cookieStore.Options.HttpOnly = config.GetConfig().SessionCookiesHTTPOnly

		if config.GetConfig().SessionCookiesSameSiteNone {
			cookieStore.Options.SameSite = http.SameSiteNoneMode
		}

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

		r.Use(httpcontext.Middleware())
		r.Use(sessions.Middleware(cookieStore))
		r.Use(monitoring.Middleware())

		r.Handle(
			"/",
			playground.GraphQLHandler(
				resolver,
				errors.Middleware(entry, localHub),
			),
		)

	})

	embedsHandler := controller.NewEmbedsHandler(storage.GetStorage(), config.GetConfig().PlaygroundBaseURL)
	router.Handle("/embed", embedsHandler)

	utilsHandler := controller.NewUtilsHandler()
	router.Route("/utils", func(r chi.Router) {
		// Add CORS middleware around every request
		// See https://github.com/rs/cors for full option listing
		r.Use(cors.New(cors.Options{
			AllowedOrigins: config.GetConfig().AllowedOrigins,
		}).Handler)

		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.HandleFunc("/version", utilsHandler.VersionHandler)
	})

	router.HandleFunc("/ping", ping)

	logStartMessage(build.Version())

	log.Printf("Connect to http://localhost:%d/ for GraphQL playground", config.GetConfig().Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.GetConfig().Port), router))
}

func ping(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
	_, _ = w.Write([]byte("ok"))
}

func logStartMessage(version *semver.Version) {
	if version != nil {
		log.Printf("Starting Playground API (Version %s)", version)
	} else {
		log.Print("Starting Playground API")
	}
}
