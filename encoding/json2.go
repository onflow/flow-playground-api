package encoding

import (
	"fmt"

	"github.com/dapperlabs/flow-go/language/runtime"
	"github.com/dapperlabs/flow-go/language/runtime/interpreter"
)

type Value interface{}

// ConvertValue converts a runtime value to its corresponding Go representation.
func ConvertValue(value runtime.Value) (Value, error) {
	switch v := value.(type) {
	case interpreter.VoidValue:
		return "void", nil

	case interpreter.NilValue:
		return "nil", nil

	case *interpreter.SomeValue:
		return ConvertValue(v.Value)

	case interpreter.BoolValue:
		return fmt.Sprintf("%v", v), nil

	case *interpreter.StringValue:
		return fmt.Sprintf("%s", v.Str), nil

	case *interpreter.ArrayValue:
		return convertArrayValue(v)

	case interpreter.IntValue:
		return fmt.Sprintf("%s", v), nil

	case interpreter.UIntValue:
		return fmt.Sprintf("%s", v), nil

	case interpreter.Int8Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.Int16Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.Int32Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.Int64Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.Int128Value:
		return fmt.Sprintf("%s", v), nil

	case interpreter.Int256Value:
		return fmt.Sprintf("%s", v), nil

	case interpreter.UInt8Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.UInt16Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.UInt32Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.UInt64Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.UInt128Value:
		return fmt.Sprintf("%s", v), nil

	case interpreter.UInt256Value:
		return fmt.Sprintf("%s", v), nil

	case interpreter.Word8Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.Word16Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.Word32Value:
		return fmt.Sprintf("%d", v), nil

	case interpreter.Word64Value:
		return fmt.Sprintf("%d", v), nil

	case *interpreter.CompositeValue:
		return convertCompositeValue(v)

	case *interpreter.DictionaryValue:
		return convertDictionaryValue(v)

	case interpreter.AddressValue:
		return v.Hex(), nil

	case *interpreter.StorageReferenceValue:
		return fmt.Sprintf("Ref(%s, %s)", v.TargetStorageAddress, v.TargetKey), nil
	}

	return nil, fmt.Errorf("cannot convert value of type %T", value)
}

func convertArrayValue(v *interpreter.ArrayValue) (Value, error) {
	vals := make([]Value, len(v.Values))

	for i, value := range v.Values {
		convertedValue, err := ConvertValue(value)
		if err != nil {
			return nil, err
		}

		vals[i] = convertedValue
	}

	return vals, nil
}

func convertCompositeValue(v *interpreter.CompositeValue) (Value, error) {
	f := make(map[string]Value)

	for key, value := range v.Fields {
		v, _ := ConvertValue(value)
		f[key] = v
	}

	return f, nil
}

func convertDictionaryValue(v *interpreter.DictionaryValue) (Value, error) {
	f := make(map[string]Value)

	for key, value := range v.Entries {
		v, _ := ConvertValue(value)
		f[key] = v
	}

	return f, nil
}
