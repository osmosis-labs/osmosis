package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

// GetParams return all module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params, err error) {
	return k.Params.Get(ctx)
}

// SetParams sets all module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	return k.Params.Set(ctx, params)
}
