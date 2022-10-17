package model

import (
	"github.com/google/uuid"
)

// TODO: Regenerate models_gen.go

// TODO: Just have CadenceFiles rather than 3 separate ones in database (Type for FE to know what's what)
// TODO: Keep type in database only and not in File struct? Or nah.
// TODO: Update Projects with new models
// TODO: Go over flows of creating/ updating/ etc and see what needs to be updated:
// TODO: Script Executions
// TODO: Transaction Executions
// TODO: Contract Deployments

type FileType int

const (
	ContractFile FileType = iota
	TransactionFile
	ScriptFile
)

// File represents a template for a contract, transaction, or script
type File struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Title     string
	Type      FileType
	Index     int
	Script    string
}
