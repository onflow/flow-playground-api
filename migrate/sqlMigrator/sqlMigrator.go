// Package sqlMigrator is used to migrate from Google datastore to the new SQL implementation
package sqlMigrator

import (
	"context"
	"fmt"
	"github.com/dapperlabs/flow-playground-api/migrate/sqlMigrator/model"
	"github.com/dapperlabs/flow-playground-api/migrate/sqlMigrator/storage/datastore"
	sqlModel "github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage/sql"
	"github.com/google/uuid"
	"time"
)

func main() {
	// TODO: Connect to test datastore on docker container
	dstore, err := connectToDatastore()
	if err != nil {
		fmt.Println("Error: could not connect to datastore", err)
		return
	}

	// TODO: Connect to test sql db
	sqlDB := sql.NewPostgreSQL()

	fmt.Println("Obtaining projects from datastore...")
	projects := *getAllProjects(dstore)
	if projects == nil {
		fmt.Println("Failed to obtain any projects from datastore")
		return
	}

	fmt.Println("Starting migration for", len(projects), "projects...")
	numErrors := 0
	for _, proj := range projects {
		sqlProj := migrateProject(proj)
		if !sqlProj.Persist {
			continue
		}

		// Migrate transaction templates for project
		sqlTtpl := migrateTransactionTemplates(dstore, proj.ID)
		if sqlTtpl == nil {
			fmt.Println("Error: could not migrate transaction templates for project ID", proj.ID.String(),
				"skipping project")
			numErrors++
			continue
		}

		// Migrate script templates for project
		sqlStpl := migrateScriptTemplates(dstore, proj.ID)
		if sqlStpl == nil {
			fmt.Println("Error: could not migrate script templates for project ID", proj.ID.String(),
				"skipping project")
			numErrors++
			continue
		}

		// Store migrated project in SQL db
		err := sqlDB.CreateProject(sqlProj, *sqlTtpl, *sqlStpl)
		if err != nil {
			fmt.Println("Error: could not store project ID", proj.ID.String(),
				"in sql db. Skipping project.", err)
			numErrors++
			continue
		}

		// Migrate accounts for project
		sqlAccounts := migrateAccounts(dstore, proj.ID)
		if sqlAccounts == nil {
			fmt.Println("Could not retrieve any accounts for project ID", proj.ID)
			continue
		}
		err = sqlDB.InsertAccounts(*sqlAccounts)
		if err != nil {
			fmt.Println("Error on migrate accounts for project ID", proj.ID.String(), err)
			numErrors++
		}
	}
	fmt.Println("Migration finished with", numErrors, "errors")
}

func connectToDatastore() (*datastore.Datastore, error) {
	store, err := datastore.NewDatastore(context.Background(), &datastore.Config{
		DatastoreProjectID: "test-project", // "dl-flow",
		DatastoreTimeout:   time.Second * 5,
	})
	if err != nil {
		return nil, err
	}
	return store, nil
}

// getAllProjects returns a list of all projects in the datastore
func getAllProjects(store *datastore.Datastore) *[]*model.InternalProject {
	var projects []*model.InternalProject
	err := store.GetAllProjects(&projects)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &projects
}

// migrateAccounts converts models of datastore accounts to sql accounts
func migrateAccounts(dstore *datastore.Datastore, projID uuid.UUID) *[]*sqlModel.Account {
	var accounts []*model.InternalAccount
	err := dstore.GetAccountsForProject(projID, &accounts)
	if err != nil {
		return nil
	}

	var sqlAccounts []*sqlModel.Account
	for _, acc := range accounts {
		tmp := sqlModel.Account{}
		tmp.ID = acc.ID
		tmp.ProjectID = acc.ProjectID
		tmp.Address = sqlModel.Address(acc.Address)
		tmp.Index = acc.Index
		tmp.DraftCode = acc.DraftCode
		sqlAccounts = append(sqlAccounts, &tmp)
	}
	return &sqlAccounts
}

// migrateTransactionTemplates converts models of datastore transaction templates to sql transaction templates
func migrateTransactionTemplates(dstore *datastore.Datastore, projID uuid.UUID) *[]*sqlModel.TransactionTemplate {
	var ttpl []*model.TransactionTemplate
	err := dstore.GetTransactionTemplatesForProject(projID, &ttpl)
	if err != nil {
		fmt.Println(err)
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
func migrateScriptTemplates(dstore *datastore.Datastore, projID uuid.UUID) *[]*sqlModel.ScriptTemplate {
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

// migrateProject converts models of datastore project to sql project
func migrateProject(proj *model.InternalProject) *sqlModel.Project {
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
	sqlProj.TransactionCount = proj.TransactionCount
	sqlProj.TransactionExecutionCount = proj.TransactionExecutionCount
	sqlProj.TransactionExecutionCount = proj.TransactionExecutionCount
	sqlProj.TransactionTemplateCount = proj.TransactionTemplateCount
	sqlProj.ScriptTemplateCount = proj.ScriptTemplateCount
	sqlProj.Persist = proj.Persist
	sqlProj.CreatedAt = proj.CreatedAt
	sqlProj.UpdatedAt = proj.UpdatedAt
	sqlProj.Version = proj.Version
	return sqlProj
}
