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
