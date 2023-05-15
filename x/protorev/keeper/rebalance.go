package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// IterateRoutes checks the profitability of every single route that is passed in
// and returns the optimal route if there is one
func (k Keeper) IterateRoutes(ctx sdk.Context, routes []RouteMetaData, remainingTxPoolPoints, remainingBlockPoolPoints *uint64) (sdk.Coin, sdk.Int, poolmanagertypes.SwapAmountInRoutes) {
	var optimalRoute poolmanagertypes.SwapAmountInRoutes
	var maxProfitInputCoin sdk.Coin
	maxProfit := sdk.ZeroInt()

	// Iterate through the routes and find the optimal route for the given swap
	for index := 0; index < len(routes) && *remainingTxPoolPoints > 0; index++ {
		// If the route consumes more pool points than we have remaining then we skip it
		if routes[index].PoolPoints > *remainingTxPoolPoints {
			continue
		}

		// Find the max profit for the route if it exists
		inputCoin, profit, err := k.FindMaxProfitForRoute(ctx, routes[index], remainingTxPoolPoints, remainingBlockPoolPoints)
		if err != nil {
			k.Logger(ctx).Error("Error finding max profit for route: ", err)
			continue
		}

		// If the profit is greater than zero, then we convert the profits to uosmo and compare profits in terms of uosmo
		if profit.GT(sdk.ZeroInt()) {
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
				optimalRoute = routes[index].Route
				maxProfit = profit
				maxProfitInputCoin = inputCoin
			}
		}
	}

	return maxProfitInputCoin, maxProfit, optimalRoute
}

// ConvertProfits converts the profit denom to uosmo to allow for a fair comparison of profits
func (k Keeper) ConvertProfits(ctx sdk.Context, inputCoin sdk.Coin, profit sdk.Int) (sdk.Int, error) {
	// Get highest liquidity pool ID for the input coin and uosmo
	conversionPoolID, err := k.GetPoolForDenomPair(ctx, types.OsmosisDenomination, inputCoin.Denom)
	if err != nil {
		return profit, err
	}

	// Get the pool
	conversionPool, err := k.poolmanagerKeeper.GetPool(ctx, conversionPoolID)
	if err != nil {
		return profit, err
	}

	swapModule, err := k.poolmanagerKeeper.GetPoolModule(ctx, conversionPoolID)
	if err != nil {
		return profit, err
	}

	// Calculate the amount of uosmo that we can get if we swapped the
	// profited amount of the orignal asset through the highest uosmo liquidity pool
	conversionTokenOut, err := swapModule.CalcOutAmtGivenIn(
		ctx,
		conversionPool,
		sdk.NewCoin(inputCoin.Denom, profit),
		types.OsmosisDenomination,
		conversionPool.GetSwapFee(ctx),
	)
	if err != nil {
		return profit, err
	}

	// return the profit denominated in uosmo
	return conversionTokenOut.Amount, nil
}

// EstimateMultihopProfit estimates the profit for a given route
// by estimating the amount out given the amount in for the first pool in the route
// and then subtracting the amount in from the amount out to get the profit
func (k Keeper) EstimateMultihopProfit(ctx sdk.Context, inputDenom string, amount sdk.Int, route poolmanagertypes.SwapAmountInRoutes) (sdk.Coin, sdk.Int, error) {
	tokenIn := sdk.NewCoin(inputDenom, amount)
	amtOut, err := k.poolmanagerKeeper.MultihopEstimateOutGivenExactAmountIn(ctx, route, tokenIn)
	if err != nil {
		return sdk.Coin{}, sdk.ZeroInt(), err
	}
	profit := amtOut.Sub(tokenIn.Amount)
	return tokenIn, profit, nil
}

// FindMaxProfitRoute runs a binary search to find the max profit for a given route
func (k Keeper) FindMaxProfitForRoute(ctx sdk.Context, route RouteMetaData, remainingTxPoolPoints, remainingBlockPoolPoints *uint64) (sdk.Coin, sdk.Int, error) {
	// Track the tokenIn amount/denom and the profit
	tokenIn := sdk.Coin{}
	profit := sdk.ZeroInt()

	// Track the left and right bounds of the binary search
	curLeft := sdk.OneInt()
	curRight := types.MaxInputAmount

	// Input denom used for cyclic arbitrage
	inputDenom := route.Route[route.Route.Length()-1].TokenOutDenom

	// If a cyclic arb exists with an optimal amount in above our minimum amount in,
	// then inputting the minimum amount in will result in a profit. So we check for that first.
	// If there is no profit, then we can return early and not run the binary search.
	_, minInProfit, err := k.EstimateMultihopProfit(ctx, inputDenom, curLeft.Mul(route.StepSize), route.Route)
	if err != nil {
		return sdk.Coin{}, sdk.ZeroInt(), err
	} else if minInProfit.LTE(sdk.ZeroInt()) {
		return sdk.Coin{}, sdk.ZeroInt(), nil
	}

	// Decrement the number of pool points remaining since we know this route will be profitable
	*remainingTxPoolPoints -= route.PoolPoints
	*remainingBlockPoolPoints -= route.PoolPoints

	// Increment the number of pool points consumed since we know this route will be profitable
	if err := k.IncrementPointCountForBlock(ctx, route.PoolPoints); err != nil {
		return sdk.Coin{}, sdk.ZeroInt(), err
	}

	// Extend the search range if the max input amount is too small
	curLeft, curRight = k.ExtendSearchRangeIfNeeded(ctx, route, inputDenom, curLeft, curRight)

	// Binary search to find the max profit
	for iteration := 0; curLeft.LT(curRight) && iteration < types.MaxIterations; iteration++ {
		curMid := (curLeft.Add(curRight)).Quo(sdk.NewInt(2))
		curMidPlusOne := curMid.Add(sdk.OneInt())

		// Short circuit profit searching if there is an error in the GAMM module
		tokenInMid, profitMid, err := k.EstimateMultihopProfit(ctx, inputDenom, curMid.Mul(route.StepSize), route.Route)
		if err != nil {
			return sdk.Coin{}, sdk.ZeroInt(), err
		}

		// Short circuit profit searching if there is an error in the GAMM module
		tokenInMidPlusOne, profitMidPlusOne, err := k.EstimateMultihopProfit(ctx, inputDenom, curMidPlusOne.Mul(route.StepSize), route.Route)
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
			tokenIn = tokenInMid
			profit = profitMid
		}
	}

	return tokenIn, profit, nil
}

