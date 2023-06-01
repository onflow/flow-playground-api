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

package contracts

import (
	"embed"
	"fmt"
)

// Embed all contracts in this folder
//
//go:embed *.cdc
var contracts embed.FS

var include = []string{
	"FungibleToken",
	"NonFungibleToken",
	"FlowToken",
	"MetadataViews",
	// Add more standard contracts here
	// Note: Adding more contracts will change the initial block height
}

func Included() []string {
	return include
}

func Read(name string) ([]byte, error) {
	return contracts.ReadFile(fmt.Sprintf("%s.cdc", name))
}
