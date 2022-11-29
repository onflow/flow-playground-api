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
	"github.com/dapperlabs/flow-playground-api/server/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"log"
	"net/http"
	"strings"
	"time"

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
	playground "github.com/dapperlabs/flow-playground-api"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/httplog"
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

type SentryConfig struct {
	Dsn              string `default:"https://e8ff473e48aa4962b1a518411489ec5d@o114654.ingest.sentry.io/6398442"`
	Debug            bool   `default:"true"`
	AttachStacktrace bool   `default:"true"`
}

const sessionName = "flow-playground"

var tracer = otel.Tracer("playground-api")

func main() {
	var sentryConf SentryConfig

	if err := envconfig.Process("SENTRY", &sentryConf); err != nil {
		log.Fatal(err)
	}

	semVer := ""
	if build.Version() != nil {
		semVer = build.Version().String()
	}

	err := sentry.Init(sentry.ClientOptions{
		Release:          semVer,
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

	if strings.EqualFold(conf.StorageBackend, storage.PostgreSQL) {
		var datastoreConf storage.DatabaseConfig
		if err := envconfig.Process("FLOW_DB", &datastoreConf); err != nil {
			log.Fatal(err)
		}

		store = storage.NewPostgreSQL(&datastoreConf)
	} else {
		store = storage.NewInMemory()
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

		telemetry.Register()
		initTracer()

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

	utilsHandler := controller.NewUtilsHandler()
	router.Route("/utils", func(r chi.Router) {
		// Add CORS middleware around every request
		// See https://github.com/rs/cors for full option listing
		// test
		r.Use(cors.New(cors.Options{
			AllowedOrigins: conf.AllowedOrigins,
		}).Handler)

		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.HandleFunc("/version", utilsHandler.VersionHandler)
	})

	router.HandleFunc("/ping", ping)
	router.Handle("/metrics", promhttp.Handler())

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

func initTracer() {
	traceExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		log.Fatalf("failed to initialize stdouttrace export pipeline: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSyncer(traceExporter),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}
