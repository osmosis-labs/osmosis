package keeper

import (
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v32/x/pool-incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetDistrInfo returns the distribution records. Since minted coins are no longer allocated to
// gauges, the records are informational only.
func (k Keeper) GetDistrInfo(ctx sdk.Context) types.DistrInfo {
	store := ctx.KVStore(k.storeKey)
	distrInfo := types.DistrInfo{}
	osmoutils.MustGet(store, types.DistrInfoKey, &distrInfo)
	return distrInfo
}

func (k Keeper) SetDistrInfo(ctx sdk.Context, distrInfo types.DistrInfo) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.DistrInfoKey, &distrInfo)
}
