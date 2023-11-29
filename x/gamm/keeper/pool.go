package keeper

import (
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func (k Keeper) MarshalPool(pool poolmanagertypes.PoolI) ([]byte, error) {
	return k.cdc.MarshalInterface(pool)
}

func (k Keeper) UnmarshalPool(bz []byte) (types.CFMMPoolI, error) {
	var acc types.CFMMPoolI
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}

// GetPool returns a pool with a given id.
func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error) {
	return k.GetPoolAndPoke(ctx, poolId)
}

// GetPoolAndPoke returns a PoolI based on it's identifier if one exists. If poolId corresponds
// to a pool with weights (e.g. balancer), the weights of the pool are updated via PokePool prior to returning.
// TODO: Consider rename to GetPool due to downstream API confusion.
func (k Keeper) GetPoolAndPoke(ctx sdk.Context, poolId uint64) (types.CFMMPoolI, error) {
	store := ctx.KVStore(k.storeKey)
	poolKey := types.GetKeyPrefixPools(poolId)
	if !store.Has(poolKey) {
		return nil, types.PoolDoesNotExistError{PoolId: poolId}
	}

	bz := store.Get(poolKey)

	pool, err := k.UnmarshalPool(bz)
	if err != nil {
		return nil, err
	}

	if pokePool, ok := pool.(types.WeightedPoolExtension); ok {
		pokePool.PokePool(ctx.BlockTime())
	}

	return pool, nil
}

// Get pool and check if the pool is active, i.e. allowed to be swapped against.
func (k Keeper) getPoolForSwap(ctx sdk.Context, poolId uint64) (types.CFMMPoolI, error) {
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return &balancer.Pool{}, err
	}

	if !pool.IsActive(ctx) {
		return &balancer.Pool{}, sdkerrors.Wrapf(types.ErrPoolLocked, "swap on inactive pool")
	}
	return pool, nil
}

func (k Keeper) iterator(ctx sdk.Context, prefix []byte) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, prefix)
}

func (k Keeper) GetPoolsAndPoke(ctx sdk.Context) (res []types.CFMMPoolI, err error) {
	iter := k.iterator(ctx, types.KeyPrefixPools)
	defer iter.Close() //nolint:errcheck

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()

		pool, err := k.UnmarshalPool(bz)
		if err != nil {
			return nil, err
		}

		if pokePool, ok := pool.(types.WeightedPoolExtension); ok {
			pokePool.PokePool(ctx.BlockTime())
		}
		res = append(res, pool)
	}

	return res, nil
}

func (k Keeper) setPool(ctx sdk.Context, pool poolmanagertypes.PoolI) error {
	bz, err := k.MarshalPool(pool)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	poolKey := types.GetKeyPrefixPools(pool.GetId())
	store.Set(poolKey, bz)

	return nil
}

func (k Keeper) DeletePool(ctx sdk.Context, poolId uint64) error {
	store := ctx.KVStore(k.storeKey)
	poolKey := types.GetKeyPrefixPools(poolId)
	if !store.Has(poolKey) {
		return fmt.Errorf("pool with ID %d does not exist", poolId)
	}

	store.Delete(poolKey)
	return nil
}

// GetPoolDenom retrieves the pool based on PoolId and
// returns the coin denoms that it holds.
func (k Keeper) GetPoolDenoms(ctx sdk.Context, poolId uint64) ([]string, error) {
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return nil, err
	}

	denoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))
	return denoms, err
}

// setNextPoolId sets next pool Id.
func (k Keeper) setNextPoolId(ctx sdk.Context, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: poolId})
	store.Set(types.KeyNextGlobalPoolId, bz)
}

// Deprecated: pool id index has been moved to x/poolmanager.
func (k Keeper) GetNextPoolId(ctx sdk.Context) uint64 {
	var nextPoolId uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyNextGlobalPoolId)
	if bz == nil {
		panic(fmt.Errorf("pool has not been initialized -- Should have been done in InitGenesis"))
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.Unmarshal(bz, &val)
		if err != nil {
			panic(err)
		}

		nextPoolId = val.GetValue()
	}
	return nextPoolId
}

func (k Keeper) GetPoolType(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolType, error) {
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return -1, err
	}

	switch pool := pool.(type) {
	case *balancer.Pool:
		return poolmanagertypes.Balancer, nil
	default:
		errMsg := fmt.Sprintf("unrecognized %s pool type: %T", types.ModuleName, pool)
		return -1, sdkerrors.Wrap(sdkerrors.ErrUnpackAny, errMsg)
	}
}

// convertToCFMMPool converts PoolI to CFMMPoolI by casting the input.
// Returns the pool of the CFMMPoolI or error if the given pool does not implement
// CFMMPoolI.
func convertToCFMMPool(pool poolmanagertypes.PoolI) (types.CFMMPoolI, error) {
	cfmmPool, ok := pool.(types.CFMMPoolI)
	if !ok {
		return nil, fmt.Errorf("given pool does not implement CFMMPoolI, implements %T", pool)
	}
	return cfmmPool, nil
}
