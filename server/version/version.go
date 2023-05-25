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

package version

import (
	"errors"
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/build"
	"github.com/go-chi/render"
	"github.com/icza/bitio"
	"github.com/onflow/cadence"
	"net/http"
	"runtime/debug"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	version := struct {
		API      string
		Cadence  string
		Emulator string
	}{
		API:      "n/a",
		Cadence:  "n/a",
		Emulator: "n/a",
	}

	apiVer := build.Version()
	if apiVer != nil {
		version.API = apiVer.String()
	}

	cadenceVer := semver.MustParse(cadence.Version)
	if cadenceVer != nil {
		version.Cadence = cadenceVer.String()
	}

	emulatorVer, err := getDependencyVersion("github.com/onflow/flow-emulator")
	if err == nil {
		version.Emulator = semver.MustParse(emulatorVer).String()
	}

	render.JSON(w, r, version)
}

func getDependencyVersion(path string) (string, error) {
	_ = bitio.NewReader
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "", errors.New("failed to read build info")
	}

	for _, dep := range bi.Deps {
		if dep.Path == path {
			return dep.Version, nil
		}
	}

	return "", errors.New("dependency not found")
}
