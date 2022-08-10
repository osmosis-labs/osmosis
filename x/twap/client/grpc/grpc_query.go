package grpc

import (
	context "context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v10/x/twap"
	"github.com/osmosis-labs/osmosis/v10/x/twap/client/grpc/grpcproto"
)

// This file should evolve to being code gen'd, off of `proto/twap/v1beta/query.yml`

type Querier struct {
	K twap.Keeper
}

var _ grpcproto.QueryServer = Querier{}

func (q Querier) GetArithmeticTwap(grpcCtx context.Context,
	req *grpcproto.GetArithmeticTwapRequest,
) (*grpcproto.GetArithmeticTwapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if (req.EndTime == nil || *req.EndTime == time.Time{}) {
		*req.EndTime = time.Now()
	}

	ctx := sdk.UnwrapSDKContext(grpcCtx)
	twap, err := q.K.GetArithmeticTwap(ctx, req.PoolId, req.BaseAsset, req.QuoteAsset, req.StartTime, *req.EndTime)
	if err != nil {
		return nil, err
	}
	return &grpcproto.GetArithmeticTwapResponse{ArithmeticTwap: twap}, nil
}
