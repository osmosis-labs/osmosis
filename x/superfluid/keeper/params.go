package keeper

import (
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set params.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// GetParams returns the total set params.
func (k Keeper) GetEpochIdentifier(ctx sdk.Context) (epochIdentifier string) {
	return k.ik.GetParams(ctx).DistrEpochIdentifier
}

// SetParams sets the total set of params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
