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
	vm     *fvm.VirtualMachine
	vmCtx  fvm.Context
	cache  *LedgerCache
	logger zerolog.Logger
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
		vm:     vm,
		vmCtx:  vmCtx,
		cache:  cache,
		logger: logger,
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
	prog := programs.NewEmptyPrograms()

	err = c.vm.Run(ctx, proc, view, prog)
	if err != nil {
		return nil, errors.Wrap(err, "vm failed to execute transaction")
	}

	d := view.Delta()

	states := c.extractStateChangesFromDelta(d, prog)

	ledger.ApplyDelta(d)

	c.cache.Set(projectID, ledger, transactionNumber)

	result := TransactionResult{
		Err:    proc.Err,
		Logs:   proc.Logs,
		Events: proc.Events,
		Delta:  d,
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

func (c *Computer) extractStateChangesFromDelta(d delta.Delta, p *programs.Programs) AccountStates {

	runtimeContext := runtime.Context{Interface: &apiEnv{
		Delta:    &d,
		Programs: p,
	}}

	getStored := func(address common.Address, key string) (value cadence.Value, err error) {
		pathParts := strings.Split(key, "\x1F")

		if len(pathParts) != 2 {
			// Not a cadence path value. Safe to ignore.
			return nil, nil
		}

		path := cadence.Path{
			Domain:     pathParts[0],
			Identifier: pathParts[1],
		}

		domain := common.PathDomainFromIdentifier(path.Domain)
		if domain == common.PathDomainUnknown {
			// Not a cadence path value. Safe to ignore.
			return nil, nil
		}

		defer func() {
			if r := recover(); r != nil {
				// Something went wrong, could be that this isn't a cadence value.
				c.logger.Debug().
					Err(r.(error)).
					Str("address", address.Hex()).
					Str("key", key).
					Msgf("error decoding value")

				value = nil
				err = nil
			}
		}()

		return c.vm.Runtime.ReadStored(address, path, runtimeContext)
	}

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
			// Not a cadence value or problem getting value.
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

func (a *apiEnv) ResolveLocation(_ []runtime.Identifier, _ runtime.Location) ([]runtime.ResolvedLocation, error) {
	panic("implement ResolveLocation")
}

func (a *apiEnv) GetCode(_ runtime.Location) ([]byte, error) {
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

func (a *apiEnv) SetValue(_, _, _ []byte) (err error) {
	panic("implement SetValue")
}

func (a *apiEnv) CreateAccount(_ runtime.Address) (address runtime.Address, err error) {
	panic("implement CreateAccount")
}

func (a *apiEnv) AddEncodedAccountKey(_ runtime.Address, _ []byte) error {
	panic("implement AddEncodedAccountKey")
}

func (a *apiEnv) RevokeEncodedAccountKey(_ runtime.Address, _ int) (publicKey []byte, err error) {
	panic("implement RevokeEncodedAccountKey")
}

func (a *apiEnv) AddAccountKey(_ runtime.Address, _ *runtime.PublicKey, _ runtime.HashAlgorithm, _ int) (*runtime.AccountKey, error) {
	panic("implement AddAccountKey")
}

func (a *apiEnv) GetAccountKey(_ runtime.Address, _ int) (*runtime.AccountKey, error) {
	panic("implement GetAccountKey")
}

func (a *apiEnv) RevokeAccountKey(_ runtime.Address, _ int) (*runtime.AccountKey, error) {
	panic("implement RevokeAccountKey")
}

func (a *apiEnv) UpdateAccountContractCode(_ runtime.Address, _ string, _ []byte) (err error) {
	panic("implement UpdateAccountContractCode")
}

func (a *apiEnv) GetAccountContractCode(address runtime.Address, name string) (code []byte, err error) {
	addr := string(flow.BytesToAddress(address.Bytes()).Bytes())
	v, _ := a.Delta.Get(addr, addr, state.ContractKey(name))
	return v, nil
}

func (a *apiEnv) RemoveAccountContractCode(_ runtime.Address, _ string) (err error) {
	panic("implement RemoveAccountContractCode")
}

func (a *apiEnv) GetSigningAccounts() ([]runtime.Address, error) {
	panic("implement GetSigningAccounts")
}

func (a *apiEnv) ProgramLog(_ string) error {
	panic("implement ProgramLog")
}

func (a *apiEnv) EmitEvent(_ cadence.Event) error {
	panic("implement EmitEvent")
}

func (a *apiEnv) ValueExists(_, _ []byte) (exists bool, err error) {
	panic("implement ValueExists")
}

func (a *apiEnv) GenerateUUID() (uint64, error) {
	panic("implement GenerateUUID")
}

func (a *apiEnv) GetComputationLimit() uint64 {
	return math.MaxUint64
}

func (a *apiEnv) SetComputationUsed(_ uint64) error {
	return nil
}

func (a *apiEnv) DecodeArgument(_ []byte, _ cadence.Type) (cadence.Value, error) {
	panic("implement DecodeArgument")
}

func (a *apiEnv) GetCurrentBlockHeight() (uint64, error) {
	panic("implement GetCurrentBlockHeight")
}

func (a *apiEnv) GetBlockAtHeight(_ uint64) (block runtime.Block, exists bool, err error) {
	panic("implement GetBlockAtHeight")
}

func (a *apiEnv) UnsafeRandom() (uint64, error) {
	panic("implement UnsafeRandom")
}

func (a *apiEnv) VerifySignature(_ []byte, _ string, _ []byte, _ []byte, _ runtime.SignatureAlgorithm, _ runtime.HashAlgorithm) (bool, error) {
	panic("implement VerifySignature")
}

func (a *apiEnv) Hash(_ []byte, _ string, _ runtime.HashAlgorithm) ([]byte, error) {
	panic("implement Hash")
}

func (a *apiEnv) GetAccountBalance(_ common.Address) (value uint64, err error) {
	panic("implement GetAccountBalance")
}

func (a *apiEnv) GetAccountAvailableBalance(_ common.Address) (value uint64, err error) {
	panic("implement GetAccountAvailableBalance")
}

func (a *apiEnv) GetStorageUsed(_ runtime.Address) (value uint64, err error) {
	panic("implement GetStorageUsed")
}

func (a *apiEnv) GetStorageCapacity(_ runtime.Address) (value uint64, err error) {
	panic("implement GetStorageCapacity")
}

func (a *apiEnv) ImplementationDebugLog(_ string) error {
	panic("implement ImplementationDebugLog")
}

func (a *apiEnv) ValidatePublicKey(_ *runtime.PublicKey) (bool, error) {
	panic("implement ValidatePublicKey")
}
