package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/c-osmosis/osmosis/x/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Pool(
	ctx context.Context,
	req *types.QueryPoolRequest,
) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	switch pool := pool.(type) {
	case *types.PoolAccount:
		return &types.QueryPoolResponse{Pool: *pool}, nil
	default:
		return nil, status.Error(codes.Internal, "invalid type of pool account")
	}
}

func (k Keeper) Pools(
	ctx context.Context,
	req *types.QueryPoolsRequest,
) (*types.QueryPoolsResponse, error) {
	panic("implement me")
}

func (k Keeper) PoolParams(ctx context.Context, req *types.QueryPoolParamsRequest) (*types.QueryPoolParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPoolParamsResponse{
		Params: pool.GetPoolParams(),
	}, nil
}

func (k Keeper) TotalShare(ctx context.Context, req *types.QueryTotalShareRequest) (*types.QueryTotalShareResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTotalShareResponse{
		TotalShare: pool.GetTotalShare(),
	}, nil
}

func (k Keeper) Records(ctx context.Context, req *types.QueryRecordsRequest) (*types.QueryRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.GetPool(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryRecordsResponse{
		Records: pool.GetAllRecords(),
	}, nil
}

func (k Keeper) SpotPrice(ctx context.Context, request *types.QuerySpotPriceRequest) (*types.QuerySpotPriceResponse, error) {
	panic("implement me")
}

func (k Keeper) EstimateSwapExactAmountIn(ctx context.Context, request *types.QuerySwapExactAmountInRequest) (*types.QuerySwapExactAmountInResponse, error) {
	panic("implement me")
}

func (k Keeper) EstimateSwapExactAmountOut(ctx context.Context, request *types.QuerySwapExactAmountOutRequest) (*types.QuerySwapExactAmountOutResponse, error) {
	panic("implement me")
}
