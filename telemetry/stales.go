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

package telemetry

import (
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/server/config"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"time"
)

// staleDuration is the amount a time before a project is considered stale if not accessed
var staleDuration = (time.Hour * 24) * time.Duration(config.Playground().StaleProjectDays)

// staleProjectScanner scans for stale projects in the database
var staleProjectScanner func(stale time.Duration, projs *[]*model.Project) error = nil

func SetStaleProjectScanner(scanner func(stale time.Duration, projs *[]*model.Project) error) {
	staleProjectScanner = scanner
}

// StaleProjectCounter returns the number of stale projects
func StaleProjectCounter() float64 {
	if staleProjectScanner == nil {
		sentry.CaptureException(errors.New("stale project scanner not set"))
		return 0
	}

	var stales []*model.Project
	err := staleProjectScanner(staleDuration, &stales)
	if err != nil {
		sentry.CaptureException(errors.Wrap(err, "failed to get stale projects"))
		return 0
	}

	return float64(len(stales))
}
