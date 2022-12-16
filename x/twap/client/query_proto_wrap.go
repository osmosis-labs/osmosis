package client

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/twap"
	"github.com/osmosis-labs/osmosis/v13/x/twap/client/queryproto"
)

// This file should evolve to being code gen'd, off of `proto/twap/v1beta/query.yml`

type Querier struct {
	K twap.Keeper
}

func (q Querier) ArithmeticTwap(ctx sdk.Context,
	req queryproto.ArithmeticTwapRequest, // nolint: staticcheck
) (*queryproto.ArithmeticTwapResponse, error) {
	if req.EndTime == nil {
		req.EndTime = &time.Time{}
	}
	if (*req.EndTime == time.Time{}) {
		*req.EndTime = ctx.BlockTime()
	}

	twap, err := q.K.GetArithmeticTwap(ctx, req.PoolId, req.BaseAsset, req.QuoteAsset, req.StartTime, *req.EndTime)

	// nolint: staticcheck
	return &queryproto.ArithmeticTwapResponse{ArithmeticTwap: twap}, err
}

func (q Querier) ArithmeticTwapToNow(ctx sdk.Context,
	req queryproto.ArithmeticTwapToNowRequest, // nolint: staticcheck
) (*queryproto.ArithmeticTwapToNowResponse, error) {
	twap, err := q.K.GetArithmeticTwapToNow(ctx, req.PoolId, req.BaseAsset, req.QuoteAsset, req.StartTime)

	// nolint: staticcheck
	return &queryproto.ArithmeticTwapToNowResponse{ArithmeticTwap: twap}, err
}

func (q Querier) Params(ctx sdk.Context,
	req queryproto.ParamsRequest,
) (*queryproto.ParamsResponse, error) {
	params := q.K.GetParams(ctx)
	return &queryproto.ParamsResponse{Params: params}, nil
}
