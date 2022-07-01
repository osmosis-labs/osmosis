package keeper

import (
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (k Keeper) MarshalPool(pool types.PoolI) ([]byte, error) {
	return k.cdc.MarshalInterface(pool)
}

func (k Keeper) UnmarshalPool(bz []byte) (types.PoolI, error) {
	var acc types.PoolI
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}

// GetPoolAndPoke returns a PoolI based on it's identifier if one exists. Prior
// to returning the pool, the weights of the pool are updated via PokePool.
func (k Keeper) GetPoolAndPoke(ctx sdk.Context, poolId uint64) (types.PoolI, error) {
	store := ctx.KVStore(k.storeKey)
	poolKey := types.GetKeyPrefixPools(poolId)
	if !store.Has(poolKey) {
		return nil, fmt.Errorf("pool with ID %d does not exist", poolId)
	}

	bz := store.Get(poolKey)

	pool, err := k.UnmarshalPool(bz)
	if err != nil {
		return nil, err
	}

	pool.PokePool(ctx.BlockTime())

	return pool, nil
}

// Get pool and check if the pool is active, i.e. allowed to be swapped against.
func (k Keeper) getPoolForSwap(ctx sdk.Context, poolId uint64) (types.PoolI, error) {
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

func (k Keeper) GetPoolsAndPoke(ctx sdk.Context) (res []types.PoolI, err error) {
	iter := k.iterator(ctx, types.KeyPrefixPools)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()

		pool, err := k.UnmarshalPool(bz)
		if err != nil {
			return nil, err
		}

		pool.PokePool(ctx.BlockTime())
		res = append(res, pool)
	}

	return res, nil
}

func (k Keeper) SetPool(ctx sdk.Context, pool types.PoolI) error {
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

// SetNextPoolNumber sets next pool number.
func (k Keeper) SetNextPoolNumber(ctx sdk.Context, poolNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: poolNumber})
	store.Set(types.KeyNextGlobalPoolNumber, bz)
}

// GetNextPoolNumberAndIncrement returns the next pool number, and increments the corresponding state entry.
func (k Keeper) GetNextPoolNumberAndIncrement(ctx sdk.Context) uint64 {
	var poolNumber uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyNextGlobalPoolNumber)
	if bz == nil {
		panic(fmt.Errorf("pool has not been initialized -- Should have been done in InitGenesis"))
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.Unmarshal(bz, &val)
		if err != nil {
			panic(err)
		}

		poolNumber = val.GetValue()
	}

	k.SetNextPoolNumber(ctx, poolNumber+1)
	return poolNumber
}

// set ScalingFactors in stable stableswap pools
func (k *Keeper) SetStableSwapScalingFactors(ctx sdk.Context, scalingFactors []uint64, poolId uint64, scalingFactorGovernor string) error {
	poolI, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return err
	}

	stableswapPool, ok := poolI.(*stableswap.Pool)
	if !ok {
		return types.ErrNotStableSwapPool
	}

	if scalingFactorGovernor != stableswapPool.ScalingFactorGovernor {
		return types.ErrNotScalingFactorGovernor
	}

	if len(scalingFactors) != stableswapPool.PoolLiquidity.Len() {
		return types.ErrInvalidStableswapScalingFactors
	}

	stableswapPool.ScalingFactor = scalingFactors

	if err = k.SetPool(ctx, stableswapPool); err != nil {
		return err
	}
	return nil
}
