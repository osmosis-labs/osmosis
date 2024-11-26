package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ibcratelimit "github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit"
	"github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/client/queryproto"
)

// This file should evolve to being code gen'd, off of `proto/twap/v1beta/query.yml`

type Querier struct {
	K ibcratelimit.ICS4Wrapper
}

func (q Querier) Params(ctx sdk.Context,
	req queryproto.ParamsRequest,
) (*queryproto.ParamsResponse, error) {
	params := q.K.GetParams(ctx)
	return &queryproto.ParamsResponse{Params: params}, nil
}
