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

package e2eTest

import (
	"encoding/json"
	"github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAccountDeployedContracts(t *testing.T) {
	c := newClient()

	project := createProject(t, c)
	account := project.Accounts[0]

	contractA := `
	pub contract HelloWorldA {
		pub var A: String
		pub init() { self.A = "HelloWorldA" }
	}`

	contractB := `
	pub contract HelloWorldB {
		pub var B: String
		pub init() { self.B = "HelloWorldB" }
	}`

	var respA CreateContractDeploymentResponse
	err := c.Post(
		MutationCreateContractDeployment,
		&respA,
		client.Var("projectId", project.ID),
		client.Var("script", contractA),
		client.Var("address", addr1),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	var respB CreateContractDeploymentResponse
	err = c.Post(
		MutationCreateContractDeployment,
		&respB,
		client.Var("projectId", project.ID),
		client.Var("script", contractB),
		client.Var("address", addr1),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	var accResp GetAccountResponse
	err = c.Post(
		QueryGetAccount,
		&accResp,
		client.Var("projectId", project.ID),
		client.Var("address", account.Address),
	)
	require.NoError(t, err)

	require.EqualValues(t, []string{"HelloWorldA", "HelloWorldB"}, accResp.Account.DeployedContracts)
}

func TestAccountStorage(t *testing.T) {
	c := newClient()

	project := createProject(t, c)
	account := project.Accounts[0]

	var accResp GetAccountResponse

	err := c.Post(
		QueryGetAccount,
		&accResp,
		client.Var("projectId", project.ID),
		client.Var("address", account.Address),
	)
	require.NoError(t, err)

	assert.Equal(t, account.Address, accResp.Account.Address)
	assert.Equal(t, `{}`, accResp.Account.State)

	var resp CreateTransactionExecutionResponse

	const script = `
		transaction {
		  prepare(signer: AuthAccount) {
			  	signer.save("storage value", to: /storage/storageTest)
 				signer.link<&String>(/public/publicTest, target: /storage/storageTest)
				signer.link<&String>(/private/privateTest, target: /storage/storageTest)
		  }
   		}`

	err = c.Post(
		MutationCreateTransactionExecution,
		&resp,
		client.Var("projectId", project.ID),
		client.Var("script", script),
		client.Var("signers", []string{account.Address}),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	err = c.Post(
		QueryGetAccount,
		&accResp,
		client.Var("projectId", project.ID),
		client.Var("address", account.Address),
	)
	require.NoError(t, err)

	assert.Equal(t, account.Address, accResp.Account.Address)
	assert.NotEmpty(t, accResp.Account.State)

	type accountStorage struct {
		Private map[string]any
		Public  map[string]any
		Storage map[string]any
	}

	var accStorage accountStorage
	err = json.Unmarshal([]byte(accResp.Account.State), &accStorage)
	require.NoError(t, err)

	assert.Equal(t, "storage value", accStorage.Storage["storageTest"])
	assert.NotEmpty(t, accStorage.Private["privateTest"])
	assert.NotEmpty(t, accStorage.Public["publicTest"])

	assert.NotContains(t, accStorage.Public, "flowTokenBalance")
	assert.NotContains(t, accStorage.Public, "flowTokenReceiver")
	assert.NotContains(t, accStorage.Storage, "flowTokenVault")
}
