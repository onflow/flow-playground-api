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
//
// This function returns a boolean indicating whether or not the project was migrated.
func (m *Migrator) MigrateProject(id uuid.UUID, from, to *semver.Version) (bool, error) {
	from = sanitizeVersion(from)
	to = sanitizeVersion(to)

	if !from.LessThan(to) {
		return false, nil
	}

	if from.LessThan(V0_1_0) {
		err := m.migrateToV0_1_0(id)
		if err != nil {
			return false, errors.Wrapf(err, "failed to migrate project from %s to %s", V0, V0_1_0)
		}
	}

	// If no migration steps are left, set project version to latest.
	err := m.setProjectVersion(id, to)
	if err != nil {
		return false, err
	}

	return true, nil
}

// migrateToV0_1_0 migrates a project from v0.0.0 to v0.1.0.
//
// Steps:
// - 1. Reset project state and recreate initial accounts
// - 2. Update project version tag
func (m *Migrator) migrateToV0_1_0(id uuid.UUID) error {
	proj := model.InternalProject{
		ID: id,
	}

	// TODO:
	//  Update storage interface to allow atomic transactions.
	//  Ideally the project state should be wiped and the version incremented in the
	//  same transaction.

	// Step 1/2
	err := m.projects.Reset(&proj)
	if err != nil {
		return errors.Wrap(err, "failed to reset project state")
	}

	// Step 2/2
	err = m.projects.UpdateVersion(id, V0_1_0)
	if err != nil {
		return errors.Wrap(err, "failed to update project version")
	}

	return nil
}

func (m *Migrator) setProjectVersion(id uuid.UUID, version *semver.Version) error {
	err := m.projects.UpdateVersion(id, version)
	if err != nil {
		return errors.Wrap(err, "failed to update project version")
	}

	return nil
}

func sanitizeVersion(version *semver.Version) *semver.Version {
	if version == nil {
		return V0
	}

	return version
}
