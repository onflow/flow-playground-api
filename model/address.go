package model

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/pkg/errors"
)

const addressLength = 20

type Address [addressLength]byte

func NewAddressFromBytes(b []byte) Address {
	var address Address
	copy(address[addressLength-len(b):], b[:])
	return address
}

func (a *Address) ToFlowAddress() flow.Address {
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
	io.WriteString(w, str)
}
