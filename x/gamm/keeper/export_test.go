package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// SetPool adds an existing pool to the keeper store.
func (k Keeper) SetPool(ctx sdk.Context, pool types.PoolI) error {
	return k.setPool(ctx, pool)
}

// SetNumPools sets the number of pools to a given number.
func (k Keeper) SetPoolCount(ctx sdk.Context, poolCount uint64) {
	k.initializePoolCount(ctx)
	for i := uint64(0); i < poolCount; i++ {
		k.incrementPoolCount(ctx)
	}
}
