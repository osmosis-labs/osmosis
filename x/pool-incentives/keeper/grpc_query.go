package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	incentivetypes "github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/pool-incentives keeper providing gRPC
// method handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

func (q Querier) GaugeIds(ctx context.Context, req *types.QueryGaugeIdsRequest) (*types.QueryGaugeIdsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	lockableDurations := q.Keeper.GetLockableDurations(sdkCtx)
	distrInfo := q.Keeper.GetDistrInfo(sdkCtx)
	gaugeIdsWithDuration := make([]*types.QueryGaugeIdsResponse_GaugeIdWithDuration, len(lockableDurations))

	totalWeightDec := distrInfo.TotalWeight.ToDec()
	incentivePercentage := sdk.NewDec(0)
	percentMultiplier := sdk.NewInt(100)

	for i, duration := range lockableDurations {
		gaugeId, err := q.Keeper.GetPoolGaugeId(sdkCtx, req.PoolId, duration)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		for _, record := range distrInfo.Records {
			if record.GaugeId == gaugeId {
				// Pool incentive % = (gauge_id_weight / sum_of_all_pool_gauge_weight) * 100
				incentivePercentage = record.Weight.ToDec().Quo(totalWeightDec).MulInt(percentMultiplier)
			}
		}

		gaugeIdsWithDuration[i] = &types.QueryGaugeIdsResponse_GaugeIdWithDuration{
			GaugeId:                  gaugeId,
			Duration:                 duration,
			GaugeIncentivePercentage: incentivePercentage.String(),
		}
	}
	return &types.QueryGaugeIdsResponse{GaugeIdsWithDuration: gaugeIdsWithDuration}, nil
}

func (q Querier) DistrInfo(ctx context.Context, _ *types.QueryDistrInfoRequest) (*types.QueryDistrInfoResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryDistrInfoResponse{DistrInfo: q.Keeper.GetDistrInfo(sdkCtx)}, nil
}

func (q Querier) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryParamsResponse{Params: q.Keeper.GetParams(sdkCtx)}, nil
}

func (q Querier) LockableDurations(ctx context.Context, _ *types.QueryLockableDurationsRequest) (*types.QueryLockableDurationsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryLockableDurationsResponse{LockableDurations: q.Keeper.GetLockableDurations(sdkCtx)}, nil
}

func (q Querier) IncentivizedPools(ctx context.Context, _ *types.QueryIncentivizedPoolsRequest) (*types.QueryIncentivizedPoolsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	lockableDurations := q.Keeper.GetLockableDurations(sdkCtx)
	distrInfo := q.Keeper.GetDistrInfo(sdkCtx)

	// While there are exceptions, typically the number of incentivizedPools
	// equals to the number of incentivized gauges / number of lockable durations.
	incentivizedPools := make([]types.IncentivizedPool, 0, len(distrInfo.Records)/len(lockableDurations))

	for _, record := range distrInfo.Records {
		for _, lockableDuration := range lockableDurations {
			poolId, err := q.Keeper.GetPoolIdFromGaugeId(sdkCtx, record.GaugeId, lockableDuration)
			if err == nil {
				incentivizedPool := types.IncentivizedPool{
					PoolId:           poolId,
					LockableDuration: lockableDuration,
					GaugeId:          record.GaugeId,
				}

				incentivizedPools = append(incentivizedPools, incentivizedPool)
			}
		}
	}

	return &types.QueryIncentivizedPoolsResponse{
		IncentivizedPools: incentivizedPools,
	}, nil
}

// ExternalIncentiveGauges iterates over all gauges, returns gauges externally
// incentivized, excluding default gauges created with pool.
func (q Querier) ExternalIncentiveGauges(ctx context.Context, req *types.QueryExternalIncentiveGaugesRequest) (*types.QueryExternalIncentiveGaugesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(q.Keeper.storeKey)
	prefixStore := prefix.NewStore(store, []byte("pool-incentives"))

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	// map true to default gauges created with pool
	poolGaugeIds := make(map[uint64]bool)
	for ; iterator.Valid(); iterator.Next() {
		poolGaugeIds[sdk.BigEndianToUint64(iterator.Value())] = true
	}

	// iterate over all gauges, exclude default created gauges, leaving externally incentivized gauges
	allGauges := q.Keeper.GetAllGauges(sdkCtx)
	gauges := []incentivetypes.Gauge{}
	for _, gauge := range allGauges {
		if _, ok := poolGaugeIds[gauge.Id]; !ok {
			gauges = append(gauges, gauge)
		}
	}

	return &types.QueryExternalIncentiveGaugesResponse{Data: gauges}, nil
}
