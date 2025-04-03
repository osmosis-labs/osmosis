package concentrated_liquidity

import (
	fmt "fmt"

	db "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	events "github.com/osmosis-labs/osmosis/v27/x/poolmanager/events"

	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// SwapState defines the state of a swap.
// It is initialized as the swap begins and is updated after every swap step.
// Once the swap is complete, this state is either returned to the estimate
// swap querier or committed to state.
type SwapState struct {
	// Remaining amount of specified token.
	// if out given in, amount of token being swapped in.
	// if in given out, amount of token being swapped out.
	// Initialized to the amount of the token specified by the user.
	// Updated after every swap step.
	amountSpecifiedRemaining osmomath.Dec

	// Amount of the other token that is calculated from the specified token.
	// if out given in, amount of token swapped out.
	// if in given out, amount of token swapped in.
	// Initialized to zero.
	// Updated after every swap step.
	amountCalculated osmomath.Dec

	// Current sqrt price while calculating swap.
	// Initialized to the pool's current sqrt price.
	// Updated after every swap step.
	sqrtPrice osmomath.BigDec

	// Current tick while calculating swap.
	// Initialized to the pool's current tick.
	// Updated each time a tick is crossed.
	tick int64

	// Current liqudiity within the active tick.
	// Initialized to the pool's current tick's liquidity.
	// Updated each time a tick is crossed.
	liquidity osmomath.Dec

	// Global spread reward growth per-current swap.
	// Initialized to zero.
	// Updated after every swap step.
	// This value is persisted in the pool's tick accumulators.
	globalSpreadRewardGrowthPerUnitLiquidity osmomath.Dec

	// Global spread reward growth
	// This value is not persisted in the pool's tick accumulators, and is instead
	// used to determine how much the user must send to the spread reward address.
	//
	// NOTE: Even if there is a scaling factor, this value is not scaled.
	// The scaling factor only applies when we are updating the pool's tick accumulators.
	globalSpreadRewardGrowth osmomath.Dec

	swapStrategy swapstrategy.SwapStrategy
}

// SwapResult represents the result from computing
// the swap state.
type SwapResult struct {
	AmountIn      osmomath.Int
	AmountOut     osmomath.Int
	SpreadRewards osmomath.Dec
}

// swapNoProgressLimit is the maximum number of iterations that can be performed
// without progressing the swap state. If this limit is reached, the swap is
// considered to have failed.
// Note, the value is chosen arbitrarily.
// From tests, there should be no reason for a swap to make more than 2 iterations without
// progress. However, we leave a buffer of 1_000 to account for any unforeseen edge cases.
const swapNoProgressLimit = 100

func newSwapState(specifiedAmount osmomath.Int, p types.ConcentratedPoolExtension, strategy swapstrategy.SwapStrategy) SwapState {
	return SwapState{
		amountSpecifiedRemaining:                 specifiedAmount.ToLegacyDec(),
		amountCalculated:                         osmomath.ZeroDec(),
		sqrtPrice:                                p.GetCurrentSqrtPrice(),
		tick:                                     p.GetCurrentTick(),
		liquidity:                                p.GetLiquidity(),
		globalSpreadRewardGrowthPerUnitLiquidity: osmomath.ZeroDec(),
		globalSpreadRewardGrowth:                 osmomath.ZeroDec(),
		swapStrategy:                             strategy,
	}
}

type SwapDetails struct {
	Sender   sdk.AccAddress
	TokenIn  sdk.Coin
	TokenOut sdk.Coin
}

type PoolUpdates struct {
	NewCurrentTick int64
	NewLiquidity   osmomath.Dec
	NewSqrtPrice   osmomath.BigDec
}

var (
	smallestDec = osmomath.SmallestDec()
)

// updateSpreadRewardGrowthGlobal updates the swap state's spread reward growth global per unit of liquidity
// when liquidity is positive.
//
// If the liquidity is zero, this is a no-op. This case may occur when there is no liquidity
// between the ticks. This is possible when there are only 2 positions with no overlapping ranges.
// As a result, the range from the end of position one to the beginning of position
// two has no liquidity and can be skipped.
//
// Returns the spread factors accrued per unit of liquidity.
func (ss *SwapState) updateSpreadRewardGrowthGlobal(spreadRewardChargeTotal osmomath.Dec, scalingFactor osmomath.Dec) (osmomath.Dec, error) {
	// Initialize spreadRewardChargeTotalScaled with the original unscaled value
	spreadRewardChargeTotalScaled := spreadRewardChargeTotal

	// Only scale if the scaling factor is not equal to oneDecScalingFactor
	if !scalingFactor.Equal(oneDecScalingFactor) {
		var err error
		spreadRewardChargeTotalScaled, err = scaleUpTotalEmittedAmount(spreadRewardChargeTotal, scalingFactor)
		if err != nil {
			return osmomath.ZeroDec(), fmt.Errorf("failed to scale up spread reward charge: %w", err)
		}
	}

	// Update global spread reward growth with the UNSCALED total
	// We could scale this up, but we would need to scale it right back down at the end of the swap.
	// The globalSpreadRewardGrowth is used to determine how much the user sends to the spread reward address,
	// and is not persisted in the pool's tick accumulators.
	ss.globalSpreadRewardGrowth = ss.globalSpreadRewardGrowth.Add(spreadRewardChargeTotal)

	// If liquidity is zero, return early to avoid division by zero
	if ss.liquidity.IsZero() {
		return osmomath.ZeroDec(), nil
	}

	// Calculate spread factors accrued per unit of liquidity, rounding down to avoid overdistribution
	spreadFactorsAccruedPerUnitOfLiquidityScaled := spreadRewardChargeTotalScaled.QuoTruncate(ss.liquidity)

	// Update global spread reward growth per unit liquidity
	ss.globalSpreadRewardGrowthPerUnitLiquidity.AddMut(spreadFactorsAccruedPerUnitOfLiquidityScaled)

	return spreadFactorsAccruedPerUnitOfLiquidityScaled, nil
}

func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolI poolmanagertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount osmomath.Int,
	spreadFactor osmomath.Dec,
) (tokenOutAmount osmomath.Int, err error) {
	if tokenIn.Denom == tokenOutDenom {
		return osmomath.Int{}, types.DenomDuplicatedError{TokenInDenom: tokenIn.Denom, TokenOutDenom: tokenOutDenom}
	}

	// Convert pool interface to CL pool type
	pool, err := asConcentrated(poolI)
	if err != nil {
		return osmomath.Int{}, err
	}

	// Trigger before hook for SwapExactAmountIn prior to mutating state.
	// If no contract is set, this will be a no-op.
	err = k.BeforeSwapExactAmountIn(ctx, pool.GetId(), sender, tokenIn, tokenOutDenom, tokenOutMinAmount, spreadFactor)
	if err != nil {
		return osmomath.Int{}, err
	}

	// Determine if we are swapping asset0 for asset1 or vice versa
	zeroForOne := getZeroForOne(tokenIn.Denom, pool.GetToken0())

	// Change priceLimit based on which direction we are swapping
	priceLimit := swapstrategy.GetPriceLimit(zeroForOne)
	tokenIn, tokenOut, _, err := k.swapOutAmtGivenIn(ctx, sender, pool, tokenIn, tokenOutDenom, spreadFactor, priceLimit)
	if err != nil {
		return osmomath.Int{}, err
	}
	tokenOutAmount = tokenOut.Amount

	// price impact protection.
	if tokenOutAmount.LT(tokenOutMinAmount) {
		return osmomath.Int{}, types.AmountLessThanMinError{TokenAmount: tokenOutAmount, TokenMin: tokenOutMinAmount}
	}

	k.RecordTotalLiquidityIncrease(ctx, sdk.NewCoins(tokenIn))
	k.RecordTotalLiquidityDecrease(ctx, sdk.NewCoins(tokenOut))

	// Trigger after hook for SwapExactAmountIn after mutating state.
	// If no contract is set, this will be a no-op.
	err = k.AfterSwapExactAmountIn(ctx, pool.GetId(), sender, tokenIn, tokenOutDenom, tokenOutMinAmount, spreadFactor)
	if err != nil {
		return osmomath.Int{}, err
	}

	return tokenOutAmount, nil
}

