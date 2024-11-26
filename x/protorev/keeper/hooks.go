package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

type Hooks struct {
	k Keeper
}

var (
	_ gammtypes.GammHooks = Hooks{}
)

// Create new ProtoRev hooks.
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// ----------------------------------------------------------------------------
// GAMM HOOKS
// ----------------------------------------------------------------------------

// AfterCFMMPoolCreated hook checks and potentially stores the pool via the highest liquidity method.
func (h Hooks) AfterCFMMPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	h.k.AfterPoolCreatedWithCoins(ctx, poolId)
}

// AfterJoinPool stores swaps to be checked by protorev given the coins entered into the pool.
func (h Hooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount osmomath.Int) {
	// Checked to avoid future unintended behavior based on how the hook is called
	if len(enterCoins) != 1 {
		return
	}

	h.k.StoreJoinExitPoolSwaps(ctx, sender, poolId, enterCoins[0].Denom, true)
}

// AfterExitPool stores swaps to be checked by protorev given the coins exited from the pool.
func (h Hooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount osmomath.Int, exitCoins sdk.Coins) {
	// Added due to ExitSwapShareAmountIn both calling
	// ExitPoolHook with all denoms of the pool and then also
	// Swapping which triggers the after swap hook.
	// So this filters out the exit pool hook call with all denoms
	if len(exitCoins) != 1 {
		return
	}

	h.k.StoreJoinExitPoolSwaps(ctx, sender, poolId, exitCoins[0].Denom, false)
}

// AfterCFMMSwap stores swaps to be checked by protorev given the coins swapped in the pool.
func (h Hooks) AfterCFMMSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	// Checked to avoid future unintended behavior based on how the hook is called
	if len(input) != 1 || len(output) != 1 {
		return
	}

	h.k.StoreSwap(ctx, poolId, input[0].Denom, output[0].Denom)
}

// ----------------------------------------------------------------------------
// CONCENTRATED LIQUIDITY HOOKS
// ----------------------------------------------------------------------------

// AfterConcentratedPoolCreated is a noop.
func (h Hooks) AfterConcentratedPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
}

// AfterInitialPoolPositionCreated checks and potentially stores the pool via the highest liquidity method.
func (h Hooks) AfterInitialPoolPositionCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	h.k.AfterPoolCreatedWithCoins(ctx, poolId)
}

// AfterLastPoolPositionRemoved is a noop.
func (h Hooks) AfterLastPoolPositionRemoved(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
}

// AfterConcentratedPoolSwap stores swaps to be checked by protorev given the coins swapped in the pool.
func (h Hooks) AfterConcentratedPoolSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	// Checked to avoid future unintended behavior based on how the hook is called
	if len(input) != 1 || len(output) != 1 {
		return
	}

	h.k.StoreSwap(ctx, poolId, input[0].Denom, output[0].Denom)
}

// ----------------------------------------------------------------------------
// HELPER METHODS
// ----------------------------------------------------------------------------

// StoreSwap stores a swap to be checked by protorev when attempting backruns.
func (k Keeper) StoreSwap(ctx sdk.Context, poolId uint64, tokenIn, tokenOut string) {
	swapToBackrun := types.Trade{
		Pool:     poolId,
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
	}

	if err := k.AddSwapsToSwapsToBackrun(ctx, []types.Trade{swapToBackrun}); err != nil {
		ctx.Logger().Error("Protorev error adding swap to backrun from storeSwap: " + err.Error()) // Does not return since logging is last thing in the function
	}
}

// GetComparablePoolLiquidity gets the comparable liquidity of a pool by multiplying the amounts of the pool coins.
func (k Keeper) GetComparablePoolLiquidity(ctx sdk.Context, poolId uint64) (comparableLiquidity osmomath.Int, err error) {
	coins, err := k.poolmanagerKeeper.GetTotalPoolLiquidity(ctx, poolId)
	if err != nil {
		return osmomath.Int{}, err
	}

	// Recover from overflow panic
	defer func() {
		if r := recover(); r != nil {
			comparableLiquidity = osmomath.Int{}
			err = errors.New("Int overflow in GetComparablePoolLiquidity")
		}
	}()

	comparableLiquidity = coins[0].Amount.Mul(coins[1].Amount)

	return comparableLiquidity, nil
}

