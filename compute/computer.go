/*
 * Flow Playground
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package compute

import (
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/flow-go/fvm/programs"
	"github.com/onflow/flow-go/fvm/state"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/onflow/cadence/runtime"
	"github.com/onflow/flow-go/engine/execution/state/delta"
	"github.com/onflow/flow-go/fvm"
	"github.com/onflow/flow-go/model/flow"

	"github.com/dapperlabs/flow-playground-api/model"
)

type Computer struct {
	vm    *fvm.VirtualMachine
	vmCtx fvm.Context
	cache *LedgerCache
}

type TransactionResult struct {
	Err    error
	Logs   []string
	Events []flow.Event
	Delta  delta.Delta
	States AccountStates
}

type ScriptResult struct {
	Value  cadence.Value
	Err    error
	Logs   []string
	Events []flow.Event
}

type AccountStates map[model.Address]model.AccountState

func NewComputer(logger zerolog.Logger, cacheSize int) (*Computer, error) {
	rt := runtime.NewInterpreterRuntime()
	vm := fvm.NewVirtualMachine(rt)

	vmCtx := fvm.NewContext(
		logger,
		fvm.WithChain(flow.MonotonicEmulator.Chain()),
		fvm.WithServiceAccount(false),
		fvm.WithRestrictedAccountCreation(false),
		fvm.WithRestrictedDeployment(false),
		fvm.WithTransactionProcessors(
			fvm.NewTransactionInvocator(logger),
		),
		fvm.WithCadenceLogging(true),
		fvm.WithAccountStorageLimit(false),
	)

	cache, err := NewLedgerCache(cacheSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate LRU cache")
	}

	return &Computer{
		vm:    vm,
		vmCtx: vmCtx,
		cache: cache,
	}, nil
}

func (c *Computer) ExecuteTransaction(
	projectID uuid.UUID,
	transactionNumber int,
	getRegisterDeltas func() ([]*model.RegisterDelta, error),
	txBody *flow.TransactionBody,
) (*TransactionResult, error) {
	ledger, err := c.cache.GetOrCreate(projectID, transactionNumber, getRegisterDeltas)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ledger for project")
	}

	ctx := fvm.NewContextFromParent(
		c.vmCtx,
	)

	// Use the default gas limit
	txBody.GasLimit = ctx.GasLimit

	proc := fvm.Transaction(txBody, 0)

	view := ledger.NewView()

	err = c.vm.Run(ctx, proc, view, programs.NewEmptyPrograms())
	if err != nil {
		return nil, errors.Wrap(err, "vm failed to execute transaction")
	}

	delta := view.Delta()
	prog := programs.NewEmptyPrograms()

	states := extractStateChangesFromDelta(delta, func(address common.Address, key string) (value cadence.Value, err error) {
		pathParts := strings.Split(key, "\x1F")

		if len(pathParts) != 2 {
			// not a cadence path value
			return nil, nil
		}

		path := cadence.Path{
			Domain:     pathParts[0],
			Identifier: pathParts[1],
		}

		defer func() {
			if r := recover(); r != nil {
				value = nil
				err = nil
			}
		}()

		value, err = c.vm.Runtime.ReadStored(address, path, runtime.Context{Interface: &apiEnv{
			Delta:    &delta,
			Programs: prog,
		}})
		return
	})

	ledger.ApplyDelta(delta)

	c.cache.Set(projectID, ledger, transactionNumber)

	result := TransactionResult{
		Err:    proc.Err,
		Logs:   proc.Logs,
		Events: proc.Events,
		Delta:  delta,
		States: states,
	}

	return &result, nil
}

func (c *Computer) ExecuteScript(
	projectID uuid.UUID,
	transactionNumber int,
	getRegisterDeltas func() ([]*model.RegisterDelta, error),
	script string,
	arguments []string,
) (*ScriptResult, error) {
	ledger, err := c.cache.GetOrCreate(projectID, transactionNumber, getRegisterDeltas)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ledger for project")
	}

	view := ledger.NewView()

	rawArguments := make([][]byte, len(arguments))
	for i, argument := range arguments {
		rawArguments[i] = []byte(argument)
	}

	proc := fvm.Script([]byte(script)).
		WithArguments(rawArguments...)

	err = c.vm.Run(c.vmCtx, proc, view, programs.NewEmptyPrograms())
	if err != nil {
		return nil, errors.Wrap(err, "vm failed to execute script")
	}

	result := ScriptResult{
		Value:  proc.Value,
		Err:    proc.Err,
		Logs:   proc.Logs,
		Events: proc.Events,
	}

	return &result, nil
}

func (c *Computer) ClearCache() {
	c.cache.Clear()
}

func (c *Computer) ClearCacheForProject(projectID uuid.UUID) {
	c.cache.Delete(projectID)
}

func extractStateChangesFromDelta(d delta.Delta, getStored func(address common.Address, key string) (cadence.Value, error)) AccountStates {
	states := make(AccountStates)

	ids, _ := d.RegisterUpdates()

	for _, id := range ids {
		addressBytes := []byte(id.Owner)
		if len(addressBytes) != flow.AddressLength {
			continue
		}
		commonAddress := common.BytesToAddress(addressBytes)
		modelAddress := model.NewAddressFromBytes(addressBytes)

		value, err := getStored(commonAddress, id.Key)

		if err != nil || value == nil {
			// problem getting stored value
			continue
		}

		if states[modelAddress] == nil {
			states[modelAddress] = make(model.AccountState)
		}

		states[modelAddress][id.Key] = value
	}

	return states
}

var _ runtime.Interface = &apiEnv{}

type apiEnv struct {
	Delta    *delta.Delta
	Programs *programs.Programs
}

func (a *apiEnv) ResolveLocation(identifiers []runtime.Identifier, location runtime.Location) ([]runtime.ResolvedLocation, error) {
	panic("implement ResolveLocation")
}

func (a *apiEnv) GetCode(location runtime.Location) ([]byte, error) {
	panic("implement GetCode")
}

func (a *apiEnv) GetProgram(location runtime.Location) (*interpreter.Program, error) {
	p, _, _ := a.Programs.Get(location)
	return p, nil
}

func (a *apiEnv) SetProgram(location runtime.Location, program *interpreter.Program) error {
	a.Programs.Set(location, program, nil)
	return nil
}

func (a *apiEnv) GetValue(owner, key []byte) (value []byte, err error) {
	v, _ := a.Delta.Get(string(owner), "", string(key))
	return v, nil
}

func (a *apiEnv) SetValue(owner, key, value []byte) (err error) {
	panic("implement SetValue")
}

func (a *apiEnv) CreateAccount(payer runtime.Address) (address runtime.Address, err error) {
	panic("implement CreateAccount")
}

func (a *apiEnv) AddEncodedAccountKey(address runtime.Address, publicKey []byte) error {
	panic("implement AddEncodedAccountKey")
}

func (a *apiEnv) RevokeEncodedAccountKey(address runtime.Address, index int) (publicKey []byte, err error) {
	panic("implement RevokeEncodedAccountKey")
}

func (a *apiEnv) AddAccountKey(address runtime.Address, publicKey *runtime.PublicKey, hashAlgo runtime.HashAlgorithm, weight int) (*runtime.AccountKey, error) {
	panic("implement AddAccountKey")
}

func (a *apiEnv) GetAccountKey(address runtime.Address, index int) (*runtime.AccountKey, error) {
	panic("implement GetAccountKey")
}

func (a *apiEnv) RevokeAccountKey(address runtime.Address, index int) (*runtime.AccountKey, error) {
	panic("implement RevokeAccountKey")
}

func (a *apiEnv) UpdateAccountContractCode(address runtime.Address, name string, code []byte) (err error) {
	panic("implement UpdateAccountContractCode")
}

func (a *apiEnv) GetAccountContractCode(address runtime.Address, name string) (code []byte, err error) {
	addr := string(flow.BytesToAddress(address.Bytes()).Bytes())
	v, _ := a.Delta.Get(addr, addr, state.ContractKey(name))
	return v, nil
}

func (a *apiEnv) RemoveAccountContractCode(address runtime.Address, name string) (err error) {
	panic("implement RemoveAccountContractCode")
}

func (a *apiEnv) GetSigningAccounts() ([]runtime.Address, error) {
	panic("implement GetSigningAccounts")
}

func (a *apiEnv) ProgramLog(s string) error {
	panic("implement ProgramLog")
}

func (a *apiEnv) EmitEvent(event cadence.Event) error {
	panic("implement EmitEvent")
}

func (a *apiEnv) ValueExists(owner, key []byte) (exists bool, err error) {
	panic("implement ValueExists")
}

func (a *apiEnv) GenerateUUID() (uint64, error) {
	panic("implement GenerateUUID")
}

func (a *apiEnv) GetComputationLimit() uint64 {
	return math.MaxUint64
}

func (a *apiEnv) SetComputationUsed(used uint64) error {
	return nil
}

func (a *apiEnv) DecodeArgument(argument []byte, argumentType cadence.Type) (cadence.Value, error) {
	panic("implement DecodeArgument")
}

func (a *apiEnv) GetCurrentBlockHeight() (uint64, error) {
	panic("implement GetCurrentBlockHeight")
}

func (a *apiEnv) GetBlockAtHeight(height uint64) (block runtime.Block, exists bool, err error) {
	panic("implement GetBlockAtHeight")
}

func (a *apiEnv) UnsafeRandom() (uint64, error) {
	panic("implement UnsafeRandom")
}

func (a *apiEnv) VerifySignature(signature []byte, tag string, signedData []byte, publicKey []byte, signatureAlgorithm runtime.SignatureAlgorithm, hashAlgorithm runtime.HashAlgorithm) (bool, error) {
	panic("implement VerifySignature")
}

func (a *apiEnv) Hash(data []byte, tag string, hashAlgorithm runtime.HashAlgorithm) ([]byte, error) {
	panic("implement Hash")
}

func (a *apiEnv) GetAccountBalance(address common.Address) (value uint64, err error) {
	panic("implement GetAccountBalance")
}

func (a *apiEnv) GetAccountAvailableBalance(address common.Address) (value uint64, err error) {
	panic("implement GetAccountAvailableBalance")
}

func (a *apiEnv) GetStorageUsed(address runtime.Address) (value uint64, err error) {
	panic("implement GetStorageUsed")
}

func (a *apiEnv) GetStorageCapacity(address runtime.Address) (value uint64, err error) {
	panic("implement GetStorageCapacity")
}

func (a *apiEnv) ImplementationDebugLog(message string) error {
	panic("implement ImplementationDebugLog")
}

func (a *apiEnv) ValidatePublicKey(key *runtime.PublicKey) (bool, error) {
	panic("implement ValidatePublicKey")
}
