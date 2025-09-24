package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/mint/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/mint keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// Params returns params of the mint module.
func (q Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// EpochProvisions returns minter.EpochProvisions of the mint module.
func (q Querier) EpochProvisions(c context.Context, _ *types.QueryEpochProvisionsRequest) (*types.QueryEpochProvisionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := q.Keeper.GetMinter(ctx)

	return &types.QueryEpochProvisionsResponse{EpochProvisions: minter.EpochProvisions}, nil
}

// Inflation returns the current minting inflation value.
func (q Querier) Inflation(c context.Context, _ *types.QueryInflationRequest) (*types.QueryInflationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	inflation, err := q.Keeper.GetInflation(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryInflationResponse{Inflation: inflation}, nil
}
