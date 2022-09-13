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

func NewAddressFromBytes(b []byte) Address {
	return shiftAddressFromFlow(b)
}

func NewAddressFromString(address string) Address {
	addr := flow.HexToAddress(address)
	var newAddress Address
	copy(newAddress[:], addr[:])
	return newAddress
}

func (a Address) ToFlowAddress() flow.Address {
	addr := shiftAddressToFlow(a)
	return flow.BytesToAddress(addr[len(addr)-flow.AddressLength:])
}

func (a Address) ToFlowAddressWithoutTranslation() flow.Address {
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

const NumberOfServiceAccounts = 4

// shiftAddressToFlow adds numberOfAccounts to the address since it was provided by the user
// and was previously shifted by shiftAddressFromFlow.
func shiftAddressToFlow(address Address) Address {
	var b Address // create a copy
	copy(b[:], address[:])
	b[len(b)-1] = b[len(b)-1] + NumberOfServiceAccounts
	return b
}

// shiftAddressFromFlow subtracts numberOfAccounts that were created during
// bootstrap automatically by emulator, so the user see the numberOfAccounts+1 as
// the first account
func shiftAddressFromFlow(a []byte) Address {
	var b Address
	copy(b[addressLength-len(a):], a[:])
	if b[len(b)-1] < NumberOfServiceAccounts { // ignore service account conversion
		return b
	}
	b[len(b)-1] = b[len(b)-1] - NumberOfServiceAccounts
	return b
}
