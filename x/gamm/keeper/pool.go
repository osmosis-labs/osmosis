package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/v3/x/gamm/types"
)

func (k Keeper) MarshalPool(pool types.PoolI) ([]byte, error) {
	return k.cdc.MarshalInterface(pool)
}

func (k Keeper) UnmarshalPool(bz []byte) (types.PoolI, error) {
	var acc types.PoolI
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}

func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (types.PoolI, error) {
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

	pool.PokeTokenWeights(ctx.BlockTime())

	return pool, nil
}

func (k Keeper) iterator(ctx sdk.Context, prefix []byte) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, prefix)
}

func (k Keeper) GetPools(ctx sdk.Context) (res []types.PoolI, err error) {
	iter := k.iterator(ctx, types.KeyPrefixPools)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()

		pool, err := k.UnmarshalPool(bz)
		if err != nil {
			return nil, err
		}

		pool.PokeTokenWeights(ctx.BlockTime())

		res = append(res, pool)
	}

	return
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

// newPool is an internal function that creates a new Pool object with the provided
// parameters, initial assets, and future governor.
func (k Keeper) newPool(ctx sdk.Context, poolParams types.PoolParams, assets []types.PoolAsset, futureGovernor string) (types.PoolI, error) {
	poolId := k.GetNextPoolNumber(ctx)

	pool, err := types.NewPool(poolId, poolParams, assets, futureGovernor, ctx.BlockTime())
	if err != nil {
		return nil, err
	}

	acc := k.accountKeeper.GetAccount(ctx, pool.GetAddress())
	if acc != nil {
		return nil, sdkerrors.Wrapf(types.ErrPoolAlreadyExist, "pool %d already exist", poolId)
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return nil, err
	}

	// Create and save corresponding module account to the account keeper
	acc = k.accountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(
		authtypes.NewBaseAccountWithAddress(
			pool.GetAddress(),
		),
		pool.GetAddress().String(),
	))
	k.accountKeeper.SetAccount(ctx, acc)

	return pool, nil
}

// SetNextPoolNumber sets next pool number
func (k Keeper) SetNextPoolNumber(ctx sdk.Context, poolNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&gogotypes.UInt64Value{Value: poolNumber})
	store.Set(types.KeyNextGlobalPoolNumber, bz)
}

// GetNextPoolNumber returns the next pool number
func (k Keeper) GetNextPoolNumber(ctx sdk.Context) uint64 {
	var poolNumber uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyNextGlobalPoolNumber)
	if bz == nil {
		// initialize the pool id numbers
		poolNumber = 1
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.UnmarshalBinaryBare(bz, &val)
		if err != nil {
			panic(err)
		}

		poolNumber = val.GetValue()
	}

	k.SetNextPoolNumber(ctx, poolNumber+1)
	return poolNumber
}

func (k Keeper) getPoolAndInOutAssets(
	ctx sdk.Context, poolId uint64,
	tokenInDenom string,
	tokenOutDenom string) (
	pool types.PoolI,
	inAsset types.PoolAsset,
	outAsset types.PoolAsset,
	err error,
) {
	pool, err = k.GetPool(ctx, poolId)
	if err != nil {
		return
	}

	inAsset, err = pool.GetPoolAsset(tokenInDenom)
	if err != nil {
		return
	}

	outAsset, err = pool.GetPoolAsset(tokenOutDenom)
	return
}
