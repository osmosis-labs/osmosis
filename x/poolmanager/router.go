package poolmanager

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var (
	// 1 << 256 - 1 where 256 is the max bit length defined for osmomath.Int
	intMaxValue = osmomath.NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)))
	// lessPoolIFunc is used for sorting pools by poolID
	lessPoolIFunc = func(i, j types.PoolI) bool {
		return i.GetId() < j.GetId()
	}
)

func (k *Keeper) GetPoolModuleAndPool(ctx sdk.Context, poolId uint64) (swapModule types.PoolModuleI, pool types.PoolI, err error) {
	// Get the pool-specific module implementation to ensure that
	// swaps are routed to the pool type corresponding to pool ID's pool.
	swapModule, err = k.GetPoolModule(ctx, poolId)
	if err != nil {
		return
	}

	// Get pool as a general pool type. Note that the underlying function used
	// still varies with the pool type.
	pool, err = swapModule.GetPool(ctx, poolId)
	if err != nil {
		return
	}
	return
}

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
	// Ensure that provided route is not empty and has valid denom format.
	if err := types.SwapAmountInRoutes(route).Validate(); err != nil {
		return osmomath.Int{}, err
	}

	totalTakerFeesCharged := sdk.Coins{}
	denomsInvolvedInRoute := []string{tokenIn.Denom}

	// Iterate through the route and execute a series of swaps through each pool.
	for i, routeStep := range route {
		// To prevent the multihop swap from being interrupted prematurely, we keep
		// the minimum expected output at a very low number until the last pool
		_outMinAmount := osmomath.NewInt(1)
		if len(route)-1 == i {
			_outMinAmount = tokenOutMinAmount
		}

		var takerFeeCharged sdk.Coin
		tokenOutAmount, takerFeeCharged, err = k.SwapExactAmountIn(ctx, sender, routeStep.PoolId, tokenIn, routeStep.TokenOutDenom, _outMinAmount)
		if err != nil {
			return osmomath.Int{}, err
		}

		// Chain output of current pool as the input for the next routed pool
		tokenIn = sdk.NewCoin(routeStep.TokenOutDenom, tokenOutAmount)

		// Track taker fees charged
		totalTakerFeesCharged = totalTakerFeesCharged.Add(takerFeeCharged)

		// Add the token out denom to the denoms involved in the route, IFF it is not already in the slice
		if !osmoutils.Contains(denomsInvolvedInRoute, routeStep.TokenOutDenom) {
			denomsInvolvedInRoute = append(denomsInvolvedInRoute, routeStep.TokenOutDenom)
		}
	}

	// Run taker fee skim logic
	err = k.TakerFeeSkim(ctx, denomsInvolvedInRoute, totalTakerFeesCharged)
	if err != nil {
		return osmomath.Int{}, err
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
) (tokenOutAmount osmomath.Int, takerFeeCharged sdk.Coin, err error) {
	swapModule, pool, err := k.GetPoolModuleAndPool(ctx, poolId)
	if err != nil {
		return osmomath.Int{}, sdk.Coin{}, err
	}

	// Check if pool has swaps enabled.
	if !pool.IsActive(ctx) {
		return osmomath.Int{}, sdk.Coin{}, fmt.Errorf("pool %d is not active", pool.GetId())
	}

	tokenInAfterSubTakerFee, takerFeeCharged, err := k.chargeTakerFee(ctx, tokenIn, tokenOutDenom, sender, true)
	if err != nil {
		return osmomath.Int{}, sdk.Coin{}, err
	}

	// routeStep to the pool-specific SwapExactAmountIn implementation.
	tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenInAfterSubTakerFee, tokenOutDenom, tokenOutMinAmount, pool.GetSpreadFactor(ctx))
	if err != nil {
		return osmomath.Int{}, sdk.Coin{}, err
	}

	// Track volume for volume-splitting incentives
	k.trackVolume(ctx, pool.GetId(), tokenIn)

	return tokenOutAmount, takerFeeCharged, nil
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
	swapModule, pool, err := k.GetPoolModuleAndPool(ctx, poolId)
	if err != nil {
		return osmomath.Int{}, err
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

	// Track volume for volume-splitting incentives
	k.trackVolume(ctx, pool.GetId(), tokenIn)

	return tokenOutAmount, nil
}

