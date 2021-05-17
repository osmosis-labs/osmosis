package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/pool-incentives/types"
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
