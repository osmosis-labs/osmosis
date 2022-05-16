package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	incentivetypes "github.com/osmosis-labs/osmosis/v8/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v8/x/pool-incentives/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) GaugeIds(ctx context.Context, req *types.QueryGaugeIdsRequest) (*types.QueryGaugeIdsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	lockableDurations := k.GetLockableDurations(sdkCtx)

	gaugeIdsWithDuration := make([]*types.QueryGaugeIdsResponse_GaugeIdWithDuration, len(lockableDurations))

	for i, duration := range lockableDurations {
		gaugeId, err := k.GetPoolGaugeId(sdkCtx, req.PoolId, duration)

		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		gaugeIdsWithDuration[i] = &types.QueryGaugeIdsResponse_GaugeIdWithDuration{
			GaugeId:  gaugeId,
			Duration: duration,
		}
	}

	return &types.QueryGaugeIdsResponse{GaugeIdsWithDuration: gaugeIdsWithDuration}, nil
}

func (k Keeper) DistrInfo(ctx context.Context, _ *types.QueryDistrInfoRequest) (*types.QueryDistrInfoResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryDistrInfoResponse{DistrInfo: k.GetDistrInfo(sdkCtx)}, nil
}

func (k Keeper) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryParamsResponse{Params: k.GetParams(sdkCtx)}, nil
}

func (k Keeper) LockableDurations(ctx context.Context, _ *types.QueryLockableDurationsRequest) (*types.QueryLockableDurationsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryLockableDurationsResponse{LockableDurations: k.GetLockableDurations(sdkCtx)}, nil
}

func (k Keeper) IncentivizedPools(ctx context.Context, _ *types.QueryIncentivizedPoolsRequest) (*types.QueryIncentivizedPoolsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	lockableDurations := k.GetLockableDurations(sdkCtx)

	distrInfo := k.GetDistrInfo(sdkCtx)

	// While there are exceptions, typically the number of incentivizedPools equals to the number of incentivized gauges / number of lockable durations.
	incentivizedPools := make([]types.IncentivizedPool, 0, len(distrInfo.Records)/len(lockableDurations))

	for _, record := range distrInfo.Records {
		for _, lockableDuration := range lockableDurations {
			poolId, err := k.GetPoolIdFromGaugeId(sdkCtx, record.GaugeId, lockableDuration)
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

// ExternalIncentiveGauges iterates over all gauges, returns gauges externally incentivized,
// excluding default gagues created with pool
func (k Keeper) ExternalIncentiveGauges(ctx context.Context, req *types.QueryExternalIncentiveGaugesRequest) (*types.QueryExternalIncentiveGaugesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte("pool-incentives"))

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	// map true to default gauges created with pool
	poolGaugeIds := make(map[uint64]bool)
	for ; iterator.Valid(); iterator.Next() {
		poolGaugeIds[sdk.BigEndianToUint64(iterator.Value())] = true
	}

	// iterate over all gauges, exclude default created gauges, leaving externally incentivized gauges
	allGauges := k.GetAllGauges(sdkCtx)
	gauges := []incentivetypes.Gauge{}
	for _, gauge := range allGauges {
		if _, ok := poolGaugeIds[gauge.Id]; !ok {
			gauges = append(gauges, gauge)
		}
	}

	return &types.QueryExternalIncentiveGaugesResponse{Data: gauges}, nil
}
