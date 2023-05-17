package blockchain

import (
	"context"
	"github.com/onflow/cadence"
	kit "github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/accounts"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/gateway"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/flowkit/tests"
	"github.com/onflow/flow-cli/flowkit/transactions"
	emu "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-emulator/storage/memstore"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/pkg/errors"
)

var _ blockchain = &flowKit{}

type flowKit struct {
	blockchain *kit.Flowkit
}

func newFlowkit() (*flowKit, error) {
	readerWriter, _ := tests.ReaderWriter()
	state, err := kit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create flow-kit state")
	}

	gw := gateway.NewEmulatorGatewayWithOpts(
		&gateway.EmulatorKey{
			PublicKey: emu.DefaultServiceKey().PublicKey,
			SigAlgo:   emu.DefaultServiceKeySigAlgo,
			HashAlgo:  emu.DefaultServiceKeyHashAlgo,
		},
		gateway.WithEmulatorOptions(
			emu.WithStore(memstore.New()),
			emu.WithTransactionValidationEnabled(false),
			emu.WithSimpleAddresses(),
			emu.WithStorageLimitEnabled(false),
			emu.WithTransactionFeesEnabled(false),
			emu.WithContractRemovalEnabled(true),
		),
	)

	return &flowKit{
		blockchain: kit.NewFlowkit(
			state,
			config.EmulatorNetwork,
			gw,
			output.NewStdoutLogger(output.NoneLog)),
	}, nil
}

func (fk *flowKit) executeTransaction(
	script string,
	arguments []string,
	authorizers []flow.Address,
) (*flow.Transaction, *flow.TransactionResult, error) {
	tx := &flow.Transaction{}
	tx.Script = []byte(script)

	args, err := parseArguments(arguments)
	if err != nil {
		return nil, nil, err
	}
	tx.Arguments = args

	return fk.sendTransaction(tx, authorizers)
}

func (fk *flowKit) executeScript(script string, arguments []string) (cadence.Value, error) {
	cadenceArgs := make([]cadence.Value, len(arguments))
	for i, arg := range arguments {
		val, err := cadence.NewValue(arg)
		if err != nil {
			return nil, err
		}
		cadenceArgs[i] = val
	}

	return fk.blockchain.ExecuteScript(
		context.Background(),
		kit.Script{
			Code:     []byte(script),
			Args:     cadenceArgs,
			Location: "",
		},
		kit.LatestScriptQuery)
}

func (fk *flowKit) createAccount() (*flow.Account, error) {
	state, err := fk.blockchain.State()
	if err != nil {
		return nil, err
	}

	service, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}
	serviceKey, err := service.Key.PrivateKey()
	if err != nil {
		return nil, err
	}

	account, _, err := fk.blockchain.CreateAccount(
		context.Background(),
		service,
		[]accounts.PublicKey{{
			Public:   (*serviceKey).PublicKey(),
			Weight:   flow.AccountKeyWeightThreshold,
			SigAlgo:  crypto.ECDSA_P256,
			HashAlgo: crypto.SHA3_256,
		}},
	)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (fk *flowKit) getAccount(address flow.Address) (*flow.Account, *emu.AccountStorage, error) {
	account, err := fk.blockchain.GetAccount(context.Background(), address)
	// TODO: How do we get account storage from emulator?
	state, err := fk.blockchain.State()
	if err != nil {
		return nil, nil, err
	}

	storage, err := e.blockchain.GetAccountStorage(address)
	if err != nil {
		return nil, nil, err
	}

	account, err := e.blockchain.GetAccount(address)
	if err != nil {
		return nil, nil, err
	}

	return account, storage, nil
}

/*

func (e *emulator) deployContract(
	address flow.Address,
	script string,
) (*types.TransactionResult, *flow.Transaction, error) {
	contractName, err := parseContractName(script)
	if err != nil {
		return nil, nil, err
	}

	tx := templates.AddAccountContract(address, templates.Contract{
		Name:   contractName,
		Source: script,
	})

	return e.sendTransaction(tx, nil)
}

func (e *emulator) removeContract(
	address flow.Address,
	contractName string,
) (*types.TransactionResult, *flow.Transaction, error) {
	tx := templates.RemoveAccountContract(address, contractName)
	return e.sendTransaction(tx, nil)
}
*/

func (fk *flowKit) sendTransaction(
	tx *flow.Transaction,
	authorizers []flow.Address,
) (*flow.Transaction, *flow.TransactionResult, error) {
	state, err := fk.blockchain.State()
	if err != nil {
		return nil, nil, err
	}

	service, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, nil, err
	}

	var accountRoles transactions.AccountRoles
	accountRoles.Payer = *service
	accountRoles.Proposer = *service

	for _, auth := range authorizers {
		acc, _ := state.Accounts().ByAddress(auth)
		accountRoles.Authorizers = append(accountRoles.Authorizers, *acc)
	}

	args := make([]cadence.Value, len(tx.Arguments))
	for i := range tx.Arguments {
		arg, err := tx.Argument(i)
		if err != nil {
			return nil, nil, err
		}
		args[i] = arg
	}

	return fk.blockchain.SendTransaction(
		context.Background(),
		accountRoles,
		kit.Script{
			Code:     tx.Script,
			Args:     args,
			Location: "", // TODO: Do we need this?
		},
		tx.GasLimit,
	)
}

func (fk *flowKit) getLatestBlockHeight() (int, error) {
	block, err := fk.blockchain.Gateway().GetLatestBlock()
	if err != nil {
		return 0, err
	}
	return int(block.BlockHeader.Height), nil
}
