package poolmanager

import (
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v19/app/params"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

// 1 << 256 - 1 where 256 is the max bit length defined for osmomath.Int
var intMaxValue = osmomath.NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)))

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
	tokenOutMinAmount osmomath.Int,
) (tokenOutAmount osmomath.Int, err error) {
	var (
		isMultiHopRouted   bool
		routeSpreadFactor  osmomath.Dec
		sumOfSpreadFactors osmomath.Dec
	)

	// Ensure that provided route is not empty and has valid denom format.
	routeStep := types.SwapAmountInRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return osmomath.Int{}, err
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
			return osmomath.Int{}, err
		}
	}

	// Iterate through the route and execute a series of swaps through each pool.
	for i, routeStep := range route {
		// To prevent the multihop swap from being interrupted prematurely, we keep
		// the minimum expected output at a very low number until the last pool
		_outMinAmount := osmomath.NewInt(1)
		if len(route)-1 == i {
			_outMinAmount = tokenOutMinAmount
		}

		// Get underlying pool type corresponding to the pool ID at the current routeStep.
		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return osmomath.Int{}, err
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := swapModule.GetPool(ctx, routeStep.PoolId)
		if poolErr != nil {
			return osmomath.Int{}, poolErr
		}

		// Check if pool has swaps enabled.
		if !pool.IsActive(ctx) {
			return osmomath.Int{}, types.InactivePoolError{PoolId: pool.GetId()}
		}

		spreadFactor := pool.GetSpreadFactor(ctx)

		// If we determined the route is an osmo multi-hop and both routes are incentivized,
		// we modify the spread factor accordingly.
		if isMultiHopRouted {
			spreadFactor = routeSpreadFactor.MulRoundUp((spreadFactor.QuoRoundUp(sumOfSpreadFactors)))
		}

		tokenInAfterSubTakerFee, err := k.chargeTakerFee(ctx, tokenIn, routeStep.TokenOutDenom, sender, true)
		if err != nil {
			return osmomath.Int{}, err
		}

		tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenInAfterSubTakerFee, routeStep.TokenOutDenom, _outMinAmount, spreadFactor)
		if err != nil {
			return osmomath.Int{}, err
		}

		// Track volume for volume-splitting incentives
		k.trackVolume(ctx, pool.GetId(), tokenIn)

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
	tokenOutMinAmount osmomath.Int,
) (osmomath.Int, error) {
	if err := types.ValidateSwapAmountInSplitRoute(routes); err != nil {
		return osmomath.Int{}, err
	}

	var (
		// We start the multihop min amount as zero because we want
		// to perform a price impact protection check on the combination of tokens out
		// from all multihop paths.
		multihopStartTokenOutMinAmount = osmomath.ZeroInt()
		totalOutAmount                 = osmomath.ZeroInt()
	)

	for _, multihopRoute := range routes {
		tokenOutAmount, err := k.RouteExactAmountIn(
			ctx,
			sender,
			types.SwapAmountInRoutes(multihopRoute.Pools),
			sdk.NewCoin(tokenInDenom, multihopRoute.TokenInAmount),
			multihopStartTokenOutMinAmount)
		if err != nil {
			return osmomath.Int{}, err
		}

		totalOutAmount = totalOutAmount.Add(tokenOutAmount)
	}

	if !totalOutAmount.IsPositive() {
		return osmomath.Int{}, types.FinalAmountIsNotPositiveError{IsAmountOut: true, Amount: totalOutAmount}
	}

	if totalOutAmount.LT(tokenOutMinAmount) {
		return osmomath.Int{}, types.PriceImpactProtectionExactInError{Actual: totalOutAmount, MinAmount: tokenOutMinAmount}
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
	tokenOutMinAmount osmomath.Int,
) (tokenOutAmount osmomath.Int, err error) {
	// Get the pool-specific module implementation to ensure that
	// swaps are routed to the pool type corresponding to pool ID's pool.
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return osmomath.Int{}, err
	}

	// Get pool as a general pool type. Note that the underlying function used
	// still varies with the pool type.
	pool, poolErr := swapModule.GetPool(ctx, poolId)
	if poolErr != nil {
		return osmomath.Int{}, poolErr
	}

	// Check if pool has swaps enabled.
	if !pool.IsActive(ctx) {
		return osmomath.Int{}, fmt.Errorf("pool %d is not active", pool.GetId())
	}

	tokenInAfterSubTakerFee, err := k.chargeTakerFee(ctx, tokenIn, tokenOutDenom, sender, true)
	if err != nil {
		return osmomath.Int{}, err
	}

	// routeStep to the pool-specific SwapExactAmountIn implementation.
	tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenInAfterSubTakerFee, tokenOutDenom, tokenOutMinAmount, pool.GetSpreadFactor(ctx))
	if err != nil {
		return osmomath.Int{}, err
	}

	return tokenOutAmount, nil
}

