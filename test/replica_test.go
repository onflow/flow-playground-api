package test

/* TODO: FIX
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
			assert.Contains(t, resp.CreateTransactionExecution.Logs, "\"Hello, World!\"")
			assert.Equal(t, script, resp.CreateTransactionExecution.Script)
		}
	})

	t.Run("Re-deploy contracts on multiple replicas to initial accounts", func(t *testing.T) {
		var contract = "pub contract Foo {}"

		for i := 0; i < 10; i++ {
			accountIdx := i % project.NumberOfAccounts

			var deployResp CreateContractDeploymentResponse
			err := loadBalancer().Post(
				MutationCreateContractDeployment,
				&deployResp,
				client.Var("projectId", project.ID),
				client.Var("address", model.NewAddressFromString("0x0"+strconv.Itoa(accountIdx))),
				client.Var("code", contract),
				client.AddCookie(cookie),
			)
			require.NoError(t, err)

			/*
				var updateResp UpdateAccountResponse
				err := loadBalancer().Post(
					MutationUpdateAccountDeployedCode,
					&updateResp,
					client.Var("projectId", project.ID),
					client.Var("accountId", account.ID),
					client.Var("code", contract),
					client.AddCookie(cookie),
				)
				require.NoError(t, err)
				assert.Equal(t, contract, updateResp.UpdateAccount.DeployedCode)
				assert.Equal(t, updateResp.UpdateAccount.DeployedCode, contract)

		}
	})
}
*/
