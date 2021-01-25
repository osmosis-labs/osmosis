package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// LockStoreKey returns action store key from ID
func LockStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixPeriodLock, sdk.Uint64ToBigEndian(ID))
}

// GetLockRefs get lock IDs specified on the prefix and timestamp key
func (k Keeper) GetLockRefs(ctx sdk.Context, key []byte) types.LockIDs {
	store := ctx.KVStore(k.storeKey)
	timeLock := types.LockIDs{}
	if store.Has(key) {
		bz := store.Get(key)
		err := json.Unmarshal(bz, &timeLock)
		if err != nil {
			panic(err)
		}
	}
	return timeLock
}

func (k Keeper) AppendLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) {
	store := ctx.KVStore(k.storeKey)
	timeLock := k.GetLockRefs(ctx, key)
	timeLock.IDs = append(timeLock.IDs, lockID)
	bz, err := json.Marshal(timeLock)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

func (k Keeper) DeleteLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) {
	var index = -1
	store := ctx.KVStore(k.storeKey)
	timeLock := k.GetLockRefs(ctx, key)
	timeLock.IDs, index = removeValue(timeLock.IDs, lockID)
	if index < 0 {
		panic(fmt.Sprintf("specific lock with ID %d not found", lockID))
	}
	if len(timeLock.IDs) == 0 {
		store.Delete(key)
	} else {
		bz, err := json.Marshal(timeLock)
		if err != nil {
			panic(err)
		}
		store.Set(key, bz)
	}
}
