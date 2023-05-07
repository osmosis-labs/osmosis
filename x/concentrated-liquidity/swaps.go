package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	events "github.com/osmosis-labs/osmosis/v15/x/poolmanager/events"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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
	tick sdk.Int

	// Current liqudiity within the active tick.
	// Initialized to the pool's current tick's liquidity.
	// Updated each time a tick is crossed.
	liquidity sdk.Dec

	// Global fee growth per-current swap.
	// Initialized to zero.
	// Updated after every swap step.
	feeGrowthGlobal sdk.Dec
}

// updateFeeGrowthGlobal updates the swap state's fee growth global per unit of liquidity
// when liquidity is positive.
//
// If the liquidity is zero, this is a no-op. This case may occur when there is no liquidity
// between the ticks.This is possible when there are only 2 positions with no overlapping ranges.
// As a result, the range from the end of position one to the beginning of position
// two has no liquidity and can be skipped.
// TODO: test
func (ss *SwapState) updateFeeGrowthGlobal(feeChargeTotal sdk.Dec) {
	if !ss.liquidity.IsZero() {
		feeChargePerUnitOfLiquidity := feeChargeTotal.Quo(ss.liquidity)
		ss.feeGrowthGlobal = ss.feeGrowthGlobal.Add(feeChargePerUnitOfLiquidity)
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
	swapFee sdk.Dec,
) (tokenOutAmount sdk.Int, err error) {
	if tokenIn.Denom == tokenOutDenom {
		return sdk.Int{}, types.DenomDuplicatedError{TokenInDenom: tokenIn.Denom, TokenOutDenom: tokenOutDenom}
	}

	pool, err := convertPoolInterfaceToConcentrated(poolI)
	if err != nil {
		return sdk.Int{}, err
	}

	// determine if we are swapping asset0 for asset1 or vice versa
	asset0 := pool.GetToken0()
	zeroForOne := tokenIn.Denom == asset0

	// change priceLimit based on which direction we are swapping
	priceLimit := swapstrategy.GetPriceLimit(zeroForOne)
	tokenIn, tokenOut, _, _, _, err := k.swapOutAmtGivenIn(ctx, sender, pool, tokenIn, tokenOutDenom, swapFee, priceLimit)
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

func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolI poolmanagertypes.PoolI,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	swapFee sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	if tokenOut.Denom == tokenInDenom {
		return sdk.Int{}, types.DenomDuplicatedError{TokenInDenom: tokenInDenom, TokenOutDenom: tokenOut.Denom}
	}

	pool, err := convertPoolInterfaceToConcentrated(poolI)
	if err != nil {
		return sdk.Int{}, err
	}

	// determine if we are swapping asset0 for asset1 or vice versa
	asset1 := pool.GetToken1()
	// if swapping asset0 (in) for asset1 (out), zeroForOne is true
	zeroForOne := tokenOut.Denom == asset1

	// change priceLimit based on which direction we are swapping
	priceLimit := swapstrategy.GetPriceLimit(zeroForOne)
	tokenIn, tokenOut, _, _, _, err := k.swapInAmtGivenOut(ctx, sender, pool, tokenOut, tokenInDenom, swapFee, priceLimit)
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

// SwapOutAmtGivenIn is the internal mutative method for CalcOutAmtGivenIn. Utilizing CalcOutAmtGivenIn's output, this function applies the
// new tick, liquidity, and sqrtPrice to the respective pool
func (k Keeper) swapOutAmtGivenIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.ConcentratedPoolExtension,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec,
) (calcTokenIn, calcTokenOut sdk.Coin, currentTick sdk.Int, liquidity, sqrtPrice sdk.Dec, err error) {
	writeCtx, tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, err := k.calcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, priceLimit, pool.GetId())
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}

	if !tokenOut.Amount.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, types.InvalidAmountCalculatedError{Amount: tokenOut.Amount}
	}

	// N.B. making the call below ensures that any mutations done inside calcOutAmtGivenIn
	// are written to store. If this call were skipped, calcOutAmtGivenIn would be non-mutative.
	// An example of a store write done in calcOutAmtGivenIn is updating ticks as we cross them.
	writeCtx()

	// Settles balances between the tx sender and the pool to match the swap that was executed earlier.
	// Also emits swap event and updates related liquidity metrics
	if err := k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}

	return tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, nil
}

