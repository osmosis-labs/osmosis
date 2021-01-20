package keeper

import (
	"fmt"
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetModuleBalance Returns full balance of the module
func (k Keeper) GetModuleBalance(ctx sdk.Context) sdk.Coins {
	acc := k.ak.GetModuleAccount(ctx, types.ModuleName)
	addr := acc.GetAddress()
	return k.bk.GetAllBalances(ctx, addr)
}

// GetModuleLockedAmount Returns locked balance of the module
func (k Keeper) GetModuleLockedAmount(ctx sdk.Context) sdk.Coins {
	// TODO: should iterate all the accounts
	return sdk.Coins{}
}

// GetAccountUnlockableCoins Returns whole unlockable coins which are not withdrawn yet
func (k Keeper) GetAccountUnlockableCoins(ctx sdk.Context, account sdk.AccAddress) sdk.Coins {
	return sdk.Coins{}
}

// GetAccountLockedCoins Returns a locked coins that can't be withdrawn
func (k Keeper) GetAccountLockedCoins(ctx sdk.Context, account sdk.AccAddress) sdk.Coins {
	return sdk.Coins{}
}

// GetAccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp
func (k Keeper) GetAccountLockedPastTime(ctx sdk.Context, account sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	return []types.PeriodLock{}
}

// GetAccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp
func (k Keeper) GetAccountUnlockedBeforeTime(ctx sdk.Context, account sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	return []types.PeriodLock{}
}

// GetAccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific
func (k Keeper) GetAccountLockedPastTimeDenom(ctx sdk.Context, account sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock {
	return []types.PeriodLock{}
}

// GetAccountLockPeriod Returns the length of the initial lock time when the lock was created
func (k Keeper) GetAccountLockPeriod(ctx sdk.Context, account sdk.AccAddress, lockID uint64) time.Duration {
	return time.Second
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

// UnlockAllUnlockableCoins Unlock all unlockable coins
func (k Keeper) UnlockAllUnlockableCoins(context sdk.Context, account sdk.AccAddress) (sdk.Coins, error) {
	return sdk.Coins{}, nil
}

// UnlockPeriodLockByID unlock by period lock ID
func (k Keeper) UnlockPeriodLockByID(context sdk.Context, LockID uint64) (types.PeriodLock, error) {
	return types.PeriodLock{}, nil
}
