package cosmwasmpool

import (
	"bytes"

	"cosmossdk.io/store/prefix"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v29/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v29/x/cosmwasmpool/types"
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

var orderKeyBytes = []byte{0, 6, 111}
var orderIdBytes = []byte{123, 34, 116}

func (k Keeper) GetOrderbookOrdersRaw(ctx sdk.Context, poolId uint64) ([][]byte, error) {
	pool, err := k.GetPoolById(ctx, poolId)
	if err != nil {
		return nil, err
	}
	// TODO: Verify the provided address is a cosmwasm pool

	contractAddress := sdk.MustAccAddressFromBech32(pool.GetContractAddress())

	orders := [][]byte{}
	k.wasmKeeper.IterateContractState(ctx, contractAddress, func(key, value []byte) bool {
		if !bytes.HasPrefix(key, orderKeyBytes) {
			return false
		}
		if bytes.HasPrefix(value, orderIdBytes) {
			orders = append(orders, value)
			return false
		}
		return false
	})

	return orders, nil
}
