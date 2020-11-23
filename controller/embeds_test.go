package controller

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-go/engine/execution/state/delta"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

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

	return fmt.Errorf("can't find script template")
}

func TestEmbedsHandler_ServeHTTP(t *testing.T) {
	mockStore := MockStore{}
	playgroundBase := "http://localhost:3000"

	handler := NewEmbedsHandler(&mockStore, playgroundBase)
	url := "http://playground.api.com/embed/24278e82-9316-4559-96f2-573ec58f618f/script/9473b82c-36ea-4810-ad3f-7ea5497d9cae"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))

	assert.NotNil(t, body)
}
