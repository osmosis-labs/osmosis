package grpc 

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/poolmanager/v1beta1/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/client"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/client/queryproto"
)

type Querier struct {
	Q client.Querier
}

var _ queryproto.QueryServer = Querier{}

func (q Querier) TotalPoolLiquidity(grpcCtx context.Context,
	req *queryproto.TotalPoolLiquidityRequest,
) (*queryproto.TotalPoolLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TotalPoolLiquidity(ctx, *req)
}

func (q Querier) SpotPrice(grpcCtx context.Context,
	req *queryproto.SpotPriceRequest,
) (*queryproto.SpotPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.SpotPrice(ctx, *req)
}

func (q Querier) Pool(grpcCtx context.Context,
	req *queryproto.PoolRequest,
) (*queryproto.PoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.Pool(ctx, *req)
}

func (q Querier) Params(grpcCtx context.Context,
	req *queryproto.ParamsRequest,
) (*queryproto.ParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.Params(ctx, *req)
}

func (q Querier) NumPools(grpcCtx context.Context,
	req *queryproto.NumPoolsRequest,
) (*queryproto.NumPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.NumPools(ctx, *req)
}

func (q Querier) EstimateSwapExactAmountOut(grpcCtx context.Context,
	req *queryproto.EstimateSwapExactAmountOutRequest,
) (*queryproto.EstimateSwapExactAmountOutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateSwapExactAmountOut(ctx, *req)
}

func (q Querier) EstimateSwapExactAmountIn(grpcCtx context.Context,
	req *queryproto.EstimateSwapExactAmountInRequest,
) (*queryproto.EstimateSwapExactAmountInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateSwapExactAmountIn(ctx, *req)
}

func (q Querier) EstimateSinglePoolSwapExactAmountOut(grpcCtx context.Context,
	req *queryproto.EstimateSinglePoolSwapExactAmountOutRequest,
) (*queryproto.EstimateSwapExactAmountOutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateSinglePoolSwapExactAmountOut(ctx, *req)
}

func (q Querier) EstimateSinglePoolSwapExactAmountIn(grpcCtx context.Context,
	req *queryproto.EstimateSinglePoolSwapExactAmountInRequest,
) (*queryproto.EstimateSwapExactAmountInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateSinglePoolSwapExactAmountIn(ctx, *req)
}

func (q Querier) AllPools(grpcCtx context.Context,
	req *queryproto.AllPoolsRequest,
) (*queryproto.AllPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.AllPools(ctx, *req)
}

