/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package controller

import (
	"fmt"
	"testing"

	"github.com/dapperlabs/flow-playground-api/model"

	"github.com/stretchr/testify/assert"
)

func Test_ContentAdapter(t *testing.T) {
	testVectors := []struct {
		in        string
		fromInput bool
		out       string
	}{{
		in:        "0x01",
		out:       "0x05",
		fromInput: true,
	}, {
		in:        "0x0000000000000001",
		out:       "0x0000000000000005",
		fromInput: true,
	}, {
		in:        "0x1",
		out:       "0x5",
		fromInput: true,
	}, {
		in: `
			import Foo from 0x01
			import Zoo from 0x02
			import Goo from 0x03
			pub struct Bar {}
		`,
		out: `
			import Foo from 0x05
			import Zoo from 0x06
			import Goo from 0x07
			pub struct Bar {}
		`,
		fromInput: true,
	}, {
		in:        "0x05",
		out:       "0x01",
		fromInput: false,
	}, { // don't convert service addresses
		in:        "0x01",
		out:       "0x01",
		fromInput: false,
	}}

	for _, vector := range testVectors {
		out := contentAdapter(vector.in, vector.fromInput)
		assert.Equal(t, vector.out, out, fmt.Sprintf("problem with input %v", vector))
	}
}

func Test_AddressAdapter(t *testing.T) {
	t.Run("adapt from input", func(t *testing.T) {
		testVectors := [][]model.Address{
			{model.NewAddressFromString("0x01"), model.NewAddressFromString("0x05")},
			{model.NewAddressFromString("0x03"), model.NewAddressFromString("0x07")},
		}

		for _, vector := range testVectors {
			out := addressFromInput(vector[0])
			assert.Equal(t, vector[1], out)
		}
	})

	t.Run("adapt to output", func(t *testing.T) {
		testVectors := [][]model.Address{
			{model.NewAddressFromString("0x05"), model.NewAddressFromString("0x01")},
			{model.NewAddressFromString("0x07"), model.NewAddressFromString("0x03")},
		}

		for _, vector := range testVectors {
			out := addressToOutput(vector[0])
			assert.Equal(t, vector[1], out)
		}
	})
}
