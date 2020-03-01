package model

import (
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/dapperlabs/flow-go/engine/execution/execution/state"
	"github.com/google/uuid"
)

type TransactionTemplate struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Index     int
	Script    string
}

func (t *TransactionTemplate) NameKey() *datastore.Key {
	return datastore.NameKey("TransactionTemplate", t.ID.String(), nil)
}

type TransactionExecution struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	Index            int
	Script           string
	SignerAccountIDs []uuid.UUID
	Error            *string
	Events           []Event
	Logs             []string
}

func (t *TransactionExecution) NameKey() *datastore.Key {
	return datastore.NameKey("TransactionExecution", t.ID.String(), nil)
}

type RegisterDelta struct {
	ProjectID uuid.UUID
	Index     int
	Delta     state.Delta
}

func (r *RegisterDelta) NameKey() *datastore.Key {
	return datastore.NameKey("RegisterDelta", fmt.Sprintf("%s-%d", r.ProjectID.String(), r.Index), nil)
}
