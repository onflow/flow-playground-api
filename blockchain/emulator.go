package blockchain

import (
	"fmt"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-emulator/storage/memstore"
	"github.com/onflow/flow-emulator/types"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/pkg/errors"
)

type Blockchain interface {
	ExecuteTransaction(
		script string,
		arguments []string,
		authorizers []model.Address,
	) (*types.TransactionResult, error)
	ExecuteScript(
		script string,
		arguments []string,
	) (*types.ScriptResult, error)
	CreateAccount() (*flowsdk.Account, *types.TransactionResult, error)
	GetAccount(address model.Address) (*flowsdk.Account, error)
	DeployContract(address model.Address, script string) (*types.TransactionResult, string, error)
}

var _ Blockchain = &Emulator{}

type Emulator struct {
	blockchain *emulator.Blockchain
}

// todo think about using flowkit

func NewEmulator() (*Emulator, error) {
	blockchain, err := emulator.NewBlockchain(
		emulator.WithStore(memstore.New()),
		emulator.WithTransactionValidationEnabled(false),
		emulator.WithSimpleAddresses(),
		emulator.WithStorageLimitEnabled(false),
		emulator.WithTransactionFeesEnabled(false),
	)
	if err != nil {
		return nil, err
	}

	return &Emulator{
		blockchain: blockchain,
	}, nil
}

func (e *Emulator) ExecuteTransaction(
	script string,
	arguments []string,
	authorizers []model.Address,
) (*types.TransactionResult, error) {
	tx := &flowsdk.Transaction{}
	tx.Script = []byte(script)

	args, err := parseArguments(arguments)
	if err != nil {
		return nil, err
	}
	tx.Arguments = args

	return e.sendTransaction(tx, authorizers)
}

func (e *Emulator) ExecuteScript(script string, arguments []string) (*types.ScriptResult, error) {
	args, err := parseArguments(arguments)
	if err != nil {
		return nil, err
	}

	return e.blockchain.ExecuteScript([]byte(script), args)
}

func (e *Emulator) CreateAccount() (*flowsdk.Account, *types.TransactionResult, error) {
	payer := e.blockchain.ServiceKey().Address
	key := flowsdk.NewAccountKey()
	key.FromPrivateKey(e.blockchain.ServiceKey().PrivateKey)
	key.HashAlgo = crypto.SHA3_256
	key.SetWeight(1000)

	tx, err := templates.CreateAccount([]*flowsdk.AccountKey{key}, nil, payer)
	if err != nil {
		return nil, nil, err
	}

	result, err := e.sendTransaction(tx, nil)
	if err != nil {
		return nil, nil, err
	}

	var address flowsdk.Address
	for _, event := range result.Events {
		if event.Type == flowsdk.EventAccountCreated {
			addressValue := event.Value.Fields[0].(cadence.Address)
			address = flowsdk.HexToAddress(addressValue.Hex())
			break
		}
	}

	return &flowsdk.Account{
		Address: address,
	}, result, nil
}

func (e *Emulator) GetAccount(address model.Address) (*flowsdk.Account, error) {
	return e.blockchain.GetAccount(address.ToFlowAddress())
}

func (e *Emulator) DeployContract(address model.Address, script string) (*types.TransactionResult, string, error) {
	contractName, err := getSourceContractName(script)
	if err != nil {
		return nil, "", err
	}

	tx := templates.AddAccountContract(address.ToFlowAddress(), templates.Contract{
		Name:   contractName,
		Source: script,
	})

	result, err := e.sendTransaction(tx, nil)
	if err != nil {
		return nil, "", err
	}

	return result, contractName, nil
}

func (e *Emulator) sendTransaction(tx *flowsdk.Transaction, authorizers []model.Address) (*types.TransactionResult, error) {
	signer, err := e.blockchain.ServiceKey().Signer()
	if err != nil {
		return nil, err
	}

	for _, auth := range authorizers {
		tx.AddAuthorizer(auth.ToFlowAddress())
	}
	tx.SetPayer(e.blockchain.ServiceKey().Address)

	for _, auth := range authorizers {
		err := tx.SignPayload(auth.ToFlowAddress(), 0, signer)
		if err != nil {
			return nil, err
		}
	}

	err = tx.SignEnvelope(e.blockchain.ServiceKey().Address, e.blockchain.ServiceKey().Index, signer)
	if err != nil { // todo should we return as transaction result error
		return nil, err
	}

	err = e.blockchain.AddTransaction(*tx)
	if err != nil { // return as transaction result error
		return &types.TransactionResult{
			Error: err,
		}, nil
	}

	_, res, err := e.blockchain.ExecuteAndCommitBlock()
	if err != nil {
		return nil, err
	}

	if len(res) != 1 {
		return nil, fmt.Errorf("failure during transaction execution")
	}

	return res[0], nil
}

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

func getSourceContractName(code string) (string, error) {
	program, err := parser.ParseProgram(code, nil)
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
