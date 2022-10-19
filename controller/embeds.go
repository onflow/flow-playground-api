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

package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/google/uuid"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Snippet struct {
	html      string
	styles    string
	themeName string
}

type EmbedsHandler struct {
	store             storage.Store
	playgroundBaseURL string
}

func NewEmbedsHandler(
	store storage.Store,
	playgroundBaseURL string,
) *EmbedsHandler {
	return &EmbedsHandler{
		store:             store,
		playgroundBaseURL: playgroundBaseURL,
	}
}

func (e *EmbedsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get project UUID and check that it's not empty
	projectID, projectIDErr := getUUID("project", r)
	if projectIDErr != nil {
		http.Error(w, "invalid project ID", http.StatusBadRequest)
		return
	}

	// Get child UUID - will be used to find model - and check that it's not empty
	scriptID, scriptIDErr := getUUID("id", r)
	if scriptIDErr != nil {
		http.Error(w, "invalid script ID", http.StatusBadRequest)
		return
	}

	// Get script type - account code is retrieved in a different way - and check that it's not empty
	scriptType, scriptTypeErr := getURLParam("type", r)
	if scriptTypeErr != nil {
		http.Error(w, "invalid script type", http.StatusBadRequest)
		return
	}

	// Get script code
	code, getErr := e.GetCode(scriptID, projectID, scriptType)
	if getErr != nil {
		http.Error(w, "could not get script with specified parameters", http.StatusBadRequest)
		return
	}

	// Get theme - if any - from url
	theme := r.URL.Query().Get("theme")

	// Use chroma to return CSS and HTML blocks with theme name
	snippet, snippetErr := createSnippet(code, theme)
	if snippetErr != nil {
		http.Error(w, "can't create snippet with specified parameters", http.StatusBadRequest)
		return
	}

	playgroundURL := fmt.Sprintf("%s/%s?type=%s&id=%s", e.playgroundBaseURL, projectID.String(), scriptType, scriptID)

	// Create injectable Javascript blocks, which will be written in response
	wrapperStyleInjection := createCodeStyles(snippet.styles, snippet.themeName)
	snippetStyleInjection := createSnippetStyles()
	htmlInjection := wrapCodeBlock(snippet.html, snippet.themeName, playgroundURL)

	w.Header().Set("Content-Type", "application/javascript")
	_, err := w.Write([]byte(wrapperStyleInjection))
	if err != nil {
		return
	}
	_, err = w.Write([]byte(snippetStyleInjection))
	if err != nil {
		return
	}
	_, err = w.Write([]byte(htmlInjection))
	if err != nil {
		return
	}
}

func (e *EmbedsHandler) GetCode(id, pID uuid.UUID, scriptType string) (string, error) {

	switch scriptType {
	case "script":
		return e.GetScriptTemplate(id, pID)
	case "transaction":
		return e.GetTransactionTemplate(id, pID)
	case "contract":
		return e.GetContractTemplate(id, pID)
	default:
		return "", fmt.Errorf("invalid script type: %s", scriptType)
	}

}

func (e *EmbedsHandler) GetScriptTemplate(id, pID uuid.UUID) (string, error) {
	var tmpl model.ScriptTemplate

	err := e.store.GetFile(id, pID, &tmpl)
	if err != nil {
		return "", err
	}

	return tmpl.Script, nil
}

func (e *EmbedsHandler) GetTransactionTemplate(id, pID uuid.UUID) (string, error) {
	var tmpl model.TransactionTemplate

	err := e.store.GetFile(id, pID, &tmpl)
	if err != nil {
		return "", err
	}

	return tmpl.Script, nil
}

func (e *EmbedsHandler) GetContractTemplate(id, pID uuid.UUID) (string, error) {
	var tmpl model.ContractTemplate

	err := e.store.GetFile(id, pID, &tmpl)
	if err != nil {
		return "", err
	}

	return tmpl.Script, nil
}

func getUUID(paramName string, r *http.Request) (id uuid.UUID, err error) {
	rawID, err := getURLParam(paramName, r)
	if err != nil {
		return uuid.Nil, err
	}
	return model.UnmarshalUUID(rawID)
}

func getURLParam(paramName string, r *http.Request) (string, error) {
	param := r.URL.Query().Get(paramName)
	if param == "" {
		return "", fmt.Errorf("failed to decode URL param %s: can't be empty", paramName)
	}
	return param, nil
}

