package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/store"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
	db "github.com/tendermint/tm-db"
)

func (k Keeper) getLocksFromIterator(ctx sdk.Context, iterator db.Iterator) []types.PeriodLock {
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

func (k Keeper) beginUnlockFromIterator(ctx sdk.Context, iterator db.Iterator) ([]types.PeriodLock, sdk.Coins, error) {
	coins := sdk.Coins{}
	locks := k.getLocksFromIterator(ctx, iterator)
	for _, lock := range locks {
		err := k.BeginUnlock(ctx, lock)
		if err != nil {
			return locks, coins, err
		}
		// sum up all coins begin unlocking
		coins = coins.Add(lock.Coins...)
	}
	return locks, coins, nil
}

func (k Keeper) addLockRefs(ctx sdk.Context, lockRefPrefix []byte, lock types.PeriodLock) error {
	refKeys, err := lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		if err := k.addLockRefByKey(ctx, combineKeys(lockRefPrefix, refKey), lock.ID); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) deleteLockRefs(ctx sdk.Context, lockRefPrefix []byte, lock types.PeriodLock) error {
	refKeys, err := lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		if err := k.deleteLockRefByKey(ctx, combineKeys(lockRefPrefix, refKey), lock.ID); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) unlockFromIterator(ctx sdk.Context, iterator db.Iterator) ([]types.PeriodLock, sdk.Coins) {
	coins := sdk.Coins{}
	locks := k.getLocksFromIterator(ctx, iterator)
	for _, lock := range locks {
		err := k.Unlock(ctx, lock)
		if err != nil {
			panic(err)
		}
		// sum up all coins unlocked
		coins = coins.Add(lock.Coins...)
	}
	return locks, coins
}

func (k Keeper) getCoinsFromLocks(locks []types.PeriodLock) sdk.Coins {
	coins := sdk.Coins{}
	for _, lock := range locks {
		coins = coins.Add(lock.Coins...)
	}
	return coins
}

func (k Keeper) getCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getCoinsFromLocks(k.getLocksFromIterator(ctx, iterator))
}

func (k Keeper) accumulationStore(ctx sdk.Context, denom string) store.Tree {
	return store.NewTree(prefix.NewStore(ctx.KVStore(k.storeKey), accumulationStorePrefix(denom)), 10)
}

// GetModuleBalance Returns full balance of the module
func (k Keeper) GetModuleBalance(ctx sdk.Context) sdk.Coins {
	// TODO: should add invariant test for module balance and lock items
	acc := k.ak.GetModuleAccount(ctx, types.ModuleName)
	return k.bk.GetAllBalances(ctx, acc.GetAddress())
}

// GetModuleLockedCoins Returns locked balance of the module
func (k Keeper) GetModuleLockedCoins(ctx sdk.Context) sdk.Coins {
	// all not unlocking + not finished unlocking
	notUnlockingCoins := k.getCoinsFromIterator(ctx, k.LockIterator(ctx, false))
	unlockingCoins := k.getCoinsFromIterator(ctx, k.LockIteratorAfterTime(ctx, true, ctx.BlockTime()))
	return notUnlockingCoins.Add(unlockingCoins...)
}

// GetAccountUnlockableCoins Returns whole unlockable coins which are not withdrawn yet
func (k Keeper) GetAccountUnlockableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.getCoinsFromIterator(ctx, k.AccountLockIteratorBeforeTime(ctx, true, addr, ctx.BlockTime()))
}

// GetAccountUnlockingCoins Returns whole unlocking coins
func (k Keeper) GetAccountUnlockingCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.getCoinsFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, true, addr, ctx.BlockTime()))
}

// GetAccountLockedCoins Returns a locked coins that can't be withdrawn
func (k Keeper) GetAccountLockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	// all account unlocking + not finished unlocking
	notUnlockingCoins := k.getCoinsFromIterator(ctx, k.AccountLockIterator(ctx, false, addr))
	unlockingCoins := k.getCoinsFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, true, addr, ctx.BlockTime()))
	return notUnlockingCoins.Add(unlockingCoins...)
}

// GetAccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp
func (k Keeper) GetAccountLockedPastTime(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	// unlockings finish after specific time + not started locks that will finish after the time even though it start now
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, true, addr, timestamp))
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedPastTimeNotUnlockingOnly Returns the total locks of an account whose unlock time is beyond timestamp
func (k Keeper) GetAccountLockedPastTimeNotUnlockingOnly(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
}

// GetAccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp
func (k Keeper) GetAccountUnlockedBeforeTime(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	// unlockings finish before specific time + not started locks that can finish before the time if start now
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorBeforeTime(ctx, true, addr, timestamp))
	if timestamp.Before(ctx.BlockTime()) {
		return unlockings
	}
	duration := timestamp.Sub(ctx.BlockTime())
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorShorterThanDuration(ctx, false, addr, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific
func (k Keeper) GetAccountLockedPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock {
	// unlockings finish after specific time + not started locks that will finish after the time even though it start now
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorAfterTimeDenom(ctx, true, addr, denom, timestamp))
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDurationDenom(ctx, false, addr, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedLongerDuration Returns account locked with duration longer than specified
func (k Keeper) GetAccountLockedLongerDuration(ctx sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodLock {
	// it does not matter started unlocking or not for duration query
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, true, addr, duration))
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedLongerDurationNotUnlockingOnly Returns account locked with duration longer than specified
func (k Keeper) GetAccountLockedLongerDurationNotUnlockingOnly(ctx sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
}

// GetAccountLockedLongerDurationDenom Returns account locked with duration longer than specified with specific denom
func (k Keeper) GetAccountLockedLongerDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock {
	// it does not matter started unlocking or not for duration query
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDurationDenom(ctx, true, addr, denom, duration))
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDurationDenom(ctx, false, addr, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetLocksPastTimeDenom Returns the locks whose unlock time is beyond timestamp
func (k Keeper) GetLocksPastTimeDenom(ctx sdk.Context, denom string, timestamp time.Time) []types.PeriodLock {
	// returns both unlocking started and not started assuming it started unlocking current time
	unlockings := k.getLocksFromIterator(ctx, k.LockIteratorAfterTimeDenom(ctx, true, denom, timestamp))
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	notUnlockings := k.getLocksFromIterator(ctx, k.LockIteratorLongerThanDurationDenom(ctx, false, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetLocksLongerThanDurationDenom Returns the locks whose unlock duration is longer than duration
func (k Keeper) GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []types.PeriodLock {
	// returns both unlocking started and not started
	unlockings := k.getLocksFromIterator(ctx, k.LockIteratorLongerThanDurationDenom(ctx, true, denom, duration))
	notUnlockings := k.getLocksFromIterator(ctx, k.LockIteratorLongerThanDurationDenom(ctx, false, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetLockByID Returns lock from lockID
func (k Keeper) GetLockByID(ctx sdk.Context, lockID uint64) (*types.PeriodLock, error) {
	lock := types.PeriodLock{}
	store := ctx.KVStore(k.storeKey)
	lockKey := lockStoreKey(lockID)
	if !store.Has(lockKey) {
		return nil, fmt.Errorf("lock with ID %d does not exist", lockID)
	}
	bz := store.Get(lockKey)
	err := proto.Unmarshal(bz, &lock)
	return &lock, err
}

// GetPeriodLocks Returns the period locks on pool
func (k Keeper) GetPeriodLocks(ctx sdk.Context) ([]types.PeriodLock, error) {
	unlockings := k.getLocksFromIterator(ctx, k.LockIterator(ctx, true))
	notUnlockings := k.getLocksFromIterator(ctx, k.LockIterator(ctx, false))
	return combineLocks(notUnlockings, unlockings), nil
}

// GetAccountPeriodLocks Returns the period locks associated to an account
func (k Keeper) GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []types.PeriodLock {
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIterator(ctx, true, addr))
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIterator(ctx, false, addr))
	return combineLocks(notUnlockings, unlockings)
}

// GetPeriodLocksByDuration returns the total amount of query.Denom tokens locked for longer than
// query.Duration
func (k Keeper) GetPeriodLocksAccumulation(ctx sdk.Context, query types.QueryCondition) sdk.Int {
	beginKey := accumulationKey(query.Duration, 0)
	return k.accumulationStore(ctx, query.Denom).SubsetAccumulation(beginKey, nil)
}

// BeginUnlockAllNotUnlockings begins unlock for all not unlocking coins
func (k Keeper) BeginUnlockAllNotUnlockings(ctx sdk.Context, account sdk.AccAddress) ([]types.PeriodLock, sdk.Coins, error) {
	locks, coins, err := k.beginUnlockFromIterator(ctx, k.AccountLockIterator(ctx, false, account))
	return locks, coins, err
}

// UnlockAllUnlockableCoins Unlock all unlockable coins
func (k Keeper) UnlockAllUnlockableCoins(ctx sdk.Context, account sdk.AccAddress) ([]types.PeriodLock, sdk.Coins, error) {
	locks, coins := k.unlockFromIterator(ctx, k.AccountLockIteratorBeforeTime(ctx, true, account, ctx.BlockTime()))
	return locks, coins, nil
}

// BeginUnlockPeriodLockByID begin unlock by period lock ID
func (k Keeper) BeginUnlockPeriodLockByID(ctx sdk.Context, LockID uint64) (*types.PeriodLock, error) {
	lock, err := k.GetLockByID(ctx, LockID)
	if err != nil {
		return lock, err
	}
	err = k.BeginUnlock(ctx, *lock)
	return lock, err
}

// UnlockPeriodLockByID unlock by period lock ID
func (k Keeper) UnlockPeriodLockByID(ctx sdk.Context, LockID uint64) (*types.PeriodLock, error) {
	lock, err := k.GetLockByID(ctx, LockID)
	if err != nil {
		return lock, err
	}
	err = k.Unlock(ctx, *lock)
	return lock, err
}

// LockTokens lock tokens from an account for specified duration
func (k Keeper) LockTokens(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.PeriodLock, error) {
	ID := k.GetLastLockID(ctx) + 1
	// unlock time is set at the beginning of unlocking time
	lock := types.NewPeriodLock(ID, owner, duration, time.Time{}, coins)
	return lock, k.Lock(ctx, lock)
}

// ResetLock reset lock to lock's previous state on InitGenesis
func (k Keeper) ResetLock(ctx sdk.Context, lock types.PeriodLock) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&lock)
	if err != nil {
		return err
	}
	store.Set(lockStoreKey(lock.ID), bz)

	// store refs by the status of unlock
	if lock.IsUnlocking() {
		return k.addLockRefs(ctx, types.KeyPrefixUnlocking, lock)
	}

	// add to accumulation store when unlocking is not started
	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Set(accumulationKey(lock.Duration, lock.ID), coin.Amount)
	}

	return k.addLockRefs(ctx, types.KeyPrefixNotUnlocking, lock)
}

// Lock is a utility to lock coins into module account
func (k Keeper) Lock(ctx sdk.Context, lock types.PeriodLock) error {
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
		if err := k.addLockRefByKey(ctx, combineKeys(types.KeyPrefixNotUnlocking, refKey), lockID); err != nil {
			return err
		}
	}

	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Set(accumulationKey(lock.Duration, lock.ID), coin.Amount)
	}

	k.hooks.OnTokenLocked(ctx, owner, lock.ID, lock.Coins, lock.Duration, lock.EndTime)
	return nil
}

// BeginUnlock is a utility to start unlocking coins from NotUnlocking queue
func (k Keeper) BeginUnlock(ctx sdk.Context, lock types.PeriodLock) error {
	lockID := lock.ID
	refKeys, err := lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		err := k.deleteLockRefByKey(ctx, combineKeys(types.KeyPrefixNotUnlocking, refKey), lockID)
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
		if err := k.addLockRefByKey(ctx, combineKeys(types.KeyPrefixUnlocking, refKey), lockID); err != nil {
			return err
		}
	}

	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Remove(accumulationKey(lock.Duration, lock.ID))
	}

	return nil
}

// Unlock is a utility to unlock coins from module account
func (k Keeper) Unlock(ctx sdk.Context, lock types.PeriodLock) error {
	// validation for current time and unlock time
	curTime := ctx.BlockTime()
	if !lock.IsUnlocking() {
		return fmt.Errorf("lock hasn't started unlocking yet")
	}
	if curTime.Before(lock.EndTime) {
		return fmt.Errorf("lock is not unlockable yet: %s >= %s", curTime.String(), lock.EndTime.String())
	}

	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return err
	}

	// send coins back to owner
	if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, lock.Coins); err != nil {
		return err
	}

	lockID := lock.ID
	store := ctx.KVStore(k.storeKey)
	store.Delete(lockStoreKey(lockID)) // remove lock from store

	refKeys, err := lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		err := k.deleteLockRefByKey(ctx, combineKeys(types.KeyPrefixUnlocking, refKey), lockID)
		if err != nil {
			return err
		}
	}

	k.hooks.OnTokenUnlocked(ctx, owner, lock.ID, lock.Coins, lock.Duration, lock.EndTime)
	return nil
}
