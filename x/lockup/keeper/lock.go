package keeper

import (
	"fmt"
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

// GetModuleBalance Returns full balance of the module
func (k Keeper) GetModuleBalance(context sdk.Context) sdk.Coins {
	return sdk.Coins{}
}

// GetModuleLockedAmount Returns locked balance of the module
func (k Keeper) GetModuleLockedAmount(context sdk.Context) sdk.Coins {
	return sdk.Coins{}
}

// GetAccountUnlockableCoins Returns whole unlockable coins which are not withdrawn yet
func (k Keeper) GetAccountUnlockableCoins(context sdk.Context, account sdk.AccAddress) sdk.Coins {
	return sdk.Coins{}
}

// GetAccountLockedCoins Returns a locked coins that can't be withdrawn
func (k Keeper) GetAccountLockedCoins(context sdk.Context, account sdk.AccAddress) sdk.Coins {
	return sdk.Coins{}
}

// GetAccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp
func (k Keeper) GetAccountLockedPastTime(context sdk.Context, account sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	return []types.PeriodLock{}
}

// GetAccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp
func (k Keeper) GetAccountUnlockedBeforeTime(context sdk.Context, account sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	return []types.PeriodLock{}
}

// GetAccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific
func (k Keeper) GetAccountLockedPastTimeDenom(context sdk.Context, account sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock {
	return []types.PeriodLock{}
}

// IteratorAccountsLockedPastTimeDenom Get iterator for all locks of a denom token that unlocks after timestamp
func (k Keeper) IteratorAccountsLockedPastTimeDenom(context sdk.Context, denom string, timestamp time.Time) db.Iterator {
	return nil
}

// IteratorLockPeriodsDenom Returns all the accounts that locked coins for longer than time.Duration.  Doesn't matter how long is left until unlock.  Only based on initial locktimes
func (k Keeper) IteratorLockPeriodsDenom(context sdk.Context, denom string, duration time.Duration) []types.PeriodLock {
	return []types.PeriodLock{}
}

// GetAccountLockPeriod Returns the length of the initial lock time when the lock was created
func (k Keeper) GetAccountLockPeriod(context sdk.Context, account sdk.AccAddress, lockID uint64) time.Duration {
	return time.Second
}

// GetPeriodLockKey returns the prefix key used for getting a set of period locks
// where unlockTime is after a specific time
func (k Keeper) GetPeriodLockKey(timestamp time.Time) []byte {
	timeBz := sdk.FormatTimeBytes(timestamp)
	timeBzL := len(timeBz)
	prefixL := len(types.KeyPrefixPeriodLock)

	bz := make([]byte, prefixL+8+timeBzL)

	// copy the prefix
	copy(bz[:prefixL], types.KeyPrefixPeriodLock)

	// copy the encoded time bytes length
	copy(bz[prefixL:prefixL+8], sdk.Uint64ToBigEndian(uint64(timeBzL)))

	// copy the encoded time bytes
	copy(bz[prefixL+8:prefixL+8+timeBzL], timeBz)
	return bz
}

// GetPeriodLocks Returns the period locks of the account
func (k Keeper) GetPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) ([]types.PeriodLock, error) {
	locks := []types.PeriodLock{}
	if k.ak.GetAccount(ctx, addr) == nil {
		return locks, fmt.Errorf("account %s not found", addr.String())
	}
	iterator := k.PeriodLockIterator(ctx, addr)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		lock := types.PeriodLock{}
		k.cdc.MustUnmarshalJSON(iterator.Value(), &lock)
		locks = append(locks, lock)
	}
	return locks, nil
}

// Lock is a utility to lock coins into module account
func (k Keeper) Lock(ctx sdk.Context, lock types.PeriodLock) error {
	if err := k.bk.SendCoinsFromAccountToModule(ctx, lock.Owner, types.ModuleName, lock.Coins); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, lock.Owner)
	key := k.GetPeriodLockKey(lock.EndTime)
	prefixStore.Set(key, k.cdc.MustMarshalJSON(&lock))
	return nil
}

// UnlockAllUnlockableCoins Unlock all unlockable coins
func (k Keeper) UnlockAllUnlockableCoins(context sdk.Context, account sdk.AccAddress) (sdk.Coins, error) {
	return sdk.Coins{}, nil
}

// UnlockPeriodLockByID unlock by period lock ID
func (k Keeper) UnlockPeriodLockByID(context sdk.Context, LockID uint64) (types.PeriodLock, error) {
	return types.PeriodLock{}, nil
}

// PeriodLockIteratorAfter returns the iterator used for getting a set of locks by timestamp
// where unlock time is after a specific time
func (k Keeper) PeriodLockIteratorAfter(ctx sdk.Context, addr sdk.AccAddress, time time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, addr)
	key := k.GetPeriodLockKey(time)
	return prefixStore.Iterator(key, storetypes.PrefixEndBytes(types.KeyPrefixPeriodLock))
}

// PeriodLockIteratorBefore returns the iterator used for getting a set of locks by timestamp
// where unlock time is before a specific time
func (k Keeper) PeriodLockIteratorBefore(ctx sdk.Context, addr sdk.AccAddress, time time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, addr)
	key := k.GetPeriodLockKey(time)
	return prefixStore.Iterator([]byte{}, key)
}

// PeriodLockIterator returns the iterator used for getting a all locks
func (k Keeper) PeriodLockIterator(ctx sdk.Context, addr sdk.AccAddress) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, addr)
	return sdk.KVStorePrefixIterator(prefixStore, types.KeyPrefixPeriodLock)
}
