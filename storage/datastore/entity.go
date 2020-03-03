package datastore

import "cloud.google.com/go/datastore"

type DatastoreEntity interface {
	NameKey() *datastore.Key
}
