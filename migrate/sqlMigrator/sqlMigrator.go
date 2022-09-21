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
	// Connect to datastore
	store, err := datastore.NewDatastore(context.Background(), &datastore.Config{
		DatastoreProjectID: "dl-flow",
		DatastoreTimeout:   time.Second * 5,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// TODO: Connect to SQL db
	sqldb := sql.NewPostgreSQL()

	fmt.Println("Obtaining all projects from datastore...")

	// TODO: Get all projects for migration
	id, _ := uuid.NewUUID()

	var test = []int{1, 2, 3}

	// Get each project from datastore and migrate to sql
	for i, _ := range test {
		proj := &model.InternalProject{}
		err := store.GetProject(id, proj)
		if err != nil {
			// TODO: handle error
			fmt.Println("Could not get project", err)
		}

		sqlProj := migrateProject(proj)

		// Drop projects that aren't supposed to be persisted
		if !sqlProj.Persist {
			continue
		}

		// Get script and transaction templates
		var ttpl []*model.TransactionTemplate
		err = store.GetTransactionTemplatesForProject(id, &ttpl)
		if err != nil {
			// TODO: handle error
		}
		// TODO: convert ttpl []*model.TransactionTemplate to sqlttpl []*sqlmodel.TransactionTemplate
		//sqlttpl := migrateTransactionTemplate(ttpl)

		// TODO: stpl migration
		var stpl []*model.ScriptTemplate
		store.GetScriptTemplatesForProject(id, &stpl)

		// Store sqlProj in SQL db
		err := sqldb.CreateProject(sqlProj, sqlttpl, stpl)
	}

}

func migrateTransactionTemplate(ttpl *[]*model.TransactionTemplate) *[]*sqlModel.TransactionTemplate {
	var sqlttpl []*sqlModel.TransactionTemplate

	for i, ttp := range *ttpl {
		sqlttp := sqlModel.TransactionTemplate{}
		sqlttp.ID = ttp.ID
		sqlttp.ProjectID = ttp.ID
		sqlttp.Title = ttp.Title
		sqlttp.Index = ttp.Index
		sqlttp.Script = ttp.Script
		sqlttpl = append(sqlttpl, &sqlttp)
	}

	return &sqlttpl
}

func migrateAccount(acc *model.Account) *sqlModel.Account {

}

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
