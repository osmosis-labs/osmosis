
package grpcv2

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/poolmanager/v2/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v26/x/poolmanager/client"
	"github.com/osmosis-labs/osmosis/v26/x/poolmanager/client/queryprotov2"
)

type Querier struct {
	Q client.QuerierV2
}

var _ queryprotov2.QueryServer = Querier{}

func (q Querier) SpotPriceV2(grpcCtx context.Context,
	req *queryprotov2.SpotPriceRequest,
) (*queryprotov2.SpotPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.SpotPriceV2(ctx, *req)
}

func (q Querier) IsAffiliated(grpcCtx context.Context,
	req *queryprotov2.IsAffiliatedRequest,
) (*queryprotov2.IsAffiliatedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	isAffilated, err := q.Q.K.IsAffiliated(ctx, sdk.AccAddress(req.Address))
	if err != nil {
		return nil, err
	}
	return &queryprotov2.IsAffiliatedResponse{
		IsAffiliated: isAffilated,
	}, nil
}

func (q Querier) RevenueShareSummary(grpcCtx context.Context,
	req *queryprotov2.RevenueShareSummaryRequest,
) (*queryprotov2.RevenueShareSummaryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	revenueShareSummary, err := q.Q.K.GetRevenueShareSummary(ctx, sdk.AccAddress(req.Address))
	if err != nil {
		return nil, err
	}
	return revenueShareSummary, nil
}

func (q Querier) RevenueShareLeaderboard(grpcCtx context.Context,
	req *queryprotov2.RevenueShareLeaderboardRequest,
) (*queryprotov2.RevenueShareLeaderboardResponse, error) {
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	leaderboard, err := q.Q.K.GetRevenueShareLeaderboard(ctx)
	if err != nil {
		return nil, err
	}
	return leaderboard, nil
}
