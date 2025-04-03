package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"
)

// querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over q
type querier struct {
	Keeper
}

// NewQuerier returns an implementation of the market QueryServer interface
// for the provided Keeper.
func NewQuerier(keeper Keeper) types.QueryServer {
	return &querier{Keeper: keeper}
}

var _ types.QueryServer = querier{}

// Params queries params of distribution module
func (q querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}

// TaxRate return the current tax rate
func (q querier) TaxRate(c context.Context, _ *types.QueryTaxRateRequest) (*types.QueryTaxRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryTaxRateResponse{TaxRate: q.GetTaxRate(ctx)}, nil
}
