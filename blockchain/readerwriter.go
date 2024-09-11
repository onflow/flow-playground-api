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
	"os"

	kit "github.com/onflow/flowkit/v2"
)

type InternalReaderWriter struct {
	data []byte
}

func (rw *InternalReaderWriter) Stat(path string) (os.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

var _ kit.ReaderWriter = &InternalReaderWriter{}

func NewInternalReaderWriter() *InternalReaderWriter {
	return &InternalReaderWriter{}
}

func (rw *InternalReaderWriter) ReadFile(path string) ([]byte, error) {
	if path == "flow.json" {
		return []byte(`
{
	"contracts": {
	},
	"networks": {
		"emulator": "127.0.0.1:3569",
		"mainnet": "access.mainnet.nodes.onflow.org:9000",
		"testing": "127.0.0.1:3569",
		"testnet": "access.devnet.nodes.onflow.org:9000"
	},
	"accounts": {
		"emulator-account": {
			"address": "0000000000000001",
			"key": "0x0d866eb285a9bdb29730a1ca37bd7201fb5bd1a922632b1d5a784b6bc3c216b9"
		}
	}
}`), nil
	}
	return rw.data, nil
}

func (rw *InternalReaderWriter) WriteFile(_ string, data []byte, _ os.FileMode) error {
	rw.data = data
	return nil
}

func (rw *InternalReaderWriter) MkdirAll(_ string, _ os.FileMode) error {
	return nil
}
