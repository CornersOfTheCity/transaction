package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func WriteLedger(obj interface{}, stub shim.ChaincodeStubInterface, objectType string, keys []string) error {
	var key string
	// 账本数据是以Key-Value形式进行存储的。一般使用主键ID（唯一）为Key，数据为Value进行存储。
	// 如用户数据，Key为ID，User为Value。这时若有需求需要使用姓名来查询用户，由于姓名可能会重复，账本数据就会被覆盖。
	// 因为一个Key只能存储条数据，相同的Key只会将之前的那条数据覆盖掉。复合键在存入账本后会使用字典序来排序。
	if val, err := stub.CreateCompositeKey(objectType, keys); err != nil {
		return errors.New(fmt.Sprintf("%s-创建复合主键出错 %s", objectType, err))
	} else {
		key = val
	}

	bytes, err := json.Marshal(obj)
	if err != nil {
		return errors.New(fmt.Sprintf("%s-序列化json数据失败出错: %s", objectType, err))
	}

	if err := stub.PutState(key, bytes); err != nil {
		return errors.New(fmt.Sprintf("%s-写入区块链账本出错: %s", objectType, err))
	}
	return nil
}

func DelLedger(stub shim.ChaincodeStubInterface, objectType string, keys []string) error {
	var key string
	if val, err := stub.CreateCompositeKey(objectType, keys); err != nil {
		return errors.New(fmt.Sprintf("%s-创建复合主键出错 %s", objectType, err))
	} else {
		key = val
	}

	if err := stub.DelState(key); err != nil {
		return errors.New(fmt.Sprintf("%s-删除区块链账本出错: %s", objectType, err))
	}
	return nil
}

func GetStateByPartialCompositeKeys(stub shim.ChaincodeStubInterface, objectType string, keys []string) (results [][]byte, err error) {
	if len(keys) == 0 {
		//GetStateByPartialCompositeKey方法获取有keys的集合迭代器
		resultIterator, err := stub.GetStateByPartialCompositeKey(objectType, keys)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s-获取全部数据出错: %s", objectType, err))
		}
		defer resultIterator.Close()

		//遍历
		for resultIterator.HasNext() {
			val, err := resultIterator.Next()
			if err != nil {
				return nil, errors.New(fmt.Sprintf("%s-返回的数据出错: %s", objectType, err))
			}
			results = append(results, val.GetValue())
		}
	} else {
		for _, v := range keys {
			key, err := stub.CreateCompositeKey(objectType, []string{v})
			if err != nil {
				return nil, errors.New(fmt.Sprintf("%s-创建组合键出错: %s", objectType, err))
			}

			bytes, err := stub.GetState(key)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("%s-获取数据出错: %s", objectType, err))
			}

			if bytes != nil {
				results = append(results, bytes)
			}
		}
	}
	return results, nil
}

func GetStateByPartialCompositeKeys2(stub shim.ChaincodeStubInterface, objectType string, keys []string) (results [][]byte, err error) {
	resultIterator, err := stub.GetStateByPartialCompositeKey(objectType, keys)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s-获取全部数据出错: %s", objectType, err))
	}
	defer resultIterator.Close()

	for resultIterator.HasNext() {
		val, err := resultIterator.Next()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s-返回的数据出错: %s", objectType, err))
		}
		results = append(results, val.GetValue())
	}
	return results, nil
}
