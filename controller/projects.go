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

package controller

import (
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Projects struct {
	version     *semver.Version
	store       storage.Store
	numAccounts int
	blockchain  blockchain.Blockchain
}

func NewProjects(
	version *semver.Version,
	store storage.Store,
	numAccounts int,
	blockchain blockchain.Blockchain,
) *Projects {
	return &Projects{
		version:     version,
		store:       store,
		numAccounts: numAccounts,
		blockchain:  blockchain,
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

	accounts, err := p.createInitialAccounts(proj.ID)
	if err != nil {
		return nil, err
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

	err = p.store.CreateProject(proj, accounts, ttpls, stpls)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create project")
	}

	return proj, nil
}

func (p *Projects) createInitialAccounts(projectID uuid.UUID) ([]*model.InternalAccount, error) {
	addresses, err := p.deployInitialAccounts()
	if err != nil {
		return nil, err
	}

	accounts := make([]*model.InternalAccount, len(addresses))

	for i, address := range addresses {
		account := model.InternalAccount{
			ProjectChildID: model.ProjectChildID{
				ID:        uuid.New(),
				ProjectID: projectID,
			},
			Address: address,
		}

		accounts[i] = &account
	}

	return accounts, nil
}

func (p *Projects) deployInitialAccounts() ([]model.Address, error) {
	addresses := make([]model.Address, p.numAccounts)

	for i := 0; i < p.numAccounts; i++ {
		account, _, err := p.blockchain.CreateAccount()
		if err != nil {
			return nil, err
		}

		addresses[i] = model.Address(account.Address)
	}

	return addresses, nil
}

func (p *Projects) Get(id uuid.UUID, proj *model.InternalProject) error {
	err := p.store.GetProject(id, proj)
	if err != nil {
		return errors.Wrap(err, "failed to get project")
	}

	return nil
}

func (p *Projects) Update(input model.UpdateProject, proj *model.InternalProject) error {
	err := p.store.UpdateProject(input, proj)
	if err != nil {
		return errors.Wrap(err, "failed to update project")
	}

	return nil
}

func (p *Projects) UpdateVersion(id uuid.UUID, version *semver.Version) error {
	err := p.store.UpdateProjectVersion(id, version)
	if err != nil {
		return errors.Wrap(err, "failed to save project version")
	}

	return nil
}

func (p *Projects) Reset(proj *model.InternalProject) error {
	// todo reset emulator state

	_, err := p.deployInitialAccounts()
	if err != nil {
		return err
	}

	err = p.store.ResetProjectState(proj)
	if err != nil {
		return err
	}

	// todo clear cache

	return nil
}
