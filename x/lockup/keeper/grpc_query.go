package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
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

// AccountUnlockableCoins returns unlockable coins which are not withdrawn yet
func (k Keeper) AccountUnlockableCoins(goCtx context.Context, req *types.AccountUnlockableCoinsRequest) (*types.AccountUnlockableCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.AccountUnlockableCoinsResponse{Coins: k.GetAccountUnlockableCoins(ctx, req.Owner)}, nil
}

// AccountUnlockingCoins returns whole unlocking coins
func (k Keeper) AccountUnlockingCoins(goCtx context.Context, req *types.AccountUnlockingCoinsRequest) (*types.AccountUnlockingCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.AccountUnlockingCoinsResponse{Coins: k.GetAccountUnlockingCoins(ctx, req.Owner)}, nil
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

// LockedByID Returns lock by lock ID
func (k Keeper) LockedByID(goCtx context.Context, req *types.LockedRequest) (*types.LockedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	lock, err := k.GetLockByID(ctx, req.LockId)
	return &types.LockedResponse{Lock: lock}, err
}

// AccountLockedLongerDuration Returns account locked with duration longer than specified
func (k Keeper) AccountLockedLongerDuration(goCtx context.Context, req *types.AccountLockedLongerDurationRequest) (*types.AccountLockedLongerDurationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	locks := k.GetAccountLockedLongerDuration(ctx, req.Owner, req.Duration)
	return &types.AccountLockedLongerDurationResponse{Locks: locks}, nil
}

// AccountLockedLongerDurationDenom Returns account locked with duration longer than specified with specific denom
func (k Keeper) AccountLockedLongerDurationDenom(goCtx context.Context, req *types.AccountLockedLongerDurationDenomRequest) (*types.AccountLockedLongerDurationDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	locks := k.GetAccountLockedLongerDurationDenom(ctx, req.Owner, req.Denom, req.Duration)
	return &types.AccountLockedLongerDurationDenomResponse{Locks: locks}, nil
}

// AccountLockedPastTimeNotUnlockingOnly Returns locked records of an account with unlock time beyond timestamp excluding tokens started unlocking
func (k Keeper) AccountLockedPastTimeNotUnlockingOnly(goCtx context.Context, req *types.AccountLockedPastTimeNotUnlockingOnlyRequest) (*types.AccountLockedPastTimeNotUnlockingOnlyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.AccountLockedPastTimeNotUnlockingOnlyResponse{Locks: k.GetAccountLockedPastTimeNotUnlockingOnly(ctx, req.Owner, req.Timestamp)}, nil
}

// AccountLockedLongerDurationNotUnlockingOnly Returns account locked records with longer duration excluding tokens started unlocking
func (k Keeper) AccountLockedLongerDurationNotUnlockingOnly(goCtx context.Context, req *types.AccountLockedLongerDurationNotUnlockingOnlyRequest) (*types.AccountLockedLongerDurationNotUnlockingOnlyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.AccountLockedLongerDurationNotUnlockingOnlyResponse{Locks: k.GetAccountLockedLongerDurationNotUnlockingOnly(ctx, req.Owner, req.Duration)}, nil
}
