package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/99designs/gqlgen-contrib/prometheus"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	playground "github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/auth"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/storage/datastore"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
	"github.com/dapperlabs/flow-playground-api/vm"
)

const defaultPort = "8080"
const defaultAllowedOrigins = "http://localhost:3000"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var allowedOriginList []string
	if allowedOrigins == "" {
		allowedOriginList = []string{defaultAllowedOrigins}
	} else {
		allowedOriginList = strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	}

	storeBackend := os.Getenv("STORE_BACKEND")
	var store storage.Store
	if strings.EqualFold(storeBackend, "datastore") {
		var err error
		projectID := os.Getenv("DATASTORE_PROJECT_ID")
		store, err = datastore.NewDatastore(context.Background(), &datastore.Config{DatastoreProjectID: projectID, DatastoreTimeout: time.Second * 5})
		if err != nil {
			// If datastore is expected, panic when we can't init
			panic(err)
		}
	} else {
		store = memory.NewStore()
	}

	computer := vm.NewComputer(store)

	resolver := playground.NewResolver(store, computer)

	// Register gql metrics
	prometheus.Register()

	router := chi.NewRouter()

	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   allowedOriginList,
		AllowCredentials: true,
		Debug:            true,
	}).Handler)

	router.Use(auth.Middleware())

	router.Handle("/", handler.Playground("GraphQL playground", "/query"))
	router.Handle("/query", handler.GraphQL(playground.NewExecutableSchema(playground.Config{Resolvers: resolver})))
	router.Handle("/metrics", promhttp.Handler())

	router.HandleFunc("/ping", ping)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}
