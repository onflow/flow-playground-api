package vm

import (
	"github.com/dapperlabs/flow-go/engine/execution/state/delta"
)

type Ledger map[string][]byte

func (l Ledger) NewView() *delta.View {
	return delta.NewView(func(key []byte) ([]byte, error) {
		return l[string(key)], nil
	})
}

func (l Ledger) ApplyDelta(delta delta.Delta) {
	for key, value := range delta {
		l[key] = value
		// TODO: support delete
	}
}
