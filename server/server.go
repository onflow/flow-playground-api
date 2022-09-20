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
	"github.com/dapperlabs/flow-playground-api/telemetry"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/httplog"

	playground "github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/auth"
	"github.com/dapperlabs/flow-playground-api/build"
	"github.com/dapperlabs/flow-playground-api/middleware/errors"
	"github.com/dapperlabs/flow-playground-api/middleware/httpcontext"
	"github.com/dapperlabs/flow-playground-api/middleware/sessions"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/storage/datastore"
	"github.com/dapperlabs/flow-playground-api/storage/memory"

	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/middleware/monitoring"

	gqlPlayground "github.com/99designs/gqlgen/graphql/playground"
	"github.com/Masterminds/semver"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	gsessions "github.com/gorilla/sessions"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Port                       int           `default:"8080"`
	Debug                      bool          `default:"false"`
	AllowedOrigins             []string      `default:"http://localhost:3000"`
	SessionAuthKey             string        `default:"428ce08c21b93e5f0eca24fbeb0c7673"`
	SessionMaxAge              time.Duration `default:"157680000s"`
	SessionCookiesSecure       bool          `default:"true"`
	SessionCookiesHTTPOnly     bool          `default:"true"`
	SessionCookiesSameSiteNone bool          `default:"false"`
	LedgerCacheSize            int           `default:"128"`
	PlaygroundBaseURL          string        `default:"http://localhost:3000"`
	StorageBackend             string
}

type DatastoreConfig struct {
	GCPProjectID string        `required:"true"`
	Timeout      time.Duration `default:"5s"`
}

type SentryConfig struct {
	Dsn              string `default:"https://e8ff473e48aa4962b1a518411489ec5d@o114654.ingest.sentry.io/6398442"`
	Debug            bool   `default:"true"`
	AttachStacktrace bool   `default:"true"`
}

const sessionName = "flow-playground"

func main() {
	var sentryConf SentryConfig

	if err := envconfig.Process("SENTRY", &sentryConf); err != nil {
		log.Fatal(err)
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryConf.Dsn,
		Debug:            sentryConf.Debug,
		AttachStacktrace: sentryConf.AttachStacktrace,
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

	var conf Config

	if err := envconfig.Process("FLOW", &conf); err != nil {
		log.Fatal(err)
	}

	var store storage.Store

	if strings.EqualFold(conf.StorageBackend, "datastore") {
		var datastoreConf DatastoreConfig

		if err := envconfig.Process("FLOW_DATASTORE", &datastoreConf); err != nil {
			log.Fatal(err)
		}

		var err error
		store, err = datastore.NewDatastore(
			context.Background(),
			&datastore.Config{
				DatastoreProjectID: datastoreConf.GCPProjectID,
				DatastoreTimeout:   datastoreConf.Timeout,
			},
		)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		store = memory.NewStore()
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

	logger := logrus.StandardLogger()
	logger.Formatter = stackdriver.NewFormatter(stackdriver.WithService("flow-playground"))
	entry := logrus.NewEntry(logger)

	telemetry.DebugLog("server startup")

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

		telemetry.DebugLog("GraphQL request")
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

	utilsHandler := controller.NewUtilsHandler()
	router.Route("/utils", func(r chi.Router) {
		// Add CORS middleware around every request
		// See https://github.com/rs/cors for full option listing
		r.Use(cors.New(cors.Options{
			AllowedOrigins: conf.AllowedOrigins,
		}).Handler)

		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.HandleFunc("/version", utilsHandler.VersionHandler)
	})

	router.HandleFunc("/ping", ping)

	logStartMessage(build.Version())

	log.Printf("Connect to http://localhost:%d/ for GraphQL playground", conf.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), router))
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
