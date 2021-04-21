package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/claim/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

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

// Withdrawable returns withdrawable amount per user
func (k Keeper) Withdrawable(
	goCtx context.Context,
	req *types.WithdrawableRequest,
) (*types.WithdrawableResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	coins, err := k.GetWithdrawableByActivity(ctx, req.Sender)
	return &types.WithdrawableResponse{Coins: coins}, err
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
	withdrawnActions := k.GetWithdrawnActions(ctx, address)
	return &types.ActivitiesResponse{
		All:       types.ActionToNames(allActions),
		Completed: types.ActionToNames(completedActions),
		Withdrawn: types.ActionToNames(withdrawnActions),
	}, err
}
