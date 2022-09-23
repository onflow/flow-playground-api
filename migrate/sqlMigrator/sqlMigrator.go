// Package sqlMigrator is used to migrate from Google datastore to the new SQL implementation
package sqlMigrator

import (
	"context"
	"fmt"
	"github.com/dapperlabs/flow-playground-api/migrate/sqlMigrator/model"
	"github.com/dapperlabs/flow-playground-api/migrate/sqlMigrator/storage/datastore"
	sqlModel "github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"time"
)

// numErrors counts the errors that occur during migration
var numErrors = 0

func main() {
	dstore := connectToDatastore()
	sqlDB := connectToSQL()

	projects := *getAllProjects(dstore)

	fmt.Println("Starting migration for", len(projects), "projects...")
	for _, proj := range projects {
		if !proj.Persist {
			continue
		}
		migrateProject(dstore, sqlDB, proj) // Includes transaction & script execution templates
		migrateAccounts(dstore, sqlDB, proj.ID)
		migrateUser(dstore, sqlDB, proj)
		migrateScriptExecutions(dstore, sqlDB, proj.ID)
		migrateTransactionExecutions(dstore, sqlDB, proj.ID)
	}
	fmt.Println("Migration finished with", numErrors, "errors")
}

func connectToDatastore() *datastore.Datastore {
	// TODO: connect to actual datastore
	store, err := datastore.NewDatastore(context.Background(), &datastore.Config{
		DatastoreProjectID: "test-project", // "dl-flow",
		DatastoreTimeout:   time.Second * 5,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to datastore")
	return store
}

func connectToSQL() *storage.SQL {
	// TODO: connect to the real postgres database
	sqlDB := storage.NewPostgreSQL(&storage.DatabaseConfig{
		User:     "newuser", // test db with newuser / password
		Password: "password",
		Name:     "postgres",
		Port:     5432,
	})
	fmt.Println("Connected to SQL database")
	return sqlDB
}

// getAllProjects returns a list of all projects in the datastore
func getAllProjects(store *datastore.Datastore) *[]*model.InternalProject {
	fmt.Println("Obtaining projects from datastore...")
	var projects []*model.InternalProject
	err := store.GetAllProjects(&projects)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return &projects
}

// migrateAccounts migrates models of datastore accounts to sql accounts
func migrateAccounts(dstore *datastore.Datastore, sqlDB *storage.SQL, projID uuid.UUID) {
	var accounts []*model.InternalAccount
	err := dstore.GetAccountsForProject(projID, &accounts)
	if err != nil {
		numErrors++
		return
	}

	var sqlAccounts []*sqlModel.Account
	for _, acc := range accounts {
		tmp := sqlModel.Account{}
		tmp.ID = acc.ProjectChildID.ID
		tmp.ProjectID = acc.ProjectChildID.ProjectID
		tmp.Address = sqlModel.Address(acc.Address)
		tmp.Index = acc.Index
		tmp.DraftCode = acc.DraftCode
		sqlAccounts = append(sqlAccounts, &tmp)
	}

	err = sqlDB.InsertAccounts(sqlAccounts)
	if err != nil {
		fmt.Println("Error on migrate accounts for project ID", projID.String(), err)
		numErrors++
	}
}

// convertTransactionTemplates Retrieves and converts models of datastore transaction templates
// to sql transaction templates
func convertTransactionTemplates(dstore *datastore.Datastore, projID uuid.UUID) *[]*sqlModel.TransactionTemplate {
	var ttpl []*model.TransactionTemplate
	err := dstore.GetTransactionTemplatesForProject(projID, &ttpl)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Error: could retrieve transaction templates for project ID", projID,
			"skipping project")
		numErrors++
		return nil
	}

	var sqlTtpl []*sqlModel.TransactionTemplate
	for _, ttp := range ttpl {
		tmp := sqlModel.TransactionTemplate{}
		tmp.ID = ttp.ID
		tmp.ProjectID = ttp.ID
		tmp.Title = ttp.Title
		tmp.Index = ttp.Index
		tmp.Script = ttp.Script
		sqlTtpl = append(sqlTtpl, &tmp)
	}
	return &sqlTtpl
}

// migrateScriptTemplates converts models of datastore script templates to sql script templates
func convertScriptTemplates(dstore *datastore.Datastore, projID uuid.UUID) *[]*sqlModel.ScriptTemplate {
	var stpl []*model.ScriptTemplate
	err := dstore.GetScriptTemplatesForProject(projID, &stpl)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var sqlStpl []*sqlModel.ScriptTemplate
	for _, ttp := range stpl {
		tmp := sqlModel.ScriptTemplate{}
		tmp.ID = ttp.ID
		tmp.ProjectID = ttp.ID
		tmp.Title = ttp.Title
		tmp.Index = ttp.Index
		tmp.Script = ttp.Script
		sqlStpl = append(sqlStpl, &tmp)
	}
	return &sqlStpl
}

