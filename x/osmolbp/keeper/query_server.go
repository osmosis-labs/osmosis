package keeper

import (
	"context"

	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

func (k Keeper) LBPs(ctx context.Context, q *api.QueryLBPs) (*api.QueryLBPsResponse, error) {
	return nil, nil
}

func (k Keeper) LBP(ctx context.Context, q *api.QueryLBP) (*api.QueryLBPResponse, error) {
	return nil, nil
}

func (k Keeper) UserPosition(ctx context.Context, q *api.QueryLBP) (*api.QueryLBPResponse, error) {
	return nil, nil
}
