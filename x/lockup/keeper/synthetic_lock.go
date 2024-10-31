package keeper

import (
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

// A synthetic lock object is a lock object used for the superfluid module.
// Each synthetic lock object is stored in state using lock id and synthetic denom
// as it's key, where a synthetic denom would be consisted of the original denom of the lock,
// validator address, and the staking positiion of the lock.
// Unlike the original lock objects, synthetic locks are mainly used to indicate the staking
// position of the lock.
// Synthetic use different accumulation store from the original lock objects.
// Note that locks with synthetic objects cannot be directly deleted or cannot directly start
// unlocking. locks with synthetic lock objects are to be unlocked via superfluid module.
// The Endtime and the Duration fields of the synthetic locks do not need to have the same values
// as the underlying lock objects.

// GetSyntheticLockup gets the synthetic lock object using lock ID and synthetic denom as key.
func (k Keeper) GetSyntheticLockup(ctx sdk.Context, lockID uint64, synthdenom string) (*types.SyntheticLock, error) {
	synthLock := types.SyntheticLock{}
	store := ctx.KVStore(k.storeKey)
	synthLockKey := syntheticLockStoreKey(lockID, synthdenom)
	if !store.Has(synthLockKey) {
		return nil, fmt.Errorf("synthetic lock with ID %d and synth denom %s does not exist", lockID, synthdenom)
	}
	bz := store.Get(synthLockKey)
	err := proto.Unmarshal(bz, &synthLock)
	return &synthLock, err
}

// Error is returned if:
// - there are more than one synthetic lockup objects with the same underlying lock ID.
// - there is no synthetic lockup object with the given underlying lock ID.
// Returns (syntheticLockup, found, error)
// intended behavior for most callers is to check:
// if !found || err != nil { handle_it }
func (k Keeper) GetSyntheticLockupByUnderlyingLockId(ctx sdk.Context, lockID uint64) (types.SyntheticLock, bool, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, combineKeys(types.KeyPrefixSyntheticLockup, sdk.Uint64ToBigEndian(lockID)))
	defer iterator.Close()

	synthLocks := []types.SyntheticLock{}
	for ; iterator.Valid(); iterator.Next() {
		synthLock := types.SyntheticLock{}
		err := proto.Unmarshal(iterator.Value(), &synthLock)
		if err != nil {
			return types.SyntheticLock{}, true, err
		}
		synthLocks = append(synthLocks, synthLock)
	}
	if len(synthLocks) > 1 {
		return types.SyntheticLock{}, true, fmt.Errorf("synthetic lockup with same lock id should not exist")
	}
	if len(synthLocks) == 0 {
		return types.SyntheticLock{}, false, nil
	}
	return synthLocks[0], true, nil
}

// GetAllSyntheticLockupsByAddr gets all the synthetic lockups from all the locks owned by the given address.
func (k Keeper) GetAllSyntheticLockupsByAddr(ctx sdk.Context, owner sdk.AccAddress) []types.SyntheticLock {
	synthLocks := []types.SyntheticLock{}
	locks := k.GetAccountPeriodLocks(ctx, owner)
	for _, lock := range locks {
		synthLock, found, err := k.GetSyntheticLockupByUnderlyingLockId(ctx, lock.ID)
		if err != nil {
			panic(err)
		}
		if found {
			synthLocks = append(synthLocks, synthLock)
		}
	}
	return synthLocks
}

// HasAnySyntheticLockups returns true if the lock has a synthetic lock.
func (k Keeper) HasAnySyntheticLockups(ctx sdk.Context, lockID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, combineKeys(types.KeyPrefixSyntheticLockup, sdk.Uint64ToBigEndian(lockID)))
	defer iterator.Close()
	return iterator.Valid()
}

// GetAllSyntheticLockups gets all synthetic locks within the store.
func (k Keeper) GetAllSyntheticLockups(ctx sdk.Context) []types.SyntheticLock {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixSyntheticLockup)
	defer iterator.Close()

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

