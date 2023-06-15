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
	"context"
	"fmt"
	"github.com/onflow/flow-cli/flowkit/accounts"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test_NewEmulator tests creating a large number of new accounts and validates corresponding storage addresses
func Test_NewFlowkit(t *testing.T) {
	fk, err := newFlowkit()
	assert.NoError(t, err)

	var accountList []*flow.Account

	const testAccounts int = 10

	for i := 0; i < testAccounts; i++ {
		state, err := fk.blockchain.State()
		assert.NoError(t, err)

		service, err := state.EmulatorServiceAccount()
		assert.NoError(t, err)

		serviceAccount := &accounts.Account{
			Name:    "Service Account",
			Address: flow.HexToAddress("0x01"),
			Key:     service.Key,
		}

		account, _, err := fk.blockchain.CreateAccount(
			context.Background(),
			serviceAccount,
			[]accounts.PublicKey{},
		)
		assert.NoError(t, err)
		accountList = append(accountList, account)
	}

	for i := 0; i < testAccounts; i++ {
		// TODO: Verify account storage
		account, _, err := fk.getAccount(accountList[i].Address)
		//_, accountStorage, err := emu.getAccount(accountList[i].Address)
		assert.NoError(t, err)
		assert.Equal(t, account.Address, accountList[i].Address)
		//assert.Equal(t, accountStorage.Account.Address.String(), accountList[i].Address.String())
		//assert.Equal(t, accountStorage.Account.Address.Hex(), accountList[i].Address.Hex())
		//assert.Equal(t, accountStorage.Account.Address.Bytes(), accountList[i].Address.Bytes())
	}
}

func Test_FlowJsonExport(t *testing.T) {
	fk, err := newFlowkit()
	assert.NoError(t, err)

	blockHeight, err := fk.getLatestBlockHeight()
	assert.NoError(t, err)
	assert.Equal(t, fk.initBlockHeight(), blockHeight)

	flowJson, err := fk.getFlowJson()
	assert.NoError(t, err)

	const CoreContracts = `"contracts": {
		"FlowToken": {
			"source": "",
			"aliases": {
				"emulator": "0ae53cb6e3f42a79"
			}
		},
		"FungibleToken": {
			"source": "",
			"aliases": {
				"emulator": "ee82856bf20e2aa6"
			}
		},
		"MetadataViews": {
			"source": "",
			"aliases": {
				"emulator": "f8d6e0586b0a20c7"
			}
		},
		"NonFungibleToken": {
			"source": "",
			"aliases": {
				"emulator": "f8d6e0586b0a20c7"
			}
		}
	}`

	const Networks = `"networks": {
		"emulator": "127.0.0.1:3569",
		"mainnet": "access.mainnet.nodes.onflow.org:9000",
		"sandboxnet": "access.sandboxnet.nodes.onflow.org:9000",
		"testnet": "access.devnet.nodes.onflow.org:9000"
	}`

	fmt.Println(flowJson)
	assert.Contains(t, flowJson, CoreContracts)
	assert.Contains(t, flowJson, Networks)

	// Accounts
	assert.Contains(t, flowJson, "Account 0x01")
	assert.Contains(t, flowJson, "Account 0x02")
	assert.Contains(t, flowJson, "Account 0x03")
	assert.Contains(t, flowJson, "Account 0x04")
	assert.Contains(t, flowJson, "Account 0x05")
	assert.Contains(t, flowJson, "Service Account")
	assert.Contains(t, flowJson, "emulator-account")
}