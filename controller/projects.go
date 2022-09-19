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

package controller

import (
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/server/storage"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/model"
)

type Projects struct {
	version    *semver.Version
	store      storage.Store
	blockchain *blockchain.Projects
}

func NewProjects(
	version *semver.Version,
	store storage.Store,
	blockchain *blockchain.Projects,
) *Projects {
	return &Projects{
		version:    version,
		store:      store,
		blockchain: blockchain,
	}
}

func (p *Projects) Create(user *model.User, input model.NewProject) (*model.InternalProject, error) {
	proj := &model.InternalProject{
		ID:          uuid.New(),
		Secret:      uuid.New(),
		PublicID:    uuid.New(),
		ParentID:    input.ParentID,
		Seed:        input.Seed,
		Title:       input.Title,
		Description: input.Description,
		Readme:      input.Readme,
		Persist:     false,
		Version:     p.version,
	}

	ttpls := make([]*model.TransactionTemplate, len(input.TransactionTemplates))

	for i, tpl := range input.TransactionTemplates {
		ttpl := &model.TransactionTemplate{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Title:  tpl.Title,
			Script: tpl.Script,
		}

		ttpls[i] = ttpl
	}

	stpls := make([]*model.ScriptTemplate, len(input.ScriptTemplates))

	for i, tpl := range input.ScriptTemplates {
		stpl := &model.ScriptTemplate{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: proj.ID,
			},
			Title:  tpl.Title,
			Script: tpl.Script,
		}

		stpls[i] = stpl
	}

	proj.UserID = user.ID

	err := p.store.CreateProject(proj, ttpls, stpls)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create project")
	}

	accounts, err := p.blockchain.CreateInitialAccounts(proj.ID)
	if err != nil {
		return nil, err
	}

	for i, account := range accounts {
		if i < len(input.Accounts) {
			account.DraftCode = input.Accounts[i]
		}

		err := p.store.InsertAccount(account)
		if err != nil {
			sentry.CaptureException(err)
			return nil, err
		}
	}

	return proj, nil
}

func (p *Projects) Get(id uuid.UUID) (*model.InternalProject, error) {
	var proj model.InternalProject
	err := p.store.GetProject(id, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	return &proj, nil
}

func (p *Projects) Update(input model.UpdateProject) (*model.InternalProject, error) {
	var proj model.InternalProject
	err := p.store.UpdateProject(input, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update project")
	}

	return &proj, nil
}

func (p *Projects) UpdateVersion(id uuid.UUID, version *semver.Version) error {
	err := p.store.UpdateProjectVersion(id, version)
	if err != nil {
		return errors.Wrap(err, "failed to save project version")
	}

	return nil
}

func (p *Projects) Reset(proj *model.InternalProject) error {
	err := p.blockchain.Reset(proj)
	if err != nil {
		return err
	}

	return nil
}
