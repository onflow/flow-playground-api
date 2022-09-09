package blockchain

import (
	"testing"

	"github.com/dapperlabs/flow-playground-api/model"

	"github.com/stretchr/testify/assert"
)

func Test_TranslateAddress(t *testing.T) {
	assert.Equal(t, NumberOfServiceAccounts, model.NumberOfServiceAccounts) // avoid circular deps

	inputs := [][][]byte{{
		[]byte(`
			import Foo from 0x02
			pub contract Bar {}
		`), []byte(`
			import Foo from 0x0000000000000006
			pub contract Bar {}
		`),
	}, {
		[]byte(`
			import Zoo from 0x0000000000000001
			pub fun main() {}
		`), []byte(`
			import Zoo from 0x0000000000000005
			pub fun main() {}
		`),
	}, {
		[]byte(`
			import Crypto
			pub fun main() {}
		`), []byte(`
			import Crypto
			pub fun main() {}
		`),
	}}

	for _, in := range inputs {
		out := translateAddresses(in[0])
		assert.Equal(t, string(in[1]), string(out))
	}

}
