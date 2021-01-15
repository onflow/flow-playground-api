package compute

import (
	"errors"

	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/ast"
	runtimeErrors "github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/flow-go/fvm"

	"github.com/dapperlabs/flow-playground-api/model"
)

func ExtractProgramErrors(err error) (result []model.ProgramError) {
	// set the default return value
	result = []model.ProgramError{
		convertProgramError(err),
	}

	// TODO: remove once fvm.ExecutionError implements Wrapper
	executionError, ok := err.(*fvm.ExecutionError)
	if !ok {
		return
	}
	err = executionError.Err

	var parsingCheckingError *runtime.ParsingCheckingError
	if errors.As(err, &parsingCheckingError) {
		err = parsingCheckingError.Err
	}

	var parentError runtimeErrors.ParentError
	if !errors.As(err, &parentError) {
		return
	}

	return convertProgramErrors(parentError.ChildErrors())
}

func convertProgramErrors(errors []error) []model.ProgramError {
	result := make([]model.ProgramError, len(errors))

	for i, err := range errors {
		result[i] = convertProgramError(err)
	}

	return result
}

func convertProgramError(err error) model.ProgramError {
	programError := model.ProgramError{
		Message: err.Error(),
	}

	if position, ok := err.(ast.HasPosition); ok {
		programError.StartPosition = convertPosition(position.StartPosition())
		programError.EndPosition = convertPosition(position.EndPosition())
	}

	return programError
}

func convertPosition(astPosition ast.Position) *model.ProgramPosition {
	programPosition := model.ProgramPosition(astPosition)
	return &programPosition
}
