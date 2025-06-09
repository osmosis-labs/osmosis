
package grpc

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/symphony/downtimedetector/v1beta1/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client/queryproto"
)

type Querier struct {
	Q client.Querier
}

var _ queryproto.QueryServer = Querier{}

func (q Querier) RecoveredSinceDowntimeOfLength(grpcCtx context.Context,
	req *queryproto.RecoveredSinceDowntimeOfLengthRequest,
) (*queryproto.RecoveredSinceDowntimeOfLengthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.RecoveredSinceDowntimeOfLength(ctx, *req)
}

