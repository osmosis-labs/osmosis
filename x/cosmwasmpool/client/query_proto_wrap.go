package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/client/queryproto"
)

// This file should evolve to being code gen'd, off of `proto/poolmanager/v1beta/query.yml`

type Querier struct {
	K cosmwasmpool.Keeper
}

func NewQuerier(k cosmwasmpool.Keeper) Querier {
	return Querier{k}
}

func (q Querier) Params(ctx sdk.Context,
	req queryproto.ParamsRequest,
) (*queryproto.ParamsResponse, error) {
	params := q.K.GetParams(ctx)
	return &queryproto.ParamsResponse{Params: params}, nil
}
