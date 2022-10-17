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
	// TODO implement
	return nil, nil
}
