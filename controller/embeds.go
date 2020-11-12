package controller

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
)

type EmbedsHandler struct {
	store storage.Store
}

func NewEmbedsHandler(
	store storage.Store,
) *EmbedsHandler {
	return &EmbedsHandler{
		store: store,
	}
}

func (e *EmbedsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectId := getUUID("projectID", w, r)
	childId := getUUID("scriptId", w, r)
	scriptType := getURLParam("scriptType", w, r)

	scriptId := model.ProjectChildID{
		ID:        childId,
		ProjectID: projectId,
	}

	code, getErr := getCode(e, scriptId, scriptType)

	if getErr != nil {
		w.Write([]byte(getErr.Error()))
		http.Error(w, http.StatusText(422), 422)
		return
	}

	styles, html := createSnippet(code, w)

	// initial styles go with no padding, so we need to update the container to include some
	entryPoint := ".chroma { color"
	withPadding := ".chroma { padding: 1.5em; color"
	// We need to prepend styles strings with "`" in order for it to work inside of embedded javascript
	styles = "`" + strings.Replace(styles, entryPoint, withPadding, 1) + "`"

	stylesInjection := fmt.Sprintf(`
		var newStyleTag = document.createElement('style');
  		newStyleTag.innerHTML = %s
		document.head.appendChild(newStyleTag);
	`, styles)

	scriptInjection := fmt.Sprintf("document.write(`%s`)", html)

	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(stylesInjection))
	w.Write([]byte(scriptInjection))
}

func getCode(e *EmbedsHandler, id model.ProjectChildID, scriptType string) (string, error) {
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
	var tmpl model.ScriptTemplate
	err := e.store.GetScriptTemplate(id, &tmpl)
	log.Println(tmpl.Script)
	if err != nil {
		return "", err
	} else {
		return tmpl.Script, nil
	}
}

func getTransactionTemplate(e *EmbedsHandler, id model.ProjectChildID) (string, error) {
	var tmpl model.TransactionTemplate
	err := e.store.GetTransactionTemplate(id, &tmpl)
	if err != nil {
		return "", err
	} else {
		return tmpl.Script, nil
	}
}

func getAccountTemplate(e *EmbedsHandler, id model.ProjectChildID) (string, error) {
	var tmpl model.InternalAccount
	err := e.store.GetAccount(id, &tmpl)
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

func createSnippet(code string, w http.ResponseWriter) (string, string) {
	lexer := lexers.Get("swift")
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := html.New(html.WithClasses(true))

	var stylesBuffer bytes.Buffer
	// TODO: catch error here
	formatter.WriteCSS(&stylesBuffer, style)
	exportStyles := stylesBuffer.String()

	// TODO: Catch error here
	iterator, _ := lexer.Tokenise(nil, code)

	var htmlBuffer bytes.Buffer
	formatter.Format(&htmlBuffer, style, iterator)
	// TODO: Catch error here
	exportHTML := htmlBuffer.String()

	return exportStyles, exportHTML
}
