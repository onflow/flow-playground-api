package model

import (
	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
)

type Account struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	Index        int
	Address      Address
	DraftCode    string
	DeployedCode string
}

func (a *Account) NameKey() *datastore.Key {
	return datastore.NameKey("Account", a.ID.String(), nil)
}
