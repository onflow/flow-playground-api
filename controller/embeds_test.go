package controller

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/alecthomas/assert"
	"github.com/dapperlabs/flow-go/engine/execution/state/delta"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TODO: can we move this to another file?
// Mock store requres implementation of all those methods
type MockStore struct {
}

func (ms *MockStore) InsertUser(user *model.User) error {
	panic("implement me")
}

func (ms *MockStore) GetUser(id uuid.UUID, user *model.User) error {
	panic("implement me")
}

func (ms *MockStore) CreateProject(proj *model.InternalProject, registerDeltas []delta.Delta, accounts []*model.InternalAccount, ttpl []*model.TransactionTemplate, stpl []*model.ScriptTemplate) error {
	panic("implement me")
}

func (ms *MockStore) UpdateProject(input model.UpdateProject, proj *model.InternalProject) error {
	panic("implement me")
}

func (ms *MockStore) UpdateProjectOwner(id, userID uuid.UUID) error {
	panic("implement me")
}

func (ms *MockStore) UpdateProjectVersion(id uuid.UUID, version *semver.Version) error {
	panic("implement me")
}

func (ms *MockStore) ResetProjectState(newDeltas []delta.Delta, proj *model.InternalProject) error {
	panic("implement me")
}

func (ms *MockStore) GetProject(id uuid.UUID, proj *model.InternalProject) error {
	panic("implement me")
}

func (ms *MockStore) InsertAccount(acc *model.InternalAccount) error {
	panic("implement me")
}

func (ms *MockStore) GetAccount(id model.ProjectChildID, acc *model.InternalAccount) error {
	panic("implement me")
}

func (ms *MockStore) UpdateAccount(input model.UpdateAccount, acc *model.InternalAccount) error {
	panic("implement me")
}

func (ms *MockStore) UpdateAccountAfterDeployment(input model.UpdateAccount, states map[uuid.UUID]model.AccountState, delta delta.Delta, acc *model.InternalAccount) error {
	panic("implement me")
}

func (ms *MockStore) GetAccountsForProject(projectID uuid.UUID, accs *[]*model.InternalAccount) error {
	panic("implement me")
}

func (ms *MockStore) DeleteAccount(id model.ProjectChildID) error {
	panic("implement me")
}

func (ms *MockStore) InsertTransactionTemplate(tpl *model.TransactionTemplate) error {
	panic("implement me")
}

func (ms *MockStore) UpdateTransactionTemplate(input model.UpdateTransactionTemplate, tpl *model.TransactionTemplate) error {
	panic("implement me")
}

func (ms *MockStore) GetTransactionTemplate(id model.ProjectChildID, tpl *model.TransactionTemplate) error {
	panic("implement me")
}

func (ms *MockStore) GetTransactionTemplatesForProject(projectID uuid.UUID, tpls *[]*model.TransactionTemplate) error {
	panic("implement me")
}

func (ms *MockStore) DeleteTransactionTemplate(id model.ProjectChildID) error {
	panic("implement me")
}

func (ms *MockStore) InsertTransactionExecution(exe *model.TransactionExecution, states map[uuid.UUID]model.AccountState, delta delta.Delta) error {
	panic("implement me")
}

func (ms *MockStore) GetTransactionExecutionsForProject(projectID uuid.UUID, exes *[]*model.TransactionExecution) error {
	panic("implement me")
}

func (ms *MockStore) InsertScriptTemplate(tpl *model.ScriptTemplate) error {
	panic("implement me")
}

func (ms *MockStore) UpdateScriptTemplate(input model.UpdateScriptTemplate, tpl *model.ScriptTemplate) error {
	panic("implement me")
}

func (ms *MockStore) GetScriptTemplatesForProject(projectID uuid.UUID, tpls *[]*model.ScriptTemplate) error {
	panic("implement me")
}

func (ms *MockStore) DeleteScriptTemplate(id model.ProjectChildID) error {
	panic("implement me")
}

func (ms *MockStore) InsertScriptExecution(exe *model.ScriptExecution) error {
	panic("implement me")
}

func (ms *MockStore) GetScriptExecutionsForProject(projectID uuid.UUID, exes *[]*model.ScriptExecution) error {
	panic("implement me")
}

func (ms *MockStore) GetRegisterDeltasForProject(projectID uuid.UUID, deltas *[]*model.RegisterDelta) error {
	panic("implement me")
}

func (ms *MockStore) GetScriptTemplate(id model.ProjectChildID, tmpl *model.ScriptTemplate) error {
	tmpl.Title = "test contract"
	tmpl.Index = 1
	tmpl.Script = `
		pub fun main(): Int {
			return 42
		}
	`

	return nil
}

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
	mockStore := MockStore{}
	playgroundBase := "http://localhost:3000"
	embedsHandler := NewEmbedsHandler(&mockStore, playgroundBase)

	r := chi.NewRouter()
	r.Get("/embed/{projectID}/{scriptType}/{scriptId}", embedsHandler.ServeHTTP)

	ts := httptest.NewServer(r)
	defer ts.Close()

	snippetUrl := "/embed/24278e82-9316-4559-96f2-573ec58f618f/script/9473b82c-36ea-4810-ad3f-7ea5497d9cae"

	response, _ := testRequest(t, ts, "GET", snippetUrl, nil)

	assert.Equal(t, response.StatusCode, http.StatusOK)
}

func TestEmbedsHandler_getUUID(t *testing.T) {
	projectId := "24278e82-9316-4559-96f2-573ec58f618f"
	scriptType := "script"
	scriptId := "9473b82c-36ea-4810-ad3f-7ea5497d9cae"

	r := httptest.NewRequest("GET", "http://play.onflow.org", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectID", projectId)
	rctx.URLParams.Add("scriptType", scriptType)
	rctx.URLParams.Add("scriptId", scriptId)

	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	// Check that projectUUID extracted properly
	projectUUID, err := getUUID("projectID", r)
	if err != nil {
		t.Fail()
	}
	expected, _ := uuid.Parse(projectId)
	assert.Equal(t, projectUUID, expected, projectUUID)

	// Check that scriptId extracted properly
	scriptUUID, err := getUUID("scriptId", r)
	if err != nil {
		t.Fail()
	}
	expected, _ = uuid.Parse(scriptId)
	assert.Equal(t, scriptUUID, expected, projectUUID)

	scriptTypeParam, _ := getURLParam("scriptType", r)
	assert.Equal(t, scriptTypeParam, scriptType)
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
