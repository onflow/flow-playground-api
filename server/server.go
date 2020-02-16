package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/handler"
	"github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
	"github.com/dapperlabs/flow-playground-api/vm"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	store := memory.NewStore()
	computer := vm.NewComputer(store)

	resolver := playground.NewResolver(store, computer)

	http.Handle("/", handler.Playground("GraphQL playground", "/query"))
	http.Handle("/query", handler.GraphQL(playground.NewExecutableSchema(playground.Config{Resolvers: resolver})))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
