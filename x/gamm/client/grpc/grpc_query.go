package grpc

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/gamm/v1beta1/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/client"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/client/queryproto"
)

type Querier struct {
	Q client.Querier
}

var _ queryproto.QueryServer = Querier{}

func (q Querier) Pools(grpcCtx context.Context,
	req *queryproto.QueryPoolsRequest,
) (*queryproto.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.Pools(ctx, *req)
}
func (q Querier) NumPools(grpcCtx context.Context,
	req *queryproto.QueryNumPoolsRequest,
) (*queryproto.QueryNumPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.NumPools(ctx, *req)
}
func (q Querier) SpotPrice(grpcCtx context.Context,
	req *queryproto.QuerySpotPriceRequest,
) (*queryproto.QuerySpotPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.SpotPrice(ctx, *req)
}
func (q Querier) TotalLiquidity(grpcCtx context.Context,
	req *queryproto.QueryTotalLiquidityRequest,
) (*queryproto.QueryTotalLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TotalLiquidity(ctx, *req)
}
func (q Querier) EstimateSwapExactAmountOut(grpcCtx context.Context,
	req *queryproto.QuerySwapExactAmountOutRequest,
) (*queryproto.QuerySwapExactAmountOutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateSwapExactAmountOut(ctx, *req)
}
func (q Querier) Pool(grpcCtx context.Context,
	req *queryproto.QueryPoolRequest,
) (*queryproto.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.Pool(ctx, *req)
}
func (q Querier) PoolParams(grpcCtx context.Context,
	req *queryproto.QueryPoolParamsRequest,
) (*queryproto.QueryPoolParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.PoolParams(ctx, *req)
}
func (q Querier) TotalPoolLiquidity(grpcCtx context.Context,
	req *queryproto.QueryTotalPoolLiquidityRequest,
) (*queryproto.QueryTotalPoolLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TotalPoolLiquidity(ctx, *req)
}
func (q Querier) TotalShares(grpcCtx context.Context,
	req *queryproto.QueryTotalSharesRequest,
) (*queryproto.QueryTotalSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TotalShares(ctx, *req)
}
func (q Querier) EstimateSwapExactAmountIn(grpcCtx context.Context,
	req *queryproto.QuerySwapExactAmountInRequest,
) (*queryproto.QuerySwapExactAmountInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateSwapExactAmountIn(ctx, *req)
}
