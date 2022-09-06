/*
 * Flow Playground
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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
	b = shiftAddressFromFlow(b)
	var address Address
	copy(address[addressLength-len(b):], b[:])
	return address
}

func (a *Address) ToFlowAddress() flow.Address {
	addr := shiftAddressToFlow(a[:])
	return flow.BytesToAddress(addr[len(addr)-flow.AddressLength:])
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

const numberOfAccounts = 4

// shiftAddressToFlow adds numberOfAccounts to the address since it was provided by the user
// and was previously shifted by shiftAddressFromFlow.
func shiftAddressToFlow(a []byte) []byte {
	var b [8]byte // create a copy
	copy(b[:], a[:])
	b[len(b)-1] = b[len(b)-1] + numberOfAccounts
	return b[:]
}

// shiftAddressFromFlow subtracts numberOfAccounts that were created during
// bootstrap automatically by emulator, so the user see the numberOfAccounts+1 as
// the first account
func shiftAddressFromFlow(a []byte) []byte {
	var b [8]byte
	copy(b[:], a[:])
	if b[len(b)-1] < numberOfAccounts { // ignore service account conversion
		return b[:]
	}
	b[len(b)-1] = b[len(b)-1] - numberOfAccounts
	return b[:]
}
