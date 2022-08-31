/*
 * Flow Playground
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

import (
	"cloud.google.com/go/datastore"
	"github.com/pkg/errors"
)

// todo ScriptTemplate and TransactionTemplate could be refactored into generic Templates

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
		return errors.Wrap(err, "failed to decode script template UUID")
	}
	if err := s.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode project UUID")
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
	Index     int
	Script    string
	Arguments []string
	Value     string
	Errors    []ProgramError
	Logs      []string
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
		Arguments []string
		Value     string
		Logs      []string
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := s.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode script execution UUID")
	}
	if err := s.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode project UUID")
	}
	s.Index = tmp.Index
	s.Script = tmp.Script
	s.Arguments = tmp.Arguments
	s.Value = tmp.Value
	s.Logs = tmp.Logs
	return nil
}

func (s *ScriptExecution) Save() ([]datastore.Property, error) {

	logs := make([]interface{}, 0, len(s.Logs))
	for _, log := range s.Logs {
		logs = append(logs, log)
	}

	arguments := make([]interface{}, 0, len(s.Arguments))
	for _, argument := range s.Arguments {
		arguments = append(arguments, argument)
	}

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
			Name:    "Arguments",
			Value:   arguments,
			NoIndex: true,
		},
		{
			Name:  "Value",
			Value: s.Value,
		},
		{
			Name:  "Logs",
			Value: logs,
		},
	}, nil
}
