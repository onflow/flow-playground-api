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

package client

import "net/http"

// Var adds a variable into the outgoing request
func Var(name string, value interface{}) Option {
	return func(bd *Request) {
		if bd.Variables == nil {
			bd.Variables = map[string]interface{}{}
		}

		bd.Variables[name] = value
	}
}

// Operation sets the operation name for the outgoing request
func Operation(name string) Option {
	return func(bd *Request) {
		bd.OperationName = name
	}
}

// Path sets the url that this request will be made against, useful if you are mounting your entire router
// and need to specify the url to the graphql endpoint.
func Path(url string) Option {
	return func(bd *Request) {
		bd.HTTP.URL.Path = url
	}
}

// AddHeader adds a header to the outgoing request. This is useful for setting expected Authentication headers for example.
func AddHeader(key string, value string) Option {
	return func(bd *Request) {
		bd.HTTP.Header.Add(key, value)
	}
}

// BasicAuth authenticates the request using http basic auth.
func BasicAuth(username, password string) Option {
	return func(bd *Request) {
		bd.HTTP.SetBasicAuth(username, password)
	}
}

// AddCookie adds a cookie to the outgoing request
func AddCookie(cookie *http.Cookie) Option {
	return func(bd *Request) {
		bd.HTTP.AddCookie(cookie)
	}
}
