package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/incentives/types"
)

var _ types.QueryServer = Keeper{}

// ModuleToDistributeCoins returns coins that is going to be distributed
func (k Keeper) ModuleToDistributeCoins(goCtx context.Context, req *types.ModuleToDistributeCoinsRequest) (*types.ModuleToDistributeCoinsResponse, error) {
	return nil, nil
}

// ModuleDistributedCoins returns coins that are distributed by module so far
func (k Keeper) ModuleDistributedCoins(goCtx context.Context, req *types.ModuleDistributedCoinsRequest) (*types.ModuleDistributedCoinsResponse, error) {
	return nil, nil
}

// PotByID returns Pot by id
func (k Keeper) PotByID(goCtx context.Context, req *types.PotByIDRequest) (*types.PotByIDResponse, error) {
	return nil, nil
}

// Pots returns pots both upcoming and active
func (k Keeper) Pots(goCtx context.Context, req *types.PotsRequest) (*types.PotsResponse, error) {
	return nil, nil
}

// ActivePots returns active pots
func (k Keeper) ActivePots(goCtx context.Context, req *types.ActivePotsRequest) (*types.ActivePotsResponse, error) {
	return nil, nil
}

// UpcomingPots returns scheduled pots
func (k Keeper) UpcomingPots(goCtx context.Context, req *types.UpcomingPotsRequest) (*types.UpcomingPotsResponse, error) {
	return nil, nil
}

// RewardsEst returns rewards estimation at a future specific time
func (k Keeper) RewardsEst(goCtx context.Context, req *types.RewardsEstRequest) (*types.RewardsEstResponse, error) {
	return nil, nil
}
