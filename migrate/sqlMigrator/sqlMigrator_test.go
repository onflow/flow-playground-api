package sqlMigrator

import (
	"fmt"
	"github.com/dapperlabs/flow-playground-api/build"
	"github.com/dapperlabs/flow-playground-api/migrate/sqlMigrator/model"
	"github.com/google/uuid"
	"strconv"
	"testing"
	"time"
)

const runPopulate = false

// Run $(gcloud beta emulators datastore env-init) to set env variables for datastore
func Test_RunSQLMigration(t *testing.T) {
	if runPopulate {
		populateDatastore()
	}
	main()
}

func populateDatastore() {
	fmt.Println("Populating datastore...")
	dstore, err := connectToDatastore()
	if err != nil {
		fmt.Println("Error: could not connect to datastore", err)
		return
	}

	for i := 0; i < 10; i++ {
		proj, ttpl, stpl := generateProject(i)
		err = dstore.CreateProject(proj, *ttpl, *stpl)
	}
}

func generateProject(projectGenCount int) (*model.InternalProject, *[]*model.TransactionTemplate, *[]*model.ScriptTemplate) {
	projID := uuid.New()
	userID := uuid.New()
	Secret := uuid.New()
	PublicID := uuid.New()

	proj := &model.InternalProject{
		ID:                        projID,
		UserID:                    userID,
		Secret:                    Secret,
		PublicID:                  PublicID,
		ParentID:                  nil,
		Title:                     "Project number " + strconv.Itoa(projectGenCount),
		Description:               "",
		Readme:                    "",
		Seed:                      0,
		TransactionCount:          0,
		TransactionExecutionCount: 0,
		TransactionTemplateCount:  0,
		ScriptTemplateCount:       0,
		Persist:                   true,
		CreatedAt:                 time.Time{},
		UpdatedAt:                 time.Time{},
		Version:                   build.Version(),
	}

	var ttpls []*model.TransactionTemplate
	var stpls []*model.ScriptTemplate

	return proj, &ttpls, &stpls
}
