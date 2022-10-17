package model

import (
	"github.com/google/uuid"
	"github.com/onflow/flow-emulator/types"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
)

// ContractTemplate TODO: Why is this being re-declared in models_gen but script and trans are not?

type ContractTemplate = File

type ContractDeployment struct {
	File
	Address Address        `json:"address"`
	Errors  []ProgramError `gorm:"serializer:json"`
	Events  []Event        `gorm:"serializer:json"`
	Logs    []string       `gorm:"serializer:json"`
}

func ContractDeploymentFromFlow(
	// TODO: Do we need this function and what does it need to do?
	projectID uuid.UUID,
	address Address,
	result *types.TransactionResult,
	tx *flowsdk.Transaction,
) *ContractDeployment {
	script := ""
	// transaction could be nil in case where we get transaction result errors
	if tx != nil {
		/*
			for _, a := range tx.Arguments {
				args = append(args, string(a))
			}

			for _, a := range tx.Authorizers {
				signers = append(signers, NewAddressFromBytes(a.Bytes()))
			}
		*/

		script = string(tx.Script)
	}

	exe := &ContractDeployment{
		File: File{
			ID:        uuid.New(),
			ProjectID: projectID,
			Type:      TransactionFile,
			Script:    script,
		},
		Address: address,
		Errors:  nil,
		Events:  nil,
		Logs:    nil,
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

func (u *UpdateContractTemplate) Validate() error {
	if u.Title == nil && u.Index == nil && u.Script == nil {
		return errors.Wrap(missingValuesError, "title, index, script")
	}
	return nil
}