// SwapExactAmountOut allows users to specify the output token amount they want to receive from a swap and get the exact
// input token amount they need to provide based on the current pool prices and any applicable spread factors.
func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolI poolmanagertypes.PoolI,
	tokenInDenom string,
	tokenInMaxAmount osmomath.Int,
	tokenOut sdk.Coin,
	spreadFactor osmomath.Dec,
) (tokenInAmount osmomath.Int, err error) {
	if tokenOut.Denom == tokenInDenom {
		return osmomath.Int{}, types.DenomDuplicatedError{TokenInDenom: tokenInDenom, TokenOutDenom: tokenOut.Denom}
	}

	pool, err := asConcentrated(poolI)
	if err != nil {
		return osmomath.Int{}, err
	}

	// Trigger before hook for SwapExactAmountOut prior to mutating state.
	// If no contract is set, this will be a no-op.
	err = k.BeforeSwapExactAmountOut(ctx, pool.GetId(), sender, tokenInDenom, tokenInMaxAmount, tokenOut, spreadFactor)
	if err != nil {
		return osmomath.Int{}, err
	}

	zeroForOne := getZeroForOne(tokenInDenom, pool.GetToken0())

	// change priceLimit based on which direction we are swapping
	// if zeroForOne == true, use MinSpotPrice else use MaxSpotPrice
	priceLimit := swapstrategy.GetPriceLimit(zeroForOne)
	tokenIn, tokenOut, _, err := k.swapInAmtGivenOut(ctx, sender, pool, tokenOut, tokenInDenom, spreadFactor, priceLimit)
	if err != nil {
		return osmomath.Int{}, err
	}
	tokenInAmount = tokenIn.Amount

	// price impact protection.
	if tokenInAmount.GT(tokenInMaxAmount) {
		return osmomath.Int{}, types.AmountGreaterThanMaxError{TokenAmount: tokenInAmount, TokenMax: tokenInMaxAmount}
	}

	k.RecordTotalLiquidityIncrease(ctx, sdk.NewCoins(tokenIn))
	k.RecordTotalLiquidityDecrease(ctx, sdk.NewCoins(tokenOut))

	// Trigger after hook for SwapExactAmountOut after mutating state.
	// If no contract is set, this will be a no-op.
	err = k.AfterSwapExactAmountOut(ctx, pool.GetId(), sender, tokenInDenom, tokenInMaxAmount, tokenOut, spreadFactor)
	if err != nil {
		return osmomath.Int{}, err
	}

	return tokenInAmount, nil
}

func (k Keeper) swapOutAmtGivenIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.ConcentratedPoolExtension,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	spreadFactor osmomath.Dec,
	priceLimit osmomath.BigDec,
) (calcTokenIn, calcTokenOut sdk.Coin, poolUpdates PoolUpdates, err error) {
	swapResult, poolUpdates, err := k.computeOutAmtGivenIn(ctx, pool.GetId(), tokenIn, tokenOutDenom, spreadFactor, priceLimit, true)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, PoolUpdates{}, err
	}
	tokenIn = sdk.NewCoin(tokenIn.Denom, swapResult.AmountIn)
	tokenOut := sdk.NewCoin(tokenOutDenom, swapResult.AmountOut)

	if !tokenOut.Amount.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, PoolUpdates{}, types.InvalidAmountCalculatedError{Amount: tokenOut.Amount}
	}

	// Settles balances between the tx sender and the pool to match the swap that was executed earlier.
	// Also emits swap event and updates related liquidity metrics
	if err := k.updatePoolForSwap(ctx, pool, SwapDetails{sender, tokenIn, tokenOut}, poolUpdates, swapResult.SpreadRewards); err != nil {
		return sdk.Coin{}, sdk.Coin{}, PoolUpdates{}, err
	}

	return tokenIn, tokenOut, poolUpdates, nil
}

