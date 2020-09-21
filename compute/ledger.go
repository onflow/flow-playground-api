package compute

import (
	"github.com/dapperlabs/flow-go/engine/execution/state/delta"
	"github.com/dapperlabs/flow-go/fvm/state"
)

type Ledger map[string][]byte

func (l Ledger) NewView() *delta.View {
	return delta.NewView(func(owner, controller, key string) ([]byte, error) {
		id := state.RegisterID(owner, controller, key)
		return l[string(id)], nil
	})
}

func (l Ledger) ApplyDelta(delta delta.Delta) {
	for id, value := range delta.Data {
		l[id] = value
		// TODO: support delete
	}
}