func (k *Keeper) swapInAmtGivenOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.ConcentratedPoolExtension,
	desiredTokenOut sdk.Coin,
	tokenInDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec,
) (calcTokenIn, calcTokenOut sdk.Coin, currentTick sdk.Int, liquidity, sqrtPrice sdk.Dec, err error) {
	writeCtx, tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, err := k.calcInAmtGivenOut(ctx, desiredTokenOut, tokenInDenom, swapFee, priceLimit, pool.GetId())
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}

	// check that the tokenOut calculated is both valid and less than specified limit
	if !tokenIn.Amount.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, types.InvalidAmountCalculatedError{Amount: tokenIn.Amount}
	}

	// N.B. making the call below ensures that any mutations done inside calcInAmtGivenOut
	// are written to store. If this call were skipped, calcInAmtGivenOut would be non-mutative.
	// An example of a store write done in calcInAmtGivenOut is updating ticks as we cross them.
	writeCtx()

	// Settles balances between the tx sender and the pool to match the swap that was executed earlier.
	// Also emits swap event and updates related liquidity metrics
	if err := k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}

	return tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, nil
}

func (k Keeper) CalcOutAmtGivenIn(
	ctx sdk.Context,
	poolI poolmanagertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	swapFee sdk.Dec,
) (tokenOut sdk.Coin, err error) {
	_, _, tokenOut, _, _, _, err = k.calcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, sdk.ZeroDec(), poolI.GetId())
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
	swapFee sdk.Dec,
) (tokenIn sdk.Coin, err error) {
	_, tokenIn, _, _, _, _, err = k.calcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, sdk.ZeroDec(), poolI.GetId())
	if err != nil {
		return sdk.Coin{}, err
	}
	return tokenIn, nil
}

