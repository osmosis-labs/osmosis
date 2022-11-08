package concentrated_liquidity

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) MarshalPool(pool types.PoolI) ([]byte, error) {
	return k.cdc.MarshalInterface(pool)
}

func (k Keeper) UnmarshalPool(bz []byte) (types.PoolI, error) {
	var acc types.PoolI
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}

func (k Keeper) getPoolbyId(ctx sdk.Context, poolId uint64) (types.PoolI, error) {
	store := ctx.KVStore(k.storeKey)

	key := types.KeyPool(poolId)
	if !store.Has(key) {
		return nil, fmt.Errorf("pool does not exist")
	}

	bz := store.Get(key)
	pool, err := k.UnmarshalPool(bz)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (k Keeper) setPool(ctx sdk.Context, pool types.PoolI) error {
	bz, err := k.MarshalPool(pool)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPool(pool.GetId())
	store.Set(key, bz)
	return nil
}
