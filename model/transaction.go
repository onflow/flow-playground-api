package model

import (
	"github.com/dapperlabs/flow-go/engine/execution/execution/state"
	"github.com/google/uuid"
)

type TransactionTemplate struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Index     int
	Script    string
}

type TransactionExecution struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	Index            int
	Script           string
	PayerAccountID   uuid.UUID
	SignerAccountIDs []uuid.UUID
	Error            string
	Events           []string
}

type RegisterDelta struct {
	ProjectID uuid.UUID
	Index     int
	Delta     state.Delta
}
