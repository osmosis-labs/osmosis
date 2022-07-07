package keeper

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the incentives module keeper providing gRPC method handlers.
type Querier struct {
	Keeper
}

// NewQuerier creates a new Querier struct.
func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// ModuleToDistributeCoins returns coins that are going to be distributed.
func (q Querier) ModuleToDistributeCoins(goCtx context.Context, _ *types.ModuleToDistributeCoinsRequest) (*types.ModuleToDistributeCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleToDistributeCoinsResponse{Coins: q.Keeper.GetModuleToDistributeCoins(ctx)}, nil
}

// ModuleDistributedCoins returns coins that have been distributed by module already.
func (q Querier) ModuleDistributedCoins(goCtx context.Context, _ *types.ModuleDistributedCoinsRequest) (*types.ModuleDistributedCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleDistributedCoinsResponse{Coins: q.Keeper.GetModuleDistributedCoins(ctx)}, nil
}

// GaugeByID takes a gaugeID and returns its respective gauge.
func (q Querier) GaugeByID(goCtx context.Context, req *types.GaugeByIDRequest) (*types.GaugeByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	gauge, err := q.Keeper.GetGaugeByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.GaugeByIDResponse{Gauge: gauge}, nil
}

// Gauges returns all upcoming and active gauges.
func (q Querier) Gauges(goCtx context.Context, req *types.GaugesRequest) (*types.GaugesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pageRes, gauges, err := q.filterByPrefixAndDenom(ctx, types.KeyPrefixGauges, "", req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.GaugesResponse{Data: gauges, Pagination: pageRes}, nil
}

// ActiveGauges returns all active gauges.
func (q Querier) ActiveGauges(goCtx context.Context, req *types.ActiveGaugesRequest) (*types.ActiveGaugesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pageRes, gauges, err := q.filterByPrefixAndDenom(ctx, types.KeyPrefixActiveGauges, "", req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.ActiveGaugesResponse{Data: gauges, Pagination: pageRes}, nil
}

// ActiveGaugesPerDenom returns all active gauges for the specified denom.
func (q Querier) ActiveGaugesPerDenom(goCtx context.Context, req *types.ActiveGaugesPerDenomRequest) (*types.ActiveGaugesPerDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pageRes, gauges, err := q.filterByPrefixAndDenom(ctx, types.KeyPrefixActiveGauges, req.Denom, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.ActiveGaugesPerDenomResponse{Data: gauges, Pagination: pageRes}, nil
}

// UpcomingGauges returns all upcoming gauges.
func (q Querier) UpcomingGauges(goCtx context.Context, req *types.UpcomingGaugesRequest) (*types.UpcomingGaugesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pageRes, gauges, err := q.filterByPrefixAndDenom(ctx, types.KeyPrefixUpcomingGauges, "", req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.UpcomingGaugesResponse{Data: gauges, Pagination: pageRes}, nil
}

// UpcomingGaugesPerDenom returns all upcoming gauges for the specified denom.
func (q Querier) UpcomingGaugesPerDenom(goCtx context.Context, req *types.UpcomingGaugesPerDenomRequest) (*types.UpcomingGaugesPerDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	pageRes, gauges, err := q.filterByPrefixAndDenom(ctx, types.KeyPrefixUpcomingGauges, req.Denom, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.UpcomingGaugesPerDenomResponse{UpcomingGauges: gauges, Pagination: pageRes}, nil
}

// RewardsEst returns rewards estimation at a future specific time (by epoch).
func (q Querier) RewardsEst(goCtx context.Context, req *types.RewardsEstRequest) (*types.RewardsEstResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	diff := req.EndEpoch - q.Keeper.GetEpochInfo(ctx).CurrentEpoch
	if diff > 365 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "end epoch out of ranges")
	}

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	locks := make([]lockuptypes.PeriodLock, 0, len(req.LockIds))
	for _, lockId := range req.LockIds {
		lock, err := q.Keeper.lk.GetLockByID(ctx, lockId)
		if err != nil {
			return nil, err
		}
		locks = append(locks, *lock)
	}

	return &types.RewardsEstResponse{Coins: q.Keeper.GetRewardsEst(ctx, owner, locks, req.EndEpoch)}, nil
}

// LockableDurations returns all of the allowed lockable durations on chain.
func (q Querier) LockableDurations(ctx context.Context, _ *types.QueryLockableDurationsRequest) (*types.QueryLockableDurationsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryLockableDurationsResponse{LockableDurations: q.Keeper.GetLockableDurations(sdkCtx)}, nil
}

// getGaugeFromIDJsonBytes returns gauges from the json bytes of gaugeIDs.
func (q Querier) getGaugeFromIDJsonBytes(ctx sdk.Context, refValue []byte) ([]types.Gauge, error) {
	gauges := []types.Gauge{}
	gaugeIDs := []uint64{}

	err := json.Unmarshal(refValue, &gaugeIDs)
	if err != nil {
		return gauges, err
	}

	for _, gaugeID := range gaugeIDs {
		gauge, err := q.Keeper.GetGaugeByID(ctx, gaugeID)
		if err != nil {
			return []types.Gauge{}, err
		}

		gauges = append(gauges, *gauge)
	}

	return gauges, nil
}

// filterByPrefixAndDenom filters gauges based on a given key prefix and denom
func (q Querier) filterByPrefixAndDenom(ctx sdk.Context, prefixType []byte, denom string, pagination *query.PageRequest) (*query.PageResponse, []types.Gauge, error) {
	gauges := []types.Gauge{}
	store := ctx.KVStore(q.Keeper.storeKey)
	valStore := prefix.NewStore(store, prefixType)

	pageRes, err := query.FilteredPaginate(valStore, pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		// this may return multiple gauges at once if two gauges start at the same time.
		// for now this is treated as an edge case that is not of importance
		newGauges, err := q.getGaugeFromIDJsonBytes(ctx, value)
		if err != nil {
			return false, err
		}
		if accumulate {
			if denom != "" {
				for _, gauge := range newGauges {
					if gauge.DistributeTo.Denom != denom {
						return false, nil
					}
					gauges = append(gauges, gauge)
				}
			} else {
				gauges = append(gauges, newGauges...)
			}
		}
		return true, nil
	})
	return pageRes, gauges, err
}
