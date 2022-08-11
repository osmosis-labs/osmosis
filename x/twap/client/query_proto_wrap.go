package client

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/twap"
	"github.com/osmosis-labs/osmosis/v10/x/twap/client/queryproto"
)

// This file should evolve to being code gen'd, off of `proto/twap/v1beta/query.yml`

type Querier struct {
	K twap.Keeper
}

func (q Querier) GetArithmeticTwap(ctx sdk.Context,
	req queryproto.GetArithmeticTwapRequest,
) (*queryproto.GetArithmeticTwapResponse, error) {
	if (req.EndTime == nil || *req.EndTime == time.Time{}) {
		*req.EndTime = time.Now()
	}

	twap, err := q.K.GetArithmeticTwap(ctx, req.PoolId, req.BaseAsset, req.QuoteAsset, req.StartTime, *req.EndTime)
	if err != nil {
		return nil, err
	}
	return &queryproto.GetArithmeticTwapResponse{ArithmeticTwap: twap}, nil
}