func createSnippet(code string, theme string) (Snippet, error) {
	lexer := lexers.Get("swift")
	lexer = chroma.Coalesce(lexer)

	style := styles.Get(theme)
	if style == nil {
		style = styles.Fallback
	}

	formatter := html.New(
		html.WithClasses(true),
		html.WithLineNumbers(true),
	)

	var stylesBuffer bytes.Buffer
	cssError := formatter.WriteCSS(&stylesBuffer, style)
	if cssError != nil {
		return Snippet{}, cssError
	}

	exportStyles := stylesBuffer.String()

	iterator, lexerErr := lexer.Tokenise(nil, code)
	if lexerErr != nil {
		return Snippet{}, lexerErr
	}

	var htmlBuffer bytes.Buffer
	formatterError := formatter.Format(&htmlBuffer, style, iterator)
	if formatterError != nil {
		return Snippet{}, formatterError
	}
	exportHTML := htmlBuffer.String()

	snippet := Snippet{
		html:      exportHTML,
		styles:    exportStyles,
		themeName: style.Name,
	}
	return snippet, nil
}

func createCodeStyles(styles string, styleName string) string {
	// We need to prepend styles strings with "`" in order for it to work inside of embedded javascript
	codeStyles := fmt.Sprintf("`%s`", styles)
	// This will allow multiple styles per page, if user desires such outcome
	codeStyles = strings.ReplaceAll(codeStyles, ".chroma", ".chroma."+styleName)

	stylesInjection := fmt.Sprintf(`
	if (!document.getElementById("theme-%s")){
		var newStyleTag = document.createElement('style');
		newStyleTag.id = "theme-%s"
		newStyleTag.innerHTML = %s
		document.head.appendChild(newStyleTag);
	}
	`, styleName, styleName, codeStyles)

	return stylesInjection
}

