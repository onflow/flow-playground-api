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

package blockchain

import (
	"context"
	"fmt"
	userErr "github.com/dapperlabs/flow-playground-api/middleware/errors"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
	kit "github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/accounts"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/gateway"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/flowkit/transactions"
	emu "github.com/onflow/flow-emulator/emulator"
	"github.com/onflow/flow-emulator/storage/memstore"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// blockchain interface defines an abstract API for communication with the blockchain. It hides complexity from the
// consumer and communicates using flow native types.
type blockchain interface {
	// executeTransaction builds and executes a transaction and uses provided authorizers for signing.
	executeTransaction(
		script string,
		arguments []string,
		authorizers []flow.Address,
	) (*flow.Transaction, *flow.TransactionResult, Logs, error)

	// executeScript executes a provided script with the arguments.
	executeScript(script string, arguments []string) (cadence.Value, Logs, error)

	// createAccount creates a new account and returns it along with transaction and result.
	createAccount() (*flow.Account, error)

	// getAccount gets an account by the address
	getAccount(address flow.Address) (*flow.Account, error)

	// getAccountStorage gets storage for an account by the address
	getAccountStorage(address flow.Address) (string, error)

	// deployContract deploys a contract on the provided address and returns transaction and result.
	deployContract(
		address flow.Address,
		script string,
		arguments []string,
	) (*flow.Transaction, *flow.TransactionResult, Logs, error)

	// getLatestBlock height from the network.
	getLatestBlockHeight() (int, error)

	initBlockHeight() int

	numAccounts() int

	getFlowJson() (string, error)
}

var _ blockchain = &flowKit{}

// NumEmulatorAccounts 4 accounts with bootstrapped contracts: 0x01, 0x02, 0x03, 0x04
const NumEmulatorAccounts = 4

// initialAccounts number of accounts to bootstrap for users
const initialAccounts = 5

type flowKit struct {
	blockchain     *kit.Flowkit
	logInterceptor *Interceptor
}

func newFlowkit() (*flowKit, error) {
	readerWriter := NewInternalReaderWriter()
	state, err := kit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create flow-kit state")
	}

	interceptor := NewInterceptor()
	emulatorLogger := zerolog.New(interceptor)

	emulator := gateway.NewEmulatorGatewayWithOpts(
		&gateway.EmulatorKey{
			PublicKey: emu.DefaultServiceKey().AccountKey().PublicKey,
			SigAlgo:   emu.DefaultServiceKeySigAlgo,
			HashAlgo:  emu.DefaultServiceKeyHashAlgo,
		},
		gateway.WithEmulatorOptions(
			emu.WithLogger(emulatorLogger),
			emu.WithStore(memstore.New()),
			emu.WithTransactionValidationEnabled(false),
			emu.WithStorageLimitEnabled(false),
			emu.WithTransactionFeesEnabled(false),
			emu.WithSimpleAddresses(),
		),
	)

	fk := &flowKit{
		blockchain: kit.NewFlowkit(
			state,
			config.EmulatorNetwork,
			emulator,
			output.NewStdoutLogger(output.NoneLog)),
		logInterceptor: interceptor,
	}

	err = fk.bootstrap()
	if err != nil {
		return nil, err
	}

	return fk, nil
}

func (fk *flowKit) bootstrap() error {
	err := fk.boostrapAccounts()
	if err != nil {
		return err
	}

	err = fk.bootstrapContracts()
	if err != nil {
		return err
	}

	return nil
}

func (fk *flowKit) boostrapAccounts() error {
	err := fk.createServiceAccount()
	if err != nil {
		return err
	}

	err = fk.createEmulatorAccounts()
	if err != nil {
		return err
	}

	for i := 0; i < initialAccounts; i++ {
		_, err := fk.createAccount()
		if err != nil {
			return err
		}
	}
	return nil
}

