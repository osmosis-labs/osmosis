package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/claim/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the mint module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// Claimable returns claimable amount per user
func (k Keeper) Claimable(
	goCtx context.Context,
	req *types.ClaimableRequest,
) (*types.ClaimableResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	coins, err := k.GetClaimable(ctx, req.Sender)
	return &types.ClaimableResponse{Coins: coins}, err
}

// Activities returns activities
func (k Keeper) Activities(
	goCtx context.Context,
	req *types.ActivitiesRequest,
) (*types.ActivitiesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	address, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, err
	}

	allActions := []types.Action{
		types.ActionAddLiquidity,
		types.ActionSwap,
		types.ActionVote,
		types.ActionDelegateStake,
	}
	completedActions := k.GetUserActions(ctx, address)
	return &types.ActivitiesResponse{
		All:       types.ActionToNames(allActions),
		Completed: types.ActionToNames(completedActions),
	}, err
}
