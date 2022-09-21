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
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-emulator/types"
)

type ScriptTemplate struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Title     string
	Index     int
	Script    string
}

func ScriptExecutionFromFlow(result *types.ScriptResult, projectID uuid.UUID, script string, arguments []string) *ScriptExecution {
	exe := &ScriptExecution{
		ID:        uuid.New(),
		ProjectID: projectID,
		Script:    script,
		Arguments: arguments,
		Logs:      result.Logs,
	}

	if result.Error != nil {
		exe.Errors = ProgramErrorFromFlow(result.Error)
	} else {
		enc, _ := jsoncdc.Encode(result.Value)
		exe.Value = string(enc)
	}

	return exe
}

type ScriptExecution struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Index     int
	Script    string
	Arguments []string
	Value     string
	Errors    []ProgramError
	Logs      []string
}
