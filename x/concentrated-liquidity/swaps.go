package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	events "github.com/osmosis-labs/osmosis/v16/x/poolmanager/events"

	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
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
	amountSpecifiedRemaining sdk.Dec

	// Amount of the other token that is calculated from the specified token.
	// if out given in, amount of token swapped out.
	// if in given out, amount of token swapped in.
	// Initialized to zero.
	// Updated after every swap step.
	amountCalculated sdk.Dec

	// Current sqrt price while calculating swap.
	// Initialized to the pool's current sqrt price.
	// Updated after every swap step.
	sqrtPrice sdk.Dec

	// Current tick while calculating swap.
	// Initialized to the pool's current tick.
	// Updated each time a tick is crossed.
	tick int64

	// Current liqudiity within the active tick.
	// Initialized to the pool's current tick's liquidity.
	// Updated each time a tick is crossed.
	liquidity sdk.Dec

	// Global spread reward growth per-current swap.
	// Initialized to zero.
	// Updated after every swap step.
	spreadRewardGrowthGlobal sdk.Dec
}

func newSwapState(specifiedAmount sdk.Int, p types.ConcentratedPoolExtension, strategy swapstrategy.SwapStrategy) SwapState {
	return SwapState{
		amountSpecifiedRemaining: specifiedAmount.ToDec(),
		amountCalculated:         sdk.ZeroDec(),
		sqrtPrice:                p.GetCurrentSqrtPrice(),
		tick:                     strategy.InitializeTickValue(p.GetCurrentTick()),
		liquidity:                p.GetLiquidity(),
		spreadRewardGrowthGlobal: sdk.ZeroDec(),
	}
}

var (
	smallestDec = sdk.SmallestDec()
)

// updateSpreadRewardGrowthGlobal updates the swap state's spread reward growth global per unit of liquidity
// when liquidity is positive.
//
// If the liquidity is zero, this is a no-op. This case may occur when there is no liquidity
// between the ticks. This is possible when there are only 2 positions with no overlapping ranges.
// As a result, the range from the end of position one to the beginning of position
// two has no liquidity and can be skipped.
func (ss *SwapState) updateSpreadRewardGrowthGlobal(spreadRewardChargeTotal sdk.Dec) {
	if !ss.liquidity.IsZero() {
		// We round down here since we want to avoid overdistributing (the "spread factor charge" refers to
		// the total spread factors that will be accrued to the spread factor accumulator)
		spreadFactorssAccruedPerUnitOfLiquidity := spreadRewardChargeTotal.QuoTruncate(ss.liquidity)
		ss.spreadRewardGrowthGlobal.AddMut(spreadFactorssAccruedPerUnitOfLiquidity)
		return
	}
}

func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolI poolmanagertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	spreadFactor sdk.Dec,
) (tokenOutAmount sdk.Int, err error) {
	if tokenIn.Denom == tokenOutDenom {
		return sdk.Int{}, types.DenomDuplicatedError{TokenInDenom: tokenIn.Denom, TokenOutDenom: tokenOutDenom}
	}

	// Convert pool interface to CL pool type
	pool, err := asConcentrated(poolI)
	if err != nil {
		return sdk.Int{}, err
	}

	// Determine if we are swapping asset0 for asset1 or vice versa
	zeroForOne := getZeroForOne(tokenIn.Denom, pool.GetToken0())

	// Change priceLimit based on which direction we are swapping
	priceLimit := swapstrategy.GetPriceLimit(zeroForOne)
	tokenIn, tokenOut, _, _, _, err := k.swapOutAmtGivenIn(ctx, sender, pool, tokenIn, tokenOutDenom, spreadFactor, priceLimit)
	if err != nil {
		return sdk.Int{}, err
	}
	tokenOutAmount = tokenOut.Amount

	// price impact protection.
	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, types.AmountLessThanMinError{TokenAmount: tokenOutAmount, TokenMin: tokenOutMinAmount}
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
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	spreadFactor sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	if tokenOut.Denom == tokenInDenom {
		return sdk.Int{}, types.DenomDuplicatedError{TokenInDenom: tokenInDenom, TokenOutDenom: tokenOut.Denom}
	}

	pool, err := asConcentrated(poolI)
	if err != nil {
		return sdk.Int{}, err
	}

	zeroForOne := getZeroForOne(tokenInDenom, pool.GetToken0())

	// change priceLimit based on which direction we are swapping
	// if zeroForOne == true, use MinSpotPrice else use MaxSpotPrice
	priceLimit := swapstrategy.GetPriceLimit(zeroForOne)
	tokenIn, tokenOut, _, _, _, err := k.swapInAmtGivenOut(ctx, sender, pool, tokenOut, tokenInDenom, spreadFactor, priceLimit)
	if err != nil {
		return sdk.Int{}, err
	}
	tokenInAmount = tokenIn.Amount

	// price impact protection.
	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, types.AmountGreaterThanMaxError{TokenAmount: tokenInAmount, TokenMax: tokenInMaxAmount}
	}

	return tokenInAmount, nil
}

