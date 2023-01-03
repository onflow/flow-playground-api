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
	NumberOfAccounts          int
	TransactionExecutionCount int
	Persist                   bool
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	AccessedAt                time.Time
	Version                   *semver.Version `gorm:"serializer:json"`
	Mutable                   bool            // todo don't persist this
}

func (p *Project) IsOwnedBy(userID uuid.UUID) bool {
	return p.UserID == userID
}

// ExportPublicMutable converts the internal project to its public representation
// and marks it as mutable.
func (p *Project) ExportPublicMutable() *Project {
	return &Project{
		ID:               p.ID,
		Title:            p.Title,
		Description:      p.Description,
		Readme:           p.Readme,
		PublicID:         p.PublicID,
		ParentID:         p.ParentID,
		Persist:          p.Persist,
		Seed:             p.Seed,
		NumberOfAccounts: p.NumberOfAccounts,
		Version:          p.Version,
		UpdatedAt:        p.UpdatedAt,
		Mutable:          true,
	}
}

// ExportPublicImmutable converts the internal project to its public representation
// and marks it as immutable.
func (p *Project) ExportPublicImmutable() *Project {
	return &Project{
		ID:               p.ID,
		Title:            p.Title,
		Description:      p.Description,
		Readme:           p.Readme,
		PublicID:         p.PublicID,
		ParentID:         p.ParentID,
		Persist:          p.Persist,
		Seed:             p.Seed,
		NumberOfAccounts: p.NumberOfAccounts,
		Version:          p.Version,
		UpdatedAt:        p.UpdatedAt,
		Mutable:          false,
	}
}

func (u *UpdateProject) Validate() error {
	if u.Title == nil && u.Readme == nil && u.Persist == nil && u.Description == nil {
		return errors.Wrap(missingValuesError, "title, readme, persist, description")
	}
	return nil
}
