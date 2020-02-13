package model

import (
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
