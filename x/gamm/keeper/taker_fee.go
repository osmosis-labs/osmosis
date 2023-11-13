package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

const (
	defaultTakerFeeDenom = "udym"
)

// This function is a helper function to support the SwapAmountOut msg
// it finds the route to the defaultTakerFeeDenom to be used to swap the taker fee
func RouteToBaseDenomFromOutRoutes(routes poolmanagertypes.SwapAmountOutRoutes, denomOut string) []poolmanagertypes.SwapAmountInRoute {
	var newRoutes []poolmanagertypes.SwapAmountInRoute

	//if the denomOut is the defaultTakerFeeDenom, we need to swap all the routes
	if denomOut == defaultTakerFeeDenom {
		for i, route := range routes[:len(routes)-1] {
			newRoutes = append(newRoutes, poolmanagertypes.SwapAmountInRoute{
				PoolId:        route.PoolId,
				TokenOutDenom: routes[i+1].TokenInDenom,
			})
		}
		newRoutes = append(newRoutes, poolmanagertypes.SwapAmountInRoute{
			PoolId:        routes[len(routes)-1].PoolId,
			TokenOutDenom: denomOut,
		})
		return newRoutes
	}

	var found bool
	var idx int
	// check where it swapped to defaultTakerFeeDenom on the routes
	for i, route := range routes {
		if route.TokenInDenom == defaultTakerFeeDenom {
			found = true
			idx = i
			break
		}
	}

	//if not found, return empty as we can't swap to the defaultTakerFeeDenom
	if idx == 0 || !found {
		return newRoutes
	}

	for i, route := range routes[:idx] {
		newRoutes = append(newRoutes, poolmanagertypes.SwapAmountInRoute{
			PoolId:        route.PoolId,
			TokenOutDenom: routes[i+1].TokenInDenom,
		})
	}

	return newRoutes
}

func (k Keeper) chargeTakerFeeSwapAmountOut(ctx sdk.Context, takerFeeCoin sdk.Coin, sender sdk.AccAddress, outRoutes []poolmanagertypes.SwapAmountOutRoute, denomOut string) error {
	if takerFeeCoin.Denom == defaultTakerFeeDenom {
		return k.burnTakerFee(ctx, takerFeeCoin, sender)
	}

	routes := RouteToBaseDenomFromOutRoutes(outRoutes, denomOut)
	if len(routes) == 0 {
		ctx.Logger().Error("failed to swap taker fee to base denom")
		return k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(takerFeeCoin), sender)
	}

	err := k.swapAndBurn(ctx, sender, routes, takerFeeCoin)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) chargeTakerFeeSwapAmountIn(ctx sdk.Context, takerFeeCoin sdk.Coin, sender sdk.AccAddress, routes []poolmanagertypes.SwapAmountInRoute) error {
	if takerFeeCoin.Denom == defaultTakerFeeDenom {
		return k.burnTakerFee(ctx, takerFeeCoin, sender)
	}

	//build new subroute to swap takerFeeCoin to base denom
	var newRoutes []poolmanagertypes.SwapAmountInRoute
	for i, route := range routes {
		if route.TokenOutDenom == defaultTakerFeeDenom {
			newRoutes = routes[:i]
			break
		}
	}
	if len(newRoutes) == 0 {
		ctx.Logger().Error("failed to swap taker fee to base denom")
		return k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(takerFeeCoin), sender)
	}

	err := k.swapAndBurn(ctx, sender, newRoutes, takerFeeCoin)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) swapAndBurn(ctx sdk.Context, sender sdk.AccAddress, routes []poolmanagertypes.SwapAmountInRoute, tokenIn sdk.Coin) error {
	minAmountOut := sdk.ZeroInt()
	// Do the swap of this fee token denom to base denom.
	out, err := k.poolManager.RouteExactAmountIn(ctx, sender, routes, tokenIn, minAmountOut)
	if err != nil {
		return err
	}

	return k.burnTakerFee(ctx, sdk.NewCoin(defaultTakerFeeDenom, out), sender)
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
