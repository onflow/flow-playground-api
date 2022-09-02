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
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type InternalAccount struct {
	ProjectChildID
	Address   Address
	DraftCode string
}

func (a *InternalAccount) NameKey() *datastore.Key {
	return datastore.NameKey("Account", a.ID.String(), ProjectNameKey(a.ProjectID))
}

func (a *InternalAccount) Load(ps []datastore.Property) error {
	tmp := struct {
		ID        string
		ProjectID string
		Address   []byte
		DraftCode string
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

	copy(a.Address[:], tmp.Address[:])

	a.DraftCode = tmp.DraftCode

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
			Name:  "Address",
			Value: a.Address[:],
		},
		{
			Name:  "DraftCode",
			Value: a.DraftCode,
		},
	}, nil
}

func (a *InternalAccount) Export() *Account {
	return &Account{
		ID:        a.ID,
		ProjectID: a.ProjectID,
		Address:   a.Address,
		DraftCode: a.DraftCode,
	}
}

type Account struct {
	ID                uuid.UUID
	ProjectID         uuid.UUID
	Address           Address
	DraftCode         string
	DeployedCode      string
	DeployedContracts []string
	State             string
}

type UpdateAccount struct {
	ID                uuid.UUID `json:"id"`
	ProjectID         uuid.UUID `json:"projectId"`
	DraftCode         *string   `json:"draftCode"`
	DeployedCode      *string   `json:"deployedCode"`
	DeployedContracts *[]string
}
