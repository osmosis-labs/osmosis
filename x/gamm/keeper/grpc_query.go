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

func (k keeper) Pools(
	ctx context.Context,
	req *types.QueryPoolsRequest,
) (*types.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pools, err := k.poolService.GetPools(sdkCtx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPoolsResponse{Pools: pools}, nil
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

func (k keeper) MaxSwappableLP(ctx context.Context, req *types.QueryMaxSwappableLPRequest) (*types.QueryMaxSwappableLPResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	maxLP, err := k.GetMaxSwappableLP(sdkCtx, req.PoolId, req.Tokens)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryMaxSwappableLPResponse{MaxLP: maxLP}, nil
}

func (k keeper) EstimateSwapExactAmountIn(ctx context.Context, req *types.QuerySwapExactAmountInRequest) (*types.QuerySwapExactAmountInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid bech32 address")
	}

	tokenIn, err := sdk.ParseCoinNormalized(req.TokenIn)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid coin format")
	}

	// TODO: Max price를 제대로 넣으려면 256비트 int의 최대값을 구해야하는데...
	tokenAmountOut, spotPriceAfter, err := k.SwapExactAmountIn(sdkCtx, sender, req.PoolId, tokenIn, req.TokenOutDenom, sdk.ZeroInt(), sdk.NewIntWithDecimal(1000000, 18))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QuerySwapExactAmountInResponse{
		TokenAmountOut: tokenAmountOut.String(),
		SpotPriceAfter: spotPriceAfter.String(),
	}, nil
}

func (k keeper) EstimateSwapExactAmountOut(ctx context.Context, req *types.QuerySwapExactAmountOutRequest) (*types.QuerySwapExactAmountOutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid bech32 address")
	}

	tokenOut, err := sdk.ParseCoinNormalized(req.TokenOut)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid coin format")
	}

	// TODO: Max price를 제대로 넣으려면 256비트 int의 최대값을 구해야하는데...
	tokenAmountIn, spotPriceAfter, err := k.SwapExactAmountOut(sdkCtx, sender, req.PoolId, req.TokenInDenom, sdk.NewIntWithDecimal(1000000, 18), tokenOut, sdk.NewIntWithDecimal(1000000, 18))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QuerySwapExactAmountOutResponse{
		TokenAmountIn:  tokenAmountIn.String(),
		SpotPriceAfter: spotPriceAfter.String(),
	}, nil
}
