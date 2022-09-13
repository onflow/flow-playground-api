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
	"testing"

	"github.com/dapperlabs/flow-playground-api/model"

	"github.com/stretchr/testify/assert"
)

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

func Test_BasicAccounts(t *testing.T) {
	emu, err := newEmulator()
	assert.NoError(t, err)

	account1, _, _, err := emu.createAccount()
	assert.NoError(t, err)

	fmt.Println("Account1 address:", account1.Address)

	account2, _, err := emu.getAccount(account1.Address)

	fmt.Println("Account2 address:", account2.Address)

	assert.Equal(t, account1, account2)
}
