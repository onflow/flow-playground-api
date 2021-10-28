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

type Contract struct {
	ProjectChildID
	AccountID      uuid.UUID
	Title          string
	Index          int
	Script         string
	DeployedScript string
}

func (c *Contract) NameKey() *datastore.Key {
	return datastore.NameKey("Contract", c.ID.String(), ProjectNameKey(c.ProjectID))
}

func (c *Contract) Load(ps []datastore.Property) error {
	tmp := struct {
		ID             string
		ProjectID      string
		AccountID      string
		Title          string
		Index          int
		Script         string
		DeployedScript string
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := c.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode contract UUID")
	}
	if err := c.ProjectID.UnmarshalText([]byte(tmp.ProjectID)); err != nil {
		return errors.Wrap(err, "failed to decode project UUID")
	}
	if err := c.AccountID.UnmarshalText([]byte(tmp.AccountID)); err != nil {
		return errors.Wrap(err, "failed to decode account UUID")
	}

	c.Title = tmp.Title
	c.Index = tmp.Index
	c.Script = tmp.Script
	c.DeployedScript = tmp.DeployedScript
	return nil
}

func (c *Contract) Save() ([]datastore.Property, error) {
	return []datastore.Property{
		{
			Name:  "ID",
			Value: c.ID.String(),
		},
		{
			Name:  "ProjectID",
			Value: c.ProjectID.String(),
		},
		{
			Name:  "AccountID",
			Value: c.AccountID.String(),
		},
		{
			Name:  "Title",
			Value: c.Title,
		},
		{
			Name:  "Index",
			Value: c.Index,
		},
		{
			Name:    "Script",
			Value:   c.Script,
			NoIndex: true,
		},
		{
			Name:    "DeployedScript",
			Value:   c.DeployedScript,
			NoIndex: true,
		},
	}, nil
}