func (k Keeper) MultihopEstimateOutGivenExactAmountInNoTakerFee(
	ctx sdk.Context,
	route []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
) (tokenOutAmount osmomath.Int, err error) {
	return k.multihopEstimateOutGivenExactAmountInInternal(ctx, route, tokenIn, false)
}

func (k Keeper) MultihopEstimateOutGivenExactAmountIn(
	ctx sdk.Context,
	route []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
) (tokenOutAmount osmomath.Int, err error) {
	return k.multihopEstimateOutGivenExactAmountInInternal(ctx, route, tokenIn, true)
}

func (k Keeper) multihopEstimateOutGivenExactAmountInInternal(
	ctx sdk.Context,
	route []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
	applyTakerFee bool,
) (tokenOutAmount osmomath.Int, err error) {
	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			tokenOutAmount = osmomath.Int{}
			if isErr, d := osmoutils.IsOutOfGasError(r); isErr {
				err = fmt.Errorf("function MultihopEstimateOutGivenExactAmountIn failed due to lack of gas: %v", d)
			} else {
				err = fmt.Errorf("function MultihopEstimateOutGivenExactAmountIn failed due to internal reason: %v", r)
			}
		}
	}()

	if err := types.SwapAmountInRoutes(route).Validate(); err != nil {
		return osmomath.Int{}, err
	}

	for _, routeStep := range route {
		swapModule, poolI, err := k.GetPoolModuleAndPool(ctx, routeStep.PoolId)
		if err != nil {
			return osmomath.Int{}, err
		}

		spreadFactor := poolI.GetSpreadFactor(ctx)

		actualTokenIn := tokenIn
		// apply taker fee if applicable
		if applyTakerFee {
			takerFee, err := k.GetTradingPairTakerFee(ctx, tokenIn.Denom, routeStep.TokenOutDenom)
			if err != nil {
				return osmomath.Int{}, err
			}

			actualTokenIn, _ = CalcTakerFeeExactIn(tokenIn, takerFee)
		}

		tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, actualTokenIn, routeStep.TokenOutDenom, spreadFactor)
		if err != nil {
			return osmomath.Int{}, err
		}

		tokenOutAmount = tokenOut.Amount
		if !tokenOutAmount.IsPositive() {
			return osmomath.Int{}, errors.New("token amount must be positive")
		}

		// Chain output of current pool as the input for the next routed pool
		// We don't need to validate the denom,
		// as CalcOutAmtGivenIn is responsible for ensuring the denom exists in the pool.
		tokenIn = sdk.Coin{Denom: routeStep.TokenOutDenom, Amount: tokenOutAmount}
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
	if err := types.SwapAmountOutRoutes(route).Validate(); err != nil {
		return osmomath.Int{}, err
	}

	defer func() {
		if r := recover(); r != nil {
			tokenInAmount = osmomath.Int{}
			if isErr, d := osmoutils.IsOutOfGasError(r); isErr {
				err = fmt.Errorf("function RouteExactAmountOut failed due to lack of gas: %v", d)
			} else {
				err = fmt.Errorf("function RouteExactAmountOut failed due to internal reason: %v", r)
			}
		}
	}()

	var insExpected []osmomath.Int
	insExpected, err = k.createMultihopExpectedSwapOuts(ctx, route, tokenOut)

	if err != nil {
		return osmomath.Int{}, err
	}
	if len(insExpected) == 0 {
		return osmomath.Int{}, nil
	}
	insExpected[0] = tokenInMaxAmount

	totalTakerFeesCharged := sdk.Coins{}
	denomsInvolvedInRoute := []string{tokenOut.Denom}

	// Iterates through each routed pool and executes their respective swaps. Note that all of the work to get the return
	// value of this method is done when we calculate insExpected – this for loop primarily serves to execute the actual
	// swaps on each pool.
	for i, routeStep := range route {
		swapModule, pool, err := k.GetPoolModuleAndPool(ctx, routeStep.PoolId)
		if err != nil {
			return osmomath.Int{}, err
		}

		_tokenOut := tokenOut

		// If there is one pool left in the routeStep, set the expected output of the current swap
		// to the estimated input of the final pool.
		if i != len(route)-1 {
			_tokenOut = sdk.NewCoin(route[i+1].TokenInDenom, insExpected[i+1])
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
		tokenInAfterAddTakerFee, takerFeeCharged, err := k.chargeTakerFee(ctx, tokenIn, _tokenOut.Denom, sender, false)
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

		// Track taker fees charged
		totalTakerFeesCharged = totalTakerFeesCharged.Add(takerFeeCharged)

		// Add the token in denom to the denoms involved in the route, IFF it is not already in the slice
		if !osmoutils.Contains(denomsInvolvedInRoute, routeStep.TokenInDenom) {
			denomsInvolvedInRoute = append(denomsInvolvedInRoute, routeStep.TokenInDenom)
		}
	}

	// Run taker fee skim logic
	err = k.TakerFeeSkim(ctx, denomsInvolvedInRoute, totalTakerFeesCharged)
	if err != nil {
		return osmomath.Int{}, err
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
) (price osmomath.BigDec, err error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	price, err = swapModule.CalculateSpotPrice(ctx, poolId, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	return price, nil
}

func (k Keeper) MultihopEstimateInGivenExactAmountOut(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) (tokenInAmount osmomath.Int, err error) {
	var insExpected []osmomath.Int

	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			insExpected = []osmomath.Int{}
			if isErr, d := osmoutils.IsOutOfGasError(r); isErr {
				err = fmt.Errorf("function MultihopEstimateInGivenExactAmountOut failed due to lack of gas: %v", d)
			} else {
				err = fmt.Errorf("function MultihopEstimateInGivenExactAmountOut failed due to internal reason: %v", r)
			}
		}
	}()

	routeStep := types.SwapAmountOutRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return osmomath.Int{}, err
	}

	// Determine what the estimated input would be for each pool along the multi-hop route
	insExpected, err = k.createMultihopExpectedSwapOuts(ctx, route, tokenOut)
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
	_, pool, err := k.GetPoolModuleAndPool(ctx, poolId)
	return pool, err
}

