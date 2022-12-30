package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// IterateRoutes checks the profitability of every single route that is passed in
// and returns the optimal route if there is one
func (k Keeper) IterateRoutes(ctx sdk.Context, routes []swaproutertypes.SwapAmountInRoutes) (sdk.Coin, sdk.Int, swaproutertypes.SwapAmountInRoutes) {
	var optimalRoute swaproutertypes.SwapAmountInRoutes
	var maxProfitInputCoin sdk.Coin
	maxProfit := sdk.ZeroInt()

	for _, route := range routes {
		// Find the max profit for the route using the token out denom of the last pool in the route as the input token denom
		inputCoin, profit, err := k.FindMaxProfitForRoute(ctx, route, route[2].TokenOutDenom)
		if err != nil {
			k.Logger(ctx).Error("Error finding max profit for route: ", err)
			continue
		}

		// Filter out routes that don't have any profit
		if profit.LTE(sdk.ZeroInt()) {
			continue
		}

		// If arb doesn't start and end with uosmo, then we convert the profit to uosmo, and compare profits in terms of uosmo
		if inputCoin.Denom != types.OsmosisDenomination {
			uosmoProfit, err := k.ConvertProfits(ctx, inputCoin, profit)
			if err != nil {
				k.Logger(ctx).Error("Error converting profits: ", err)
				continue
			}
			profit = uosmoProfit
		}

		// Select the optimal route King of the Hill style (route with the highest profit will be executed)
		if profit.GT(maxProfit) {
			optimalRoute = route
			maxProfit = profit
			maxProfitInputCoin = inputCoin
		}
	}

	return maxProfitInputCoin, maxProfit, optimalRoute
}

// ConvertProfits converts the profit denom to uosmo to allow for a fair comparison of profits
func (k Keeper) ConvertProfits(ctx sdk.Context, inputCoin sdk.Coin, profit sdk.Int) (sdk.Int, error) {
	// Get highest liquidity pool ID for the input coin and uosmo
	conversionPoolID, err := k.GetOsmoPool(ctx, inputCoin.Denom)
	if err != nil {
		return profit, err
	}

	// Get the pool
	conversionPool, err := k.gammKeeper.GetPoolAndPoke(ctx, conversionPoolID)
	if err != nil {
		return profit, err
	}

	// Calculate the amount of uosmo that we can get if we swapped the
	// profited amount of the orignal asset through the highest uosmo liquidity pool
	conversionTokenOut, err := conversionPool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(sdk.NewCoin(inputCoin.Denom, profit)),
		types.OsmosisDenomination, conversionPool.GetSwapFee(ctx))
	if err != nil {
		return profit, err
	}

	// return the profit denominated in uosmo
	return conversionTokenOut.Amount, nil
}

// EstimateMultihopProfit estimates the profit for a given route
// by estimating the amount out given the amount in for the first pool in the route
// and then subtracting the amount in from the amount out to get the profit
func (k Keeper) EstimateMultihopProfit(ctx sdk.Context, inputDenom string, amount sdk.Int, route swaproutertypes.SwapAmountInRoutes) (sdk.Coin, sdk.Int, error) {
	tokenIn := sdk.NewCoin(inputDenom, amount)
	amtOut, err := k.swaprouterKeeper.MultihopEstimateOutGivenExactAmountIn(ctx, route, tokenIn)
	if err != nil {
		return sdk.Coin{}, sdk.ZeroInt(), err
	}
	profit := amtOut.Sub(tokenIn.Amount)
	return tokenIn, profit, nil
}

// FindMaxProfitRoute runs a binary search to find the max profit for a given route
func (k Keeper) FindMaxProfitForRoute(ctx sdk.Context, route swaproutertypes.SwapAmountInRoutes, inputDenom string) (sdk.Coin, sdk.Int, error) {
	tokenIn := sdk.Coin{}
	profit := sdk.ZeroInt()

	curLeft := sdk.OneInt()
	curRight := types.MaxInputAmount
	iteration := 0

	for curLeft.LT(curRight) && iteration < types.MaxIterations {
		iteration++

		curMid := (curLeft.Add(curRight)).Quo(sdk.NewInt(2))
		curMidPlusOne := curMid.Add(sdk.OneInt())

		// Short circuit profit searching if there is an error in the GAMM module
		_, profitMid, err := k.EstimateMultihopProfit(ctx, inputDenom, curMid.Mul(types.StepSize), route)
		if err != nil {
			return sdk.Coin{}, sdk.ZeroInt(), err
		}

		// Short circuit profit searching if there is an error in the GAMM module
		tokenInMidPlusOne, profitMidPlusOne, err := k.EstimateMultihopProfit(ctx, inputDenom, curMidPlusOne.Mul(types.StepSize), route)
		if err != nil {
			return sdk.Coin{}, sdk.ZeroInt(), err
		}

		// Reduce subspace to search for max profit
		if profitMid.LTE(profitMidPlusOne) {
			curLeft = curMidPlusOne
			tokenIn = tokenInMidPlusOne
			profit = profitMidPlusOne
		} else {
			curRight = curMid
		}
	}

	return tokenIn, profit, nil
}

// ExecuteTrade inputs a route, amount in, and rebalances the pool
func (k Keeper) ExecuteTrade(ctx sdk.Context, route swaproutertypes.SwapAmountInRoutes, inputCoin sdk.Coin, poolId uint64) error {
	// Get the module address which will execute the trade
	protorevModuleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)

	// Mint the module account the input coin to trade
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(inputCoin)); err != nil {
		return err
	}

	// Use the inputCoin.Amount as the min amount out to ensure profitability
	tokenOutAmount, err := k.swaprouterKeeper.RouteExactAmountIn(ctx, protorevModuleAddress, route, inputCoin, inputCoin.Amount)
	if err != nil {
		return err
	}

	// Burn the coins from the module account after the trade and leave all remaining coins in the module account
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(inputCoin)); err != nil {
		return err
	}

	// Update the module statistics stores
	if err = k.UpdateStatistics(ctx, route, inputCoin, tokenOutAmount); err != nil {
		return err
	}

	return nil
}
