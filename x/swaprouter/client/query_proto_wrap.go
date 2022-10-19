package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/swaprouter"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/client/queryproto"
)

// This file should evolve to being code gen'd, off of `proto/swaprouter/v1beta/query.yml`

type Querier struct {
	K swaprouter.Keeper
}

func (q Querier) Params(ctx sdk.Context,
	req queryproto.ParamsRequest,
) (*queryproto.ParamsResponse, error) {
	params := q.K.GetParams(ctx)
	return &queryproto.ParamsResponse{Params: params}, nil
}
