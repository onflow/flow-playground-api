package vm

import (
	"math/rand"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-go/engine/execution/computation/virtualmachine"
	"github.com/dapperlabs/flow-go/engine/execution/state"
	"github.com/dapperlabs/flow-go/language/runtime"
	"github.com/dapperlabs/flow-go/model/flow"

	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/dapperlabs/flow-playground-api/storage"
)

type Computer struct {
	store        storage.Store
	blockContext virtualmachine.BlockContext
	cache        *LedgerCache
}

func NewComputer(store storage.Store, cacheSize int) (*Computer, error) {
	rt := runtime.NewInterpreterRuntime()
	vm := virtualmachine.New(rt)

	blockContext := vm.NewBlockContext(&flow.Header{Height: 0})

	cache, err := NewLedgerCache(cacheSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate LRU cache")
	}

	return &Computer{
		store:        store,
		blockContext: blockContext,
		cache:        cache,
	}, nil
}

type AccountState map[model.Address]map[string][]byte

func (c *Computer) ExecuteTransaction(
	projectID uuid.UUID,
	transactionCount int,
	getRegisterDeltas func() ([]state.Delta, error),
	script string,
	signers []model.Address,
) (
	*virtualmachine.TransactionResult,
	state.Delta,
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
		scriptAccounts[i] = flow.Address(signer)
	}

	transactionBody := flow.TransactionBody{
		Nonce:          rand.Uint64(),
		Script:         []byte(script),
		ScriptAccounts: scriptAccounts,
	}

	data := AccountState{}

	valueHandler := func(owner, controller, key, value []byte) {
		address := model.Address(flow.BytesToAddress(owner))

		if _, ok := data[address]; !ok {
			data[address] = map[string][]byte{}
		}

		data[address][string(key)] = value
	}

	setValueHandler := func(context *virtualmachine.TransactionContext) {
		context.OnSetValueHandler = valueHandler
	}

	result, err := c.blockContext.ExecuteTransaction(view, &transactionBody, setValueHandler)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "vm failed to execute transaction")
	}

	delta := view.Delta()

	ledgerItem.ledger.ApplyDelta(delta)
	ledgerItem.count++

	c.cache.Set(projectID, ledgerItem)

	return result, delta, data, nil
}

func (c *Computer) ClearCache() {
	c.cache.Clear()
}

func (c *Computer) getOrCreateLedger(
	projectID uuid.UUID,
	transactionCount int,
	getRegisterDeltas func() ([]state.Delta, error),
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
		ledger.ApplyDelta(delta)
	}

	ledgerItem = LedgerCacheItem{
		ledger: ledger,
		count:  transactionCount,
	}

	c.cache.Set(projectID, ledgerItem)

	return ledgerItem, nil
}

func (c *Computer) ExecuteScript(
	projectID uuid.UUID,
	transactionCount int,
	getRegisterDeltas func() ([]state.Delta, error),
	script string,
) (*virtualmachine.ScriptResult, error) {
	ledgerItem, err := c.getOrCreateLedger(projectID, transactionCount, getRegisterDeltas)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ledger for project")
	}

	view := ledgerItem.ledger.NewView()

	result, err := c.blockContext.ExecuteScript(view, []byte(script))
	if err != nil {
		return nil, errors.Wrap(err, "vm failed to execute script")
	}

	return result, nil
}

func (c *Computer) ClearCacheForProject(projectID uuid.UUID) {
	c.cache.Delete(projectID)
}
