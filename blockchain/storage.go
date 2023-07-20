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