// swapOutAmtGivenIn is the internal mutative method for CalcOutAmtGivenIn. Utilizing CalcOutAmtGivenIn's output, this function applies the
// new tick, liquidity, and sqrtPrice to the respective pool
func (k Keeper) swapOutAmtGivenIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.ConcentratedPoolExtension,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	spreadFactor sdk.Dec,
	priceLimit sdk.Dec,
) (calcTokenIn, calcTokenOut sdk.Coin, currentTick int64, liquidity, sqrtPrice sdk.Dec, err error) {
	tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, totalSpreadFactors, err := k.computeOutAmtGivenIn(ctx, pool.GetId(), tokenIn, tokenOutDenom, spreadFactor, priceLimit)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, err
	}

	if !tokenOut.Amount.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, types.InvalidAmountCalculatedError{Amount: tokenOut.Amount}
	}

	// Settles balances between the tx sender and the pool to match the swap that was executed earlier.
	// Also emits swap event and updates related liquidity metrics
	if err := k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, totalSpreadFactors); err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, err
	}

	return tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, nil
}

// swapInAmtGivenOut is the internal mutative method for calcInAmtGivenOut. Utilizing calcInAmtGivenOut's output, this function applies the
// new tick, liquidity, and sqrtPrice to the respective pool.
func (k *Keeper) swapInAmtGivenOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.ConcentratedPoolExtension,
	desiredTokenOut sdk.Coin,
	tokenInDenom string,
	spreadFactor sdk.Dec,
	priceLimit sdk.Dec,
) (calcTokenIn, calcTokenOut sdk.Coin, currentTick int64, liquidity, sqrtPrice sdk.Dec, err error) {
	tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, totalSpreadFactors, err := k.computeInAmtGivenOut(ctx, desiredTokenOut, tokenInDenom, spreadFactor, priceLimit, pool.GetId())
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, err
	}

	// check that the tokenOut calculated is both valid and less than specified limit
	if !tokenIn.Amount.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, types.InvalidAmountCalculatedError{Amount: tokenIn.Amount}
	}

	// Settles balances between the tx sender and the pool to match the swap that was executed earlier.
	// Also emits swap event and updates related liquidity metrics
	if err := k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, totalSpreadFactors); err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, err
	}

	return tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, nil
}

func (k Keeper) CalcOutAmtGivenIn(
	ctx sdk.Context,
	poolI poolmanagertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	spreadFactor sdk.Dec,
) (tokenOut sdk.Coin, err error) {
	cacheCtx, _ := ctx.CacheContext()
	_, tokenOut, _, _, _, _, err = k.computeOutAmtGivenIn(cacheCtx, poolI.GetId(), tokenIn, tokenOutDenom, spreadFactor, sdk.ZeroDec())
	if err != nil {
		return sdk.Coin{}, err
	}
	return tokenOut, nil
}

