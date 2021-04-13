/*
 * Flow Playground
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package compute

import (
	"errors"

	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/ast"
	runtimeErrors "github.com/onflow/cadence/runtime/errors"
	errors2 "github.com/onflow/flow-go/fvm/errors"

	"github.com/dapperlabs/flow-playground-api/model"
)

func ExtractProgramErrors(err error) (result []model.ProgramError) {
	// set the default return value
	result = []model.ProgramError{
		convertProgramError(err),
	}

	// TODO: remove once fvm.ExecutionError implements Wrapper
	executionError, ok := err.(*errors2.ExecutionError)
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
