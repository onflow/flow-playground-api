package model

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"cloud.google.com/go/datastore"
	"github.com/dapperlabs/flow-go/language/runtime/interpreter"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/encoding"
)

type InternalAccount struct {
	ID                uuid.UUID
	ProjectID         uuid.UUID
	Index             int
	Address           Address
	DraftCode         string
	DeployedCode      string
	DeployedContracts []string
	State             map[string][]byte
}

type UpdateAccount struct {
	ID                uuid.UUID `json:"id"`
	DraftCode         *string   `json:"draftCode"`
	DeployedCode      *string   `json:"deployedCode"`
	DeployedContracts *[]string
}

func (a *InternalAccount) NameKey() *datastore.Key {
	return datastore.NameKey("Account", a.ID.String(), nil)
}

func (a *InternalAccount) Load(ps []datastore.Property) error {
	tmp := struct {
		ID           string
		ProjectID    string
		Index        int
		Address      []byte
		DraftCode    string
		DeployedCode string
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
	return nil
}

func (a *InternalAccount) Save() ([]datastore.Property, error) {
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
			Name:  "DraftCode",
			Value: a.DraftCode,
		},
		{
			Name:  "DeployedCode",
			Value: a.DeployedCode,
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
	state := make(map[string]encoding.Value, len(a.State))

	for key, valueData := range a.State {
		if len(valueData) == 0 {
			continue
		}

		var interpreterValue interpreter.Value

		decoder := gob.NewDecoder(bytes.NewReader(valueData))
		err := decoder.Decode(&interpreterValue)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode value")
		}

		convertedValue, err := encoding.ConvertValue(interpreterValue)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert value")
		}

		state[key] = convertedValue
	}

	encoded, err := json.Marshal(state)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode to JSON")
	}

	exported := a.Export()
	exported.State = string(encoded)

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
