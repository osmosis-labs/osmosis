package keeper

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

// GetLastLockID returns ID used last time
func (k Keeper) GetLastLockID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastLockID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLastLockID save ID used by last lock
func (k Keeper) SetLastLockID(ctx sdk.Context, ID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastLockID, sdk.Uint64ToBigEndian(ID))
}

// lockStoreKey returns action store key from ID
func lockStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixPeriodLock, sdk.Uint64ToBigEndian(ID))
}

// getLockRefs get lock IDs specified on the prefix and timestamp key
func (k Keeper) getLockRefs(ctx sdk.Context, key []byte) []uint64 {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, key)
	defer iterator.Close()

	lockIDs := []uint64{}
	for ; iterator.Valid(); iterator.Next() {
		lockID := sdk.BigEndianToUint64(iterator.Value())
		lockIDs = append(lockIDs, lockID)
	}
	return lockIDs
}

// addLockRefByKey make a lockID iterable with the prefix `key`
func (k Keeper) addLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) error {
	store := ctx.KVStore(k.storeKey)
	lockIDBz := sdk.Uint64ToBigEndian(lockID)
	endKey := combineKeys(key, lockIDBz)
	if store.Has(endKey) {
		return fmt.Errorf("lock with same ID exist: %d", lockID)
	}
	store.Set(endKey, lockIDBz)
	return nil
}

// deleteLockRefByKey removes lock ID from an array associated to provided key
func (k Keeper) deleteLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) error {
	store := ctx.KVStore(k.storeKey)
	lockIDKey := sdk.Uint64ToBigEndian(lockID)
	store.Delete(combineKeys(key, lockIDKey))
	return nil
}

func accumulationStorePrefix(denom string) (res []byte) {
	capacity := len(types.KeyPrefixLockAccumulation) + len(denom) + 1
	res = make([]byte, len(types.KeyPrefixLockAccumulation), capacity)
	copy(res, types.KeyPrefixLockAccumulation)
	res = append(res, []byte(denom+"/")...)
	return
}

// accumulationKey should return sort key upon duration.
func accumulationKey(duration time.Duration) (res []byte) {
	res = make([]byte, 8)
	binary.BigEndian.PutUint64(res[:8], uint64(duration))
	return
}

func (k Keeper) ClearAllAccumulationStores(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixLockAccumulation)
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
