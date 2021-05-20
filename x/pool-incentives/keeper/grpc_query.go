package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/pool-incentives/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) PotIds(ctx context.Context, req *types.QueryPotIdsRequest) (*types.QueryPotIdsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	lockableDurations := k.GetLockableDurations(sdkCtx)

	potIdsWithDuration := make([]*types.QueryPotIdsResponse_PotIdWithDuration, len(lockableDurations))

	for i, duration := range lockableDurations {
		potId, err := k.GetPoolPotId(sdkCtx, req.PoolId, duration)

		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		potIdsWithDuration[i] = &types.QueryPotIdsResponse_PotIdWithDuration{
			PotId:    potId,
			Duration: duration,
		}
	}

	return &types.QueryPotIdsResponse{PotIdsWithDuration: potIdsWithDuration}, nil
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

	// While there are exceptions, typically the number of incentivizedPools equals to the number of incentivized pots / number of lockable durations.
	incentivizedPools := make([]types.IncentivizedPool, 0, len(distrInfo.Records)/len(lockableDurations))

	for _, record := range distrInfo.Records {
		for _, lockableDuration := range lockableDurations {
			poolId, err := k.GetPoolIdFromPotId(sdkCtx, record.PotId, lockableDuration)
			if err == nil {
				incentivizedPool := types.IncentivizedPool{
					PoolId:           poolId,
					LockableDuration: lockableDuration,
					PotId:            record.PotId,
				}

				incentivizedPools = append(incentivizedPools, incentivizedPool)
			}
		}

	}

	return &types.QueryIncentivizedPoolsResponse{
		IncentivizedPools: incentivizedPools,
	}, nil
}
