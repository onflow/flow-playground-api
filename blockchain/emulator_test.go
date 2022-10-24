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
	flowsdk "github.com/onflow/flow-go-sdk"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_NewEmulator tests creating a large number of new accounts and validates corresponding storage addresses
func Test_NewEmulator(t *testing.T) {
	emu, err := newEmulator()
	assert.NoError(t, err)

	var accountList []*flowsdk.Account

	const testAccounts int = 10

	for i := 0; i < testAccounts; i++ {
		account, _, _, err := emu.createAccount()
		assert.NoError(t, err)
		accountList = append(accountList, account)
	}

	for i := 0; i < testAccounts; i++ {
		_, accountStorage, err := emu.getAccount(accountList[i].Address)
		assert.NoError(t, err)
		assert.Equal(t, accountStorage.Account.Address.String(), accountList[i].Address.String())
		assert.Equal(t, accountStorage.Account.Address.Hex(), accountList[i].Address.Hex())
		assert.Equal(t, accountStorage.Account.Address.Bytes(), accountList[i].Address.Bytes())
	}
}

// Test_DeployContracts tests deployment of different contracts
func Test_DeployContracts(t *testing.T) {
	// Test deploying an empty contract
	t.Run("Empty Contract", func(t *testing.T) {
		emu, err := newEmulator()
		assert.NoError(t, err)
		account, _, _, err := emu.createAccount()
		assert.NoError(t, err)

		// TODO: WHY IS THERE NO ERROR WHEN DEPLOYING AN EMPTY SCRIPT?!
		_, _, err = emu.deployContract(account.Address, "", "")
		assert.Error(t, err)
	})

	// Test deploying many contracts to a single account
	t.Run("Basic Contracts", func(t *testing.T) {
		emu, err := newEmulator()
		assert.NoError(t, err)
		account, _, _, err := emu.createAccount()
		assert.NoError(t, err)

		const numContracts int = 10

		const baseName string = "Foo"
		var deployedContracts []string

		for i := 0; i < numContracts; i++ {
			name := baseName + strconv.Itoa(i)
			contract := "pub contract " + name + "{}"
			deployedContracts = append(deployedContracts, name)

			_, tx, err := emu.deployContract(account.Address, contract, name)
			assert.NoError(t, err)
			assert.Equal(t, tx.Authorizers[0], account.Address)
		}

		account, _, err = emu.getAccount(account.Address)
		assert.NoError(t, err)

		keys := make([]string, 0, len(account.Contracts))
		for k := range account.Contracts {
			keys = append(keys, k)
		}

		// Verify that every deployed contract is included
		for _, deployed := range deployedContracts {
			contains := false
			for _, contract := range keys {
				if contract == deployed {
					contains = true
					break
				}
			}
			assert.Equal(t, contains, true)
		}
	})
}

// Test_ParseContractName tests contract name parsing returns the correct name
func Test_ParseContractName(t *testing.T) {
	type testCase struct {
		inputContract string
		expected      string
		errExpected   bool
	}

	longName := "foo" + strings.Repeat("long", 100000)

	tests := []testCase{
		{"pub contract foo {}", "foo", false},
		{"pub contract " + longName + " {}", longName, false},
		{"pub contract foo_bar {}", "foo_bar", false},
		{"pub contract foo bar {}", "", true},
		{"pub contract {}", "", true},
		{"pub contract 123foo {}", "", true},
		{"pub contract foo! {}", "", true},
	}

	for _, tc := range tests {
		name, err := parseContractName(tc.inputContract)
		if tc.errExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, name, tc.expected)
	}
}
