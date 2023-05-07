package grpc 

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/concentrated-liquidity/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/client"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/client/queryproto"
)

type Querier struct {
	Q client.Querier
}

var _ queryproto.QueryServer = Querier{}

func (q Querier) UserPositions(grpcCtx context.Context,
	req *queryproto.UserPositionsRequest,
) (*queryproto.UserPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.UserPositions(ctx, *req)
}

func (q Querier) PositionById(grpcCtx context.Context,
	req *queryproto.PositionByIdRequest,
) (*queryproto.PositionByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.PositionById(ctx, *req)
}

func (q Querier) Pools(grpcCtx context.Context,
	req *queryproto.PoolsRequest,
) (*queryproto.PoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.Pools(ctx, *req)
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

func (q Querier) LiquidityPerTickRange(grpcCtx context.Context,
	req *queryproto.LiquidityPerTickRangeRequest,
) (*queryproto.LiquidityPerTickRangeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.LiquidityPerTickRange(ctx, *req)
}

func (q Querier) LiquidityNetInDirection(grpcCtx context.Context,
	req *queryproto.LiquidityNetInDirectionRequest,
) (*queryproto.LiquidityNetInDirectionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.LiquidityNetInDirection(ctx, *req)
}

func (q Querier) ClaimableIncentives(grpcCtx context.Context,
	req *queryproto.ClaimableIncentivesRequest,
) (*queryproto.ClaimableIncentivesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.ClaimableIncentives(ctx, *req)
}

func (q Querier) ClaimableFees(grpcCtx context.Context,
	req *queryproto.ClaimableFeesRequest,
) (*queryproto.ClaimableFeesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.ClaimableFees(ctx, *req)
}

