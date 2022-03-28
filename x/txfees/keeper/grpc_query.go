package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) FeeTokens(ctx context.Context, _ *types.QueryFeeTokensRequest) (*types.QueryFeeTokensResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	feeTokens := k.GetFeeTokens(sdkCtx)

	return &types.QueryFeeTokensResponse{FeeTokens: feeTokens}, nil
}

func (k Keeper) TxFeesByDenom(ctx context.Context, req *types.QueryTxFeesByDenomRequest) (*types.QueryTxFeesByDenomResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	baseDenom, err := k.GetBaseDenom(sdkCtx)
	if err != nil {
		return nil, err
	}

	spotPrice, err := k.CalcFeeSpotPrice(sdkCtx, req.Denom)
	if err != nil {
		return nil, err
	}

	feeToken, err := k.GetFeeToken(sdkCtx, req.GetDenom())
	if err != nil {
		return nil, err
	}

	return &types.QueryTxFeesByDenomResponse{PoolID: feeToken.PoolID, BaseDenom: baseDenom, SpotPrice: spotPrice}, nil
}

func (k Keeper) DenomPoolId(ctx context.Context, req *types.QueryDenomPoolIdRequest) (*types.QueryDenomPoolIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	feeToken, err := k.GetFeeToken(sdkCtx, req.GetDenom())
	if err != nil {
		return nil, err
	}

	return &types.QueryDenomPoolIdResponse{PoolID: feeToken.GetPoolID()}, nil
}

func (k Keeper) BaseDenom(ctx context.Context, _ *types.QueryBaseDenomRequest) (*types.QueryBaseDenomResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	baseDenom, err := k.GetBaseDenom(sdkCtx)
	if err != nil {
		return nil, err
	}

	return &types.QueryBaseDenomResponse{BaseDenom: baseDenom}, nil
}
