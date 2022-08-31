package controller

import (
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Transactions struct {
	store      storage.Store
	blockchain blockchain.Blockchain
}

func NewTransactions(
	store storage.Store,
	blockchain blockchain.Blockchain,
) *Transactions {
	return &Transactions{
		store:      store,
		blockchain: blockchain,
	}
}

func (t *Transactions) CreateTemplate(projectID uuid.UUID, title string, script string) (*model.TransactionTemplate, error) {
	tpl := model.TransactionTemplate{
		ProjectChildID: model.ProjectChildID{
			ID:        uuid.New(),
			ProjectID: projectID,
		},
		Title:  title,
		Script: script,
	}

	err := t.store.InsertTransactionTemplate(&tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store script template")
	}

	return &tpl, nil
}

func (t *Transactions) UpdateTemplate(input model.UpdateTransactionTemplate) (*model.TransactionTemplate, error) {
	var tpl model.TransactionTemplate

	err := t.store.UpdateTransactionTemplate(input, &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update transaction template")
	}

	return &tpl, nil
}

func (t *Transactions) DeleteTemplate(transactionID, projectID uuid.UUID) error {
	err := t.store.DeleteTransactionTemplate(model.NewProjectChildID(transactionID, projectID))
	if err != nil {
		return errors.Wrap(err, "failed to delete transaction template")
	}

	return nil
}

func (t *Transactions) TemplateByID(ID uuid.UUID, projectID uuid.UUID) (*model.TransactionTemplate, error) {
	var tpl model.TransactionTemplate

	err := t.store.GetTransactionTemplate(model.NewProjectChildID(ID, projectID), &tpl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction template")
	}

	return &tpl, nil
}

func (t *Transactions) AllTemplatesForProjectID(ID uuid.UUID) ([]*model.TransactionTemplate, error) {
	var templates []*model.TransactionTemplate
	err := t.store.GetTransactionTemplatesForProject(ID, &templates)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction templates")
	}

	return templates, nil
}

func (t *Transactions) AllExecutionsForProjectID(ID uuid.UUID) ([]*model.TransactionExecution, error) {
	var exes []*model.TransactionExecution

	err := t.store.GetTransactionExecutionsForProject(ID, &exes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction executions")
	}

	return exes, nil
}

func (t *Transactions) CreateTransactionExecution(input model.NewTransactionExecution) (*model.TransactionExecution, error) {
	result, err := t.blockchain.ExecuteTransaction(input.Script, input.Arguments, input.Signers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute transaction")
	}

	exe, err := model.TransactionExecutionFromFlow(
		result,
		input.ProjectID,
		input.Script,
		input.Arguments,
		input.Signers,
	)
	if err != nil {
		return nil, err
	}

	err = t.store.InsertTransactionExecution(exe)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert transaction execution record")
	}

	return exe, nil
}
