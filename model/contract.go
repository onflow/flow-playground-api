package model

import (
	"github.com/google/uuid"
	"github.com/onflow/flow-emulator/types"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
)

// ContractTemplate TODO: Why is this being re-declared in models_gen but script and trans are not?
type ContractTemplate = File

func ContractDeploymentFromFlow(
	// TODO: Do we need this function and what does it need to do?
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
		File: File{
			ID:        uuid.New(),
			ProjectID: projectID,
			Type:      TransactionFile,
			Script:    script,
		},
		Arguments: nil,
		Signers:   nil,
		Errors:    nil,
		Events:    nil,
		Logs:      nil,
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
