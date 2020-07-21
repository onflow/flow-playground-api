package migrate_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapperlabs/flow-playground-api/compute"
	"github.com/dapperlabs/flow-playground-api/controller"
	"github.com/dapperlabs/flow-playground-api/migrate"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage/memory"
)

const (
	numAccounts = 4
	cacheSize   = 256
)

func TestV0ToV0_1_0(t *testing.T) {
	store := memory.NewStore()
	computer, err := compute.NewComputer(cacheSize)
	require.NoError(t, err)

	scripts := controller.NewScripts(store, computer)
	projects := controller.NewProjects(migrate.V0, store, computer, numAccounts)

	m := migrate.NewMigrator(projects)

	user := model.User{
		ID: uuid.New(),
	}

	err = store.InsertUser(&user)
	require.NoError(t, err)

	proj, err := projects.Create(&user, model.NewProject{})
	require.NoError(t, err)

	assertAllAccountsExist(t, scripts, proj)

	migrated, err := m.MigrateProject(proj.ID, proj.Version, migrate.V0_1_0)
	require.NoError(t, err)
	assert.True(t, migrated)

	assertAllAccountsExist(t, scripts, proj)
}

func assertAllAccountsExist(t *testing.T, scripts *controller.Scripts, proj *model.InternalProject) {
	for i := 0; i < numAccounts; i++ {
		script := fmt.Sprintf(`pub fun main() { getAccount(0x0%d) }`, i+1)

		result, err := scripts.CreateExecution(proj, script)
		require.NoError(t, err)

		assert.Nil(t, result.Error)
	}
}
