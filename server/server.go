package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen-contrib/prometheus"
	"github.com/99designs/gqlgen/handler"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

	playground "github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/middleware"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/storage/datastore"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
	"github.com/dapperlabs/flow-playground-api/vm"
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
	StorageBackend             string
}

type DatastoreConfig struct {
	GCPProjectID string        `required:"true"`
	Timeout      time.Duration `default:"5s"`
}

func main() {
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
			// If datastore is expected, panic when we can't init
			panic(err)
		}
	} else {
		store = memory.NewStore()
	}

	computer, err := vm.NewComputer(conf.LedgerCacheSize)
	if err != nil {
		panic(err)
	}

	resolver := playground.NewResolver(store, computer)

	// Register gql metrics
	prometheus.Register()

	router := chi.NewRouter()

	if conf.Debug {
		router.Handle("/", handler.Playground("GraphQL playground", "/query"))
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
			Debug:            conf.Debug,
		}).Handler)

		cookieStore := sessions.NewCookieStore([]byte(conf.SessionAuthKey))
		cookieStore.MaxAge(int(conf.SessionMaxAge.Seconds()))

		cookieStore.Options.Secure = conf.SessionCookiesSecure
		cookieStore.Options.HttpOnly = conf.SessionCookiesHTTPOnly

		if conf.SessionCookiesSameSiteNone {
			cookieStore.Options.SameSite = http.SameSiteNoneMode
		}

		r.Use(middleware.ProjectSessions(cookieStore))

		r.Handle("/", handler.GraphQL(
			playground.NewExecutableSchema(playground.Config{Resolvers: resolver}),
			handler.RequestMiddleware(middleware.ErrorMiddleware(entry)),
			handler.RequestMiddleware(prometheus.RequestMiddleware()),
			handler.ResolverMiddleware(prometheus.ResolverMiddleware()),
		))
	})

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/ping", ping)

	log.Printf("connect to http://localhost:%d/ for GraphQL playground", conf.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), router))
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}
