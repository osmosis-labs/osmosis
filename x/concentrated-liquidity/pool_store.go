package concentrated_liquidity

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) marshalPool(pool types.ConcentratedPoolExtension) ([]byte, error) {
	return k.cdc.MarshalInterface(pool)
}

func (k Keeper) unmarshalPool(bz []byte) (types.ConcentratedPoolExtension, error) {
	var acc types.ConcentratedPoolExtension
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}

func (k Keeper) getPoolById(ctx sdk.Context, poolId uint64) (types.ConcentratedPoolExtension, error) {
	store := ctx.KVStore(k.storeKey)

	key := types.KeyPool(poolId)
	if !store.Has(key) {
		return nil, fmt.Errorf("pool does not exist")
	}

	bz := store.Get(key)
	pool, err := k.unmarshalPool(bz)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (k Keeper) setPool(ctx sdk.Context, pool types.ConcentratedPoolExtension) error {
	bz, err := k.marshalPool(pool)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPool(pool.GetId())
	store.Set(key, bz)
	return nil
}
