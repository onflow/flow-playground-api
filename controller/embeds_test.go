package controller

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/alecthomas/assert"
	"github.com/dapperlabs/flow-go/engine/execution/state/delta"
	"github.com/stretchr/testify/require"
	// playground "github.com/dapperlabs/flow-playground-api"
	// "github.com/dapperlabs/flow-playground-api/auth"
	// "github.com/dapperlabs/flow-playground-api/compute"
	"github.com/dapperlabs/flow-playground-api/model"
	// "github.com/dapperlabs/flow-playground-api/storage/datastore"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	// "github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	// "time"
)

// version to create project
var version, _ = semver.NewVersion("0.1.0")

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
	deltas := make([]delta.Delta, 0)
	ttpls := make([]*model.TransactionTemplate, 0)
	stpls := make([]*model.ScriptTemplate, 0)

	internalProj.UserID = user.ID

	projErr := store.CreateProject(internalProj, deltas, accounts, ttpls, stpls)
	if projErr != nil {
		t.Fail()
	}

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
		wrongScriptId := model.MarshalUUID(uuid.New())

		snippetUrl := fmt.Sprintf("/embed?project=%s&type=%s&id=%s", projectID, scriptType, wrongScriptId)

		response, _ := testRequest(t, ts, "GET", snippetUrl, nil)

		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	})

	t.Run("Shall get 400 on non-existing project id", func(t *testing.T) {
		wrongProjectId := model.MarshalUUID(uuid.New())

		snippetUrl := fmt.Sprintf("/embed?project=%s&type=%s&id=%s", wrongProjectId, scriptType, scriptID)

		response, _ := testRequest(t, ts, "GET", snippetUrl, nil)

		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	})

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

func TestMethods(t *testing.T) {
	t.Run("Generate wrapper styles", func(t *testing.T) {
		generatedStyles := generateWrapperStyles()
		assertHaveWrapper(t, generatedStyles)
	})
	t.Run("Create snippet styles", func(t *testing.T) {
		snippetStyles := createSnippetStyles()
		generatedStyles := generateWrapperStyles()

		haveProperId := strings.Contains(snippetStyles, "cadence-styles")
		assert.True(t, haveProperId)

		haveStylesHTML := strings.Contains(snippetStyles, generatedStyles)
		assert.True(t, haveStylesHTML)
	})
	t.Run("Create code styles", func(t *testing.T) {
		const styles = ".chroma { background: red }"
		const styleName = "red"
		codeStyle := createCodeStyles(styles, styleName)

		// Check that id is created properly
		assertHaveTheme(t, codeStyle, styleName)

		// Chroma class names shall have another class to ensure themes are working properly
		adjustedClassName := fmt.Sprintf(".chroma.%s", styleName)
		adjustedStyles := strings.ReplaceAll(styles, ".chroma", adjustedClassName)

		// Check that inserted HTML have adjusted styles
		haveStylesHTML := strings.Contains(codeStyle, adjustedStyles)
		assert.True(t, haveStylesHTML)
	})
	t.Run("Wrap code block", func(t *testing.T) {
		htmlBlock := "<div>no code</div>"
		styleName := "red"
		playgroundUrl := "https://play.onflow.org/"

		wrappedCodeBlock := wrapCodeBlock(htmlBlock, styleName, playgroundUrl)

		haveWrappedCode := strings.Contains(wrappedCodeBlock, htmlBlock)
		assert.True(t, haveWrappedCode)

		haveProperUrl := strings.Contains(wrappedCodeBlock, playgroundUrl)
		assert.True(t, haveProperUrl)
	})
}
