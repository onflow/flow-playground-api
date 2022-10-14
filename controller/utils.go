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
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/build"
	"github.com/onflow/cadence"
	"net/http"

	"github.com/go-chi/render"
)

type UtilsHandler struct{}

func NewUtilsHandler() *UtilsHandler {
	return &UtilsHandler{}
}

func (u *UtilsHandler) VersionHandler(w http.ResponseWriter, r *http.Request) {
	version := struct {
		API     string
		cadence string
	}{
		API:     "n/a",
		cadence: "n/a",
	}

	apiVer := build.Version()
	if apiVer != nil {
		version.API = apiVer.String()
	}

	cadenceVer := semver.MustParse(cadence.Version)
	if cadenceVer != nil {
		version.cadence = cadenceVer.String()
	}

	render.JSON(w, r, version)
}
