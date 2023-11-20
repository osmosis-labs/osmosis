package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/client/queryproto"
	keeper "github.com/osmosis-labs/osmosis/v15/x/poolmanager/keeper"
)

// This file should evolve to being code gen'd, off of `proto/poolmanager/v1beta/query.yml`

type Querier struct {
	K keeper.Keeper
}

func NewQuerier(k keeper.Keeper) Querier {
	return Querier{k}
}

func (q Querier) Params(ctx sdk.Context,
	req queryproto.ParamsRequest,
) (*queryproto.ParamsResponse, error) {
	params := q.K.GetParams(ctx)
	return &queryproto.ParamsResponse{Params: params}, nil
}
