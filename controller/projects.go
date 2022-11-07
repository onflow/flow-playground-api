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
	"context"
	"github.com/Masterminds/semver"
	"github.com/dapperlabs/flow-playground-api/auth"
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"strconv"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

// MaxProjectsLimit limit on the number of projects a user can create
const MaxProjectsLimit = 10

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
	var projectCount int64
	err := p.store.GetProjectCountForUser(user.ID, &projectCount)
	if projectCount >= MaxProjectsLimit {
		return nil, errors.New("maximum number of" +
			strconv.Itoa(MaxProjectsLimit) + "projects reached.")
	}

	proj := &model.Project{
		ID:               uuid.New(),
		Secret:           uuid.New(),
		PublicID:         uuid.New(),
		ParentID:         input.ParentID,
		Seed:             input.Seed,
		Title:            input.Title,
		Description:      input.Description,
		Readme:           input.Readme,
		Persist:          false,
		NumberOfAccounts: input.NumberOfAccounts,
		Version:          p.version,
		UserID:           user.ID,
	}

	files := make([]*model.File, 0)

	for _, tpl := range input.ContractTemplates {
		files = append(files, &model.File{
			ID:        uuid.New(),
			ProjectID: proj.ID,
			Title:     tpl.Title,
			Script:    tpl.Script,
			Type:      model.ContractFile,
		})
	}

	for _, tpl := range input.TransactionTemplates {
		files = append(files, &model.File{
			ID:        uuid.New(),
			ProjectID: proj.ID,
			Title:     tpl.Title,
			Script:    tpl.Script,
			Type:      model.TransactionFile,
		})
	}

	for _, tpl := range input.ScriptTemplates {
		files = append(files, &model.File{
			ID:        uuid.New(),
			ProjectID: proj.ID,
			Title:     tpl.Title,
			Script:    tpl.Script,
			Type:      model.ScriptFile,
		})
	}

	err = p.store.CreateProject(proj, files)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create project")
	}

	_, err = p.blockchain.CreateInitialAccounts(proj.ID)
	if err != nil {
		return nil, err
	}

	return proj, nil
}

func (p *Projects) Delete(id uuid.UUID) error {
	var proj model.Project
	err := p.store.GetProject(id, &proj)
	if err != nil {
		return err
	}

	err = p.store.DeleteProject(id)
	if err != nil {
		return err
	}

	return nil
}

func (p *Projects) Get(id uuid.UUID) (*model.Project, error) {
	var proj model.Project
	err := p.store.GetProject(id, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	return &proj, nil
}

func (p *Projects) GetProjectListForUser(userID uuid.UUID, auth *auth.Authenticator, ctx context.Context) (*model.ProjectList, error) {
	var projects []*model.Project
	err := p.store.GetAllProjectsForUser(userID, &projects)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get projects for user "+userID.String())
	}

	exportedProjects := make([]*model.Project, len(projects))

	for i, proj := range projects {
		if err := auth.CheckProjectAccess(ctx, proj); err != nil {
			exportedProjects[i] = proj.ExportPublicImmutable()
		} else {
			exportedProjects[i] = proj.ExportPublicMutable()
		}
	}

	return &model.ProjectList{Projects: exportedProjects}, nil
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
	return p.blockchain.Reset(proj.ID, nil)
}