// StoreJoinExitPoolSwaps stores the swaps associated with GAMM join/exit pool messages in the store, depending on if it is a join or exit.
func (k Keeper) StoreJoinExitPoolSwaps(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, denom string, isJoin bool) {
	pool, err := k.gammKeeper.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return
	}

	// Get all the pool coins and iterate to get the denoms that make up the swap
	coins := pool.GetTotalPoolLiquidity(ctx)

	// Create swaps to backrun being the join coin swapped against all other pool coins
	for _, coin := range coins {
		if coin.Denom == denom {
			continue
		}
		// Join messages swap in the denom, exit messages swap out the denom
		if isJoin {
			k.StoreSwap(ctx, poolId, denom, coin.Denom)
		} else {
			k.StoreSwap(ctx, poolId, coin.Denom, denom)
		}
	}
}

// AfterPoolCreatedWithCoins checks if the new pool should be stored as the highest liquidity pool
// for any of the base denoms, and stores it if so.
func (k Keeper) AfterPoolCreatedWithCoins(ctx sdk.Context, poolId uint64) {
	baseDenoms, err := k.GetAllBaseDenoms(ctx)
	if err != nil {
		ctx.Logger().Error("Protorev error getting base denoms in AfterCFMMPoolCreated hook: " + err.Error())
		return
	}

	baseDenomMap := make(map[string]bool)
	for _, baseDenom := range baseDenoms {
		baseDenomMap[baseDenom.Denom] = true
	}

	pool, err := k.poolmanagerKeeper.GetPool(ctx, poolId)
	if err != nil {
		ctx.Logger().Error("Protorev error getting pool in AfterCFMMPoolCreated hook: " + err.Error())
		return
	}

	denoms, err := k.poolmanagerKeeper.RouteGetPoolDenoms(ctx, poolId)
	if err != nil {
		ctx.Logger().Error("Protorev error getting pool liquidity in afterPoolCreated: " + err.Error())
		return
	}

	// Pool must be active and the number of denoms must be 2
	if pool.IsActive(ctx) && len(denoms) == 2 {
		// Check if either of the denoms are base denoms (denoms in which we store highest liquidity
		// pools for to create backrun routes). If so, we call CompareAndStorePool which will check
		// if a pool already exists for the base denom pair, and if not, stores the new pool.
		// If a pool does already exist for the base denom pair, it will compare the liquidity
		// of the new pool with the stored pool, and store the new pool if it has more liquidity.
		if _, ok := baseDenomMap[denoms[0]]; ok {
			k.CompareAndStorePool(ctx, poolId, denoms[0], denoms[1])
		}
		if _, ok := baseDenomMap[denoms[1]]; ok {
			k.CompareAndStorePool(ctx, poolId, denoms[1], denoms[0])
		}
	}
}

// CompareAndStorePool compares the liquidity of the new pool with the liquidity of the stored pool, and stores the new pool if it has more liquidity.
func (k Keeper) CompareAndStorePool(ctx sdk.Context, poolId uint64, baseDenom, otherDenom string) {
	storedPoolId, err := k.GetPoolForDenomPair(ctx, baseDenom, otherDenom)
	if err != nil {
		// Error means no pool exists for this pair, so we set it
		k.SetPoolForDenomPair(ctx, baseDenom, otherDenom, poolId)
		return
	}

	// Get comparable liquidity for the new pool
	newPoolLiquidity, err := k.GetComparablePoolLiquidity(ctx, poolId)
	if err != nil {
		ctx.Logger().Error("Protorev error getting newPoolLiquidity in compareAndStorePool: " + err.Error())
		return
	}

	// Get comparable liquidity for the stored pool
	storedPoolLiquidity, err := k.GetComparablePoolLiquidity(ctx, storedPoolId)
	if err != nil {
		ctx.Logger().Error("Protorev error getting storedPoolLiquidity in compareAndStorePool: " + err.Error())
		return
	}

	// If the new pool has more liquidity, we set it
	if newPoolLiquidity.GT(storedPoolLiquidity) {
		k.SetPoolForDenomPair(ctx, baseDenom, otherDenom, poolId)
	}
}
