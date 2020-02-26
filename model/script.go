package model

import "github.com/google/uuid"

type ScriptTemplate struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Index     int
	Script    string
}

type ScriptExecution struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Index     int
	Script    string
	Value     XDRValue
	Error     *string
	Logs      []string
}
