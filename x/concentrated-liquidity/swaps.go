package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/model"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

type SwapState struct {
	amountSpecifiedRemaining sdk.Dec // remaining amount of tokens that need to be bought by the pool
	amountCalculated         sdk.Dec // amount out
	sqrtPrice                sdk.Dec // new current price when swap is done
	tick                     sdk.Int // new tick when swap is done
	liquidity                sdk.Dec // new liquidity when swap is done
}

func (k Keeper) CreateNewConcentratedLiquidityPool(ctx sdk.Context, poolId uint64, denom0, denom1 string, currSqrtPrice sdk.Dec, currTick sdk.Int) (types.ConcentratedPoolExtension, error) {
	denom0, denom1, err := types.OrderInitialPoolDenoms(denom0, denom1)
	if err != nil {
		return nil, err
	}
	pool := &model.Pool{
		// TODO: move gammtypes.NewPoolAddress(poolId) to swaproutertypes
		Address:          gammtypes.NewPoolAddress(poolId).String(),
		Id:               poolId,
		CurrentSqrtPrice: currSqrtPrice,
		CurrentTick:      currTick,
		Liquidity:        sdk.ZeroDec(),
		Token0:           denom0,
		Token1:           denom1,
	}

	err = k.setPool(ctx, pool)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool swaproutertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	swapFee sdk.Dec,
) (sdk.Int, error) {
	return sdk.Int{}, nil
}

func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool swaproutertypes.PoolI,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	swapFee sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	return sdk.Int{}, nil
}

// SwapOutAmtGivenIn is the internal mutative method for CalcOutAmtGivenIn. Utilizing CalcOutAmtGivenIn's output, this function applies the
// new tick, liquidity, and sqrtPrice to the respective pool
func (k Keeper) SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec, priceLimit sdk.Dec, poolId uint64) (tokenOut sdk.Coin, err error) {
	tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, err := k.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, priceLimit, poolId)
	if err != nil {
		return sdk.Coin{}, err
	}
	// applySwap mutates the pool state to apply the new tick, liquidity and sqrtPrice
	err = k.applySwap(ctx, tokenIn, tokenOut, poolId, newLiquidity, newCurrentTick, newSqrtPrice)
	if err != nil {
		return sdk.Coin{}, err
	}

	return tokenOut, nil
}

