package vm

import (
	"fmt"
	"math/rand"

	"github.com/dapperlabs/flow-go/engine/execution/execution/state"
	"github.com/dapperlabs/flow-go/engine/execution/execution/virtualmachine"
	"github.com/dapperlabs/flow-go/language/runtime"
	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/google/uuid"
	"github.com/pkg/errors"

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

	blockContext := vm.NewBlockContext(&flow.Header{Number: 0})

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

func (c *Computer) ExecuteTransaction(
	projectID uuid.UUID,
	script string,
	signers []model.Address,
) (*virtualmachine.TransactionResult, state.Delta, error) {
	ledgerItem, err := c.getOrCreateLedger(projectID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get ledger for project")
	}

	view := ledgerItem.ledger.NewView()

	scriptAccounts := make([]flow.Address, len(signers))
	for i, signer := range signers {
		scriptAccounts[i] = flow.Address(signer)
	}

	result, err := c.blockContext.ExecuteTransaction(view, &flow.TransactionBody{
		Nonce:          rand.Uint64(),
		Script:         []byte(script),
		ScriptAccounts: scriptAccounts,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "vm failed to execute transaction")
	}

	delta := view.Delta()

	ledgerItem.ledger.ApplyDelta(delta)
	ledgerItem.count++

	c.cache.Set(projectID, ledgerItem)

	return result, delta, nil
}

func (c *Computer) ClearCache() {
	c.cache.Clear()
}

func (c *Computer) getOrCreateLedger(projectID uuid.UUID) (LedgerCacheItem, error) {
	var proj model.InternalProject
	err := c.store.GetProject(projectID, &proj)
	if err != nil {
		return LedgerCacheItem{}, errors.Wrap(err, "failed to load project")
	}

	if proj.TransactionCount == 0 {
		return LedgerCacheItem{
			ledger: make(Ledger),
			count:  0,
		}, nil
	}

	ledgerItem, ok := c.cache.Get(projectID)
	if ok && ledgerItem.count == proj.TransactionCount {
		return ledgerItem, nil
	}

	ledger := make(Ledger)

	var deltas []state.Delta

	err = c.store.GetRegisterDeltasForProject(projectID, &deltas)
	if err != nil {
		return LedgerCacheItem{}, errors.Wrap(err, "failed to load register deltas for project")
	}

	fmt.Println("REBUILDING CACHE", deltas)

	for _, delta := range deltas {
		ledger.ApplyDelta(delta)
	}

	ledgerItem = LedgerCacheItem{
		ledger: ledger,
		count:  proj.TransactionCount,
	}

	c.cache.Set(projectID, ledgerItem)

	return ledgerItem, nil
}

func (c *Computer) ExecuteScript(
	projectID uuid.UUID,
	script string,
) (*virtualmachine.ScriptResult, error) {
	ledgerItem, err := c.getOrCreateLedger(projectID)
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
