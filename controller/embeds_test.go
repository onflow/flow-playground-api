package controller

import (
	"context"
	"fmt"
	"github.com/alecthomas/assert"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Utility method to send requests
func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}

// Unit tests start here
func TestEmbedsHandler_ServeHTTP(t *testing.T) {
	store := memory.NewStore()
	playgroundBaseURL := "http://localhost:3000"
	embedsHandler := NewEmbedsHandler(store, playgroundBaseURL)

	projectID := "24278e82-9316-4559-96f2-573ec58f618f"
	scriptType := "script"
	scriptID := "9473b82c-36ea-4810-ad3f-7ea5497d9cae"

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

	snippetUrl := fmt.Sprintf("/embed?project=%s&type=%s&id=%s", projectID, scriptType, scriptID)

	response, _ := testRequest(t, ts, "GET", snippetUrl, nil)

	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestEmbedsHandler_getUUID(t *testing.T) {
	projectId := "24278e82-9316-4559-96f2-573ec58f618f"
	scriptType := "script"
	scriptId := "9473b82c-36ea-4810-ad3f-7ea5497d9cae"

	requestURL := fmt.Sprintf("http://playground-api.com/embed?project=%s&type=%s&id=%s", projectId, scriptType, scriptId)
	r := httptest.NewRequest("GET", requestURL, nil)

	rctx := chi.NewRouteContext()

	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	// Check that projectUUID extracted properly
	projectUUID, err := getUUID("projectID", r)
	if err != nil {
		t.Fail()
	}
	expected, _ := uuid.Parse(projectId)
	assert.Equal(t, expected, projectUUID)

	// Check that scriptId extracted properly
	scriptUUID, err := getUUID("scriptId", r)
	if err != nil {
		t.Fail()
	}
	expected, _ = uuid.Parse(scriptId)
	assert.Equal(t, expected, scriptUUID)

	scriptTypeParam, _ := getURLParam("scriptType", r)
	assert.Equal(t, scriptType, scriptTypeParam)
}

func TestEmbedsHandler_createCodeStyles(t *testing.T) {
	const styles = ".chroma { background: red }"
	const styleName = "red"
	codeStyle := createCodeStyles(styles, styleName)

	// Check that id is created properly
	themeName := fmt.Sprintf("theme-%s", styleName)
	haveThemeString := strings.Contains(codeStyle, themeName)
	assert.True(t, haveThemeString)

	// Chroma class names shall have another class to ensure themes are working properly
	adjustedClassName := fmt.Sprintf(".chroma.%s", styleName)
	adjustedStyles := strings.ReplaceAll(styles, ".chroma", adjustedClassName)

	// Check that inserted HTML have adjusted styles
	haveStylesHTML := strings.Contains(codeStyle, adjustedStyles)
	assert.True(t, haveStylesHTML)
}

func TestEmbedsHandler_generateWrapperStyles(t *testing.T) {
	generatedStyles := generateWrapperStyles()

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
		assert.True(t, strings.Contains(generatedStyles, value))
	}
}

func TestEmbedsHandler_createSnippetStyles(t *testing.T) {
	snippetStyles := createSnippetStyles()
	generatedStyles := generateWrapperStyles()

	haveProperId := strings.Contains(snippetStyles, "cadence-styles")
	assert.True(t, haveProperId)

	haveStylesHTML := strings.Contains(snippetStyles, generatedStyles)
	assert.True(t, haveStylesHTML)
}

func TestEmbedsHandler_wrapCodeBlock(t *testing.T) {
	htmlBlock := "<div>no code</div>"
	styleName := "red"
	playgroundUrl := "https://play.onflow.org/"

	wrappedCodeBlock := wrapCodeBlock(htmlBlock, styleName, playgroundUrl)

	haveWrappedCode := strings.Contains(wrappedCodeBlock, htmlBlock)
	assert.True(t, haveWrappedCode)

	haveProperUrl := strings.Contains(wrappedCodeBlock, playgroundUrl)
	assert.True(t, haveProperUrl)
}