func generateWrapperStyles() string {
	wrapperStyles := `
	.cadence-snippet{
		width: 100%;
		overflow: hidden;
		border-radius: 5px;
		box-shadow:0 3px 5px rgba(0,0,0,0.1), 0 0 0 1px #ccc;
		margin-bottom: 1.5em;
	}
	
	.cadence-snippet pre.chroma{
		margin: 0;
		margin-bottom: 0;
		padding: 20px;
		overflow-x: auto;
	}
	
	.cadence-code-block .chroma{
		padding: 1em;
	}
	
	.cadence-info-block{
		padding: 12px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		border-top: 1px solid #ccc;
	}
	
	
	.cadence-info-block img{
		height: 24px;
		width: auto;
	}
	
	.cadence-info-block a{
		display: flex;
		align-items: center;
		width: 100%;
		justify-content: flex-start;
		text-decoration: none;
		font-family: sans-serif;
		color: #231f20;
		font-size: 14px;
		color: #5a6270;
	}
	
	.flow-playground-logo {
		background-image: url(data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPHN2ZyB3aWR0aD0iMTE0cHgiIGhlaWdodD0iMTE0cHgiIHZpZXdCb3g9IjAgMCAxMTQgMTE0IiB2ZXJzaW9uPSIxLjEiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGxpbmsiPgogICAgPHRpdGxlPnBsYXlncm91bmQuc3ZnPC90aXRsZT4KICAgIDxkZWZzPgogICAgICAgIDxsaW5lYXJHcmFkaWVudCB4MT0iMjUwLjM3NjkxOSUiIHkxPSItMTM3LjUzNDI2NSUiIHgyPSIxMS43MzU4ODI3JSIgeTI9IjkwLjczNDY0OTElIiBpZD0ibGluZWFyR3JhZGllbnQtMSI+CiAgICAgICAgICAgIDxzdG9wIHN0b3AtY29sb3I9IiNGRkZGRkYiIHN0b3Atb3BhY2l0eT0iMCIgb2Zmc2V0PSIwJSI+PC9zdG9wPgogICAgICAgICAgICA8c3RvcCBzdG9wLWNvbG9yPSIjRkE2QjAwIiBvZmZzZXQ9IjEyLjgzMjk4NDMlIj48L3N0b3A+CiAgICAgICAgICAgIDxzdG9wIHN0b3AtY29sb3I9IiNGRkNGMDAiIG9mZnNldD0iMTAwJSI+PC9zdG9wPgogICAgICAgIDwvbGluZWFyR3JhZGllbnQ+CiAgICAgICAgPGNpcmNsZSBpZD0icGF0aC0yIiBjeD0iNTciIGN5PSI1NyIgcj0iNTciPjwvY2lyY2xlPgogICAgPC9kZWZzPgogICAgPGcgaWQ9IlBhZ2UtMSIgc3Ryb2tlPSJub25lIiBzdHJva2Utd2lkdGg9IjEiIGZpbGw9Im5vbmUiIGZpbGwtcnVsZT0iZXZlbm9kZCI+CiAgICAgICAgPGcgaWQ9IkZsb3ctUGxheWdyb3VuZCI+CiAgICAgICAgICAgIDxjaXJjbGUgaWQ9IkNvbG9yIiBmaWxsPSJ1cmwoI2xpbmVhckdyYWRpZW50LTEpIiBjeD0iNTciIGN5PSI1NyIgcj0iNTciPjwvY2lyY2xlPgogICAgICAgICAgICA8ZyBpZD0iYmVhY2gtKDEpIj4KICAgICAgICAgICAgICAgIDxtYXNrIGlkPSJtYXNrLTMiIGZpbGw9IndoaXRlIj4KICAgICAgICAgICAgICAgICAgICA8dXNlIHhsaW5rOmhyZWY9IiNwYXRoLTIiPjwvdXNlPgogICAgICAgICAgICAgICAgPC9tYXNrPgogICAgICAgICAgICAgICAgPGcgaWQ9Ik92YWwiPjwvZz4KICAgICAgICAgICAgICAgIDxwYXRoIGQ9Ik04MC42NzM1Mjk3LDkzLjI5NjIwNiBDNzIuMDg1NDcyOSw4OC41MjYwMjI2IDYyLjU4MTcyMDksODYuMTA0ODkwMyA1My4wNTA4MzU4LDg2LjE1Mjk0NDUgTDU5Ljc2MTg2MTIsNTUuODEyOTA0MiBMNjkuMzc3Miw1NC4wNDE5OTIyIEM2OS43OTMwMDA0LDQ3LjU3MDgwMDggNzAuMTk1MzEyNSw0My4xNzE4NzUgNzAuMTk1MzEyNSwzNS40OTc1NTg2IEM3Ny40MTY3NDUzLDM5Ljk5MTIxMDkgODMuMjI2MTkxOCw0Ny41NzA4MDA4IDg2LjM4NDc2NTYsNTggQzc1LjA0NjM4ODYsNTQuNTkwODIwMyA2OS4zNzcyLDUyLjg4NjIzMDUgNjkuMzc3Miw1Mi44ODYyMzA1IEM2OS4zNzcyLDUyLjg4NjIzMDUgNTkuMDM5MzMxNiw1NS42MTQ0ODAzIDU5Ljc2MTg2MTIsNTUuODEyOTA0MiBMOTAuODgxMzAzMiw2NC4zNTk1NDk2IEM5MC45ODgwMzg1LDY0LjM4ODc4MjUgOTEuMDk2MjExMyw2NC40MDMxOTg4IDkxLjIwMzY2NTMsNjQuNDAzMTk4OCBDOTEuNTI2MjA3MSw2NC40MDMxOTg4IDkxLjg0MjI4MDEsNjQuMjc0MDUzMSA5Mi4wOTA2MTAzLDY0LjAzMjM4MDMgQzkyLjQyMTU5NzUsNjMuNzEwMDE2NSA5Mi41OTA4NjQ1LDYzLjIyODA3MjYgOTIuNTQzNzg2LDYyLjc0MDkyMjkgQzkxLjg1MjE2Myw1NS41NzM0MzQgODguNTQzMTg5OCw0OC4xMjMwMjU4IDgzLjIyNjE5MTgsNDEuNzYxNjQ2NyBDNzguMjQ4MjY3LDM1LjgwNTkyNTQgNzIuMDM3ODU1MywzMS40NDMyMDE4IDY1LjYxMjcxNTIsMjkuMzYxMjUyMiBMNjYuMTY1NjE4MywyNi44NjEyMzA4IEM2Ni4zNDM2OTAxLDI2LjA1NTkyMiA2NS45MDIxOTQyLDI1LjI0MjIwMzcgNjUuMTc5NjY0NiwyNS4wNDM3Nzk3IEM2NC40NTcxMzUxLDI0Ljg0NTk1NjUgNjMuNzI2Njk5MywyNS4zMzcxMTA4IDYzLjU0ODYyNzUsMjYuMTQyNDE5NiBMNjIuOTk1NTQ0NywyOC42NDI0NDEgQzU2LjQzNzk3MzgsMjcuMTU4NzY2NiA0OS4xMTk0MTk5LDI3LjgwNTg5NjkgNDIuMjI3ODA3MSwzMC41MDE3MzkzIEMzNC44NjcwMjYzLDMzLjM4MTE4ODkgMjguODI5NjU1LDM4LjI2NDkwMDQgMjUuMjI4MTQ4MSw0NC4yNTMyNTg2IEMyNC45ODM0MTE3LDQ0LjY2MDExNzggMjQuOTMyMjAwMyw0NS4xNzQ0OTgzIDI1LjA5MDg2NTcsNDUuNjMxNDEzOSBDMjUuMjQ5NTMxMSw0Ni4wODgzMjk2IDI1LjU5ODMwNzUsNDYuNDMwMTE1MyAyNi4wMjUwNjg5LDQ2LjU0NzQ0NzcgTDU3LjE0NDY5MDYsNTUuMDk0MDkzIEw1MC4yNTU3NzMyLDg2LjIzODQ0MSBDNDcuMDYzOTU3Miw4Ni40MTY4NDIzIDQzLjg3ODc4OTcsODYuODcwMzU0MSA0MC43MzM2OTI5LDg3LjYxMTM5MDQgQzQwLjAwNTQxMzMsODcuNzgyOTg0MSAzOS41NDAwMTg4LDg4LjU3OTg4MzQgMzkuNjkzODMyNiw4OS4zOTExOTkgQzM5Ljg0NzY0NjQsOTAuMjAyNTE0NiA0MC41NjE5MTAyLDkwLjcyMDg5OTcgNDEuMjkxMDg4Miw5MC41NDk5MDY3IEM0NC4wMzA2MjcsODkuOTA0Mzc4MiA0Ni44MDIxNTAzLDg5LjQ4NjMwNjQgNDkuNTgxMjIwNiw4OS4yODc2ODIyIEw1Mi4zODU2MjcxLDg5LjE2MTUzOTkgQzYxLjcyMzg4NTUsODguOTkzNzUwNSA3MS4wNTcyOTIzLDkxLjMxMDM2NDkgNzkuNDcxNDEwMSw5NS45ODQyMzk2IEwxMDAuODE5MDA0LDEwNy44NDIwMjIgQzEwMS4wMTIxNywxMDcuOTQ5MzQzIDEwMS4yMTcxOTUsMTA4IDEwMS40MTkxNjYsMTA4IEMxMDEuOTE0MDI5LDEwOCAxMDIuMzkwNTY0LDEwNy42OTUwNTYgMTAyLjYyNjMxNiwxMDcuMTY4MDYxIEMxMDIuOTU4MzgyLDEwNi40MjU4MjMgMTAyLjY4NzU5LDEwNS41MjQyMDYgMTAyLjAyMTQ4MywxMDUuMTUzOTg4IEM5MS4zNDkwNDgzLDk5LjIyMjMyMTkgOTEuMzQ3NTQ3Miw5OS4yMjUwMjM1IDgwLjY3MzUyOTcsOTMuMjk2MjA2IFogTTUyLjcyODUxNTYsNDguNTMzNjkxNCBDNTcuNjk1MzEyNSwzNi4wMjA5OTYxIDYxLjIyOTY4NzIsMzIuMTM3NzQ1MiA2My42MTg2NTIzLDMyLjc4NjEzMjggQzY2LjAwNzYxNzUsMzMuNDMxNDkxNSA2Ni4yMDYyOTg2LDM4LjY3OTgxNDkgNjQuNzY1NjI1LDUyLjAyMjk0OTIgTDUyLjcyODUxNTYsNDguNTMzNjkxNCBaIE0zMi4xNTgyMDMxLDQyLjc3OTc4NTIgQzM4LjA3NzYzNjcsMzQuNzAzNjEzMyA0Ni4yNjE3MTg4LDMyLjUxNTYyNSA1Ni4xNDU1MDc4LDMyLjUxNTYyNSBDNTUuMjQwMjYyNiwzMy40MTc4Mjk2IDQ5LjUzMTA4ODksNDEuMzQ3MjY5IDQ3LjU2NDQ1MzEsNDcuMDM3MTA5NCBMMzIuMTU4MjAzMSw0Mi43Nzk3ODUyIFoiIGlkPSJTaGFwZSIgZmlsbD0iI0ZGRkZGRiIgZmlsbC1ydWxlPSJub256ZXJvIiBtYXNrPSJ1cmwoI21hc2stMykiPjwvcGF0aD4KICAgICAgICAgICAgICAgIDxwYXRoIGQ9Ik0xMTYuMjA0Nzg0LDExMy4wNjc4MjIgTDEwNy4wNzYyMDksMTA4LjE2MjkxIEMxMDYuMzY2NTYyLDEwNy43ODE0ODQgMTA1LjUwNDk1NiwxMDguMDkyNzcyIDEwNS4xNTEwOSwxMDguODU3Mjc2IEMxMDQuNzk3NDE0LDEwOS42MjIxOTIgMTA1LjA4NTgyOCwxMTAuNTUwOTAxIDEwNS43OTUyODQsMTEwLjkzMjMyOCBMMTE0LjkyMzg2LDExNS44MzcyMzkgQzExNS4xMjk1OTYsMTE1Ljk0NzgwOSAxMTUuMzQ3OTY0LDExNiAxMTUuNTYzMDc4LDExNiBDMTE2LjA5MDE0NiwxMTYgMTE2LjU5NzY5MywxMTUuNjg1ODIzIDExNi44NDg3ODgsMTE1LjE0Mjg3MyBDMTE3LjIwMjY1NCwxMTQuMzc3OTU3IDExNi45MTQyNCwxMTMuNDQ5MDQyIDExNi4yMDQ3ODQsMTEzLjA2NzgyMiBMMTE2LjIwNDc4NCwxMTMuMDY3ODIyIFoiIGlkPSJQYXRoIiBmaWxsPSIjRkZGRkZGIiBmaWxsLXJ1bGU9Im5vbnplcm8iIG1hc2s9InVybCgjbWFzay0zKSI+PC9wYXRoPgogICAgICAgICAgICAgICAgPHBhdGggZD0iTTM1LjEyNDcxNDYsOTAuMDY1OTA2MiBDMzMuNjYyNDQxLDkwLjUxMTA3NTQgMzIuMTk4NDQ3OSw5MS4wMTE0ODcxIDMwLjc3MzA0NjMsOTEuNTUzMzMwNyBMMjUuODk1NDI1Nyw5My40MDc1NDQ1IEMyNS4xNjE5OTYzLDkzLjY4NjQ0NzggMjQuODA4MTc3Myw5NC40NzA2MDU1IDI1LjEwNTA2NDMsOTUuMTU5MzQ0MiBDMjUuMzMwNjkwOCw5NS42ODI1MzQ1IDI1Ljg2NzE1MDcsOTYgMjYuNDMzNjA1MSw5NiBDMjYuNjEyNjE2MSw5NiAyNi43OTQ2ODM5LDk1Ljk2ODI1MzUgMjYuOTcwODI5Miw5NS45MDEzNTI1IEwzMS44NDg2NDA5LDk0LjA0NzEzODggQzMzLjIxMTE4OCw5My41MjkxNDk5IDM0LjYxMDc5ODIsOTMuMDUwNzk5MyAzNi4wMDg2ODksOTIuNjI1MTgwMyBDMzYuNzYxNDE0Miw5Mi4zOTU5NTk1IDM3LjE3Mzg4NDYsOTEuNjM3MjcwOCAzNi45Mjk3MjY2LDkwLjkzMDU5NjEgQzM2LjY4NTc1OTUsOTAuMjIzOTIxNSAzNS44Nzc4MjE4LDg5LjgzNjg2NDcgMzUuMTI0NzE0Niw5MC4wNjU5MDYyIFoiIGlkPSJQYXRoIiBmaWxsPSIjRkZGRkZGIiBmaWxsLXJ1bGU9Im5vbnplcm8iIG1hc2s9InVybCgjbWFzay0zKSI+PC9wYXRoPgogICAgICAgICAgICAgICAgPHBhdGggZD0iTTkxLjUwNjAzNTYsMTA3IEw4OS40OTM5NjQ0LDEwNyBDODguNjY4ODk3NywxMDcgODgsMTA3LjY3MTYgODgsMTA4LjUgQzg4LDEwOS4zMjg0IDg4LjY2ODg5NzcsMTEwIDg5LjQ5Mzk2NDQsMTEwIEw5MS41MDYwMzU2LDExMCBDOTIuMzMxMTAyMywxMTAgOTMsMTA5LjMyODQgOTMsMTA4LjUgQzkzLDEwNy42NzE2IDkyLjMzMTEwMjMsMTA3IDkxLjUwNjAzNTYsMTA3IFoiIGlkPSJQYXRoIiBmaWxsPSIjRkZGRkZGIiBmaWxsLXJ1bGU9Im5vbnplcm8iIG1hc2s9InVybCgjbWFzay0zKSI+PC9wYXRoPgogICAgICAgICAgICAgICAgPHBhdGggZD0iTTEwLjY5OTIxODgsMTAyLjMwMzcxMSBDMTIuOTEwODA3Myw5OS44MjU4NDY0IDE3LjY3NzczNDQsOTcuMDU3OTQyNyAyNSw5NCBDMzUuOTgzMzk4NCw4OS40MTMwODU5IDQ0LjcwMDE5NTMsODYuNzM2MzI4MSA2MS4wNzcxNDg0LDg5LjQxMzA4NTkgQzcxLjk5NTExNzIsOTEuMTk3NTkxMSA4Ni42MzYwNjc3LDk4LjM5MzIyOTIgMTA1LDExMSIgaWQ9IlBhdGgtMiIgc3Ryb2tlPSIjRkZGRkZGIiBzdHJva2Utd2lkdGg9IjYiIG1hc2s9InVybCgjbWFzay0zKSI+PC9wYXRoPgogICAgICAgICAgICAgICAgPGxpbmUgeDE9IjU4LjUiIHkxPSI1NC4wNDE5OTIyIiB4Mj0iNTAuNjg0MDgyIiB5Mj0iODguMzQ3NzE5NSIgaWQ9IlBhdGgtMyIgc3Ryb2tlPSIjRkZGRkZGIiBzdHJva2Utd2lkdGg9IjUiIG1hc2s9InVybCgjbWFzay0zKSI+PC9saW5lPgogICAgICAgICAgICA8L2c+CiAgICAgICAgPC9nPgogICAgPC9nPgo8L3N2Zz4=);
	}
	
	.cadence-info-block .umbrella{
		height: 24px;
		width: 24px;
		background-size: 100%;
		margin-right: 5px;
		display: inline-block;
	}
	`
	return wrapperStyles
}

