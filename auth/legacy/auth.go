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

package legacy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	gorillasessions "github.com/gorilla/sessions"

	"github.com/dapperlabs/flow-playground-api/middleware/sessions"
	"github.com/dapperlabs/flow-playground-api/model"
)

const projectSecretKeyName = "project-secret"

// ProjectInSession returns true if the given project is authorized in the current session.
//
// A project is authorized in a session if the session contains a reference to the
// project's secret.
func ProjectInSession(ctx context.Context, proj *model.InternalProject) bool {
	session := sessions.Get(ctx, getProjectSessionName(proj))

	secret, ok := session.Values[projectSecretKeyName]
	if !ok {
		return false
	}

	secretStr, ok := secret.(string)
	if !ok {
		return false
	}

	return secretStr == proj.Secret.String()
}

func getProjectSessionName(proj *model.InternalProject) string {
	return getProjectSessionNameFromString(proj.ID.String())
}

func getProjectSessionNameFromString(projectID string) string {
	return fmt.Sprintf("flow-%s", projectID)
}

const mockSessionAuthenticationKey = "1bbcf50e2e5f3e2d1801db50742f6a97"

func mockCookieStore() *gorillasessions.CookieStore {
	return gorillasessions.NewCookieStore([]byte(mockSessionAuthenticationKey))
}

// MockProjectSessions returns project sessions middleware to be used for testing.
func MockProjectSessions() func(http.Handler) http.Handler {
	return sessions.Middleware(mockCookieStore())
}

// MockProjectSessionCookie returns a session cookie that provides access to the given project.
func MockProjectSessionCookie(projectID, secret string) *http.Cookie {
	store := mockCookieStore()

	r := &http.Request{}
	w := httptest.NewRecorder()

	session, _ := store.Get(r, getProjectSessionNameFromString(projectID))

	session.Values[projectSecretKeyName] = secret

	err := session.Save(r, w)
	if err != nil {
		panic(err)
	}

	return w.Result().Cookies()[0]
}
