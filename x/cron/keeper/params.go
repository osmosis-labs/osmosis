package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v29/x/cron/types"
)

// GetParams returns the current x/cron module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	k.paramstore.GetParamSet(ctx, &p)
	return
}

// SetParams sets the x/cron module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.paramstore.SetParamSet(ctx, &p)
}