// CreateSyntheticLockup create synthetic lockup with lock id and synthdenom.
func (k Keeper) CreateSyntheticLockup(ctx sdk.Context, lockID uint64, synthDenom string, unlockDuration time.Duration, isUnlocking bool) error {
	// Note: synthetic lockup is doing everything same as lockup except coin movement
	// There is no relationship between unbonding and bonding synthetic lockup, it's managed separately
	// A separate accumulation store is incremented with the synth denom.

	_, found, err := k.GetSyntheticLockupByUnderlyingLockId(ctx, lockID)
	if err != nil {
		return err
	}
	if found {
		return types.ErrSyntheticLockupAlreadyExists
	}

	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	endTime := time.Time{}
	if isUnlocking { // end time is set automatically if it's unlocking lockup
		if unlockDuration > lock.Duration {
			return types.ErrSyntheticDurationLongerThanNative
		}
		endTime = ctx.BlockTime().Add(unlockDuration)
	}

	// set synthetic lockup object
	synthLock := types.SyntheticLock{
		UnderlyingLockId: lockID,
		SynthDenom:       synthDenom,
		EndTime:          endTime,
		Duration:         unlockDuration,
	}
	err = k.setSyntheticLockupObject(ctx, &synthLock)
	if err != nil {
		return err
	}

	// add lock refs into not unlocking queue
	err = k.addSyntheticLockRefs(ctx, *lock, synthLock)
	if err != nil {
		return err
	}

	coin, err := lock.SingleCoin()
	if err != nil {
		return err
	}

	k.accumulationStore(ctx, synthLock.SynthDenom).Increase(accumulationKey(unlockDuration), coin.Amount)
	return nil
}

// DeleteSyntheticLockup delete synthetic lockup with lock id and synthdenom.
// Synthetic lock has three relevant state entries.
// - synthetic lock object itself
// - synthetic lock refs
// - accumulation store for the synthetic lock.
// all of which are deleted within this method.
func (k Keeper) DeleteSyntheticLockup(ctx sdk.Context, lockID uint64, synthdenom string) error {
	synthLock, err := k.GetSyntheticLockup(ctx, lockID, synthdenom)
	if err != nil {
		return err
	}

	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	// update lock for synthetic lock
	lock.EndTime = synthLock.EndTime

	k.deleteSyntheticLockupObject(ctx, lockID, synthdenom)

	// delete lock refs from the unlocking queue
	err = k.deleteSyntheticLockRefs(ctx, *lock, *synthLock)
	if err != nil {
		return err
	}

	// remove from accumulation store
	coin, err := lock.SingleCoin()
	if err != nil {
		return err
	}
	k.accumulationStore(ctx, synthLock.SynthDenom).Decrease(accumulationKey(lock.Duration), coin.Amount)
	return nil
}

// DeleteAllMaturedSyntheticLocks deletes all matured synthetic locks.
func (k Keeper) DeleteAllMaturedSyntheticLocks(ctx sdk.Context) {
	iterator := k.iteratorBeforeTime(ctx, combineKeys(types.KeyPrefixSyntheticLockTimestamp), ctx.BlockTime())
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		synthLock := types.SyntheticLock{}
		err := proto.Unmarshal(iterator.Value(), &synthLock)
		if err != nil {
			panic(err)
		}
		err = k.DeleteSyntheticLockup(ctx, synthLock.UnderlyingLockId, synthLock.SynthDenom)
		if err != nil {
			// TODO: When underlying lock is deleted for a reason while synthetic lockup exists, panic could happen
			panic(err)
		}
	}
}

func (k Keeper) GetUnderlyingLock(ctx sdk.Context, synthlock types.SyntheticLock) types.PeriodLock {
	lock, err := k.GetLockByID(ctx, synthlock.UnderlyingLockId)
	if err != nil {
		panic(err) // Synthetic lock MUST have underlying lock
	}
	return *lock
}

func (k Keeper) setSyntheticLockupObject(ctx sdk.Context, synthLock *types.SyntheticLock) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(synthLock)
	if err != nil {
		return err
	}
	store.Set(syntheticLockStoreKey(synthLock.UnderlyingLockId, synthLock.SynthDenom), bz)
	if !synthLock.EndTime.Equal(time.Time{}) {
		store.Set(syntheticLockTimeStoreKey(synthLock.UnderlyingLockId, synthLock.SynthDenom, synthLock.EndTime), bz)
	}
	return nil
}

func (k Keeper) deleteSyntheticLockupObject(ctx sdk.Context, lockID uint64, synthdenom string) {
	store := ctx.KVStore(k.storeKey)
	synthLock, _ := k.GetSyntheticLockup(ctx, lockID, synthdenom)
	if synthLock != nil && !synthLock.EndTime.Equal(time.Time{}) {
		store.Delete(syntheticLockTimeStoreKey(lockID, synthdenom, synthLock.EndTime))
	}
	store.Delete(syntheticLockStoreKey(lockID, synthdenom))
}
