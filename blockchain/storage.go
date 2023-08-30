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

import (
	"errors"
	"fmt"
	"github.com/onflow/cadence"
	"strings"
)

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

type AccountStorage struct {
	storage []StorageItem
	public  []StorageItem
	private []StorageItem
}

type StorageItem struct {
	Value string
	Type  string
	Path  string
}

// ParseAccountStorage parses the account storage returned by the StorageIteration script
// and returns the storage as a list of StorageItem
func ParseAccountStorage(rawStorage cadence.Value) (storage *AccountStorage, err error) {
	defer func() {
		r := recover()
		if r != nil {
			storage = nil
			err = errors.New("failed to parse account storage")
		}
	}()
	storage = &AccountStorage{}

	// Storage item parts
	const (
		ValuePrefix = `value: `
		PathPrefix  = `path: `
		TypePrefix  = `type: `
	)

	items := strings.Split(rawStorage.String(), "},")
	for _, item := range items {
		storageItem := StorageItem{}
		item = strings.TrimPrefix(item, "[")
		item = strings.TrimPrefix(item, "{")

		// Extract parts of value, path, type for current item
		prevPart := ""
		itemParts := strings.Split(item, ",")
		for _, part := range itemParts {
			part = strings.TrimPrefix(part, " ")
			part = strings.TrimPrefix(part, "{")
			part = strings.TrimSuffix(part, "}]")
			part = strings.ReplaceAll(part, `"`, ``)

			if strings.HasPrefix(part, ValuePrefix) {
				prevPart = ValuePrefix
				storageItem.Value = strings.TrimPrefix(part, ValuePrefix)
			} else if strings.HasPrefix(part, PathPrefix) {
				prevPart = PathPrefix
				storageItem.Path = strings.TrimPrefix(part, PathPrefix)
			} else if strings.HasPrefix(part, TypePrefix) {
				prevPart = TypePrefix
				storageItem.Type = strings.TrimPrefix(part, TypePrefix)
			} else {
				// Add to previous part
				if prevPart == ValuePrefix {
					storageItem.Value += `, ` + part
				} else if prevPart == PathPrefix {
					storageItem.Path += `, ` + part
				} else if prevPart == TypePrefix {
					storageItem.Type += `, ` + part
				} else {
					// Shouldn't happen
					continue
				}
			}
		}

		if strings.HasPrefix(storageItem.Path, `/public/`) {
			storage.public = append(storage.public, storageItem)
		} else if strings.HasPrefix(storageItem.Path, `/private/`) {
			storage.private = append(storage.private, storageItem)
		} else {
			storage.storage = append(storage.storage, storageItem)
		}
	}

	return storage, nil
}

func StorageItemListToJson(itemList []StorageItem) string {
	json := ``
	for i, item := range itemList {
		if i != 0 {
			json += `,`
		}
		json += item.ToJsonString()
	}
	return json
}

func (storage *AccountStorage) ToJsonString() string {
	jsonStorage := StorageItemListToJson(storage.storage)
	jsonPublic := StorageItemListToJson(storage.public)
	jsonPrivate := StorageItemListToJson(storage.private)

	return fmt.Sprintf(`{"storage":[%s], "public":[%s], "private":[%s]}`,
		jsonStorage,
		jsonPublic,
		jsonPrivate,
	)
}

func (item *StorageItem) ToJsonString() string {
	return fmt.Sprintf(
		`{"value":"%s", "type":"%s", "path":"%s"}`,
		item.Value,
		item.Type,
		item.Path,
	)
}
