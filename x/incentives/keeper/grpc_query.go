package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

// ModuleToDistributeCoins returns coins that is going to be distributed
func (k Keeper) ModuleToDistributeCoins(goCtx context.Context, req *types.ModuleToDistributeCoinsRequest) (*types.ModuleToDistributeCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleToDistributeCoinsResponse{Coins: k.GetModuleToDistributeCoins(ctx)}, nil
}

// ModuleDistributedCoins returns coins that are distributed by module so far
func (k Keeper) ModuleDistributedCoins(goCtx context.Context, req *types.ModuleDistributedCoinsRequest) (*types.ModuleDistributedCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleDistributedCoinsResponse{Coins: k.GetModuleDistributedCoins(ctx)}, nil
}

// PotByID returns Pot by id
func (k Keeper) PotByID(goCtx context.Context, req *types.PotByIDRequest) (*types.PotByIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	pot, err := k.GetPotByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.PotByIDResponse{Pot: pot}, nil
}

// Pots returns pots both upcoming and active
func (k Keeper) Pots(goCtx context.Context, req *types.PotsRequest) (*types.PotsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.PotsResponse{Data: k.GetPots(ctx)}, nil
}

// ActivePots returns active pots
func (k Keeper) ActivePots(goCtx context.Context, req *types.ActivePotsRequest) (*types.ActivePotsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ActivePotsResponse{Data: k.GetPots(ctx)}, nil
}

// UpcomingPots returns scheduled pots
func (k Keeper) UpcomingPots(goCtx context.Context, req *types.UpcomingPotsRequest) (*types.UpcomingPotsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.UpcomingPotsResponse{Data: k.GetPots(ctx)}, nil
}

// RewardsEst returns rewards estimation at a future specific time
func (k Keeper) RewardsEst(goCtx context.Context, req *types.RewardsEstRequest) (*types.RewardsEstResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.RewardsEstResponse{Coins: k.GetRewardsEst(ctx)}, nil
}