func (k Keeper) CalcInAmtGivenOut(
	ctx sdk.Context,
	poolI poolmanagertypes.PoolI,
	tokenOut sdk.Coin,
	tokenInDenom string,
	spreadFactor sdk.Dec,
) (tokenIn sdk.Coin, err error) {
	cacheCtx, _ := ctx.CacheContext()
	tokenIn, _, _, _, _, _, err = k.computeInAmtGivenOut(cacheCtx, tokenOut, tokenInDenom, spreadFactor, sdk.ZeroDec(), poolI.GetId())
	if err != nil {
		return sdk.Coin{}, err
	}
	return tokenIn, nil
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
	spreadFactor sdk.Dec,
	priceLimit sdk.Dec,
) (tokenIn, tokenOut sdk.Coin, updatedTick int64, updatedLiquidity, updatedSqrtPrice sdk.Dec, totalSpreadFactors sdk.Dec, err error) {
	// Get pool and asset info
	p, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
	}

	if err := checkDenomValidity(tokenInMin.Denom, tokenOutDenom, p.GetToken0(), p.GetToken1()); err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
	}

	swapStrategy, sqrtPriceLimit, err := k.setupSwapStrategy(ctx, p, spreadFactor, tokenInMin.Denom, priceLimit)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
	}

	spreadRewardAccumulator, uptimeAccums, err := k.getSwapAccumulators(ctx, poolId)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
	}

	// initialize swap state with the following parameters:
	// as we iterate through the following for loop, this swap state will get updated after each required iteration
	swapState := newSwapState(tokenInMin.Amount, p, swapStrategy)

	nextTickIter := swapStrategy.InitializeNextTickIterator(ctx, poolId, swapState.tick)
	defer nextTickIter.Close()

	totalSpreadFactors = sdk.ZeroDec()
	// Iterate and update swapState until we swap all tokenIn or we reach the specific sqrtPriceLimit
	// TODO: for now, we check if amountSpecifiedRemaining is GT 0.0000001. This is because there are times when the remaining
	// amount may be extremely small, and that small amount cannot generate and amountIn/amountOut and we are therefore left
	// in an infinite loop.
	for swapState.amountSpecifiedRemaining.GT(smallestDec) && !swapState.sqrtPrice.Equal(sqrtPriceLimit) {
		// Log the sqrtPrice we start the iteration with
		sqrtPriceStart := swapState.sqrtPrice

		// Iterator must be valid to be able to retrieve the next tick from it below.
		if !nextTickIter.Valid() {
			return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, types.RanOutOfTicksForPoolError{PoolId: poolId}
		}

		// We first check to see what the position of the nearest initialized tick is
		// if zeroForOneStrategy, we look to the left of the tick the current sqrt price is at
		// if oneForZeroStrategy, we look to the right of the tick the current sqrt price is at
		// if no ticks are initialized (no users have created liquidity positions) then we return an error
		nextTick, err := types.TickIndexFromBytes(nextTickIter.Key())
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
		}

		// Utilizing the next initialized tick, we find the corresponding nextPrice (the target price).
		_, nextTickSqrtPrice, err := math.TickToSqrtPrice(nextTick)
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("could not convert next tick (%v) to nextSqrtPrice", nextTick)
		}

		// If nextSqrtPrice exceeds the price limit, we set the nextSqrtPrice to the price limit.
		sqrtPriceTarget := swapStrategy.GetSqrtTargetPrice(nextTickSqrtPrice)

		// Utilizing the bucket's liquidity and knowing the price target, we calculate the how much tokenOut we get from the tokenIn
		// we also calculate the swap state's new sqrtPrice after this swap
		sqrtPrice, amountIn, amountOut, spreadRewardCharge := swapStrategy.ComputeSwapStepOutGivenIn(
			swapState.sqrtPrice,
			sqrtPriceTarget,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
		)

		// Update the spread reward growth for the entire swap using the total spread factors charged.
		swapState.updateSpreadRewardGrowthGlobal(spreadRewardCharge)

		ctx.Logger().Debug("cl calc out given in")
		ctx.Logger().Debug("start sqrt price", swapState.sqrtPrice)
		ctx.Logger().Debug("reached sqrt price", sqrtPrice)
		ctx.Logger().Debug("liquidity", swapState.liquidity)
		ctx.Logger().Debug("amountIn", amountIn)
		ctx.Logger().Debug("amountOut", amountOut)
		ctx.Logger().Debug("spreadRewardCharge", spreadRewardCharge)

		// Update the swapState with the new sqrtPrice from the above swap
		swapState.sqrtPrice = sqrtPrice
		// We deduct the amount of tokens we input in the computeSwapStep above from the user's defined tokenIn amount
		swapState.amountSpecifiedRemaining.SubMut(amountIn.Add(spreadRewardCharge))
		// We add the amount of tokens we received (amountOut) from the computeSwapStep above to the amountCalculated accumulator
		swapState.amountCalculated.AddMut(amountOut)
		totalSpreadFactors = totalSpreadFactors.Add(spreadRewardCharge)

		// If the computeSwapStep calculated a sqrtPrice that is equal to the nextSqrtPrice, this means all liquidity in the current
		// tick has been consumed and we must move on to the next tick to complete the swap
		if nextTickSqrtPrice.Equal(sqrtPrice) {
			swapState, err = k.swapCrossTickLogic(ctx, swapState, swapStrategy,
				nextTick, nextTickIter, p, spreadRewardAccumulator, uptimeAccums, tokenInMin.Denom)
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
			}
		} else if !sqrtPriceStart.Equal(sqrtPrice) {
			// Otherwise if the sqrtPrice calculated from computeSwapStep does not equal the sqrtPrice we started with at the
			// beginning of this iteration, we set the swapState tick to the corresponding tick of the sqrtPrice calculated from computeSwapStep
			price := sqrtPrice.Mul(sqrtPrice)
			swapState.tick, err = math.PriceToTickRoundDown(price, p.GetTickSpacing())
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
			}
		}
	}

	// Add spread reward growth per share to the pool-global spread reward accumulator.
	spreadRewardAccumulator.AddToAccumulator(sdk.NewDecCoins(sdk.NewDecCoinFromDec(tokenInMin.Denom, swapState.spreadRewardGrowthGlobal)))

	// Coin amounts require int values
	// Round amountIn up to avoid under charging
	amt0 := (tokenInMin.Amount.ToDec().Sub(swapState.amountSpecifiedRemaining)).Ceil().TruncateInt()
	// Round amountOut down to avoid over distributing.
	amt1 := swapState.amountCalculated.TruncateInt()

	ctx.Logger().Debug("final amount in", amt0)
	ctx.Logger().Debug("final amount out", amt1)

	tokenIn = sdk.NewCoin(tokenInMin.Denom, amt0)
	tokenOut = sdk.NewCoin(tokenOutDenom, amt1)

	return tokenIn, tokenOut, swapState.tick, swapState.liquidity, swapState.sqrtPrice, totalSpreadFactors, nil
}

