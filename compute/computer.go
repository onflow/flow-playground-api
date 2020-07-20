package compute

import (
	"github.com/google/uuid"
	"github.com/onflow/cadence"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-go/engine/execution/state/delta"
	"github.com/dapperlabs/flow-go/fvm"
	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/onflow/cadence/runtime"

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
	Events []cadence.Event
	Delta  delta.Delta
	States AccountStates
}

type ScriptResult struct {
	Value  cadence.Value
	Err    error
	Logs   []string
	Events []cadence.Event
}

type AccountStates map[model.Address]model.AccountState

func NewComputer(cacheSize int) (*Computer, error) {
	rt := runtime.NewInterpreterRuntime()
	vm := fvm.New(rt)

	vmCtx := fvm.NewContext(
		fvm.WithChain(flow.MonotonicEmulator.Chain()),
		fvm.WithServiceAccount(false),
		fvm.WithRestrictedAccountCreation(false),
		fvm.WithRestrictedDeployment(false),
		fvm.WithTransactionProcessors([]fvm.TransactionProcessor{
			fvm.NewTransactionInvocator(),
		}),
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
	transactionCount int,
	getRegisterDeltas func() ([]*model.RegisterDelta, error),
	txBody *flow.TransactionBody,
) (*TransactionResult, error) {
	ledgerItem, err := c.cache.GetOrCreate(projectID, transactionCount, getRegisterDeltas)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ledger for project")
	}

	states := make(AccountStates)

	ctx := fvm.NewContextFromParent(
		c.vmCtx,
		fvm.WithSetValueHandler(newValueHandler(states)),
	)

	proc := fvm.Transaction(txBody)

	view := ledgerItem.ledger.NewView()

	err = c.vm.Run(ctx, proc, view)
	if err != nil {
		return nil, errors.Wrap(err, "vm failed to execute transaction")
	}

	delta := view.Delta()

	ledgerItem.ledger.ApplyDelta(delta)
	ledgerItem.index++

	c.cache.Set(projectID, ledgerItem)

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
	transactionCount int,
	getRegisterDeltas func() ([]*model.RegisterDelta, error),
	script string,
) (*ScriptResult, error) {
	ledgerItem, err := c.cache.GetOrCreate(projectID, transactionCount, getRegisterDeltas)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ledger for project")
	}

	view := ledgerItem.ledger.NewView()

	proc := fvm.Script([]byte(script))

	err = c.vm.Run(c.vmCtx, proc, view)
	if err != nil {
		return nil, errors.Wrap(err, "vm failed to execute script")
	}

	result := ScriptResult{
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

func newValueHandler(states AccountStates) func(owner flow.Address, key string, value cadence.Value) error {
	return func(owner flow.Address, key string, value cadence.Value) error {
		// TODO: Remove address conversion
		address := model.NewAddressFromBytes(owner.Bytes())

		if _, ok := states[address]; !ok {
			states[address] = make(map[string]cadence.Value)
		}

		states[address][key] = value

		return nil
	}
}