// AllPools returns all pools sorted by their ids
// from every pool module registered in the
// pool manager keeper.
func (k Keeper) AllPools(
	ctx sdk.Context,
) ([]types.PoolI, error) {
	//	Allocate the slice with the exact capacity to avoid reallocations.
	poolCount := k.GetNextPoolId(ctx)
	sortedPools := make([]types.PoolI, 0, poolCount)
	for _, poolModule := range k.poolModules {
		currentModulePools, err := poolModule.GetPools(ctx)
		if err != nil {
			return nil, err
		}

		sortedPools = osmoutils.MergeSlices(sortedPools, currentModulePools, lessPoolIFunc)
	}

	return sortedPools, nil
}

// ListPoolsByDenom returns all pools by denom sorted by their ids
// from every pool module registered in the
// pool manager keeper.
// N.B. It is possible for incorrectly implemented pools to be skipped
func (k Keeper) ListPoolsByDenom(
	ctx sdk.Context,
	denom string,
) ([]types.PoolI, error) {
	var sortedPools []types.PoolI
	for _, poolModule := range k.poolModules {
		currentModulePools, err := poolModule.GetPools(ctx)
		if err != nil {
			return nil, err
		}

		var poolsByDenom []types.PoolI
		for _, pool := range currentModulePools {
			// If the pool is incorrectly implemented and we can't get the PoolDenoms
			// skip the pool.
			poolDenoms, err := poolModule.GetPoolDenoms(ctx, pool.GetId())
			if err != nil {
				ctx.Logger().Debug(fmt.Sprintf("Error getting pool denoms for pool %d: %s", pool.GetId(), err.Error()))
				continue
			}
			if osmoutils.Contains(poolDenoms, denom) {
				poolsByDenom = append(poolsByDenom, pool)
			}
		}
		sortedPools = osmoutils.MergeSlices(sortedPools, poolsByDenom, lessPoolIFunc)
	}
	return sortedPools, nil
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

		swapModule, poolI, err := k.GetPoolModuleAndPool(ctx, routeStep.PoolId)
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

		tokenInAfterTakerFee, _ := CalcTakerFeeExactOut(tokenIn, takerFee)

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

// nolint: unused
// trackVolume converts the input token into OSMO units and adds it to the global tracked volume for the given pool ID.
// Fails quietly if an OSMO paired pool cannot be found, although this should only happen in rare scenarios where OSMO is
// removed as a base denom from the protorev module (which this function relies on).
//
// CONTRACT: `volumeGenerated` corresponds to one of the denoms in the pool
// CONTRACT: pool with `poolId` exists
func (k Keeper) trackVolume(ctx sdk.Context, poolId uint64, volumeGenerated sdk.Coin) {
	// If the denom is already denominated in uosmo, we can just use it directly
	OSMO, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		panic(err)
	}
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
	volumeInOsmo := osmomath.BigDecFromSDKInt(volumeGenerated.Amount).Mul(osmoPerInputToken).Dec().TruncateInt()

	// Add this new volume to the global tracked volume for the pool ID
	k.addVolume(ctx, poolId, sdk.NewCoin(OSMO, volumeInOsmo))
}

