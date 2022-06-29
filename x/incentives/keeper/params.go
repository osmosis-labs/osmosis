package keeper

import (
	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Returns all of the parameters in the incentive module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// Sets all of the parameters in the incentive module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
