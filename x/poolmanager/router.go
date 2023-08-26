package poolmanager

import (
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	appparams "github.com/osmosis-labs/osmosis/v19/app/params"
	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

// 1 << 256 - 1 where 256 is the max bit length defined for sdk.Int
var intMaxValue = sdk.NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)))

// RouteExactAmountIn processes a swap along the given route using the swap function
// corresponding to poolID's pool type. It takes in the input denom and amount for
// the initial swap against the first pool and chains the output as the input for the
// next routed pool until the last pool is reached.
// Transaction succeeds if final amount out is greater than tokenOutMinAmount defined
// and no errors are encountered along the way.
func (k Keeper) RouteExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	route []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	var (
		isMultiHopRouted   bool
		routeSpreadFactor  sdk.Dec
		sumOfSpreadFactors sdk.Dec
	)

	// Ensure that provided route is not empty and has valid denom format.
	routeStep := types.SwapAmountInRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return sdk.Int{}, err
	}

	// In this loop (isOsmoRoutedMultihop), we check if:
	// - the routeStep is of length 2
	// - routeStep 1 and routeStep 2 don't trade via the same pool
	// - routeStep 1 contains uosmo
	// - both routeStep 1 and routeStep 2 are incentivized pools
	//
	// If all of the above is true, then we collect the additive and max fee between the
	// two pools to later calculate the following:
	// total_spread_factor = max(spread_factor1, spread_factor2)
	// fee_per_pool = total_spread_factor * ((pool_fee) / (spread_factor1 + spread_factor2))
	if k.isOsmoRoutedMultihop(ctx, routeStep, route[0].TokenOutDenom, tokenIn.Denom) {
		isMultiHopRouted = true
		routeSpreadFactor, sumOfSpreadFactors, err = k.getOsmoRoutedMultihopTotalSpreadFactor(ctx, routeStep)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	// Iterate through the route and execute a series of swaps through each pool.
	for i, routeStep := range route {
		// To prevent the multihop swap from being interrupted prematurely, we keep
		// the minimum expected output at a very low number until the last pool
		_outMinAmount := sdk.NewInt(1)
		if len(route)-1 == i {
			_outMinAmount = tokenOutMinAmount
		}

		// Get underlying pool type corresponding to the pool ID at the current routeStep.
		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := swapModule.GetPool(ctx, routeStep.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		// Check if pool has swaps enabled.
		if !pool.IsActive(ctx) {
			return sdk.Int{}, types.InactivePoolError{PoolId: pool.GetId()}
		}

		spreadFactor := pool.GetSpreadFactor(ctx)

		// If we determined the route is an osmo multi-hop and both routes are incentivized,
		// we modify the spread factor accordingly.
		if isMultiHopRouted {
			spreadFactor = routeSpreadFactor.MulRoundUp((spreadFactor.QuoRoundUp(sumOfSpreadFactors)))
		}

		tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, routeStep.TokenOutDenom, _outMinAmount, spreadFactor)
		if err != nil {
			return sdk.Int{}, err
		}

		// Chain output of current pool as the input for the next routed pool
		tokenIn = sdk.NewCoin(routeStep.TokenOutDenom, tokenOutAmount)
	}
	return tokenOutAmount, nil
}

