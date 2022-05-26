package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

// GetParams returns the total set params
func (k Keeper) GetParams(ctx sdk.Context) (params api.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of params
func (k Keeper) SetParams(ctx sdk.Context, params api.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
