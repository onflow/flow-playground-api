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

const StorageIteration = `
access(all) fun main(address: Address) : AnyStruct{

	var res :  [{String:AnyStruct}] = []

	getAuthAccount<auth(Storage) &Account>(address).storage.forEachStored(fun (path: StoragePath, type: Type): Bool {
		res.append(
		{
			"path" : path, 
			"type" : type.identifier,
			"value":  type.isSubtype(of: Type<AnyStruct>()) ?
							getAuthAccount<auth(Storage) &Account>(address).storage.borrow<&AnyStruct>(from: path)! as AnyStruct
							: getAuthAccount<auth(Storage) &Account>(address).storage.borrow<&AnyResource>(from: path)! as AnyStruct
		})
		return true
	})

	


	return res
}`
