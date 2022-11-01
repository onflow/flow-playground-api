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

package model

import (
	"github.com/google/uuid"
	flowsdk "github.com/onflow/flow-go-sdk"
)

type Account struct {
	ProjectID         uuid.UUID
	Address           Address
	DeployedContracts []string
	State             string
}

func AccountFromFlow(account *flowsdk.Account, projectID uuid.UUID) *Account {
	contractNames := make([]string, 0)
	for name := range account.Contracts {
		contractNames = append(contractNames, name)
	}

	return &Account{
		ProjectID:         projectID,
		Address:           NewAddressFromBytes(account.Address.Bytes()),
		DeployedContracts: contractNames,
	}
}

func (a *Account) Export() *Account {
	return &Account{
		ProjectID:         a.ProjectID,
		Address:           a.Address,
		DeployedContracts: a.DeployedContracts,
		State:             a.State,
	}
}