// swapInAmtGivenOut is the internal mutative method for calcInAmtGivenOut. Utilizing calcInAmtGivenOut's output, this function applies the
// new tick, liquidity, and sqrtPrice to the respective pool.
func (k *Keeper) swapInAmtGivenOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.ConcentratedPoolExtension,
	desiredTokenOut sdk.Coin,
	tokenInDenom string,
	spreadFactor osmomath.Dec,
	priceLimit osmomath.BigDec,
) (calcTokenIn, calcTokenOut sdk.Coin, poolUpdates PoolUpdates, err error) {
	swapResult, poolUpdates, err := k.computeInAmtGivenOut(ctx, desiredTokenOut, tokenInDenom, spreadFactor, priceLimit, pool.GetId(), true)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, PoolUpdates{}, err
	}
	tokenIn := sdk.NewCoin(tokenInDenom, swapResult.AmountIn)
	tokenOut := sdk.NewCoin(desiredTokenOut.Denom, swapResult.AmountOut)

	// check that the tokenOut calculated is both valid and less than specified limit
	if !tokenIn.Amount.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, PoolUpdates{}, types.InvalidAmountCalculatedError{Amount: tokenIn.Amount}
	}

	// Settles balances between the tx sender and the pool to match the swap that was executed earlier.
	// Also emits swap event and updates related liquidity metrics
	if err := k.updatePoolForSwap(ctx, pool, SwapDetails{sender, tokenIn, tokenOut}, poolUpdates, swapResult.SpreadRewards); err != nil {
		return sdk.Coin{}, sdk.Coin{}, PoolUpdates{}, err
	}

	return tokenIn, tokenOut, poolUpdates, nil
}

var unboundedPriceLimit = osmomath.ZeroBigDec()

func (k Keeper) CalcOutAmtGivenIn(
	ctx sdk.Context,
	poolI poolmanagertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	spreadFactor osmomath.Dec,
) (tokenOut sdk.Coin, err error) {
	cacheCtx, _ := ctx.CacheContext()
	swapResult, _, err := k.computeOutAmtGivenIn(cacheCtx, poolI.GetId(), tokenIn, tokenOutDenom, spreadFactor, unboundedPriceLimit, false)
	if err != nil {
		return sdk.Coin{}, err
	}
	return sdk.NewCoin(tokenOutDenom, swapResult.AmountOut), nil
}

func (k Keeper) CalcInAmtGivenOut(
	ctx sdk.Context,
	poolI poolmanagertypes.PoolI,
	tokenOut sdk.Coin,
	tokenInDenom string,
	spreadFactor osmomath.Dec,
) (sdk.Coin, error) {
	cacheCtx, _ := ctx.CacheContext()
	swapResult, _, err := k.computeInAmtGivenOut(cacheCtx, tokenOut, tokenInDenom, spreadFactor, unboundedPriceLimit, poolI.GetId(), false)
	if err != nil {
		return sdk.Coin{}, err
	}
	return sdk.NewCoin(tokenInDenom, swapResult.AmountIn), nil
}

func (k Keeper) swapSetup(ctx sdk.Context,
	poolId uint64,
	tokenInDenom string,
	tokenOutDenom string,
	getAccumulators bool) (pool types.ConcentratedPoolExtension, spreadAccum *accum.AccumulatorObject, err error) {
	pool, err = k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return pool, spreadAccum, err
	}
	if err := checkDenomValidity(tokenInDenom, tokenOutDenom, pool.GetToken0(), pool.GetToken1()); err != nil {
		return pool, spreadAccum, err
	}
	if getAccumulators {
		spreadAccum, err = k.GetSpreadRewardAccumulator(ctx, poolId)
	}
	return pool, spreadAccum, err
}

// returns next initialized tick, next initialized tick sqrt price, implied sqrt price target, and error
// the next initialized tick sqrt price return arg is unfortunately there due to some bigdec situations.
// poolId as an argument is for debug info and should be removed in the future
func iteratorToNextInitializedTickSqrtPriceTarget(nextInitTickIter db.Iterator, _poolId uint64, swapstrat swapstrategy.SwapStrategy) (int64, osmomath.BigDec, osmomath.BigDec, error) {
	// Iterator must be valid to be able to retrieve the next tick from it below.
	if !nextInitTickIter.Valid() {
		return 0, osmomath.BigDec{}, osmomath.BigDec{}, types.RanOutOfTicksForPoolError{PoolId: _poolId}
	}
	// We first check to see what the position of the nearest initialized tick is
	// if zeroForOneStrategy, we look to the left of the tick the current sqrt price is at
	// if oneForZeroStrategy, we look to the right of the tick the current sqrt price is at
	// if no ticks are initialized (no users have created liquidity positions) then we return an error
	nextInitializedTick, err := types.TickIndexFromBytes(nextInitTickIter.Key())
	if err != nil {
		return 0, osmomath.BigDec{}, osmomath.BigDec{}, err
	}

	// Utilizing the next initialized tick, we find the corresponding nextInitializedTickSqrtPrice (the target sqrt price).
	nextInitializedTickSqrtPrice, err := math.TickToSqrtPrice(nextInitializedTick)
	if err != nil {
		return 0, osmomath.BigDec{}, osmomath.BigDec{}, fmt.Errorf("could not convert next tick (%v) to nextSqrtPrice", nextInitializedTick)
	}

	// If nextInitializedTickSqrtPrice exceeds the given price limit, we set the sqrtPriceTarget to the price limit.
	sqrtPriceTarget := swapstrat.GetSqrtTargetPrice(nextInitializedTickSqrtPrice)
	return nextInitializedTick, nextInitializedTickSqrtPrice, sqrtPriceTarget, nil
}

