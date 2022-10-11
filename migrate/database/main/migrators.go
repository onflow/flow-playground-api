package main

// Migrator functions to move data from datastore to sql database

import (
	"github.com/dapperlabs/flow-playground-api/migrate/database/model"
	"github.com/dapperlabs/flow-playground-api/migrate/database/storage/datastore"
	sqlModel "github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/telemetry"
	"github.com/google/uuid"
	"strings"
)

// migrateAccounts migrates models of datastore accounts to sql accounts
func migrateAccounts(dstore *datastore.Datastore, sqlDB *storage.SQL, projID uuid.UUID) {
	var accounts []*model.InternalAccount
	err := dstore.GetAccountsForProject(projID, &accounts)
	if err != nil {
		numErrors++
		return
	}

	if len(accounts) == 0 {
		return
	}

	var sqlAccounts []*sqlModel.Account
	for _, acc := range accounts {
		sqlAccounts = append(sqlAccounts, &sqlModel.Account{
			ID:                acc.ProjectChildID.ID,
			ProjectID:         acc.ProjectChildID.ProjectID,
			Index:             acc.Index,
			Address:           sqlModel.Address(acc.Address),
			DraftCode:         acc.DraftCode,
			DeployedCode:      "",
			DeployedContracts: nil,
			State:             "",
		})
	}

	err = sqlDB.InsertAccounts(sqlAccounts)
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		telemetry.DebugLog("Error on migrate accounts for project ID " + projID.String() + " " + err.Error())
		numErrors++
	}
}

// convertTransactionTemplates Retrieves and converts models of datastore transaction templates
// to sql transaction templates
func convertTransactionTemplates(dstore *datastore.Datastore, projID uuid.UUID) (*[]*sqlModel.TransactionTemplate, error) {
	var ttpl []*model.TransactionTemplate
	err := dstore.GetTransactionTemplatesForProject(projID, &ttpl)
	if err != nil {
		telemetry.DebugLog(err.Error())
		telemetry.DebugLog("Error: could retrieve transaction templates for project ID " +
			projID.String() + " skipping project")
		numErrors++
		return nil, err
	}

	var sqlTtpl []*sqlModel.TransactionTemplate
	for _, ttp := range ttpl {
		sqlTtpl = append(sqlTtpl, &sqlModel.TransactionTemplate{
			ID:        ttp.ID,
			ProjectID: ttp.ProjectID,
			Title:     ttp.Title,
			Index:     ttp.Index,
			Script:    ttp.Script,
		})
	}
	return &sqlTtpl, nil
}

// migrateScriptTemplates converts models of datastore script templates to sql script templates
func convertScriptTemplates(dstore *datastore.Datastore, projID uuid.UUID) (*[]*sqlModel.ScriptTemplate, error) {
	var stpl []*model.ScriptTemplate
	err := dstore.GetScriptTemplatesForProject(projID, &stpl)
	if err != nil {
		telemetry.DebugLog(err.Error())
		return nil, err
	}

	var sqlStpl []*sqlModel.ScriptTemplate
	for _, ttp := range stpl {
		sqlStpl = append(sqlStpl, &sqlModel.ScriptTemplate{
			ID:        ttp.ID,
			ProjectID: ttp.ProjectID,
			Title:     ttp.Title,
			Index:     ttp.Index,
			Script:    ttp.Script,
		})
	}
	return &sqlStpl, nil
}

func convertProject(proj *model.InternalProject) *sqlModel.Project {
	return &sqlModel.Project{
		ID:                        proj.ID,
		UserID:                    proj.UserID,
		Secret:                    proj.Secret,
		PublicID:                  proj.PublicID,
		ParentID:                  proj.ParentID,
		Title:                     proj.Title,
		Description:               proj.Description,
		Readme:                    proj.Readme,
		Seed:                      proj.Seed,
		TransactionExecutionCount: proj.TransactionExecutionCount,
		Persist:                   proj.Persist,
		CreatedAt:                 proj.CreatedAt,
		UpdatedAt:                 proj.UpdatedAt,
		Version:                   proj.Version,
		Mutable:                   false,
	}
}

