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

// Package build contains information about the build that is injected at build-time.
//
// To use this package, simply import it in your program, then add build
// arguments like the following:
//
//   go build -ldflags "-X github.com/dapperlabs/flow-playground-api/build.version=v1.0.0"
//
package build

import (
	"github.com/Masterminds/semver"
)

// Default value for build-time-injected version strings.
const undefined = "undefined"

// The following variables are injected at build-time using ldflags.
var version string

// Version returns the semantic version of this build.
func Version() *semver.Version {
	if version == undefined {
		return nil
	}

	return semver.MustParse(version)
}

// If any of the build-time-injected variables are empty at initialization,
// mark them as undefined.
func init() {
	if len(version) == 0 {
		version = undefined
	}
}