// computeOutAmtGivenIn calculates tokens to be swapped out given the provided amount and spread factor deducted. It also returns
// what the updated tick, liquidity, and currentSqrtPrice for the pool would be after this swap.
// Note this method is mutative, some of the tick and accumulator updates get written to store.
// However, there are no token transfers or pool updates done in this method. These mutations are performed in swapInAmtGivenOut.
// Note that passing in 0 for `priceLimit` will result in the price limit being set to the max/min value based on swap direction
func (k Keeper) computeOutAmtGivenIn(
	ctx sdk.Context,
	poolId uint64,
	tokenInMin sdk.Coin,
	tokenOutDenom string,
	spreadFactor osmomath.Dec,
	priceLimit osmomath.BigDec,
	updateAccumulators bool,
) (swapResult SwapResult, poolUpdates PoolUpdates, err error) {
	p, spreadRewardAccumulator, err := k.swapSetup(ctx, poolId, tokenInMin.Denom, tokenOutDenom, updateAccumulators)
	if err != nil {
		return SwapResult{}, PoolUpdates{}, err
	}
	var uptimeAccums []*accum.AccumulatorObject

	swapStrategy, sqrtPriceLimit, err := k.setupSwapStrategy(p, spreadFactor, tokenInMin.Denom, priceLimit)
	if err != nil {
		return SwapResult{}, PoolUpdates{}, err
	}

	var scalingFactor osmomath.Dec
	if updateAccumulators {
		// We only need the scaling factor if we are updating the accumulators
		scalingFactor, err = k.getSpreadFactorScalingFactorForPool(ctx, poolId)
		if err != nil {
			return SwapResult{}, PoolUpdates{}, err
		}
	}

	// initialize swap state with the following parameters:
	// as we iterate through the following for loop, this swap state will get updated after each required iteration
	swapState := newSwapState(tokenInMin.Amount, p, swapStrategy)

	nextInitTickIter := swapStrategy.InitializeNextTickIterator(ctx, poolId, swapState.tick)
	defer nextInitTickIter.Close()

	// Iterate and update swapState until we swap all tokenIn or we reach the specific sqrtPriceLimit
	swapNoProgressIterationCount := 0
	// TODO: for now, we check if amountSpecifiedRemaining is GT 0.0000001. This is because there are times when the remaining
	// amount may be extremely small, and that small amount cannot generate and amountIn/amountOut and we are therefore left
	// in an infinite loop.
	for swapState.amountSpecifiedRemaining.GT(smallestDec) && !swapState.sqrtPrice.Equal(sqrtPriceLimit) {
		// Log the sqrtPrice we start the iteration with
		sqrtPriceStart := swapState.sqrtPrice

		// get next initialized tick, and its implied sqrtPriceTarget
		nextInitializedTick, nextInitializedTickSqrtPrice, sqrtPriceTarget, err := iteratorToNextInitializedTickSqrtPriceTarget(nextInitTickIter, poolId, swapStrategy)
		if err != nil {
			return SwapResult{}, PoolUpdates{}, err
		}

		// Utilizing the bucket's liquidity and knowing the sqrt price target, we calculate how much tokenOut we get from the tokenIn
		// we also calculate the swap state's new computedSqrtPrice after this swap
		computedSqrtPrice, amountIn, amountOut, spreadRewardCharge := swapStrategy.ComputeSwapWithinBucketOutGivenIn(
			swapState.sqrtPrice,
			sqrtPriceTarget,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
		)

		if err := validateSwapProgressAndAmountConsumption(computedSqrtPrice, sqrtPriceStart, amountIn, amountOut); err != nil {
			return SwapResult{}, PoolUpdates{}, err
		}

		if updateAccumulators {
			// Calculate the spread reward growth for the swap step using the total spread factors charged, scaling the spread reward by the scaling factor.
			spreadFactorsAccruedPerUnitOfLiquidityScaled, err := swapState.updateSpreadRewardGrowthGlobal(spreadRewardCharge, scalingFactor)
			if err != nil {
				return SwapResult{}, PoolUpdates{}, err
			}

			// Emit telemetry to detect spread reward truncation.
			emitAccumulatorUpdateTelemetry(types.SpreadFactorTruncationTelemetryName, spreadFactorsAccruedPerUnitOfLiquidityScaled, spreadRewardCharge, poolId, swapState.liquidity, "is_out_given_in", "true")
		}

		ctx.Logger().Debug("cl calc out given in")
		emitSwapDebugLogs(ctx, swapState, computedSqrtPrice, amountIn, amountOut, spreadRewardCharge)

		// Update the swapState with the new sqrtPrice from the above swap
		swapState.sqrtPrice = computedSqrtPrice
		// We deduct the amount of tokens we input in ComputeSwapWithinBucketOutGivenIn(...) above from the user's defined tokenIn amount
		swapState.amountSpecifiedRemaining.SubMut(amountIn.Add(spreadRewardCharge))
		// We add the amount of tokens we received (amountOut) from the ComputeSwapWithinBucketOutGivenIn(...) above to the amountCalculated accumulator
		swapState.amountCalculated.AddMut(amountOut)

		// If ComputeSwapWithinBucketOutGivenIn(...) calculated a computedSqrtPrice that is equal to the nextInitializedTickSqrtPrice, this means all liquidity in the current
		// bucket has been consumed and we must move on to the next bucket to complete the swap
		if nextInitializedTickSqrtPrice.Equal(computedSqrtPrice) {
			swapState, err = k.swapCrossTickLogic(ctx, swapState, swapStrategy,
				nextInitializedTick, nextInitTickIter, p, spreadRewardAccumulator, &uptimeAccums, tokenInMin.Denom, updateAccumulators)
			if err != nil {
				return SwapResult{}, PoolUpdates{}, err
			}
		} else if edgeCaseInequalityBasedOnSwapStrategy(swapStrategy.ZeroForOne(), nextInitializedTickSqrtPrice, computedSqrtPrice) {
			// If, based on the swap strategy, the computedSqrtPrice matches the edge case inequality, we return an error.
			// This is an edge case that occurs when swapping at/near tick boundaries that will be fixed in the next release.
			// For now, we return an error and ask the user to try again with a different swap amount.
			return SwapResult{}, PoolUpdates{}, types.ComputedSqrtPriceInequalityError{IsZeroForOne: swapStrategy.ZeroForOne(), ComputedSqrtPrice: computedSqrtPrice, NextInitializedTickSqrtPrice: nextInitializedTickSqrtPrice}
		} else if !sqrtPriceStart.Equal(computedSqrtPrice) {
			// Otherwise if the sqrtPrice calculated from ComputeSwapWithinBucketOutGivenIn(...) does not equal the sqrtPriceStart we started with at the
			// beginning of this iteration, we set the swapState tick to the corresponding tick of the computedSqrtPrice calculated from ComputeSwapWithinBucketOutGivenIn(...)
			newTick, err := math.CalculateSqrtPriceToTick(computedSqrtPrice)
			if err != nil {
				return SwapResult{}, PoolUpdates{}, err
			}
			swapState.tick = newTick
		}

		// If nothing was consumed from swapState.amountSpecifiedRemaining, we increment the swapNoProgressIterationCount.
		// See definition of swapNoProgressLimit for more details.
		if amountIn.IsZero() {
			if swapNoProgressIterationCount >= swapNoProgressLimit {
				return SwapResult{}, PoolUpdates{}, types.SwapNoProgressError{PoolId: poolId, UserProvidedCoin: tokenInMin}
			}
			swapNoProgressIterationCount++
		}
	}

	// Note, this should be impossible to reach but we leave it as a defense-in-depth measure.
	if swapState.amountSpecifiedRemaining.IsNegative() {
		return SwapResult{}, PoolUpdates{}, types.OverChargeSwapOutGivenInError{AmountSpecifiedRemaining: swapState.amountSpecifiedRemaining}
	}

	// Add spread reward growth per share to the pool-global spread reward accumulator.
	if updateAccumulators {
		spreadRewardGrowth := sdk.DecCoin{Denom: tokenInMin.Denom, Amount: swapState.globalSpreadRewardGrowthPerUnitLiquidity}
		spreadRewardAccumulator.AddToAccumulator(sdk.NewDecCoins(spreadRewardGrowth))
	}

	// Coin amounts require int values
	// Round amountIn up to avoid under charging
	amountIn := (tokenInMin.Amount.ToLegacyDec().SubMut(swapState.amountSpecifiedRemaining)).Ceil().TruncateInt()
	// Round amountOut down to avoid over distributing.
	amountOut := swapState.amountCalculated.TruncateInt()

	ctx.Logger().Debug("final amount in", amountIn)
	ctx.Logger().Debug("final amount out", amountOut)

	return SwapResult{
		AmountIn:      amountIn,
		AmountOut:     amountOut,
		SpreadRewards: swapState.globalSpreadRewardGrowth,
	}, PoolUpdates{swapState.tick, swapState.liquidity, swapState.sqrtPrice}, nil
}

