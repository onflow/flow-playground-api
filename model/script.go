package model

import "github.com/google/uuid"

type ScriptTemplate struct {
	ID     uuid.UUID
	Index  int
	Script string
}

type ScriptExecution struct {
	ID     uuid.UUID
	Index  int
	Script string
}
