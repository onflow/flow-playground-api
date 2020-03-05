package model

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-go/engine/execution/state"
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

type RegisterDelta struct {
	ProjectID         uuid.UUID
	Index             int
	Delta             state.Delta
	IsAccountCreation bool
}

func (r *RegisterDelta) NameKey() *datastore.Key {
	return datastore.NameKey("RegisterDelta", fmt.Sprintf("%s-%d", r.ProjectID.String(), r.Index), ProjectNameKey(r.ProjectID))
}

func (r *RegisterDelta) Load(ps []datastore.Property) error {
	tmp := struct {
		ProjectID         string
		Index             int
		Delta             []byte
		IsAccountCreation bool
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := r.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}
	r.Index = tmp.Index

	var delta state.Delta

	decoder := gob.NewDecoder(bytes.NewReader(tmp.Delta))
	err := decoder.Decode(&delta)
	if err != nil {
		return errors.Wrap(err, "failed to decode Delta")
	}

	r.Delta = delta

	r.IsAccountCreation = tmp.IsAccountCreation

	return nil
}

func (r *RegisterDelta) Save() ([]datastore.Property, error) {
	w := new(bytes.Buffer)

	encoder := gob.NewEncoder(w)
	err := encoder.Encode(&r.Delta)
	if err != nil {
		return nil, err
	}

	delta := w.Bytes()

	fmt.Println("IS ACCOUNT CREATION", r.IsAccountCreation)

	return []datastore.Property{
		{
			Name:  "ProjectID",
			Value: r.ProjectID.String(),
		},
		{
			Name:  "Index",
			Value: r.Index,
		},
		{
			Name:    "Delta",
			Value:   delta,
			NoIndex: true,
		},
		{
			Name:  "IsAccountCreation",
			Value: r.IsAccountCreation,
		},
	}, nil
}
