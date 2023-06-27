package osmoutils

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/gogo/protobuf/proto"
)

var (
	ErrNoValuesInRange = errors.New("No values in range")
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
	return gatherValuesFromIterator(iterator, parseValue, noStopFn)
}

// GatherValuesFromStorePrefix is a decorator around GatherValuesFromStorePrefixWithKeyParser. It overwrites the parse function to
// disable parsing keys, only keeping values
func GatherValuesFromStorePrefix[T any](storeObj store.KVStore, prefix []byte, parseValue func([]byte) (T, error)) ([]T, error) {
	// Replace a callback with the one that takes both key and value
	// but ignores the key.
	parseOnlyValue := func(_ []byte, value []byte) (T, error) {
		return parseValue(value)
	}
	return GatherValuesFromStorePrefixWithKeyParser(storeObj, prefix, parseOnlyValue)
}

// GatherValuesFromStorePrefixWithKeyParser is a helper function that gathers values from a given store prefix. While iterating through
// the entries, it parses both key and the value using the provided parse function to return the desired type.
// Returns error if:
// - the parse function returns an error.
// - internal database error
func GatherValuesFromStorePrefixWithKeyParser[T any](storeObj store.KVStore, prefix []byte, parse func(key []byte, value []byte) (T, error)) ([]T, error) {
	iterator := sdk.KVStorePrefixIterator(storeObj, prefix)
	defer iterator.Close()
	return gatherValuesFromIteratorWithKeyParser(iterator, parse, noStopFn)
}

func GetValuesUntilDerivedStop[T any](storeObj store.KVStore, keyStart []byte, stopFn func([]byte) bool, parseValue func([]byte) (T, error)) ([]T, error) {
	// SDK iterator is broken for nil end time, and non-nil start time
	// https://github.com/cosmos/cosmos-sdk/issues/12661
	// hence we use []byte{0xff}
	keyEnd := []byte{0xff}
	return GetIterValuesWithStop(storeObj, keyStart, keyEnd, false, stopFn, parseValue)
}

func makeIterator(storeObj store.KVStore, keyStart []byte, keyEnd []byte, reverse bool) store.Iterator {
	if reverse {
		return storeObj.ReverseIterator(keyStart, keyEnd)
	}
	return storeObj.Iterator(keyStart, keyEnd)
}

func GetIterValuesWithStop[T any](
	storeObj store.KVStore,
	keyStart []byte,
	keyEnd []byte,
	reverse bool,
	stopFn func([]byte) bool,
	parseValue func([]byte) (T, error),
) ([]T, error) {
	iter := makeIterator(storeObj, keyStart, keyEnd, reverse)
	defer iter.Close()

	return gatherValuesFromIterator(iter, parseValue, stopFn)
}

// HasAnyAtPrefix returns true if there is at least one value in the given prefix.
func HasAnyAtPrefix[T any](storeObj store.KVStore, prefix []byte, parseValue func([]byte) (T, error)) (bool, error) {
	_, err := GetFirstValueInRange(storeObj, prefix, sdk.PrefixEndBytes(prefix), false, parseValue)
	if err != nil {
		if err == ErrNoValuesInRange {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func GetFirstValueAfterPrefixInclusive[T any](storeObj store.KVStore, keyStart []byte, parseValue func([]byte) (T, error)) (T, error) {
	// SDK iterator is broken for nil end time, and non-nil start time
	// https://github.com/cosmos/cosmos-sdk/issues/12661
	// hence we use []byte{0xff}
	return GetFirstValueInRange(storeObj, keyStart, []byte{0xff}, false, parseValue)
}

func GetFirstValueInRange[T any](storeObj store.KVStore, keyStart []byte, keyEnd []byte, reverseIterate bool, parseValue func([]byte) (T, error)) (T, error) {
	iterator := makeIterator(storeObj, keyStart, keyEnd, reverseIterate)
	defer iterator.Close()

	if !iterator.Valid() {
		var blankValue T
		return blankValue, ErrNoValuesInRange
	}

	return parseValue(iterator.Value())
}

func gatherValuesFromIterator[T any](iterator db.Iterator, parseValue func([]byte) (T, error), stopFn func([]byte) bool) ([]T, error) {
	// Replace a callback with the one that takes both key and value
	// but ignores the key.
	parseKeyValue := func(_ []byte, value []byte) (T, error) {
		return parseValue(value)
	}
	return gatherValuesFromIteratorWithKeyParser(iterator, parseKeyValue, stopFn)
}

func gatherValuesFromIteratorWithKeyParser[T any](iterator db.Iterator, parse func(key []byte, value []byte) (T, error), stopFn func([]byte) bool) ([]T, error) {
	values := []T{}
	for ; iterator.Valid(); iterator.Next() {
		if stopFn(iterator.Key()) {
			break
		}
		val, err := parse(iterator.Key(), iterator.Value())
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}
	return values, nil
}

func noStopFn([]byte) bool {
	return false
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

// MustGet gets key from store by mutating result
// Panics on any error.
func MustGet(store store.KVStore, key []byte, result proto.Message) {
	b := store.Get(key)
	if b == nil {
		panic(fmt.Errorf("getting at key (%v) should not have been nil", key))
	}
	if err := proto.Unmarshal(b, result); err != nil {
		panic(err)
	}
}

// MustSetDec sets dec value to store at key. Panics on any error.
func MustSetDec(store store.KVStore, key []byte, value sdk.Dec) {
	MustSet(store, key, &sdk.DecProto{
		Dec: value,
	})
}

// MustGetDec gets dec value from store at key. Panics on any error.
func MustGetDec(store store.KVStore, key []byte) sdk.Dec {
	result := &sdk.DecProto{}
	MustGet(store, key, result)
	return result.Dec
}

// GetDec gets dec value from store at key. Returns error if:
// - database error occurs.
// - no value at given key is found.
func GetDec(store store.KVStore, key []byte) (sdk.Dec, error) {
	result := &sdk.DecProto{}
	isFound, err := Get(store, key, result)
	if err != nil {
		return sdk.Dec{}, err
	}
	if !isFound {
		return sdk.Dec{}, DecNotFoundError{Key: string(key)}
	}
	return result.Dec, nil
}

// Get returns a value at key by mutating the result parameter. Returns true if the value was found and the
// result mutated correctly. If the value is not in the store, returns false.
// Returns error only when database or serialization errors occur. (And when an error occurs, returns false)
func Get(store store.KVStore, key []byte, result proto.Message) (found bool, err error) {
	b := store.Get(key)
	if b == nil {
		return false, nil
	}
	if err := proto.Unmarshal(b, result); err != nil {
		return true, err
	}
	return true, nil
}

// DeleteAllKeysFromPrefix deletes all store records that contains the given prefixKey.
func DeleteAllKeysFromPrefix(ctx sdk.Context, store store.KVStore, prefixKey []byte) {
	prefixStore := prefix.NewStore(store, prefixKey)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		prefixStore.Delete(iter.Key())
	}
}