// Determine if the binary search range needs to be extended
func (k Keeper) ExtendSearchRangeIfNeeded(ctx sdk.Context, route RouteMetaData, inputDenom string, curLeft, curRight sdk.Int) (sdk.Int, sdk.Int) {
	// Get the profit for the maximum amount in
	_, maxInProfit, err := k.EstimateMultihopProfit(ctx, inputDenom, curRight.Mul(route.StepSize), route.Route)
	if err != nil {
		return curLeft, curRight
	}

	// If the profit for the maximum amount in is still increasing, then we can increase the range of the binary search
	if maxInProfit.GTE(sdk.ZeroInt()) {
		// Get the profit for the maximum amount in + 1
		_, maxInProfitPlusOne, err := k.EstimateMultihopProfit(ctx, inputDenom, curRight.Add(sdk.OneInt()).Mul(route.StepSize), route.Route)
		if err != nil {
			return curLeft, curRight
		}

		// Change the range of the binary search if the profit is still increasing
		if maxInProfitPlusOne.GT(maxInProfit) {
			curLeft = curRight
			curRight = types.ExtendedMaxInputAmount
		}
	}

	return curLeft, curRight
}

// ExecuteTrade inputs a route, amount in, and rebalances the pool
func (k Keeper) ExecuteTrade(ctx sdk.Context, route poolmanagertypes.SwapAmountInRoutes, inputCoin sdk.Coin, pool SwapToBackrun, remainingTxPoolPoints, remainingBlockPoolPoints uint64) error {
	// Get the module address which will execute the trade
	protorevModuleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)

	// Mint the module account the input coin to trade
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(inputCoin)); err != nil {
		return err
	}

	// Use the inputCoin.Amount as the min amount out to ensure profitability
	tokenOutAmount, err := k.poolmanagerKeeper.RouteExactAmountIn(ctx, protorevModuleAddress, route, inputCoin, inputCoin.Amount)
	if err != nil {
		return err
	}

	// Burn the coins from the module account after the trade and leave all remaining coins in the module account
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(inputCoin)); err != nil {
		return err
	}

	// Profit from the trade
	profit := tokenOutAmount.Sub(inputCoin.Amount)

	// Update the module statistics stores
	if err = k.UpdateStatistics(ctx, route, inputCoin.Denom, profit); err != nil {
		return err
	}

	// Update the developer fees
	if err = k.UpdateDeveloperFees(ctx, inputCoin.Denom, profit); err != nil {
		return err
	}

	// Create and emit the backrun event and add it to the context
	EmitBackrunEvent(ctx, pool, inputCoin, profit, tokenOutAmount, remainingTxPoolPoints, remainingBlockPoolPoints)

	return nil
}

// RemainingPoolPointsForTx calculates the number of pool points that can be consumed in the transaction and block.
// When the remaining pool points for the block is less than the remaining pool points for the transaction, then both
// returned values will be the same, which will be the remaining pool points for the block.
func (k Keeper) GetRemainingPoolPoints(ctx sdk.Context) (uint64, uint64, error) {
	maxPoolPointsPerTx, err := k.GetMaxPointsPerTx(ctx)
	if err != nil {
		return 0, 0, err
	}

	maxPoolPointsPerBlock, err := k.GetMaxPointsPerBlock(ctx)
	if err != nil {
		return 0, 0, err
	}

	currentPoolPointsUsedForBlock, err := k.GetPointCountForBlock(ctx)
	if err != nil {
		return 0, 0, err
	}

	// Edge case where the number of pool points consumed in the current block is greater than the max number of routes per block
	// This should never happen, but we need to handle it just in case (deal with overflow)
	if currentPoolPointsUsedForBlock >= maxPoolPointsPerBlock {
		return 0, 0, nil
	}

	// Calculate the number of pool points that can be iterated over
	numberOfAvailablePoolPointsForBlock := maxPoolPointsPerBlock - currentPoolPointsUsedForBlock
	if numberOfAvailablePoolPointsForBlock > maxPoolPointsPerTx {
		return maxPoolPointsPerTx, numberOfAvailablePoolPointsForBlock, nil
	}

	return numberOfAvailablePoolPointsForBlock, numberOfAvailablePoolPointsForBlock, nil
}
