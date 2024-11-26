package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

var zeroInt = osmomath.ZeroInt()

// IterateRoutes checks the profitability of every single route that is passed in
// and returns the optimal route if there is one
func (k Keeper) IterateRoutes(ctx sdk.Context, routes []RouteMetaData, remainingTxPoolPoints, remainingBlockPoolPoints *uint64) (sdk.Coin, osmomath.Int, poolmanagertypes.SwapAmountInRoutes) {
	var optimalRoute poolmanagertypes.SwapAmountInRoutes
	var maxProfitInputCoin sdk.Coin
	maxProfit := osmomath.ZeroInt()

	// Iterate through the routes and find the optimal route for the given swap
	for index := 0; index < len(routes) && *remainingTxPoolPoints > 0; index++ {
		// If the route consumes more pool points than we have remaining then we skip it
		if routes[index].PoolPoints > *remainingTxPoolPoints {
			continue
		}

		// Find the max profit for the route if it exists
		inputCoin, profit, err := k.FindMaxProfitForRoute(ctx, routes[index], remainingTxPoolPoints, remainingBlockPoolPoints)
		if err != nil {
			k.Logger(ctx).Debug("Error finding max profit for route: " + err.Error())
			continue
		}

		// If the profit is greater than zero, then we convert the profits to uosmo and compare profits in terms of uosmo
		if profit.GT(zeroInt) {
			profit, err := k.ConvertProfits(ctx, inputCoin, profit)
			if err != nil {
				k.Logger(ctx).Error("Error converting profits: " + err.Error())
				continue
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
//
// NOTE: This does not check the underlying pool before swapping so this may go over the MaxTicksCrossed.
func (k Keeper) ConvertProfits(ctx sdk.Context, inputCoin sdk.Coin, profit osmomath.Int) (osmomath.Int, error) {
	if inputCoin.Denom == types.OsmosisDenomination {
		return profit, nil
	}

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
	// profited amount of the original asset through the highest uosmo liquidity pool
	conversionTokenOut, err := swapModule.CalcOutAmtGivenIn(
		ctx,
		conversionPool,
		sdk.NewCoin(inputCoin.Denom, profit),
		types.OsmosisDenomination,
		conversionPool.GetSpreadFactor(ctx),
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
func (k Keeper) EstimateMultihopProfit(ctx sdk.Context, inputDenom string, amount osmomath.Int, route poolmanagertypes.SwapAmountInRoutes) (sdk.Coin, osmomath.Int, error) {
	tokenIn := sdk.Coin{Denom: inputDenom, Amount: amount}
	amtOut, err := k.poolmanagerKeeper.MultihopEstimateOutGivenExactAmountInNoTakerFee(ctx, route, tokenIn)
	if err != nil {
		return sdk.Coin{}, osmomath.ZeroInt(), err
	}
	profit := amtOut.Sub(tokenIn.Amount)
	return tokenIn, profit, nil
}

var oneInt, twoInt = osmomath.OneInt(), osmomath.NewInt(2)

// FindMaxProfitRoute runs a binary search to find the max profit for a given route
func (k Keeper) FindMaxProfitForRoute(ctx sdk.Context, route RouteMetaData, remainingTxPoolPoints, remainingBlockPoolPoints *uint64) (sdk.Coin, osmomath.Int, error) {
	// Track the tokenIn amount/denom and the profit
	tokenIn := sdk.Coin{}
	profit := osmomath.ZeroInt()

	// Track the left and right bounds of the binary search
	curLeft := osmomath.OneInt()
	curRight := types.MaxInputAmount

	// Input denom used for cyclic arbitrage
	inputDenom := route.Route[route.Route.Length()-1].TokenOutDenom

	// If a cyclic arb exists with an optimal amount in above our minimum amount in,
	// then inputting the minimum amount in will result in a profit. So we check for that first.
	// If there is no profit, then we can return early and not run the binary search.
	_, minInProfit, err := k.EstimateMultihopProfit(ctx, inputDenom, curLeft.Mul(route.StepSize), route.Route)
	if err != nil {
		return sdk.Coin{}, osmomath.ZeroInt(), err
	} else if minInProfit.LTE(osmomath.ZeroInt()) {
		return sdk.Coin{}, osmomath.ZeroInt(), nil
	}

	// Decrement the number of pool points remaining since we know this route will be profitable
	*remainingTxPoolPoints -= route.PoolPoints
	*remainingBlockPoolPoints -= route.PoolPoints

	// Increment the number of pool points consumed since we know this route will be profitable
	if err := k.IncrementPointCountForBlock(ctx, route.PoolPoints); err != nil {
		return sdk.Coin{}, osmomath.ZeroInt(), err
	}

	// Update the search range if the max input amount is too small/large
	curLeft, curRight, err = k.UpdateSearchRangeIfNeeded(ctx, route, inputDenom, curLeft, curRight)
	if err != nil {
		return sdk.Coin{}, osmomath.ZeroInt(), err
	}

	// Binary search to find the max profit
	for iteration := 0; curLeft.LT(curRight) && iteration < types.MaxIterations; iteration++ {
		curMid := (curLeft.Add(curRight)).Quo(twoInt)
		curMidPlusOne := curMid.Add(oneInt)

		// Short circuit profit searching if there is an error in the GAMM module
		tokenInMid, profitMid, err := k.EstimateMultihopProfit(ctx, inputDenom, curMid.Mul(route.StepSize), route.Route)
		if err != nil {
			return sdk.Coin{}, osmomath.ZeroInt(), err
		}

		// Short circuit profit searching if there is an error in the GAMM module
		tokenInMidPlusOne, profitMidPlusOne, err := k.EstimateMultihopProfit(ctx, inputDenom, curMidPlusOne.Mul(route.StepSize), route.Route)
		if err != nil {
			return sdk.Coin{}, osmomath.ZeroInt(), err
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

// UpdateSearchRangeIfNeeded updates the search range for the binary search. First, we check if there are any
// concentrated liquidity pools in the route. If there are, then we may need to reduce the upper bound of the
// binary search since it is gas intensive to move across several ticks. Next, we determine if the current bound
// includes the optimal amount in. If it does not, then we can extend the search range to capture more profits.
func (k Keeper) UpdateSearchRangeIfNeeded(
	ctx sdk.Context,
	route RouteMetaData,
	inputDenom string,
	curLeft, curRight osmomath.Int,
) (osmomath.Int, osmomath.Int, error) {
	// If there are concentrated liquidity pools in the route, then we may need to reduce the upper bound of the binary search.
	updatedMax, err := k.CalculateUpperBoundForSearch(ctx, route, inputDenom)
	if err != nil {
		return osmomath.ZeroInt(), osmomath.ZeroInt(), err
	}

	// In the case where the updated upper bound is less than the current upper bound, we know we will not extend
	// the search range so we can short-circuit return.
	if updatedMax.LT(curRight) {
		return curLeft, updatedMax, nil
	}

	return k.ExtendSearchRangeIfNeeded(ctx, route, inputDenom, curLeft, curRight, updatedMax)
}

// CalculateUpperBoundForSearch returns the max amount in that can be used for the binary search
// respecting the max ticks moved across all concentrated liquidity pools in the route.
func (k Keeper) CalculateUpperBoundForSearch(
	ctx sdk.Context,
	route RouteMetaData,
	inputDenom string,
) (osmomath.Int, error) {
	var intermidiateCoin sdk.Coin

	poolInfo := k.GetInfoByPoolType(ctx)

	// Iterate through all CL pools and determine the maximal amount of input that can be used
	// respecting the max ticks moved.
	for index := route.Route.Length() - 1; index >= 0; index-- {
		hop := route.Route[index]
		pool, err := k.poolmanagerKeeper.GetPool(ctx, hop.PoolId)
		if err != nil {
			return osmomath.ZeroInt(), err
		}

		tokenInDenom := inputDenom
		if index > 0 {
			tokenInDenom = route.Route[index-1].TokenOutDenom
		}

		switch {
		case pool.GetType() == poolmanagertypes.Concentrated:
			// If the pool is a concentrated liquidity pool, then check the maximum amount in that can be used
			// and determine what this amount is as an input at the previous pool (working all the way up to the
			// beginning of the route).
			maxTokenIn, maxTokenOut, err := k.concentratedLiquidityKeeper.ComputeMaxInAmtGivenMaxTicksCrossed(
				ctx,
				pool.GetId(),
				tokenInDenom,
				poolInfo.Concentrated.MaxTicksCrossed,
			)
			if err != nil {
				return osmomath.ZeroInt(), err
			}

			// if there have been no other CL pools in the route, then we can set the intermediate coin to the max input amount.
			// Additionally, if the amount of the previous token is greater than the possible amount from this pool, then we
			// can set the intermediate coin to the max input amount (from the current pool). Otherwise we have to do a
			// safe swap given the previous max amount.
			if intermidiateCoin.IsNil() || maxTokenOut.Amount.LT(intermidiateCoin.Amount) {
				intermidiateCoin = maxTokenIn
				continue
			}

			// In the scenario where there are multiple CL pools in a route, we select the one that inputs
			// the smaller amount to ensure we do not overstep the max ticks moved.
			intermidiateCoin, err = k.executeSafeSwap(ctx, pool.GetId(), intermidiateCoin, tokenInDenom)
			if err != nil {
				return osmomath.ZeroInt(), err
			}
		case !intermidiateCoin.IsNil():
			// If we have already seen a CL pool in the route, then simply propagate the intermediate coin up
			// the route.
			intermidiateCoin, err = k.executeSafeSwap(ctx, pool.GetId(), intermidiateCoin, tokenInDenom)
			if err != nil {
				return osmomath.ZeroInt(), err
			}
		}
	}

	// In the case where there are no CL pools, we want to return the extended max input amount
	if intermidiateCoin.IsNil() {
		return types.ExtendedMaxInputAmount, nil
	}

	// Ensure that the normalized upper bound is not greater than the extended max input amount
	upperBound := intermidiateCoin.Amount.Quo(route.StepSize)
	if upperBound.GT(types.ExtendedMaxInputAmount) {
		return types.ExtendedMaxInputAmount, nil
	}

	return upperBound, nil
}

// executeSafeSwap executes a safe swap by first ensuring the swap amount is less than the
// amount of total liquidity in the pool.
func (k Keeper) executeSafeSwap(
	ctx sdk.Context,
	poolID uint64,
	outputCoin sdk.Coin,
	tokenInDenom string,
) (sdk.Coin, error) {
	liquidity, err := k.poolmanagerKeeper.GetTotalPoolLiquidity(ctx, poolID)
	if err != nil {
		return sdk.NewCoin(tokenInDenom, osmomath.ZeroInt()), err
	}

	// At most we can swap half of the liquidity in the pool
	liquidTokenAmt := liquidity.AmountOf(outputCoin.Denom).Quo(osmomath.NewInt(4))
	if liquidTokenAmt.LT(outputCoin.Amount) {
		outputCoin.Amount = liquidTokenAmt
	}

	amt, err := k.poolmanagerKeeper.MultihopEstimateInGivenExactAmountOut(
		ctx,
		poolmanagertypes.SwapAmountOutRoutes{
			{
				PoolId:       poolID,
				TokenInDenom: tokenInDenom,
			},
		},
		outputCoin,
	)
	if err != nil {
		return sdk.NewCoin(tokenInDenom, osmomath.ZeroInt()), err
	}

	return sdk.NewCoin(tokenInDenom, amt), nil
}

// Determine if the binary search range needs to be extended
func (k Keeper) ExtendSearchRangeIfNeeded(
	ctx sdk.Context,
	route RouteMetaData,
	inputDenom string,
	curLeft, curRight, updatedMax osmomath.Int,
) (osmomath.Int, osmomath.Int, error) {
	// Get the profit for the maximum amount in
	_, maxInProfit, err := k.EstimateMultihopProfit(ctx, inputDenom, curRight.Mul(route.StepSize), route.Route)
	if err != nil {
		return osmomath.ZeroInt(), osmomath.ZeroInt(), err
	}

	// If the profit for the maximum amount in is still increasing, then we can increase the range of the binary search
	if maxInProfit.GTE(osmomath.ZeroInt()) {
		// Get the profit for the maximum amount in + 1
		_, maxInProfitPlusOne, err := k.EstimateMultihopProfit(ctx, inputDenom, curRight.Add(osmomath.OneInt()).Mul(route.StepSize), route.Route)
		if err != nil {
			return osmomath.ZeroInt(), osmomath.ZeroInt(), err
		}

		// Change the range of the binary search if the profit is still increasing
		if maxInProfitPlusOne.GT(maxInProfit) {
			curLeft = curRight
			curRight = updatedMax
		}
	}

	return curLeft, curRight, nil
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