// computeInAmtGivenOut calculates tokens to be swapped in given the desired token out and spread factor deducted. It also returns
// what the updated tick, liquidity, and currentSqrtPrice for the pool would be after this swap.
// Note this method is mutative, some of the tick and accumulator updates get written to store.
// However, there are no token transfers or pool updates done in this method. These mutations are performed in swapOutAmtGivenIn.
func (k Keeper) computeInAmtGivenOut(
	ctx sdk.Context,
	desiredTokenOut sdk.Coin,
	tokenInDenom string,
	spreadFactor sdk.Dec,
	priceLimit sdk.Dec,
	poolId uint64,
) (tokenIn, tokenOut sdk.Coin, updatedTick int64, updatedLiquidity, updatedSqrtPrice sdk.Dec, totalSpreadFactors sdk.Dec, err error) {
	p, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
	}

	if err := checkDenomValidity(tokenInDenom, desiredTokenOut.Denom, p.GetToken0(), p.GetToken1()); err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
	}

	swapStrategy, sqrtPriceLimit, err := k.setupSwapStrategy(ctx, p, spreadFactor, tokenInDenom, priceLimit)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
	}

	// initialize swap state with the following parameters:
	// as we iterate through the following for loop, this swap state will get updated after each required iteration
	swapState := newSwapState(desiredTokenOut.Amount, p, swapStrategy)

	nextTickIter := swapStrategy.InitializeNextTickIterator(ctx, poolId, swapState.tick)
	defer nextTickIter.Close()

	spreadRewardAccumulator, uptimeAccums, err := k.getSwapAccumulators(ctx, poolId)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
	}

	totalSpreadFactors = sdk.ZeroDec()

	// TODO: This should be GT 0 but some instances have very small remainder
	// need to look into fixing this
	for swapState.amountSpecifiedRemaining.GT(smallestDec) && !swapState.sqrtPrice.Equal(sqrtPriceLimit) {
		// log the sqrtPrice we start the iteration with
		sqrtPriceStart := swapState.sqrtPrice

		// Iterator must be valid to be able to retrieve the next tick from it below.
		if !nextTickIter.Valid() {
			return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, types.RanOutOfTicksForPoolError{PoolId: poolId}
		}

		// we first check to see what the position of the nearest initialized tick is
		// if zeroForOne is false, we look to the left of the tick the current sqrt price is at
		// if zeroForOne is true, we look to the right of the tick the current sqrt price is at
		// if no ticks are initialized (no users have created liquidity positions) then we return an error
		nextTick, err := types.TickIndexFromBytes(nextTickIter.Key())
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
		}

		// utilizing the next initialized tick, we find the corresponding nextPrice (the target price)
		_, sqrtPriceNextTick, err := math.TickToSqrtPrice(nextTick)
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, types.TickToSqrtPriceConversionError{NextTick: nextTick}
		}

		sqrtPriceTarget := swapStrategy.GetSqrtTargetPrice(sqrtPriceNextTick)

		// utilizing the bucket's liquidity and knowing the price target, we calculate the how much tokenOut we get from the tokenIn
		// we also calculate the swap state's new sqrtPrice after this swap
		sqrtPrice, amountOut, amountIn, spreadRewardChargeTotal := swapStrategy.ComputeSwapStepInGivenOut(
			swapState.sqrtPrice,
			sqrtPriceTarget,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
		)

		swapState.updateSpreadRewardGrowthGlobal(spreadRewardChargeTotal)

		ctx.Logger().Debug("cl calc in given out")
		ctx.Logger().Debug("start sqrt price", swapState.sqrtPrice)
		ctx.Logger().Debug("reached sqrt price", sqrtPrice)
		ctx.Logger().Debug("liquidity", swapState.liquidity)
		ctx.Logger().Debug("amountIn", amountIn)
		ctx.Logger().Debug("amountOut", amountOut)
		ctx.Logger().Debug("spreadRewardChargeTotal", spreadRewardChargeTotal)

		// update the swapState with the new sqrtPrice from the above swap
		swapState.sqrtPrice = sqrtPrice
		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.SubMut(amountOut)
		swapState.amountCalculated = swapState.amountCalculated.AddMut(amountIn.Add(spreadRewardChargeTotal))
		totalSpreadFactors = totalSpreadFactors.Add(spreadRewardChargeTotal)

		// if the computeSwapStep calculated a sqrtPrice that is equal to the nextSqrtPrice, this means all liquidity in the current
		// tick has been consumed and we must move on to the next tick to complete the swap
		if sqrtPriceNextTick.Equal(sqrtPrice) {
			swapState, err = k.swapCrossTickLogic(ctx, swapState, swapStrategy,
				nextTick, nextTickIter, p, spreadRewardAccumulator, uptimeAccums, tokenInDenom)
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
			}
		} else if !sqrtPriceStart.Equal(sqrtPrice) {
			// otherwise if the sqrtPrice calculated from computeSwapStep does not equal the sqrtPrice we started with at the
			// beginning of this iteration, we set the swapState tick to the corresponding tick of the sqrtPrice calculated from computeSwapStep
			price := sqrtPrice.Mul(sqrtPrice)
			swapState.tick, err = math.PriceToTickRoundDown(price, p.GetTickSpacing())
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, 0, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
			}
		}
	}

	// Add spread reward growth per share to the pool-global spread reward accumulator.
	spreadRewardAccumulator.AddToAccumulator(sdk.NewDecCoins(sdk.NewDecCoinFromDec(tokenInDenom, swapState.spreadRewardGrowthGlobal)))

	// coin amounts require int values
	// Round amount in up to avoid under charging the user.
	amt0 := swapState.amountCalculated.Ceil().TruncateInt()
	// Round amount out down to avoid over charging the pool.
	amt1 := desiredTokenOut.Amount.ToDec().Sub(swapState.amountSpecifiedRemaining).TruncateInt()

	ctx.Logger().Debug("final amount in", amt0)
	ctx.Logger().Debug("final amount out", amt1)

	tokenIn = sdk.NewCoin(tokenInDenom, amt0)
	tokenOut = sdk.NewCoin(desiredTokenOut.Denom, amt1)

	return tokenIn, tokenOut, swapState.tick, swapState.liquidity, swapState.sqrtPrice, totalSpreadFactors, nil
}