// addVolume adds the given volume to the global tracked volume for the given pool ID.
func (k Keeper) addVolume(ctx sdk.Context, poolId uint64, volumeGenerated sdk.Coin) {
	// Get the current volume for the pool ID
	currentTotalVolume := k.GetTotalVolumeForPool(ctx, poolId)

	// Add newly generated volume to existing volume and set updated volume in state
	newTotalVolume := currentTotalVolume.Add(volumeGenerated)
	k.SetVolume(ctx, poolId, newTotalVolume)
}

// SetVolume sets the given volume to the global tracked volume for the given pool ID.
// Note that this function is exported for cross-module testing purposes and should not be
// called directly from other modules.
func (k Keeper) SetVolume(ctx sdk.Context, poolId uint64, totalVolume sdk.Coins) {
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
	OSMO, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		panic(err)
	}
	return totalVolume.AmountOf(OSMO)
}

// EstimateTradeBasedOnPriceImpactBalancerPool estimates a trade based on price impact for a balancer pool type.
// For a balancer pool if an amount entered is greater than the total pool liquidity the trade estimated would be
// the full liquidity of the other token. If the amount is small it would return a close 1:1 trade of the
// smallest units.
func (k Keeper) EstimateTradeBasedOnPriceImpactBalancerPool(
	ctx sdk.Context,
	req queryproto.EstimateTradeBasedOnPriceImpactRequest,
	spotPrice, adjustedMaxPriceImpact osmomath.Dec,
	swapModule types.PoolModuleI,
	poolI types.PoolI,
) (*queryproto.EstimateTradeBasedOnPriceImpactResponse, error) {
	tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, req.FromCoin, req.ToCoinDenom, types.ZeroDec)
	if err != nil {
		if errors.Is(err, gammtypes.ErrInvalidMathApprox) {
			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
				OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
			}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if tokenOut.IsZero() {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
			OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
		}, nil
	}

	// Validate if the trade as is respects the price impact, if it does re-estimate it with a swap fee and return
	// the result.
	priceDeviation := calculatePriceDeviation(req.FromCoin, tokenOut, spotPrice)
	if priceDeviation.LTE(adjustedMaxPriceImpact) {
		tokenOut, err = swapModule.CalcOutAmtGivenIn(
			ctx, poolI, req.FromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
		)
		if err != nil {
			if errors.Is(err, gammtypes.ErrInvalidMathApprox) {
				return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
					InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
					OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
				}, nil
			}
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  req.FromCoin,
			OutputCoin: tokenOut,
		}, nil
	}

	// Define low and high amount to search between. Start from 1 and req.FromCoin.Amount as initial range.
	lowAmount := osmomath.OneInt()
	highAmount := req.FromCoin.Amount
	currFromCoin := req.FromCoin

	// Repeat the above process using the binary search algorithm which iteratively narrows down the optimal trade
	// amount within a given maximum price impact range.
	//
	// The algorithm iteratively:
	// 1) Calculates the middle amount of the current range ('midAmount').
	// 2) Tries to execute a trade using this middle amount.
	// 3) Calculates the resulting price deviation between the spot price and the
	//    price of the tried trade.
	//
	// Depending on whether the price deviation is within the allowed 'adjustedMaxPriceImpact',
	// the algorithm adjusts the 'lowAmount' or 'highAmount' for the next iteration.
	//
	// This process continues until 'lowAmount' is greater than 'highAmount', at which
	// point the optimal amount respecting the max price impact will have been found.
	for lowAmount.LTE(highAmount) {
		// Calculate currFromCoin as the new middle amount to try trade.
		midAmount := lowAmount.Add(highAmount).Quo(osmomath.NewInt(2))
		currFromCoin = sdk.NewCoin(req.FromCoin.Denom, midAmount)

		tokenOut, err := swapModule.CalcOutAmtGivenIn(
			ctx, poolI, currFromCoin, req.ToCoinDenom, types.ZeroDec,
		)
		if err != nil {
			if errors.Is(err, gammtypes.ErrInvalidMathApprox) {
				return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
					InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
					OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
				}, nil
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		if tokenOut.IsZero() {
			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
				OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
			}, nil
		}

		priceDeviation := calculatePriceDeviation(currFromCoin, tokenOut, spotPrice)
		if priceDeviation.LTE(adjustedMaxPriceImpact) {
			lowAmount = midAmount.Add(osmomath.OneInt())
		} else {
			highAmount = midAmount.Sub(osmomath.OneInt())
		}
	}

	// highAmount is 0 it means the loop has iterated to the end without finding a viable trade that respects
	// the price impact.
	if highAmount.IsZero() {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
			OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
		}, nil
	}

	tokenOut, err = swapModule.CalcOutAmtGivenIn(
		ctx, poolI, currFromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
		InputCoin:  currFromCoin,
		OutputCoin: tokenOut,
	}, nil
}