func convertProject(proj *model.InternalProject) *sqlModel.Project {
	sqlProj := &sqlModel.Project{}
	sqlProj.ID = proj.ID
	sqlProj.UserID = proj.UserID
	sqlProj.Secret = proj.Secret
	sqlProj.PublicID = proj.PublicID
	sqlProj.ParentID = proj.ParentID
	sqlProj.Title = proj.Title
	sqlProj.Description = proj.Description
	sqlProj.Readme = proj.Readme
	sqlProj.Seed = proj.Seed
	sqlProj.TransactionExecutionCount = proj.TransactionExecutionCount
	sqlProj.TransactionExecutionCount = proj.TransactionExecutionCount
	sqlProj.Persist = proj.Persist
	sqlProj.CreatedAt = proj.CreatedAt
	sqlProj.UpdatedAt = proj.UpdatedAt
	sqlProj.Version = proj.Version
	sqlProj.Mutable = false
	return sqlProj
}

func migrateScriptExecutions(dstore *datastore.Datastore, sqlDB *storage.SQL, projID uuid.UUID) {
	var exes []*model.ScriptExecution
	err := dstore.GetScriptExecutionsForProject(projID, &exes)
	if err != nil {
		fmt.Println("Error: could not get script executions for project ID: ", projID, err)
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

		tmp := sqlModel.ScriptExecution{
			ID:        exe.ID,
			ProjectID: exe.ProjectID,
			Index:     exe.Index,
			Script:    exe.Script,
			Arguments: exe.Arguments,
			Value:     exe.Value,
			Errors:    sqlErrors,
			Logs:      exe.Logs,
		}

		err := sqlDB.InsertScriptExecution(&tmp)
		if err != nil {
			fmt.Println("Error: could not insert script execution", tmp.ID, "into project ID", projID, err)
		}
	}
}

func migrateTransactionExecutions(dstore *datastore.Datastore, sqlDB *storage.SQL, projID uuid.UUID) {
	var exes []*model.TransactionExecution
	err := dstore.GetTransactionExecutionsForProject(projID, &exes)
	if err != nil {
		fmt.Println("Error: could not get transaction executions for project ID: ", projID, err)
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

		tmp := sqlModel.TransactionExecution{
			ID:        exe.ID,
			ProjectID: exe.ProjectID,
			Index:     exe.Index,
			Script:    exe.Script,
			Arguments: exe.Arguments,
			Signers:   sqlSigners,
			Errors:    sqlErrors,
			Events:    sqlEvents,
			Logs:      exe.Logs,
		}

		err := sqlDB.InsertTransactionExecution(&tmp)
		if err != nil {
			fmt.Println("Error: could not insert transaction execution", tmp.ID, "into project ID", projID, err)
		}
	}
}

// migrateProject Migrates a project and corresponding transaction & script execution templates
func migrateProject(dstore *datastore.Datastore, sqlDB *storage.SQL, proj *model.InternalProject) {
	sqlProj := convertProject(proj)

	sqlTtpl := convertTransactionTemplates(dstore, proj.ID)
	if sqlTtpl == nil {
		fmt.Println("Error: could not migrate transaction templates for project ID", proj.ID.String(),
			"skipping project")
		numErrors++
		return
	}

	// Convert script templates for project
	sqlStpl := convertScriptTemplates(dstore, proj.ID)
	if sqlStpl == nil {
		fmt.Println("Error: could not migrate script templates for project ID", proj.ID.String(),
			"skipping project")
		numErrors++
		return
	}

	// Store migrated project in SQL db
	err := sqlDB.CreateProject(sqlProj, *sqlTtpl, *sqlStpl)
	if err != nil {
		fmt.Println("Error: could not store project ID", proj.ID.String(),
			"in sql db. Skipping project.", err)
		numErrors++
	}
}

func migrateUser(dstore *datastore.Datastore, sqlDB *storage.SQL, proj *model.InternalProject) {
	user := model.User{}
	err := dstore.GetUser(proj.UserID, &user)
	if err != nil {
		fmt.Println("Error: could not get user for project ID", proj.ID, err)
		numErrors++
		return
	}

	err = sqlDB.InsertUser(&sqlModel.User{ID: user.ID})
	if err != nil {
		fmt.Println("Error on insert user for project ID", proj.ID.String(), err)
		numErrors++
	}
}