// logic for crossing a tick during a swap
func (k Keeper) swapCrossTickLogic(ctx sdk.Context,
	swapState SwapState, swapStrategy swapstrategy.SwapStrategy,
	nextTick int64, nextTickIter db.Iterator,
	p types.ConcentratedPoolExtension,
	spreadRewardAccum accum.AccumulatorObject, uptimeAccums []accum.AccumulatorObject,
	tokenInDenom string) (SwapState, error) {
	nextTickInfo, err := ParseTickFromBz(nextTickIter.Value())
	if err != nil {
		return swapState, err
	}

	if err := k.updateGivenPoolUptimeAccumulatorsToNow(ctx, p, uptimeAccums); err != nil {
		return swapState, err
	}

	// Retrieve the liquidity held in the next closest initialized tick
	liquidityNet, err := k.crossTick(ctx, p.GetId(), nextTick, &nextTickInfo, sdk.NewDecCoinFromDec(tokenInDenom, swapState.spreadRewardGrowthGlobal), spreadRewardAccum.GetValue(), uptimeAccums)
	if err != nil {
		return swapState, err
	}

	// Move next tick iterator to the next tick as the tick is crossed.
	nextTickIter.Next()

	liquidityNet = swapStrategy.SetLiquidityDeltaSign(liquidityNet)
	// Update the swapState's liquidity with the new tick's liquidity
	newLiquidity := swapState.liquidity.AddMut(liquidityNet)
	swapState.liquidity = newLiquidity

	// Update the swapState's tick with the tick we retrieved liquidity from
	swapState.tick = nextTick
	return swapState, nil
}

