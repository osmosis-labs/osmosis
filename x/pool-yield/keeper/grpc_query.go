package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/pool-yield/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) FarmIds(ctx context.Context, req *types.QueryFarmIdsRequest) (*types.QueryFarmIdsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	lockableDurations := k.GetLockableDurations(sdkCtx)

	farmIdsWithDuration := make([]*types.QueryFarmIdsResponse_FarmIdWithDuration, len(lockableDurations))

	for i, duration := range lockableDurations {
		farmId, err := k.GetPoolFarmId(sdkCtx, req.PoolId, duration)

		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		farmIdsWithDuration[i] = &types.QueryFarmIdsResponse_FarmIdWithDuration{
			FarmId:   farmId,
			Duration: duration,
		}
	}

	return &types.QueryFarmIdsResponse{FarmIdsWithDuration: farmIdsWithDuration}, nil
}
