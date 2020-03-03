package encoding

import (
	"fmt"

	"github.com/dapperlabs/flow-go/language"
)

type Value interface{}

type Literal string

func (s Literal) String() string {
	return string(s)
}

type AtomicValue struct {
	Type  string
	Value Value
}

type ListValue []Value

type CompositeValue map[string]Value

func Encode(value language.Value) Value {
	switch v := value.(type) {
	case language.Void:
		return AtomicValue{
			Type:  "Void",
			Value: Literal("void"),
		}
	case language.Nil:
		return AtomicValue{
			Type:  "Nil",
			Value: Literal("nil"),
		}
	case language.Bool:
		return AtomicValue{
			Type:  "Bool",
			Value: Literal(fmt.Sprintf("%s", v.ToGoValue())),
		}
	case language.String:
		return AtomicValue{
			Type:  "String",
			Value: Literal(fmt.Sprintf("%s", v.ToGoValue())),
		}
	case language.VariableSizedArray:
		return encodeArray(v.Values)
	case language.ConstantSizedArray:
		return encodeArray(v.Values)
	case language.Int:
		return AtomicValue{
			Type:  "Int",
			Value: Literal(fmt.Sprintf("%d", v.ToGoValue())),
		}
	case language.Int8:
		return AtomicValue{
			Type:  "UInt64",
			Value: Literal(fmt.Sprintf("%d", v.ToGoValue())),
		}
	case language.Int16:
		return AtomicValue{
			Type:  "UInt64",
			Value: Literal(fmt.Sprintf("%d", v.ToGoValue())),
		}
	case language.Int32:
		return AtomicValue{
			Type:  "UInt64",
			Value: Literal(fmt.Sprintf("%d", v.ToGoValue())),
		}
	case language.Int64:
		return AtomicValue{
			Type:  "UInt64",
			Value: Literal(fmt.Sprintf("%d", v.ToGoValue())),
		}
	case language.UInt8:
		return AtomicValue{
			Type:  "UInt64",
			Value: Literal(fmt.Sprintf("%d", v.ToGoValue())),
		}
	case language.UInt16:
		return AtomicValue{
			Type:  "UInt64",
			Value: Literal(fmt.Sprintf("%d", v.ToGoValue())),
		}
	case language.UInt32:
		return AtomicValue{
			Type:  "UInt64",
			Value: Literal(fmt.Sprintf("%d", v.ToGoValue())),
		}
	case language.UInt64:
		return AtomicValue{
			Type:  "UInt64",
			Value: Literal(fmt.Sprintf("%d", v.ToGoValue())),
		}
	case language.Composite:
		// fields := v.Type().(language.CompositeType).Fields
		// return encodeComposite(fields, v.Fields)
		return encodeArray(v.Fields)
	case language.Dictionary:
		return encodeDictionary(v.Pairs)
	case language.Address:
		return AtomicValue{
			Type:  "Address",
			Value: Literal(fmt.Sprintf("%x", v.ToGoValue())),
		}
	}

	return Literal("UNKNOWN")
}

func encodeArray(values []language.Value) ListValue {
	l := make(ListValue, len(values))

	for i, v := range values {
		l[i] = Encode(v)
	}

	return l
}

func encodeComposite(fields []language.Field, values []language.Value) CompositeValue {
	c := make(CompositeValue)

	for i, value := range values {
		keyStr := fields[i].Identifier
		c[keyStr] = Encode(value)
	}

	return c
}

func encodeDictionary(pairs []language.KeyValuePair) CompositeValue {
	c := make(CompositeValue)

	for _, pair := range pairs {
		keyStr := fmt.Sprintf("%s", pair.Key.ToGoValue())
		c[keyStr] = Encode(pair.Value)
	}

	return c
}
