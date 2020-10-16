package model

import (
	"encoding/json"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
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
		return errors.Wrap(err, "failed to decode UUID")
	}
	if err := t.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
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

type TransactionExecution struct {
	ProjectChildID
	Index            int
	Script           string
	Arguments        []string
	SignerAccountIDs []uuid.UUID
	Error            *string
	Events           []Event
	Logs             []string
}

func (t *TransactionExecution) NameKey() *datastore.Key {
	return datastore.NameKey("TransactionExecution", t.ID.String(), ProjectNameKey(t.ProjectID))
}

func (t *TransactionExecution) Load(ps []datastore.Property) error {
	tmp := struct {
		ID               string
		ProjectID        string
		Index            int
		Script           string
		Arguments        []string
		SignerAccountIDs []string
		Error            *string
		Events           string
		Logs             []string
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := t.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}
	if err := t.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}

	for _, aID := range tmp.SignerAccountIDs {
		signer := uuid.UUID{}
		if err := signer.UnmarshalText([]byte(aID)); err != nil {
			return errors.Wrap(err, "failed to decode UUID")
		}
		t.SignerAccountIDs = append(t.SignerAccountIDs, signer)
	}

	if err := json.Unmarshal([]byte(tmp.Events), &t.Events); err != nil {
		return errors.Wrap(err, "failed to decode Events")
	}

	t.Index = tmp.Index
	t.Script = tmp.Script
	t.Arguments = tmp.Arguments
	t.Error = tmp.Error
	t.Logs = tmp.Logs
	return nil
}

func (t *TransactionExecution) Save() ([]datastore.Property, error) {
	signerAccountIDs := make([]interface{}, 0, len(t.Events))
	for _, aID := range t.SignerAccountIDs {
		signerAccountIDs = append(signerAccountIDs, aID.String())
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
			Name:  "Index",
			Value: t.Index,
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
			Name:  "SignerAccountIDs",
			Value: signerAccountIDs,
		},
		{
			Name:  "Error",
			Value: t.Error,
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
