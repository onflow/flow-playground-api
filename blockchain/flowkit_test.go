package blockchain

import (
	emu "github.com/onflow/flow-emulator"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test_NewEmulator tests creating a large number of new accounts and validates corresponding storage addresses
func Test_NewFlowkit(t *testing.T) {
	fk, err := newFlowkit()
	assert.NoError(t, err)

	var accountList []*flowsdk.Account

	const testAccounts int = 10

	for i := 0; i < testAccounts; i++ {
		fk.blockchain.CreateAccount() //TODO?
		account, _, _, err := emu.createAccount()
		assert.NoError(t, err)
		accountList = append(accountList, account)
	}

	for i := 0; i < testAccounts; i++ {
		_, accountStorage, err := emu.getAccount(accountList[i].Address)
		assert.NoError(t, err)
		assert.Equal(t, accountStorage.Account.Address.String(), accountList[i].Address.String())
		assert.Equal(t, accountStorage.Account.Address.Hex(), accountList[i].Address.Hex())
		assert.Equal(t, accountStorage.Account.Address.Bytes(), accountList[i].Address.Bytes())
	}
}
