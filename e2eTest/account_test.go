package e2eTest

import (
	"encoding/json"
	client2 "github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAccountStorage(t *testing.T) {
	c := newClient()

	project := createProject(t, c)
	account := project.Accounts[0]

	var accResp GetAccountResponse

	err := c.Post(
		QueryGetAccount,
		&accResp,
		client2.Var("projectId", project.ID),
		client2.Var("accountId", account.Address),
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
		client2.Var("projectId", project.ID),
		client2.Var("script", script),
		client2.Var("signers", []string{account.Address}),
		client2.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	err = c.Post(
		QueryGetAccount,
		&accResp,
		client2.Var("projectId", project.ID),
		client2.Var("accountId", account.Address),
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
