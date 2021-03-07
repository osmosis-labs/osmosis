package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

func (k Keeper) getLocksFromIterator(ctx sdk.Context, iterator db.Iterator) []types.PeriodLock {
	locks := []types.PeriodLock{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		timeLock := types.LockIDs{}
		err := json.Unmarshal(iterator.Value(), &timeLock)
		if err != nil {
			panic(err)
		}
		for _, lockID := range timeLock.IDs {
			lock, err := k.GetLockByID(ctx, lockID)
			if err != nil {
				panic(err)
			}
			locks = append(locks, *lock)
		}
	}
	return locks
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

// GetModuleBalance Returns full balance of the module
func (k Keeper) GetModuleBalance(ctx sdk.Context) sdk.Coins {
	// TODO: should add invariant test for module balance and lock items
	acc := k.ak.GetModuleAccount(ctx, types.ModuleName)
	return k.bk.GetAllBalances(ctx, acc.GetAddress())
}

// GetModuleLockedCoins Returns locked balance of the module
func (k Keeper) GetModuleLockedCoins(ctx sdk.Context) sdk.Coins {
	return k.getCoinsFromIterator(ctx, k.LockIteratorAfterTime(ctx, ctx.BlockTime()))
}

// GetAccountUnlockableCoins Returns whole unlockable coins which are not withdrawn yet
func (k Keeper) GetAccountUnlockableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.getCoinsFromIterator(ctx, k.AccountLockIteratorBeforeTime(ctx, addr, ctx.BlockTime()))
}

// GetAccountLockedCoins Returns a locked coins that can't be withdrawn
func (k Keeper) GetAccountLockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.getCoinsFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, addr, ctx.BlockTime()))
}

// GetAccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp
func (k Keeper) GetAccountLockedPastTime(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, addr, timestamp))
}

// GetAccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp
func (k Keeper) GetAccountUnlockedBeforeTime(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorBeforeTime(ctx, addr, timestamp))
}

// GetAccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific
func (k Keeper) GetAccountLockedPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorAfterTimeDenom(ctx, addr, denom, timestamp))
}

// GetAccountLockedLongerThanDuration Returns account locked with duration longer than specified
func (k Keeper) GetAccountLockedLongerThanDuration(ctx sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerThanDuration(ctx, addr, duration))
}

// GetAccountLockedLongerThanDurationDenom Returns account locked with duration longer than specified with specific denom
func (k Keeper) GetAccountLockedLongerThanDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerThanDurationDenom(ctx, addr, denom, duration))
}

// GetLocksPastTimeDenom Returns the locks whose unlock time is beyond timestamp
func (k Keeper) GetLocksPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.LockIteratorAfterTimeDenom(ctx, denom, timestamp))
}

// GetLocksLongerThanDurationDenom Returns the locks whose unlock duration is longer than duration
func (k Keeper) GetLocksLongerThanDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.LockIteratorLongerThanDurationDenom(ctx, denom, duration))
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
	k.cdc.MustUnmarshalJSON(bz, &lock)
	return &lock, nil
}

// GetPeriodLocks Returns the period locks on pool
func (k Keeper) GetPeriodLocks(ctx sdk.Context) ([]types.PeriodLock, error) {
	return k.getLocksFromIterator(ctx, k.LockIterator(ctx)), nil
}

// GetAccountPeriodLocks Returns the period locks associated to an account
func (k Keeper) GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) ([]types.PeriodLock, error) {
	return k.getLocksFromIterator(ctx, k.AccountLockIterator(ctx, addr)), nil
}

// UnlockAllUnlockableCoins Unlock all unlockable coins
func (k Keeper) UnlockAllUnlockableCoins(ctx sdk.Context, account sdk.AccAddress) ([]types.PeriodLock, sdk.Coins, error) {
	locks, coins := k.unlockFromIterator(ctx, k.AccountLockIteratorBeforeTime(ctx, account, ctx.BlockTime()))
	return locks, coins, nil
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
	ID := k.getLastLockID(ctx) + 1
	lock := types.NewPeriodLock(ID, owner, duration, ctx.BlockTime().Add(duration), coins)
	return lock, k.Lock(ctx, lock)
}

// Lock is a utility to lock coins into module account
func (k Keeper) Lock(ctx sdk.Context, lock types.PeriodLock) error {
	if err := k.bk.SendCoinsFromAccountToModule(ctx, lock.Owner, types.ModuleName, lock.Coins); err != nil {
		return err
	}

	lockID := lock.ID
	store := ctx.KVStore(k.storeKey)
	store.Set(lockStoreKey(lockID), k.cdc.MustMarshalJSON(&lock))
	k.setLastLockID(ctx, lockID)

	refKeys := lockRefKeys(lock)
	for _, refKey := range refKeys {
		if err := k.addLockRefByKey(ctx, refKey, lockID); err != nil {
			return err
		}
	}
	return nil
}

// Unlock is a utility to unlock coins from module account
func (k Keeper) Unlock(ctx sdk.Context, lock types.PeriodLock) error {
	// validation for current time and unlock time
	curTime := ctx.BlockTime()
	if curTime.Before(lock.EndTime) {
		return fmt.Errorf("lock is not unlockable yet: %s >= %s", curTime.String(), lock.EndTime.String())
	}

	// send coins back to owner
	if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lock.Owner, lock.Coins); err != nil {
		return err
	}

	lockID := lock.ID
	store := ctx.KVStore(k.storeKey)
	store.Delete(lockStoreKey(lockID)) // remove lock from store

	refKeys := lockRefKeys(lock)
	for _, refKey := range refKeys {
		k.deleteLockRefByKey(ctx, refKey, lockID)
	}
	return nil
}
