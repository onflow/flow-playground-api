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
	"github.com/robfig/cron"
	"time"
)

// staleDuration is the amount a time before a project is considered stale if not accessed
var staleDuration = (time.Hour * 24) * time.Duration(config.Playground().StaleProjectDays)

// staleProjectScanner scans for stale projects in the database
var staleProjectScanner func(stale time.Duration, projs *[]*model.Project) error = nil

// totalProjectCounter queries for the total number of projects in the database
var totalProjectCounter func(totalProjects *int64) error = nil

func SetStaleProjectScanner(scanner func(stale time.Duration, projs *[]*model.Project) error) {
	staleProjectScanner = scanner
}

func SetTotalProjectCounter(counter func(totalProjects *int64) error) {
	totalProjectCounter = counter
}

// StaleProjectCounter returns the number of stale projects
func StaleProjectCounter() (int, error) {
	if staleProjectScanner == nil {
		return 0, errors.New("stale project scanner not set")
	}

	var stales []*model.Project
	err := staleProjectScanner(staleDuration, &stales)
	if err != nil {
		return 0, err
	}

	return len(stales), nil
}

func TotalProjectCounter() (int, error) {
	if totalProjectCounter == nil {
		return 0, errors.New("total project counter not set")
	}

	var count int64
	err := totalProjectCounter(&count)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// registerStaleProjectJob registers recurring database query for project metrics
func registerProjectJobs() error {
	job := cron.New()

	err := job.AddFunc(
		config.Telemetry().ProjectQueryTime,
		func() {
			staleCount, err := StaleProjectCounter()
			if err != nil {
				sentry.CaptureException(err)
				return
			}
			staleProjectGauge.Set(float64(staleCount))

			totalProjects, err := TotalProjectCounter()
			if err != nil {
				sentry.CaptureException(err)
				return
			}
			totalProjectGauge.Set(float64(totalProjects))
		},
	)
	if err != nil {
		return err
	}

	job.Start()
	job.Run()

	return nil
}