// calcOutAmtGivenIn calculates tokens to be swapped out given the provided amount and fee deducted. It also returns
// what the updated tick, liquidity, and currentSqrtPrice for the pool would be after this swap.
// Note this method is non-mutative, so the values returned by CalcOutAmtGivenIn do not get stored
// Instead, we return writeCtx function so that the caller of this method can decide to write the cached ctx to store or not.
func (k Keeper) calcOutAmtGivenIn(ctx sdk.Context,
	tokenInMin sdk.Coin,
	tokenOutDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec,
	poolId uint64,
) (writeCtx func(), tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error) {
	ctx, writeCtx = ctx.CacheContext()
	p, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}
	asset0 := p.GetToken0()
	asset1 := p.GetToken1()
	tokenAmountInSpecified := tokenInMin.Amount.ToDec()

	// if swapping asset0 for asset1, zeroForOne is true
	zeroForOne := tokenInMin.Denom == asset0

	// if priceLimit not set, set to max/min value based on swap direction
	if zeroForOne && priceLimit.Equal(sdk.ZeroDec()) {
		priceLimit = types.MinSpotPrice
	} else if !zeroForOne && priceLimit.Equal(sdk.ZeroDec()) {
		priceLimit = types.MaxSpotPrice
	}

	// take provided price limit and turn this into a sqrt price limit since formulas use sqrtPrice
	sqrtPriceLimit, err := priceLimit.ApproxSqrt()
	if err != nil {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("issue calculating square root of price limit")
	}

	// set the swap strategy
	swapStrategy := swapstrategy.New(zeroForOne, sqrtPriceLimit, k.storeKey, swapFee, p.GetTickSpacing())

	// get current sqrt price from pool
	curSqrtPrice := p.GetCurrentSqrtPrice()
	if err := swapStrategy.ValidateSqrtPrice(sqrtPriceLimit, curSqrtPrice); err != nil {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}

	// check that the specified tokenIn matches one of the assets in the specified pool
	if tokenInMin.Denom != asset0 && tokenInMin.Denom != asset1 {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, types.TokenInDenomNotInPoolError{TokenInDenom: tokenInMin.Denom}
	}
	// check that the specified tokenOut matches one of the assets in the specified pool
	if tokenOutDenom != asset0 && tokenOutDenom != asset1 {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, types.TokenOutDenomNotInPoolError{TokenOutDenom: tokenOutDenom}
	}
	// check that token in and token out are different denominations
	if tokenInMin.Denom == tokenOutDenom {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, types.DenomDuplicatedError{TokenInDenom: tokenInMin.Denom, TokenOutDenom: tokenOutDenom}
	}

	// initialize swap state with the following parameters:
	// as we iterate through the following for loop, this swap state will get updated after each required iteration
	swapState := SwapState{
		amountSpecifiedRemaining: tokenAmountInSpecified, // tokenIn
		amountCalculated:         sdk.ZeroDec(),          // tokenOut
		sqrtPrice:                curSqrtPrice,
		tick:                     swapStrategy.InitializeTickValue(p.GetCurrentTick()),
		liquidity:                p.GetLiquidity(),
		feeGrowthGlobal:          sdk.ZeroDec(),
	}

	// iterate and update swapState until we swap all tokenIn or we reach the specific sqrtPriceLimit
	// TODO: for now, we check if amountSpecifiedRemaining is GT 0.0000001. This is because there are times when the remaining
	// amount may be extremely small, and that small amount cannot generate and amountIn/amountOut and we are therefore left
	// in an infinite loop.
	for swapState.amountSpecifiedRemaining.GT(sdk.SmallestDec()) && !swapState.sqrtPrice.Equal(sqrtPriceLimit) {
		// log the sqrtPrice we start the iteration with
		sqrtPriceStart := swapState.sqrtPrice

		// we first check to see what the position of the nearest initialized tick is
		// if zeroForOneStrategy, we look to the left of the tick the current sqrt price is at
		// if oneForZeroStrategy, we look to the right of the tick the current sqrt price is at
		// if no ticks are initialized (no users have created liquidity positions) then we return an error
		nextTick, ok := swapStrategy.NextInitializedTick(ctx, poolId, swapState.tick.Int64())
		if !ok {
			return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("there are no more ticks initialized to fill the swap")
		}

		// utilizing the next initialized tick, we find the corresponding nextPrice (the target price)
		nextTickSqrtPrice, err := math.TickToSqrtPrice(nextTick)
		if err != nil {
			return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("could not convert next tick (%v) to nextSqrtPrice", nextTick)
		}

		sqrtPriceTarget := swapStrategy.GetSqrtTargetPrice(nextTickSqrtPrice)

		// utilizing the bucket's liquidity and knowing the price target, we calculate the how much tokenOut we get from the tokenIn
		// we also calculate the swap state's new sqrtPrice after this swap
		sqrtPrice, amountIn, amountOut, feeCharge := swapStrategy.ComputeSwapStepOutGivenIn(
			swapState.sqrtPrice,
			sqrtPriceTarget,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
		)

		// Update the fee growth for the entire swap using the total fees charged.
		swapState.updateFeeGrowthGlobal(feeCharge)

		ctx.Logger().Debug("cl calc out given in")
		ctx.Logger().Debug("start sqrt price", swapState.sqrtPrice)
		ctx.Logger().Debug("reached sqrt price", sqrtPrice)
		ctx.Logger().Debug("liquidity", swapState.liquidity)
		ctx.Logger().Debug("amountIn", amountIn)
		ctx.Logger().Debug("amountOut", amountOut)
		ctx.Logger().Debug("feeCharge", feeCharge)

		// update the swapState with the new sqrtPrice from the above swap
		swapState.sqrtPrice = sqrtPrice
		// we deduct the amount of tokens we input in the computeSwapStep above from the user's defined tokenIn amount
		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.Sub(amountIn.Add(feeCharge))
		// we add the amount of tokens we received (amountOut) from the computeSwapStep above to the amountCalculated accumulator
		swapState.amountCalculated = swapState.amountCalculated.Add(amountOut)

		// if the computeSwapStep calculated a sqrtPrice that is equal to the nextSqrtPrice, this means all liquidity in the current
		// tick has been consumed and we must move on to the next tick to complete the swap
		if nextTickSqrtPrice.Equal(sqrtPrice) {
			// retrieve the liquidity held in the next closest initialized tick
			liquidityNet, err := k.crossTick(ctx, p.GetId(), nextTick.Int64(), sdk.NewDecCoinFromDec(tokenInMin.Denom, swapState.feeGrowthGlobal))
			if err != nil {
				return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
			}
			liquidityNet = swapStrategy.SetLiquidityDeltaSign(liquidityNet)
			// update the swapState's liquidity with the new tick's liquidity
			newLiquidity := math.AddLiquidity(swapState.liquidity, liquidityNet)
			swapState.liquidity = newLiquidity

			// update the swapState's tick with the tick we retrieved liquidity from
			swapState.tick = nextTick
		} else if !sqrtPriceStart.Equal(sqrtPrice) {
			// otherwise if the sqrtPrice calculated from computeSwapStep does not equal the sqrtPrice we started with at the
			// beginning of this iteration, we set the swapState tick to the corresponding tick of the sqrtPrice calculated from computeSwapStep
			price := sqrtPrice.Mul(sqrtPrice)
			swapState.tick, err = math.PriceToTickRoundDown(price, p.GetTickSpacing())
			if err != nil {
				return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
			}
		}
	}

	if err := k.chargeFee(ctx, poolId, sdk.NewDecCoinFromDec(tokenInMin.Denom, swapState.feeGrowthGlobal)); err != nil {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}

	// coin amounts require int values
	// round amountIn up to avoid under charging
	amt0 := tokenAmountInSpecified.Sub(swapState.amountSpecifiedRemaining).Ceil().TruncateInt()
	// round amountOut down to avoid over refunding.
	amt1 := swapState.amountCalculated.TruncateInt()

	ctx.Logger().Debug("final amount in", amt0)
	ctx.Logger().Debug("final amount out", amt1)

	tokenIn = sdk.NewCoin(tokenInMin.Denom, amt0)
	tokenOut = sdk.NewCoin(tokenOutDenom, amt1)

	return writeCtx, tokenIn, tokenOut, swapState.tick, swapState.liquidity, swapState.sqrtPrice, nil
}

