package cmd

// Database migrator is used to migrate from Google datastore to the new SQL postgres database

import (
	"context"
	"github.com/dapperlabs/flow-playground-api/migrate/database/storage/datastore"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/telemetry"
	"strconv"
	"time"
)

// numErrors counts the errors that occur during migration
var numErrors = 0

func main() {
	dstore := connectToDatastore()
	telemetry.DebugLog("Connected to datastore")

	sqlDB := connectToSQL()
	telemetry.DebugLog("Connected to SQL database")

	telemetry.DebugLog("Starting migration...")
	numProjects := 0
	for p := datastore.CreateIterator(dstore, 100); p.HasNext(); p.GetNext() {
		numProjects += len(p.Projects)
		telemetry.DebugLog("Migrating projects " + strconv.Itoa(p.GetIndex()) +
			" - " + strconv.Itoa(numProjects))
		for _, proj := range p.Projects {
			if !proj.Persist {
				continue
			}
			migrateProject(dstore, sqlDB, proj) // Includes transaction & script execution templates
			migrateAccounts(dstore, sqlDB, proj.ID)
			migrateUser(dstore, sqlDB, proj)
			migrateScriptExecutions(dstore, sqlDB, proj.ID)
			migrateTransactionExecutions(dstore, sqlDB, proj.ID)
		}
	}
	telemetry.DebugLog("Migration of " + strconv.Itoa(numProjects) +
		" projects finished with " + strconv.Itoa(numErrors) + " errors")
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
	return store
}

func connectToSQL() *storage.SQL {
	// TODO: connect to the real postgres database
	/*
		var datastoreConf storage.DatabaseConfig
		if err := envconfig.Process("FLOW_DB", &datastoreConf); err != nil {
			log.Fatal(err)
		}
	*/

	sqlDB := storage.NewPostgreSQL(&storage.DatabaseConfig{
		User:     "newuser", // test db with newuser / password
		Password: "password",
		Name:     "postgres",
		Port:     5432,
	})
	return sqlDB
}
