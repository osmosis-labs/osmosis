package grpc

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/poolmanager/v1beta1/query.yml`

import (
	context "context"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"

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
	routeReq := &queryproto.EstimateSwapExactAmountOutRequest{
		Sender:   "",
		PoolId:   req.PoolId,
		TokenOut: req.TokenOut,
		Routes:   types.SwapAmountOutRoutes{{PoolId: req.PoolId, TokenInDenom: req.TokenInDenom}},
	}
	return q.Q.EstimateSwapExactAmountOut(ctx, *routeReq)
}

func (q Querier) EstimateSinglePoolSwapExactAmountIn(grpcCtx context.Context,
	req *queryproto.EstimateSinglePoolSwapExactAmountInRequest,
) (*queryproto.EstimateSwapExactAmountInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	routeReq := &queryproto.EstimateSwapExactAmountInRequest{
		Sender:  "",
		PoolId:  req.PoolId,
		TokenIn: req.TokenIn,
		Routes:  types.SwapAmountInRoutes{{PoolId: req.PoolId, TokenOutDenom: req.TokenOutDenom}},
	}

	return q.Q.EstimateSwapExactAmountIn(ctx, *routeReq)
}
