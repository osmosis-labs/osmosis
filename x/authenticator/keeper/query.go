package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) GetAuthenticators(
	ctx context.Context,
	request *types.GetAuthenticatorsRequest,
) (*types.GetAuthenticatorsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	acc, err := sdk.AccAddressFromBech32(request.Account)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	authenticators, err := k.GetAuthenticatorDataForAccount(sdkCtx, acc)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.GetAuthenticatorsResponse{AccountAuthenticators: authenticators}, nil

}
