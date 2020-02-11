package model

import "github.com/google/uuid"

type TransactionTemplate struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Index     int
	Script    string
}

type TransactionExecution struct {
	ID               uuid.UUID
	Index            int
	Script           string
	PayerAccountID   uuid.UUID
	SignerAccountIDs []uuid.UUID
}
