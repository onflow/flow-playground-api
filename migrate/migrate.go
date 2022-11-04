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

package migrate

// This package migrates users and projects from stable playground v1 to playground v2
// TODO: Migration entry points:
// TODO:    Create Project / Delete Project (needs to migrate the user if needed)
// TODO:    Opening a project needs to migrate the project (and user if needed)
// TODO:    ProjectList needs to migrate each project (and user if needed)
// TODO: Everything else can't really be called before these?

import (
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Migrator struct {
	store    storage.Store
	projects *controller.Projects
}

var V0 = semver.MustParse("v0.0.0")
var V1 = semver.MustParse("v1.0.0")
var V2 = semver.MustParse("v2.0.0")

func NewMigrator(store storage.Store, projects *controller.Projects) *Migrator {
	return &Migrator{
		store:    store,
		projects: projects,
	}
}

// MigrateProject migrates a project from one API version to another.
//
// This function only support projects upgrades (from lower to higher version).
//
// This function returns a boolean indicating whether the project was migrated or not.
func (m *Migrator) MigrateProject(id uuid.UUID, from, to *semver.Version) (bool, error) {
	from = sanitizeVersion(from)
	to = sanitizeVersion(to)

	if !from.LessThan(to) {
		// No migration work to do
		return false, nil
	}

	if from.LessThan(V1) {
		// Current version is too old to migrate to v2
		return false, errors.New("Current project version " + from.String() + " cannot be migrated to " + to.String())
	}

	if from.LessThan(V2) {
		err := m.migrateV1ProjectToV2(id)
		if err != nil {
			return false, errors.Wrapf(err, "failed to migrate project from %s to %s", from.String(), to.String())
		}
	}

	return true, nil
}

func (m *Migrator) MigrateUser(userID uuid.UUID, from, to *semver.Version) (bool, error) {
	return false, nil
}

func sanitizeVersion(version *semver.Version) *semver.Version {
	if version == nil {
		return V0
	}

	return version
}
