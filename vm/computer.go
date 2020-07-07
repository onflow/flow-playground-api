package vm

import (
	"github.com/google/uuid"
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

func NewComputer(cacheSize int) (*Computer, error) {
	rt := runtime.NewInterpreterRuntime()
	vm := fvm.New(rt)

	vmCtx := fvm.NewContext(
		fvm.WithChain(flow.MonotonicEmulator.Chain()),
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

type AccountState map[model.Address]map[string][]byte

func (c *Computer) ExecuteTransaction(
	projectID uuid.UUID,
	transactionCount int,
	getRegisterDeltas func() ([]*model.RegisterDelta, error),
	script string,
	signers []model.Address,
) (
	*fvm.TransactionProcedure,
	delta.Delta,
	AccountState,
	error,
) {
	ledgerItem, err := c.getOrCreateLedger(projectID, transactionCount, getRegisterDeltas)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get ledger for project")
	}

	view := ledgerItem.ledger.NewView()

	scriptAccounts := make([]flow.Address, len(signers))
	for i, signer := range signers {
		// TODO: Remove address conversion
		scriptAccounts[i] = signer.ToFlowAddress()
	}

	txBody := flow.NewTransactionBody().
		SetScript([]byte(script))

	for _, authorizer := range scriptAccounts {
		txBody.AddAuthorizer(authorizer)
	}

	data := AccountState{}

	// TODO: capture account resources
	// valueHandler := func(owner, controller, key, value []byte) {
	// 	// TODO: Remove address conversion
	// 	address := model.NewAddressFromBytes(owner)
	//
	// 	if _, ok := data[address]; !ok {
	// 		data[address] = map[string][]byte{}
	// 	}
	//
	// 	data[address][string(key)] = value
	// }
	//
	// setValueHandler := func(context *fvm.Context) {
	// 	context.OnSetValueHandler = valueHandler
	// }

	tx := fvm.Transaction(txBody)

	err = c.vm.Run(c.vmCtx, tx, view)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "vm failed to execute transaction")
	}

	delta := view.Delta()

	ledgerItem.ledger.ApplyDelta(delta)
	ledgerItem.count++

	c.cache.Set(projectID, ledgerItem)

	return tx, delta, data, nil
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
