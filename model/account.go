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

/*
import (
	"github.com/google/uuid"
	flowsdk "github.com/onflow/flow-go-sdk"
)

func (a *Account) Export() *Account {
	return &Account{
		ID:                a.ID,
		ProjectID:         a.ProjectID,
		Address:           a.Address,
		DeployedContracts: a.DeployedContracts,
		State:             a.State,
	}
}

type Account struct {
	ID                uuid.UUID
	ProjectID         uuid.UUID
	Index             int
	Address           Address  `gorm:"serializer:json"`
	DeployedContracts []string `gorm:"serializer:json"`
	State             string
}

func (a *Account) MergeFromStore(acc *Account) *Account {
	a.ID = acc.ID
	return a
}

type UpdateAccount struct {
	ID                uuid.UUID `json:"id"`
	ProjectID         uuid.UUID `json:"projectId"`
	DeployedContracts *[]string
}


func (u *UpdateAccount) Validate() error {

		if u.DeployedCode == nil && u.DraftCode == nil {
			return errors.Wrap(missingValuesError, "deployed code, draft code")
		}
		if u.DeployedCode != nil && u.DraftCode != nil {
			return fmt.Errorf("can only provide deployed code or draft code")
		}


	return nil
}

func AccountFromFlow(account *flowsdk.Account, projectID uuid.UUID) *Account {
	contractNames := make([]string, 0)
	contractCode := ""
	for name, code := range account.Contracts {
		contractNames = append(contractNames, name)
		contractCode = string(code)
		break // we only allow one deployed contract on account so only get the first if present
	}

	_ = contractCode

	return &Account{
		ProjectID:         projectID,
		Address:           NewAddressFromBytes(account.Address.Bytes()),
		DeployedContracts: contractNames,
	}
}
*/
