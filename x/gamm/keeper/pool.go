package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
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

func (k Keeper) DeletePool(ctx sdk.Context, poolId uint64) error {
	store := ctx.KVStore(k.storeKey)
	poolKey := types.GetKeyPrefixPools(poolId)
	if !store.Has(poolKey) {
		return fmt.Errorf("pool with ID %d does not exist", poolId)
	}

	store.Delete(poolKey)
	return nil
}

// CleanupBalancerPool destructs a pool and refund all the assets according to
// the shares held by the accounts. CleanupBalancerPool should not be called during
// the chain execution time, as it iterates the entire account balances.
// TODO: once SDK v0.46.0, use https://github.com/cosmos/cosmos-sdk/pull/9611
//
// All locks on this pool share must be unlocked prior to execution. Use LockupKeeper.ForceUnlock
// on remaining locks before calling this function.
func (k Keeper) CleanupBalancerPool(ctx sdk.Context, poolIds []uint64, excludedModules []string) (err error) {
	pools := make(map[string]types.PoolI)
	totalShares := make(map[string]sdk.Int)
	for _, poolId := range poolIds {
		pool, err := k.GetPool(ctx, poolId)
		if err != nil {
			return err
		}
		shareDenom := pool.GetTotalShares().Denom
		pools[shareDenom] = pool
		totalShares[shareDenom] = pool.GetTotalShares().Amount
	}

	moduleAccounts := make(map[string]string)
	for _, module := range excludedModules {
		moduleAccounts[string(authtypes.NewModuleAddress(module))] = module
	}

	// first iterate through the share holders and burn them
	k.bankKeeper.IterateAllBalances(ctx, func(addr sdk.AccAddress, coin sdk.Coin) (stop bool) {
		if coin.Amount.IsZero() {
			return
		}

		pool, ok := pools[coin.Denom]
		if !ok {
			return
		}

		// track the iterated shares
		pool.SubTotalShares(coin.Amount)
		pools[coin.Denom] = pool

		// check if the shareholder is a module
		if _, ok = moduleAccounts[coin.Denom]; ok {
			return
		}

		// Burn the share tokens
		err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, sdk.Coins{coin})
		if err != nil {
			return true
		}

		err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{coin})
		if err != nil {
			return true
		}

		// Refund assets
		for _, asset := range pool.GetAllPoolAssets() {
			// lpShareEquivalentTokens = (amount in pool) * (your shares) / (total shares)
			lpShareEquivalentTokens := asset.Token.Amount.Mul(coin.Amount).Quo(totalShares[coin.Denom])
			if lpShareEquivalentTokens.IsZero() {
				continue
			}
			err = k.bankKeeper.SendCoins(
				ctx, pool.GetAddress(), addr, sdk.Coins{{asset.Token.Denom, lpShareEquivalentTokens}})
			if err != nil {
				return true
			}
		}

		return false
	})

	if err != nil {
		return err
	}

	for _, pool := range pools {
		// sanity check
		if !pool.GetTotalShares().IsZero() {
			panic("pool total share should be zero after cleanup")
		}

		err = k.DeletePool(ctx, pool.GetId())
		if err != nil {
			return err
		}
	}

	return nil
}

// newBalancerPool is an internal function that creates a new Balancer Pool object with the provided
// parameters, initial assets, and future governor.
func (k Keeper) newBalancerPool(ctx sdk.Context, balancerPoolParams balancer.PoolParams, assets []types.PoolAsset, futureGovernor string) (types.PoolI, error) {
	poolId := k.GetNextPoolNumberAndIncrement(ctx)

	pool, err := balancer.NewBalancerPool(poolId, balancerPoolParams, assets, futureGovernor, ctx.BlockTime())
	if err != nil {
		return nil, err
	}

	acc := k.accountKeeper.GetAccount(ctx, pool.GetAddress())
	if acc != nil {
		return nil, sdkerrors.Wrapf(types.ErrPoolAlreadyExist, "pool %d already exist", poolId)
	}

	err = k.SetPool(ctx, &pool)
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

	return &pool, nil
}

// SetNextPoolNumber sets next pool number
func (k Keeper) SetNextPoolNumber(ctx sdk.Context, poolNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: poolNumber})
	store.Set(types.KeyNextGlobalPoolNumber, bz)
}

// GetNextPoolNumberAndIncrement returns the next pool number, and increments the corresponding state entry
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
