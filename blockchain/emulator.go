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
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
	emu "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-emulator/storage/memstore"
	"github.com/onflow/flow-emulator/types"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/pkg/errors"
)

// blockchain interface defines an abstract API for communication with the blockchain. It hides complexity from the
// consumer and communicates using flow native types.
type blockchain interface {
	// executeTransaction builds and executes a transaction and uses provided authorizers for signing.
	executeTransaction(
		script string,
		arguments []string,
		authorizers []flowsdk.Address,
	) (*types.TransactionResult, *flowsdk.Transaction, error)

	// executeScript executes a provided script with the arguments.
	executeScript(
		script string,
		arguments []string,
	) (*types.ScriptResult, error)

	// createAccount creates a new account and returns it along with transaction and result.
	createAccount() (*flowsdk.Account, *flowsdk.Transaction, *types.TransactionResult, error)

	// getAccount gets an account by the address and also returns its storage.
	getAccount(address flowsdk.Address) (*flowsdk.Account, *emu.AccountStorage, error)

	// deployContract deploys a contract on the provided address and returns transaction and result.
	deployContract(address flowsdk.Address, script string) (*types.TransactionResult, *flowsdk.Transaction, error)

	// removeContract removes specified contract from provided address and returns transaction and result.
	removeContract(address flowsdk.Address, contractName string) (*types.TransactionResult, *flowsdk.Transaction, error)

	// getLatestBlock height from the network.
	getLatestBlockHeight() (int, error)
}

var _ blockchain = &emulator{}

type emulator struct {
	blockchain *emu.Blockchain
}

func newEmulator() (*emulator, error) {
	blockchain, err := emu.NewBlockchain(
		emu.WithStore(memstore.New()),
		emu.WithTransactionValidationEnabled(false),
		emu.WithSimpleAddresses(),
		emu.WithStorageLimitEnabled(false),
		emu.WithTransactionFeesEnabled(false),
		emu.WithContractRemovalEnabled(true),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new emulator instance")
	}

	return &emulator{
		blockchain: blockchain,
	}, nil
}

func (e *emulator) executeTransaction(
	script string,
	arguments []string,
	authorizers []flowsdk.Address,
) (result *types.TransactionResult, tx *flowsdk.Transaction, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Panic occurred during transaction execution")
		}
	}()

	tx = &flowsdk.Transaction{}
	tx.Script = []byte(script)

	args, err := parseArguments(arguments)
	if err != nil {
		return nil, nil, err
	}
	tx.Arguments = args

	return e.sendTransaction(tx, authorizers)
}

func (e *emulator) executeScript(script string, arguments []string) (*types.ScriptResult, error) {
	args, err := parseArguments(arguments)
	if err != nil {
		return nil, err
	}

	return e.blockchain.ExecuteScript([]byte(script), args)
}

func (e *emulator) createAccount() (*flowsdk.Account, *flowsdk.Transaction, *types.TransactionResult, error) {
	payer := e.blockchain.ServiceKey().Address

	key := flowsdk.NewAccountKey().
		FromPrivateKey(e.blockchain.ServiceKey().PrivateKey).
		SetHashAlgo(crypto.SHA3_256).
		SetWeight(flowsdk.AccountKeyWeightThreshold)

	tx, err := templates.CreateAccount([]*flowsdk.AccountKey{key}, nil, payer)
	if err != nil {
		return nil, nil, nil, err
	}

	result, tx, err := e.sendTransaction(tx, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	account := &flowsdk.Account{
		Address: parseEventAddress(result.Events),
	}

	return account, tx, result, nil
}

func (e *emulator) getAccount(address flowsdk.Address) (*flowsdk.Account, *emu.AccountStorage, error) {
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

func (e *emulator) deployContract(
	address flowsdk.Address,
	script string,
) (*types.TransactionResult, *flowsdk.Transaction, error) {
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
	address flowsdk.Address,
	contractName string,
) (*types.TransactionResult, *flowsdk.Transaction, error) {
	tx := templates.RemoveAccountContract(address, contractName)
	return e.sendTransaction(tx, nil)
}

func (e *emulator) sendTransaction(
	tx *flowsdk.Transaction,
	authorizers []flowsdk.Address,
) (*types.TransactionResult, *flowsdk.Transaction, error) {
	signer, err := e.blockchain.ServiceKey().Signer()
	if err != nil {
		return nil, nil, errors.Wrap(err, "error getting service signer")
	}

	for _, auth := range authorizers {
		tx.AddAuthorizer(auth)
	}
	tx.SetPayer(e.blockchain.ServiceKey().Address)

	for _, auth := range authorizers {
		if len(authorizers) == 1 && tx.Payer == authorizers[0] {
			break // don't sign if we have same authorizer and payer, only sign envelope
		}

		err := tx.SignPayload(auth, 0, signer)
		if err != nil {
			return nil, nil, errors.Wrap(err, "error signing payload")
		}
	}

	err = tx.SignEnvelope(e.blockchain.ServiceKey().Address, e.blockchain.ServiceKey().Index, signer)
	if err != nil { // todo should we return as transaction result error
		return nil, nil, errors.Wrap(err, "error signing the envelope")
	}

	err = e.blockchain.AddTransaction(*tx)
	if err != nil {
		return &types.TransactionResult{
			Error: err,
		}, nil, nil
	}

	_, res, err := e.blockchain.ExecuteAndCommitBlock()
	if err != nil {
		return nil, nil, errors.Wrap(err, "error executing the block")
	}

	// there should always be just one transaction per block execution, if not the case fail
	if len(res) != 1 {
		sentry.CaptureMessage(fmt.Sprintf("%d transactions were executed: %v", len(res), res))
		return nil, nil, fmt.Errorf("failure during transaction execution, multiple transactions executed")
	}

	return res[0], tx, nil
}

func (e *emulator) getLatestBlockHeight() (int, error) {
	block, err := e.blockchain.GetLatestBlock()
	if err != nil {
		return 0, err
	}
	return int(block.Header.Height), nil
}

// parseEventAddress gets an address out of the account creation events payloads
func parseEventAddress(events []flowsdk.Event) flowsdk.Address {
	for _, event := range events {
		if event.Type == flowsdk.EventAccountCreated {
			addressValue := event.Value.Fields[0].(cadence.Address)
			return flowsdk.HexToAddress(addressValue.Hex())
		}
	}
	return flowsdk.EmptyAddress
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
