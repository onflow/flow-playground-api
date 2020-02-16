package model

import "github.com/google/uuid"

type Project struct {
	ID               uuid.UUID
	TransactionCount int
}
