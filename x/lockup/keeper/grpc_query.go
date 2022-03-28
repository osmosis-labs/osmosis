package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/claim keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// ModuleBalance Return full balance of the module.
func (q Querier) ModuleBalance(goCtx context.Context, req *types.ModuleBalanceRequest) (*types.ModuleBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleBalanceResponse{Coins: q.Keeper.GetModuleBalance(ctx)}, nil
}

// ModuleLockedAmount Returns locked balance of the module.
func (q Querier) ModuleLockedAmount(goCtx context.Context, req *types.ModuleLockedAmountRequest) (*types.ModuleLockedAmountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleLockedAmountResponse{Coins: q.Keeper.GetModuleLockedCoins(ctx)}, nil
}

// AccountUnlockableCoins returns unlockable coins which are not withdrawn yet.
func (q Querier) AccountUnlockableCoins(goCtx context.Context, req *types.AccountUnlockableCoinsRequest) (*types.AccountUnlockableCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if req.Owner == "" {
		return nil, errors.New("empty address")
	} else if err != nil {
		return nil, err
	}

	return &types.AccountUnlockableCoinsResponse{Coins: q.Keeper.GetAccountUnlockableCoins(ctx, owner)}, nil
}

// AccountUnlockingCoins returns whole unlocking coins.
func (q Querier) AccountUnlockingCoins(goCtx context.Context, req *types.AccountUnlockingCoinsRequest) (*types.AccountUnlockingCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if req.Owner == "" {
		return nil, errors.New("empty address")
	} else if err != nil {
		return nil, err
	}

	return &types.AccountUnlockingCoinsResponse{Coins: q.Keeper.GetAccountUnlockingCoins(ctx, owner)}, nil
}

// AccountLockedCoins Returns a locked coins that can't be withdrawn.
func (q Querier) AccountLockedCoins(goCtx context.Context, req *types.AccountLockedCoinsRequest) (*types.AccountLockedCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if req.Owner == "" {
		return nil, errors.New("empty address")
	} else if err != nil {
		return nil, err
	}

	return &types.AccountLockedCoinsResponse{Coins: q.Keeper.GetAccountLockedCoins(ctx, owner)}, nil
}

// AccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp.
func (q Querier) AccountLockedPastTime(goCtx context.Context, req *types.AccountLockedPastTimeRequest) (*types.AccountLockedPastTimeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if req.Owner == "" {
		return nil, errors.New("empty address")
	} else if err != nil {
		return nil, err
	}

	return &types.AccountLockedPastTimeResponse{Locks: q.Keeper.GetAccountLockedPastTime(ctx, owner, req.Timestamp)}, nil
}

// AccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp.
func (q Querier) AccountUnlockedBeforeTime(goCtx context.Context, req *types.AccountUnlockedBeforeTimeRequest) (*types.AccountUnlockedBeforeTimeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if req.Owner == "" {
		return nil, errors.New("empty address")
	} else if err != nil {
		return nil, err
	}

	return &types.AccountUnlockedBeforeTimeResponse{Locks: q.Keeper.GetAccountUnlockedBeforeTime(ctx, owner, req.Timestamp)}, nil
}

// AccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific.
func (q Querier) AccountLockedPastTimeDenom(goCtx context.Context, req *types.AccountLockedPastTimeDenomRequest) (*types.AccountLockedPastTimeDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if req.Owner == "" {
		return nil, errors.New("empty address")
	} else if err != nil {
		return nil, err
	}

	return &types.AccountLockedPastTimeDenomResponse{Locks: q.Keeper.GetAccountLockedPastTimeDenom(ctx, owner, req.Denom, req.Timestamp)}, nil
}

// LockedByID Returns lock by lock ID.
func (q Querier) LockedByID(goCtx context.Context, req *types.LockedRequest) (*types.LockedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	lock, err := q.Keeper.GetLockByID(ctx, req.LockId)
	return &types.LockedResponse{Lock: lock}, err
}

// SyntheticLockupsByLockupID returns synthetic lockups by native lockup id.
func (q Querier) SyntheticLockupsByLockupID(goCtx context.Context, req *types.SyntheticLockupsByLockupIDRequest) (*types.SyntheticLockupsByLockupIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	synthLocks := q.Keeper.GetAllSyntheticLockupsByLockup(ctx, req.LockId)
	return &types.SyntheticLockupsByLockupIDResponse{SyntheticLocks: synthLocks}, nil
}

// AccountLockedLongerDuration Returns account locked with duration longer than specified.
func (q Querier) AccountLockedLongerDuration(goCtx context.Context, req *types.AccountLockedLongerDurationRequest) (*types.AccountLockedLongerDurationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if req.Owner == "" {
		return nil, errors.New("empty address")
	} else if err != nil {
		return nil, err
	}

	locks := q.Keeper.GetAccountLockedLongerDuration(ctx, owner, req.Duration)
	return &types.AccountLockedLongerDurationResponse{Locks: locks}, nil
}

// AccountLockedLongerDurationDenom Returns account locked with duration longer than specified with specific denom.
func (q Querier) AccountLockedLongerDurationDenom(goCtx context.Context, req *types.AccountLockedLongerDurationDenomRequest) (*types.AccountLockedLongerDurationDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if req.Owner == "" {
		return nil, errors.New("empty address")
	} else if err != nil {
		return nil, err
	}

	locks := q.Keeper.GetAccountLockedLongerDurationDenom(ctx, owner, req.Denom, req.Duration)
	return &types.AccountLockedLongerDurationDenomResponse{Locks: locks}, nil
}

// AccountLockedPastTimeNotUnlockingOnly Returns locked records of an account with unlock time beyond timestamp excluding tokens started unlocking.
func (q Querier) AccountLockedPastTimeNotUnlockingOnly(goCtx context.Context, req *types.AccountLockedPastTimeNotUnlockingOnlyRequest) (*types.AccountLockedPastTimeNotUnlockingOnlyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if req.Owner == "" {
		return nil, errors.New("empty address")
	} else if err != nil {
		return nil, err
	}

	return &types.AccountLockedPastTimeNotUnlockingOnlyResponse{Locks: q.Keeper.GetAccountLockedPastTimeNotUnlockingOnly(ctx, owner, req.Timestamp)}, nil
}

// AccountLockedLongerDurationNotUnlockingOnly Returns account locked records with longer duration excluding tokens started unlocking.
func (q Querier) AccountLockedLongerDurationNotUnlockingOnly(goCtx context.Context, req *types.AccountLockedLongerDurationNotUnlockingOnlyRequest) (*types.AccountLockedLongerDurationNotUnlockingOnlyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if req.Owner == "" {
		return nil, errors.New("empty address")
	} else if err != nil {
		return nil, err
	}

	return &types.AccountLockedLongerDurationNotUnlockingOnlyResponse{Locks: q.Keeper.GetAccountLockedLongerDurationNotUnlockingOnly(ctx, owner, req.Duration)}, nil
}

func (q Querier) LockedDenom(goCtx context.Context, req *types.LockedDenomRequest) (*types.LockedDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.LockedDenomResponse{Amount: q.Keeper.GetLockedDenom(ctx, req.Denom, req.Duration)}, nil
}
