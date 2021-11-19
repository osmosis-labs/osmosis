package keeper

import (
	"context"
	"encoding/json"

	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
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

// GaugeByID returns Gauge by id
func (k Keeper) GaugeByID(goCtx context.Context, req *types.GaugeByIDRequest) (*types.GaugeByIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	gauge, err := k.GetGaugeByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.GaugeByIDResponse{Gauge: gauge}, nil
}

// Gauges returns gauges both upcoming and active
func (k Keeper) Gauges(goCtx context.Context, req *types.GaugesRequest) (*types.GaugesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	gauges := []types.Gauge{}
	store := ctx.KVStore(k.storeKey)
	valStore := prefix.NewStore(store, types.KeyPrefixGauges)

	pageRes, err := query.FilteredPaginate(valStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		newGauges, err := k.getGaugeFromIDJsonBytes(ctx, value)
		if err != nil {
			panic(err)
		}
		gauges = append(gauges, newGauges...)

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.GaugesResponse{Data: gauges, Pagination: pageRes}, nil
}

// ActiveGauges returns active gauges
func (k Keeper) ActiveGauges(goCtx context.Context, req *types.ActiveGaugesRequest) (*types.ActiveGaugesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	gauges := []types.Gauge{}
	store := ctx.KVStore(k.storeKey)
	valStore := prefix.NewStore(store, types.KeyPrefixActiveGauges)

	pageRes, err := query.FilteredPaginate(valStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		newGauges, err := k.getGaugeFromIDJsonBytes(ctx, value)
		if err != nil {
			panic(err)
		}
		gauges = append(gauges, newGauges...)

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.ActiveGaugesResponse{Data: gauges, Pagination: pageRes}, nil
}

// ActiveGaugesPerDenom returns active gauges for the specified denom
func (k Keeper) ActiveGaugesPerDenom(goCtx context.Context, req *types.ActiveGaugesPerDenomRequest) (*types.ActiveGaugesPerDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	gauges := []types.Gauge{}
	store := ctx.KVStore(k.storeKey)
	valStore := prefix.NewStore(store, types.KeyPrefixActiveGauges)

	pageRes, err := query.FilteredPaginate(valStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		activeGauges := k.GetActiveGauges(ctx)
		for _, gauge := range activeGauges {
			if gauge.DistributeTo.Denom == req.Denom {
				gauges = append(gauges, gauge)
			}
		}
		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.ActiveGaugesPerDenomResponse{Data: gauges, Pagination: pageRes}, nil
}

// UpcomingGauges returns scheduled gauges
func (k Keeper) UpcomingGauges(goCtx context.Context, req *types.UpcomingGaugesRequest) (*types.UpcomingGaugesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	gauges := []types.Gauge{}
	store := ctx.KVStore(k.storeKey)
	valStore := prefix.NewStore(store, types.KeyPrefixUpcomingGauges)

	pageRes, err := query.FilteredPaginate(valStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		newGauges, err := k.getGaugeFromIDJsonBytes(ctx, value)
		if err != nil {
			panic(err)
		}
		gauges = append(gauges, newGauges...)

		return true, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.UpcomingGaugesResponse{Data: gauges, Pagination: pageRes}, nil
}

// RewardsEst returns rewards estimation at a future specific time
func (k Keeper) RewardsEst(goCtx context.Context, req *types.RewardsEstRequest) (*types.RewardsEstResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	diff := req.EndEpoch - k.GetEpochInfo(ctx).CurrentEpoch
	if diff > 365 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "end epoch out of ranges")
	}
	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}
	locks := make([]lockuptypes.PeriodLock, 0, len(req.LockIds))
	for _, lockId := range req.LockIds {
		lock, err := k.lk.GetLockByID(ctx, lockId)
		if err != nil {
			return nil, err
		}
		locks = append(locks, *lock)
	}
	return &types.RewardsEstResponse{Coins: k.GetRewardsEst(ctx, owner, locks, req.EndEpoch)}, nil
}

func (k Keeper) LockableDurations(ctx context.Context, _ *types.QueryLockableDurationsRequest) (*types.QueryLockableDurationsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryLockableDurationsResponse{LockableDurations: k.GetLockableDurations(sdkCtx)}, nil
}

// getGaugeFromIDJsonBytes returns gauges from gauge id json bytes
func (k Keeper) getGaugeFromIDJsonBytes(ctx sdk.Context, refValue []byte) ([]types.Gauge, error) {
	gauges := []types.Gauge{}
	gaugeIDs := []uint64{}
	err := json.Unmarshal(refValue, &gaugeIDs)
	if err != nil {
		return gauges, err
	}
	for _, gaugeID := range gaugeIDs {
		gauge, err := k.GetGaugeByID(ctx, gaugeID)
		if err != nil {
			return []types.Gauge{}, err
		}
		gauges = append(gauges, *gauge)
	}
	return gauges, nil
}
