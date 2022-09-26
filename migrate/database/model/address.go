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

package model

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
)

const addressLength = 8

type Address [addressLength]byte

func NewAddressFromBytes(a []byte) Address {
	var b Address
	copy(b[addressLength-len(a):], a[:])
	return b
}

func NewAddressFromString(address string) Address {
	addr := flow.HexToAddress(address)
	var newAddress Address
	copy(newAddress[:], addr[:])
	return newAddress
}

func (a Address) ToFlowAddress() flow.Address {
	return flow.BytesToAddress(a[len(a)-flow.AddressLength:])
}

func (a *Address) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("addresses must be hex strings")
	}

	b, err := hex.DecodeString(str)
	if err != nil {
		return errors.Wrap(err, "failed to decode hex string")
	}

	if len(b) != addressLength {
		return fmt.Errorf("addresses must be %d bytes", addressLength)
	}

	copy(a[:], b[:])

	return nil
}

func (a Address) MarshalGQL(w io.Writer) {
	str := fmt.Sprintf("\"%x\"", a)
	_, _ = io.WriteString(w, str)
}
