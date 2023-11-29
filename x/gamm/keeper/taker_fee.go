package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

const (
	defaultTakerFeeDenom = "udym"
)

// possibilites for takerFeeCoin:
// 1. DYM -> taker fee is already in DYM
// 2. any other denom -> DYM (taker fee will be swapped and burned)
// 3. Base denom -> find pool with DYM
// 4. any other denom -> make swap on first pool (to swap for base denom) than find pool with DYM
func (k Keeper) chargeTakerFeeSwapAmountOut(ctx sdk.Context, takerFeeCoin sdk.Coin, sender sdk.AccAddress, outRoutes []poolmanagertypes.SwapAmountOutRoute, denomOut string) error {
	if takerFeeCoin.Denom == defaultTakerFeeDenom {
		return k.swapAndBurn(ctx, sender, nil, takerFeeCoin)
	}

	//transcode the outRoutes to inRoutes
	firstPool := poolmanagertypes.SwapAmountInRoute{}
	firstPool.PoolId = outRoutes[0].PoolId
	if len(outRoutes) > 1 {
		firstPool.TokenOutDenom = outRoutes[1].TokenInDenom
	} else {
		firstPool.TokenOutDenom = denomOut
	}

	return k.swapAndBurn(ctx, sender, []poolmanagertypes.SwapAmountInRoute{firstPool}, takerFeeCoin)
}

// possibilites for takerFeeCoin:
// 1. DYM -> taker fee is already in DYM
// 2. any other denom -> DYM (taker fee will be swapped and burned)
// 3. Base denom -> find pool with DYM
// 4. any other denom -> make swap on first pool (to swap for base denom) than find pool with DYM
func (k Keeper) chargeTakerFeeSwapAmountIn(ctx sdk.Context, takerFeeCoin sdk.Coin, sender sdk.AccAddress, routes []poolmanagertypes.SwapAmountInRoute) error {
	if takerFeeCoin.Denom == defaultTakerFeeDenom {
		return k.swapAndBurn(ctx, sender, nil, takerFeeCoin)
	}

	return k.swapAndBurn(ctx, sender, []poolmanagertypes.SwapAmountInRoute{routes[0]}, takerFeeCoin)
}

// swapAndBurn swaps the taker fee coin to DYM and then burns it.
// if routes is nil, it directly burns the taker fee coin witout swaps
// if routes is not nil, it uses the first route to swap the taker fee coin to base denom
// if the first route's output denom is not DYM but other base denom, it finds the pool with this base denom and DYM
func (k Keeper) swapAndBurn(ctx sdk.Context, sender sdk.AccAddress, routes []poolmanagertypes.SwapAmountInRoute, tokenIn sdk.Coin) error {
	if len(routes) == 0 {
		return k.burnTakerFee(ctx, tokenIn, sender)
	}

	var routeForTakerFee []poolmanagertypes.SwapAmountInRoute
	firstPool := routes[0]

	// Do the swap of this fee token denom to base denom.
	if firstPool.TokenOutDenom == defaultTakerFeeDenom {
		routeForTakerFee = []poolmanagertypes.SwapAmountInRoute{firstPool}
	} else {
		params := k.GetParams(ctx)
		//if tokenIn is base denom, we need to find the pool with DYM
		isBaseDenom, _ := params.PoolCreationFee.Find(tokenIn.Denom)
		if isBaseDenom {
			route, err := k.findPoolWithTakerFeeDenom(ctx, tokenIn.Denom)
			if err != nil {
				ctx.Logger().Error("failed to find swapping route to DYM", "error", err)
				return k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(tokenIn), sender)
			}
			routeForTakerFee = []poolmanagertypes.SwapAmountInRoute{route}
		} else {
			//If swap needed, add the first pool to the route (to swap for base denom)
			route, err := k.findPoolWithTakerFeeDenom(ctx, firstPool.TokenOutDenom)
			if err != nil {
				ctx.Logger().Error("failed to find swapping route to DYM", "error", err)
				return k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(tokenIn), sender)
			}
			routeForTakerFee = []poolmanagertypes.SwapAmountInRoute{firstPool, route}
		}
	}

	//double check the denom before burning
	if routeForTakerFee[len(routeForTakerFee)-1].TokenOutDenom != defaultTakerFeeDenom {
		ctx.Logger().Error("wrong route to burn Taker Fee", "tokenIn", tokenIn, "routes", routeForTakerFee)
		return k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(tokenIn), sender)
	}

	minAmountOut := sdk.ZeroInt()
	burnTokens := sdk.Coin{}
	out, err := k.poolManager.RouteExactAmountIn(ctx, sender, routeForTakerFee, tokenIn, minAmountOut)
	if err != nil {
		return err
	}
	burnTokens.Amount = out
	burnTokens.Denom = defaultTakerFeeDenom

	return k.burnTakerFee(ctx, burnTokens, sender)
}

