package migrate

import (
	"github.com/Masterminds/semver"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/model"
)

type Migrator struct {
	projects *controller.Projects
}

var V0 = semver.MustParse("v0.0.0")
var V0_1_0 = semver.MustParse("v0.1.0")

func NewMigrator(projects *controller.Projects) *Migrator {
	return &Migrator{
		projects: projects,
	}
}

// MigrateProject migrates a project from one API version to another.
//
// This function only support projects upgrades (from lower to higher version). Project
// downgrades are not yet supported.
func (m *Migrator) MigrateProject(id uuid.UUID, from, to *semver.Version) error {
	from = sanitizeVersion(from)
	to = sanitizeVersion(to)

	if !from.LessThan(to) {
		return nil
	}

	if from.LessThan(V0_1_0) {
		err := m.migrateToV0_1_0(id)
		if err != nil {
			return errors.Wrapf(err, "failed to migrate project from %s to %s", V0, V0_1_0)
		}
	}

	return nil
}

// migrateToV0_1_0 migrates a project from v0.0.0 to v0.1.0.
//
// Steps:
// - 1. Reset project state and recreate initial accounts
func (m *Migrator) migrateToV0_1_0(id uuid.UUID) error {
	proj := model.InternalProject{
		ID: id,
	}

	// Step 1/1
	err := m.projects.Reset(&proj)
	if err != nil {
		return errors.Wrap(err, "failed to reset project state")
	}

	return nil
}

func sanitizeVersion(version *semver.Version) *semver.Version {
	if version == nil {
		return V0
	}

	return version
}
