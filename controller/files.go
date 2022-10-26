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
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Files struct {
	store      storage.Store
	blockchain *blockchain.Projects
}

func NewFiles(
	store storage.Store,
	blockchain *blockchain.Projects,
) *Files {
	return &Files{
		store:      store,
		blockchain: blockchain,
	}
}

func (f *Files) CreateFile(projectID uuid.UUID, input model.NewFile, fileType model.FileType) (*model.File, error) {
	file := model.File{
		ID:        uuid.New(),
		ProjectID: projectID,
		Title:     input.Title,
		Type:      fileType,
		Script:    input.Script,
	}

	err := f.store.InsertCadenceFile(&file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store file")
	}

	return &file, nil
}

func (f *Files) UpdateFile(input model.UpdateFile) (*model.File, error) {
	var file model.File
	err := f.store.UpdateCadenceFile(input, &file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update cadence file")
	}

	return &file, nil
}

func (f *Files) DeleteFile(scriptID, projectID uuid.UUID) error {
	err := f.store.DeleteCadenceFile(scriptID, projectID)
	if err != nil {
		return errors.Wrap(err, "failed to delete cadence file")
	}

	return nil
}

func (f *Files) CreateScriptExecution(input model.NewScriptExecution) (*model.ScriptExecution, error) {
	if len(input.Script) == 0 {
		return nil, errors.New("cannot execute empty script")
	}

	execution, err := f.blockchain.ExecuteScript(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute script")
	}
	return execution, nil
}

func (f *Files) CreateTransactionExecution(input model.NewTransactionExecution) (*model.TransactionExecution, error) {
	if len(input.Script) == 0 {
		return nil, errors.New("cannot execute empty transaction script")
	}

	exe, err := f.blockchain.ExecuteTransaction(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute transaction")
	}

	return exe, nil
}

func (f *Files) DeployContract(input model.NewContractDeployment) (*model.ContractDeployment, error) {
	if len(input.Script) == 0 {
		return nil, errors.New("cannot deploy empty contract")
	}

	deploy, err := f.blockchain.DeployContract(input.ProjectID, input.Address, input.Script)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy contract")
	}

	return deploy, nil
}

func (f *Files) GetFilesForProject(projID uuid.UUID, fileType model.FileType) ([]*model.File, error) {
	var files []*model.File

	err := f.store.GetFilesForProject(projID, &files, fileType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get files")
	}

	return files, nil
}

func (f *Files) GetFile(id uuid.UUID, projID uuid.UUID) (*model.File, error) {
	var file model.File
	err := f.store.GetFile(id, projID, &file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get file")
	}

	return &file, nil
}
