package keeper

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
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
	var ownerAddress sdk.AccAddress
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 && len(req.LockIds) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	diff := req.EndEpoch - q.Keeper.GetEpochInfo(ctx).CurrentEpoch
	if diff > 365 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "end epoch out of ranges")
	}

	if len(req.Owner) != 0 {
		owner, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, err
		}
		ownerAddress = owner
	}

	locks := make([]lockuptypes.PeriodLock, 0, len(req.LockIds))
	for _, lockId := range req.LockIds {
		lock, err := q.Keeper.lk.GetLockByID(ctx, lockId)
		if err != nil {
			return nil, err
		}
		locks = append(locks, *lock)
	}

	return &types.RewardsEstResponse{Coins: q.Keeper.GetRewardsEst(ctx, ownerAddress, locks, req.EndEpoch)}, nil
}

// LockableDurations returns all of the allowed lockable durations on chain.
func (q Querier) LockableDurations(ctx context.Context, _ *types.QueryLockableDurationsRequest) (*types.QueryLockableDurationsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryLockableDurationsResponse{LockableDurations: q.Keeper.GetLockableDurations(sdkCtx)}, nil
}

// AllGroups returns all groups that exist on chain.
func (q Querier) AllGroups(goCtx context.Context, req *types.QueryAllGroupsRequest) (*types.QueryAllGroupsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groups, err := q.Keeper.GetAllGroups(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllGroupsResponse{Groups: groups}, nil
}

// AllGroupsGauges returns all group gauges that exist on chain.
func (q Querier) AllGroupsGauges(goCtx context.Context, req *types.QueryAllGroupsGaugesRequest) (*types.QueryAllGroupsGaugesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	gauges, err := q.Keeper.GetAllGroupsGauges(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllGroupsGaugesResponse{Gauges: gauges}, nil
}

// AllGroupsWithGauge returns all groups with their respective gauge that exist on chain.
func (q Querier) AllGroupsWithGauge(goCtx context.Context, req *types.QueryAllGroupsWithGaugeRequest) (*types.QueryAllGroupsWithGaugeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupsWithGauge, err := q.Keeper.GetAllGroupsWithGauge(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllGroupsWithGaugeResponse{GroupsWithGauge: groupsWithGauge}, nil
}

// GroupByGroupGaugeID retrieves a group by its associated gauge ID.
// If the group cannot be found or an error occurs during the operation, it returns an error.
func (q Querier) GroupByGroupGaugeID(goCtx context.Context, req *types.QueryGroupByGroupGaugeIDRequest) (*types.QueryGroupByGroupGaugeIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	group, err := q.Keeper.GetGroupByGaugeID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGroupByGroupGaugeIDResponse{Group: group}, nil
}

func (q Querier) CurrentWeightByGroupGaugeID(goCtx context.Context, req *types.QueryCurrentWeightByGroupGaugeIDRequest) (*types.QueryCurrentWeightByGroupGaugeIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	group, err := q.Keeper.GetGroupByGaugeID(ctx, req.GroupGaugeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	gaugeWeights, err := q.Keeper.queryWeightSplitGroup(ctx, group)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryCurrentWeightByGroupGaugeIDResponse{GaugeWeight: gaugeWeights}, nil
}

func (q Querier) Params(goCtx context.Context, req *types.ParamsRequest) (*types.ParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)
	return &types.ParamsResponse{
		Params: params,
	}, nil
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

// queryWeightSplitGroup calculates the ratio of volume for each gauge in a group since the last epoch.
// It first updates the group weights based on the pool volumes.
// Then, for each gauge in the updated group, it calculates the ratio of the gauge's current weight to the total weight of the group.
// If the total weight of the group is zero, the ratio of volume for the gauge is set to zero.
// The function returns a slice of GaugeVolume, each representing a gauge and its ratio of volume.
// It returns an error if there is an issue updating the group weights.
func (k Keeper) queryWeightSplitGroup(ctx sdk.Context, group types.Group) ([]types.GaugeWeight, error) {
	updatedGroup, err := k.calculateGroupWeights(ctx, group)
	if err != nil {
		return nil, err
	}

	gaugeVolumes := make([]types.GaugeWeight, len(updatedGroup.InternalGaugeInfo.GaugeRecords))

	for i, gaugeRecord := range updatedGroup.InternalGaugeInfo.GaugeRecords {
		if updatedGroup.InternalGaugeInfo.TotalWeight.IsZero() {
			gaugeVolumes[i] = types.GaugeWeight{
				GaugeId:     gaugeRecord.GaugeId,
				WeightRatio: osmomath.ZeroDec(),
			}
		} else {
			gaugeVolumes[i] = types.GaugeWeight{
				GaugeId:     gaugeRecord.GaugeId,
				WeightRatio: gaugeRecord.CurrentWeight.ToLegacyDec().Quo(updatedGroup.InternalGaugeInfo.TotalWeight.ToLegacyDec()),
			}
		}
	}

	return gaugeVolumes, nil
}