// computeInAmtGivenOut calculates tokens to be swapped in given the desired token out and spread factor deducted. It also returns
// what the updated tick, liquidity, and currentSqrtPrice for the pool would be after this swap.
// Note this method is mutative, some of the tick and accumulator updates get written to store.
// However, there are no token transfers or pool updates done in this method. These mutations are performed in swapOutAmtGivenIn.
func (k Keeper) computeInAmtGivenOut(
	ctx sdk.Context,
	desiredTokenOut sdk.Coin,
	tokenInDenom string,
	spreadFactor osmomath.Dec,
	priceLimit osmomath.BigDec,
	poolId uint64,
	updateAccumulators bool,
) (swapResult SwapResult, poolUpdates PoolUpdates, err error) {
	p, spreadRewardAccumulator, err := k.swapSetup(ctx, poolId, tokenInDenom, desiredTokenOut.Denom, updateAccumulators)
	if err != nil {
		return SwapResult{}, PoolUpdates{}, err
	}
	var uptimeAccums []*accum.AccumulatorObject

	swapStrategy, sqrtPriceLimit, err := k.setupSwapStrategy(p, spreadFactor, tokenInDenom, priceLimit)
	if err != nil {
		return SwapResult{}, PoolUpdates{}, err
	}

	var scalingFactor osmomath.Dec
	if updateAccumulators {
		// We only need the scaling factor if we are updating the accumulators
		scalingFactor, err = k.getSpreadFactorScalingFactorForPool(ctx, poolId)
		if err != nil {
			return SwapResult{}, PoolUpdates{}, err
		}
	}

	// initialize swap state with the following parameters:
	// as we iterate through the following for loop, this swap state will get updated after each required iteration
	swapState := newSwapState(desiredTokenOut.Amount, p, swapStrategy)

	nextInitTickIter := swapStrategy.InitializeNextTickIterator(ctx, poolId, swapState.tick)
	defer nextInitTickIter.Close()

	swapNoProgressIterationCount := 0
	// TODO: for now, we check if amountSpecifiedRemaining is GT 10^-18. This is because there are times when the remaining
	// amount may be extremely small, and that small amount cannot generate and amountIn/amountOut and we are therefore left
	// in an infinite loop.
	for swapState.amountSpecifiedRemaining.GT(smallestDec) && !swapState.sqrtPrice.Equal(sqrtPriceLimit) {
		// log the sqrtPrice we start the iteration with
		sqrtPriceStart := swapState.sqrtPrice

		nextInitializedTick, nextInitializedTickSqrtPrice, sqrtPriceTarget, err := iteratorToNextInitializedTickSqrtPriceTarget(nextInitTickIter, poolId, swapStrategy)
		if err != nil {
			return SwapResult{}, PoolUpdates{}, err
		}

		// Utilizing the bucket's liquidity and knowing the sqrt price target, we calculate the how much tokenOut we get from the tokenIn
		// we also calculate the swap state's new computedSqrtPrice after this swap
		computedSqrtPrice, amountOut, amountIn, spreadRewardChargeTotal := swapStrategy.ComputeSwapWithinBucketInGivenOut(
			swapState.sqrtPrice,
			sqrtPriceTarget,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
		)

		if err := validateSwapProgressAndAmountConsumption(computedSqrtPrice, sqrtPriceStart, amountIn, amountOut); err != nil {
			return SwapResult{}, PoolUpdates{}, err
		}

		if updateAccumulators {
			// Calculate the spread reward growth for the swap step using the total spread factors charged, scaling the spread reward by the scaling factor.
			spreadFactorsAccruedPerUnitOfLiquidityScaled, err := swapState.updateSpreadRewardGrowthGlobal(spreadRewardChargeTotal, scalingFactor)
			if err != nil {
				return SwapResult{}, PoolUpdates{}, err
			}

			// Emit telemetry to detect spread reward truncation.
			emitAccumulatorUpdateTelemetry(types.SpreadFactorTruncationTelemetryName, spreadFactorsAccruedPerUnitOfLiquidityScaled, spreadRewardChargeTotal, poolId, swapState.liquidity, "is_out_given_in", "false")
		}

		ctx.Logger().Debug("cl calc in given out")
		emitSwapDebugLogs(ctx, swapState, computedSqrtPrice, amountIn, amountOut, spreadRewardChargeTotal)

		// Update the swapState with the new sqrtPrice from the above swap
		swapState.sqrtPrice = computedSqrtPrice
		swapState.amountSpecifiedRemaining.SubMut(amountOut)
		swapState.amountCalculated.AddMut(amountIn.Add(spreadRewardChargeTotal))

		// If the ComputeSwapWithinBucketInGivenOut(...) calculated a computedSqrtPrice that is equal to the nextInitializedTickSqrtPrice, this means all liquidity in the current
		// bucket has been consumed and we must move on to the next bucket by crossing a tick to complete the swap
		if nextInitializedTickSqrtPrice.Equal(computedSqrtPrice) {
			swapState, err = k.swapCrossTickLogic(ctx, swapState, swapStrategy,
				nextInitializedTick, nextInitTickIter, p, spreadRewardAccumulator, &uptimeAccums, tokenInDenom, updateAccumulators)
			if err != nil {
				return SwapResult{}, PoolUpdates{}, err
			}
		} else if edgeCaseInequalityBasedOnSwapStrategy(swapStrategy.ZeroForOne(), nextInitializedTickSqrtPrice, computedSqrtPrice) {
			// If, based on the swap strategy, the computedSqrtPrice matches the edge case inequality, we return an error.
			// This is an edge case that occurs when swapping at/near tick boundaries that will be fixed in the next release.
			// For now, we return an error and ask the user to try again with a different swap amount.
			return SwapResult{}, PoolUpdates{}, types.ComputedSqrtPriceInequalityError{IsZeroForOne: swapStrategy.ZeroForOne(), ComputedSqrtPrice: computedSqrtPrice, NextInitializedTickSqrtPrice: nextInitializedTickSqrtPrice}
		} else if !sqrtPriceStart.Equal(computedSqrtPrice) {
			// Otherwise, if the computedSqrtPrice calculated from ComputeSwapWithinBucketInGivenOut(...) does not equal the sqrtPriceStart we started with at the
			// beginning of this iteration, we set the swapState tick to the corresponding tick of the computedSqrtPrice calculated from ComputeSwapWithinBucketInGivenOut(...)
			swapState.tick, err = math.CalculateSqrtPriceToTick(computedSqrtPrice)
			if err != nil {
				return SwapResult{}, PoolUpdates{}, err
			}
		}

		// If nothing was consumed from swapState.amountSpecifiedRemaining, we increment the swapNoProgressIterationCount.
		// See definition of swapNoProgressLimit for more details.
		if amountOut.IsZero() {
			if swapNoProgressIterationCount >= swapNoProgressLimit {
				return SwapResult{}, PoolUpdates{}, types.SwapNoProgressError{PoolId: poolId, UserProvidedCoin: desiredTokenOut}
			}
			swapNoProgressIterationCount++
		}
	}

	// Note, this should be impossible to reach but we leave it as a defense-in-depth measure.
	if swapState.amountSpecifiedRemaining.IsNegative() {
		return SwapResult{}, PoolUpdates{}, fmt.Errorf("over charged problem swap in given out by %s", swapState.amountSpecifiedRemaining)
	}

	// Add spread reward growth per share to the pool-global spread reward accumulator.
	if updateAccumulators {
		spreadRewardAccumulator.AddToAccumulator(sdk.NewDecCoins(sdk.NewDecCoinFromDec(tokenInDenom, swapState.globalSpreadRewardGrowthPerUnitLiquidity)))
	}

	// coin amounts require int values
	// Round amount in up to avoid under charging the user.
	amountIn := swapState.amountCalculated.Ceil().TruncateInt()

	// Round amount out down to avoid over charging the pool.
	amountOut := desiredTokenOut.Amount.ToLegacyDec().SubMut(swapState.amountSpecifiedRemaining).TruncateInt()

	ctx.Logger().Debug("final amount in", amountIn)
	ctx.Logger().Debug("final amount out", amountOut)

	return SwapResult{
		AmountIn:      amountIn,
		AmountOut:     amountOut,
		SpreadRewards: swapState.globalSpreadRewardGrowth,
	}, PoolUpdates{swapState.tick, swapState.liquidity, swapState.sqrtPrice}, nil
}

