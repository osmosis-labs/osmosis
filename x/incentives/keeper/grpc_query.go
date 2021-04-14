package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	pots := []types.Pot{}
	store := ctx.KVStore(k.storeKey)
	valStore := prefix.NewStore(store, types.KeyPrefixPots)

	pageRes, err := query.FilteredPaginate(valStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		newPots, err := k.GetPotFromIDs(ctx, value)
		if err != nil {
			panic(err)
		}
		pots = append(pots, newPots...)

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.PotsResponse{Data: pots, Pagination: pageRes}, nil
}

// ActivePots returns active pots
func (k Keeper) ActivePots(goCtx context.Context, req *types.ActivePotsRequest) (*types.ActivePotsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	pots := []types.Pot{}
	store := ctx.KVStore(k.storeKey)
	valStore := prefix.NewStore(store, types.KeyPrefixActivePots)

	pageRes, err := query.FilteredPaginate(valStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		newPots, err := k.GetPotFromIDs(ctx, value)
		if err != nil {
			panic(err)
		}
		pots = append(pots, newPots...)

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.ActivePotsResponse{Data: pots, Pagination: pageRes}, nil
}

// UpcomingPots returns scheduled pots
func (k Keeper) UpcomingPots(goCtx context.Context, req *types.UpcomingPotsRequest) (*types.UpcomingPotsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	pots := []types.Pot{}
	store := ctx.KVStore(k.storeKey)
	valStore := prefix.NewStore(store, types.KeyPrefixUpcomingPots)

	pageRes, err := query.FilteredPaginate(valStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		newPots, err := k.GetPotFromIDs(ctx, value)
		if err != nil {
			panic(err)
		}
		pots = append(pots, newPots...)

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.UpcomingPotsResponse{Data: pots, Pagination: pageRes}, nil
}

// RewardsEst returns rewards estimation at a future specific time
func (k Keeper) RewardsEst(goCtx context.Context, req *types.RewardsEstRequest) (*types.RewardsEstResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.RewardsEstResponse{Coins: k.GetRewardsEst(ctx, req.Owner, req.Locks, req.Pots, req.EndEpoch)}, nil
}
