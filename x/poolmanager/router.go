package poolmanager

import (
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	appparams "github.com/osmosis-labs/osmosis/v15/app/params"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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
		isMultiHopRouted bool
		routeSwapFee     sdk.Dec
		sumOfSwapFees    sdk.Dec
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
	// total_swap_fee = max(swapfee1, swapfee2)
	// fee_per_pool = total_swap_fee * ((pool_fee) / (swapfee1 + swapfee2))
	if k.isOsmoRoutedMultihop(ctx, routeStep, route[0].TokenOutDenom, tokenIn.Denom) {
		isMultiHopRouted = true
		routeSwapFee, sumOfSwapFees, err = k.getOsmoRoutedMultihopTotalSwapFee(ctx, routeStep)
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

		swapFee := pool.GetSwapFee(ctx)

		// If we determined the routeStep is an osmo multi-hop and both route are incentivized,
		// we modify the swap fee accordingly.
		if isMultiHopRouted {
			swapFee = routeSwapFee.MulRoundUp((swapFee.QuoRoundUp(sumOfSwapFees)))
		}

		tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, routeStep.TokenOutDenom, _outMinAmount, swapFee)
		if err != nil {
			return sdk.Int{}, err
		}

		// Chain output of current pool as the input for the next routed pool
		tokenIn = sdk.NewCoin(routeStep.TokenOutDenom, tokenOutAmount)
	}
	return tokenOutAmount, nil
}

// SplitRouteExactAmountIn route the swap across multiple multihop paths
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
	route []types.SwapAmountInSplitRoute,
	tokenInDenom string,
	tokenOutMinAmount sdk.Int,
) (sdk.Int, error) {
	if err := types.ValidateSwapAmountInSplitRoute(route); err != nil {
		return sdk.Int{}, err
	}

	var (
		// We start the multihop min amount as zero because we want
		// to perform a price impact protection check on the combination of tokens out
		// from all multihop paths.
		multihopStartTokenOutMinAmount = sdk.ZeroInt()
		totalOutAmount                 = sdk.ZeroInt()
	)

	for _, multihopRoute := range route {
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

	swapFee := pool.GetSwapFee(ctx)

	// routeStep to the pool-specific SwapExactAmountIn implementation.
	tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, swapFee)
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
		isMultiHopRouted bool
		routeSwapFee     sdk.Dec
		sumOfSwapFees    sdk.Dec
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
		routeSwapFee, sumOfSwapFees, err = k.getOsmoRoutedMultihopTotalSwapFee(ctx, routeStep)
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

		swapFee := poolI.GetSwapFee(ctx)

		// If we determined the routeStep is an osmo multi-hop and both route are incentivized,
		// we modify the swap fee accordingly.
		if isMultiHopRouted {
			swapFee = routeSwapFee.Mul((swapFee.Quo(sumOfSwapFees)))
		}

		tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, tokenIn, routeStep.TokenOutDenom, swapFee)
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