func (fk *flowKit) bootstrapContracts() error {
	/* Bootstrapped Emulator Contracts
	0x01: FlowIDTableStaking, FlowStorageFees, MetadataViews, NonFungibleToken, ViewResolver
	FlowDKG, FlowEpoch, FlowServiceAccount, FlowStakingCollection, LockedTokens,
	NodeVersionBeacon, StakingProxy, FlowClusterQC
	0x02: FungibleToken, FungibleTokenMetadataViews
	0x03: FlowToken
	0x04: FlowFees
	*/

	state, err := fk.blockchain.State()
	if err != nil {
		return err
	}

	for i := 1; i <= NumEmulatorAccounts; i++ {
		var emulatorContracts []config.ContractDeployment

		// Get all bootstrapped contracts from each emulator account
		addr := flow.HexToAddress(fmt.Sprintf("0x0%d", i))
		account, err := fk.blockchain.GetAccount(context.Background(), addr)
		if err != nil {
			return err
		}

		// Create flowKit deployments for each contract to make them accessible
		for name := range account.Contracts {
			state.Contracts().AddOrUpdate(config.Contract{
				Name:     name,
				Location: "",
				Aliases:  nil,
			})

			emulatorContracts = append(emulatorContracts, config.ContractDeployment{
				Name: name,
				Args: nil, // TODO: Do we need to know what the arguments are?
			})
		}

		state.Deployments().AddOrUpdate(config.Deployment{
			Network:   config.EmulatorNetwork.Name,
			Account:   fmt.Sprintf("Emulator Account 0x%d", i),
			Contracts: emulatorContracts,
		})
	}

	return nil
}

// initBlockHeight returns what the bootstrapped block height should be
func (fk *flowKit) initBlockHeight() int {
	return initialAccounts
}

func (fk *flowKit) numAccounts() int {
	return initialAccounts
}

func (fk *flowKit) getFlowJson() (string, error) {
	// TODO: Maybe we shouldn't provide flow.json since it's too complicated and not useful?
	state, err := fk.blockchain.State()
	if err != nil {
		return "", err
	}

	err = state.Save("flow.json")
	if err != nil {
		return "", err
	}

	flowJson, err := state.ReaderWriter().ReadFile("")
	if err != nil {
		return "", err
	}

	return string(flowJson), nil
}

func (fk *flowKit) executeTransaction(
	script string,
	arguments []string,
	authorizers []flow.Address,
) (*flow.Transaction, *flow.TransactionResult, Logs, error) {
	tx := &flow.Transaction{}
	tx.Script = []byte(script)

	args, err := parseArguments(arguments)
	if err != nil {
		return nil, nil, nil, err
	}
	tx.Arguments = args

	return fk.sendTransaction(tx, authorizers)
}

func (fk *flowKit) executeScript(script string, arguments []string) (cadence.Value, Logs, error) {
	cadenceArgs := make([]cadence.Value, len(arguments))

	// Encode arguments using a transaction
	args, err := parseArguments(arguments)
	if err != nil {
		return nil, nil, err
	}
	tx := &flow.Transaction{
		Arguments: args,
	}

	for i := range arguments {
		cadenceArgs[i], err = tx.Argument(i)
		if err != nil {
			return nil, nil, err
		}
	}

	fk.logInterceptor.ClearLogs()

	val, err := fk.blockchain.ExecuteScript(
		context.Background(),
		kit.Script{
			Code:     []byte(script),
			Args:     cadenceArgs,
			Location: "",
		},
		kit.LatestScriptQuery)
	if err != nil {
		return nil, nil, userErr.NewUserError(err.Error())
	}

	logs := fk.logInterceptor.GetCadenceLogs()

	return val, logs, nil
}

