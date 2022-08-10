package grpc

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/twap/client"
	"github.com/osmosis-labs/osmosis/v10/x/twap/client/queryproto"
)

// This file should evolve to being code gen'd, off of `proto/twap/v1beta/query.yml`

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