// SplitRouteExactAmountIn routes the swap across multiple multihop paths
// to get the desired token out. This is useful for achieving the most optimal execution. However, note that the responsibility
// of determining the optimal split is left to the client. This method simply route the swap across the given route.
// The route must end with the same token out and begin with the same token in.
//
// It performs the price impact protection check on the combination of tokens out from all multihop paths. The given tokenOutMinAmount
// is used for comparison.
//
// Returns error if:
//   - route are empty
//   - route contain duplicate multihop paths
//   - last token out denom is not the same for all multihop paths in routeStep
//   - one of the multihop swaps fails for internal reasons
//   - final token out computed is not positive
//   - final token out computed is smaller than tokenOutMinAmount
func (k Keeper) SplitRouteExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountInSplitRoute,
	tokenInDenom string,
	tokenOutMinAmount sdk.Int,
) (sdk.Int, error) {
	if err := types.ValidateSwapAmountInSplitRoute(routes); err != nil {
		return sdk.Int{}, err
	}

	var (
		// We start the multihop min amount as zero because we want
		// to perform a price impact protection check on the combination of tokens out
		// from all multihop paths.
		multihopStartTokenOutMinAmount = sdk.ZeroInt()
		totalOutAmount                 = sdk.ZeroInt()
	)

	for _, multihopRoute := range routes {
		tokenOutAmount, err := k.RouteExactAmountIn(
			ctx,
			sender,
			types.SwapAmountInRoutes(multihopRoute.Pools),
			sdk.NewCoin(tokenInDenom, multihopRoute.TokenInAmount),
			multihopStartTokenOutMinAmount)
		if err != nil {
			return sdk.Int{}, err
		}

		totalOutAmount = totalOutAmount.Add(tokenOutAmount)
	}

	if !totalOutAmount.IsPositive() {
		return sdk.Int{}, types.FinalAmountIsNotPositiveError{IsAmountOut: true, Amount: totalOutAmount}
	}

	if totalOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, types.PriceImpactProtectionExactInError{Actual: totalOutAmount, MinAmount: tokenOutMinAmount}
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSplitRouteSwapExactAmountIn,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
			sdk.NewAttribute(types.AttributeKeyTokensOut, totalOutAmount.String()),
		),
	})

	return totalOutAmount, nil
}

// SwapExactAmountIn is an API for swapping an exact amount of tokens
// as input to a pool to get a minimum amount of the desired token out.
// The method succeeds when tokenOutAmount is greater than tokenOutMinAmount defined.
// Errors otherwise. Also, errors if the pool id is invalid, if tokens do not belong to the pool with given
// id or if sender does not have the swapped-in tokenIn.
func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	// Get the pool-specific module implementation to ensure that
	// swaps are routed to the pool type corresponding to pool ID's pool.
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	// Get pool as a general pool type. Note that the underlying function used
	// still varies with the pool type.
	pool, poolErr := swapModule.GetPool(ctx, poolId)
	if poolErr != nil {
		return sdk.Int{}, poolErr
	}

	// Check if pool has swaps enabled.
	if !pool.IsActive(ctx) {
		return sdk.Int{}, fmt.Errorf("pool %d is not active", pool.GetId())
	}

	spreadFactor := pool.GetSpreadFactor(ctx)

	// routeStep to the pool-specific SwapExactAmountIn implementation.
	tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, spreadFactor)
	if err != nil {
		return sdk.Int{}, err
	}

	return tokenOutAmount, nil
}

func (k Keeper) MultihopEstimateOutGivenExactAmountIn(
	ctx sdk.Context,
	route []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
) (tokenOutAmount sdk.Int, err error) {
	var (
		isMultiHopRouted   bool
		routeSpreadFactor  sdk.Dec
		sumOfSpreadFactors sdk.Dec
	)

	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			tokenOutAmount = sdk.Int{}
			err = fmt.Errorf("function MultihopEstimateOutGivenExactAmountIn failed due to internal reason: %v", r)
		}
	}()

	routeStep := types.SwapAmountInRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return sdk.Int{}, err
	}

	if k.isOsmoRoutedMultihop(ctx, routeStep, route[0].TokenOutDenom, tokenIn.Denom) {
		isMultiHopRouted = true
		routeSpreadFactor, sumOfSpreadFactors, err = k.getOsmoRoutedMultihopTotalSpreadFactor(ctx, routeStep)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	for _, routeStep := range route {
		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}

		// Execute the expected swap on the current routed pool
		poolI, poolErr := swapModule.GetPool(ctx, routeStep.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		spreadFactor := poolI.GetSpreadFactor(ctx)

		// If we determined the routeStep is an osmo multi-hop and both route are incentivized,
		// we modify the swap fee accordingly.
		if isMultiHopRouted {
			spreadFactor = routeSpreadFactor.Mul((spreadFactor.Quo(sumOfSpreadFactors)))
		}

		tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, tokenIn, routeStep.TokenOutDenom, spreadFactor)
		if err != nil {
			return sdk.Int{}, err
		}

		tokenOutAmount = tokenOut.Amount
		if !tokenOutAmount.IsPositive() {
			return sdk.Int{}, errors.New("token amount must be positive")
		}

		// Chain output of current pool as the input for the next routed pool
		tokenIn = sdk.NewCoin(routeStep.TokenOutDenom, tokenOutAmount)
	}
	return tokenOutAmount, err
}

