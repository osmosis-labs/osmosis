package client

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/client/queryproto"
)

// This file should evolve to being code gen'd, off of `proto/swaprouter/v1beta/query.yml`

type Querier struct {
}

// nolint: unused
var sdkIntMaxValue = sdk.NewInt(0)

func init() {
	maxInt := big.NewInt(2)
	maxInt = maxInt.Exp(maxInt, big.NewInt(256), nil)

	_sdkIntMaxValue, ok := sdk.NewIntFromString(maxInt.Sub(maxInt, big.NewInt(1)).String())
	if !ok {
		panic("Failed to calculate the max value of sdk.Int")
	}

	sdkIntMaxValue = _sdkIntMaxValue
}

func (q Querier) Params(ctx sdk.Context,
	req queryproto.ParamsRequest,
) (*queryproto.ParamsResponse, error) {
	panic("not implemented")
}

// EstimateSwapExactAmountIn estimates input token amount for a swap.
func (q Querier) EstimateSwapExactAmountIn(ctx sdk.Context, req queryproto.EstimateSwapExactAmountInRequest) (*queryproto.EstimateSwapExactAmountInResponse, error) {
	panic("not implemented")
}

// EstimateSwapExactAmountOut estimates token output amount for a swap.
func (q Querier) EstimateSwapExactAmountOut(ctx sdk.Context, req queryproto.EstimateSwapExactAmountOutRequest) (*queryproto.EstimateSwapExactAmountOutResponse, error) {
	panic("not implemented")
}

// NumPools returns total number of pools.
func (q Querier) NumPools(ctx sdk.Context, _ queryproto.NumPoolsRequest) (*queryproto.NumPoolsResponse, error) {
	panic("not implemented")
}
