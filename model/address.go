package model

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const addressLength = 20

type Address [addressLength]byte

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
