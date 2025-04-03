package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/pool-incentives keeper providing gRPC
// method handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// Params return pool-incentives module params.
func (q Querier) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryParamsResponse{Params: q.Keeper.GetParams(sdkCtx)}, nil
}
