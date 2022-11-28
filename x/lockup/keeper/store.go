package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v3/x/lockup/types"
)

// getLastLockID returns ID used last time
func (k Keeper) getLastLockID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastLockID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// setLastLockID save ID used by last lock
func (k Keeper) setLastLockID(ctx sdk.Context, ID uint64) {
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
	lockIDs := []uint64{}
	if store.Has(key) {
		bz := store.Get(key)
		err := json.Unmarshal(bz, &lockIDs)
		if err != nil {
			panic(err)
		}
	}
	return lockIDs
}

// addLockRefByKey append lock ID into an array associated to provided key
func (k Keeper) addLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) error {
	store := ctx.KVStore(k.storeKey)
	lockIDs := k.getLockRefs(ctx, key)
	if findIndex(lockIDs, lockID) > -1 {
		return fmt.Errorf("lock with same ID exist: %d", lockID)
	}
	lockIDs = append(lockIDs, lockID)
	bz, err := json.Marshal(lockIDs)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// deleteLockRefByKey removes lock ID from an array associated to provided key
//nolint:ineffassign
func (k Keeper) deleteLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) error {
	index := -1
	store := ctx.KVStore(k.storeKey)
	lockIDs := k.getLockRefs(ctx, key)
	lockIDs, index = removeValue(lockIDs, lockID)
	if index < 0 {
		return fmt.Errorf("specific lock with ID %d not found", lockID)
	}
	if len(lockIDs) == 0 {
		if store.Has(key) {
			store.Delete(key)
		}
	} else {
		bz, err := json.Marshal(lockIDs)
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	return nil
}

//nolint:makezero
func accumulationStorePrefix(denom string) (res []byte) {
	res = make([]byte, len(types.KeyPrefixLockAccumulation))
	copy(res, types.KeyPrefixLockAccumulation)
	res = append(res, []byte(denom+"/")...)
	return
}

// accumulationKey should return sort key upon duration.
// lockID is for preventing key duplication.
func accumulationKey(duration time.Duration, lockID uint64) (res []byte) {
	res = make([]byte, 16)
	binary.BigEndian.PutUint64(res[:8], uint64(duration))
	binary.BigEndian.PutUint64(res[8:], lockID)
	return
}
