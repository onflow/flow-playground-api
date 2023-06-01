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

package e2eTest

import (
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api"
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/dapperlabs/flow-playground-api/middleware/errors"
	"github.com/dapperlabs/flow-playground-api/server/config"
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

	"github.com/dapperlabs/flow-playground-api/auth"
	legacyauth "github.com/dapperlabs/flow-playground-api/auth/legacy"
	"github.com/dapperlabs/flow-playground-api/middleware/httpcontext"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Client struct {
	client        *client.Client
	resolver      *playground.Resolver
	sessionCookie *http.Cookie
	projects      *blockchain.Projects
	store         storage.Store
}

func (c *Client) Post(query string, response interface{}, options ...client.Option) error {
	w := httptest.NewRecorder()

	err := c.client.Post(w, query, response, options...)

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == sessionName {
			c.sessionCookie = cookie
		}
	}

	return err
}

func (c *Client) MustPost(query string, response interface{}, options ...client.Option) {
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

const sessionName = "flow-playground-e2eTest"

var version, _ = semver.NewVersion("0.1.0")

// keep same instance of store due to connection pool exhaustion
var store storage.Store

func newStore() storage.Store {
	if store != nil {
		return store
	}

	if strings.EqualFold(os.Getenv("FLOW_STORAGEBACKEND"), storage.PostgreSQL) {
		var datastoreConf config.DatabaseConfig
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
		client:   client.New(router),
		resolver: resolver,
	}
}

func createProject(t *testing.T, c *Client) Project {
	var resp CreateProjectResponse

	err := c.Post(
		MutationCreateProject,
		&resp,
		client.Var("title", "foo"),
		client.Var("seed", 42),
		client.Var("description", "desc"),
		client.Var("readme", "rtfm"),
		client.Var("numberOfAccounts", 5),
		client.Var("accounts", []string{}),
		client.Var("transactionTemplates", []string{}),
		client.Var("scriptTemplates", []string{}),
		client.Var("contractTemplates", []string{}),
	)
	require.NoError(t, err)

	proj := resp.CreateProject
	internalProj := c.resolver.LastCreatedProject()

	proj.Secret = internalProj.Secret.String()

	return proj
}
