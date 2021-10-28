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

package controller

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/alecthomas/assert"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/onflow/flow-go/engine/execution/state/delta"
	"github.com/stretchr/testify/require"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
)

// version to create project
var version, _ = semver.NewVersion("0.1.0")

// Utility method to send requests
func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func assertHaveWrapper(t *testing.T, body string) {
	testClasses := [...]string{
		"cadence-snippet",
		"cadence-snippet pre.chroma",
		".cadence-code-block .chroma",
		".cadence-info-block",
		".cadence-info-block img",
		".cadence-info-block a",
		".flow-playground-logo",
		".cadence-info-block .umbrella",
	}

	for _, value := range testClasses {
		assert.True(t, strings.Contains(body, value))
	}
}

func assertHaveTheme(t *testing.T, body string, theme string) {
	checkTheme := theme

	// when theme is not provided it will default to "swapoff"
	if theme == "" {
		checkTheme = "swapoff"
	}

	themeName := fmt.Sprintf("theme-%s", checkTheme)
	haveThemeString := strings.Contains(body, themeName)
	assert.True(t, haveThemeString)
}

// Unit tests start here
func TestEmbedsHandler_ServeHTTP(t *testing.T) {

	store := memory.NewStore()
	playgroundBaseURL := "http://localhost:3000"
	embedsHandler := NewEmbedsHandler(store, playgroundBaseURL)

	projectID := "24278e82-9316-4559-96f2-573ec58f618f"
	scriptType := "script"
	scriptID := "9473b82c-36ea-4810-ad3f-7ea5497d9cae"

	parentID := uuid.New()

	user := model.User{
		ID: uuid.New(),
	}

	internalProj := &model.InternalProject{
		ID:       uuid.MustParse(projectID),
		Secret:   uuid.New(),
		PublicID: uuid.New(),
		ParentID: &parentID,
		Seed:     0,
		Title:    "test-project",
		Persist:  false,
		Version:  version,
	}

	accounts := make([]*model.InternalAccount, 0)
	cons := make([]*model.Contract, 0)
	deltas := make([]delta.Delta, 0)
	ttpls := make([]*model.TransactionTemplate, 0)
	stpls := make([]*model.ScriptTemplate, 0)

	internalProj.UserID = user.ID

	projErr := store.CreateProject(internalProj, deltas, accounts, cons, ttpls, stpls)
	require.NoError(t, projErr)

	script := `
	pub fun main(): Int {
	  return 42
	}
	`

	scriptTemplate := model.ScriptTemplate{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.MustParse(scriptID),
			ProjectID: uuid.MustParse(projectID),
		},
		Title:  "test contract",
		Script: script,
	}

	// insert your mock data
	err := store.InsertScriptTemplate(&scriptTemplate)
	require.NoError(t, err)

	r := chi.NewRouter()
	r.Get("/embed", embedsHandler.ServeHTTP)

	ts := httptest.NewServer(r)
	defer ts.Close()

	t.Run("Shall get existing script", func(t *testing.T) {
		snippetUrl := fmt.Sprintf("/embed?project=%s&type=%s&id=%s", projectID, scriptType, scriptID)

		response, body := testRequest(t, ts, "GET", snippetUrl, nil)

		assert.Equal(t, http.StatusOK, response.StatusCode)

		assertHaveWrapper(t, body)
		assertHaveTheme(t, body, "")
	})

	t.Run("Shall get 400 on wrong script type", func(t *testing.T) {
		wrongScriptType := "not-the-droid-you-are-looking-for"

		snippetUrl := fmt.Sprintf("/embed?project=%s&type=%s&id=%s", projectID, wrongScriptType, scriptID)

		response, _ := testRequest(t, ts, "GET", snippetUrl, nil)

		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	})

	t.Run("Shall get 400 on non-existing script id", func(t *testing.T) {
		wrongScriptID := model.MarshalUUID(uuid.New())

		snippetUrl := fmt.Sprintf("/embed?project=%s&type=%s&id=%s", projectID, scriptType, wrongScriptID)

		response, _ := testRequest(t, ts, "GET", snippetUrl, nil)

		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	})

	t.Run("Shall get 400 on non-existing project id", func(t *testing.T) {
		wrongProjectID := model.MarshalUUID(uuid.New())

		snippetUrl := fmt.Sprintf("/embed?project=%s&type=%s&id=%s", wrongProjectID, scriptType, scriptID)

		response, _ := testRequest(t, ts, "GET", snippetUrl, nil)

		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	})

}

func TestGetUUID(t *testing.T) {
	projectID := "24278e82-9316-4559-96f2-573ec58f618f"
	scriptType := "script"
	scriptID := "9473b82c-36ea-4810-ad3f-7ea5497d9cae"

	requestURL := fmt.Sprintf("http://playground-api.com/embed?project=%s&type=%s&id=%s", projectID, scriptType, scriptID)
	r := httptest.NewRequest("GET", requestURL, nil)

	rctx := chi.NewRouteContext()

	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	// Check that projectUUID extracted properly
	projectUUID, err := getUUID("project", r)
	require.NoError(t, err)

	expected, _ := uuid.Parse(projectID)
	assert.Equal(t, expected, projectUUID)

	// Check that scriptID extracted properly
	scriptUUID, err := getUUID("id", r)
	require.NoError(t, err)

	expected, _ = uuid.Parse(scriptID)
	assert.Equal(t, expected, scriptUUID)

	scriptTypeParam, _ := getURLParam("type", r)
	assert.Equal(t, scriptType, scriptTypeParam)
}

func TestGenerateWrapperStyles(t *testing.T) {
	generatedStyles := generateWrapperStyles()
	assertHaveWrapper(t, generatedStyles)
}

func TestCreateSnippetStyles(t *testing.T) {
	snippetStyles := createSnippetStyles()
	generatedStyles := generateWrapperStyles()

	haveProperId := strings.Contains(snippetStyles, "cadence-styles")
	assert.True(t, haveProperId)

	haveStylesHTML := strings.Contains(snippetStyles, generatedStyles)
	assert.True(t, haveStylesHTML)
}

func TestCreateCodeStyles(t *testing.T) {
	styles := ".chroma { background: red }"
	styleName := "red"

	codeStyle := createCodeStyles(styles, styleName)
	expectedCodeStyle := ".chroma.red { background: red }"
	haveUpdatedStyle := strings.Contains(codeStyle, expectedCodeStyle)
	assert.True(t, haveUpdatedStyle)

	// Check that id is created properly
	assertHaveTheme(t, codeStyle, styleName)

	// Chroma class names shall have another class to ensure themes are working properly
	adjustedClassName := fmt.Sprintf(".chroma.%s", styleName)
	adjustedStyles := strings.ReplaceAll(styles, ".chroma", adjustedClassName)

	// Check that inserted HTML have adjusted styles
	haveStylesHTML := strings.Contains(codeStyle, adjustedStyles)
	assert.True(t, haveStylesHTML)
}

func TestWrapCodeBlock(t *testing.T) {
	htmlBlock := "<div>no code</div>"
	styleName := "red"
	playgroundURL := "https://play.onflow.org/"

	wrappedCodeBlock := wrapCodeBlock(htmlBlock, styleName, playgroundURL)

	haveWrappedCode := strings.Contains(wrappedCodeBlock, htmlBlock)
	assert.True(t, haveWrappedCode)

	haveProperURL := strings.Contains(wrappedCodeBlock, playgroundURL)
	assert.True(t, haveProperURL)
}