// RouteExactAmountOut processes a swap along the given route using the swap function corresponding
// to poolID's pool type. This function is responsible for computing the optimal output amount
// for a given input amount when swapping tokens, taking into account the current price of the
// tokens in the pool and any slippage.
// Transaction succeeds if the calculated tokenInAmount of the first pool is less than the defined
// tokenInMaxAmount defined.
func (k Keeper) RouteExactAmountOut(ctx sdk.Context,
	sender sdk.AccAddress,
	route []types.SwapAmountOutRoute,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	isMultiHopRouted, routeSpreadFactor, sumOfSpreadFactors := false, sdk.Dec{}, sdk.Dec{}
	// Ensure that provided route is not empty and has valid denom format.
	routeStep := types.SwapAmountOutRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return sdk.Int{}, err
	}

	defer func() {
		if r := recover(); r != nil {
			tokenInAmount = sdk.Int{}
			err = fmt.Errorf("function RouteExactAmountOut failed due to internal reason: %v", r)
		}
	}()

	// In this loop (isOsmoRoutedMultihop), we check if:
	// - the routeStep is of length 2
	// - routeStep 1 and routeStep 2 don't trade via the same pool
	// - routeStep 1 contains uosmo
	// - both routeStep 1 and routeStep 2 are incentivized pools
	//
	// if all of the above is true, then we collect the additive and max fee between the two pools to later calculate the following:
	// total_spread_factor = total_spread_factor = max(spread_factor1, spread_factor2)
	// fee_per_pool = total_spread_factor * ((pool_fee) / (spread_factor1 + spread_factor2))
	var insExpected []sdk.Int
	isMultiHopRouted = k.isOsmoRoutedMultihop(ctx, routeStep, route[0].TokenInDenom, tokenOut.Denom)

	// Determine what the estimated input would be for each pool along the multi-hop routeStep
	// if we determined the routeStep is an osmo multi-hop and both route are incentivized,
	// we utilize a separate function that calculates the discounted swap fees
	if isMultiHopRouted {
		routeSpreadFactor, sumOfSpreadFactors, err = k.getOsmoRoutedMultihopTotalSpreadFactor(ctx, routeStep)
		if err != nil {
			return sdk.Int{}, err
		}
		insExpected, err = k.createOsmoMultihopExpectedSwapOuts(ctx, route, tokenOut, routeSpreadFactor, sumOfSpreadFactors)
	} else {
		insExpected, err = k.createMultihopExpectedSwapOuts(ctx, route, tokenOut)
	}

	if err != nil {
		return sdk.Int{}, err
	}
	if len(insExpected) == 0 {
		return sdk.Int{}, nil
	}
	insExpected[0] = tokenInMaxAmount

	// Iterates through each routed pool and executes their respective swaps. Note that all of the work to get the return
	// value of this method is done when we calculate insExpected – this for loop primarily serves to execute the actual
	// swaps on each pool.
	for i, routeStep := range route {
		// Get underlying pool type corresponding to the pool ID at the current routeStep.
		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}

		_tokenOut := tokenOut

		// If there is one pool left in the routeStep, set the expected output of the current swap
		// to the estimated input of the final pool.
		if i != len(route)-1 {
			_tokenOut = sdk.NewCoin(route[i+1].TokenInDenom, insExpected[i+1])
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := swapModule.GetPool(ctx, routeStep.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		// check if pool is active, if not error
		if !pool.IsActive(ctx) {
			return sdk.Int{}, types.InactivePoolError{PoolId: pool.GetId()}
		}

		spreadFactor := pool.GetSpreadFactor(ctx)
		// If we determined the routeStep is an osmo multi-hop and both route are incentivized,
		// we modify the swap fee accordingly.
		if isMultiHopRouted {
			spreadFactor = routeSpreadFactor.Mul((spreadFactor.Quo(sumOfSpreadFactors)))
		}

		_tokenInAmount, swapErr := swapModule.SwapExactAmountOut(ctx, sender, pool, routeStep.TokenInDenom, insExpected[i], _tokenOut, spreadFactor)
		if swapErr != nil {
			return sdk.Int{}, swapErr
		}

		// Sets the final amount of tokens that need to be input into the first pool. Even though this is the final return value for the
		// whole method and will not change after the first iteration, we still iterate through the rest of the pools to execute their respective
		// swaps.
		if i == 0 {
			tokenInAmount = _tokenInAmount
		}
	}

	return tokenInAmount, nil
}

