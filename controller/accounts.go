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
	var acc model.Account

	err := a.store.GetAccount(ID, projectID, &acc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account")
	}

	account, err := a.blockchain.GetAccount(projectID, acc.Address)
	if err != nil {
		return nil, err
	}

	return account.
		MergeFromStore(&acc).
		Export(), nil
}

func (a *Accounts) AllForProjectID(projectID uuid.UUID) ([]*model.Account, error) {
	var accounts []*model.Account

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

		acc.MergeFromStore(account)
		exported[i] = acc.Export()
	}

	return exported, nil
}

func (a *Accounts) Update(input model.UpdateAccount) (*model.Account, error) {
	if input.UpdateCode() {
		return a.updateCode(input)
	}

	return a.deployCode(input)
}

// updateCode only updates the database code of an account.
func (a *Accounts) updateCode(input model.UpdateAccount) (*model.Account, error) {
	var acc model.Account
	err := a.store.UpdateAccount(input, &acc)
	if err != nil {
		return nil, err
	}

	return acc.Export(), nil
}

// deployCode deploys code on the flow network.
func (a *Accounts) deployCode(input model.UpdateAccount) (*model.Account, error) {
	var dbAccount model.Account
	err := a.store.GetAccount(input.ID, input.ProjectID, &dbAccount)
	if err != nil {
		return nil, err
	}

	flowAccount, err := a.blockchain.GetAccount(input.ProjectID, dbAccount.Address)
	if err != nil {
		return nil, err
	}

	// reset the state first if already contains deployed code
	if flowAccount.HasDeployedCode() {
		var proj model.Project
		err := a.store.GetProject(input.ProjectID, &proj)
		if err != nil {
			return nil, err
		}

		_, err = a.blockchain.Reset(&proj)
		if err != nil {
			return nil, err
		}
	}

	flowAccount, err = a.blockchain.DeployContract(input.ProjectID, dbAccount.Address, *input.DeployedCode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy account code")
	}

	return flowAccount.
		MergeFromStore(&dbAccount).
		Export(), nil
}
