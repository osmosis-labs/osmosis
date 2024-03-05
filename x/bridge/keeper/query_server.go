package keeper

import (
	"context"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

func (q queryServer) Params(ctx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	params, err := q.Keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.ParamsResponse{Params: params}, nil
}
