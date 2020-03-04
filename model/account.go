package model

import (
	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Account struct {
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

func (a *Account) NameKey() *datastore.Key {
	return datastore.NameKey("Account", a.ID.String(), nil)
}

func (a *Account) Load(ps []datastore.Property) error {
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

func (a *Account) Save() ([]datastore.Property, error) {
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
