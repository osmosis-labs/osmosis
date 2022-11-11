package client

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/twap"
	"github.com/osmosis-labs/osmosis/v12/x/twap/client/v2queryproto"
)

// This file should evolve to being code gen'd, off of `proto/twap/v1beta/query.yml`

type QuerierV2 struct {
	K twap.Keeper
}

func (q QuerierV2) ArithmeticTwap(ctx sdk.Context,
	req v2queryproto.ArithmeticTwapRequest,
) (*v2queryproto.ArithmeticTwapResponse, error) {
	if (req.EndTime == nil || *req.EndTime == time.Time{}) {
		*req.EndTime = ctx.BlockTime()
	}

	twap, err := q.K.GetArithmeticTwap(ctx, req.PoolId, req.QuoteAsset, req.BaseAsset, req.StartTime, *req.EndTime)
	return &v2queryproto.ArithmeticTwapResponse{ArithmeticTwap: twap}, err
}

func (q QuerierV2) ArithmeticTwapToNow(ctx sdk.Context,
	req v2queryproto.ArithmeticTwapToNowRequest,
) (*v2queryproto.ArithmeticTwapToNowResponse, error) {
	twap, err := q.K.GetArithmeticTwapToNow(ctx, req.PoolId, req.QuoteAsset, req.BaseAsset, req.StartTime)
	return &v2queryproto.ArithmeticTwapToNowResponse{ArithmeticTwap: twap}, err
}
