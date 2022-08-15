package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/gamm/types"
)

// SetParams sets the total set of params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.setParams(ctx, params)
}

// SetPool adds an existing pool to the keeper store.
func (k Keeper) SetPool(ctx sdk.Context, pool types.PoolI) error {
	return k.setPool(ctx, pool)
}

func (k Keeper) GetNextPoolNumberAndIncrement(ctx sdk.Context) uint64 {
	return k.getNextPoolNumberAndIncrement(ctx)
}
