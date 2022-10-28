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
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Accounts struct {
	blockchain *blockchain.Projects
}

func NewAccounts(
	blockchain *blockchain.Projects,
) *Accounts {
	return &Accounts{
		blockchain: blockchain,
	}
}

func (a *Accounts) GetByAddress(address model.Address, projectID uuid.UUID) (*model.Account, error) {
	var acc model.Account

	/*
		err := a.store.GetAccount(ID, projectID, &acc)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get account")
		}
	*/

	account, err := a.blockchain.GetAccount(projectID, acc.Address)
	if err != nil {
		return nil, err
	}

	return account.Export(), nil
}

func (a *Accounts) AllForProjectID(projectID uuid.UUID) ([]*model.Account, error) {
	var accounts []*model.Account

	// TODO FIX
	err := a.store.GetAccountsForProject(projectID, &accounts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get accounts")
	}

	addresses := make([]model.Address, len(accounts))
	for i, account := range accounts {
		addresses[i] = account.Address
	}

	accs, err := a.blockchain.GetAccounts(projectID, addresses)
	if err != nil {
		return nil, err
	}

	exported := make([]*model.Account, len(accounts))
	for i, account := range accounts {
		accs[i].MergeFromStore(account)
		exported[i] = accs[i].Export()
	}

	return exported, nil
}