// MultihopSwapExactAmountOut defines the output denom and output amount for the last pool.
// Calculation starts by providing the tokenOutAmount of the final pool to calculate the required tokenInAmount
// the calculated tokenInAmount is used as defined tokenOutAmount of the previous pool, calculating in reverse order of the swap
// Transaction succeeds if the calculated tokenInAmount of the first pool is less than the defined tokenInMaxAmount defined.
func (k Keeper) RouteExactAmountOut(ctx sdk.Context,
	sender sdk.AccAddress,
	route []types.SwapAmountOutRoute,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	isMultiHopRouted, routeSwapFee, sumOfSwapFees := false, sdk.Dec{}, sdk.Dec{}
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

	// in this loop, we check if:
	// - the routeStep is of length 2
	// - routeStep 1 and routeStep 2 don't trade via the same pool
	// - routeStep 1 contains uosmo
	// - both routeStep 1 and routeStep 2 are incentivized pools
	// if all of the above is true, then we collect the additive and max fee between the two pools to later calculate the following:
	// total_swap_fee = total_swap_fee = max(swapfee1, swapfee2)
	// fee_per_pool = total_swap_fee * ((pool_fee) / (swapfee1 + swapfee2))
	if k.isOsmoRoutedMultihop(ctx, routeStep, route[0].TokenInDenom, tokenOut.Denom) {
		isMultiHopRouted = true
		routeSwapFee, sumOfSwapFees, err = k.getOsmoRoutedMultihopTotalSwapFee(ctx, routeStep)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	// Determine what the estimated input would be for each pool along the multi-hop routeStep
	// if we determined the routeStep is an osmo multi-hop and both route are incentivized,
	// we utilize a separate function that calculates the discounted swap fees
	var insExpected []sdk.Int
	if isMultiHopRouted {
		insExpected, err = k.createOsmoMultihopExpectedSwapOuts(ctx, route, tokenOut, routeSwapFee, sumOfSwapFees)
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
			return sdk.Int{}, fmt.Errorf("pool %d is not active", pool.GetId())
		}

		swapFee := pool.GetSwapFee(ctx)
		if isMultiHopRouted {
			swapFee = routeSwapFee.Mul((swapFee.Quo(sumOfSwapFees)))
		}

		_tokenInAmount, swapErr := swapModule.SwapExactAmountOut(ctx, sender, pool, routeStep.TokenInDenom, insExpected[i], _tokenOut, swapFee)
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
	isMultiHopRouted, routeSwapFee, sumOfSwapFees := false, sdk.Dec{}, sdk.Dec{}
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
		routeSwapFee, sumOfSwapFees, err = k.getOsmoRoutedMultihopTotalSwapFee(ctx, routeStep)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	// Determine what the estimated input would be for each pool along the multi-hop routeStep
	// if we determined the routeStep is an osmo multi-hop and both route are incentivized,
	// we utilize a separate function that calculates the discounted swap fees
	if isMultiHopRouted {
		insExpected, err = k.createOsmoMultihopExpectedSwapOuts(ctx, route, tokenOut, routeSwapFee, sumOfSwapFees)
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

// getOsmoRoutedMultihopTotalSwapFee calculates and returns the average swap fee and the sum of swap fees for
// a given route. For the former, it sets a lower bound of the highest swap fee pool in the route to ensure total
// swap fees for a route are never more than halved.
func (k Keeper) getOsmoRoutedMultihopTotalSwapFee(ctx sdk.Context, route types.MultihopRoute) (
	totalPathSwapFee sdk.Dec, sumOfSwapFees sdk.Dec, err error,
) {
	additiveSwapFee := sdk.ZeroDec()
	highestSwapFee := sdk.ZeroDec()

	for _, poolId := range route.PoolIds() {
		swapModule, err := k.GetPoolModule(ctx, poolId)
		if err != nil {
			return sdk.Dec{}, sdk.Dec{}, err
		}

		pool, poolErr := swapModule.GetPool(ctx, poolId)
		if poolErr != nil {
			return sdk.Dec{}, sdk.Dec{}, poolErr
		}
		swapFee := pool.GetSwapFee(ctx)
		additiveSwapFee = additiveSwapFee.Add(swapFee)
		highestSwapFee = sdk.MaxDec(highestSwapFee, swapFee)
	}

	// We divide by 2 to get the average since OSMO-routed multihops always have exactly 2 pools.
	averageSwapFee := additiveSwapFee.QuoInt64(2)

	// We take the max here as a guardrail to ensure that there is a lowerbound on the swap fee for the
	// whole route equivalent to the highest fee pool
	routeSwapFee := sdk.MaxDec(highestSwapFee, averageSwapFee)

	return routeSwapFee, additiveSwapFee, nil
}

// createMultihopExpectedSwapOuts defines the output denom and output amount for the last pool in
// the routeStep of pools the caller is intending to hop through in a fixed-output multihop tx. It estimates the input
// amount for this last pool and then chains that input as the output of the previous pool in the routeStep, repeating
// until the first pool is reached. It returns an array of inputs, each of which correspond to a pool ID in the
// routeStep of pools for the original multihop transaction.
// TODO: test this.
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

		tokenIn, err := swapModule.CalcInAmtGivenOut(ctx, poolI, tokenOut, routeStep.TokenInDenom, poolI.GetSwapFee(ctx))
		if err != nil {
			return nil, err
		}

		insExpected[i] = tokenIn.Amount
		tokenOut = tokenIn
	}

	return insExpected, nil
}

// createOsmoMultihopExpectedSwapOuts does the same as createMultihopExpectedSwapOuts, however discounts the swap fee
func (k Keeper) createOsmoMultihopExpectedSwapOuts(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
	cumulativeRouteSwapFee, sumOfSwapFees sdk.Dec,
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

		swapFee := poolI.GetSwapFee(ctx)
		tokenIn, err := swapModule.CalcInAmtGivenOut(ctx, poolI, tokenOut, routeStep.TokenInDenom, cumulativeRouteSwapFee.Mul((swapFee.Quo(sumOfSwapFees))))
		if err != nil {
			return nil, err
		}

		insExpected[i] = tokenIn.Amount
		tokenOut = tokenIn
	}

	return insExpected, nil
}

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
