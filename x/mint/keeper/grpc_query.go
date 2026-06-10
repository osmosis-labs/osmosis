package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v31/x/mint/types"
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

// BurnedSupply returns the amount of mint-denom held in the burn address.
func (q Querier) BurnedSupply(c context.Context, _ *types.QueryBurnedRequest) (*types.QueryBurnedResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryBurnedResponse{Burned: q.Keeper.GetBurnedSupply(ctx)}, nil
}

// TotalSupply returns the total supply (minted - burned).
func (q Querier) TotalSupply(c context.Context, _ *types.QueryTotalSupplyRequest) (*types.QueryTotalSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryTotalSupplyResponse{TotalSupply: q.Keeper.GetTotalSupply(ctx)}, nil
}

// RestrictedSupply returns the supply held in restricted addresses.
func (q Querier) RestrictedSupply(c context.Context, _ *types.QueryRestrictedSupplyRequest) (*types.QueryRestrictedSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryRestrictedSupplyResponse{RestrictedSupply: q.Keeper.GetRestrictedSupply(ctx)}, nil
}

// CirculatingSupply returns the circulating supply (minted - burned - restricted).
func (q Querier) CirculatingSupply(c context.Context, _ *types.QueryCirculatingSupplyRequest) (*types.QueryCirculatingSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryCirculatingSupplyResponse{CirculatingSupply: q.Keeper.GetCirculatingSupply(ctx)}, nil
}
