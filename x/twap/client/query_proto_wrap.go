package client

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/twap"
	"github.com/osmosis-labs/osmosis/v27/x/twap/client/queryproto"
)

// This file should evolve to being code gen'd, off of `proto/twap/v1beta/query.yml`

type Querier struct {
	K twap.Keeper
}

func (q Querier) ArithmeticTwap(ctx sdk.Context,
	req queryproto.ArithmeticTwapRequest,
) (*queryproto.ArithmeticTwapResponse, error) {
	if req.EndTime == nil {
		req.EndTime = &time.Time{}
	}
	if (*req.EndTime == time.Time{}) {
		*req.EndTime = ctx.BlockTime()
	}

	twap, err := q.K.GetArithmeticTwap(ctx, req.PoolId, req.BaseAsset, req.QuoteAsset, req.StartTime, *req.EndTime)

	return &queryproto.ArithmeticTwapResponse{ArithmeticTwap: twap}, err
}

func (q Querier) ArithmeticTwapToNow(ctx sdk.Context,
	req queryproto.ArithmeticTwapToNowRequest,
) (*queryproto.ArithmeticTwapToNowResponse, error) {
	twap, err := q.K.GetArithmeticTwapToNow(ctx, req.PoolId, req.BaseAsset, req.QuoteAsset, req.StartTime)

	return &queryproto.ArithmeticTwapToNowResponse{ArithmeticTwap: twap}, err
}

func (q Querier) GeometricTwap(ctx sdk.Context,
	req queryproto.GeometricTwapRequest,
) (*queryproto.GeometricTwapResponse, error) {
	if req.EndTime == nil {
		req.EndTime = &time.Time{}
	}
	if (*req.EndTime == time.Time{}) {
		*req.EndTime = ctx.BlockTime()
	}

	twap, err := q.K.GetGeometricTwap(ctx, req.PoolId, req.BaseAsset, req.QuoteAsset, req.StartTime, *req.EndTime)

	return &queryproto.GeometricTwapResponse{GeometricTwap: twap}, err
}

func (q Querier) GeometricTwapToNow(ctx sdk.Context,
	req queryproto.GeometricTwapToNowRequest,
) (*queryproto.GeometricTwapToNowResponse, error) {
	twap, err := q.K.GetGeometricTwapToNow(ctx, req.PoolId, req.BaseAsset, req.QuoteAsset, req.StartTime)

	return &queryproto.GeometricTwapToNowResponse{GeometricTwap: twap}, err
}

func (q Querier) Params(ctx sdk.Context,
	req queryproto.ParamsRequest,
) (*queryproto.ParamsResponse, error) {
	params := q.K.GetParams(ctx)
	return &queryproto.ParamsResponse{Params: params}, nil
}
