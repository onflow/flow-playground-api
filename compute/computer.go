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
	"github.com/google/uuid"
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/flow-go/fvm/programs"
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

	states := extractStateChangesFromDelta(delta)

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

func extractStateChangesFromDelta(d delta.Delta) AccountStates {
	states := make(AccountStates)

	ids, values := d.RegisterUpdates()

	for i, id := range ids {
		addressBytes := []byte(id.Owner)
		if len(addressBytes) != flow.AddressLength {
			continue
		}
		commonAddress := common.BytesToAddress(addressBytes)

		storedData, version := interpreter.StripMagic(values[i])
		storedValue, err := interpreter.DecodeValue(storedData, &commonAddress, []string{id.Key}, version, nil)
		if err != nil {
			// skip this value, as it cannot be decoded
			continue
		}

		modelAddress := model.NewAddressFromBytes(addressBytes)

		if _, ok := states[modelAddress]; !ok {
			states[modelAddress] = make(map[string]cadence.Value)
		}

		inter, err := interpreter.NewInterpreter(nil, common.StringLocation("extractStateChangesFromDelta"))

		if err != nil {
			panic(err)
		}

		// How to convert interpreter.Value to cadence.Value
		states[modelAddress][id.Key] = runtime.ExportValue(storedValue, inter)
	}

	return states
}
