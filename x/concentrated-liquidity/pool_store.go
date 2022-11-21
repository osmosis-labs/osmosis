package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) getPoolbyId(ctx sdk.Context, poolId uint64) (Pool, error) {
	store := ctx.KVStore(k.storeKey)
	pool := Pool{}
	key := types.KeyPool(poolId)
	found, err := osmoutils.GetIfFound(store, key, &pool)
	if err != nil {
		panic(err)
	}
	if !found {
		return Pool{}, types.PoolNotFoundError{PoolId: poolId}
	}
	return pool, nil
}

func (k Keeper) setPoolById(ctx sdk.Context, poolId uint64, pool Pool) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPool(poolId)
	osmoutils.MustSet(store, key, &pool)
}
