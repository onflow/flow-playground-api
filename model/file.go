package model

import (
	"github.com/google/uuid"
)

type FileType int

const (
	ContractFile FileType = iota
	TransactionFile
	ScriptFile
)

// File represents a template for a contract, transaction, or script
type File struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"projectId"`
	Title     string    `json:"title"`
	Type      FileType  `json:"type"`
	Index     int       `json:"index"`
	Script    string    `json:"script"`
}