// CalcOutAmtGivenIn calculates tokens to be swapped out given the provided amount and fee deducted. It also returns
// what the updated tick, liquidity, and currentSqrtPrice for the pool would be after this swap.
// Note this method is non-mutative, so the values returned by CalcOutAmtGivenIn do not get stored
func (k Keeper) CalcOutAmtGivenIn(ctx sdk.Context,
	tokenInMin sdk.Coin,
	tokenOutDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec,
	poolId uint64) (tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error) {
	p, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}
	asset0 := p.GetToken0()
	asset1 := p.GetToken1()
	tokenAmountInAfterFee := tokenInMin.Amount.ToDec().Mul(sdk.OneDec().Sub(swapFee))

	// if swapping asset0 for asset1, zeroForOne is true
	zeroForOne := tokenInMin.Denom == asset0
	swapStrategy := math.NewSwapStrategy(zeroForOne)

	// get current sqrt price from pool
	curSqrtPrice := p.GetCurrentSqrtPrice()

	// take provided price limit and turn this into a sqrt price limit since formulas use sqrtPrice
	sqrtPriceLimit, err := priceLimit.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("issue calculating square root of price limit")
	}

	// check that the price limit is below the current sqrt price but not lower than the minimum sqrt ratio if we are swapping asset0 for asset1
	// check that the price limit is above the current sqrt price but not greater than the maximum sqrt ratio if we are swapping asset1 for asset0
	if (zeroForOne && (sqrtPriceLimit.GT(p.GetCurrentSqrtPrice()) || sqrtPriceLimit.LT(types.MinSqrtRatio))) ||
		(!zeroForOne && (sqrtPriceLimit.LT(p.GetCurrentSqrtPrice()) || sqrtPriceLimit.GT(types.MaxSqrtRatio))) {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("invalid price limit (%s)", priceLimit.String())
	}
	// check that the specified tokenIn matches one of the assets in the specified pool
	if tokenInMin.Denom != asset0 && tokenInMin.Denom != asset1 {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("tokenIn (%s) does not match any asset in pool", tokenInMin.Denom)
	}
	// check that the specified tokenOut matches one of the assets in the specified pool
	if tokenOutDenom != asset0 && tokenOutDenom != asset1 {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("tokenOutDenom (%s) does not match any asset in pool", tokenOutDenom)
	}
	// check that token in and token out are different denominations
	if tokenInMin.Denom == tokenOutDenom {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("tokenIn (%s) cannot be the same as tokenOut (%s)", tokenInMin.Denom, tokenOutDenom)
	}

	// initialize swap state with the following parameters:
	// as we iterate through the following for loop, this swap state will get updated after each required iteration
	swapState := SwapState{
		amountSpecifiedRemaining: tokenAmountInAfterFee, // tokenIn
		amountCalculated:         sdk.ZeroDec(),         // tokenOut
		sqrtPrice:                curSqrtPrice,
		tick:                     p.GetCurrentTick(),
		liquidity:                p.GetLiquidity(),
	}

	// iterate and update swapState until we swap all tokenIn or we reach the specific sqrtPriceLimit
	// TODO: for now, we check if amountSpecifiedRemaining is GT 0.0000001. This is because there are times when the remaining
	// amount may be extremely small, and that small amount cannot generate and amountIn/amountOut and we are therefore left
	// in an infinite loop.
	for swapState.amountSpecifiedRemaining.GT(sdk.MustNewDecFromStr("0.0000001")) && !swapState.sqrtPrice.Equal(sqrtPriceLimit) {
		// log the sqrtPrice we start the iteration with
		sqrtPriceStart := swapState.sqrtPrice

		// we first check to see what the position of the nearest initialized tick is
		// if zeroForOne is false, we look to the left of the tick the current sqrt price is at
		// if zeroForOne is true, we look to the right of the tick the current sqrt price is at
		// if no ticks are initialized (no users have created liquidity positions) then we return an error
		nextTick, ok := k.NextInitializedTick(ctx, poolId, swapState.tick.Int64(), zeroForOne)
		if !ok {
			return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("there are no more ticks initialized to fill the swap")
		}

		// utilizing the next initialized tick, we find the corresponding nextSqrtPrice (the target sqrtPrice)
		nextSqrtPrice, err := math.TickToSqrtPrice(sdk.NewInt(nextTick))
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("could not convert next tick (%v) to nextSqrtPrice", sdk.NewInt(nextTick))
		}

		var sqrtPriceTarget sdk.Dec
		// as long as the nextSqrtPrice (calculated above) is within the user defined price limit, we set it as the target sqrtPrice
		// if it is outside the user defined price limit, we set the target sqrtPrice to the user defined price limit
		if zeroForOne && nextSqrtPrice.LT(sqrtPriceLimit) || !zeroForOne && nextSqrtPrice.GT(sqrtPriceLimit) {
			sqrtPriceTarget = sqrtPriceLimit
		} else {
			sqrtPriceTarget = nextSqrtPrice
		}

		// utilizing the bucket's liquidity and knowing the price target, we calculate the how much tokenOut we get from the tokenIn
		// we also calculate the swap state's new sqrtPrice after this swap
		sqrtPrice, amountIn, amountOut := swapStrategy.ComputeSwapStep(
			swapState.sqrtPrice,
			sqrtPriceTarget,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
		)

		// update the swapState with the new sqrtPrice from the above swap
		swapState.sqrtPrice = sqrtPrice
		// we deduct the amount of tokens we input in the computeSwapStep above from the user's defined tokenIn amount
		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.Sub(amountIn)
		// we add the amount of tokens we received (amountOut) from the computeSwapStep above to the amountCalculated accumulator
		swapState.amountCalculated = swapState.amountCalculated.Add(amountOut)

		// if the computeSwapStep calculated a sqrtPrice that is equal to the nextSqrtPrice, this means all liquidity in the current
		// tick has been consumed and we must move on to the next tick to complete the swap
		if nextSqrtPrice.Equal(sqrtPrice) {
			newSwapStrategy := math.NewSwapStrategy(zeroForOne)
			// retrieve the liquidity held in the next closest initialized tick
			liquidityNet, err := k.crossTick(ctx, p.GetId(), nextTick)
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
			}
			liquidityNet = newSwapStrategy.SetLiquidityDeltaSign(liquidityNet)
			// update the swapState's liquidity with the new tick's liquidity
			newLiquidity := math.AddLiquidity(swapState.liquidity, liquidityNet)
			swapState.liquidity = newLiquidity

			// update the swapState's tick with the tick we retrieved liquidity from
			swapState.tick = newSwapStrategy.SetNextTick(nextTick)
		} else if !sqrtPriceStart.Equal(sqrtPrice) {
			// otherwise if the sqrtPrice calculated from computeSwapStep does not equal the sqrtPrice we started with at the
			// beginning of this iteration, we set the swapState tick to the corresponding tick of the sqrtPrice calculated from computeSwapStep
			swapState.tick = math.PriceToTick(sqrtPrice.Power(2))
		}
	}

	// coin amounts require int values
	// round amountIn up to avoid under charging
	amt0 := tokenAmountInAfterFee.Sub(swapState.amountSpecifiedRemaining).RoundInt()
	amt1 := swapState.amountCalculated.TruncateInt()

	tokenIn = sdk.NewCoin(tokenInMin.Denom, amt0)
	tokenOut = sdk.NewCoin(tokenOutDenom, amt1)

	return tokenIn, tokenOut, swapState.tick, swapState.liquidity, swapState.sqrtPrice, nil
}

