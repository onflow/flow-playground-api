package test

import (
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/middleware/errors"
	client2 "github.com/dapperlabs/flow-playground-api/test/client"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	playground "github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/auth"
	legacyauth "github.com/dapperlabs/flow-playground-api/auth/legacy"
	"github.com/dapperlabs/flow-playground-api/middleware/httpcontext"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Client struct {
	client        *client2.Client
	resolver      *playground.Resolver
	sessionCookie *http.Cookie
	projects      *blockchain.Projects
	store         storage.Store
}

func (c *Client) Post(query string, response interface{}, options ...client2.Option) error {
	w := httptest.NewRecorder()

	err := c.client.Post(w, query, response, options...)

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == sessionName {
			c.sessionCookie = cookie
		}
	}

	return err
}

func (c *Client) MustPost(query string, response interface{}, options ...client2.Option) {
	err := c.Post(query, response, options...)
	if err != nil {
		panic(err)
	}
}

func (c *Client) SessionCookie() *http.Cookie {
	return c.sessionCookie
}

func (c *Client) ClearSessionCookie() {
	c.sessionCookie = nil
}

const sessionName = "flow-playground-test"

var version, _ = semver.NewVersion("0.1.0")

// keep same instance of store due to connection pool exhaustion
var store storage.Store

func newStore() storage.Store {
	if store != nil {
		return store
	}

	if strings.EqualFold(os.Getenv("FLOW_STORAGEBACKEND"), storage.PostgreSQL) {
		var datastoreConf storage.DatabaseConfig
		if err := envconfig.Process("FLOW_DB", &datastoreConf); err != nil {
			panic(err)
		}

		store = storage.NewPostgreSQL(&datastoreConf)
	} else {
		store = storage.NewSqlite()
	}

	return store
}

func newClient() *Client {
	store := newStore()
	authenticator := auth.NewAuthenticator(store, sessionName)
	chain := blockchain.NewProjects(store, initAccounts)
	resolver := playground.NewResolver(version, store, authenticator, chain)

	c := newClientWithResolver(resolver)
	c.store = store
	c.projects = chain
	return c
}

func newClientWithResolver(resolver *playground.Resolver) *Client {
	router := chi.NewRouter()
	router.Use(httpcontext.Middleware())
	router.Use(legacyauth.MockProjectSessions())

	localHub := sentry.CurrentHub().Clone()
	logger := logrus.StandardLogger()
	entry := logrus.NewEntry(logger)
	router.Handle("/", playground.GraphQLHandler(resolver, errors.Middleware(entry, localHub)))

	return &Client{
		client:   client2.New(router),
		resolver: resolver,
	}
}

func createProject(t *testing.T, c *Client) Project {
	var resp CreateProjectResponse

	err := c.Post(
		MutationCreateProject,
		&resp,
		client2.Var("title", "foo"),
		client2.Var("seed", 42),
		client2.Var("description", "desc"),
		client2.Var("readme", "rtfm"),
		client2.Var("numberOfAccounts", 5),
		client2.Var("transactionTemplates", []string{}),
		client2.Var("scriptTemplates", []string{}),
		client2.Var("contractTemplates", []string{}),
	)
	require.NoError(t, err)

	proj := resp.CreateProject
	internalProj := c.resolver.LastCreatedProject()

	proj.Secret = internalProj.Secret.String()

	return proj
}
