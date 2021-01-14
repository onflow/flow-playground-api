package compute

import (
	"github.com/onflow/flow-go/engine/execution/state/delta"
	"github.com/onflow/flow-go/model/flow"
)

type Ledger map[string]flow.RegisterEntry

func (l Ledger) NewView() *delta.View {
	return delta.NewView(func(owner, controller, key string) ([]byte, error) {
		id := flow.RegisterID{
			Owner:      owner,
			Controller: controller,
			Key:        key,
		}
		return l[id.String()].Value, nil
	})
}

func (l Ledger) ApplyDelta(delta delta.Delta) {
	for id, value := range delta.Data {
		l[id] = value
		// TODO: support delete
	}
}