// SplitRouteExactAmountOut route the swap across multiple multihop paths
// to get the desired token in. This is useful for achieving the most optimal execution. However, note that the responsibility
// of determining the optimal split is left to the client. This method simply route the swap across the given route.
// The route must end with the same token out and begin with the same token in.
//
// It performs the price impact protection check on the combination of tokens in from all multihop paths. The given tokenInMaxAmount
// is used for comparison.
//
// Returns error if:
//   - route are empty
//   - route contain duplicate multihop paths
//   - last token out denom is not the same for all multihop paths in routeStep
//   - one of the multihop swaps fails for internal reasons
//   - final token out computed is not positive
//   - final token out computed is smaller than tokenInMaxAmount
func (k Keeper) SplitRouteExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	route []types.SwapAmountOutSplitRoute,
	tokenOutDenom string,
	tokenInMaxAmount sdk.Int,
) (sdk.Int, error) {
	if err := types.ValidateSwapAmountOutSplitRoute(route); err != nil {
		return sdk.Int{}, err
	}

	var (
		// We start the multihop min amount as int max value
		// that is defined as one under the max bit length of sdk.Int
		// which is 256. This is to ensure that we utilize price impact protection
		// on the total of in amount from all multihop paths.
		multihopStartTokenInMaxAmount = intMaxValue
		totalInAmount                 = sdk.ZeroInt()
	)

	for _, multihopRoute := range route {
		tokenOutAmount, err := k.RouteExactAmountOut(
			ctx,
			sender,
			types.SwapAmountOutRoutes(multihopRoute.Pools),
			multihopStartTokenInMaxAmount,
			sdk.NewCoin(tokenOutDenom, multihopRoute.TokenOutAmount))
		if err != nil {
			return sdk.Int{}, err
		}

		totalInAmount = totalInAmount.Add(tokenOutAmount)
	}

	if !totalInAmount.IsPositive() {
		return sdk.Int{}, types.FinalAmountIsNotPositiveError{IsAmountOut: false, Amount: totalInAmount}
	}

	if totalInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, types.PriceImpactProtectionExactOutError{Actual: totalInAmount, MaxAmount: tokenInMaxAmount}
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSplitRouteSwapExactAmountOut,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
			sdk.NewAttribute(types.AttributeKeyTokensOut, totalInAmount.String()),
		),
	})

	return totalInAmount, nil
}

func (k Keeper) RouteGetPoolDenoms(
	ctx sdk.Context,
	poolId uint64,
) (denoms []string, err error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return []string{}, err
	}

	denoms, err = swapModule.GetPoolDenoms(ctx, poolId)
	if err != nil {
		return []string{}, err
	}

	return denoms, nil
}

