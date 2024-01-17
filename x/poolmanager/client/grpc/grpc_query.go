
package grpc

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/poolmanager/v1beta1/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v21/x/poolmanager/client"
	"github.com/osmosis-labs/osmosis/v21/x/poolmanager/client/queryproto"
)

type Querier struct {
	Q client.Querier
}

var _ queryproto.QueryServer = Querier{}

func (q Querier) TradingPairTakerFee(grpcCtx context.Context,
	req *queryproto.TradingPairTakerFeeRequest,
) (*queryproto.TradingPairTakerFeeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TradingPairTakerFee(ctx, *req)
}

func (q Querier) TotalVolumeForPool(grpcCtx context.Context,
	req *queryproto.TotalVolumeForPoolRequest,
) (*queryproto.TotalVolumeForPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TotalVolumeForPool(ctx, *req)
}

func (q Querier) TotalPoolLiquidity(grpcCtx context.Context,
	req *queryproto.TotalPoolLiquidityRequest,
) (*queryproto.TotalPoolLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TotalPoolLiquidity(ctx, *req)
}

func (q Querier) TotalLiquidity(grpcCtx context.Context,
	req *queryproto.TotalLiquidityRequest,
) (*queryproto.TotalLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TotalLiquidity(ctx, *req)
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

func (q Querier) ListPoolsByDenom(grpcCtx context.Context,
	req *queryproto.ListPoolsByDenomRequest,
) (*queryproto.ListPoolsByDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.ListPoolsByDenom(ctx, *req)
}

func (q Querier) EstimateTradeBasedOnPriceImpact(grpcCtx context.Context,
	req *queryproto.EstimateTradeBasedOnPriceImpactRequest,
) (*queryproto.EstimateTradeBasedOnPriceImpactResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateTradeBasedOnPriceImpact(ctx, *req)
}

func (q Querier) EstimateSwapExactAmountOutWithPrimitiveTypes(grpcCtx context.Context,
	req *queryproto.EstimateSwapExactAmountOutWithPrimitiveTypesRequest,
) (*queryproto.EstimateSwapExactAmountOutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateSwapExactAmountOutWithPrimitiveTypes(ctx, *req)
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

func (q Querier) EstimateSwapExactAmountInWithPrimitiveTypes(grpcCtx context.Context,
	req *queryproto.EstimateSwapExactAmountInWithPrimitiveTypesRequest,
) (*queryproto.EstimateSwapExactAmountInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateSwapExactAmountInWithPrimitiveTypes(ctx, *req)
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

