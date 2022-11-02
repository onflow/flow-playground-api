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

package client_test

import (
	"encoding/json"
	"github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		require.Equal(t, `{"query":"user(id:$id){name}","variables":{"id":1}}`, string(b))

		err = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"name": "bob",
			},
		})
		if err != nil {
			panic(err)
		}
	})

	c := client.New(h)

	var resp struct {
		Name string
	}

	w := httptest.NewRecorder()

	err := c.Post(w, "user(id:$id){name}", &resp, client.Var("id", 1))
	require.NoError(t, err)

	require.Equal(t, "bob", resp.Name)
}

func TestAddHeader(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "ASDF", r.Header.Get("Test-Key"))

		_, _ = w.Write([]byte(`{}`))
	})

	c := client.New(h)

	var resp struct{}

	w := httptest.NewRecorder()

	err := c.Post(w, "{ id }", &resp,
		client.AddHeader("Test-Key", "ASDF"),
	)
	require.NoError(t, err)
}

func TestAddClientHeader(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "ASDF", r.Header.Get("Test-Key"))

		_, _ = w.Write([]byte(`{}`))
	})

	c := client.New(h, client.AddHeader("Test-Key", "ASDF"))

	var resp struct{}

	w := httptest.NewRecorder()

	err := c.Post(w, "{ id }", &resp)
	require.NoError(t, err)
}

func TestBasicAuth(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "user", user)
		require.Equal(t, "pass", pass)

		_, _ = w.Write([]byte(`{}`))
	})

	c := client.New(h)

	var resp struct{}

	w := httptest.NewRecorder()

	err := c.Post(w, "{ id }", &resp,
		client.BasicAuth("user", "pass"),
	)
	require.NoError(t, err)
}

func TestAddCookie(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("foo")
		require.NoError(t, err)
		require.Equal(t, "value", c.Value)

		_, _ = w.Write([]byte(`{}`))
	})

	c := client.New(h)

	var resp struct{}

	w := httptest.NewRecorder()

	err := c.Post(w, "{ id }", &resp,
		client.AddCookie(&http.Cookie{Name: "foo", Value: "value"}),
	)
	require.NoError(t, err)
}
