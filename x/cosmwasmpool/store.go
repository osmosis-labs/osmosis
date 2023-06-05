package cosmwasmpool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/types"
)

// SetPool stores the given pool in state.
func (k Keeper) SetPool(ctx sdk.Context, pool types.CosmWasmExtension) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.FormatPoolsPrefix(pool.GetId()), pool.GetStoreModel())
}

// GetPoolById returns a CosmWasmExtension that corresponds to the requested pool id. Returns error if pool id is not found.
func (k Keeper) GetPoolById(ctx sdk.Context, poolId uint64) (types.CosmWasmExtension, error) {
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
