package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/osmosis-labs/osmosis/v13/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// TODO spec and tests
func (k Keeper) InitializePool(ctx sdk.Context, pool swaproutertypes.PoolI, creatorAddress sdk.AccAddress) error {
	poolId := pool.GetId()

	// Add the share token's meta data to the bank keeper.
	poolShareBaseDenom := types.GetPoolShareDenom(poolId)
	poolShareDisplayDenom := fmt.Sprintf("GAMM-%d", poolId)
	k.bankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
		Description: fmt.Sprintf("The share token of the gamm pool %d", poolId),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    poolShareBaseDenom,
				Exponent: 0,
				Aliases: []string{
					"attopoolshare",
				},
			},
			{
				Denom:    poolShareDisplayDenom,
				Exponent: types.OneShareExponent,
				Aliases:  nil,
			},
		},
		Base:    poolShareBaseDenom,
		Display: poolShareDisplayDenom,
	})

	// Mint the initial pool shares share token to the sender
	err := k.MintPoolShareToAccount(ctx, pool, creatorAddress, pool.GetTotalShares())
	if err != nil {
		return err
	}

	k.RecordTotalLiquidityIncrease(ctx, pool.GetTotalPoolLiquidity(ctx))

	return k.setPool(ctx, pool)
}

func (k Keeper) MarshalPool(pool swaproutertypes.PoolI) ([]byte, error) {
	return k.cdc.MarshalInterface(pool)
}

func (k Keeper) UnmarshalPool(bz []byte) (types.TraditionalAmmInterface, error) {
	var acc types.TraditionalAmmInterface
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}

// GetPool returns a pool with a given id.
func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (swaproutertypes.PoolI, error) {
	return k.getPoolForSwap(ctx, poolId)
}

// GetPoolAndPoke returns a PoolI based on it's identifier if one exists. If poolId corresponds
// to a pool with weights (e.g. balancer), the weights of the pool are updated via PokePool prior to returning.
// TODO: Consider rename to GetPool due to downstream API confusion.
func (k Keeper) GetPoolAndPoke(ctx sdk.Context, poolId uint64) (types.TraditionalAmmInterface, error) {
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
func (k Keeper) getPoolForSwap(ctx sdk.Context, poolId uint64) (types.TraditionalAmmInterface, error) {
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

func (k Keeper) GetPoolsAndPoke(ctx sdk.Context) (res []types.TraditionalAmmInterface, err error) {
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

func (k Keeper) setPool(ctx sdk.Context, pool swaproutertypes.PoolI) error {
	bz, err := k.MarshalPool(pool)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	poolKey := types.GetKeyPrefixPools(pool.GetId())
	store.Set(poolKey, bz)

	return nil
}

// initializePoolId initializes pool id to 0.
func (k Keeper) initializePoolId(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	poolId := &gogotypes.UInt64Value{Value: 0}
	osmoutils.MustSet(store, types.KeyNextGlobalPoolId, poolId)
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
// 	pools := make(map[string]types.PoolI)
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

// GetNextPoolId returns the next pool Id.
// TODO: remove after concentrated-liquidity upgrade.
func (k Keeper) GetNextPoolId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextPoolId := gogotypes.UInt64Value{}
	osmoutils.MustGet(store, types.KeyNextGlobalPoolId, &nextPoolId)
	return nextPoolId.Value
}

// SetNextPoolId sets the next pool Id.
// TODO: remove after concentrated-liquidity upgrade.
func (k Keeper) SetNextPoolId(ctx sdk.Context, nextPoolId uint64) {
	store := ctx.KVStore(k.storeKey)
	nextPoolIdState := gogotypes.UInt64Value{Value: nextPoolId}
	osmoutils.MustSet(store, types.KeyNextGlobalPoolId, &nextPoolIdState)
}

func (k Keeper) GetPoolType(ctx sdk.Context, poolId uint64) (string, error) {
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return "", err
	}

	switch pool := pool.(type) {
	case *balancer.Pool:
		return balancer.PoolTypeName, nil
	case *stableswap.Pool:
		return stableswap.PoolTypeName, nil
	default:
		errMsg := fmt.Sprintf("unrecognized %s pool type: %T", types.ModuleName, pool)
		return "", sdkerrors.Wrap(sdkerrors.ErrUnpackAny, errMsg)
	}
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