func emitSwapDebugLogs(ctx sdk.Context, swapState SwapState, reachedPrice osmomath.BigDec, amountIn, amountOut, spreadCharge osmomath.Dec) {
	ctx.Logger().Debug("start sqrt price", swapState.sqrtPrice)
	ctx.Logger().Debug("reached sqrt price", reachedPrice)
	ctx.Logger().Debug("liquidity", swapState.liquidity)
	ctx.Logger().Debug("amountIn", amountIn)
	ctx.Logger().Debug("amountOut", amountOut)
	ctx.Logger().Debug("spreadRewardChargeTotal", spreadCharge)
}

// logic for crossing a tick during a swap
//
// if uptimeAccums is nil, fetches uptime accumulators.
// uptime accumulators are always mutated for the tick crossing.
func (k Keeper) swapCrossTickLogic(ctx sdk.Context,
	swapState SwapState, strategy swapstrategy.SwapStrategy,
	nextInitializedTick int64, nextTickIter db.Iterator,
	p types.ConcentratedPoolExtension,
	spreadRewardAccum *accum.AccumulatorObject, uptimeAccums *[]*accum.AccumulatorObject,
	tokenInDenom string, updateAccumulators bool) (SwapState, error) {
	nextInitializedTickInfo, err := ParseTickFromBz(nextTickIter.Value())
	if err != nil {
		return swapState, err
	}
	if updateAccumulators {
		if *uptimeAccums == nil {
			uptimeAccumsRaw, err := k.GetUptimeAccumulators(ctx, p.GetId())
			if err != nil {
				return swapState, err
			}
			uptimeAccums = &uptimeAccumsRaw
		}

		if err := k.updateGivenPoolUptimeAccumulatorsToNow(ctx, p, *uptimeAccums); err != nil {
			return swapState, err
		}

		// Retrieve the liquidity held in the next closest initialized tick
		spreadRewardGrowth := sdk.DecCoin{Denom: tokenInDenom, Amount: swapState.globalSpreadRewardGrowthPerUnitLiquidity}
		err := k.crossTick(ctx, p.GetId(), nextInitializedTick, &nextInitializedTickInfo, spreadRewardGrowth, spreadRewardAccum.GetValue(), *uptimeAccums)
		if err != nil {
			return swapState, err
		}
	}
	liquidityNet := nextInitializedTickInfo.LiquidityNet

	// Move next tick iterator to the next tick as the tick is crossed.
	nextTickIter.Next()

	liquidityNet = swapState.swapStrategy.SetLiquidityDeltaSign(liquidityNet)
	// Update the swapState's liquidity with the new tick's liquidity
	swapState.liquidity.AddMut(liquidityNet)

	// Update the swapState's tick with the tick we retrieved liquidity from
	swapState.tick = strategy.UpdateTickAfterCrossing(nextInitializedTick)

	return swapState, nil
}

