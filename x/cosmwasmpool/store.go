package cosmwasmpool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"
)

func (k Keeper) SetPool(ctx sdk.Context, pool types.CosmWasmExtension) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.FormatPoolsPrefix(pool.GetId()), pool.GetStoreModel())
}

// getPoolById returns a concentratedPoolExtension that corresponds to the requested pool id. Returns error if pool id is not found.
func (k Keeper) getPoolById(ctx sdk.Context, poolId uint64) (types.CosmWasmExtension, error) {
	store := ctx.KVStore(k.storeKey)
	pool := model.CosmWasmPool{}
	key := types.FormatPoolsPrefix(poolId)
	found, err := osmoutils.Get(store, key, &pool)
	if err != nil {
		panic(err)
	}
	if !found {
		return nil, types.PoolNotFoundError{PoolId: poolId}
	}
	return &model.Pool{
		CosmWasmPool: pool,
		WasmKeeper:   k.wasmKeeper,
	}, nil
}
