package model

import (
	"cloud.google.com/go/datastore"
	"github.com/pkg/errors"
)

type ScriptTemplate struct {
	ProjectChildID
	Title  string
	Index  int
	Script string
}

func (s *ScriptTemplate) NameKey() *datastore.Key {
	return datastore.NameKey("ScriptTemplate", s.ID.String(), ProjectNameKey(s.ProjectID))
}

func (s *ScriptTemplate) Load(ps []datastore.Property) error {
	tmp := struct {
		ID        string
		ProjectID string
		Title     string
		Index     int
		Script    string
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := s.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}
	if err := s.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}
	s.Title = tmp.Title
	s.Index = tmp.Index
	s.Script = tmp.Script
	return nil
}

func (s *ScriptTemplate) Save() ([]datastore.Property, error) {
	return []datastore.Property{
		{
			Name:  "ID",
			Value: s.ID.String(),
		},
		{
			Name:  "ProjectID",
			Value: s.ProjectID.String(),
		},
		{
			Name:  "Title",
			Value: s.Title,
		},
		{
			Name:  "Index",
			Value: s.Index,
		},
		{
			Name:    "Script",
			Value:   s.Script,
			NoIndex: true,
		},
	}, nil
}

type ScriptExecution struct {
	ProjectChildID
	Index  int
	Script string
	Value  string
	Error  *string
	Logs   []string
}

func (s *ScriptExecution) NameKey() *datastore.Key {
	return datastore.NameKey("ScriptExecution", s.ID.String(), ProjectNameKey(s.ProjectID))
}

func (s *ScriptExecution) Load(ps []datastore.Property) error {
	tmp := struct {
		ID        string
		ProjectID string
		Index     int
		Script    string
		Value     string
		Error     *string
		Logs      []string
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := s.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}
	if err := s.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}
	s.Index = tmp.Index
	s.Script = tmp.Script
	s.Value = tmp.Value
	s.Error = tmp.Error
	s.Logs = tmp.Logs
	return nil
}

func (s *ScriptExecution) Save() ([]datastore.Property, error) {
	return []datastore.Property{
		{
			Name:  "ID",
			Value: s.ID.String(),
		},
		{
			Name:  "ProjectID",
			Value: s.ProjectID.String(),
		},
		{
			Name:  "Index",
			Value: s.Index,
		},
		{
			Name:    "Script",
			Value:   s.Script,
			NoIndex: true,
		},
		{
			Name:  "Value",
			Value: s.Value,
		},
		{
			Name:  "Error",
			Value: s.Error,
		},
		{
			Name:  "Logs",
			Value: s.Logs,
		},
	}, nil
}
