package datastore

// Iterator for obtaining projects in Google datastore

import (
	"cloud.google.com/go/datastore"
	"github.com/dapperlabs/flow-playground-api/migrate/database/model"
	"github.com/dapperlabs/flow-playground-api/telemetry"
	"time"
)

// maxTransactionRate limit on maximum rate of transactions reading from or writing to an entity group
const maxTransactionRate = 1 * time.Second

// prevRun is the timestamp of the last query
var prevRun time.Time

// DatastoreIterator iterates over all projects in the datastore
type DatastoreIterator struct {
	index        int
	limit        int
	dstore       *Datastore
	Projects     []*model.InternalProject
	nextProjects []*model.InternalProject
}

// CreateIterator returns an iterator containing the first group of Projects
func CreateIterator(dstore *Datastore, limit int) *DatastoreIterator {
	dIter := DatastoreIterator{
		index:        0,
		limit:        limit,
		dstore:       dstore,
		Projects:     nil,
		nextProjects: []*model.InternalProject{},
	}
	// Initialize first entries
	prevRun = time.Now()
	err := dIter.GetNext()
	if err != nil {
		panic(err)
	}
	_ = dIter.GetNext()
	if err != nil {
		panic(err)
	}
	return &dIter
}

func (d *DatastoreIterator) HasNext() bool {
	if len(d.Projects) > 0 {
		return true
	}
	return false
}

func (d *DatastoreIterator) GetNext() error {
	d.Projects = d.nextProjects
	d.nextProjects = []*model.InternalProject{}

	// Query for persisted projects only
	// Will ignoring non persisted entries skew the offset??
	// This Would mean we try the same projects again, which leads to duplicate key errors.
	query := datastore.NewQuery("Project").Filter("Persist =", true).Offset(d.index).Limit(d.limit)

	// Wait to not exceed the max transaction rate of datastore
	elapsedTime := time.Since(prevRun)
	if elapsedTime < maxTransactionRate {
		time.Sleep(maxTransactionRate - elapsedTime)
	}

	err := d.dstore.getAll(query, &d.nextProjects)
	if err != nil {
		telemetry.DebugLog("Error: failed to get projects. " + err.Error())
		return err
		//panic(err)
	}
	prevRun = time.Now()

	d.index += d.limit
	return nil
}

func (d *DatastoreIterator) GetIndex() int {
	return d.index - 2*d.limit
}
