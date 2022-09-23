package sqlMigrator

import (
	"encoding/binary"
	"fmt"
	"github.com/dapperlabs/flow-playground-api/build"
	"github.com/dapperlabs/flow-playground-api/migrate/sqlMigrator/model"
	"github.com/google/uuid"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

const runPopulate = true
const numProjectsGen = 2

// Run $(gcloud beta emulators datastore env-init) to set env variables for datastore
func Test_RunSQLMigration(t *testing.T) {
	if runPopulate {
		populateDatastore()
	}
	main()
}

func populateDatastore() {
	fmt.Println("Populating datastore...")
	dstore := connectToDatastore()

	// Generate projects
	for i := 0; i < numProjectsGen; i++ {
		proj, ttpl, stpl := generateProject(i)
		err := dstore.CreateProject(proj, *ttpl, *stpl)
		if err != nil {

		}

		// Populate projects with random number of accounts
		accounts := generateAccounts(proj.ID)

		for _, account := range *accounts {
			err := dstore.InsertAccount(account)
			if err != nil {
				fmt.Println("Error: could not insert account into project", err)
			}
		}

		// Populate projects with transaction executions
		for _, exec := range *generateTransactionExecutions(proj) {
			err = dstore.InsertTransactionExecution(&exec)
			if err != nil {

			}
		}

		// Populate projects with script executions
		for _, exec := range *generateScriptExecutions(proj) {
			err = dstore.InsertScriptExecution(&exec)
			if err != nil {

			}
		}
	}
}

func generateProject(projectGenCount int) (*model.InternalProject, *[]*model.TransactionTemplate, *[]*model.ScriptTemplate) {
	proj := &model.InternalProject{
		ID:                        uuid.New(),
		UserID:                    uuid.New(),
		Secret:                    uuid.New(),
		PublicID:                  uuid.New(),
		ParentID:                  nil,
		Title:                     "Project number " + strconv.Itoa(projectGenCount),
		Description:               "",
		Readme:                    "",
		Seed:                      0,
		TransactionCount:          0,
		TransactionExecutionCount: rand.Intn(5),
		TransactionTemplateCount:  rand.Intn(5),
		ScriptTemplateCount:       rand.Intn(5),
		Persist:                   true,
		CreatedAt:                 time.Time{},
		UpdatedAt:                 time.Time{},
		Version:                   build.Version(),
	}

	// Populate project with transaction templates
	var ttpls []*model.TransactionTemplate
	for i := 0; i < proj.TransactionTemplateCount; i++ {
		ttpl := model.TransactionTemplate{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Title:  "Transaction Template " + strconv.Itoa(i),
			Index:  i,
			Script: "Test script",
		}

		ttpls = append(ttpls, &ttpl)
	}

	// Populate project with script templates
	var stpls []*model.ScriptTemplate
	for i := 0; i < proj.ScriptTemplateCount; i++ {
		stpl := model.ScriptTemplate{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Title:  "Script Template " + strconv.Itoa(i),
			Index:  i,
			Script: "Test script",
		}

		stpls = append(stpls, &stpl)
	}

	return proj, &ttpls, &stpls
}

// generateAccounts generates a random number of accounts up to 10
func generateAccounts(projID uuid.UUID) *[]*model.InternalAccount {
	var accounts []*model.InternalAccount

	for i := 0; i < rand.Intn(10); i++ {
		// Generate address
		bn := make([]byte, 8)
		binary.BigEndian.PutUint32(bn, uint32(i))
		addr := model.NewAddressFromBytes(bn)

		account := model.InternalAccount{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: projID,
			},
			Address:   addr,
			DraftCode: "test code " + strconv.Itoa(i),
		}

		accounts = append(accounts, &account)
	}
	return &accounts
}

func generateScriptExecutions(proj *model.InternalProject) *[]model.ScriptExecution {
	var scriptExecs []model.ScriptExecution
	for i := 0; i < proj.TransactionExecutionCount; i++ {
		scriptExecs = append(scriptExecs, model.ScriptExecution{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Index:     i,
			Script:    "Test script execution",
			Arguments: nil,
			Value:     "",
			Errors:    nil,
			Logs:      nil,
		})
	}
	return &scriptExecs
}

func generateTransactionExecutions(proj *model.InternalProject) *[]model.TransactionExecution {
	var txExecs []model.TransactionExecution
	for i := 0; i < proj.TransactionExecutionCount; i++ {
		txExecs = append(txExecs, model.TransactionExecution{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Index:     i,
			Script:    "Test transaction execution",
			Arguments: nil,
			Signers:   nil,
			Errors:    nil,
			Events:    nil,
			Logs:      nil,
		})
	}
	return &txExecs
}
