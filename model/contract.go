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
	Address Address        `gorm:"serializer:json"`
	Errors  []ProgramError `gorm:"serializer:json"`
	Events  []Event        `gorm:"serializer:json"`
	Logs    []string       `gorm:"serializer:json"`
}

func ContractDeploymentFromFlow(
	projectID uuid.UUID,
	result *types.TransactionResult,
	tx *flowsdk.Transaction,
) *ContractDeployment {
	script := ""
	// transaction could be nil in case where we get transaction result errors
	if tx != nil {
		script = string(tx.Script)
	}

	exe := &ContractDeployment{
		File: File{
			ID:        uuid.New(),
			ProjectID: projectID,
			Type:      ContractFile,
			Script:    script,
		},
		Errors: nil,
		Events: nil,
		Logs:   nil,
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
