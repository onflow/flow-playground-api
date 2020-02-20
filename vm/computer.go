package vm

import (
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
	ledgerCache  map[uuid.UUID]Ledger
}

func NewComputer(store storage.Store) *Computer {
	return &Computer{
		store:        store,
		blockContext: virtualmachine.New(runtime.NewInterpreterRuntime()).NewBlockContext(&flow.Header{Number: 0}),
		// TODO: cache eviction
		ledgerCache: make(map[uuid.UUID]Ledger),
	}
}

func (c *Computer) ExecuteTransaction(
	projectID uuid.UUID,
	script string,
	signers []model.Address,
) (*virtualmachine.TransactionResult, state.Delta, error) {
	ledger, err := c.getOrCreateLedger(projectID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get ledger for project")
	}

	view := ledger.NewView()

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

	ledger.ApplyDelta(delta)

	return result, delta, nil
}

func (c *Computer) ClearCache() {
	c.ledgerCache = make(map[uuid.UUID]Ledger)
}

func (c *Computer) getOrCreateLedger(projectID uuid.UUID) (Ledger, error) {
	// TODO: check that cache is up-to-date

	ledger, ok := c.ledgerCache[projectID]
	if ok {
		return ledger, nil
	}

	ledger = make(Ledger)

	var deltas []state.Delta

	err := c.store.GetRegisterDeltasForProject(projectID, &deltas)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load register deltas for project")
	}

	for _, delta := range deltas {
		ledger.ApplyDelta(delta)
	}

	c.ledgerCache[projectID] = ledger

	return ledger, nil
}
