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
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-go-sdk"
)

// Embed all contracts in this folder
//
//go:embed *.cdc
var contracts embed.FS

// Core defines core contract to be embedded, along with their locations in the emulator
var Core = []config.Contract{
	{
		Name: "NonFungibleToken",
		Aliases: config.Aliases{
			config.Alias{
				Network: "emulator",
				Address: flow.HexToAddress("0xf8d6e0586b0a20c7"),
			},
		},
	},
	{
		Name: "FungibleToken",
		Aliases: config.Aliases{
			config.Alias{
				Network: "emulator",
				Address: flow.HexToAddress("0xee82856bf20e2aa6"),
			},
		},
	},
	// TODO: Need to support new import schema in order to resolve core contracts
	/*
		{
			Name: "FlowToken",
			Aliases: config.Aliases{
				config.Alias{
					Network: "emulator",
					Address: flow.HexToAddress("0x0ae53cb6e3f42a79"),
				},
			},
		},
		{
			Name: "MetadataViews",
			Aliases: config.Aliases{
				config.Alias{
					Network: "emulator",
					Address: flow.HexToAddress("0xf8d6e0586b0a20c7"),
				},
			},
		},
	*/
}

func Included() []string {
	var included []string
	for _, contract := range Core {
		included = append(included, contract.Name)
	}
	return included
}

func Read(name string) ([]byte, error) {
	return contracts.ReadFile(fmt.Sprintf("%s.cdc", name))
}
