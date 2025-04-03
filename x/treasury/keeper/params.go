package keeper

import (
	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WindowShort is a short period window for moving average
func (k Keeper) WindowShort(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyWindowShort, &res)
	return
}

// WindowLong is a long period window for moving average
func (k Keeper) WindowLong(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyWindowLong, &res)
	return
}

// WindowProbation is a period of time to prevent updates
func (k Keeper) WindowProbation(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyWindowProbation, &res)
	return
}

// GetParams returns the total set of treasury parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSetIfExists(ctx, &params)
	return params
}

// SetParams sets the total set of treasury parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
