package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen-contrib/prometheus"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	playground "github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/auth"
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
	// If cannot parse, just assume false
	gqlPlayground, _ := strconv.ParseBool(os.Getenv("GQL_PLAYGROUND"))

	store := memory.NewStore()
	computer := vm.NewComputer(store)

	resolver := playground.NewResolver(store, computer)

	// Register gql metrics
	prometheus.Register()

	router := chi.NewRouter()

	if gqlPlayground {
		router.Handle("/", handler.Playground("GraphQL playground", "/query"))
	}

	router.Route("/query", func(r chi.Router) {
		// Add CORS middleware around every request
		// See https://github.com/rs/cors for full option listing
		r.Use(cors.New(cors.Options{
			AllowedOrigins:   allowedOriginList,
			AllowCredentials: true,
			Debug:            gqlPlayground,
		}).Handler)

		r.Use(auth.Middleware())
		r.Handle("/", handler.GraphQL(playground.NewExecutableSchema(playground.Config{Resolvers: resolver})))
	})

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/ping", ping)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}
