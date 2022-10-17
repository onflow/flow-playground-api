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
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
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

func (p *Projects) Create(user *model.User, input model.NewProject) (*model.Project, error) {
	// TODO: Needs to take contract templates from input as well
	proj := &model.Project{
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
		UserID:      user.ID,
	}

	// TODO: Need to convert all to files
	ctrcts := make([]*model.File, len(input.ContractTemplates))

	ttpls := make([]*model.File, len(input.TransactionTemplates))
	for i, tpl := range input.TransactionTemplates {
		ttpls[i] = &model.File{
			ID:        uuid.New(),
			ProjectID: proj.ID,
			Title:     tpl.Title,
			Script:    tpl.Script,
		}
	}

	stpls := make([]*model.File, len(input.ScriptTemplates))
	for i, tpl := range input.ScriptTemplates {
		stpls[i] = &model.File{
			ID:        uuid.New(),
			ProjectID: proj.ID,
			Title:     tpl.Title,
			Script:    tpl.Script,
		}
	}

	files := make([]*model.File, len(stpls)+len(ttpls)+len())

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
	}

	err = p.store.InsertAccounts(accounts)
	if err != nil {
		sentry.CaptureException(err)
		return nil, err
	}

	return proj, nil
}

func (p *Projects) Get(id uuid.UUID) (*model.Project, error) {
	var proj model.Project
	err := p.store.GetProject(id, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	return &proj, nil
}

func (p *Projects) Update(input model.UpdateProject) (*model.Project, error) {
	var proj model.Project
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

// Reset is not used in the API but for migration
func (p *Projects) Reset(proj *model.Project) ([]*model.Account, error) {
	return p.blockchain.Reset(proj)
}