// updatePoolForSwap updates the given pool object with the results of a swap operation.
//
// The method consumes a fixed amount of gas per swap to prevent spam. It applies the swap operation to the given
// pool object by calling its ApplySwap method. It then sets the updated pool object using the setPool method
// of the keeper. Finally, it transfers the input and output tokens to and from the sender and the pool account
// using the SendCoins method of the bank keeper.
//
// Calls AfterConcentratedPoolSwap listener. Currently, it notifies twap module about
// a spot price update.
//
// If any error occurs during the swap operation, the method returns an error value indicating the cause of the error.
func (k Keeper) updatePoolForSwap(
	ctx sdk.Context,
	pool types.ConcentratedPoolExtension,
	swapDetails SwapDetails,
	poolUpdates PoolUpdates,
	totalSpreadFactors osmomath.Dec,
) error {
	// Fixed gas consumption per swap to prevent spam
	poolId := pool.GetId()
	ctx.GasMeter().ConsumeGas(types.ConcentratedGasFeeForSwap, "cl pool swap computation")
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	// Spread factors should already be rounded up to a whole number dec, but we do this as a precaution
	spreadFactorsRoundedUp := sdk.NewCoin(swapDetails.TokenIn.Denom, totalSpreadFactors.Ceil().TruncateInt())

	// Remove the spread factors from the input token
	swapDetails.TokenIn.Amount = swapDetails.TokenIn.Amount.Sub(spreadFactorsRoundedUp.Amount)

	// Send the input token from the user to the pool's primary address
	err = k.bankKeeper.SendCoins(ctx, swapDetails.Sender, pool.GetAddress(), sdk.Coins{
		swapDetails.TokenIn,
	})
	if err != nil {
		return types.InsufficientUserBalanceError{Err: err}
	}

	// Send the spread factors taken from the input token from the user to the pool's spread factor account
	if !spreadFactorsRoundedUp.IsZero() {
		err = k.bankKeeper.SendCoins(ctx, swapDetails.Sender, pool.GetSpreadRewardsAddress(), sdk.Coins{
			spreadFactorsRoundedUp,
		})
		if err != nil {
			return types.InsufficientUserBalanceError{Err: err}
		}
	}

	// Send the output token to the sender from the pool
	err = k.bankKeeper.SendCoins(ctx, pool.GetAddress(), swapDetails.Sender, sdk.Coins{
		swapDetails.TokenOut,
	})
	if err != nil {
		return types.InsufficientPoolBalanceError{Err: err}
	}

	err = pool.ApplySwap(poolUpdates.NewLiquidity, poolUpdates.NewCurrentTick, poolUpdates.NewSqrtPrice)
	if err != nil {
		return fmt.Errorf("error applying swap: %w", err)
	}

	if err := k.setPool(ctx, pool); err != nil {
		return err
	}

	k.listeners.AfterConcentratedPoolSwap(ctx, swapDetails.Sender, poolId, sdk.Coins{swapDetails.TokenIn}, sdk.Coins{swapDetails.TokenOut})

	// Emit swap event. Note that we emit these at the layer of each pool module rather than the poolmanager module
	// since poolmanager has many swap wrapper APIs that we would need to consider.
	// Search for references to this function to see where else it is used.
	// Each new pool module will have to emit this event separately
	events.EmitSwapEvent(ctx, swapDetails.Sender, pool.GetId(), sdk.Coins{swapDetails.TokenIn}, sdk.Coins{swapDetails.TokenOut})

	return err
}

func getZeroForOne(inDenom, asset0 string) bool {
	return inDenom == asset0
}

func checkDenomValidity(inDenom, outDenom, asset0, asset1 string) error {
	// check that the specified tokenOut matches one of the assets in the specified pool
	if outDenom != asset0 && outDenom != asset1 {
		return types.TokenOutDenomNotInPoolError{TokenOutDenom: outDenom}
	}
	// check that the specified tokenIn matches one of the assets in the specified pool
	if inDenom != asset0 && inDenom != asset1 {
		return types.TokenInDenomNotInPoolError{TokenInDenom: inDenom}
	}
	// check that token in and token out are different denominations
	if outDenom == inDenom {
		return types.DenomDuplicatedError{TokenInDenom: inDenom, TokenOutDenom: outDenom}
	}
	return nil
}

func (k Keeper) setupSwapStrategy(p types.ConcentratedPoolExtension, spreadFactor osmomath.Dec, tokenInDenom string, priceLimit osmomath.BigDec) (strategy swapstrategy.SwapStrategy, sqrtPriceLimit osmomath.BigDec, err error) {
	zeroForOne := getZeroForOne(tokenInDenom, p.GetToken0())

	// take provided price limit and turn this into a sqrt price limit since formulas use sqrtPrice
	sqrtPriceLimit, err = swapstrategy.GetSqrtPriceLimit(priceLimit, zeroForOne)
	if err != nil {
		return strategy, osmomath.BigDec{}, types.SqrtRootCalculationError{SqrtPriceLimit: sqrtPriceLimit}
	}

	// set the swap strategy
	swapStrategy := swapstrategy.New(zeroForOne, sqrtPriceLimit, k.storeKey, spreadFactor)

	// get current sqrt price from pool
	curSqrtPrice := p.GetCurrentSqrtPrice()
	if err := swapStrategy.ValidateSqrtPrice(sqrtPriceLimit, curSqrtPrice); err != nil {
		return strategy, osmomath.BigDec{}, err
	}

	return swapStrategy, sqrtPriceLimit, nil
}

func (k Keeper) getPoolForSwap(ctx sdk.Context, poolId uint64) (types.ConcentratedPoolExtension, error) {
	p, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return p, err
	}
	hasPositionInPool := k.PoolHasPosition(ctx, p)
	if !hasPositionInPool {
		return p, types.NoSpotPriceWhenNoLiquidityError{PoolId: poolId}
	}
	return p, nil
}

