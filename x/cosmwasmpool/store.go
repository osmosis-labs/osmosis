package cosmwasmpool

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v17/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v17/x/cosmwasmpool/types"
)

// SetPool stores the given pool in state.
func (k Keeper) SetPool(ctx sdk.Context, pool types.CosmWasmExtension) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.FormatPoolsPrefix(pool.GetId()), pool.GetStoreModel())
}

func (k Keeper) UnmarshalPool(bz []byte) (types.CosmWasmExtension, error) {
	var acc types.CosmWasmExtension
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}

// GetPoolById returns a CosmWasmExtension that corresponds to the requested pool id. Returns error if pool id is not found.
func (k Keeper) GetPoolById(ctx sdk.Context, poolId uint64) (types.CosmWasmExtension, error) {
	store := ctx.KVStore(k.storeKey)
	poolKey := types.FormatPoolsPrefix(poolId)

	bz := store.Get(poolKey)

	pool, err := k.UnmarshalPool(bz)
	if err != nil {
		return nil, err
	}

	pool.SetWasmKeeper(k.wasmKeeper)

	return pool, nil
}

// GetSerializedPools retrieves all pool objects stored in the keeper.
// Returns them as a slice of codectypes.Any for use as a response to pools queries and CLI
func (k Keeper) GetSerializedPools(ctx sdk.Context, pagination *query.PageRequest) ([]*codectypes.Any, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	poolStore := prefix.NewStore(store, types.PoolsKey)

	var anys []*codectypes.Any
	pageRes, err := query.Paginate(poolStore, pagination, func(key, _ []byte) error {
		pool := model.Pool{}
		// Get the next pool from the poolStore and pass it to the pool variable
		_, err := osmoutils.Get(poolStore, key, &pool)
		if err != nil {
			return err
		}

		// Retrieve the poolInterface from the respective pool
		poolI, err := k.GetPoolById(ctx, pool.GetId())
		if err != nil {
			return err
		}

		any, err := codectypes.NewAnyWithValue(poolI.GetStoreModel())
		if err != nil {
			return err
		}

		anys = append(anys, any)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return anys, pageRes, err
}
