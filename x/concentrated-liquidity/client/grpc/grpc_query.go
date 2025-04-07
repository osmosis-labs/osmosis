package grpc

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/concentratedliquidity/v1beta1/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client/queryproto"
)

type Querier struct {
	Q client.Querier
}

var _ queryproto.QueryServer = Querier{}

func (q Querier) UserUnbondingPositions(grpcCtx context.Context,
	req *queryproto.UserUnbondingPositionsRequest,
) (*queryproto.UserUnbondingPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.UserUnbondingPositions(ctx, *req)
}

func (q Querier) UserPositions(grpcCtx context.Context,
	req *queryproto.UserPositionsRequest,
) (*queryproto.UserPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.UserPositions(ctx, *req)
}

func (q Querier) TickAccumulatorTrackers(grpcCtx context.Context,
	req *queryproto.TickAccumulatorTrackersRequest,
) (*queryproto.TickAccumulatorTrackersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TickAccumulatorTrackers(ctx, *req)
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

func (q Querier) PoolAccumulatorRewards(grpcCtx context.Context,
	req *queryproto.PoolAccumulatorRewardsRequest,
) (*queryproto.PoolAccumulatorRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.PoolAccumulatorRewards(ctx, *req)
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

func (q Querier) NumPoolPositions(grpcCtx context.Context,
	req *queryproto.NumPoolPositionsRequest,
) (*queryproto.NumPoolPositionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.NumPoolPositions(ctx, *req)
}

func (q Querier) NumNextInitializedTicks(grpcCtx context.Context,
	req *queryproto.NumNextInitializedTicksRequest,
) (*queryproto.NumNextInitializedTicksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.NumNextInitializedTicks(ctx, *req)
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

func (q Querier) IncentiveRecords(grpcCtx context.Context,
	req *queryproto.IncentiveRecordsRequest,
) (*queryproto.IncentiveRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.IncentiveRecords(ctx, *req)
}

func (q Querier) GetTotalLiquidity(grpcCtx context.Context,
	req *queryproto.GetTotalLiquidityRequest,
) (*queryproto.GetTotalLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.GetTotalLiquidity(ctx, *req)
}

func (q Querier) ClaimableSpreadRewards(grpcCtx context.Context,
	req *queryproto.ClaimableSpreadRewardsRequest,
) (*queryproto.ClaimableSpreadRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.ClaimableSpreadRewards(ctx, *req)
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

func (q Querier) CFMMPoolIdLinkFromConcentratedPoolId(grpcCtx context.Context,
	req *queryproto.CFMMPoolIdLinkFromConcentratedPoolIdRequest,
) (*queryproto.CFMMPoolIdLinkFromConcentratedPoolIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.CFMMPoolIdLinkFromConcentratedPoolId(ctx, *req)
}
