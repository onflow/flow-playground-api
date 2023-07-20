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

package blockchain

type AccountStorage any

type StorageItem struct {
	Value interface{}
	Type  interface{}
	Path  interface{}
}

const StorageIteration = `
pub fun main(address: Address) : AnyStruct{

	var res :  [{String:AnyStruct}] = []

	getAuthAccount(address).forEachStored(fun (path: StoragePath, type: Type): Bool {
		res.append(
		{
			"path" : path,
			"type" : type.identifier,
			"value":  type.isSubtype(of: Type<AnyStruct>()) ?
							getAuthAccount(address).borrow<&AnyStruct>(from: path)! as AnyStruct
							: getAuthAccount(address).borrow<&AnyResource>(from: path)! as AnyStruct
		})
		return true
	})

	getAuthAccount(address).forEachPublic(fun (path: PublicPath, type: Type): Bool {
		res.append(
		{
			"path" : path,
			"type" : type.identifier,
			"value":  getAuthAccount(address).getLinkTarget(path)
		})
		return true
	})

	getAuthAccount(address).forEachPrivate(fun (path: PrivatePath, type: Type): Bool {
		res.append(
		{
			"path" : path,
			"type" : type.identifier,
			"value":  getAuthAccount(address).getLinkTarget(path)
		})
		return true
	})
	return res
}`

//type StorageItem map[interface{}]interface{}

/* TODO: Parse account storage into a useful format or structure
func ParseAccountStorage(rawStorage cadence.Value) (*AccountStorage, error) {
	encoded, err := jsoncdc.Encode(rawStorage)
	if err != nil {
		return nil, err
	}

	var storage AccountStorage
	err = yaml.Unmarshal(encoded, &storage)
	if err != nil {
		fmt.Println("ERROR Unmarshal", err.Error())
	}

	for key, val := range storage.(map[interface{}]interface{}) {
		fmt.Println("Key, val:", key, ",", val)
	}

	return nil, nil
}
*/