func (fk *flowKit) createServiceAccount() error {
	state, err := fk.blockchain.State()
	if err != nil {
		return err
	}

	serviceAccount, err := fk.getServiceAccount()
	if err != nil {
		return err
	}

	state.Accounts().AddOrUpdate(&accounts.Account{
		Name:    "Service Account",
		Address: flow.HexToAddress("0x01"),
		Key:     serviceAccount.Key,
	})

	state.Deployments().AddOrUpdate(config.Deployment{ // init empty deployment
		Network: config.EmulatorNetwork.Name,
		Account: "Service Account",
	})

	return nil
}

func (fk *flowKit) createEmulatorAccounts() error {
	state, err := fk.blockchain.State()
	if err != nil {
		return err
	}

	serviceAccount, err := fk.getServiceAccount()
	if err != nil {
		return err
	}

	for i := 1; i <= NumEmulatorAccounts; i++ {
		addr := fmt.Sprintf("0x%d", i)
		name := fmt.Sprintf("Emulator Account %s", addr)

		state.Accounts().AddOrUpdate(&accounts.Account{
			Name:    name,
			Address: flow.HexToAddress(addr),
			Key:     serviceAccount.Key,
		})

		state.Deployments().AddOrUpdate(config.Deployment{ // init empty deployment
			Network: config.EmulatorNetwork.Name,
			Account: name,
		})
	}

	return nil
}