func (k Keeper) RouteCalculateSpotPrice(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
) (price sdk.Dec, err error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	price, err = swapModule.CalculateSpotPrice(ctx, poolId, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	return price, nil
}

func (k Keeper) MultihopEstimateInGivenExactAmountOut(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	isMultiHopRouted, routeSpreadFactor, sumOfSpreadFactors := false, sdk.Dec{}, sdk.Dec{}
	var insExpected []sdk.Int

	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			insExpected = []sdk.Int{}
			err = fmt.Errorf("function MultihopEstimateInGivenExactAmountOut failed due to internal reason: %v", r)
		}
	}()

	routeStep := types.SwapAmountOutRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return sdk.Int{}, err
	}

	if k.isOsmoRoutedMultihop(ctx, routeStep, route[0].TokenInDenom, tokenOut.Denom) {
		isMultiHopRouted = true
		routeSpreadFactor, sumOfSpreadFactors, err = k.getOsmoRoutedMultihopTotalSpreadFactor(ctx, routeStep)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	// Determine what the estimated input would be for each pool along the multi-hop route
	// if we determined the route is an osmo multi-hop and both routes are incentivized,
	// we utilize a separate function that calculates the discounted spread factors
	if isMultiHopRouted {
		insExpected, err = k.createOsmoMultihopExpectedSwapOuts(ctx, route, tokenOut, routeSpreadFactor, sumOfSpreadFactors)
	} else {
		insExpected, err = k.createMultihopExpectedSwapOuts(ctx, route, tokenOut)
	}
	if err != nil {
		return sdk.Int{}, err
	}
	if len(insExpected) == 0 {
		return sdk.Int{}, nil
	}

	return insExpected[0], nil
}

func (k Keeper) GetPool(
	ctx sdk.Context,
	poolId uint64,
) (types.PoolI, error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return nil, err
	}

	return swapModule.GetPool(ctx, poolId)
}

// AllPools returns all pools sorted by their ids
// from every pool module registered in the
// pool manager keeper.
func (k Keeper) AllPools(
	ctx sdk.Context,
) ([]types.PoolI, error) {
	less := func(i, j types.PoolI) bool {
		return i.GetId() < j.GetId()
	}

	//	Allocate the slice with the exact capacity to avoid reallocations.
	poolCount := k.GetNextPoolId(ctx)
	sortedPools := make([]types.PoolI, 0, poolCount)
	for _, poolModule := range k.poolModules {
		currentModulePools, err := poolModule.GetPools(ctx)
		if err != nil {
			return nil, err
		}

		sortedPools = osmoutils.MergeSlices(sortedPools, currentModulePools, less)
	}

	return sortedPools, nil
}

// IsOsmoRoutedMultihop determines if a multi-hop swap involves OSMO, as one of the intermediary tokens.
func (k Keeper) isOsmoRoutedMultihop(ctx sdk.Context, route types.MultihopRoute, inDenom, outDenom string) (isRouted bool) {
	if route.Length() != 2 {
		return false
	}
	intemediateDenoms := route.IntermediateDenoms()
	if len(intemediateDenoms) != 1 || intemediateDenoms[0] != appparams.BaseCoinUnit {
		return false
	}
	if inDenom == outDenom {
		return false
	}
	poolIds := route.PoolIds()
	if poolIds[0] == poolIds[1] {
		return false
	}

	route0Incentivized := k.poolIncentivesKeeper.IsPoolIncentivized(ctx, poolIds[0])
	route1Incentivized := k.poolIncentivesKeeper.IsPoolIncentivized(ctx, poolIds[1])

	return route0Incentivized && route1Incentivized
}

// getOsmoRoutedMultihopTotalSpreadFactor calculates and returns the average swap fee and the sum of swap fees for
// a given route. For the former, it sets a lower bound of the highest swap fee pool in the route to ensure total
// swap fees for a route are never more than halved.
func (k Keeper) getOsmoRoutedMultihopTotalSpreadFactor(ctx sdk.Context, route types.MultihopRoute) (
	totalPathSpreadFactor sdk.Dec, sumOfSpreadFactors sdk.Dec, err error,
) {
	additiveSpreadFactor := sdk.ZeroDec()
	maxSpreadFactor := sdk.ZeroDec()

	for _, poolId := range route.PoolIds() {
		swapModule, err := k.GetPoolModule(ctx, poolId)
		if err != nil {
			return sdk.Dec{}, sdk.Dec{}, err
		}

		pool, poolErr := swapModule.GetPool(ctx, poolId)
		if poolErr != nil {
			return sdk.Dec{}, sdk.Dec{}, poolErr
		}
		SpreadFactor := pool.GetSpreadFactor(ctx)
		additiveSpreadFactor = additiveSpreadFactor.Add(SpreadFactor)
		maxSpreadFactor = sdk.MaxDec(maxSpreadFactor, SpreadFactor)
	}

	// We divide by 2 to get the average since OSMO-routed multihops always have exactly 2 pools.
	averageSpreadFactor := additiveSpreadFactor.QuoInt64(2)

	// We take the max here as a guardrail to ensure that there is a lowerbound on the swap fee for the
	// whole route equivalent to the highest fee pool
	routeSpreadFactor := sdk.MaxDec(maxSpreadFactor, averageSpreadFactor)

	return routeSpreadFactor, additiveSpreadFactor, nil
}

