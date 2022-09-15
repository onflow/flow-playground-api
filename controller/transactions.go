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

type Transactions struct {
	store      storage.Store
	blockchain *blockchain.Projects
}

func NewTransactions(
	store storage.Store,
	blockchain *blockchain.Projects,
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
	if input.Script == "" {
		return nil, errors.New("cannot execute empty transaction script")
	}

	exe, err := t.blockchain.ExecuteTransaction(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute transaction")
	}

	return exe, nil
}
