package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
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
	h.k.afterPoolCreated(ctx, poolId)
}

// AfterJoinPool stores swaps to be checked by protorev given the coins entered into the pool.
func (h Hooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
	// Checked to avoid future unintended behavior based on how the hook is called
	if len(enterCoins) != 1 {
		return
	}

	h.k.storeJoinExitPoolMsgs(ctx, sender, poolId, enterCoins[0].Denom, true)
}

// AfterExitPool stores swaps to be checked by protorev given the coins exited from the pool.
func (h Hooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) {
	// Added due to ExitSwapShareAmountIn both calling
	// ExitPoolHook with all denoms of the pool and then also
	// Swapping which triggers the after swap hook.
	// So this filters out the exit pool hook call with all denoms
	if len(exitCoins) != 1 {
		return
	}

	h.k.storeJoinExitPoolMsgs(ctx, sender, poolId, exitCoins[0].Denom, false)
}

// AfterCFMMSwap stores swaps to be checked by protorev given the coins swapped in the pool.
func (h Hooks) AfterCFMMSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	// Checked to avoid future unintended behavior based on how the hook is called
	if len(input) != 1 || len(output) != 1 {
		return
	}

	h.k.storeSwap(ctx, poolId, input[0].Denom, output[0].Denom)
}

// ----------------------------------------------------------------------------
// CONCENTRATED LIQUIDITY HOOKS
// ----------------------------------------------------------------------------

// AfterConcentratedPoolCreated checks and potentially stores the pool via the highest liquidity method.
func (h Hooks) AfterConcentratedPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	h.k.afterPoolCreated(ctx, poolId)
}

// AfterInitialPoolPositionCreated is a noop.
func (h Hooks) AfterInitialPoolPositionCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
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

	h.k.storeSwap(ctx, poolId, input[0].Denom, output[0].Denom)
}

// ----------------------------------------------------------------------------
// HELPER METHODS
// ----------------------------------------------------------------------------

// afterPoolCreated checks if the new pool should be stored as the highest liquidity pool
// for any of the base denoms, and stores it if so.
func (k Keeper) afterPoolCreated(ctx sdk.Context, poolId uint64) {
	baseDenoms, err := k.GetAllBaseDenoms(ctx)
	if err != nil {
		ctx.Logger().Error("Protorev error getting base denoms in AfterCFMMPoolCreated hook", err)
		return
	}

	baseDenomMap := make(map[string]bool)
	for _, baseDenom := range baseDenoms {
		baseDenomMap[baseDenom.Denom] = true
	}

	pool, err := k.poolmanagerKeeper.GetPool(ctx, poolId)
	if err != nil {
		ctx.Logger().Error("Protorev error getting pool in AfterCFMMPoolCreated hook", err)
		return
	}

	// Handle concentrated pools differently since they can have 0 liquidity
	var coins sdk.Coins
	if pool.GetType() == poolmanagertypes.Concentrated {
		clPool, ok := pool.(cltypes.ConcentratedPoolExtension)
		if !ok {
			ctx.Logger().Error("Protorev error casting pool to concentrated pool in AfterCFMMPoolCreated hook")
			return
		}

		coins, err = k.poolmanagerKeeper.GetTotalPoolLiquidity(ctx, poolId)
		if err != nil {
			ctx.Logger().Error("Protorev error getting total pool liquidity in after swap hook", err)
			return
		}

		if coins == nil {
			coins = sdk.Coins{sdk.NewCoin(clPool.GetToken0(), sdk.NewInt(0)), sdk.NewCoin(clPool.GetToken1(), sdk.NewInt(0))}
		}
	} else {
		coins, err = k.poolmanagerKeeper.GetTotalPoolLiquidity(ctx, poolId)
		if err != nil {
			ctx.Logger().Error("Protorev error getting total pool liquidity in after pool created hook", err)
			return
		}
	}

	// Pool must be active and the number of coins must be 2
	if pool.IsActive(ctx) && len(coins) == 2 {
		tokenA := coins[0]
		tokenB := coins[1]

		if baseDenomMap[tokenA.Denom] {
			k.compareAndStorePool(ctx, poolId, tokenA, tokenB)
		}
		if baseDenomMap[tokenB.Denom] {
			k.compareAndStorePool(ctx, poolId, tokenB, tokenA)
		}
	}
}

// storeSwap stores a swap to be checked by protorev when attempting backruns.
func (k Keeper) storeSwap(ctx sdk.Context, poolId uint64, tokenIn, tokenOut string) {
	swapToBackrun := types.Trade{
		Pool:     poolId,
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
	}

	if err := k.AddSwapsToSwapsToBackrun(ctx, []types.Trade{swapToBackrun}); err != nil {
		ctx.Logger().Error("Protorev error adding swap to backrun from storeSwap", err)
	}
}

// compareAndStorePool compares the liquidity of the new pool with the liquidity of the stored pool, and stores the new pool if it has more liquidity.
func (k Keeper) compareAndStorePool(ctx sdk.Context, poolId uint64, baseToken, otherToken sdk.Coin) {
	storedPoolId, err := k.GetPoolForDenomPair(ctx, baseToken.Denom, otherToken.Denom)
	if err != nil {
		// Error means no pool exists for this pair, so we set it
		k.SetPoolForDenomPair(ctx, baseToken.Denom, otherToken.Denom, poolId)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			ctx.Logger().Error("Protorev error recovering from panic in AfterCFMMPoolCreated hook, likely an overflow error", r)
		}
	}()

	liquidity := baseToken.Amount.Mul(otherToken.Amount)

	storedPool, err := k.gammKeeper.GetPoolAndPoke(ctx, storedPoolId)
	if err != nil {
		ctx.Logger().Error("Protorev error getting storedPool in AfterCFMMPoolCreated hook", err)
		return
	}

	storedPoolCoins := storedPool.GetTotalPoolLiquidity(ctx)
	storedPoolLiquidity := storedPoolCoins[0].Amount.Mul(storedPoolCoins[1].Amount)

	// If the new pool has more liquidity, we set it
	if liquidity.GT(storedPoolLiquidity) {
		k.SetPoolForDenomPair(ctx, baseToken.Denom, otherToken.Denom, poolId)
	}
}

// storeJoinExitPoolMsgs stores the join/exit pool messages in the store, depending on if it is a join or exit.
func (k Keeper) storeJoinExitPoolMsgs(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, denom string, isJoin bool) {
	pool, err := k.gammKeeper.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return
	}

	// Get all the pool coins and iterate to get the denoms that make up the swap
	coins := pool.GetTotalPoolLiquidity(ctx)

	// Create swaps to backrun being the join coin swapped against all other pool coins
	swapsToBackrun := make([]types.Trade, 0)
	for _, coin := range coins {
		if coin.Denom == denom {
			continue
		}
		// Join messages swap in the denom, exit messages swap out the denom
		if isJoin {
			swapsToBackrun = append(swapsToBackrun, types.Trade{
				Pool:     poolId,
				TokenIn:  denom,
				TokenOut: coin.Denom})
		} else {
			swapsToBackrun = append(swapsToBackrun, types.Trade{
				Pool:     poolId,
				TokenIn:  coin.Denom,
				TokenOut: denom})
		}
	}

	if err := k.AddSwapsToSwapsToBackrun(ctx, swapsToBackrun); err != nil {
		ctx.Logger().Error("Protorev error adding swaps to backrun from AfterJoin/ExitPool hook", err)
	}
}