// createMultihopExpectedSwapOuts defines the output denom and output amount for the last pool in
// the routeStep of pools the caller is intending to hop through in a fixed-output multihop tx. It estimates the input
// amount for this last pool and then chains that input as the output of the previous pool in the routeStep, repeating
// until the first pool is reached. It returns an array of inputs, each of which correspond to a pool ID in the
// routeStep of pools for the original multihop transaction.
func (k Keeper) createMultihopExpectedSwapOuts(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) ([]sdk.Int, error) {
	insExpected := make([]sdk.Int, len(route))
	for i := len(route) - 1; i >= 0; i-- {
		routeStep := route[i]

		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return nil, err
		}

		poolI, err := swapModule.GetPool(ctx, routeStep.PoolId)
		if err != nil {
			return nil, err
		}

		tokenIn, err := swapModule.CalcInAmtGivenOut(ctx, poolI, tokenOut, routeStep.TokenInDenom, poolI.GetSpreadFactor(ctx))
		if err != nil {
			return nil, err
		}

		insExpected[i] = tokenIn.Amount
		tokenOut = tokenIn
	}

	return insExpected, nil
}

// createOsmoMultihopExpectedSwapOuts does the same as createMultihopExpectedSwapOuts, however discounts the swap fee.
func (k Keeper) createOsmoMultihopExpectedSwapOuts(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
	cumulativeRouteSpreadFactor, sumOfSpreadFactors sdk.Dec,
) ([]sdk.Int, error) {
	insExpected := make([]sdk.Int, len(route))
	for i := len(route) - 1; i >= 0; i-- {
		routeStep := route[i]

		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return nil, err
		}

		poolI, err := swapModule.GetPool(ctx, routeStep.PoolId)
		if err != nil {
			return nil, err
		}

		spreadFactor := poolI.GetSpreadFactor(ctx)
		tokenIn, err := swapModule.CalcInAmtGivenOut(ctx, poolI, tokenOut, routeStep.TokenInDenom, cumulativeRouteSpreadFactor.Mul((spreadFactor.Quo(sumOfSpreadFactors))))
		if err != nil {
			return nil, err
		}

		insExpected[i] = tokenIn.Amount
		tokenOut = tokenIn
	}

	return insExpected, nil
}

// GetTotalPoolLiquidity gets the total liquidity for a given poolId.
func (k Keeper) GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return nil, err
	}

	coins, err := swapModule.GetTotalPoolLiquidity(ctx, poolId)
	if err != nil {
		return coins, err
	}

	return coins, nil
}

// TotalLiquidity gets the total liquidity across all pools.
func (k Keeper) TotalLiquidity(ctx sdk.Context) (sdk.Coins, error) {
	totalGammLiquidity, err := k.gammKeeper.GetTotalLiquidity(ctx)
	if err != nil {
		return nil, err
	}
	totalConcentratedLiquidity, err := k.concentratedKeeper.GetTotalLiquidity(ctx)
	if err != nil {
		return nil, err
	}
	totalCosmwasmLiquidity, err := k.cosmwasmpoolKeeper.GetTotalLiquidity(ctx)
	if err != nil {
		return nil, err
	}
	totalLiquidity := totalGammLiquidity.Add(totalConcentratedLiquidity...).Add(totalCosmwasmLiquidity...)
	return totalLiquidity, nil
}