// calcInAmtGivenOut calculates tokens to be swapped in given the desired token out and fee deducted. It also returns
// what the updated tick, liquidity, and currentSqrtPrice for the pool would be after this swap.
// Note this method is non-mutative, so the values returned by calcInAmtGivenOut do not get stored
// Instead, we return writeCtx function so that the caller of this method can decide to write the cached ctx to store or not.
func (k Keeper) calcInAmtGivenOut(
	ctx sdk.Context,
	desiredTokenOut sdk.Coin,
	tokenInDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec,
	poolId uint64,
) (writeCtx func(), tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error) {
	ctx, writeCtx = ctx.CacheContext()
	p, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}
	asset0 := p.GetToken0()
	asset1 := p.GetToken1()

	// if swapping asset0 (in) for asset1 (out), zeroForOne is true
	zeroForOne := desiredTokenOut.Denom == asset1

	// if priceLimit not set, set to max/min value based on swap direction
	if zeroForOne && priceLimit.Equal(sdk.ZeroDec()) {
		priceLimit = types.MinSpotPrice
	} else if !zeroForOne && priceLimit.Equal(sdk.ZeroDec()) {
		priceLimit = types.MaxSpotPrice
	}

	// take provided price limit and turn this into a sqrt price limit since formulas use sqrtPrice
	sqrtPriceLimit, err := priceLimit.ApproxSqrt()
	if err != nil {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("issue calculating square root of price limit")
	}

	// set the swap strategy
	swapStrategy := swapstrategy.New(zeroForOne, sqrtPriceLimit, k.storeKey, swapFee, p.GetTickSpacing())

	// get current sqrt price from pool
	curSqrtPrice := p.GetCurrentSqrtPrice()

	if err := swapStrategy.ValidateSqrtPrice(sqrtPriceLimit, curSqrtPrice); err != nil {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}

	// check that the specified tokenOut matches one of the assets in the specified pool
	if desiredTokenOut.Denom != asset0 && desiredTokenOut.Denom != asset1 {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, types.TokenOutDenomNotInPoolError{TokenOutDenom: desiredTokenOut.Denom}
	}
	// check that the specified tokenIn matches one of the assets in the specified pool
	if tokenInDenom != asset0 && tokenInDenom != asset1 {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, types.TokenInDenomNotInPoolError{TokenInDenom: tokenInDenom}
	}
	// check that token in and token out are different denominations
	if desiredTokenOut.Denom == tokenInDenom {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, types.DenomDuplicatedError{TokenInDenom: tokenInDenom, TokenOutDenom: desiredTokenOut.Denom}
	}

	// initialize swap state with the following parameters:
	// as we iterate through the following for loop, this swap state will get updated after each required iteration
	swapState := SwapState{
		amountSpecifiedRemaining: desiredTokenOut.Amount.ToDec(), // tokenOut
		amountCalculated:         sdk.ZeroDec(),                  // tokenIn
		sqrtPrice:                curSqrtPrice,
		tick:                     swapStrategy.InitializeTickValue(p.GetCurrentTick()),
		liquidity:                p.GetLiquidity(),
		feeGrowthGlobal:          sdk.ZeroDec(),
	}

	// TODO: This should be GT 0 but some instances have very small remainder
	// need to look into fixing this
	for swapState.amountSpecifiedRemaining.GT(sdk.SmallestDec()) && !swapState.sqrtPrice.Equal(sqrtPriceLimit) {
		// log the sqrtPrice we start the iteration with
		sqrtPriceStart := swapState.sqrtPrice

		// we first check to see what the position of the nearest initialized tick is
		// if zeroForOne is false, we look to the left of the tick the current sqrt price is at
		// if zeroForOne is true, we look to the right of the tick the current sqrt price is at
		// if no ticks are initialized (no users have created liquidity positions) then we return an error
		nextTick, ok := swapStrategy.NextInitializedTick(ctx, poolId, swapState.tick.Int64())
		if !ok {
			return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("there are no more ticks initialized to fill the swap")
		}

		// utilizing the next initialized tick, we find the corresponding nextPrice (the target price)
		sqrtPriceNextTick, err := math.TickToSqrtPrice(nextTick)
		if err != nil {
			return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("could not convert next tick (%v) to nextSqrtPrice", nextTick)
		}

		sqrtPriceTarget := swapStrategy.GetSqrtTargetPrice(sqrtPriceNextTick)

		// utilizing the bucket's liquidity and knowing the price target, we calculate the how much tokenOut we get from the tokenIn
		// we also calculate the swap state's new sqrtPrice after this swap
		sqrtPrice, amountOut, amountIn, feeChargeTotal := swapStrategy.ComputeSwapStepInGivenOut(
			swapState.sqrtPrice,
			sqrtPriceTarget,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
		)

		swapState.updateFeeGrowthGlobal(feeChargeTotal)

		ctx.Logger().Debug("cl calc in given out")
		ctx.Logger().Debug("start sqrt price", swapState.sqrtPrice)
		ctx.Logger().Debug("reached sqrt price", sqrtPrice)
		ctx.Logger().Debug("liquidity", swapState.liquidity)
		ctx.Logger().Debug("amountIn", amountIn)
		ctx.Logger().Debug("amountOut", amountOut)
		ctx.Logger().Debug("feeChargeTotal", feeChargeTotal)

		// update the swapState with the new sqrtPrice from the above swap
		swapState.sqrtPrice = sqrtPrice
		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.Sub(amountOut)
		swapState.amountCalculated = swapState.amountCalculated.Add(amountIn.Add(feeChargeTotal))

		// if the computeSwapStep calculated a sqrtPrice that is equal to the nextSqrtPrice, this means all liquidity in the current
		// tick has been consumed and we must move on to the next tick to complete the swap
		if sqrtPriceNextTick.Equal(sqrtPrice) {
			// retrieve the liquidity held in the next closest initialized tick
			liquidityNet, err := k.crossTick(ctx, p.GetId(), nextTick.Int64(), sdk.NewDecCoinFromDec(desiredTokenOut.Denom, swapState.feeGrowthGlobal))
			if err != nil {
				return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
			}
			liquidityNet = swapStrategy.SetLiquidityDeltaSign(liquidityNet)
			// update the swapState's liquidity with the new tick's liquidity
			newLiquidity := math.AddLiquidity(swapState.liquidity, liquidityNet)
			swapState.liquidity = newLiquidity

			// update the swapState's tick with the tick we retrieved liquidity from
			swapState.tick = nextTick
		} else if !sqrtPriceStart.Equal(sqrtPrice) {
			// otherwise if the sqrtPrice calculated from computeSwapStep does not equal the sqrtPrice we started with at the
			// beginning of this iteration, we set the swapState tick to the corresponding tick of the sqrtPrice calculated from computeSwapStep
			price := sqrtPrice.Mul(sqrtPrice)
			swapState.tick, err = math.PriceToTickRoundDown(price, p.GetTickSpacing())
			if err != nil {
				return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
			}
		}
	}

	if err := k.chargeFee(ctx, poolId, sdk.NewDecCoinFromDec(tokenInDenom, swapState.feeGrowthGlobal)); err != nil {
		return writeCtx, sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}

	// coin amounts require int values
	// Round amount in up to avoid under charging the user.
	amt0 := swapState.amountCalculated.Ceil().TruncateInt()
	// Round amount out down to avoid over charging the pool.
	amt1 := desiredTokenOut.Amount.ToDec().Sub(swapState.amountSpecifiedRemaining).TruncateInt()

	ctx.Logger().Debug("final amount in", amt0)
	ctx.Logger().Debug("final amount out", amt1)

	tokenIn = sdk.NewCoin(tokenInDenom, amt0)
	tokenOut = sdk.NewCoin(desiredTokenOut.Denom, amt1)

	return writeCtx, tokenIn, tokenOut, swapState.tick, swapState.liquidity, swapState.sqrtPrice, nil
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
	newCurrentTick sdk.Int,
	newLiquidity sdk.Dec,
	newSqrtPrice sdk.Dec,
) error {
	// Fixed gas consumption per swap to prevent spam
	poolId := pool.GetId()
	ctx.GasMeter().ConsumeGas(gammtypes.BalancerGasFeeForSwap, "cl pool swap computation")
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), sdk.Coins{
		tokenIn,
	})
	if err != nil {
		return types.InsufficientUserBalanceError{Err: err}
	}

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

	k.listeners.AfterConcentratedPoolSwap(ctx, sender, poolId)

	// TODO: move this to poolmanager and remove from here.
	// Also, remove from gamm.
	events.EmitSwapEvent(ctx, sender, pool.GetId(), sdk.Coins{tokenIn}, sdk.Coins{tokenOut})

	return err
}
