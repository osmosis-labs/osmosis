package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) getPoolbyId(ctx sdk.Context, poolId uint64) Pool {
	store := ctx.KVStore(k.storeKey)
	pool := Pool{}
	key := types.KeyPool(poolId)
	osmoutils.MustGet(store, key, &pool)
	return pool
}

func (k Keeper) setPoolById(ctx sdk.Context, poolId uint64, pool Pool) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPool(poolId)
	osmoutils.MustSet(store, key, &pool)
}
