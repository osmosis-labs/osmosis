package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (k Keeper) setSyntheticLockupObject(ctx sdk.Context, lockID uint64, suffix string, endTime time.Time) error {
	synthLock := &types.SyntheticLock{
		LockId:  lockID,
		Suffix:  suffix,
		EndTime: endTime,
	}
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(synthLock)
	if err != nil {
		return err
	}
	store.Set(syntheticLockStoreKey(lockID, suffix), bz)
	if !endTime.Equal(time.Time{}) {
		store.Set(syntheticLockTimeStoreKey(lockID, suffix, endTime), bz)
	}
	return nil
}

func (k Keeper) deleteSyntheticLockupObject(ctx sdk.Context, lockID uint64, suffix string) {
	store := ctx.KVStore(k.storeKey)
	synthLock, _ := k.GetSyntheticLockup(ctx, lockID, suffix)
	if synthLock != nil && !synthLock.EndTime.Equal(time.Time{}) {
		store.Delete(syntheticLockTimeStoreKey(lockID, suffix, synthLock.EndTime))
	}
	store.Delete(syntheticLockStoreKey(lockID, suffix))
}

func (k Keeper) GetSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string) (*types.SyntheticLock, error) {
	synthLock := types.SyntheticLock{}
	store := ctx.KVStore(k.storeKey)
	synthLockKey := syntheticLockStoreKey(lockID, suffix)
	if !store.Has(synthLockKey) {
		return nil, fmt.Errorf("synthetic lock with ID %d and suffix %s does not exist", lockID, suffix)
	}
	bz := store.Get(synthLockKey)
	err := proto.Unmarshal(bz, &synthLock)
	return &synthLock, err
}

func (k Keeper) GetAllSyntheticLockupsByLockup(ctx sdk.Context, lockID uint64) []types.SyntheticLock {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, combineKeys(types.KeyPrefixSyntheticLockup, sdk.Uint64ToBigEndian(lockID)))

	synthLocks := []types.SyntheticLock{}
	for ; iterator.Valid(); iterator.Next() {
		synthLock := types.SyntheticLock{}
		err := proto.Unmarshal(iterator.Value(), &synthLock)
		if err != nil {
			panic(err)
		}
		synthLocks = append(synthLocks, synthLock)
	}
	return synthLocks
}

func (k Keeper) GetAllSyntheticLockups(ctx sdk.Context) []types.SyntheticLock {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixSyntheticLockup)

	synthLocks := []types.SyntheticLock{}
	for ; iterator.Valid(); iterator.Next() {
		synthLock := types.SyntheticLock{}
		err := proto.Unmarshal(iterator.Value(), &synthLock)
		if err != nil {
			panic(err)
		}
		synthLocks = append(synthLocks, synthLock)
	}
	return synthLocks
}

// CreateSyntheticLockup create synthetic lockup with lock id and suffix
func (k Keeper) CreateSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string, unlockDuration time.Duration) error {
	// Note: synthetic lockup is doing everything same as lockup except coin movement
	// There is no relationship between unbonding and bonding synthetic lockup, it's managed separately
	// Accumulation store works without caring about unlocking synthetic or not

	isUnlocking := (unlockDuration == 0)
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	lock.Coins = syntheticCoins(lock.Coins, suffix)
	if isUnlocking { // end time is set automatically if it's unlocking lockup
		lock.EndTime = ctx.BlockTime().Add(unlockDuration)
	} else {
		lock.EndTime = time.Time{}
	}

	// set synthetic lockup object
	err = k.setSyntheticLockupObject(ctx, lockID, suffix, lock.EndTime)
	if err != nil {
		return err
	}

	unlockingPrefix := unlockingPrefix(isUnlocking)

	// add lock refs into not unlocking queue
	err = k.addSyntheticLockRefs(ctx, unlockingPrefix, *lock)
	if err != nil {
		return err
	}

	// add to accumulation store
	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(unlockDuration), coin.Amount)
	}
	return nil
}

// DeleteSyntheticLockup delete synthetic lockup with lock id and suffix
func (k Keeper) DeleteSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string) error {
	synthLock, err := k.GetSyntheticLockup(ctx, lockID, suffix)
	if err != nil {
		return err
	}

	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	// update lock for synthetic lock
	lock.Coins = syntheticCoins(lock.Coins, suffix)
	lock.EndTime = synthLock.EndTime

	k.deleteSyntheticLockupObject(ctx, lockID, suffix)

	// delete lock refs from the unlocking queue
	err = k.deleteSyntheticLockRefs(ctx, unlockingPrefix(lock.IsUnlocking()), *lock)
	if err != nil {
		return err
	}

	// remove from accumulation store
	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Decrease(accumulationKey(lock.Duration), coin.Amount)
	}
	return nil
}

// DeleteAllSyntheticLocksByLockup delete all the synthetic lockups by lockup id
func (k Keeper) DeleteAllSyntheticLocksByLockup(ctx sdk.Context, lockID uint64) error {
	syntheticLocks := k.GetAllSyntheticLockupsByLockup(ctx, lockID)
	for _, synthLock := range syntheticLocks {
		err := k.DeleteSyntheticLockup(ctx, lockID, synthLock.Suffix)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) DeleteAllMaturedSyntheticLocks(ctx sdk.Context) {
	iterator := k.iteratorBeforeTime(ctx, combineKeys(types.KeyPrefixSyntheticLockTimestamp), ctx.BlockTime())
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		synthLock := types.SyntheticLock{}
		err := proto.Unmarshal(iterator.Value(), &synthLock)
		if err != nil {
			panic(err)
		}
		err = k.DeleteSyntheticLockup(ctx, synthLock.LockId, synthLock.Suffix)
		if err != nil {
			panic(err)
		}
	}
}
