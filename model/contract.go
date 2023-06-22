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
	"github.com/pkg/errors"
)

type ContractTemplate = File

type ContractDeployment struct {
	File
	Address     Address        `gorm:"serializer:json"`
	Arguments   []string       `gorm:"serializer:json"`
	BlockHeight int            `json:"blockHeight"`
	Errors      []ProgramError `gorm:"serializer:json"`
	Events      []Event        `gorm:"serializer:json"`
	Logs        []string       `gorm:"serializer:json"`
}

func ContractDeploymentFromFlow(
	projectID uuid.UUID,
	contractName string,
	script string,
	arguments []string,
	result *flowsdk.TransactionResult,
	tx *flowsdk.Transaction,
	logs []string,
	blockHeight int,
) *ContractDeployment {
	signers := make([]Address, 0)
	// transaction could be nil in case where we get transaction result errors
	if tx != nil {
		for _, a := range tx.Authorizers {
			signers = append(signers, NewAddressFromBytes(a.Bytes()))
		}
	}

	deploy := &ContractDeployment{
		File: File{
			ID:        uuid.New(),
			Title:     contractName, // Parsed contract name
			ProjectID: projectID,
			Type:      ContractFile,
			Script:    script,
		},
		Arguments:   arguments,
		Address:     signers[0],
		BlockHeight: blockHeight,
		Errors:      nil,
		Events:      nil,
		Logs:        logs,
	}

	if result.Events != nil {
		events, _ := EventsFromFlow(result.Events)
		deploy.Events = events
	}

	if result.Error != nil {
		deploy.Errors = ProgramErrorFromFlow(result.Error)
	}

	return deploy
}

func (u *UpdateContractTemplate) Validate() error {
	if u.Title == nil && u.Index == nil && u.Script == nil {
		return errors.Wrap(missingValuesError, "title, index, script")
	}
	return nil
}
