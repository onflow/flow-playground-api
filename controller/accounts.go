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
	"github.com/dapperlabs/flow-playground-api/server/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Accounts struct {
	store      storage.Store
	blockchain *blockchain.Projects
}

func NewAccounts(
	store storage.Store,
	blockchain *blockchain.Projects,
) *Accounts {
	return &Accounts{
		store:      store,
		blockchain: blockchain,
	}
}

func (a *Accounts) GetByID(ID uuid.UUID, projectID uuid.UUID) (*model.Account, error) {
	var acc model.InternalAccount

	err := a.store.GetAccount(model.NewProjectChildID(ID, projectID), &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account")
	}

	account, err := a.blockchain.GetAccount(projectID, acc.Address)
	if err != nil {
		return nil, err
	}

	account.ID = acc.ID
	account.DraftCode = acc.DraftCode
	return account, nil
}

func (a *Accounts) AllForProjectID(projectID uuid.UUID) ([]*model.Account, error) {
	var accounts []*model.InternalAccount

	err := a.store.GetAccountsForProject(projectID, &accounts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get accounts")
	}

	exported := make([]*model.Account, len(accounts))
	for i, account := range accounts {
		acc, err := a.blockchain.GetAccount(projectID, account.Address)
		if err != nil {
			return nil, err
		}

		acc.ID = account.ID
		acc.DraftCode = account.DraftCode
		exported[i] = acc
	}

	return exported, nil
}

func (a *Accounts) Update(input model.UpdateAccount) (*model.Account, error) {
	var acc model.InternalAccount

	// if we provided draft code then just do a storage update of an account
	if input.DeployedCode == nil {
		err := a.store.UpdateAccount(input, &acc)
		if err != nil {
			return nil, err
		}

		return acc.Export(), nil
	}

	err := a.store.GetAccount(model.NewProjectChildID(input.ID, input.ProjectID), &acc)
	if err != nil {
		return nil, err
	}

	account, err := a.blockchain.GetAccount(input.ProjectID, acc.Address)
	if err != nil {
		return nil, err
	}

	if account.DeployedCode != "" {
		var proj model.InternalProject
		err := a.store.GetProject(input.ProjectID, &proj)
		if err != nil {
			return nil, err
		}

		err = a.blockchain.Reset(&proj)
		if err != nil {
			return nil, err
		}
	}

	account, err = a.blockchain.DeployContract(input.ProjectID, acc.Address, *input.DeployedCode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy account code")
	}

	account.DraftCode = acc.DraftCode
	account.ID = acc.ID
	return account, nil
}
