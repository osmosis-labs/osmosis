package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
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

func (k Keeper) GetAuthenticator(
	ctx context.Context,
	request *types.GetAuthenticatorRequest,
) (*types.GetAuthenticatorResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	acc, err := sdk.AccAddressFromBech32(request.Account)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	authenticator, err := k.GetSelectedAuthenticatorData(sdkCtx, acc, int(request.AuthenticatorId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.GetAuthenticatorResponse{AccountAuthenticator: authenticator}, nil
}
