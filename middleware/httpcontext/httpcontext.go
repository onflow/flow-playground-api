/*
 * Flow Playground
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package httpcontext

import (
	"context"
	"net/http"
)

const httpCtxKeyWriter = "http_writer"
const httpCtxKeyRequest = "http_request"

func Request(ctx context.Context) *http.Request {
	return ctx.Value(httpCtxKeyRequest).(*http.Request)
}

func Writer(ctx context.Context) http.ResponseWriter {
	return ctx.Value(httpCtxKeyWriter).(http.ResponseWriter)
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httpCtxKeyWriter, w)
			ctx = context.WithValue(ctx, httpCtxKeyRequest, r)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
