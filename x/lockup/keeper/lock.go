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
			locks = append(locks, lock)
		}
	}
	return locks
}

func (k Keeper) unlockFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
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
	return coins
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

// GetLockByID Returns lock from lockID
func (k Keeper) GetLockByID(ctx sdk.Context, lockID uint64) (types.PeriodLock, error) {
	lock := types.PeriodLock{}
	store := ctx.KVStore(k.storeKey)
	lockKey := LockStoreKey(lockID)
	if !store.Has(lockKey) {
		return lock, fmt.Errorf("lock with ID %d does not exist", lockID)
	}
	bz := store.Get(lockKey)
	k.cdc.MustUnmarshalJSON(bz, &lock)
	return lock, nil
}

// GetPeriodLocks Returns the period locks on pool
func (k Keeper) GetPeriodLocks(ctx sdk.Context) ([]types.PeriodLock, error) {
	return k.getLocksFromIterator(ctx, k.LockIterator(ctx)), nil
}

// UnlockAllUnlockableCoins Unlock all unlockable coins
func (k Keeper) UnlockAllUnlockableCoins(ctx sdk.Context, account sdk.AccAddress) (sdk.Coins, error) {
	return k.unlockFromIterator(ctx, k.LockIteratorBeforeTime(ctx, ctx.BlockTime())), nil
}

// UnlockPeriodLockByID unlock by period lock ID
func (k Keeper) UnlockPeriodLockByID(context sdk.Context, LockID uint64) (types.PeriodLock, error) {
	return types.PeriodLock{}, nil
}
