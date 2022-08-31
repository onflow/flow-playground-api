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
	"encoding/json"

	"github.com/google/uuid"
	"github.com/onflow/flow-emulator/types"

	"cloud.google.com/go/datastore"
	"github.com/pkg/errors"
)

type TransactionTemplate struct {
	ProjectChildID
	Title  string
	Index  int
	Script string
}

func (t *TransactionTemplate) NameKey() *datastore.Key {
	return datastore.NameKey("TransactionTemplate", t.ID.String(), ProjectNameKey(t.ProjectID))
}

func (t *TransactionTemplate) Load(ps []datastore.Property) error {
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

	if err := t.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode transaction template UUID")
	}
	if err := t.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode project UUID")
	}
	t.Title = tmp.Title
	t.Index = tmp.Index
	t.Script = tmp.Script
	return nil
}

func (t *TransactionTemplate) Save() ([]datastore.Property, error) {
	return []datastore.Property{
		{
			Name:  "ID",
			Value: t.ID.String(),
		},
		{
			Name:  "ProjectID",
			Value: t.ProjectID.String(),
		},
		{
			Name:  "Title",
			Value: t.Title,
		},
		{
			Name:  "Index",
			Value: t.Index,
		},
		{
			Name:    "Script",
			Value:   t.Script,
			NoIndex: true,
		},
	}, nil
}

func TransactionExecutionFromFlow(
	result *types.TransactionResult,
	projectID uuid.UUID,
	script string,
	args []string,
	signers []Address,
) (*TransactionExecution, error) {
	id := ProjectChildID{
		ID:        uuid.New(),
		ProjectID: projectID,
	}

	exe := &TransactionExecution{
		ProjectChildID: id,
		Script:         script,
		Arguments:      args,
		Signers:        signers,
	}

	events, err := EventsFromFlow(result.Events)
	if err != nil {
		return nil, err
	}
	exe.Events = events

	exe.Errors = ProgramErrorFromFlow(result.Error)

	return exe, nil
}

type TransactionExecution struct {
	ProjectChildID
	Index     int
	Script    string
	Arguments []string
	Signers   []Address
	Errors    []ProgramError
	Events    []Event
	Logs      []string
}

func (t *TransactionExecution) NameKey() *datastore.Key {
	return datastore.NameKey("TransactionExecution", t.ID.String(), ProjectNameKey(t.ProjectID))
}

func (t *TransactionExecution) Load(ps []datastore.Property) error {
	tmp := struct {
		ID        string
		ProjectID string
		Index     int
		Script    string
		Arguments []string
		Signers   [][]byte
		Events    string
		Logs      []string
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := t.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode transaction execution UUID")
	}
	if err := t.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode project UUID")
	}

	for _, sig := range tmp.Signers {
		var signer Address
		copy(signer[:], sig[:])
		t.Signers = append(t.Signers, signer)
	}

	if err := json.Unmarshal([]byte(tmp.Events), &t.Events); err != nil {
		return errors.Wrap(err, "failed to decode Events")
	}

	t.Index = tmp.Index
	t.Script = tmp.Script
	t.Arguments = tmp.Arguments
	t.Logs = tmp.Logs
	return nil
}

func (t *TransactionExecution) Save() ([]datastore.Property, error) {
	signers := make([]interface{}, 0, len(t.Signers))
	for _, sig := range t.Signers {
		signers = append(signers, sig.ToFlowAddress().Bytes())
	}

	events, err := json.Marshal(t.Events)
	if err != nil {
		return nil, err
	}

	logs := make([]interface{}, 0, len(t.Logs))
	for _, log := range t.Logs {
		logs = append(logs, log)
	}

	arguments := make([]interface{}, 0, len(t.Arguments))
	for _, argument := range t.Arguments {
		arguments = append(arguments, argument)
	}

	return []datastore.Property{
		{
			Name:  "ID",
			Value: t.ID.String(),
		},
		{
			Name:  "ProjectID",
			Value: t.ProjectID.String(),
		},
		{
			Name:    "Script",
			Value:   t.Script,
			NoIndex: true,
		},
		{
			Name:    "Arguments",
			Value:   arguments,
			NoIndex: true,
		},
		{
			Name:  "Signers",
			Value: signers,
		},
		{
			Name:    "Events",
			Value:   string(events),
			NoIndex: true,
		},
		{
			Name:  "Logs",
			Value: logs,
		},
	}, nil
}
