package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/lockup/types"
)

var _ types.QueryServer = Keeper{}

// ModuleBalance Return full balance of the module
func (k Keeper) ModuleBalance(ctx context.Context, req *types.ModuleBalanceRequest) (*types.ModuleBalanceResponse, error) {
	return nil, nil
}

// ModuleLockedAmount Returns locked balance of the module
func (k Keeper) ModuleLockedAmount(ctx context.Context, req *types.ModuleLockedAmountRequest) (*types.ModuleLockedAmountResponse, error) {
	return nil, nil
}

// AccountUnlockableCoins Returns whole unlockable coins which are not withdrawn yet
func (k Keeper) AccountUnlockableCoins(ctx context.Context, req *types.AccountUnlockableCoinsRequest) (*types.AccountUnlockableCoinsResponse, error) {
	return nil, nil
}

// AccountLockedCoins Returns a locked coins that can't be withdrawn
func (k Keeper) AccountLockedCoins(ctx context.Context, req *types.AccountLockedCoinsRequest) (*types.AccountLockedCoinsResponse, error) {
	return nil, nil
}

// AccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp
func (k Keeper) AccountLockedPastTime(ctx context.Context, req *types.AccountLockedPastTimeRequest) (*types.AccountLockedPastTimeResponse, error) {
	return nil, nil
}

// AccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp
func (k Keeper) AccountUnlockedBeforeTime(ctx context.Context, req *types.AccountUnlockedBeforeTimeRequest) (*types.AccountUnlockedBeforeTimeResponse, error) {
	return nil, nil
}

// AccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific
func (k Keeper) AccountLockedPastTimeDenom(ctx context.Context, req *types.AccountLockedPastTimeDenomRequest) (*types.AccountLockedPastTimeDenomResponse, error) {
	return nil, nil
}

// Locked Returns lock by lock ID
func (k Keeper) Locked(ctx context.Context, req *types.LockedRequest) (*types.LockedResponse, error) {
	return nil, nil
}

// AccountLockedLongerThanDuration Returns account locked with duration longer than specified
func (k Keeper) AccountLockedLongerThanDuration(ctx context.Context, req *types.AccountLockedLongerDurationRequest) (*types.AccountLockedLongerDurationResponse, error) {
	return nil, nil
}

// AccountLockedLongerThanDurationDenom Returns account locked with duration longer than specified with specific denom
func (k Keeper) AccountLockedLongerThanDurationDenom(ctx context.Context, req *types.AccountLockedLongerDurationDenomRequest) (*types.AccountLockedLongerDurationDenomResponse, error) {
	return nil, nil
}
