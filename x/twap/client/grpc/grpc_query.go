package grpc 

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/twap/v1beta1/query.yml`

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/twap/client"
	"github.com/osmosis-labs/osmosis/v10/x/twap/client/queryproto"
)

type Querier struct {
	Q client.Querier
}

var _ queryproto.QueryServer = Querier{}

func (q Querier) GetArithmeticTwap(grpcCtx context.Context,
	req *queryproto.GetArithmeticTwapRequest,
) (*queryproto.GetArithmeticTwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.GetArithmeticTwap(ctx, req)
}
