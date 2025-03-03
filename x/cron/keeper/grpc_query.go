package keeper

import (
	"context"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/osmosis-labs/osmosis/v29/x/cron/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = QueryServer{}

type QueryServer struct {
	Keeper
}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &QueryServer{
		Keeper: k,
	}
}

func (q QueryServer) Cron(c context.Context, req *types.QueryCronRequest) (*types.QueryCronResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	cronJob, found := q.Keeper.GetCronJob(ctx, req.Id)
	if !found {
		return nil, status.Error(codes.NotFound, "cron job not found")
	}
	return &types.QueryCronResponse{CronJob: cronJob}, nil
}

func (q QueryServer) Crons(c context.Context, req *types.QueryCronsRequest) (*types.QueryCronsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	var (
		items []types.CronJob
		ctx   = sdk.UnwrapSDKContext(c)
	)
	pagination, err := query.FilteredPaginate(
		prefix.NewStore(q.Store(ctx), types.CronJobKeyPrefix),
		req.Pagination,
		func(_, value []byte, accumulate bool) (bool, error) {
			var item types.CronJob
			if err := q.cdc.Unmarshal(value, &item); err != nil {
				return false, err
			}
			if accumulate {
				items = append(items, item)
			}
			return true, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryCronsResponse{CronJobs: items, Pagination: pagination}, nil
}

func (q QueryServer) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.Keeper.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}
