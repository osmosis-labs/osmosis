package keeper

import (
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
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

func (k Keeper) GetPools(ctx sdk.Context) ([]poolmanagertypes.PoolI, error) {
	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), types.KeyPrefixPools, func(bz []byte) (poolmanagertypes.PoolI, error) {
		pool, err := k.UnmarshalPool(bz)
		if err != nil {
			return nil, err
		}

		if pokePool, ok := pool.(types.WeightedPoolExtension); ok {
			pokePool.PokePool(ctx.BlockTime())
		}

		return pool, nil
	})
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

// GetCFMMPool gets CFMMPool and checks if the pool is active, i.e. allowed to be swapped against.
// The difference from GetPools is that this function returns an error if the pool is inactive.
// Additionally, it returns x/gamm specific CFMMPool type.
func (k Keeper) GetCFMMPool(ctx sdk.Context, poolId uint64) (types.CFMMPoolI, error) {
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return &balancer.Pool{}, err
	}

	if !pool.IsActive(ctx) {
		return &balancer.Pool{}, errorsmod.Wrapf(types.ErrPoolLocked, "swap on inactive pool")
	}
	return pool, nil
}

func (k Keeper) iterator(ctx sdk.Context, prefix []byte) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, prefix)
}

func (k Keeper) GetPoolsAndPoke(ctx sdk.Context) (res []types.CFMMPoolI, err error) {
	iter := k.iterator(ctx, types.KeyPrefixPools)
	defer iter.Close()

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

// OverwritePoolV15MigrationUnsafe is a temporary method for calling from the v15 upgrade handler
// for balancer to stableswap pool migration. Do not use for other purposes.
func (k Keeper) OverwritePoolV15MigrationUnsafe(ctx sdk.Context, pool poolmanagertypes.PoolI) error {
	return k.setPool(ctx, pool)
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
// func (k Keeper) CleanupBalancerPool(ctx sdk.Context, poolIds []uint64, excludedModules []string) (err error) {
// 	pools := make(map[string]types.CFMMPoolI)
// 	totalShares := make(map[string]sdk.Int)
// 	for _, poolId := range poolIds {
// 		pool, err := k.GetPool(ctx, poolId)
// 		if err != nil {
// 			return err
// 		}
// 		shareDenom := pool.GetTotalShares().Denom
// 		pools[shareDenom] = pool
// 		totalShares[shareDenom] = pool.GetTotalShares().Amount
// 	}

// 	moduleAccounts := make(map[string]string)
// 	for _, module := range excludedModules {
// 		moduleAccounts[string(authtypes.NewModuleAddress(module))] = module
// 	}

// 	// first iterate through the share holders and burn them
// 	k.bankKeeper.IterateAllBalances(ctx, func(addr sdk.AccAddress, coin sdk.Coin) (stop bool) {
// 		if coin.Amount.IsZero() {
// 			return
// 		}

// 		pool, ok := pools[coin.Denom]
// 		if !ok {
// 			return
// 		}

// 		// track the iterated shares
// 		pool.SubTotalShares(coin.Amount)
// 		pools[coin.Denom] = pool

// 		// check if the shareholder is a module
// 		if _, ok = moduleAccounts[coin.Denom]; ok {
// 			return
// 		}

// 		// Burn the share tokens
// 		err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, sdk.Coins{coin})
// 		if err != nil {
// 			return true
// 		}

// 		err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{coin})
// 		if err != nil {
// 			return true
// 		}

// 		// Refund assets
// 		for _, asset := range pool.GetAllPoolAssets() {
// 			// lpShareEquivalentTokens = (amount in pool) * (your shares) / (total shares)
// 			lpShareEquivalentTokens := asset.Token.Amount.Mul(coin.Amount).Quo(totalShares[coin.Denom])
// 			if lpShareEquivalentTokens.IsZero() {
// 				continue
// 			}
// 			err = k.bankKeeper.SendCoins(
// 				ctx, pool.GetAddress(), addr, sdk.Coins{{asset.Token.Denom, lpShareEquivalentTokens}})
// 			if err != nil {
// 				return true
// 			}
// 		}

// 		return false
// 	})

// 	if err != nil {
// 		return err
// 	}

// 	for _, pool := range pools {
// 		// sanity check
// 		if !pool.GetTotalShares().IsZero() {
// 			panic("pool total share should be zero after cleanup")
// 		}

// 		err = k.DeletePool(ctx, pool.GetId())
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

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
	case *stableswap.Pool:
		return poolmanagertypes.Stableswap, nil
	default:
		errMsg := fmt.Sprintf("unrecognized %s pool type: %T", types.ModuleName, pool)
		return -1, errorsmod.Wrap(sdkerrors.ErrUnpackAny, errMsg)
	}
}

// GetTotalPoolLiquidity returns the coins in the pool owned by all LPs
func (k Keeper) GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error) {
	pool, err := k.GetCFMMPool(ctx, poolId)
	if err != nil {
		return nil, err
	}
	return pool.GetTotalPoolLiquidity(ctx), nil
}

// GetTotalPoolShares returns the total number of pool shares for the given pool.
func (k Keeper) GetTotalPoolShares(ctx sdk.Context, poolId uint64) (sdk.Int, error) {
	pool, err := k.GetCFMMPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	return pool.GetTotalShares(), nil
}

// setStableSwapScalingFactors sets the stable swap scaling factors.
// errors if the pool does not exist, the sender is not the scaling factor controller, or due to other
// internal errors.
func (k Keeper) setStableSwapScalingFactors(ctx sdk.Context, poolId uint64, scalingFactors []uint64, sender string) error {
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return err
	}
	stableswapPool, ok := pool.(*stableswap.Pool)
	if !ok {
		return fmt.Errorf("pool id %d is not of type stableswap pool", poolId)
	}
	if err := stableswapPool.SetScalingFactors(ctx, scalingFactors, sender); err != nil {
		return err
	}

	return k.setPool(ctx, stableswapPool)
}

// asCFMMPool converts PoolI to CFMMPoolI by casting the input.
// Returns the pool of the CFMMPoolI or error if the given pool does not implement
// CFMMPoolI.
func asCFMMPool(pool poolmanagertypes.PoolI) (types.CFMMPoolI, error) {
	cfmmPool, ok := pool.(types.CFMMPoolI)
	if !ok {
		return nil, fmt.Errorf("given pool does not implement CFMMPoolI, implements %T", pool)
	}
	return cfmmPool, nil
}
