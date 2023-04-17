package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
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

// AfterCFMMPoolCreated hook is a noop.
func (h Hooks) AfterCFMMPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	baseDenoms, err := h.k.GetAllBaseDenoms(ctx)
	if err != nil {
		ctx.Logger().Error("Protorev error getting base denoms in AfterCFMMPoolCreated hook", err)
		return
	}

	baseDenomMap := make(map[string]bool)
	for _, baseDenom := range baseDenoms {
		baseDenomMap[baseDenom.Denom] = true
	}

	pool, err := h.k.gammKeeper.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		ctx.Logger().Error("Protorev error getting pool in AfterCFMMPoolCreated hook", err)
		return
	}

	coins := pool.GetTotalPoolLiquidity(ctx)

	// Pool must be active and the number of coins must be 2
	if pool.IsActive(ctx) && len(coins) == 2 {
		tokenA := coins[0]
		tokenB := coins[1]

		liquidity := tokenA.Amount.Mul(tokenB.Amount)

		if baseDenomMap[tokenA.Denom] {
			h.k.compareAndStorePool(ctx, poolId, liquidity, tokenA.Denom, tokenB.Denom)
		}
		if baseDenomMap[tokenB.Denom] {
			h.k.compareAndStorePool(ctx, poolId, liquidity, tokenB.Denom, tokenA.Denom)
		}
	}
}

// AfterJoinPool hook is a noop.
func (h Hooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
	// TODO: Probably generalize to allow for more join denoms
	// Right now, the main message only allows for a single join denom
	// But I imagine we will want to allow for multiple join denoms in the future
	// Since the function allows for multile join denoms, just not the main interface
	joinDenom := enterCoins[0].Denom // As of v15, it is safe to assume only one input token

	h.k.storeJoinExitPoolMsgs(ctx, sender, poolId, joinDenom, true)
}

// AfterExitPool hook is a noop.
func (h Hooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) {
	fmt.Println("AfterExitPool hook is a noop in Protorev.")

	// TODO: Probably generalize to allow for more join denoms
	// Right now, the main message only allows for a single exit denom
	// But I imagine we will want to allow for multiple exit denoms in the future
	// Since the function allows for multile exit denoms, just not the main interface
	exitDenom := exitCoins[0].Denom // As of v15, it is safe to assume only one output token

	h.k.storeJoinExitPoolMsgs(ctx, sender, poolId, exitDenom, false)
}

// AfterCFMMSwap hook is a noop.
func (h Hooks) AfterCFMMSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	fmt.Println("AfterCFMMSwap hook is a noop in Protorev.")

	swapToBackrun := types.Trade{
		Pool:     poolId,
		TokenIn:  input[0].Denom,  // As of v15, it is safe to assume only one input token
		TokenOut: output[0].Denom, // As of v15, it is safe to assume only one output token
	}

	if err := h.k.AddSwapsToSwapsToBackrun(ctx, []types.Trade{swapToBackrun}); err != nil {
		ctx.Logger().Error("Protorev error adding swap to backrun from AfterCFMMSwap hook", err)
	}
}

// AfterConcentratedPoolCreated creates a single gauge for the concentrated liquidity pool.
func (h Hooks) AfterConcentratedPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	fmt.Println("AfterConcentratedPoolCreated hook is a noop in Protorev.")
}

// AfterInitialPoolPositionCreated is a noop.
func (h Hooks) AfterInitialPoolPositionCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	fmt.Println("AfterInitialPoolPositionCreated hook is a noop in Protorev.")
}

// AfterLastPoolPositionRemoved is a noop.
func (h Hooks) AfterLastPoolPositionRemoved(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	fmt.Println("AfterLastPoolPositionRemoved hook is a noop in Protorev.")
}

// AfterConcentratedPoolSwap is a noop.
func (h Hooks) AfterConcentratedPoolSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	fmt.Println("AfterConcentratedPoolSwap hook is a noop in Protorev.")
}

func (k Keeper) compareAndStorePool(ctx sdk.Context, poolId uint64, liquidity sdk.Int, baseDenom, otherDenom string) {
	storedPoolId, err := k.GetPoolForDenomPair(ctx, baseDenom, otherDenom)
	if err != nil {
		// Error means no pool exists for this pair, so we set it
		k.SetPoolForDenomPair(ctx, baseDenom, otherDenom, poolId)
		return
	}

	storedPool, err := k.gammKeeper.GetPoolAndPoke(ctx, storedPoolId)
	if err != nil {
		ctx.Logger().Error("Protorev error getting storedPool in AfterCFMMPoolCreated hook", err)
		return
	}

	storedPoolCoins := storedPool.GetTotalPoolLiquidity(ctx)
	storedPoolLiquidity := storedPoolCoins[0].Amount.Mul(storedPoolCoins[1].Amount)

	// If the new pool has more liquidity, we set it
	if liquidity.GT(storedPoolLiquidity) {
		k.SetPoolForDenomPair(ctx, baseDenom, otherDenom, poolId)
	}
}

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
