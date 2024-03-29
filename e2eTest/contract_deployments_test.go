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
	"fmt"
	"github.com/dapperlabs/flow-playground-api/blockchain"
	"github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var InitBlockHeight int = blockchain.GetInitialBlockHeightForTesting()

func TestContractDeployments(t *testing.T) {
	t.Run("Create deployment for non-existent project", func(t *testing.T) {
		c := newClient()

		badID := uuid.New().String()

		contractA := `
		pub contract HelloWorldA {
			pub var A: String
			pub init() { self.A = "HelloWorldA" }
		}`

		var resp CreateContractDeploymentResponse
		err := c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", badID),
			client.Var("script", contractA),
			client.Var("address", addr1),
		)

		assert.Error(t, err)
	})

	t.Run("Create deployment with initialization arguments", func(t *testing.T) {
		c := newClient()
		project := createProject(t, c)

		const contract = `
		pub contract HelloWorld {
			pub var A: Int
			pub init(a: Int) { self.A = a }
		}`

		args := []string{
			`{"type":"Int","value":"42"}`,
		}

		var resp CreateContractDeploymentResponse
		err := c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", contract),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		assert.Error(t, err)

		err = c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", contract),
			client.Var("address", addr1),
			client.Var("arguments", args),
			client.AddCookie(c.SessionCookie()),
		)
		assert.NoError(t, err)
		assert.Equal(t, args, resp.CreateContractDeployment.Arguments)
	})

}

