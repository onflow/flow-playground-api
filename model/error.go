package model

import (
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/ast"
	runtimeErrors "github.com/onflow/cadence/runtime/errors"
	"github.com/pkg/errors"
)

func ProgramErrorFromFlow(err error) (result []ProgramError) {
	// set the default return value
	result = []ProgramError{
		convertProgramError(err),
	}

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

func convertProgramErrors(errors []error) []ProgramError {
	result := make([]ProgramError, len(errors))

	for i, err := range errors {
		result[i] = convertProgramError(err)
	}

	return result
}

func convertProgramError(err error) ProgramError {
	programError := ProgramError{
		Message: err.Error(),
	}

	var unexpectedErr runtimeErrors.UnexpectedError
	if errors.As(err, &unexpectedErr) {
		programError.Message = unexpectedErr.Err.Error() // remove error stack
	}

	if position, ok := err.(ast.HasPosition); ok {
		programError.StartPosition = convertPosition(position.StartPosition())
		programError.EndPosition = convertPosition(position.EndPosition(nil))
	}

	return programError
}

func convertPosition(astPosition ast.Position) *ProgramPosition {
	programPosition := ProgramPosition(astPosition)
	return &programPosition
}
