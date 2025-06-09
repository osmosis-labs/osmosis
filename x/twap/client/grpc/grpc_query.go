
package grpc

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/symphony/twap/v1beta1/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v27/x/twap/client"
	"github.com/osmosis-labs/osmosis/v27/x/twap/client/queryproto"
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

func (q Querier) GeometricTwapToNow(grpcCtx context.Context,
	req *queryproto.GeometricTwapToNowRequest,
) (*queryproto.GeometricTwapToNowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.GeometricTwapToNow(ctx, *req)
}

func (q Querier) GeometricTwap(grpcCtx context.Context,
	req *queryproto.GeometricTwapRequest,
) (*queryproto.GeometricTwapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.GeometricTwap(ctx, *req)
}

func (q Querier) ArithmeticTwapToNow(grpcCtx context.Context,
	req *queryproto.ArithmeticTwapToNowRequest,
) (*queryproto.ArithmeticTwapToNowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.ArithmeticTwapToNow(ctx, *req)
}

func (q Querier) ArithmeticTwap(grpcCtx context.Context,
	req *queryproto.ArithmeticTwapRequest,
) (*queryproto.ArithmeticTwapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.ArithmeticTwap(ctx, *req)
}

