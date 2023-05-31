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
	"fmt"
	"github.com/google/uuid"
	"github.com/onflow/cadence"
	"github.com/pkg/errors"
)

type ScriptTemplate = File

type ScriptExecution struct {
	File
	Arguments []string `gorm:"serializer:json"`
	Value     string
	Errors    []ProgramError `gorm:"serializer:json"`
	Logs      []string       `gorm:"serializer:json"`
}

func ScriptExecutionFromFlow(result cadence.Value, projectID uuid.UUID, script string, arguments []string) *ScriptExecution {
	// TODO: Get execution result from a cadence.Value and check for errors
	fmt.Println("Script Execution Result: ", result.String())

	exe := &ScriptExecution{
		File: File{
			ID:        uuid.New(),
			ProjectID: projectID,
			Type:      ScriptFile,
			Script:    script,
		},
		Arguments: arguments,
		Value:     result.String(),
		Errors:    nil,
		Logs:      nil, // TODO: Add result
	}

	/*
		// TODO: Check result for errors
		if result.Error != nil {
			exe.Errors = ProgramErrorFromFlow(result.Error)
		} else {
			enc, _ := jsoncdc.Encode(result.Value)
			exe.Value = string(enc)
		}
	*/

	return exe
}

func (u *UpdateScriptTemplate) Validate() error {
	if u.Title == nil && u.Script == nil && u.Index == nil {
		return errors.Wrap(missingValuesError, "title, script, index")
	}
	return nil
}