func (fk *flowKit) createAccount() (*flow.Account, error) {
	serviceAccount, err := fk.getServiceAccount()
	if err != nil {
		return nil, err
	}

	pk, err := serviceAccount.Key.PrivateKey()
	if err != nil {
		return nil, err
	}

	account, _, err := fk.blockchain.CreateAccount(
		context.Background(),
		serviceAccount,
		[]accounts.PublicKey{
			{
				Public:   (*pk).PublicKey(),
				Weight:   flow.AccountKeyWeightThreshold,
				SigAlgo:  crypto.ECDSA_P256,
				HashAlgo: crypto.SHA3_256,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	// Update state with new account
	state, err := fk.blockchain.State()
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("Account 0x0%d", len(state.Accounts().Names())-1)

	state.Accounts().AddOrUpdate(&accounts.Account{
		Name:    name,
		Address: account.Address,
		Key:     serviceAccount.Key,
	})

	state.Deployments().AddOrUpdate(config.Deployment{ // init empty deployment
		Network: config.EmulatorNetwork.Name,
		Account: name,
	})

	return account, nil
}

func (fk *flowKit) getAccountStorage(address flow.Address) (string, error) {
	args := []string{fmt.Sprintf(`{"type":"Address","value":"0x%s"}`, address.Hex())}
	val, _, err := fk.executeScript(StorageIteration, args)
	if err != nil {
		return "", err
	}

	storage, err := jsoncdc.Encode(val)
	if err != nil {
		return "", err
	}
	return string(storage), nil
}

func (fk *flowKit) getAccount(address flow.Address) (*flow.Account, error) {
	account, err := fk.blockchain.GetAccount(context.Background(), address)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (fk *flowKit) deployContract(
	address flow.Address,
	script string,
	arguments []string,
) (*flow.Transaction, *flow.TransactionResult, Logs, error) {
	state, err := fk.blockchain.State()
	if err != nil {
		return nil, nil, nil, err
	}

	to, err := state.Accounts().ByAddress(address)
	if err != nil {
		return nil, nil, nil, err
	}

	args, err := parseCadenceValues(arguments)
	if err != nil {
		return nil, nil, nil, err
	}

	fk.logInterceptor.ClearLogs()

	txID, _, err := fk.blockchain.AddContract(
		context.Background(),
		to,
		kit.Script{
			Code:     []byte(script),
			Args:     args,
			Location: "",
		},
		kit.UpdateExistingContract(false),
	)
	if err != nil {
		return nil, nil, nil, err
	}

	tx, result, err := fk.blockchain.GetTransactionByID(
		context.Background(),
		txID,
		true,
	)

	logs := fk.logInterceptor.GetCadenceLogs()

	return tx, result, logs, err
}

func (fk *flowKit) sendTransaction(
	tx *flow.Transaction,
	authorizers []flow.Address,
) (*flow.Transaction, *flow.TransactionResult, Logs, error) {
	serviceAccount, err := fk.getServiceAccount()
	if err != nil {
		return nil, nil, nil, err
	}

	var accountRoles transactions.AccountRoles
	accountRoles.Payer = *serviceAccount
	accountRoles.Proposer = *serviceAccount

	state, err := fk.blockchain.State()
	if err != nil {
		return nil, nil, nil, err
	}

	for _, auth := range authorizers {
		acc, err := state.Accounts().ByAddress(auth)
		if err != nil {
			return nil, nil, nil, err
		}
		accountRoles.Authorizers = append(accountRoles.Authorizers, *acc)
	}

	args := make([]cadence.Value, len(tx.Arguments))
	for i := range tx.Arguments {
		arg, err := tx.Argument(i)
		if err != nil {
			return nil, nil, nil, err
		}
		args[i] = arg
	}

	fk.logInterceptor.ClearLogs()

	tx, result, err := fk.blockchain.SendTransaction(
		context.Background(),
		accountRoles,
		kit.Script{
			Code:     tx.Script,
			Args:     args,
			Location: "",
		},
		tx.GasLimit,
	)
	if err != nil {
		return nil, nil, nil, userErr.NewUserError(err.Error())
	}

	logs := fk.logInterceptor.GetCadenceLogs()

	return tx, result, logs, nil
}

func (fk *flowKit) getLatestBlockHeight() (int, error) {
	block, err := fk.blockchain.Gateway().GetLatestBlock()
	if err != nil {
		return 0, err
	}
	return int(block.BlockHeader.Height), nil
}

func (fk *flowKit) getServiceAccount() (*accounts.Account, error) {
	state, err := fk.blockchain.State()
	if err != nil {
		return nil, err
	}

	service, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}

	return &accounts.Account{
		Name:    "Service Account",
		Address: flow.HexToAddress("0x01"),
		Key:     service.Key,
	}, nil
}

func parseCadenceValues(args []string) ([]cadence.Value, error) {
	cadenceArgs := make([]cadence.Value, len(args))
	for i, arg := range args {
		val, err := jsoncdc.Decode(nil, []byte(arg))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode argument")
		}
		cadenceArgs[i] = val
	}
	return cadenceArgs, nil
}

// parseArguments converts string arguments list in cadence-JSON format into a byte serialised list
func parseArguments(args []string) ([][]byte, error) {
	encodedArgs := make([][]byte, len(args))
	for i, arg := range args {
		// decode and then encode again to ensure the value is valid
		val, err := jsoncdc.Decode(nil, []byte(arg))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode argument")
		}
		enc, _ := jsoncdc.Encode(val)
		encodedArgs[i] = enc
	}

	return encodedArgs, nil
}

// parseContractName extracts contract name from its source
func parseContractName(code string) (string, error) {
	program, err := parser.ParseProgram(nil, []byte(code), parser.Config{})
	if err != nil {
		return "", err
	}
	if len(program.CompositeDeclarations())+len(program.InterfaceDeclarations()) != 1 {
		return "", errors.New("the code must declare exactly one contract or contract interface")
	}

	for _, compositeDeclaration := range program.CompositeDeclarations() {
		if compositeDeclaration.CompositeKind == common.CompositeKindContract {
			return compositeDeclaration.Identifier.Identifier, nil
		}
	}

	for _, interfaceDeclaration := range program.InterfaceDeclarations() {
		if interfaceDeclaration.CompositeKind == common.CompositeKindContract {
			return interfaceDeclaration.Identifier.Identifier, nil
		}
	}

	return "", fmt.Errorf("unable to determine contract name")
}

func GetInitialBlockHeightForTesting() int {
	fk, _ := newFlowkit()
	return fk.initBlockHeight()
}