// TODO: those styles can be served via CSS file over CDN
//  this will enable browsers caching, smaller response and easier update
//  especially with big background-image for Playground umbrella logo
func createSnippetStyles() string {
	wrapperStyles := generateWrapperStyles()

	wrapperStyles = fmt.Sprintf("`%s`", wrapperStyles)

	stylesInjection := fmt.Sprintf(`
		if (!document.getElementById("cadence-styles")){
			var newStyleTag = document.createElement('style');
			newStyleTag.id = "cadence-styles"
			newStyleTag.innerHTML = %s
			document.head.appendChild(newStyleTag);
		}
	`, wrapperStyles)

	return stylesInjection
}

func wrapCodeBlock(htmlBlock string, styleName string, playgroundUrl string) string {
	sourceCode := strings.Replace(htmlBlock, "chroma", "chroma "+styleName, 1)

	wrapper := `
	<div class="cadence-snippet">
		<div class="cadence-code-block">
			%s
		</div>
		<div class="cadence-info-block">
			<a href="%s">
				<div class="flow-playground-logo umbrella"></div>
				<span>View on Playground</span>
			</a>
		</div>
	</div>
	`
	wrapper = fmt.Sprintf(wrapper, sourceCode, playgroundUrl)
	wrapper = fmt.Sprintf("document.write(`%s`)", wrapper)
	return wrapper
}
