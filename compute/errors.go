package compute

import (
	"errors"

	"github.com/onflow/flow-go/fvm"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/ast"
	runtimeErrors "github.com/onflow/cadence/runtime/errors"

	"github.com/dapperlabs/flow-playground-api/model"
)

func ExtractProgramErrors(err error) (result []model.ProgramError) {
	// set the default return value
	result = []model.ProgramError{
		{
			Message: err.Error(),
		},
	}

	// TODO: remove once fvm.ExecutionError implements Wrapper
	executionError, ok := err.(*fvm.ExecutionError)
	if !ok {
		return
	}

	var runtimeError runtime.Error
	if !errors.As(executionError.Err, &runtimeError) {
		return
	}

	parentError, ok := runtimeError.Unwrap().(runtimeErrors.ParentError)
	if !ok {
		return
	}

	childErrors := parentError.ChildErrors()

	result = make([]model.ProgramError, len(childErrors))

	for i, childError := range childErrors {

		programError := model.ProgramError{
			Message: childError.Error(),
		}

		if position, ok := childError.(ast.HasPosition); ok {
			programError.StartPosition = convertPosition(position.StartPosition())
			programError.EndPosition = convertPosition(position.EndPosition())
		}

		result[i] = programError
	}

	return result
}

func convertPosition(astPosition ast.Position) *model.ProgramPosition {
	programPosition := model.ProgramPosition(astPosition)
	return &programPosition
}
