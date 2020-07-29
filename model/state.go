package model

import (
	"encoding/json"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
)

type AccountState map[string]cadence.Value

func (a AccountState) MarshalJSON() ([]byte, error) {
	m := make(serializableAccountState, len(a))

	for key, value := range a {
		m[key] = serializableCadenceValue{Value: value}
	}

	return json.Marshal(m)
}

func (a *AccountState) UnmarshalJSON(data []byte) error {
	*a = make(AccountState)

	m := make(serializableAccountState)

	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	for key, value := range m {
		(*a)[key] = value.Value
	}

	return nil
}

type serializableCadenceValue struct {
	cadence.Value
}

func (v serializableCadenceValue) MarshalJSON() ([]byte, error) {
	if v.Value == nil {
		return json.Marshal(nil)
	}

	return jsoncdc.Encode(v.Value)
}

func (v *serializableCadenceValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	value, err := jsoncdc.Decode(data)
	if err != nil {
		return err
	}

	v.Value = value

	return nil
}

type serializableAccountState map[string]serializableCadenceValue
