package datastore

import (
	"cloud.google.com/go/datastore"
	"github.com/dapperlabs/flow-playground-api/migrate/database/model"
	"github.com/dapperlabs/flow-playground-api/telemetry"
)

// SELECT id FROM projects ORDER BY id LIMIT 10 OFFSET n

type DatastoreIterator struct {
	index        int
	limit        int
	dstore       *Datastore
	Projects     []*model.InternalProject
	nextProjects []*model.InternalProject
}

func CreateIterator(dstore *Datastore, limit int) *DatastoreIterator {
	dIter := DatastoreIterator{
		index:        0,
		limit:        limit,
		dstore:       dstore,
		Projects:     nil,
		nextProjects: []*model.InternalProject{},
	}
	dIter.GetNext()
	return &dIter
}

func (d *DatastoreIterator) HasNext() bool {
	if len(d.nextProjects) > 0 {
		return true
	}
	return false
}

func (d *DatastoreIterator) GetNext() {
	d.Projects = d.nextProjects
	d.nextProjects = []*model.InternalProject{}
	query := datastore.NewQuery("Project").Offset(d.index).Limit(d.limit)
	err := d.dstore.getAll(query, &d.nextProjects)
	if err != nil {
		telemetry.DebugLog("Error: failed to get projects. " + err.Error())
		panic(err)
	}
	d.index += d.limit
}

func (d *DatastoreIterator) GetIndex() int {
	return d.index
}
