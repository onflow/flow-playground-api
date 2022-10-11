package main

import (
	"github.com/dapperlabs/flow-playground-api/build"
	"github.com/dapperlabs/flow-playground-api/migrate/database/model"
	"github.com/dapperlabs/flow-playground-api/telemetry"
	"github.com/google/uuid"
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
	telemetry.DebugLog("Populating datastore...")
	dstore := connectToDatastore()

	// Generate projects
	for i := 0; i < numProjectsGen; i++ {
		proj, ttpl, stpl := generateProject(i)
		err := dstore.CreateProject(proj, *ttpl, *stpl)
		if err != nil {
			telemetry.DebugLog("Error: could not populate project " + err.Error())
		}

		// Populate projects with accounts
		accounts := generateAccounts(proj.ID)

		for _, account := range *accounts {
			err := dstore.InsertAccount(account)
			if err != nil {
				telemetry.DebugLog("Error: could not populate accounts " + err.Error())
			}
		}

		// Populate projects with transaction executions
		for _, exec := range *generateTransactionExecutions(proj) {
			err = dstore.InsertTransactionExecution(&exec)
			if err != nil {
				telemetry.DebugLog("Error: could not populate transaction executions " + err.Error())
			}
		}

		// Populate projects with script executions
		for _, exec := range *generateScriptExecutions(proj) {
			err = dstore.InsertScriptExecution(&exec)
			if err != nil {
				telemetry.DebugLog("Error: could not populate script executions " + err.Error())
			}
		}

		// Add a project user
		err = dstore.InsertUser(generateUser())
		if err != nil {
			telemetry.DebugLog("Error: could not populate user " + err.Error())
		}
	}
}

func generateUser() *model.User {
	return &model.User{ID: uuid.New()}
}

func generateProject(projectGenCount int) (*model.InternalProject, *[]*model.TransactionTemplate, *[]*model.ScriptTemplate) {
	proj := &model.InternalProject{
		ID:                        uuid.New(),
		UserID:                    uuid.New(),
		Secret:                    uuid.New(),
		PublicID:                  uuid.New(),
		ParentID:                  nil,
		Title:                     "Project number " + strconv.Itoa(projectGenCount),
		Description:               "Project description",
		Readme:                    "Project readme",
		Seed:                      0,
		TransactionCount:          0,
		TransactionExecutionCount: 4,
		TransactionTemplateCount:  5,
		ScriptTemplateCount:       6,
		Persist:                   true,
		CreatedAt:                 time.Time{},
		UpdatedAt:                 time.Time{},
		Version:                   build.Version(),
	}

	// Populate project with transaction templates
	var ttpls []*model.TransactionTemplate
	for i := 0; i < proj.TransactionTemplateCount; i++ {
		ttpls = append(ttpls, &model.TransactionTemplate{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Title:  "Transaction Template " + strconv.Itoa(i),
			Index:  i,
			Script: "Test script",
		})
	}

	// Populate project with script templates
	var stpls []*model.ScriptTemplate
	for i := 0; i < proj.ScriptTemplateCount; i++ {
		stpls = append(stpls, &model.ScriptTemplate{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Title:  "Script Template " + strconv.Itoa(i),
			Index:  i,
			Script: "Test script",
		})
	}

	return proj, &ttpls, &stpls
}

// generateAccounts generates accounts for a project
func generateAccounts(projID uuid.UUID) *[]*model.InternalAccount {
	var accounts []*model.InternalAccount

	for i := 0; i < 10; i++ {
		accounts = append(accounts, &model.InternalAccount{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: projID,
			},
			Address:   [8]byte{0x01},
			DraftCode: "test code " + strconv.Itoa(i),
		})
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
			Arguments: []string{"arg1", "arg2", "arg3"},
			Value:     "test value",
			Errors: []model.ProgramError{
				{
					Message: "Program error 1",
					StartPosition: &model.ProgramPosition{
						Offset: 10,
						Line:   11,
						Column: 12,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 13,
						Line:   14,
						Column: 15,
					},
				},
				{
					Message: "Program error 2",
					StartPosition: &model.ProgramPosition{
						Offset: 16,
						Line:   17,
						Column: 18,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 19,
						Line:   20,
						Column: 21,
					},
				},
			},
			Logs: []string{"log1", "log2", "log3"},
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
			Arguments: []string{"test"},
			Signers: []model.Address{
				[8]byte{0x01},
				[8]byte{0x02},
				[8]byte{0x03},
			},
			Errors: []model.ProgramError{
				{
					Message: "Test error 1",
					StartPosition: &model.ProgramPosition{
						Offset: 10,
						Line:   11,
						Column: 12,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 13,
						Line:   14,
						Column: 15,
					},
				},
				{
					Message: "Test error 2",
					StartPosition: &model.ProgramPosition{
						Offset: 15,
						Line:   16,
						Column: 17,
					},
					EndPosition: &model.ProgramPosition{
						Offset: 18,
						Line:   19,
						Column: 20,
					},
				},
			},
			Events: []model.Event{
				{
					Type:   "Test event 1",
					Values: []string{"val1", "val2", "val3"},
				},
				{
					Type:   "Test event 2",
					Values: []string{"val4", "val5", "val6"},
				},
			},
			Logs: []string{"Log 1", "Log 2", "Log 3"},
		})
	}
	return &txExecs
}
