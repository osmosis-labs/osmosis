package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/lockup/types"
)

var _ types.QueryServer = Keeper{}

// GetModuleBalance Return full balance of the module
func (k Keeper) GetModuleBalance(ctx context.Context, req *types.ModuleBalanceRequest) (*types.ModuleBalanceResponse, error) {
	return nil, nil
}

// GetModuleLockedAmount Returns locked balance of the module
func (k Keeper) GetModuleLockedAmount(ctx context.Context, req *types.ModuleLockedAmountRequest) (*types.ModuleLockedAmountResponse, error) {
	return nil, nil
}

// GetAccountUnlockableCoins Returns whole unlockable coins which are not withdrawn yet
func (k Keeper) GetAccountUnlockableCoins(ctx context.Context, req *types.AccountUnlockableCoinsRequest) (*types.AccountUnlockableCoinsResponse, error) {
	return nil, nil
}

// GetAccountLockedCoins Returns a locked coins that can't be withdrawn
func (k Keeper) GetAccountLockedCoins(ctx context.Context, req *types.AccountLockedCoinsRequest) (*types.AccountLockedCoinsResponse, error) {
	return nil, nil
}

// GetAccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp
func (k Keeper) GetAccountLockedPastTime(ctx context.Context, req *types.AccountLockedPastTimeRequest) (*types.AccountLockedPastTimeResponse, error) {
	return nil, nil
}

// GetAccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp
func (k Keeper) GetAccountUnlockedBeforeTime(ctx context.Context, req *types.AccountUnlockedBeforeTimeRequest) (*types.AccountUnlockedBeforeTimeResponse, error) {
	return nil, nil
}

// GetAccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific
func (k Keeper) GetAccountLockedPastTimeDenom(ctx context.Context, req *types.AccountLockedPastTimeDenomRequest) (*types.AccountLockedPastTimeDenomResponse, error) {
	return nil, nil
}

// GetLock Returns the length of the initial lock time when the lock was created
func (k Keeper) GetLock(ctx context.Context, req *types.LockRequest) (*types.LockResponse, error) {
	return nil, nil
}

// GetAccountLockedLongerThanDuration Returns account locked with duration longer than specified
func (k Keeper) GetAccountLockedLongerThanDuration(ctx context.Context, req *types.AccountLockedLongerDurationRequest) (*types.AccountLockedLongerDurationResponse, error) {
	return nil, nil
}

// GetAccountLockedLongerThanDurationDenom Returns account locked with duration longer than specified with specific denom
func (k Keeper) GetAccountLockedLongerThanDurationDenom(ctx context.Context, req *types.AccountLockedLongerDurationDenomRequest) (*types.AccountLockedLongerDurationDenomResponse, error) {
	return nil, nil
}
