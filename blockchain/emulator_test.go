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
	"fmt"
	flowsdk "github.com/onflow/flow-go-sdk"
	"strconv"
	"testing"

	"github.com/dapperlabs/flow-playground-api/model"

	"github.com/stretchr/testify/assert"
)

// Test_TranslateAddress tests that address translation works as expected
func Test_TranslateAddress(t *testing.T) {
	assert.Equal(t, NumberOfServiceAccounts, model.NumberOfServiceAccounts) // avoid circular deps

	inputs := [][][]byte{{
		[]byte(`
			import Foo from 0x02
			pub contract Bar {}
		`), []byte(`
			import Foo from 0x0000000000000006
			pub contract Bar {}
		`),
	}, {
		[]byte(`
			import Zoo from 0x0000000000000001
			pub fun main() {}
		`), []byte(`
			import Zoo from 0x0000000000000005
			pub fun main() {}
		`),
	}, {
		[]byte(`
			import Crypto
			pub fun main() {}
		`), []byte(`
			import Crypto
			pub fun main() {}
		`),
	}}

	for _, in := range inputs {
		out := translateAddresses(in[0])
		assert.Equal(t, string(in[1]), string(out))
	}
}

// Test_NewEmulator tests creating a large number of new accounts and validates corresponding storage addresses
func Test_NewEmulator(t *testing.T) {
	emu, err := newEmulator()
	assert.NoError(t, err)

	var accountList []*flowsdk.Account

	const testAccounts int = 1000

	fmt.Println("Creating", testAccounts, "new accounts...")
	for i := 0; i < testAccounts; i++ {
		account, _, _, err := emu.createAccount()
		assert.NoError(t, err)
		accountList = append(accountList, account)
	}

	fmt.Println("Validating account storage...")
	for i := 0; i < testAccounts; i++ {
		_, accountStorage, err := emu.getAccount(accountList[i].Address)
		assert.NoError(t, err)
		assert.Equal(t, accountStorage.Account.Address.String(), accountList[i].Address.String())
		assert.Equal(t, accountStorage.Account.Address.Hex(), accountList[i].Address.Hex())
		assert.Equal(t, accountStorage.Account.Address.Bytes(), accountList[i].Address.Bytes())
	}
}

// Test_DeployEmptyContract tests deployment of an empty contract
func Test_DeployEmptyContract(t *testing.T) {
	emu, err := newEmulator()
	assert.NoError(t, err)
	account, _, _, err := emu.createAccount()
	_, _, err = emu.deployContract(account.Address, "")
	assert.Error(t, err)
}

// Test_DeployBasicContracts tests deployment of a large number of basic contracts to a single account
func Test_DeployBasicContracts(t *testing.T) {
	emu, err := newEmulator()
	assert.NoError(t, err)
	account, _, _, err := emu.createAccount()

	const numContracts int = 1000

	const baseName string = "Foo"
	var deployedContracts []string

	for i := 0; i < numContracts; i++ {
		name := baseName + strconv.Itoa(i)
		contract := "pub contract " + name + "{}"
		deployedContracts = append(deployedContracts, name)

		_, tx, err := emu.deployContract(account.Address, contract)
		assert.NoError(t, err)
		assert.Equal(t, tx.Authorizers[0], account.Address)
	}

	account, _, err = emu.getAccount(account.Address)

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
}

// Test_ParseContractName tests contract name parsing returns the correct name
func Test_ParseContractName(t *testing.T) {
	contract := "pub contract foo {}"
	name, err := parseContractName(contract)
	assert.NoError(t, err)
	assert.Equal(t, name, "foo")

	longName := "foo"
	for i := 0; i < 100000; i++ {
		longName += "long"
	}

	contract = "pub contract " + longName + " {}"
	name, err = parseContractName(contract)
	assert.NoError(t, err)
	assert.Equal(t, name, longName)

	contract = "pub contract foo_bar {}"
	name, err = parseContractName(contract)
	assert.NoError(t, err)
	assert.Equal(t, name, "foo_bar")

	// Double name
	contract = "pub contract foo bar {}"
	name, err = parseContractName(contract)
	assert.Error(t, err)

	// No name
	contract = "pub contract {}"
	name, err = parseContractName(contract)
	assert.Error(t, err)

	// Decimal name
	contract = "pub contract 123foo {}"
	name, err = parseContractName(contract)
	assert.Error(t, err)

	// Invalid character
	contract = "pub contract foo! {}"
	name, err = parseContractName(contract)
	assert.Error(t, err)
}
