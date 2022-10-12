package main

// Database migrator is used to migrate from Google datastore to the new SQL postgres database

import (
	"context"
	"github.com/dapperlabs/flow-playground-api/migrate/database/storage/datastore"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/dapperlabs/flow-playground-api/telemetry"
	"github.com/kelseyhightower/envconfig"
	"log"
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

	telemetry.DebugLog("Clearing SQL database...")
	sqlDB.DeleteAllData()

	telemetry.DebugLog("Starting migration...")
	numProjects := 0
	var err error = nil
	exitWithError := false
	for p := datastore.CreateIterator(dstore, 100); p.HasNext(); err = p.GetNext() {
		if err != nil {
			telemetry.DebugLog("Error getting data from datastore iterator: " + err.Error())
			exitWithError = true
			break
		}

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
	if !exitWithError {
		telemetry.DebugLog("Migration of " + strconv.Itoa(numProjects) +
			" projects finished with " + strconv.Itoa(numErrors) + " potential errors")
	} else {
		telemetry.DebugLog("Migration failed after " + strconv.Itoa(numProjects) +
			"projects with the following error: " + err.Error())
	}

	for {
	} // Busy wait to prevent migration from running again
}

func connectToDatastore() *datastore.Datastore {
	store, err := datastore.NewDatastore(context.Background(), &datastore.Config{
		DatastoreProjectID: "flow-developer-playground",
		DatastoreTimeout:   time.Second * 1000, // TODO: Large timeout to avoid context deadline exceeded?
	})
	if err != nil {
		panic(err)
	}
	return store
}

func connectToSQL() *storage.SQL {
	var datastoreConf storage.DatabaseConfig
	if err := envconfig.Process("FLOW_DB", &datastoreConf); err != nil {
		log.Fatal(err)
	}

	return storage.NewPostgreSQL(&datastoreConf)
}
