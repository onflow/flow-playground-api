package model

import (
	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
)

type ScriptTemplate struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Index     int
	Script    string
}

func (s *ScriptTemplate) NameKey() *datastore.Key {
	return datastore.NameKey("ScriptTemplate", s.ID.String(), nil)
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

func (s *ScriptExecution) NameKey() *datastore.Key {
	return datastore.NameKey("ScriptExecution", s.ID.String(), nil)
}