// EstimateTradeBasedOnPriceImpactStableSwapPool estimates a trade based on price impact for a stableswap pool type.
// For a stableswap pool if an amount entered is greater than the total pool liquidity the trade estimated would
// `panic`. If the amount is small it would return an error, in the case of a `panic` we should ignore it
// and keep attempting lower input amounts while if it's a normal error we should return an empty trade.
func (k Keeper) EstimateTradeBasedOnPriceImpactStableSwapPool(
	ctx sdk.Context,
	req queryproto.EstimateTradeBasedOnPriceImpactRequest,
	spotPrice, adjustedMaxPriceImpact osmomath.Dec,
	swapModule types.PoolModuleI,
	poolI types.PoolI,
) (*queryproto.EstimateTradeBasedOnPriceImpactResponse, error) {
	var tokenOut sdk.Coin
	var err error
	err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		tokenOut, err = swapModule.CalcOutAmtGivenIn(ctx, poolI, req.FromCoin, req.ToCoinDenom, types.ZeroDec)
		return err
	})

	// Find out if the error is because the amount is too large or too little. The calculation should error
	// if the amount is too small, and it should panic if the amount is too large. If the amount is too large
	// we want to continue to iterate to find attempt to find a smaller value. StableSwap panics on amounts that
	// are too large due to the maths involved, while Balancer pool types do not.
	if err != nil && !strings.Contains(err.Error(), "panic") {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
			OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
		}, nil
	} else if err == nil {
		// Validate if the trade as is respects the price impact, if it does re-estimate it with a swap fee and return
		// the result.
		priceDeviation := calculatePriceDeviation(req.FromCoin, tokenOut, spotPrice)
		if priceDeviation.LTE(adjustedMaxPriceImpact) {
			tokenOut, err = swapModule.CalcOutAmtGivenIn(
				ctx, poolI, req.FromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
			)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}

			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  req.FromCoin,
				OutputCoin: tokenOut,
			}, nil
		}
	}

	// Define low and high amount to search between. Start from 1 and req.FromCoin.Amount as initial range.
	lowAmount := osmomath.OneInt()
	highAmount := req.FromCoin.Amount
	currFromCoin := req.FromCoin

	// Repeat the above process using the binary search algorithm which iteratively narrows down the optimal trade
	// amount within a given maximum price impact range.
	//
	// The algorithm iteratively:
	// 1) Calculates the middle amount of the current range ('midAmount').
	// 2) Tries to execute a trade using this middle amount.
	// 3) Calculates the resulting price deviation between the spot price and the
	//    price of the tried trade.
	//
	// Depending on whether the price deviation is within the allowed 'adjustedMaxPriceImpact',
	// the algorithm adjusts the 'lowAmount' or 'highAmount' for the next iteration.
	//
	// This process continues until 'lowAmount' is greater than 'highAmount', at which
	// point the optimal amount respecting the max price impact will have been found.
	for lowAmount.LTE(highAmount) {
		// Calculate currFromCoin as the new middle amount to try trade.
		midAmount := lowAmount.Add(highAmount).Quo(osmomath.NewInt(2))
		currFromCoin = sdk.NewCoin(req.FromCoin.Denom, midAmount)

		err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			tokenOut, err = swapModule.CalcOutAmtGivenIn(ctx, poolI, currFromCoin, req.ToCoinDenom, types.ZeroDec)
			return err
		})

		// If it returns an error without a panic it means the input has become too small and we should return.
		// This occurs for the StableSwap pool type due to the maths involved, this does not occur for Balancer
		// pool types.
		if err != nil && !strings.Contains(err.Error(), "panic") {
			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
				OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
			}, nil
		} else if err != nil {
			// If there is an error that does contain a panic it means the amount is still too large,
			// and we should continue halving.
			highAmount = midAmount.Sub(osmomath.OneInt())
		} else {
			priceDeviation := calculatePriceDeviation(currFromCoin, tokenOut, spotPrice)
			if priceDeviation.LTE(adjustedMaxPriceImpact) {
				lowAmount = midAmount.Add(osmomath.OneInt())
			} else {
				highAmount = midAmount.Sub(osmomath.OneInt())
			}
		}
	}

	// highAmount is 0 it means the loop has iterated to the end without finding a viable trade that respects
	// the price impact.
	if highAmount.IsZero() {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
			OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
		}, nil
	}

	tokenOut, err = swapModule.CalcOutAmtGivenIn(
		ctx, poolI, currFromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
		InputCoin:  currFromCoin,
		OutputCoin: tokenOut,
	}, nil
}

