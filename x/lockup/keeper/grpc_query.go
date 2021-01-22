package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

// ModuleBalance Return full balance of the module
func (k Keeper) ModuleBalance(goCtx context.Context, req *types.ModuleBalanceRequest) (*types.ModuleBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleBalanceResponse{Coins: k.GetModuleBalance(ctx)}, nil
}

// ModuleLockedAmount Returns locked balance of the module
func (k Keeper) ModuleLockedAmount(goCtx context.Context, req *types.ModuleLockedAmountRequest) (*types.ModuleLockedAmountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleLockedAmountResponse{Coins: k.GetModuleLockedCoins(ctx)}, nil
}

// AccountUnlockableCoins Returns whole unlockable coins which are not withdrawn yet
func (k Keeper) AccountUnlockableCoins(goCtx context.Context, req *types.AccountUnlockableCoinsRequest) (*types.AccountUnlockableCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.AccountUnlockableCoinsResponse{Coins: k.GetAccountUnlockableCoins(ctx, req.Owner)}, nil
}

// AccountLockedCoins Returns a locked coins that can't be withdrawn
func (k Keeper) AccountLockedCoins(goCtx context.Context, req *types.AccountLockedCoinsRequest) (*types.AccountLockedCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.AccountLockedCoinsResponse{Coins: k.GetAccountLockedCoins(ctx, req.Owner)}, nil
}

// AccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp
func (k Keeper) AccountLockedPastTime(goCtx context.Context, req *types.AccountLockedPastTimeRequest) (*types.AccountLockedPastTimeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.AccountLockedPastTimeResponse{Locks: k.GetAccountLockedPastTime(ctx, req.Owner, req.Timestamp)}, nil
}

// AccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp
func (k Keeper) AccountUnlockedBeforeTime(goCtx context.Context, req *types.AccountUnlockedBeforeTimeRequest) (*types.AccountUnlockedBeforeTimeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.AccountUnlockedBeforeTimeResponse{Locks: k.GetAccountUnlockedBeforeTime(ctx, req.Owner, req.Timestamp)}, nil
}

// AccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific
func (k Keeper) AccountLockedPastTimeDenom(goCtx context.Context, req *types.AccountLockedPastTimeDenomRequest) (*types.AccountLockedPastTimeDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.AccountLockedPastTimeDenomResponse{Locks: k.GetAccountLockedPastTimeDenom(ctx, req.Owner, req.Denom, req.Timestamp)}, nil
}

// Locked Returns lock by lock ID
func (k Keeper) Locked(goCtx context.Context, req *types.LockedRequest) (*types.LockedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	lock, err := k.GetLockByID(ctx, req.LockId)
	return &types.LockedResponse{Lock: lock}, err
}

// AccountLockedLongerThanDuration Returns account locked with duration longer than specified
func (k Keeper) AccountLockedLongerThanDuration(goCtx context.Context, req *types.AccountLockedLongerDurationRequest) (*types.AccountLockedLongerDurationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	locks := k.GetAccountLockedLongerThanDuration(ctx, req.Owner, req.Duration)
	return &types.AccountLockedLongerDurationResponse{Locks: locks}, nil
}

// AccountLockedLongerThanDurationDenom Returns account locked with duration longer than specified with specific denom
func (k Keeper) AccountLockedLongerThanDurationDenom(goCtx context.Context, req *types.AccountLockedLongerDurationDenomRequest) (*types.AccountLockedLongerDurationDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	locks := k.GetAccountLockedLongerThanDurationDenom(ctx, req.Owner, req.Denom, req.Duration)
	return &types.AccountLockedLongerDurationDenomResponse{Locks: locks}, nil
}
