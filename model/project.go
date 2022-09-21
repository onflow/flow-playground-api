/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/uuid"
)

type Project struct {
	ID                        uuid.UUID
	UserID                    uuid.UUID
	Secret                    uuid.UUID
	PublicID                  uuid.UUID
	ParentID                  *uuid.UUID
	Title                     string
	Description               string
	Readme                    string
	Seed                      int
	TransactionCount          int
	TransactionExecutionCount int
	TransactionTemplateCount  int
	ScriptTemplateCount       int
	Persist                   bool
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	Version                   *semver.Version
	Mutable                   bool // todo don't persist this
}

func (p *Project) IsOwnedBy(userID uuid.UUID) bool {
	return p.UserID == userID
}

// ExportPublicMutable converts the internal project to its public representation
// and marks it as mutable.
func (p *Project) ExportPublicMutable() *Project {
	return &Project{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Readme:      p.Readme,
		PublicID:    p.PublicID,
		ParentID:    p.ParentID,
		Persist:     p.Persist,
		Seed:        p.Seed,
		Version:     p.Version,
		Mutable:     true,
	}
}

// ExportPublicImmutable converts the internal project to its public representation
// and marks it as immutable.
func (p *Project) ExportPublicImmutable() *Project {
	return &Project{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Readme:      p.Readme,
		PublicID:    p.PublicID,
		ParentID:    p.ParentID,
		Persist:     p.Persist,
		Seed:        p.Seed,
		Version:     p.Version,
		Mutable:     false,
	}
}

// Load is used for datastore to SQL migration
func (p *Project) Load(ps []datastore.Property) error {
	tmp := struct {
		ID                        string
		UserID                    string
		Secret                    string
		PublicID                  string
		ParentID                  *string
		Title                     string
		Description               string
		Readme                    string
		Seed                      int
		TransactionCount          int
		TransactionExecutionCount int
		TransactionTemplateCount  int
		ScriptTemplateCount       int
		Persist                   bool
		CreatedAt                 time.Time
		UpdatedAt                 time.Time
		Version                   *string
	}{}

	if err := datastore.LoadStruct(&tmp, ps); err != nil {
		return err
	}

	if err := p.ID.UnmarshalText([]byte(tmp.ID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}

	if tmp.UserID != "" {
		if err := p.UserID.UnmarshalText([]byte(tmp.UserID)); err != nil {
			return errors.Wrap(err, "failed to decode UUID")
		}
	}

	if tmp.Secret != "" {
		if err := p.Secret.UnmarshalText([]byte(tmp.Secret)); err != nil {
			return errors.Wrap(err, "failed to decode UUID")
		}
	}

	if err := p.PublicID.UnmarshalText([]byte(tmp.PublicID)); err != nil {
		return errors.Wrap(err, "failed to decode UUID")
	}

	if tmp.ParentID != nil && len(*tmp.ParentID) != 0 {
		p.ParentID = new(uuid.UUID)
		if err := p.ParentID.UnmarshalText([]byte(*tmp.ParentID)); err != nil {
			return errors.Wrap(err, "failed to decode UUID")
		}
	} else {
		p.ParentID = nil
	}

	if tmp.Version != nil && len(*tmp.Version) != 0 {
		var err error
		p.Version, err = semver.NewVersion(*tmp.Version)
		if err != nil {
			return errors.Wrap(err, "failed to parse project version")
		}
	}

	p.Title = tmp.Title
	p.Description = tmp.Description
	p.Readme = tmp.Readme
	p.Seed = tmp.Seed
	p.TransactionCount = tmp.TransactionCount
	p.TransactionExecutionCount = tmp.TransactionExecutionCount
	p.TransactionTemplateCount = tmp.TransactionTemplateCount
	p.ScriptTemplateCount = tmp.ScriptTemplateCount
	p.Persist = tmp.Persist

	p.CreatedAt = tmp.CreatedAt
	p.UpdatedAt = tmp.UpdatedAt

	return nil
}
