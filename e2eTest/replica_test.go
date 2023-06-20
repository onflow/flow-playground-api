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
	"github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestReplicas(t *testing.T) {
	// Each replica is a different client calling the API, but also an instance of the resolver
	const numReplicas = 5

	// Create replicas
	var replicas []*Client
	for i := 0; i < numReplicas; i++ {
		replicas = append(replicas, newClient())
	}

	replicaIdx := 0 // current replica
	// loadBalancer cycles through replicas
	var loadBalancer = func() *Client {
		replicaIdx = (replicaIdx + 1) % len(replicas)
		return replicas[replicaIdx]
	}

	// Create project for all replica tests
	c := loadBalancer()
	project := createProject(t, c)
	cookie := c.SessionCookie() // Use one session cookie for everything currently

	t.Run("Execute transactions on multiple replicas", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			const script = "transaction { execute { log(\"Hello, World!\") } }"

			var resp CreateTransactionExecutionResponse
			err := loadBalancer().Post(
				MutationCreateTransactionExecution,
				&resp,
				client.Var("projectId", project.ID),
				client.Var("script", script),
				client.AddCookie(cookie),
			)

			require.NoError(t, err)
			assert.Empty(t, resp.CreateTransactionExecution.Errors)
			//assert.Contains(t, resp.CreateTransactionExecution.Logs, "\"Hello, World!\"")
			assert.Equal(t, script, resp.CreateTransactionExecution.Script)
		}
	})

	t.Run("Re-deploy contracts on multiple replicas to initial accounts", func(t *testing.T) {
		var contract = "pub contract Foo {}"

		for i := 0; i < 10; i++ {
			accountIdx := (i % project.NumberOfAccounts) + 1

			var deployResp CreateContractDeploymentResponse
			err := loadBalancer().Post(
				MutationCreateContractDeployment,
				&deployResp,
				client.Var("projectId", project.ID),
				client.Var("address", "000000000000000"+strconv.Itoa(accountIdx)),
				client.Var("script", contract),
				client.AddCookie(cookie),
			)
			require.NoError(t, err)
		}
	})
}