func TestContractTitleParsing(t *testing.T) {
	c := newClient()

	project := createProject(t, c)
	contractA := `
		pub contract HelloWorld {
			pub init() {}
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
	require.Equal(t, "HelloWorld", respA.CreateContractDeployment.Title)
}

func TestImportSyntax(t *testing.T) {
	t.Run("new import syntax", func(t *testing.T) {
		c := newClient()
		project := createProject(t, c)

		const contractA = `
		pub contract HelloWorld {
			pub var A: Int
			pub init(a: Int) { self.A = a }
		}`

		args := []string{
			`{"type":"Int","value":"42"}`,
		}

		const contractB = `
		import "HelloWorld"
		pub contract Test {
			pub var B: Int
			pub init() { self.B = HelloWorld.A }
		}`

		var resp CreateContractDeploymentResponse
		err := c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", contractA),
			client.Var("address", addr1),
			client.Var("arguments", args),
			client.AddCookie(c.SessionCookie()),
		)
		assert.NoError(t, err)

		err = c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", contractB),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		assert.NoError(t, err)
	})

	t.Run("multiple deployments with same name", func(t *testing.T) {
		c := newClient()
		project := createProject(t, c)

		const contractA = `
		pub contract HelloWorld {
			pub var A: Int
			pub init() { self.A = 5 }
		}`

		const contractB = `
		import "HelloWorld"
		pub contract Test {
			pub var B: Int
			pub init() { self.B = HelloWorld.A }
		}`

		var resp CreateContractDeploymentResponse
		err := c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", contractA),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		assert.NoError(t, err)

		err = c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", contractA),
			client.Var("address", addr2),
			client.AddCookie(c.SessionCookie()),
		)
		assert.NoError(t, err)

		// TODO: What if we deploy a different contract with the same name?!?

		err = c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", contractA),
			client.Var("address", addr3),
			client.AddCookie(c.SessionCookie()),
		)
		assert.NoError(t, err)

		err = c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", contractB),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		assert.NoError(t, err)
	})
}

func TestContractRedeployment(t *testing.T) {
	t.Run("same contract name with different arguments", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		contractA := `
		pub contract HelloWorld {
			pub var A: Int
			pub init() { self.A = 5 }
			access(all) fun returnInt(): Int {
        		return self.A
    		}
			access(all) fun setVar(a: Int) {
				self.A = a
			}
		}`

		contractB := `
		pub contract HelloWorld {
			pub var B: String
			pub init() { self.B = "HelloWorldB" }
			access(all) fun returnString(): String {
        		return self.B
    		}
			access(all) fun setVar(b: String) {
				self.B = b
			}
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
		require.Equal(t, contractA, respA.CreateContractDeployment.Script)

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
		require.Equal(t, contractB, respB.CreateContractDeployment.Script)

		var accountResp GetAccountResponse
		err = c.Post(
			QueryGetAccount,
			&accountResp,
			client.Var("address", addr1),
			client.Var("projectId", project.ID),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Equal(t, []string{"HelloWorld"}, accountResp.Account.DeployedContracts)
	})

	t.Run("Contract redeployment with resource", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		PersonContract := `
		pub contract Person {
			pub fun makeFriends(): @Friendship {
				return <-create Friendship()
			}
		
			pub resource Friendship {
				pub fun yaay() {
					log("such a nice friend") // we can log to output, useful on emualtor for debugging
				}
			}
		}`

		var createContractResp CreateContractDeploymentResponse
		err := c.Post(
			MutationCreateContractDeployment,
			&createContractResp,
			client.Var("projectId", project.ID),
			client.Var("script", PersonContract),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		MakeFriendsTransaction := `
		import Person from 0x05
		
		transaction {
			let acc: AuthAccount
		
			prepare(signer: AuthAccount) {
				self.acc = signer    
			}
			
			execute {
				self.acc.save<@Person.Friendship>(<-Person.makeFriends(), to: StoragePath(identifier: "friendship")!)
			}
		}`

		var executeTransactionResp CreateTransactionExecutionResponse
		err = c.Post(
			MutationCreateTransactionExecution,
			&executeTransactionResp,
			client.Var("projectId", project.ID),
			client.Var("script", MakeFriendsTransaction),
			client.Var("signers", []string{addr1}),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		var accResp GetAccountResponse
		err = c.Post(
			QueryGetAccount,
			&accResp,
			client.Var("projectId", project.ID),
			client.Var("address", addr1),
		)
		require.NoError(t, err)
		require.Contains(t,
			accResp.Account.State,
			`A.0000000000000005.Person.Friendship`)

		PersonContractUpdate := `
		pub contract Person { 
		// empty
		}`

		err = c.Post(
			MutationCreateContractDeployment,
			&createContractResp,
			client.Var("projectId", project.ID),
			client.Var("script", PersonContractUpdate),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Equal(t, InitBlockHeight+1, createContractResp.CreateContractDeployment.BlockHeight)

		err = c.Post(
			QueryGetAccount,
			&accResp,
			client.Var("projectId", project.ID),
			client.Var("address", addr1),
		)
		require.NoError(t, err)
		require.NotContains(t, accResp.Account.State, "Person.Friendship")
	})

	t.Run("Contract redeployment block height rollback", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		PersonContract := `pub contract Person {}`

		var createContractResp CreateContractDeploymentResponse
		err := c.Post(
			MutationCreateContractDeployment,
			&createContractResp,
			client.Var("projectId", project.ID),
			client.Var("script", PersonContract),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, InitBlockHeight+1, createContractResp.CreateContractDeployment.BlockHeight)

		err = c.Post(
			MutationCreateContractDeployment,
			&createContractResp,
			client.Var("projectId", project.ID),
			client.Var("script", PersonContract),
			client.Var("address", addr2),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, InitBlockHeight+2, createContractResp.CreateContractDeployment.BlockHeight)

		err = c.Post(
			MutationCreateContractDeployment,
			&createContractResp,
			client.Var("projectId", project.ID),
			client.Var("script", PersonContract),
			client.Var("address", addr3),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, InitBlockHeight+3, createContractResp.CreateContractDeployment.BlockHeight)

		err = c.Post(
			MutationCreateContractDeployment,
			&createContractResp,
			client.Var("projectId", project.ID),
			client.Var("script", PersonContract),
			client.Var("address", addr4),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, InitBlockHeight+4, createContractResp.CreateContractDeployment.BlockHeight)

		err = c.Post(
			MutationCreateContractDeployment,
			&createContractResp,
			client.Var("projectId", project.ID),
			client.Var("script", PersonContract),
			client.Var("address", addr5),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, InitBlockHeight+5, createContractResp.CreateContractDeployment.BlockHeight)

		var projStorage GetProjectResponse
		err = c.Post(
			QueryGetProjectStorage,
			&projStorage,
			client.Var("projectId", project.ID),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		// Rollback block height
		err = c.Post(
			MutationCreateContractDeployment,
			&createContractResp,
			client.Var("projectId", project.ID),
			client.Var("script", PersonContract),
			client.Var("address", addr3),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, InitBlockHeight+3, createContractResp.CreateContractDeployment.BlockHeight)

		err = c.Post(
			QueryGetProjectStorage,
			&projStorage,
			client.Var("projectId", project.ID),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		// Rollback block height
		err = c.Post(
			MutationCreateContractDeployment,
			&createContractResp,
			client.Var("projectId", project.ID),
			client.Var("script", PersonContract),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, InitBlockHeight+1, createContractResp.CreateContractDeployment.BlockHeight)

		err = c.Post(
			QueryGetProjectStorage,
			&projStorage,
			client.Var("projectId", project.ID),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
	})

}

func TestContractInteraction(t *testing.T) {
	c := newClient()

	project := createProject(t, c)

	var respA CreateContractDeploymentResponse

	err := c.Post(
		MutationCreateContractDeployment,
		&respA,
		client.Var("projectId", project.ID),
		client.Var("script", counterContract),
		client.Var("address", addr1),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	addScript := generateAddTwoToCounterScript(addr1)

	var respB CreateTransactionExecutionResponse

	err = c.Post(
		MutationCreateTransactionExecution,
		&respB,
		client.Var("projectId", project.ID),
		client.Var("script", addScript),
		client.Var("signers", []string{addr2}),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)
	assert.Empty(t, respB.CreateTransactionExecution.Errors)
}

func TestContractImport(t *testing.T) {
	c := newClient()

	project := createProject(t, c)

	contractA := `
	pub contract HelloWorldA {
		pub var A: String
		pub init() { self.A = "HelloWorldA" }
	}`

	contractB := `
	import HelloWorldA from 0x05
	pub contract HelloWorldB {
		pub init() {
			log(HelloWorldA.A)
		}
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
		client.Var("address", addr2),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)
	require.Empty(t, respB.CreateContractDeployment.Errors)
	require.Contains(t, respB.CreateContractDeployment.Logs[0], "HelloWorldA")
}

const counterContract = `
  pub contract Counting {

      pub event CountIncremented(count: Int)

      pub resource Counter {
          pub var count: Int

          init() {
              self.count = 0
          }

          pub fun add(_ count: Int) {
              self.count = self.count + count
              emit CountIncremented(count: self.count)
          }
      }

      pub fun createCounter(): @Counter {
          return <-create Counter()
      }
  }
`

// generateAddTwoToCounterScript generates a script that increments a counter.
// If no counter exists, it is created.
func generateAddTwoToCounterScript(counterAddress string) string {
	return fmt.Sprintf(
		`
            import 0x%s

            transaction {

                prepare(signer: AuthAccount) {
                    if signer.borrow<&Counting.Counter>(from: /storage/counter) == nil {
                        signer.save(<-Counting.createCounter(), to: /storage/counter)
                    }

                    signer.borrow<&Counting.Counter>(from: /storage/counter)!.add(2)
                }
            }
        `,
		counterAddress,
	)
}