// EstimateTradeBasedOnPriceImpactConcentratedLiquidity estimates a trade based on price impact for a concentrated
// liquidity pool type. For a concentrated liquidity pool if an amount entered is greater than the total pool liquidity
// the trade estimated would error. If the amount is small it would return tokenOut to be 0 in which case we should
// return an empty trade. If the estimate returns an error we should ignore it and continue attempting to estimate
// by halving the input.
func (k Keeper) EstimateTradeBasedOnPriceImpactConcentratedLiquidity(
	ctx sdk.Context,
	req queryproto.EstimateTradeBasedOnPriceImpactRequest,
	spotPrice, adjustedMaxPriceImpact osmomath.Dec,
	swapModule types.PoolModuleI,
	poolI types.PoolI,
) (*queryproto.EstimateTradeBasedOnPriceImpactResponse, error) {
	tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, req.FromCoin, req.ToCoinDenom, types.ZeroDec)
	// If there was no error we attempt to validate if the output is below the adjustedMaxPriceImpact.
	if err == nil {
		// If the tokenOut was returned to be zero it means the amount being traded is too small. We ignore the
		// error output here as it could mean that the input is too large.
		if tokenOut.IsZero() {
			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
				OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
			}, nil
		}

		priceDeviation := calculatePriceDeviation(req.FromCoin, tokenOut, spotPrice)
		if priceDeviation.LTE(adjustedMaxPriceImpact) {
			tokenOut, err = swapModule.CalcOutAmtGivenIn(
				ctx, poolI, req.FromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
			)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}

			return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
				InputCoin:  req.FromCoin,
				OutputCoin: tokenOut,
			}, nil
		}
	}

	// Define low and high amount to search between. Start from 1 and req.FromCoin.Amount as initial range.
	lowAmount := osmomath.OneInt()
	highAmount := req.FromCoin.Amount
	currFromCoin := req.FromCoin

	// Repeat the above process using the binary search algorithm which iteratively narrows down the optimal trade
	// amount within a given maximum price impact range.
	//
	// The algorithm iteratively:
	// 1) Calculates the middle amount of the current range ('midAmount').
	// 2) Tries to execute a trade using this middle amount.
	// 3) Calculates the resulting price deviation between the spot price and the
	//    price of the tried trade.
	//
	// Depending on whether the price deviation is within the allowed 'adjustedMaxPriceImpact',
	// the algorithm adjusts the 'lowAmount' or 'highAmount' for the next iteration.
	//
	// This process continues until 'lowAmount' is greater than 'highAmount', at which
	// point the optimal amount respecting the max price impact will have been found.
	for lowAmount.LTE(highAmount) {
		// Calculate currFromCoin as the new middle amount to try trade.
		midAmount := lowAmount.Add(highAmount).Quo(osmomath.NewInt(2))
		currFromCoin = sdk.NewCoin(req.FromCoin.Denom, midAmount)

		tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, currFromCoin, req.ToCoinDenom, types.ZeroDec)
		if err == nil {
			// If the tokenOut was returned to be zero it means the amount being traded is too small. We ignore the
			// error output here as it could mean that the input is too large.
			if tokenOut.IsZero() {
				return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
					InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
					OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
				}, nil
			}

			priceDeviation := calculatePriceDeviation(currFromCoin, tokenOut, spotPrice)
			if priceDeviation.LTE(adjustedMaxPriceImpact) {
				lowAmount = midAmount.Add(osmomath.OneInt())
			} else {
				highAmount = midAmount.Sub(osmomath.OneInt())
			}
		} else {
			highAmount = midAmount.Sub(osmomath.OneInt())
		}
	}

	// highAmount is 0 it means the loop has iterated to the end without finding a viable trade that respects
	// the price impact.
	if highAmount.IsZero() {
		return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
			InputCoin:  sdk.NewCoin(req.FromCoin.Denom, osmomath.ZeroInt()),
			OutputCoin: sdk.NewCoin(req.ToCoinDenom, osmomath.ZeroInt()),
		}, nil
	}

	tokenOut, err = swapModule.CalcOutAmtGivenIn(
		ctx, poolI, currFromCoin, req.ToCoinDenom, poolI.GetSpreadFactor(ctx),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.EstimateTradeBasedOnPriceImpactResponse{
		InputCoin:  currFromCoin,
		OutputCoin: tokenOut,
	}, nil
}

// calculatePriceDeviation calculates the price deviation between the current trade price and the spot price.
// We have an `Abs()` at the end of the priceDeviation equation as we cannot be sure if any pool types based on their
// configurations trade out more tokens than given for a trade, it is added just in-case.
func calculatePriceDeviation(currFromCoin, tokenOut sdk.Coin, spotPrice osmomath.Dec) osmomath.Dec {
	currTradePrice := osmomath.NewDec(currFromCoin.Amount.Int64()).QuoInt(tokenOut.Amount)
	priceDeviation := currTradePrice.Sub(spotPrice).Quo(spotPrice).Abs()
	return priceDeviation
}