// validateSwapProgressAndAmountConsumption validates that the swap progress and amount consumption are valid. These are valid if:
// - computedSqrtPrice is not equal to sqrtPriceStart (progress made)
// - computedSqrtPrice is equals to sqrtPriceStart and both amountIn and amountOut are zero (progress not made AND amounts are not consumed)
// If swap succeeded within the same ULP while consuming amounts, that would mean that a trader can exploit the pool to
// continue getting the same execution price. To prevent that, we add this check that fails the swap if amounts
// in or out are non-zero while swap does not move even by 1 ULP.
// If no progress is made with zero amounts, there is a risk of running into an infinite loop which we prevent
// by a constant bound on the the number of iterations without progress.
// Note that having a single iteration without progress and zero amounts is correct. One good example
// is swapping one for zero (right) directly to the tick. Then, immediately swapping zero for one (left).
// After the first swap, our sqrtPriceCurrent is the crossed tick's sqrt price. Also sqrtPriceTarget is the crossed tick's sqrt price.
// In such a case, no amounts are consumed and the swap step is allowed to succeed.
func validateSwapProgressAndAmountConsumption(computedSqrtPrice, sqrtPriceStart osmomath.BigDec, amountIn, amountOut osmomath.Dec) error {
	if computedSqrtPrice.Equal(sqrtPriceStart) && !(amountIn.IsZero() && amountOut.IsZero()) {
		return types.SwapNoProgressWithConsumptionError{ComputedSqrtPrice: computedSqrtPrice, AmountIn: amountIn, AmountOut: amountOut}
	}
	return nil
}

func edgeCaseInequalityBasedOnSwapStrategy(isZeroForOne bool, nextInitializedTickSqrtPrice, computedSqrtPrice osmomath.BigDec) bool {
	if isZeroForOne {
		return nextInitializedTickSqrtPrice.GT(computedSqrtPrice)
	}
	return nextInitializedTickSqrtPrice.LT(computedSqrtPrice)
}

// ComputeMaxInAmtGivenMaxTicksCrossed calculates the maximum amount of the tokenInDenom that can be swapped
// into the pool to swap through all the liquidity from the current tick through the maxTicksCrossed tick,
// but not exceed it.
func (k Keeper) ComputeMaxInAmtGivenMaxTicksCrossed(
	ctx sdk.Context,
	poolId uint64,
	tokenInDenom string,
	maxTicksCrossed uint64,
) (maxTokenIn, resultingTokenOut sdk.Coin, err error) {
	cacheCtx, _ := ctx.CacheContext()

	p, err := k.getPoolForSwap(cacheCtx, poolId)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// Validate tokenInDenom exists in the pool
	if tokenInDenom != p.GetToken0() && tokenInDenom != p.GetToken1() {
		return sdk.Coin{}, sdk.Coin{}, types.TokenInDenomNotInPoolError{TokenInDenom: tokenInDenom}
	}

	// Determine the tokenOutDenom based on the tokenInDenom
	var tokenOutDenom string
	if tokenInDenom == p.GetToken0() {
		tokenOutDenom = p.GetToken1()
	} else {
		tokenOutDenom = p.GetToken0()
	}

	// Setup the swap strategy
	swapStrategy, _, err := k.setupSwapStrategy(p, p.GetSpreadFactor(cacheCtx), tokenInDenom, osmomath.ZeroBigDec())
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// Initialize swap state
	// Utilize the total amount of tokenOutDenom in the pool as the specified amountOut, since we want
	// the limitation to be the tick crossing, not the amountOut.
	balance := k.bankKeeper.GetBalance(ctx, p.GetAddress(), tokenOutDenom)
	swapState := newSwapState(balance.Amount, p, swapStrategy)

	nextInitTickIter := swapStrategy.InitializeNextTickIterator(cacheCtx, poolId, swapState.tick)
	defer nextInitTickIter.Close()

	totalTokenOut := osmomath.ZeroDec()

	for i := uint64(0); i < maxTicksCrossed; i++ {
		// Check if the iterator is valid
		if !nextInitTickIter.Valid() {
			break
		}

		nextInitializedTick, nextInitializedTickSqrtPrice, sqrtPriceTarget, err := iteratorToNextInitializedTickSqrtPriceTarget(nextInitTickIter, poolId, swapStrategy)
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, err
		}

		// Compute the swap
		computedSqrtPrice, amountOut, amountIn, spreadRewardChargeTotal := swapStrategy.ComputeSwapWithinBucketInGivenOut(
			swapState.sqrtPrice,
			sqrtPriceTarget,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
		)

		swapState.sqrtPrice = computedSqrtPrice
		swapState.amountSpecifiedRemaining.SubMut(amountOut)
		swapState.amountCalculated.AddMut(amountIn.Add(spreadRewardChargeTotal))

		totalTokenOut = totalTokenOut.Add(amountOut)

		// Check if the tick needs to be updated
		// We do not need to track spread rewards or uptime accums here since we are not actually swapping.
		if nextInitializedTickSqrtPrice.Equal(computedSqrtPrice) {
			nextInitializedTickInfo, err := ParseTickFromBz(nextInitTickIter.Value())
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, err
			}
			liquidityNet := nextInitializedTickInfo.LiquidityNet

			nextInitTickIter.Next()

			liquidityNet = swapState.swapStrategy.SetLiquidityDeltaSign(liquidityNet)
			swapState.liquidity.AddMut(liquidityNet)

			swapState.tick = swapStrategy.UpdateTickAfterCrossing(nextInitializedTick)
		} else if edgeCaseInequalityBasedOnSwapStrategy(swapStrategy.ZeroForOne(), nextInitializedTickSqrtPrice, computedSqrtPrice) {
			return sdk.Coin{}, sdk.Coin{}, types.ComputedSqrtPriceInequalityError{
				IsZeroForOne:                 swapStrategy.ZeroForOne(),
				ComputedSqrtPrice:            computedSqrtPrice,
				NextInitializedTickSqrtPrice: nextInitializedTickSqrtPrice,
			}
		} else if !swapState.sqrtPrice.Equal(computedSqrtPrice) {
			newTick, err := math.CalculateSqrtPriceToTick(computedSqrtPrice)
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, err
			}
			swapState.tick = newTick
		}

		// Break the loop early if nothing was consumed from swapState.amountSpecifiedRemaining
		if amountOut.IsZero() {
			break
		}
	}

	maxAmt := swapState.amountCalculated.Ceil().TruncateInt()
	maxTokenIn = sdk.NewCoin(tokenInDenom, maxAmt)
	resultingTokenOut = sdk.NewCoin(tokenOutDenom, totalTokenOut.TruncateInt())

	return maxTokenIn, resultingTokenOut, nil
}
