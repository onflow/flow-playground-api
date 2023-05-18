package blockchain

import (
	"context"
	"github.com/onflow/flow-cli/flowkit/accounts"
	flow "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test_NewEmulator tests creating a large number of new accounts and validates corresponding storage addresses
func Test_NewFlowkit(t *testing.T) {
	fk, err := newFlowkit()
	assert.NoError(t, err)

	var accountList []*flow.Account

	const testAccounts int = 10

	for i := 0; i < testAccounts; i++ {
		state, err := fk.blockchain.State()
		assert.NoError(t, err)

		service, err := state.EmulatorServiceAccount()
		assert.NoError(t, err)

		serviceKey, err := service.Key.PrivateKey()
		assert.NoError(t, err)

		account, _, err := fk.blockchain.CreateAccount(
			context.Background(), service,
			[]accounts.PublicKey{{
				Public:   (*serviceKey).PublicKey(),
				Weight:   flow.AccountKeyWeightThreshold,
				SigAlgo:  crypto.ECDSA_P256,
				HashAlgo: crypto.SHA3_256,
			}},
		)
		assert.NoError(t, err)
		accountList = append(accountList, account)
	}

	for i := 0; i < testAccounts; i++ {
		// TODO: Verify account storage
		account, _, err := fk.getAccount(accountList[i].Address)
		//_, accountStorage, err := emu.getAccount(accountList[i].Address)
		assert.NoError(t, err)
		assert.Equal(t, account.Address, accountList[i].Address)
		//assert.Equal(t, accountStorage.Account.Address.String(), accountList[i].Address.String())
		//assert.Equal(t, accountStorage.Account.Address.Hex(), accountList[i].Address.Hex())
		//assert.Equal(t, accountStorage.Account.Address.Bytes(), accountList[i].Address.Bytes())
	}
}
