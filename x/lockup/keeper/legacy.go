package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/v6/x/lockup/types"
	db "github.com/tendermint/tm-db"
)

func legacyAccumulationKey(duration time.Duration, id uint64) []byte {
	res := make([]byte, 16)
	binary.BigEndian.PutUint64(res[:8], uint64(duration))
	binary.BigEndian.PutUint64(res[8:], id)
	return res
}

func findIndex(IDs []uint64, ID uint64) int {
	for index, id := range IDs {
		if id == ID {
			return index
		}
	}
	return -1
}

func removeValue(IDs []uint64, ID uint64) ([]uint64, int) {
	index := findIndex(IDs, ID)
	if index < 0 {
		return IDs, index
	}
	IDs[index] = IDs[len(IDs)-1] // set last element to index
	return IDs[:len(IDs)-1], index
}

// getLockRefs get lock IDs specified on the prefix and timestamp key
func (k Keeper) getLegacyLockRefs(ctx sdk.Context, key []byte) []uint64 {
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

func (k Keeper) getLegacyLocksFromIterator(ctx sdk.Context, iterator db.Iterator) []types.PeriodLock {
	locks := []types.PeriodLock{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		lockIDs := []uint64{}
		err := json.Unmarshal(iterator.Value(), &lockIDs)
		if err != nil {
			panic(err)
		}
		for _, lockID := range lockIDs {
			lock, err := k.GetLockByID(ctx, lockID)
			if err != nil {
				panic(err)
			}
			locks = append(locks, *lock)
		}
	}
	return locks
}

// GetLegacyPeriodLocks Returns the period locks on pool
func (k Keeper) GetLegacyPeriodLocks(ctx sdk.Context) ([]types.PeriodLock, error) {
	maxID := int(k.GetLastLockID(ctx) + 1)

	locks := make([]types.PeriodLock, 0, maxID)
	store := ctx.KVStore(k.storeKey)
	for lockID := 0; lockID < maxID; lockID++ {
		if lockID%10000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("Fetched %d locks", lockID))
		}
		// Copy in GetLockByID logic, with optimizations for hotloop
		lockKey := lockStoreKey(uint64(lockID))
		if !store.Has(lockKey) {
			continue
		}
		lock := types.PeriodLock{}
		bz := store.Get(lockKey)
		err := proto.Unmarshal(bz, &lock)
		if err != nil {
			return nil, err
		}
		locks = append(locks, lock)
	}
	return locks, nil
}

// addLockRefByKey append lock ID into an array associated to provided key
func (k Keeper) addLegacyLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) error {
	store := ctx.KVStore(k.storeKey)
	lockIDs := k.getLegacyLockRefs(ctx, key)
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

// deleteLegacyLockRefByKey removes lock ID from an array associated to provided key
//nolint:ineffassign
func (k Keeper) deleteLegacyLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) error {
	index := -1
	store := ctx.KVStore(k.storeKey)
	lockIDs := k.getLegacyLockRefs(ctx, key)
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

// LegacyLockTokens lock tokens from an account for specified duration
func (k Keeper) LegacyLockTokens(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.PeriodLock, error) {
	ID := k.GetLastLockID(ctx) + 1
	// unlock time is set at the beginning of unlocking time
	lock := types.NewPeriodLock(ID, owner, duration, time.Time{}, coins)
	return lock, k.LegacyLock(ctx, lock)
}

// Lock is a utility to lock coins into module account
func (k Keeper) LegacyLock(ctx sdk.Context, lock types.PeriodLock) error {
	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return err
	}
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, lock.Coins); err != nil {
		return err
	}

	lockID := lock.ID
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&lock)
	if err != nil {
		return err
	}
	store.Set(lockStoreKey(lockID), bz)
	k.SetLastLockID(ctx, lockID)

	refKeys, err := lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		if err := k.addLegacyLockRefByKey(ctx, combineKeys(types.KeyPrefixNotUnlocking, refKey), lockID); err != nil {
			return err
		}
	}

	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Set(legacyAccumulationKey(lock.Duration, lock.ID), coin.Amount)
	}

	k.hooks.OnTokenLocked(ctx, owner, lock.ID, lock.Coins, lock.Duration, lock.EndTime)
	return nil
}

// LegacyBeginUnlock is a utility to start unlocking coins from NotUnlocking queue
func (k Keeper) LegacyBeginUnlock(ctx sdk.Context, lock types.PeriodLock) error {
	lockID := lock.ID
	refKeys, err := lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		err := k.deleteLegacyLockRefByKey(ctx, combineKeys(types.KeyPrefixNotUnlocking, refKey), lockID)
		if err != nil {
			return err
		}
	}
	lock.EndTime = ctx.BlockTime().Add(lock.Duration)
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&lock)
	if err != nil {
		return err
	}
	store.Set(lockStoreKey(lockID), bz)

	refKeys, err = lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		if err := k.addLegacyLockRefByKey(ctx, combineKeys(types.KeyPrefixUnlocking, refKey), lockID); err != nil {
			return err
		}
	}

	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Remove(legacyAccumulationKey(lock.Duration, lock.ID))
	}

	return nil
}
