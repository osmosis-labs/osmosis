package grpc 

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/twap/v1beta1/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v12/x/twap/client"
	"github.com/osmosis-labs/osmosis/v12/x/twap/client/v2queryproto"
)

type QuerierV2 struct {
	Q client.QuerierV2
}

var _ v2queryproto.QueryServer = QuerierV2{}


func (q QuerierV2) ArithmeticTwapToNow(grpcCtx context.Context,
	req *v2queryproto.ArithmeticTwapToNowRequest,
) (*v2queryproto.ArithmeticTwapToNowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.ArithmeticTwapToNow(ctx, *req)
}

func (q QuerierV2) ArithmeticTwap(grpcCtx context.Context,
	req *v2queryproto.ArithmeticTwapRequest,
) (*v2queryproto.ArithmeticTwapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.ArithmeticTwap(ctx, *req)
}

