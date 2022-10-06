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
	"github.com/onflow/flow-emulator/types"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
)

type TransactionTemplate struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Title     string
	Index     int
	Script    string
}

func TransactionExecutionFromFlow(
	projectID uuid.UUID,
	result *types.TransactionResult,
	tx *flowsdk.Transaction,
) *TransactionExecution {
	args := make([]string, 0)
	signers := make([]Address, 0)
	script := ""
	// transaction could be nil in case where we get transaction result errors
	if tx != nil {
		for _, a := range tx.Arguments {
			args = append(args, string(a))
		}

		for _, a := range tx.Authorizers {
			signers = append(signers, NewAddressFromBytes(a.Bytes()))
		}

		script = string(tx.Script)
	}

	exe := &TransactionExecution{
		ID:        uuid.New(),
		ProjectID: projectID,
		Script:    script,
		Arguments: args,
		Signers:   signers,
		Logs:      result.Logs,
	}

	if result.Events != nil {
		events, _ := EventsFromFlow(result.Events)
		exe.Events = events
	}

	if result.Error != nil {
		exe.Errors = ProgramErrorFromFlow(result.Error)
	}

	return exe
}

type TransactionExecution struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Index     int
	Script    string
	Arguments []string       `gorm:"serializer:json"`
	Signers   []Address      `gorm:"serializer:json"`
	Errors    []ProgramError `gorm:"serializer:json"`
	Events    []Event        `gorm:"serializer:json"`
	Logs      []string       `gorm:"serializer:json"`
}

func (n *NewTransactionExecution) SignersToFlow() []flowsdk.Address {
	return convertSigners(n.Signers)
}

func (t *TransactionExecution) SignersToFlow() []flowsdk.Address {
	return convertSigners(t.Signers)
}

func convertSigners(signers []Address) []flowsdk.Address {
	sigs := make([]flowsdk.Address, len(signers))
	for i, sig := range signers {
		sigs[i] = sig.ToFlowAddress()
	}

	return sigs
}

func (u *UpdateTransactionTemplate) Validate() error {
	if u.Title == nil && u.Index == nil && u.Script == nil {
		return errors.Wrap(missingValuesError, "title, index, script")
	}
	return nil
}