// updatePoolForSwap updates the given pool object with the results of a swap operation.
//
// The method consumes a fixed amount of gas per swap to prevent spam. It applies the swap operation to the given
// pool object by calling its ApplySwap method. It then sets the updated pool object using the setPool method
// of the keeper. Finally, it transfers the input and output tokens to and from the sender and the pool account
// using the SendCoins method of the bank keeper.
//
// Calls AfterConcentratedPoolSwap listener. Currently, it notifies twap module about a
// a spot price update.
//
// If any error occurs during the swap operation, the method returns an error value indicating the cause of the error.
func (k Keeper) updatePoolForSwap(
	ctx sdk.Context,
	pool types.ConcentratedPoolExtension,
	sender sdk.AccAddress,
	tokenIn sdk.Coin,
	tokenOut sdk.Coin,
	newCurrentTick int64,
	newLiquidity sdk.Dec,
	newSqrtPrice sdk.Dec,
	totalSpreadFactors sdk.Dec,
) error {
	// Fixed gas consumption per swap to prevent spam
	poolId := pool.GetId()
	ctx.GasMeter().ConsumeGas(gammtypes.BalancerGasFeeForSwap, "cl pool swap computation")
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	// Spread factors should already be rounded up to a whole number dec, but we do this as a precaution
	spreadFactorsRoundedUp := sdk.NewCoin(tokenIn.Denom, totalSpreadFactors.Ceil().TruncateInt())

	// Remove the spread factors from the input token
	tokenIn.Amount = tokenIn.Amount.Sub(spreadFactorsRoundedUp.Amount)

	// Send the input token from the user to the pool's primary address
	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), sdk.Coins{
		tokenIn,
	})
	if err != nil {
		return types.InsufficientUserBalanceError{Err: err}
	}

	// Send the spread factors taken from the input token from the user to the pool's spread factor account
	if !spreadFactorsRoundedUp.IsZero() {
		err = k.bankKeeper.SendCoins(ctx, sender, pool.GetSpreadRewardsAddress(), sdk.Coins{
			spreadFactorsRoundedUp,
		})
		if err != nil {
			return types.InsufficientUserBalanceError{Err: err}
		}
	}

	// Send the output token to the sender from the pool
	err = k.bankKeeper.SendCoins(ctx, pool.GetAddress(), sender, sdk.Coins{
		tokenOut,
	})
	if err != nil {
		return types.InsufficientPoolBalanceError{Err: err}
	}

	err = pool.ApplySwap(newLiquidity, newCurrentTick, newSqrtPrice)
	if err != nil {
		return fmt.Errorf("error applying swap: %w", err)
	}

	if err := k.setPool(ctx, pool); err != nil {
		return err
	}

	k.listeners.AfterConcentratedPoolSwap(ctx, sender, poolId, sdk.Coins{tokenIn}, sdk.Coins{tokenOut})

	// TODO: move this to poolmanager and remove from here.
	// Also, remove from gamm.
	events.EmitSwapEvent(ctx, sender, pool.GetId(), sdk.Coins{tokenIn}, sdk.Coins{tokenOut})

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

