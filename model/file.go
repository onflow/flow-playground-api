package model

import "github.com/google/uuid"

// TODO: Regenerate models_gen.go

// File represents a template for a contract, transaction, or script
type File struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Title     string
	Index     int
	Script    string
}
