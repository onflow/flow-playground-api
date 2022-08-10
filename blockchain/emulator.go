package blockchain

import (
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-emulator/types"

	"github.com/google/uuid"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-emulator"
	"github.com/onflow/flow-emulator/storage/badger"
	"github.com/onflow/flow-go/model/flow"
)

type Blockchain interface {
	ExecuteTransaction(
		projectID uuid.UUID,
		script string,
		arguments []cadence.Value,
		signers []flow.Address,
	) (*TransactionResult, error)
	ExecuteScript(
		projectID uuid.UUID,
		script string,
		arguments []cadence.Value,
	) (*types.ScriptResult, error)
	CreateAccount(projectID uuid.UUID) (*Account, error)
	GetAccount(projectID uuid.UUID, address flow.Address) (*Account, error)
}

type TransactionResult struct{}

type ScriptResult struct{}

type Account struct{}

var _ Blockchain = &Emulator{}

type Emulator struct {
	blockchain *emulator.Blockchain
}

func NewEmulator() (*Emulator, error) {
	storage, err := badger.New(badger.WithPath("db-0"))
	if err != nil {
		return nil, err
	}

	blockchain, err := emulator.NewBlockchain(
		emulator.WithStore(storage),
	)
	if err != nil {
		return nil, err
	}

	return &Emulator{
		blockchain: blockchain,
	}, nil
}

func (e *Emulator) ExecuteTransaction(
	projectID uuid.UUID,
	script string,
	arguments []cadence.Value,
	signers []flow.Address,
) (*TransactionResult, error) {
	return nil, nil
}

func (e *Emulator) ExecuteScript(projectID uuid.UUID, script string, arguments []cadence.Value) (*types.ScriptResult, error) {
	encodedArgs := make([][]byte, len(arguments))
	for i, arg := range arguments {
		enc, err := jsoncdc.Encode(arg)
		if err != nil {
			return nil, err
		}
		encodedArgs[i] = enc
	}

	return e.blockchain.ExecuteScript([]byte(script), encodedArgs)
}

func (e *Emulator) CreateAccount(projectID uuid.UUID) (*Account, error) {
	return nil, nil
}

func (e *Emulator) GetAccount(projectID uuid.UUID, address flow.Address) (*Account, error) {
	return nil, nil
}
