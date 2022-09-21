// Package sqlMigrator is used to migrate from Google datastore to the new SQL implementation
package sqlMigrator

import (
	"context"
	"fmt"
	"github.com/dapperlabs/flow-playground-api/migrate/sqlMigrator/storage/datastore"
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

	fmt.Println("Obtaining all projects from datastore...")

	// Get all projects for migration
}
