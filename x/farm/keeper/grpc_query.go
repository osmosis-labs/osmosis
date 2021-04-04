package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/farm/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) PendingRewards(ctx context.Context, req *types.QueryPendingRewardsRequest) (*types.QueryPendingRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	decRewards, err := k.CalculatePendingRewards(sdkCtx, req.FarmId, address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPendingRewardsResponse{Rewards: decRewards}, nil
}
