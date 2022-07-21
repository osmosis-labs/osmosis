package osmoutils

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/gogo/protobuf/proto"
)

func GatherAllKeysFromStore(storeObj store.KVStore) []string {
	iterator := storeObj.Iterator(nil, nil)
	defer iterator.Close()

	keys := []string{}
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, string(iterator.Key()))
	}
	return keys
}

func GatherValuesFromStore[T any](storeObj store.KVStore, keyStart []byte, keyEnd []byte, parseValue func([]byte) (T, error)) ([]T, error) {
	iterator := storeObj.Iterator(keyStart, keyEnd)
	defer iterator.Close()

	values := []T{}
	for ; iterator.Valid(); iterator.Next() {
		val, err := parseValue(iterator.Value())
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}
	return values, nil
}

func GetValuesUntilDerivedStop[T any](storeObj store.KVStore, keyStart []byte, stopFn func([]byte) bool, parseValue func([]byte) (T, error)) ([]T, error) {
	// SDK iterator is broken for nil end time, and non-nil start time
	// https://github.com/cosmos/cosmos-sdk/issues/12661
	// hence we use []byte{0xff}
	iterator := storeObj.Iterator(keyStart, []byte{0xff})
	defer iterator.Close()

	values := []T{}
	fmt.Println(iterator.Valid())
	for ; iterator.Valid(); iterator.Next() {
		if stopFn(iterator.Key()) {
			break
		}
		val, err := parseValue(iterator.Value())
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}
	return values, nil
}

func GetFirstValueAfterPrefix[T any](storeObj store.KVStore, keyStart []byte, parseValue func([]byte) (T, error)) (T, error) {
	// SDK iterator is broken for nil end time, and non-nil start time
	// https://github.com/cosmos/cosmos-sdk/issues/12661
	// hence we use []byte{0xff}
	iterator := storeObj.Iterator(keyStart, []byte{0xff})
	defer iterator.Close()

	if !iterator.Valid() {
		var blankValue T
		return blankValue, errors.New("No values in iterator")
	}

	return parseValue(iterator.Value())
}

// MustSet runs store.Set(key, proto.Marshal(value))
// but panics on any error.
func MustSet(storeObj store.KVStore, key []byte, value proto.Message) {
	bz, err := proto.Marshal(value)
	if err != nil {
		panic(err)
	}

	storeObj.Set(key, bz)
}
