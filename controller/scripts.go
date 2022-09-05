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
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Scripts struct {
	store      storage.Store
	blockchain *blockchain.State
}

func NewScripts(
	store storage.Store,
	blockchain *blockchain.State,
) *Scripts {
	return &Scripts{
		store:      store,
		blockchain: blockchain,
	}
}

func (s *Scripts) CreateTemplate(projectID uuid.UUID, input model.NewScriptTemplate) (*model.ScriptTemplate, error) {
	tpl := model.ScriptTemplate{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: projectID,
		},
		Title:  input.Title,
		Script: input.Script,
	}

	err := s.store.InsertScriptTemplate(&tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store script template")
	}

	return &tpl, nil
}

func (s *Scripts) UpdateTemplate(input model.UpdateScriptTemplate) (*model.ScriptTemplate, error) {
	var tpl model.ScriptTemplate

	err := s.store.UpdateScriptTemplate(input, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update script template")
	}

	return &tpl, nil
}

func (s *Scripts) DeleteTemplate(scriptID, projectID uuid.UUID) error {
	err := s.store.DeleteScriptTemplate(model.NewProjectChildID(scriptID, projectID))
	if err != nil {
		return errors.Wrap(err, "failed to delete script template")
	}

	return nil
}

func (s *Scripts) AllTemplatesForProjectID(ID uuid.UUID) ([]*model.ScriptTemplate, error) {
	var templates []*model.ScriptTemplate
	err := s.store.GetScriptTemplatesForProject(ID, &templates)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get script templates")
	}

	return templates, nil
}

func (s *Scripts) TemplateByID(ID uuid.UUID, projectID uuid.UUID) (*model.ScriptTemplate, error) {
	var tpl model.ScriptTemplate
	err := s.store.GetScriptTemplate(model.NewProjectChildID(ID, projectID), &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get script template")
	}

	return &tpl, nil
}

func (s *Scripts) CreateExecution(script model.NewScriptExecution) (*model.ScriptExecution, error) {
	if len(script.Script) == 0 {
		return nil, errors.New("cannot execute empty script")
	}

	execution, err := s.blockchain.ExecuteScript(script)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute script")
	}
	return execution, nil
}
