package blockchain

import (
	"fmt"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"gopkg.in/yaml.v2"
	"reflect"
)

type AccountStorage any

type StorageItem struct {
	Value interface{}
	Type  interface{}
	Path  interface{}
}

//type StorageItem map[interface{}]interface{}

func ParseAccountStorage(rawStorage cadence.Value) (*AccountStorage, error) {
	fmt.Println("Storage:", rawStorage.String())

	encoded, err := jsoncdc.Encode(rawStorage)
	if err != nil {
		fmt.Println("ERROR encoding", err.Error())
	}

	var storage AccountStorage
	err = yaml.Unmarshal(encoded, &storage)
	if err != nil {
		fmt.Println("ERROR Unmarshal", err.Error())
	}

	fmt.Println("Storage:", storage)
	fmt.Println(reflect.TypeOf(storage))

	for key, val := range storage.(map[interface{}]interface{}) {
		fmt.Println("Key, val:", key, ",", val)
	}

	//fmt.Println("Unmarshalled val", item[0])
	//fmt.Println("Unmarshalled path", item[0]["path"])
	//fmt.Println("Unmarshalled type", item[0]["type"])

	return nil, nil
}
