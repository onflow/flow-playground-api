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

package blockchain

import (
	"io"
	"strings"
)

type Logs []string

// Interceptor is used to intercept Cadence runtime logs from the emulator
type Interceptor struct {
	logs Logs
}

var _ io.Writer = &Interceptor{}

func NewInterceptor() *Interceptor {
	return &Interceptor{}
}

func (logger *Interceptor) Write(p []byte) (n int, err error) {
	logger.logs = append(logger.logs, string(p))
	return len(p), nil
}

func (logger *Interceptor) ClearLogs() {
	logger.logs = Logs{}
}

func (logger *Interceptor) GetCadenceLogs() Logs {
	var filteredLogs Logs
	for _, log := range logger.logs {
		if strings.Contains(log, `"message":"Cadence log:`) {
			filteredLogs = append(filteredLogs, strings.TrimSuffix(log, "\n"))
		}
	}
	return filteredLogs
}
