package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = keeper{}

func (k keeper) Pool(
	ctx context.Context,
	req *types.QueryPoolRequest,
) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.poolService.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPoolResponse{Pool: pool}, nil
}

func (k keeper) SwapFee(
	ctx context.Context,
	req *types.QuerySwapFeeRequest,
) (*types.QuerySwapFeeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	swapFee, err := k.poolService.GetSwapFee(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QuerySwapFeeResponse{SwapFee: swapFee}, nil
}

func (k keeper) ShareInfo(
	ctx context.Context,
	req *types.QueryShareInfoRequest,
) (*types.QueryShareInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	shareInfo, err := k.poolService.GetShareInfo(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryShareInfoResponse{ShareInfo: shareInfo}, nil
}

func (k keeper) TokenBalance(
	ctx context.Context,
	req *types.QueryTokenBalanceRequest,
) (*types.QueryTokenBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	tokens, err := k.poolService.GetTokenBalance(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTokenBalanceResponse{Tokens: tokens}, nil
}

func (k keeper) SpotPrice(
	ctx context.Context,
	req *types.QuerySpotPriceRequest,
) (*types.QuerySpotPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if req.TokenIn == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid tokenIn")
	}
	if req.TokenOut == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid tokenOut")
	}

	spotPrice, err := k.GetSpotPrice(sdkCtx, req.PoolId, req.TokenIn, req.TokenOut)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QuerySpotPriceResponse{SpotPrice: spotPrice}, nil
}
