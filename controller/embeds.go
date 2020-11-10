package controller

import (
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

type EmbedsHandler struct {
	store  storage.Store
	logger *logrus.Logger
}

func NewEmbedsHandler(
	store storage.Store,
	logger *logrus.Logger,
) *EmbedsHandler {
	return &EmbedsHandler{
		store:  store,
		logger: logger,
	}
}

func (e *EmbedsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// implementation here
	log.Println("Follow the white rabbit")

	projectId := getUUID("projectID", w, r)
	childId := getUUID("scriptId", w, r)
	scriptType := getURLParam("scriptType", w, r)

	log.Println(projectId, scriptType, childId)

	scriptId := model.ProjectChildID{
		ID:        childId,
		ProjectID: projectId,
	}

	code, getErr := e.GetCode(scriptId, scriptType)
	if getErr != nil {
		w.Write([]byte(getErr.Error()))
		http.Error(w, http.StatusText(422), 422)
		return
	}

	log.Println(code)

	// createSnippet(code, w)
}

func (e *EmbedsHandler) GetCode(id model.ProjectChildID, scriptType string) (string, error) {
	var code string
	var err error

	switch scriptType {
	case "script":
		code, err = getScriptTemplate(e, id)
	case "transaction":
		code, err = getTransactionTemplate(e, id)
	case "contract":
		code, err = getAccountTemplate(e, id)
	}

	return code, err
}

func getScriptTemplate(e *EmbedsHandler, id model.ProjectChildID) (string, error) {
	var tmpl *model.ScriptTemplate
	err := e.store.GetScriptTemplate(id, tmpl)
	log.Println(tmpl.Script)
	if err != nil {
		return "", err
	} else {
		return tmpl.Script, nil
	}
}

func getTransactionTemplate(e *EmbedsHandler, id model.ProjectChildID) (string, error) {
	var tmpl *model.TransactionTemplate
	err := e.store.GetTransactionTemplate(id, tmpl)
	if err != nil {
		return "", err
	} else {
		return tmpl.Script, nil
	}
}

func getAccountTemplate(e *EmbedsHandler, id model.ProjectChildID) (string, error) {
	var tmpl *model.InternalAccount
	err := e.store.GetAccount(id, tmpl)
	if err != nil {
		return "", err
	} else {
		return tmpl.DraftCode, nil
	}
}

func getUUID(paramName string, w http.ResponseWriter, r *http.Request) (id uuid.UUID) {
	rawId := chi.URLParam(r, paramName)
	log.Println(paramName, rawId)
	id, err := model.UnmarshalUUID(rawId)
	if err != nil || rawId == "" {
		w.Write([]byte(err.Error()))
		http.Error(w, http.StatusText(422), 422)
	}
	return id
}

func getURLParam(paramName string, w http.ResponseWriter, r *http.Request) string {
	param := chi.URLParam(r, paramName)
	if param == "" {
		w.Write([]byte(param + " can't be empty"))
		http.Error(w, http.StatusText(422), 422)
	}
	return param
}

func createSnippet(code string, w http.ResponseWriter) {
	lexer := lexers.Get("swift")
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := html.New(html.WithClasses(true))

	// TODO: Catch error here
	formatter.WriteCSS(w, style)

	// TODO: Catch error here
	iterator, _ := lexer.Tokenise(nil, code)

	// TODO: Catch error here
	formatter.Format(w, style, iterator)
}
