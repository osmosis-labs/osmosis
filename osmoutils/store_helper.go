package osmoutils

import (
	"bytes"
	"errors"
	"fmt"

	"cosmossdk.io/store/prefix"
	db "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/osmomath"

	"cosmossdk.io/store"
	"github.com/cosmos/gogoproto/proto"

	lru "github.com/hashicorp/golang-lru"
)

var (
	ErrNoValuesInRange = errors.New("No values in range")
)

func GatherAllKeysFromStore(storeObj storetypes.KVStore) []string {
	iterator := storeObj.Iterator(nil, nil)
	defer iterator.Close()

	keys := []string{}
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, string(iterator.Key()))
	}
	return keys
}

func GatherValuesFromStore[T any](storeObj storetypes.KVStore, keyStart []byte, keyEnd []byte, parseValue func([]byte) (T, error)) ([]T, error) {
	iterator := storeObj.Iterator(keyStart, keyEnd)
	defer iterator.Close()
	return gatherValuesFromIterator(iterator, parseValue, noStopFn)
}

// GatherValuesFromStorePrefix is a decorator around GatherValuesFromStorePrefixWithKeyParser. It overwrites the parse function to
// disable parsing keys, only keeping values
func GatherValuesFromStorePrefix[T any](storeObj storetypes.KVStore, prefix []byte, parseValue func([]byte) (T, error)) ([]T, error) {
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
func GatherValuesFromStorePrefixWithKeyParser[T any](storeObj storetypes.KVStore, prefix []byte, parse func(key []byte, value []byte) (T, error)) ([]T, error) {
	iterator := storetypes.KVStorePrefixIterator(storeObj, prefix)
	defer iterator.Close()
	return gatherValuesFromIteratorWithKeyParser(iterator, parse, noStopFn)
}

func GetValuesUntilDerivedStop[T any](storeObj storetypes.KVStore, keyStart []byte, stopFn func([]byte) bool, parseValue func([]byte) (T, error)) ([]T, error) {
	// SDK iterator is broken for nil end time, and non-nil start time
	// https://github.com/cosmos/cosmos-sdk/issues/12661
	// hence we use []byte{0xff}
	keyEnd := []byte{0xff}
	return GetIterValuesWithStop(storeObj, keyStart, keyEnd, false, stopFn, parseValue)
}

func makeIterator(storeObj storetypes.KVStore, keyStart []byte, keyEnd []byte, reverse bool) store.Iterator {
	if reverse {
		return storeObj.ReverseIterator(keyStart, keyEnd)
	}
	return storeObj.Iterator(keyStart, keyEnd)
}

func GetIterValuesWithStop[T any](
	storeObj storetypes.KVStore,
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
func HasAnyAtPrefix[T any](storeObj storetypes.KVStore, prefix []byte, parseValue func([]byte) (T, error)) (bool, error) {
	_, err := GetFirstValueInRange(storeObj, prefix, storetypes.PrefixEndBytes(prefix), false, parseValue)
	if err != nil {
		if err == ErrNoValuesInRange {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func GetFirstValueAfterPrefixInclusive[T any](storeObj storetypes.KVStore, keyStart []byte, parseValue func([]byte) (T, error)) (T, error) {
	// SDK iterator is broken for nil end time, and non-nil start time
	// https://github.com/cosmos/cosmos-sdk/issues/12661
	// hence we use []byte{0xff}
	return GetFirstValueInRange(storeObj, keyStart, []byte{0xff}, false, parseValue)
}

func GetFirstValueInRange[T any](storeObj storetypes.KVStore, keyStart []byte, keyEnd []byte, reverseIterate bool, parseValue func([]byte) (T, error)) (T, error) {
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
func MustSet(store storetypes.KVStore, key []byte, value proto.Message) {
	bz, err := proto.Marshal(value)
	if err != nil {
		panic(err)
	}

	store.Set(key, bz)
}

// MustGet gets key from store by mutating result
// Panics on any error.
func MustGet(store storetypes.KVStore, key []byte, result proto.Message) {
	b := store.Get(key)
	if b == nil {
		panic(fmt.Errorf("getting at key (%v) should not have been nil", key))
	}
	if err := proto.Unmarshal(b, result); err != nil {
		panic(err)
	}
}

// MustSetDec sets dec value to store at key. Panics on any error.
func MustSetDec(store storetypes.KVStore, key []byte, value osmomath.Dec) {
	MustSet(store, key, &sdk.DecProto{
		Dec: value,
	})
}

// MustGetDec gets dec value from store at key. Panics on any error.
func MustGetDec(store storetypes.KVStore, key []byte) osmomath.Dec {
	result := &sdk.DecProto{}
	MustGet(store, key, result)
	return result.Dec
}

// GetDec gets dec value from store at key. Returns error if:
// - database error occurs.
// - no value at given key is found.
func GetDec(store storetypes.KVStore, key []byte) (osmomath.Dec, error) {
	result := &sdk.DecProto{}
	isFound, err := Get(store, key, result)
	if err != nil {
		return osmomath.Dec{}, err
	}
	if !isFound {
		return osmomath.Dec{}, DecNotFoundError{Key: string(key)}
	}
	return result.Dec, nil
}

// Get returns a value at key by mutating the result parameter. Returns true if the value was found and the
// result mutated correctly. If the value is not in the store, returns false.
// Returns error only when database or serialization errors occur. (And when an error occurs, returns false)
func Get(store storetypes.KVStore, key []byte, result proto.Message) (found bool, err error) {
	b := store.Get(key)
	if b == nil {
		return false, nil
	}
	if err := proto.Unmarshal(b, result); err != nil {
		return true, err
	}
	return true, nil
}

// GetWithCache returns a value at key by mutating the result parameter, leveraging hashicorp's LRU cache to avoid
// repeated proto unmarshalling. Returns true if the value was found and the result mutated correctly.
// If the value is not in the store, returns false.
// Returns error only when database or serialization errors occur. (And when an error occurs, returns false)
//
// The cache parameter should be a *lru.Cache from github.com/hashicorp/golang-lru.
// When a value is retrieved from the cache, a copy is made to avoid mutation of the cached value.
func GetWithCache(store storetypes.KVStore, key []byte, cache *lru.Cache, result proto.Message) (found bool, err error) {
	// Get from store
	b := store.Get(key)
	if b == nil {
		return false, nil
	}

	// Try to get from cache first
	// TODO: Use unsafe string conversion to avoid allocation
	cacheKey := string(b)
	if cachedValue, hit := cache.Get(cacheKey); hit {
		if cachedProto, ok := cachedValue.(proto.Message); ok {
			// Make a copy to avoid mutating the cached value
			proto.Merge(result, cachedProto)
			return true, nil
		}
	}

	// Not in cache, unmarshal and add to cache
	if err := proto.Unmarshal(b, result); err != nil {
		return true, err
	}

	// Add to cache
	clone := proto.Clone(result)
	cache.Add(cacheKey, clone)

	return true, nil
}
func SetWithCache(store storetypes.KVStore, key []byte, cache *lru.Cache, value proto.Message) {
	bz, err := proto.Marshal(value)
	if err != nil {
		panic(err)
	}

	// TODO: Use unsafe string conversion to avoid allocation
	cache.Add(string(bz), proto.Clone(value))

	store.Set(key, bz)
}

// DeleteAllKeysFromPrefix deletes all store records that contains the given prefixKey.
func DeleteAllKeysFromPrefix(store storetypes.KVStore, prefixKey []byte) {
	prefixStore := prefix.NewStore(store, prefixKey)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		prefixStore.Delete(iter.Key())
	}
}

// GetCoinArrayFromPrefix returns all coins from the store that has the given prefix.
func GetCoinArrayFromPrefix(ctx sdk.Context, storeKey storetypes.StoreKey, storePrefix []byte) []sdk.Coin {
	coinArray := make([]sdk.Coin, 0)

	store := ctx.KVStore(storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, storePrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		sdkInt := osmomath.Int{}
		if err := sdkInt.Unmarshal(bz); err == nil {
			denom := bytes.TrimPrefix(iterator.Key(), storePrefix)
			coinArray = append(coinArray, sdk.NewCoin(string(denom), sdkInt))
		}
	}

	return coinArray
}

// GetCoinByDenomFromPrefix returns the coin from the store that has the given prefix and denom.
// If the denom is not found, a zero coin is returned.
func GetCoinByDenomFromPrefix(ctx sdk.Context, storeKey storetypes.StoreKey, storePrefix []byte, denom string) (sdk.Coin, error) {
	store := prefix.NewStore(ctx.KVStore(storeKey), storePrefix)
	key := []byte(denom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return sdk.NewCoin(denom, osmomath.ZeroInt()), nil
	}

	sdkInt := osmomath.Int{}
	if err := sdkInt.Unmarshal(bz); err != nil {
		return sdk.NewCoin(denom, osmomath.ZeroInt()), err
	}

	return sdk.NewCoin(denom, sdkInt), nil
}

// IncreaseCoinByDenomFromPrefix increases the coin from the store that has the given prefix and denom by the specified amount.
func IncreaseCoinByDenomFromPrefix(ctx sdk.Context, storeKey storetypes.StoreKey, storePrefix []byte, denom string, increasedAmt osmomath.Int) error {
	store := prefix.NewStore(ctx.KVStore(storeKey), storePrefix)
	key := []byte(denom)

	coin, err := GetCoinByDenomFromPrefix(ctx, storeKey, storePrefix, denom)
	if err != nil {
		return err
	}

	coin.Amount = coin.Amount.Add(increasedAmt)
	bz, err := coin.Amount.Marshal()
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

var kvGasConfig = storetypes.KVGasConfig()

// Get returns a value at key by mutating the result parameter. Returns true if the value was found and the
// result mutated correctly. If the value is not in the store, returns false.
// Returns error only when database or serialization errors occur. (And when an error occurs, returns false)
//
// This function also returns three gas numbers:
// Gas flat, gas for key read, gas for value read.
// You must charge all 3 for the gas accounting to be correct in the current SDK version.
func TrackGasUsedInGet(store storetypes.KVStore, key []byte, result proto.Message) (found bool, gasFlat, gasKey, gasVal uint64, err error) {
	gasFlat = kvGasConfig.ReadCostFlat
	gasKey = uint64(len(key)) * kvGasConfig.ReadCostPerByte
	b := store.Get(key)
	gasVal = uint64(len(b)) * kvGasConfig.ReadCostPerByte
	if b == nil {
		return false, gasFlat, gasKey, gasVal, nil
	}
	if err := proto.Unmarshal(b, result); err != nil {
		return true, gasFlat, gasKey, gasVal, err
	}
	return true, gasFlat, gasKey, gasVal, nil
}

func ChargeMockReadGas(ctx sdk.Context, gasFlat, gasKey, gasVal uint64) {
	ctx.GasMeter().ConsumeGas(gasFlat, storetypes.GasReadCostFlatDesc)
	ctx.GasMeter().ConsumeGas(gasKey, storetypes.GasReadPerByteDesc)
	ctx.GasMeter().ConsumeGas(gasVal, storetypes.GasReadPerByteDesc)
}
