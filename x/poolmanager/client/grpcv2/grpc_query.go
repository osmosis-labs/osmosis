package grpcv2

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/poolmanager/v2/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryprotov2"
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
