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
) (
	*fvm.TransactionProcedure,
	delta.Delta,
	AccountStates,
	error,
) {
	ledgerItem, err := c.getOrCreateLedger(projectID, transactionCount, getRegisterDeltas)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get ledger for project")
	}

	view := ledgerItem.ledger.NewView()

	states := make(AccountStates)

	valueHandler := func(owner flow.Address, key string, value cadence.Value) error {

		// TODO: Remove address conversion
		address := model.NewAddressFromBytes(owner.Bytes())

		if _, ok := states[address]; !ok {
			states[address] = make(map[string]cadence.Value)
		}

		states[address][key] = value

		return nil
	}

	ctx := fvm.NewContextFromParent(c.vmCtx, fvm.WithSetValueHandler(valueHandler))

	tx := fvm.Transaction(txBody)

	err = c.vm.Run(ctx, tx, view)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "vm failed to execute transaction")
	}

	delta := view.Delta()

	ledgerItem.ledger.ApplyDelta(delta)
	ledgerItem.count++

	c.cache.Set(projectID, ledgerItem)

	return tx, delta, states, nil
}

func (c *Computer) ExecuteScript(
	projectID uuid.UUID,
	transactionCount int,
	getRegisterDeltas func() ([]*model.RegisterDelta, error),
	script string,
) (*fvm.ScriptProcedure, error) {
	ledgerItem, err := c.getOrCreateLedger(projectID, transactionCount, getRegisterDeltas)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ledger for project")
	}

	view := ledgerItem.ledger.NewView()

	scriptProc := fvm.Script([]byte(script))

	err = c.vm.Run(c.vmCtx, scriptProc, view)
	if err != nil {
		return nil, errors.Wrap(err, "vm failed to execute script")
	}

	return scriptProc, nil
}

func (c *Computer) ClearCache() {
	c.cache.Clear()
}

func (c *Computer) ClearCacheForProject(projectID uuid.UUID) {
	c.cache.Delete(projectID)
}

func (c *Computer) getOrCreateLedger(
	projectID uuid.UUID,
	transactionCount int,
	getRegisterDeltas func() ([]*model.RegisterDelta, error),
) (LedgerCacheItem, error) {
	if transactionCount == 0 {
		return LedgerCacheItem{
			ledger: make(Ledger),
			count:  0,
		}, nil
	}

	ledgerItem, ok := c.cache.Get(projectID)
	if ok && ledgerItem.count == transactionCount {
		return ledgerItem, nil
	}

	ledger := make(Ledger)

	deltas, err := getRegisterDeltas()
	if err != nil {
		return LedgerCacheItem{}, errors.Wrap(err, "failed to load register deltas for project")
	}

	for _, delta := range deltas {
		ledger.ApplyDelta(delta.Delta)
	}

	ledgerItem = LedgerCacheItem{
		ledger: ledger,
		count:  transactionCount,
	}

	c.cache.Set(projectID, ledgerItem)

	return ledgerItem, nil
}
