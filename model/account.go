package model

import (
	"encoding/json"
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type InternalAccount struct {
	ProjectChildID
	Index             int
	Address           Address
	DraftCode         string
	DeployedCode      string
	DeployedContracts []string
	marshalledState   string
	unmarshalledState AccountState
}

func (a *InternalAccount) State() (AccountState, error) {
	if a.unmarshalledState != nil {
		return a.unmarshalledState, nil
	}

	state := []byte(a.marshalledState)

	err := json.Unmarshal(state, &a.unmarshalledState)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal account state: %w", err)
	}

	return a.unmarshalledState, nil
}

func (a *InternalAccount) SetState(state AccountState) {
	a.marshalledState = ""
	a.unmarshalledState = state
}

func (a *InternalAccount) marshalAccountState() (string, error) {
	if a.marshalledState != "" {
		return a.marshalledState, nil
	}

	stateBytes, err := json.Marshal(a.unmarshalledState)
	if err != nil {
		return "", fmt.Errorf("failed to marshal account state: %w", err)
	}

	return string(stateBytes), nil
}

type UpdateAccount struct {
	ID                uuid.UUID `json:"id"`
	ProjectID         uuid.UUID `json:"projectId"`
	DraftCode         *string   `json:"draftCode"`
	DeployedCode      *string   `json:"deployedCode"`
	DeployedContracts *[]string
}

func (a *InternalAccount) NameKey() *datastore.Key {
	return datastore.NameKey("Account", a.ID.String(), ProjectNameKey(a.ProjectID))
}

func (a *InternalAccount) Load(ps []datastore.Property) error {
	tmp := struct {
		ID                string
		ProjectID         string
		Index             int
		Address           []byte
		DraftCode         string
		DeployedCode      string
		DeployedContracts []string
		State             string
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := a.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}

	if err := a.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}

	a.Index = tmp.Index
	copy(a.Address[:], tmp.Address[:])
	a.DraftCode = tmp.DraftCode
	a.DeployedCode = tmp.DeployedCode
	a.DeployedContracts = tmp.DeployedContracts
	a.marshalledState = tmp.State
	return nil
}

func (a *InternalAccount) Save() ([]datastore.Property, error) {
	marshalledState, err := a.marshalAccountState()
	if err != nil {
		return nil, err
	}

	deployedContracts := []interface{}{}
	for _, contract := range a.DeployedContracts {
		deployedContracts = append(deployedContracts, contract)
	}

	return []datastore.Property{
		{
			Name:  "ID",
			Value: a.ID.String(),
		},
		{
			Name:  "ProjectID",
			Value: a.ProjectID.String(),
		},
		{
			Name:  "Index",
			Value: a.Index,
		},
		{
			Name:  "Address",
			Value: a.Address[:],
		},
		{
			Name:    "DraftCode",
			Value:   a.DraftCode,
			NoIndex: true,
		},
		{
			Name:    "DeployedCode",
			Value:   a.DeployedCode,
			NoIndex: true,
		},
		{
			Name:  "DeployedContracts",
			Value: deployedContracts,
		},
		{
			Name:    "State",
			Value:   marshalledState,
			NoIndex: true,
		},
	}, nil
}

func (a *InternalAccount) Export() *Account {
	return &Account{
		ID:                a.ID,
		ProjectID:         a.ProjectID,
		Index:             a.Index,
		Address:           a.Address,
		DraftCode:         a.DraftCode,
		DeployedCode:      a.DeployedCode,
		DeployedContracts: a.DeployedContracts,
		// NOTE: State left intentionally blank
	}
}

func (a *InternalAccount) ExportWithJSONState() (*Account, error) {

	exported := a.Export()

	encoded, err := a.marshalAccountState()
	if err != nil {
		return nil, err
	}

	if encoded != "" {
		exported.State = encoded
	}

	return exported, nil
}

type Account struct {
	ID                uuid.UUID
	ProjectID         uuid.UUID
	Index             int
	Address           Address
	DraftCode         string
	DeployedCode      string
	DeployedContracts []string
	State             string
}
