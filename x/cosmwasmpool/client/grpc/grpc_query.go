package grpc

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/cosmwasmpool/v1beta1/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client/queryproto"
)

type Querier struct {
	Q client.Querier
}

var _ queryproto.QueryServer = Querier{}

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

func (q Querier) ContractInfoByPoolId(grpcCtx context.Context,
	req *queryproto.ContractInfoByPoolIdRequest,
) (*queryproto.ContractInfoByPoolIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.ContractInfoByPoolId(ctx, *req)
}
