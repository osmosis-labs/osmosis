package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/client/queryproto"
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

func (q Querier) Pools(ctx sdk.Context,
	req queryproto.PoolsRequest,
) (*queryproto.PoolsResponse, error) {
	pools, pageResponse, err := q.K.GetSerializedPools(ctx, req.Pagination)
	if err != nil {
		return nil, err
	}
	return &queryproto.PoolsResponse{Pools: pools, Pagination: pageResponse}, nil
}

func (q Querier) ContractInfoByPoolId(ctx sdk.Context,
	req queryproto.ContractInfoByPoolIdRequest,
) (*queryproto.ContractInfoByPoolIdResponse, error) {
	pool, err := q.K.GetPoolById(ctx, req.PoolId)
	if err != nil {
		return nil, err
	}

	return &queryproto.ContractInfoByPoolIdResponse{ContractAddress: pool.GetContractAddress(), CodeId: pool.GetCodeId()}, nil
}
