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

func (a *Accounts) GetByAddress(address model.Address, projectID uuid.UUID) (*model.Account, error) {
	account, err := a.blockchain.GetAccount(projectID, address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account by address")
	}
	return account.Export(), nil
}

func (a *Accounts) AllForProjectID(projectID uuid.UUID) ([]*model.Account, error) {
	var proj model.Project
	err := a.store.GetProject(projectID, &proj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all project accounts")
	}

	addresses := make([]model.Address, proj.NumberOfAccounts)
	for i := 0; i < proj.NumberOfAccounts; i++ {
		addresses[i] = model.NewAddressFromIndex(i)
	}

	accs, err := a.blockchain.GetAccounts(projectID, addresses)
	if err != nil {
		return nil, err
	}

	exported := make([]*model.Account, proj.NumberOfAccounts)
	for i := 0; i < proj.NumberOfAccounts; i++ {
		exported[i] = accs[i].Export()
	}

	return exported, nil
}
