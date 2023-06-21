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
	kit "github.com/onflow/flow-cli/flowkit"
	"os"
)

type InternalReaderWriter struct {
	data []byte
}

var _ kit.ReaderWriter = &InternalReaderWriter{}

func NewInternalReaderWriter() kit.ReaderWriter {
	return &InternalReaderWriter{}
}

func (rw *InternalReaderWriter) ReadFile(_ string) ([]byte, error) {
	return rw.data, nil
}

func (rw *InternalReaderWriter) WriteFile(_ string, data []byte, _ os.FileMode) error {
	rw.data = data
	return nil
}
