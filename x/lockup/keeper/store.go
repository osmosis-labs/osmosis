package keeper

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetLastLockID returns ID used last time.
func (k Keeper) GetLastLockID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastLockID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLastLockID save ID used by last lock.
func (k Keeper) SetLastLockID(ctx sdk.Context, ID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastLockID, sdk.Uint64ToBigEndian(ID))
}

// lockStoreKey returns action store key from ID.
func lockStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixPeriodLock, sdk.Uint64ToBigEndian(ID))
}

// syntheticLockStoreKey returns synthetic store key from ID and synth denom.
func syntheticLockStoreKey(lockID uint64, synthDenom string) []byte {
	return combineKeys(combineKeys(types.KeyPrefixSyntheticLockup, sdk.Uint64ToBigEndian(lockID)), []byte(synthDenom))
}

// syntheticLockTimeStoreKey returns synthetic store key from ID, synth denom and time.
func syntheticLockTimeStoreKey(lockID uint64, synthDenom string, endTime time.Time) []byte {
	return combineKeys(
		combineKeys(
			combineKeys(types.KeyPrefixSyntheticLockTimestamp, getTimeKey(endTime)),
			sdk.Uint64ToBigEndian(lockID),
		),
		[]byte(synthDenom))
}

// getLockRefs get lock IDs specified on the prefix and timestamp key.
// nolint: unused
func (k Keeper) getLockRefs(ctx sdk.Context, key []byte) []uint64 {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, key)
	defer iterator.Close()

	lockIDs := []uint64{}
	for ; iterator.Valid(); iterator.Next() {
		lockID := sdk.BigEndianToUint64(iterator.Value())
		lockIDs = append(lockIDs, lockID)
	}
	return lockIDs
}

// addLockRefByKey make a lockID iterable with the prefix `key`.
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

// deleteLockRefByKey removes lock ID from an array associated to provided key.
func (k Keeper) deleteLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) {
	store := ctx.KVStore(k.storeKey)
	lockIDKey := sdk.Uint64ToBigEndian(lockID)
	store.Delete(combineKeys(key, lockIDKey))
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

// GetAccountUnlockableCoins Returns whole unlockable coins which are not withdrawn yet.
func (k Keeper) GetAccountUnlockableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.getCoinsFromIterator(ctx, k.AccountLockIteratorBeforeTime(ctx, addr, ctx.BlockTime()))
}

// GetAccountUnlockingCoins Returns whole unlocking coins.
func (k Keeper) GetAccountUnlockingCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.getCoinsFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, addr, ctx.BlockTime()))
}

// GetAccountLockedCoins Returns a locked coins that can't be withdrawn.
func (k Keeper) GetAccountLockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	// all account unlocking + not finished unlocking
	notUnlockingCoins := k.getCoinsFromIterator(ctx, k.AccountLockIterator(ctx, false, addr))
	unlockingCoins := k.getCoinsFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, addr, ctx.BlockTime()))
	return notUnlockingCoins.Add(unlockingCoins...)
}

// GetAccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp.
func (k Keeper) GetAccountLockedPastTime(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	// unlockings finish after specific time + not started locks that will finish after the time even though it start now
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, addr, timestamp))
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedPastTimeNotUnlockingOnly Returns the total locks of an account whose unlock time is beyond timestamp.
func (k Keeper) GetAccountLockedPastTimeNotUnlockingOnly(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
}

// GetAccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp.
func (k Keeper) GetAccountUnlockedBeforeTime(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	// unlockings finish before specific time + not started locks that can finish before the time if start now
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorBeforeTime(ctx, addr, timestamp))
	if timestamp.Before(ctx.BlockTime()) {
		return unlockings
	}
	duration := timestamp.Sub(ctx.BlockTime())
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorShorterThanDuration(ctx, false, addr, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific.
func (k Keeper) GetAccountLockedPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock {
	// unlockings finish after specific time + not started locks that will finish after the time even though it start now
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorAfterTimeDenom(ctx, addr, denom, timestamp))
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDurationDenom(ctx, false, addr, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedDurationNotUnlockingOnly Returns account locked with specific duration within not unlockings.
func (k Keeper) GetAccountLockedDurationNotUnlockingOnly(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorDurationDenom(ctx, false, addr, denom, duration))
}

// GetAccountLockedLongerDuration Returns account locked with duration longer than specified.
func (k Keeper) GetAccountLockedLongerDuration(ctx sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodLock {
	// it does not matter started unlocking or not for duration query
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, true, addr, duration))
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedDuration returns locks with a specific duration for a given account.
func (k Keeper) GetAccountLockedDuration(ctx sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodLock {
	// it does not matter started unlocking or not for duration query
	unlockedLocks := k.getLocksFromIterator(ctx, k.AccountLockIteratorDuration(ctx, true, addr, duration))
	lockedLocks := k.getLocksFromIterator(ctx, k.AccountLockIteratorDuration(ctx, false, addr, duration))
	return combineLocks(unlockedLocks, lockedLocks)
}

// GetAccountLockedLongerDurationNotUnlockingOnly Returns account locked with duration longer than specified
func (k Keeper) GetAccountLockedLongerDurationNotUnlockingOnly(ctx sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
}

// GetAccountLockedLongerDurationDenom Returns account locked with duration longer than specified with specific denom.
func (k Keeper) GetAccountLockedLongerDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock {
	// it does not matter started unlocking or not for duration query
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDurationDenom(ctx, true, addr, denom, duration))
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDurationDenom(ctx, false, addr, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedLongerDurationDenom Returns account locked with duration longer than specified with specific denom.
func (k Keeper) GetAccountLockedLongerDurationDenomNotUnlockingOnly(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDurationDenom(ctx, false, addr, denom, duration))
}

// GetLocksPastTimeDenom Returns the locks whose unlock time is beyond timestamp.
func (k Keeper) GetLocksPastTimeDenom(ctx sdk.Context, denom string, timestamp time.Time) []types.PeriodLock {
	// returns both unlocking started and not started assuming it started unlocking current time
	unlockings := k.getLocksFromIterator(ctx, k.LockIteratorAfterTimeDenom(ctx, denom, timestamp))
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	notUnlockings := k.getLocksFromIterator(ctx, k.LockIteratorLongerThanDurationDenom(ctx, false, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

func (k Keeper) GetLocksDenom(ctx sdk.Context, denom string) []types.PeriodLock {
	return k.GetLocksLongerThanDurationDenom(ctx, denom, time.Duration(0))
}

// GetLockedDenom Returns the total amount of denom that are locked.
func (k Keeper) GetLockedDenom(ctx sdk.Context, denom string, duration time.Duration) osmomath.Int {
	totalAmtLocked := k.GetPeriodLocksAccumulation(ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         denom,
		Duration:      duration,
	})
	return totalAmtLocked
}

// GetLocksLongerThanDurationDenom Returns the locks whose unlock duration is longer than duration.
func (k Keeper) GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []types.PeriodLock {
	// returns both unlocking started and not started
	unlockings := k.getLocksFromIterator(ctx, k.LockIteratorLongerThanDurationDenom(ctx, true, denom, duration))
	notUnlockings := k.getLocksFromIterator(ctx, k.LockIteratorLongerThanDurationDenom(ctx, false, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetLockByID Returns lock from lockID.
func (k Keeper) GetLockByID(ctx sdk.Context, lockID uint64) (*types.PeriodLock, error) {
	lock := types.PeriodLock{}
	store := ctx.KVStore(k.storeKey)
	lockKey := lockStoreKey(lockID)
	if !store.Has(lockKey) {
		return nil, errorsmod.Wrap(types.ErrLockupNotFound, fmt.Sprintf("lock with ID %d does not exist", lockID))
	}
	bz := store.Get(lockKey)
	err := proto.Unmarshal(bz, &lock)
	return &lock, err
}

// GetLockRewardReceiver returns the reward receiver stored in state.
// Note that if the lock reward receiver address in state is an empty string literal,
// it indicates that the lock reward receiver is the owner of the lock, thus
// returns the lock owner address.
func (k Keeper) GetLockRewardReceiver(ctx sdk.Context, lockID uint64) (string, error) {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return "", err
	}

	rewardReceiverAddress := lock.RewardReceiverAddress
	if rewardReceiverAddress == "" {
		rewardReceiverAddress = lock.Owner
	}

	return rewardReceiverAddress, nil
}

// GetPeriodLocks Returns the period locks on pool.
func (k Keeper) GetPeriodLocks(ctx sdk.Context) ([]types.PeriodLock, error) {
	unlockings := k.getLocksFromIterator(ctx, k.LockIterator(ctx, true))
	notUnlockings := k.getLocksFromIterator(ctx, k.LockIterator(ctx, false))
	return combineLocks(notUnlockings, unlockings), nil
}

// GetAccountPeriodLocks Returns the period locks associated to an account.
func (k Keeper) GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []types.PeriodLock {
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIterator(ctx, true, addr))
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIterator(ctx, false, addr))
	return combineLocks(notUnlockings, unlockings)
}