func (k Keeper) setupSwapStrategy(ctx sdk.Context, p types.ConcentratedPoolExtension,
	spreadFactor sdk.Dec, tokenInDenom string,
	priceLimit sdk.Dec) (strategy swapstrategy.SwapStrategy, sqrtPriceLimit sdk.Dec, err error) {
	zeroForOne := getZeroForOne(tokenInDenom, p.GetToken0())

	// take provided price limit and turn this into a sqrt price limit since formulas use sqrtPrice
	sqrtPriceLimit, err = swapstrategy.GetSqrtPriceLimit(priceLimit, zeroForOne)
	if err != nil {
		return strategy, sdk.Dec{}, types.SqrtRootCalculationError{SqrtPriceLimit: sqrtPriceLimit}
	}

	// set the swap strategy
	swapStrategy := swapstrategy.New(zeroForOne, sqrtPriceLimit, k.storeKey, spreadFactor)

	// get current sqrt price from pool
	curSqrtPrice := p.GetCurrentSqrtPrice()
	if err := swapStrategy.ValidateSqrtPrice(sqrtPriceLimit, curSqrtPrice); err != nil {
		return strategy, sdk.Dec{}, err
	}

	return swapStrategy, sqrtPriceLimit, nil
}

func (k Keeper) getPoolForSwap(ctx sdk.Context, poolId uint64) (types.ConcentratedPoolExtension, error) {
	p, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return p, err
	}
	hasPositionInPool, err := k.HasAnyPositionForPool(ctx, poolId)
	if err != nil {
		return p, err
	}
	if !hasPositionInPool {
		return p, types.NoSpotPriceWhenNoLiquidityError{PoolId: poolId}
	}
	return p, nil
}

func (k Keeper) getSwapAccumulators(ctx sdk.Context, poolId uint64) (accum.AccumulatorObject, []accum.AccumulatorObject, error) {
	spreadAccum, err := k.GetSpreadRewardAccumulator(ctx, poolId)
	if err != nil {
		return accum.AccumulatorObject{}, []accum.AccumulatorObject{}, err
	}
	uptimeAccums, err := k.GetUptimeAccumulators(ctx, poolId)
	if err != nil {
		return accum.AccumulatorObject{}, []accum.AccumulatorObject{}, err
	}
	return spreadAccum, uptimeAccums, nil
}
