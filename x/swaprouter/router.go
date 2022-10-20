package swaprouter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: spec and tests
func (k Keeper) RouteExactAmountIn(ctx sdk.Context) (tokenOutAmount sdk.Int, err error) {
	return sdk.ZeroInt(), nil
}

// TODO: spec and tests
func (k Keeper) RouteExactAmountOut(ctx sdk.Context) (tokenInAmount sdk.Int, err error) {
	return sdk.ZeroInt(), nil
}
