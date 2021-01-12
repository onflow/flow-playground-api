package main

import (
	"context"
	"fmt"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/go-chi/render"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen-contrib/prometheus"
	"github.com/99designs/gqlgen/handler"
	"github.com/Masterminds/semver"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/go-chi/chi"
	gsessions "github.com/gorilla/sessions"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

	playground "github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/auth"
	"github.com/dapperlabs/flow-playground-api/build"
	"github.com/dapperlabs/flow-playground-api/compute"
	"github.com/dapperlabs/flow-playground-api/middleware/errors"
	"github.com/dapperlabs/flow-playground-api/middleware/httpcontext"
	"github.com/dapperlabs/flow-playground-api/middleware/sessions"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/storage/datastore"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
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

const sessionName = "flow-playground"

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

	computer, err := compute.NewComputer(conf.LedgerCacheSize)
	if err != nil {
		panic(err)
	}

	sessionAuthKey := []byte(conf.SessionAuthKey)

	authenticator := auth.NewAuthenticator(store, sessionName)

	resolver := playground.NewResolver(build.Version(), store, computer, authenticator)

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

		cookieStore := gsessions.NewCookieStore(sessionAuthKey)
		cookieStore.MaxAge(int(conf.SessionMaxAge.Seconds()))

		cookieStore.Options.Secure = conf.SessionCookiesSecure
		cookieStore.Options.HttpOnly = conf.SessionCookiesHTTPOnly

		if conf.SessionCookiesSameSiteNone {
			cookieStore.Options.SameSite = http.SameSiteNoneMode
		}

		r.Use(httpcontext.Middleware())
		r.Use(sessions.Middleware(cookieStore))

		r.Handle("/", handler.GraphQL(
			playground.NewExecutableSchema(playground.Config{Resolvers: resolver}),
			handler.RequestMiddleware(errors.Middleware(entry)),
			handler.RequestMiddleware(prometheus.RequestMiddleware()),
			handler.ResolverMiddleware(prometheus.ResolverMiddleware()),
		))
	})

	embedsHandler := controller.NewEmbedsHandler(store, conf.PlaygroundBaseURL)
	router.Handle("/embed", embedsHandler)

	utilsHandler := controller.NewUtilsHandler()
	router.Route("/utils", func(r chi.Router) {
		// Add CORS middleware around every request
		// See https://github.com/rs/cors for full option listing
		r.Use(cors.New(cors.Options{
			AllowedOrigins: conf.AllowedOrigins,
			Debug:          conf.Debug,
		}).Handler)

		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.HandleFunc("/version", utilsHandler.VersionHandler)
	})

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/ping", ping)

	logStartMessage(build.Version())

	log.Printf("Connect to http://localhost:%d/ for GraphQL playground", conf.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), router))
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func logStartMessage(version *semver.Version) {
	if version != nil {
		log.Printf("Starting Playground API (Version %s)", version)
	} else {
		log.Print("Starting Playground API")
	}
}
