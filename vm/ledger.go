package vm

import "github.com/dapperlabs/flow-go/engine/execution/execution/state"

type Ledger map[string][]byte

func (l Ledger) NewView() *state.View {
	return state.NewView(func(key []byte) ([]byte, error) {
		return l[string(key)], nil
	})
}

func (l Ledger) ApplyDelta(delta state.Delta) {
	for key, value := range delta {
		if value != nil {
			l[key] = value
		} else {
			delete(l, key)
		}
	}
}
