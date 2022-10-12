package datastore

// Iterator for obtaining projects in Google datastore

import (
	"cloud.google.com/go/datastore"
	"github.com/dapperlabs/flow-playground-api/migrate/database/model"
	"github.com/dapperlabs/flow-playground-api/telemetry"
)

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
	_ = dIter.GetNext()
	_ = dIter.GetNext()
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
	query := datastore.NewQuery("Project").Offset(d.index).Limit(d.limit)
	err := d.dstore.getAll(query, &d.nextProjects)
	if err != nil {
		telemetry.DebugLog("Error: failed to get projects. " + err.Error())
		return err
		//panic(err)
	}
	d.index += d.limit
	return nil
}

func (d *DatastoreIterator) GetIndex() int {
	return d.index - 2*d.limit
}