func (k *Keeper) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec, minPrice, maxPrice sdk.Dec, poolId uint64) (tokenIn sdk.Coin, err error) {
	tokenInCoin, newLiquidity, newCurrentTick, newCurrentSqrtPrice, err := k.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, sdk.ZeroDec(), sdk.NewDec(9999999999), poolId)
	if err != nil {
		return sdk.Coin{}, err
	}

	if err := k.applySwap(ctx, tokenInCoin, tokenOut, poolId, newLiquidity, newCurrentTick, newCurrentSqrtPrice); err != nil {
		return sdk.Coin{}, err
	}

	if tokenInCoin.Amount.GT(tokenIn.Amount) {
		return sdk.Coin{}, fmt.Errorf("tokenIn calculated is larger than tokenIn provided")
	}

	return tokenInCoin, nil
}

func (k Keeper) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec, minPrice, maxPrice sdk.Dec, poolId uint64) (sdk.Coin, sdk.Dec, sdk.Int, sdk.Dec, error) {
	tokenOutAmt := tokenOut.Amount.ToDec()
	p, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Coin{}, sdk.Dec{}, sdk.Int{}, sdk.Dec{}, err
	}
	asset0 := p.GetToken0()
	asset1 := p.GetToken1()
	zeroForOne := tokenOut.Denom == asset0
	swapStrategy := math.NewSwapStrategy(zeroForOne)

	// get current sqrt price from pool
	curSqrtPrice := p.GetCurrentSqrtPrice()

	// validation
	if tokenOut.Denom != asset0 && tokenOut.Denom != asset1 {
		return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), fmt.Errorf("tokenOut denom (%s) does not match any asset in pool", tokenOut.Denom)
	}
	if tokenInDenom != asset0 && tokenInDenom != asset1 {
		return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), fmt.Errorf("tokenInDenom (%s) does not match any asset in pool", tokenInDenom)
	}
	if tokenOut.Denom == tokenInDenom {
		return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), fmt.Errorf("tokenOut (%s) cannot be the same as tokenIn (%s)", tokenOut.Denom, tokenInDenom)
	}
	if minPrice.GTE(curSqrtPrice.Power(2)) {
		return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), fmt.Errorf("minPrice (%s) must be less than current price (%s)", minPrice, curSqrtPrice.Power(2))
	}
	if maxPrice.LTE(curSqrtPrice.Power(2)) {
		return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), fmt.Errorf("maxPrice (%s) must be greater than current price (%s)", maxPrice, curSqrtPrice.Power(2))
	}

	// sqrtPrice of upper and lower user defined price range
	sqrtPLowerTick, err := minPrice.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), fmt.Errorf("issue calculating square root of minPrice")
	}

	sqrtPUpperTick, err := maxPrice.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), fmt.Errorf("issue calculating square root of maxPrice")
	}

	// TODO: How do we remove/generalize this? I am stumped.
	amountETH := sdk.NewInt(1000000)
	amountUSDC := sdk.NewInt(5000000000)

	// find liquidity of assetA and assetB
	liq0 := math.Liquidity0(amountETH, curSqrtPrice, sqrtPUpperTick)
	liq1 := math.Liquidity1(amountUSDC, curSqrtPrice, sqrtPLowerTick)

	// utilize the smaller liquidity between assetA and assetB when performing the swap calculation
	liq := sdk.MinDec(liq0, liq1)

	swapState := SwapState{
		amountSpecifiedRemaining: tokenOutAmt,
		amountCalculated:         sdk.ZeroDec(),
		sqrtPrice:                curSqrtPrice,
		tick:                     p.GetCurrentTick(),
	}

	// TODO: This should be GT 0 but some instances have very small remainder
	// need to look into fixing this
	for swapState.amountSpecifiedRemaining.GT(sdk.NewDecWithPrec(1, 6)) {
		nextTick, ok := k.NextInitializedTick(ctx, poolId, swapState.tick.Int64(), zeroForOne)

		// TODO: we can enable this error checking once we fix tick initialization
		if !ok {
			return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), fmt.Errorf("there are no more ticks initialized to fill the swap")
		}
		nextSqrtPrice, err := math.TickToSqrtPrice(sdk.NewInt(nextTick))
		if err != nil {
			return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), err
		}

		// TODO: In and out get flipped based on if we are calculating for in or out, need to fix this
		sqrtPrice, amountIn, amountOut := swapStrategy.ComputeSwapStep(
			swapState.sqrtPrice,
			nextSqrtPrice,
			liq,
			swapState.amountSpecifiedRemaining,
		)

		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.Sub(amountIn)
		swapState.amountCalculated = swapState.amountCalculated.Add(amountOut.Quo(sdk.OneDec().Sub(swapFee)))

		if swapState.sqrtPrice.Equal(sqrtPrice) {
			newSwapStrategy := math.NewSwapStrategy(!zeroForOne)
			liquidityDelta, err := k.crossTick(ctx, p.GetId(), nextTick)
			if err != nil {
				return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), err
			}
			liquidityDelta = swapStrategy.SetLiquidityDeltaSign(liquidityDelta)
			swapState.liquidity = swapState.liquidity.Add(liquidityDelta)
			if swapState.liquidity.LTE(sdk.ZeroDec()) || swapState.liquidity.IsNil() {
				return sdk.Coin{}, sdk.ZeroDec(), sdk.ZeroInt(), sdk.ZeroDec(), fmt.Errorf("no liquidity available, cannot swap")
			}
			swapState.tick = newSwapStrategy.SetNextTick(nextTick)
		} else {
			swapState.tick = math.PriceToTick(sqrtPrice.Power(2))
		}
	}

	return sdk.NewCoin(tokenInDenom, swapState.amountCalculated.RoundInt()), swapState.liquidity, swapState.tick, swapState.sqrtPrice, nil
}

// applySwap persists the swap state and charges gas fees.
// TODO: test this and make sure that gas consumption is taken
func (k *Keeper) applySwap(ctx sdk.Context, tokenIn sdk.Coin, tokenOut sdk.Coin, poolId uint64, newLiquidity sdk.Dec, newCurrentTick sdk.Int, newCurrentSqrtPrice sdk.Dec) error {
	// Fixed gas consumption per swap to prevent spam
	ctx.GasMeter().ConsumeGas(gammtypes.BalancerGasFeeForSwap, "cl pool swap computation")
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	if err := pool.ApplySwap(newLiquidity, newCurrentTick, newCurrentSqrtPrice); err != nil {
		return err
	}

	if err := k.setPool(ctx, pool); err != nil {
		return err
	}

	return nil
}