func migrateScriptExecutions(dstore *datastore.Datastore, sqlDB *storage.SQL, projID uuid.UUID) {
	var exes []*model.ScriptExecution
	err := dstore.GetScriptExecutionsForProject(projID, &exes)
	if err != nil {
		telemetry.DebugLog("Error: could not get script executions for project ID: " +
			projID.String() + " " + err.Error())
		return
	}

	for _, exe := range exes {
		// Convert Errors
		var sqlErrors []sqlModel.ProgramError
		for _, pError := range exe.Errors {
			sqlErrors = append(sqlErrors, sqlModel.ProgramError{
				Message: pError.Message,
				StartPosition: &sqlModel.ProgramPosition{
					Offset: pError.StartPosition.Offset,
					Line:   pError.StartPosition.Line,
					Column: pError.StartPosition.Column,
				},
				EndPosition: &sqlModel.ProgramPosition{
					Offset: pError.EndPosition.Offset,
					Line:   pError.EndPosition.Line,
					Column: pError.EndPosition.Column,
				},
			})
		}

		err := sqlDB.InsertScriptExecution(&sqlModel.ScriptExecution{
			ID:        exe.ID,
			ProjectID: exe.ProjectID,
			Index:     exe.Index,
			Script:    exe.Script,
			Arguments: exe.Arguments,
			Value:     exe.Value,
			Errors:    sqlErrors,
			Logs:      exe.Logs,
		})

		if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			telemetry.DebugLog("Error: could not insert script execution " + exe.ID.String() +
				"into project ID" + projID.String() + err.Error())
		}
	}
}

func migrateTransactionExecutions(dstore *datastore.Datastore, sqlDB *storage.SQL, projID uuid.UUID) {
	var exes []*model.TransactionExecution
	err := dstore.GetTransactionExecutionsForProject(projID, &exes)
	if err != nil {
		telemetry.DebugLog("Error: could not get transaction executions for project ID: " +
			projID.String() + " " + err.Error())
		return
	}

	for _, exe := range exes {
		// Convert signers
		var sqlSigners []sqlModel.Address
		for _, signer := range exe.Signers {
			sqlSigners = append(sqlSigners, sqlModel.Address(signer))
		}

		// Convert Errors
		var sqlErrors []sqlModel.ProgramError
		for _, pError := range exe.Errors {
			sqlErrors = append(sqlErrors, sqlModel.ProgramError{
				Message: pError.Message,
				StartPosition: &sqlModel.ProgramPosition{
					Offset: pError.StartPosition.Offset,
					Line:   pError.StartPosition.Line,
					Column: pError.StartPosition.Column,
				},
				EndPosition: &sqlModel.ProgramPosition{
					Offset: pError.EndPosition.Offset,
					Line:   pError.EndPosition.Line,
					Column: pError.EndPosition.Column,
				},
			})
		}

		// Convert Events
		var sqlEvents []sqlModel.Event
		for _, event := range exe.Events {
			sqlEvents = append(sqlEvents, sqlModel.Event(event))
		}

		err := sqlDB.InsertTransactionExecution(&sqlModel.TransactionExecution{
			ID:        exe.ID,
			ProjectID: exe.ProjectID,
			Index:     exe.Index,
			Script:    exe.Script,
			Arguments: exe.Arguments,
			Signers:   sqlSigners,
			Errors:    sqlErrors,
			Events:    sqlEvents,
			Logs:      exe.Logs,
		})

		if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			telemetry.DebugLog("Error: could not insert transaction execution" + exe.ID.String() +
				"into project ID " + projID.String() + " " + err.Error())
		}
	}
}

// migrateProject Migrates a project and corresponding transaction & script execution templates
func migrateProject(dstore *datastore.Datastore, sqlDB *storage.SQL, proj *model.InternalProject) {
	sqlProj := convertProject(proj)

	sqlTtpl, err := convertTransactionTemplates(dstore, proj.ID)
	if err != nil {
		telemetry.DebugLog("Error: could not migrate transaction templates for project ID " +
			proj.ID.String() + " skipping project")
		numErrors++
		return
	}

	// Convert script templates for project
	sqlStpl, err := convertScriptTemplates(dstore, proj.ID)
	if err != nil {
		telemetry.DebugLog("Error: could not migrate script templates for project ID " +
			proj.ID.String() + " skipping project")
		numErrors++
		return
	}

	// Store migrated project in SQL db
	err = sqlDB.CreateProject(sqlProj, *sqlTtpl, *sqlStpl)
	if err != nil {
		// TODO: Other cases?
		if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			telemetry.DebugLog("Error: could not store project ID " + proj.ID.String() +
				" in sql db. Skipping project. " + err.Error())
			numErrors++
		}
		return
	}
}

func migrateUser(dstore *datastore.Datastore, sqlDB *storage.SQL, proj *model.InternalProject) {
	user := model.User{}
	err := dstore.GetUser(proj.UserID, &user)
	if err != nil {
		if !strings.Contains(err.Error(), "no such entity") {
			telemetry.DebugLog("Error: could not get user for project ID " +
				proj.ID.String() + " " + err.Error())
		}
		return
	}

	err = sqlDB.InsertUser(&sqlModel.User{ID: user.ID})
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		telemetry.DebugLog("Error on insert user for project ID " +
			proj.ID.String() + " " + err.Error())
		numErrors++
	}
}