// SwapExactAmountInNoTakerFee is an API for swapping an exact amount of tokens
// as input to a pool to get a minimum amount of the desired token out.
// This method does NOT charge a taker fee, and should only be used in txfees hooks
// when swapping taker fees. This prevents us from charging taker fees
// on top of taker fees.
func (k Keeper) SwapExactAmountInNoTakerFee(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount osmomath.Int,
) (tokenOutAmount osmomath.Int, err error) {
	// Get the pool-specific module implementation to ensure that
	// swaps are routed to the pool type corresponding to pool ID's pool.
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return osmomath.Int{}, err
	}

	// Get pool as a general pool type. Note that the underlying function used
	// still varies with the pool type.
	pool, poolErr := swapModule.GetPool(ctx, poolId)
	if poolErr != nil {
		return osmomath.Int{}, poolErr
	}

	// Check if pool has swaps enabled.
	if !pool.IsActive(ctx) {
		return osmomath.Int{}, fmt.Errorf("pool %d is not active", pool.GetId())
	}

	// routeStep to the pool-specific SwapExactAmountIn implementation.
	tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, pool.GetSpreadFactor(ctx))
	if err != nil {
		return osmomath.Int{}, err
	}

	return tokenOutAmount, nil
}

func (k Keeper) MultihopEstimateOutGivenExactAmountIn(
	ctx sdk.Context,
	route []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
) (tokenOutAmount osmomath.Int, err error) {
	var (
		isMultiHopRouted   bool
		routeSpreadFactor  osmomath.Dec
		sumOfSpreadFactors osmomath.Dec
	)

	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			tokenOutAmount = osmomath.Int{}
			err = fmt.Errorf("function MultihopEstimateOutGivenExactAmountIn failed due to internal reason: %v", r)
		}
	}()

	routeStep := types.SwapAmountInRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return osmomath.Int{}, err
	}

	if k.isOsmoRoutedMultihop(ctx, routeStep, route[0].TokenOutDenom, tokenIn.Denom) {
		isMultiHopRouted = true
		routeSpreadFactor, sumOfSpreadFactors, err = k.getOsmoRoutedMultihopTotalSpreadFactor(ctx, routeStep)
		if err != nil {
			return osmomath.Int{}, err
		}
	}

	for _, routeStep := range route {
		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return osmomath.Int{}, err
		}

		// Execute the expected swap on the current routed pool
		poolI, poolErr := swapModule.GetPool(ctx, routeStep.PoolId)
		if poolErr != nil {
			return osmomath.Int{}, poolErr
		}

		spreadFactor := poolI.GetSpreadFactor(ctx)

		// If we determined the routeStep is an osmo multi-hop and both route are incentivized,
		// we modify the swap fee accordingly.
		if isMultiHopRouted {
			spreadFactor = routeSpreadFactor.Mul((spreadFactor.Quo(sumOfSpreadFactors)))
		}

		takerFee, err := k.GetTradingPairTakerFee(ctx, routeStep.TokenOutDenom, tokenIn.Denom)
		if err != nil {
			return osmomath.Int{}, err
		}

		tokenInAfterSubTakerFee, _ := k.calcTakerFeeExactIn(tokenIn, takerFee)

		tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, tokenInAfterSubTakerFee, routeStep.TokenOutDenom, spreadFactor)
		if err != nil {
			return osmomath.Int{}, err
		}

		tokenOutAmount = tokenOut.Amount
		if !tokenOutAmount.IsPositive() {
			return osmomath.Int{}, errors.New("token amount must be positive")
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
	tokenInMaxAmount osmomath.Int,
	tokenOut sdk.Coin,
) (tokenInAmount osmomath.Int, err error) {
	isMultiHopRouted, routeSpreadFactor, sumOfSpreadFactors := false, osmomath.Dec{}, osmomath.Dec{}
	// Ensure that provided route is not empty and has valid denom format.
	routeStep := types.SwapAmountOutRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return osmomath.Int{}, err
	}

	defer func() {
		if r := recover(); r != nil {
			tokenInAmount = osmomath.Int{}
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
	var insExpected []osmomath.Int
	isMultiHopRouted = k.isOsmoRoutedMultihop(ctx, routeStep, route[0].TokenInDenom, tokenOut.Denom)

	// Determine what the estimated input would be for each pool along the multi-hop routeStep
	// if we determined the routeStep is an osmo multi-hop and both route are incentivized,
	// we utilize a separate function that calculates the discounted swap fees
	if isMultiHopRouted {
		routeSpreadFactor, sumOfSpreadFactors, err = k.getOsmoRoutedMultihopTotalSpreadFactor(ctx, routeStep)
		if err != nil {
			return osmomath.Int{}, err
		}
		insExpected, err = k.createOsmoMultihopExpectedSwapOuts(ctx, route, tokenOut, routeSpreadFactor, sumOfSpreadFactors)
	} else {
		insExpected, err = k.createMultihopExpectedSwapOuts(ctx, route, tokenOut)
	}

	if err != nil {
		return osmomath.Int{}, err
	}
	if len(insExpected) == 0 {
		return osmomath.Int{}, nil
	}
	insExpected[0] = tokenInMaxAmount

	// Iterates through each routed pool and executes their respective swaps. Note that all of the work to get the return
	// value of this method is done when we calculate insExpected – this for loop primarily serves to execute the actual
	// swaps on each pool.
	for i, routeStep := range route {
		// Get underlying pool type corresponding to the pool ID at the current routeStep.
		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return osmomath.Int{}, err
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
			return osmomath.Int{}, poolErr
		}

		// check if pool is active, if not error
		if !pool.IsActive(ctx) {
			return osmomath.Int{}, types.InactivePoolError{PoolId: pool.GetId()}
		}

		spreadFactor := pool.GetSpreadFactor(ctx)
		// If we determined the routeStep is an osmo multi-hop and both route are incentivized,
		// we modify the swap fee accordingly.
		if isMultiHopRouted {
			spreadFactor = routeSpreadFactor.Mul((spreadFactor.Quo(sumOfSpreadFactors)))
		}

		curTokenInAmount, swapErr := swapModule.SwapExactAmountOut(ctx, sender, pool, routeStep.TokenInDenom, insExpected[i], _tokenOut, spreadFactor)
		if swapErr != nil {
			return osmomath.Int{}, swapErr
		}

		tokenIn := sdk.NewCoin(routeStep.TokenInDenom, curTokenInAmount)
		tokenInAfterAddTakerFee, err := k.chargeTakerFee(ctx, tokenIn, _tokenOut.Denom, sender, false)
		if err != nil {
			return osmomath.Int{}, err
		}

		// Track volume for volume-splitting incentives
		k.trackVolume(ctx, pool.GetId(), sdk.NewCoin(routeStep.TokenInDenom, tokenIn.Amount))

		// Sets the final amount of tokens that need to be input into the first pool. Even though this is the final return value for the
		// whole method and will not change after the first iteration, we still iterate through the rest of the pools to execute their respective
		// swaps.
		if i == 0 {
			tokenInAmount = tokenInAfterAddTakerFee.Amount
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
	tokenInMaxAmount osmomath.Int,
) (osmomath.Int, error) {
	if err := types.ValidateSwapAmountOutSplitRoute(route); err != nil {
		return osmomath.Int{}, err
	}

	var (
		// We start the multihop min amount as int max value
		// that is defined as one under the max bit length of osmomath.Int
		// which is 256. This is to ensure that we utilize price impact protection
		// on the total of in amount from all multihop paths.
		multihopStartTokenInMaxAmount = intMaxValue
		totalInAmount                 = osmomath.ZeroInt()
	)

	for _, multihopRoute := range route {
		tokenOutAmount, err := k.RouteExactAmountOut(
			ctx,
			sender,
			types.SwapAmountOutRoutes(multihopRoute.Pools),
			multihopStartTokenInMaxAmount,
			sdk.NewCoin(tokenOutDenom, multihopRoute.TokenOutAmount))
		if err != nil {
			return osmomath.Int{}, err
		}

		totalInAmount = totalInAmount.Add(tokenOutAmount)
	}

	if !totalInAmount.IsPositive() {
		return osmomath.Int{}, types.FinalAmountIsNotPositiveError{IsAmountOut: false, Amount: totalInAmount}
	}

	if totalInAmount.GT(tokenInMaxAmount) {
		return osmomath.Int{}, types.PriceImpactProtectionExactOutError{Actual: totalInAmount, MaxAmount: tokenInMaxAmount}
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
) (price osmomath.Dec, err error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return osmomath.Dec{}, err
	}

	price, err = swapModule.CalculateSpotPrice(ctx, poolId, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return osmomath.Dec{}, err
	}

	return price, nil
}

func (k Keeper) MultihopEstimateInGivenExactAmountOut(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) (tokenInAmount osmomath.Int, err error) {
	isMultiHopRouted, routeSpreadFactor, sumOfSpreadFactors := false, osmomath.Dec{}, osmomath.Dec{}
	var insExpected []osmomath.Int

	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			insExpected = []osmomath.Int{}
			err = fmt.Errorf("function MultihopEstimateInGivenExactAmountOut failed due to internal reason: %v", r)
		}
	}()

	routeStep := types.SwapAmountOutRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return osmomath.Int{}, err
	}

	if k.isOsmoRoutedMultihop(ctx, routeStep, route[0].TokenInDenom, tokenOut.Denom) {
		isMultiHopRouted = true
		routeSpreadFactor, sumOfSpreadFactors, err = k.getOsmoRoutedMultihopTotalSpreadFactor(ctx, routeStep)
		if err != nil {
			return osmomath.Int{}, err
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
		return osmomath.Int{}, err
	}
	if len(insExpected) == 0 {
		return osmomath.Int{}, nil
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
	totalPathSpreadFactor osmomath.Dec, sumOfSpreadFactors osmomath.Dec, err error,
) {
	additiveSpreadFactor := osmomath.ZeroDec()
	maxSpreadFactor := osmomath.ZeroDec()

	for _, poolId := range route.PoolIds() {
		swapModule, err := k.GetPoolModule(ctx, poolId)
		if err != nil {
			return osmomath.Dec{}, osmomath.Dec{}, err
		}

		pool, poolErr := swapModule.GetPool(ctx, poolId)
		if poolErr != nil {
			return osmomath.Dec{}, osmomath.Dec{}, poolErr
		}
		spreadFactor := pool.GetSpreadFactor(ctx)
		additiveSpreadFactor = additiveSpreadFactor.Add(spreadFactor)
		maxSpreadFactor = sdk.MaxDec(maxSpreadFactor, spreadFactor)
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
) ([]osmomath.Int, error) {
	insExpected := make([]osmomath.Int, len(route))
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

		takerFee, err := k.GetTradingPairTakerFee(ctx, routeStep.TokenInDenom, tokenOut.Denom)
		if err != nil {
			return nil, err
		}

		tokenIn, err := swapModule.CalcInAmtGivenOut(ctx, poolI, tokenOut, routeStep.TokenInDenom, spreadFactor)
		if err != nil {
			return nil, err
		}

		tokenInAfterTakerFee, _ := k.calcTakerFeeExactOut(tokenIn, takerFee)

		insExpected[i] = tokenInAfterTakerFee.Amount
		tokenOut = tokenInAfterTakerFee
	}

	return insExpected, nil
}

// createOsmoMultihopExpectedSwapOuts does the same as createMultihopExpectedSwapOuts, however discounts the swap fee.
func (k Keeper) createOsmoMultihopExpectedSwapOuts(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
	cumulativeRouteSpreadFactor, sumOfSpreadFactors osmomath.Dec,
) ([]osmomath.Int, error) {
	insExpected := make([]osmomath.Int, len(route))
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

		takerFee, err := k.GetTradingPairTakerFee(ctx, routeStep.TokenInDenom, tokenOut.Denom)
		if err != nil {
			return nil, err
		}

		osmoDiscountedSpreadFactor := cumulativeRouteSpreadFactor.Mul((spreadFactor.Quo(sumOfSpreadFactors)))

		tokenIn, err := swapModule.CalcInAmtGivenOut(ctx, poolI, tokenOut, routeStep.TokenInDenom, osmoDiscountedSpreadFactor)
		if err != nil {
			return nil, err
		}

		tokenInAfterTakerFee, _ := k.calcTakerFeeExactOut(tokenIn, takerFee)

		insExpected[i] = tokenInAfterTakerFee.Amount
		tokenOut = tokenInAfterTakerFee
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

// isDenomWhitelisted checks if the denom provided exists in the list of authorized quote denoms.
// If it does, it returns true, otherwise false.
func isDenomWhitelisted(denom string, authorizedQuoteDenoms []string) bool {
	for _, authorizedQuoteDenom := range authorizedQuoteDenoms {
		if denom == authorizedQuoteDenom {
			return true
		}
	}
	return false
}

// nolint: unused
// trackVolume converts the input token into OSMO units and adds it to the global tracked volume for the given pool ID.
// Fails quietly if an OSMO paired pool cannot be found, although this should only happen in rare scenarios where OSMO is
// removed as a base denom from the protorev module (which this function relies on).
//
// CONTRACT: `volumeGenerated` corresponds to one of the denoms in the pool
// CONTRACT: pool with `poolId` exists
func (k Keeper) trackVolume(ctx sdk.Context, poolId uint64, volumeGenerated sdk.Coin) {
	// If the denom is already denominated in uosmo, we can just use it directly
	OSMO := k.stakingKeeper.BondDenom(ctx)
	if volumeGenerated.Denom == OSMO {
		k.addVolume(ctx, poolId, volumeGenerated)
		return
	}

	// Get the most liquid OSMO-paired pool with `volumeGenerated`'s denom using `GetPoolForDenomPair`
	osmoPairedPoolId, err := k.protorevKeeper.GetPoolForDenomPair(ctx, OSMO, volumeGenerated.Denom)

	// If no pool is found, fail quietly.
	//
	// This is a rare scenario that should only happen if OSMO-paired pools are all removed from the protorev module.
	// Since this removal scenario is all-or-nothing, this is functionally equiavalent to freezing the tracked volume amounts
	// where they were prior to the disabling, which seems an appropriate response.
	//
	// This branch would also get triggered in the case where there is a token that has no OSMO-paired pool on the entire chain.
	// We simply do not track volume in these cases. Importantly, volume splitting gauge logic should prevent a gauge from being
	// created for such a pool that includes such a token, although it is okay to no-op in these cases regardless.
	if err != nil {
		return
	}

	// Since we want to ultimately multiply the volume by this spot price, we want to quote OSMO in terms of the input token.
	// This is so that once we multiply the volume by the spot price, we get the volume in units of OSMO.
	osmoPerInputToken, err := k.RouteCalculateSpotPrice(ctx, osmoPairedPoolId, OSMO, volumeGenerated.Denom)

	// We expect that if a pool is found, there should always be an available spot price as well.
	// That being said, if there is an error finding the spot price, we fail quietly and leave tracked volume unchanged.
	// This is because we do not want to escalate an issue with finding spot price to locking all swaps involving the given asset.
	if err != nil {
		return
	}

	// Multiply `volumeGenerated.Amount.ToDec()` by this spot price.
	// While rounding does not particularly matter here, we round down to ensure that we do not overcount volume.
	volumeInOsmo := volumeGenerated.Amount.ToLegacyDec().Mul(osmoPerInputToken).TruncateInt()

	// Add this new volume to the global tracked volume for the pool ID
	k.addVolume(ctx, poolId, sdk.NewCoin(OSMO, volumeInOsmo))
}

// nolint: unused
// addVolume adds the given volume to the global tracked volume for the given pool ID.
func (k Keeper) addVolume(ctx sdk.Context, poolId uint64, volumeGenerated sdk.Coin) {
	// Get the current volume for the pool ID
	currentTotalVolume := k.GetTotalVolumeForPool(ctx, poolId)

	// Add newly generated volume to existing volume and set updated volume in state
	newTotalVolume := currentTotalVolume.Add(volumeGenerated)
	k.setVolume(ctx, poolId, newTotalVolume)
}

// nolint: unused
// setVolume sets the given volume to the global tracked volume for the given pool ID.
func (k Keeper) setVolume(ctx sdk.Context, poolId uint64, totalVolume sdk.Coins) {
	storedVolume := types.TrackedVolume{Amount: totalVolume}
	osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyPoolVolume(poolId), &storedVolume)
}

// GetTotalVolumeForPool gets the total historical volume in all supported denominations for a given pool ID.
func (k Keeper) GetTotalVolumeForPool(ctx sdk.Context, poolId uint64) sdk.Coins {
	var currentTrackedVolume types.TrackedVolume
	volumeFound, err := osmoutils.Get(ctx.KVStore(k.storeKey), types.KeyPoolVolume(poolId), &currentTrackedVolume)
	if err != nil {
		// We can only encounter an error if a database or serialization errors occurs, so we panic here.
		// Normally this would be handled by `osmoutils.MustGet`, but since we want to specifically use `osmoutils.Get`,
		// we also have to manually panic here.
		panic(err)
	}

	// If no volume was found, we treat the existing volume as 0.
	// While we can technically require volume to exist, we would need to store empty coins in state for each pool (past and present),
	// which is a high storage cost to pay for a weak guardrail.
	currentTotalVolume := sdk.NewCoins()
	if volumeFound {
		currentTotalVolume = currentTrackedVolume.Amount
	}

	return currentTotalVolume
}

// GetOsmoVolumeForPool gets the total OSMO-denominated historical volume for a given pool ID.
func (k Keeper) GetOsmoVolumeForPool(ctx sdk.Context, poolId uint64) osmomath.Int {
	totalVolume := k.GetTotalVolumeForPool(ctx, poolId)
	return totalVolume.AmountOf(k.stakingKeeper.BondDenom(ctx))
}