// Returns remaining amount in to swap, and takerFeeCoins.
// returns (1 - takerFee) * tokenIn, takerFee * tokenIn
func (k Keeper) SubTakerFee(tokenIn sdk.Coin, takerFee sdk.Dec) (sdk.Coin, sdk.Coin) {
	amountInAfterSubTakerFee := sdk.NewDecFromInt(tokenIn.Amount).MulTruncate(sdk.OneDec().Sub(takerFee))
	tokenInAfterSubTakerFee := sdk.NewCoin(tokenIn.Denom, amountInAfterSubTakerFee.TruncateInt())
	takerFeeCoin := sdk.NewCoin(tokenIn.Denom, tokenIn.Amount.Sub(tokenInAfterSubTakerFee.Amount))
	return tokenInAfterSubTakerFee, takerFeeCoin
}

// here we need the output to be (tokenIn / (1 - takerFee), takerFee * tokenIn)
func (k Keeper) AddTakerFee(tokenIn sdk.Coin, takerFee sdk.Dec) (sdk.Coin, sdk.Coin) {
	amountInAfterAddTakerFee := sdk.NewDecFromInt(tokenIn.Amount).Quo(sdk.OneDec().Sub(takerFee))
	tokenInAfterAddTakerFee := sdk.NewCoin(tokenIn.Denom, amountInAfterAddTakerFee.Ceil().TruncateInt())
	takerFeeCoin := sdk.NewCoin(tokenIn.Denom, tokenInAfterAddTakerFee.Amount.Sub(tokenIn.Amount))
	return tokenInAfterAddTakerFee, takerFeeCoin
}

// BurnPoolShareFromAccount burns `amount` of the given pools shares held by `addr`.
func (k Keeper) burnTakerFee(ctx sdk.Context, takerFeeCoin sdk.Coin, sender sdk.AccAddress) error {
	amt := sdk.NewCoins(takerFeeCoin)
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amt)
	if err != nil {
		return err
	}

	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, amt)
	if err != nil {
		return err
	}

	return nil
}

// find route from denom to defaultTakerFeeDenom by iterating all pools
func (k Keeper) findPoolWithTakerFeeDenom(ctx sdk.Context, fromDenom string) (poolmanagertypes.SwapAmountInRoute, error) {
	route := poolmanagertypes.SwapAmountInRoute{}

	iter := k.iterator(ctx, types.KeyPrefixPools)
	defer iter.Close() //nolint:errcheck

	for ; iter.Valid(); iter.Next() {
		pool, err := k.UnmarshalPool(iter.Value())
		if err != nil {
			return route, err
		}

		poolDenoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))

		//check if poolDenoms contains both fromDenom and defaultTakerFeeDenom
		if contains(poolDenoms, fromDenom) && contains(poolDenoms, defaultTakerFeeDenom) {
			route.PoolId = pool.GetId()
			route.TokenOutDenom = defaultTakerFeeDenom
			return route, nil
		}
	}

	return route, fmt.Errorf("failed to find pool with %s and %s", fromDenom, defaultTakerFeeDenom)
}
