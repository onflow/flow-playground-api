package test

// TODO: Test contract deployments similar to transaction executions

/*
func TestContractInteraction(t *testing.T) {
	c := newClient()

	project := createProject(t, c)

	accountA := project.Accounts[0]
	accountB := project.Accounts[1]

	var respA UpdateAccountResponse

	err := c.Post(
		MutationUpdateAccountDeployedCode,
		&respA,
		client.Var("projectId", project.ID),
		client.Var("accountId", accountA.ID),
		client.Var("code", counterContract),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	assert.Equal(t, counterContract, respA.UpdateAccount.DeployedCode)

	addScript := generateAddTwoToCounterScript(accountA.Address)

	var respB CreateTransactionExecutionResponse

	err = c.Post(
		MutationCreateTransactionExecution,
		&respB,
		client.Var("projectId", project.ID),
		client.Var("script", addScript),
		client.Var("signers", []string{accountB.Address}),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	assert.Empty(t, respB.CreateTransactionExecution.Errors)
}

func TestContractImport(t *testing.T) {
	c := newClient()

	project := createProject(t, c)

	accountA := project.Accounts[0]
	accountB := project.Accounts[1]

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

	var respA UpdateAccountResponse

	err := c.Post(
		MutationUpdateAccountDeployedCode,
		&respA,
		client.Var("projectId", project.ID),
		client.Var("accountId", accountA.ID),
		client.Var("code", contractA),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)
	assert.Equal(t, contractA, respA.UpdateAccount.DeployedCode)

	var respB UpdateAccountResponse

	err = c.Post(
		MutationUpdateAccountDeployedCode,
		&respB,
		client.Var("projectId", project.ID),
		client.Var("accountId", accountB.ID),
		client.Var("code", contractB),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)
}
*/
