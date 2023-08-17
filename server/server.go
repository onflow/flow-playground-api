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
	"context"
	"fmt"
	"github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/server/config"
	"github.com/dapperlabs/flow-playground-api/server/ping"
	"github.com/dapperlabs/flow-playground-api/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/sdk/trace"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/httplog"

	"github.com/dapperlabs/flow-playground-api/auth"
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/build"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/middleware/errors"
	"github.com/dapperlabs/flow-playground-api/middleware/httpcontext"
	"github.com/dapperlabs/flow-playground-api/middleware/monitoring"
	"github.com/dapperlabs/flow-playground-api/middleware/sessions"
	"github.com/dapperlabs/flow-playground-api/storage"

	gqlPlayground "github.com/99designs/gqlgen/graphql/playground"
	"github.com/Masterminds/semver"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	gsessions "github.com/gorilla/sessions"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

const sessionName = "flow-playground"

func main() {
	ctx := context.Background()
	semVer := ""
	if build.Version() != nil {
		semVer = build.Version().String()
	}

	platform := config.Platform()

	if platform != config.Local {
		var sentryConf = config.Sentry()
		err := sentry.Init(sentry.ClientOptions{
			Release:          semVer,
			Dsn:              sentryConf.Dsn,
			Debug:            sentryConf.Debug,
			AttachStacktrace: sentryConf.AttachStacktrace,
			Environment:      string(platform),
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				if hint.Context != nil {
					if sentryLevel, ok := errors.SentryLogLevel(hint.Context); ok {
						event.Level = sentryLevel
					}
				}
				return event
			},
		})

		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}

		defer sentry.Flush(2 * time.Second)
		defer sentry.Recover()
	}

	var conf = config.Playground()

	var store storage.Store

	if strings.EqualFold(conf.StorageBackend, storage.PostgreSQL) {
		var databaseConf = config.Database()
		store = storage.NewPostgreSQL(&databaseConf)
	} else {
		store = storage.NewSqlite()
	}

	const initAccountsNumber = 5

	sessionAuthKey := []byte(conf.SessionAuthKey)
	authenticator := auth.NewAuthenticator(store, sessionName)
	chain := blockchain.NewProjects(store, initAccountsNumber)
	resolver := playground.NewResolver(build.Version(), store, authenticator, chain)

	router := chi.NewRouter()
	router.Use(monitoring.Middleware())

	if conf.Debug {
		logger := httplog.NewLogger("playground-api", httplog.Options{Concise: true, JSON: true})
		router.Use(httplog.RequestLogger(logger))
		router.Handle("/", gqlPlayground.Handler("GraphQL playground", "/query"))
	}

	if config.Telemetry().TracingEnabled {
		tp, err := telemetry.NewProvider(ctx,
			"playground-api",
			config.Telemetry().TracingCollectorEndpoint,
			trace.ParentBased(trace.AlwaysSample()),
		)
		if err != nil {
			log.Fatal("failed to setup telemetry provider", err)
		}
		defer telemetry.CleanupTraceProvider(ctx, tp)
	}

	defer telemetry.UnRegisterMetrics()

	logger := logrus.StandardLogger()
	logger.Formatter = stackdriver.NewFormatter(stackdriver.WithService("flow-playground"))
	entry := logrus.NewEntry(logger)

	router.Route("/query", func(r chi.Router) {
		// Add CORS middleware around every request
		// See https://github.com/rs/cors for full option listing
		r.Use(cors.New(cors.Options{
			AllowedOrigins:   conf.AllowedOrigins,
			AllowCredentials: true,
		}).Handler)

		cookieStore := gsessions.NewCookieStore(sessionAuthKey)
		cookieStore.MaxAge(int(conf.SessionMaxAge.Seconds()))

		cookieStore.Options.Secure = conf.SessionCookiesSecure
		cookieStore.Options.HttpOnly = conf.SessionCookiesHTTPOnly

		if conf.SessionCookiesSameSiteNone {
			cookieStore.Options.SameSite = http.SameSiteNoneMode
		}

		sessions.SetCookieStore(cookieStore)

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

	embedsHandler := controller.NewEmbedsHandler(store, conf.PlaygroundBaseURL)
	router.Handle("/embed", embedsHandler)

	err := ping.SetPingHandlers(store.Ping)
	if err != nil {
		log.Fatal(err)
	}

	telemetry.SetStaleProjectScanner(store.GetStaleProjects)
	telemetry.SetTotalProjectCounter(store.TotalProjectCount)

	router.HandleFunc("/ping", ping.Ping)
	router.Handle("/metrics", promhttp.Handler())

	logStartMessage(build.Version())

	log.Printf("Connect to http://localhost:%d/ for GraphQL playground", conf.Port)
	log.Print("Allowed origins", conf.AllowedOrigins)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), router))
}

func logStartMessage(version *semver.Version) {
	if version != nil {
		log.Printf("Starting Playground API (Version %s)", version)
	} else {
		log.Print("Starting Playground API")
	}
}
