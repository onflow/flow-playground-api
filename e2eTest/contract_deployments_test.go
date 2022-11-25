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
	"github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

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

}

func TestContractRedeployment(t *testing.T) {
	t.Run("same contract name different arguments", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		contractA := `
		pub contract HelloWorld {
			pub var A: Int
			pub init() { self.A = 5 }
			access(all) fun returnInt(): Int {
        		return self.A
    		}
		}`

		contractB := `
		pub contract HelloWorld {
			pub var B: String
			pub init() { self.B = "HelloWorldB" }
			access(all) fun returnString(): String {
        		return self.B
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

		// TODO: Why isn't the actual contract script stored in the deployment?
		// TODO: How can I verify that the right contract was deployed?
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

		require.Equal(t, "HelloWorld", accountResp.Account.DeployedContracts)
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
	import HelloWorldA from 0x01
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
