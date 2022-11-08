package concentrated_pool

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

type SwapState struct {
	amountSpecifiedRemaining sdk.Dec // remaining amount of tokens that need to be bought by the pool
	amountCalculated         sdk.Dec // amount out
	sqrtPrice                sdk.Dec // new current price when swap is done
	tick                     sdk.Int // new tick when swap is done
	liquidity                sdk.Dec
}

func (p Pool) CalcOutAmtGivenIn(ctx sdk.Context,
	poolTickKVStore sdk.KVStore,
	tokenInMin sdk.Coin,
	tokenOutDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec,
	poolId uint64) (tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error) {
	asset0 := p.GetToken0()
	asset1 := p.GetToken1()
	tokenAmountInAfterFee := tokenInMin.Amount.ToDec().Mul(sdk.OneDec().Sub(swapFee))

	zeroForOne := tokenInMin.Denom == asset0

	// get current sqrt price from pool
	curSqrtPrice := p.GetCurrentSqrtPrice()
	sqrtPriceLimit, err := priceLimit.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("issue calculating square root of price limit")
	}
	if (zeroForOne && (sqrtPriceLimit.GT(curSqrtPrice) || sqrtPriceLimit.LT(types.MinSqrtRatio))) ||
		(!zeroForOne && (sqrtPriceLimit.LT(curSqrtPrice) || sqrtPriceLimit.GT(types.MaxSqrtRatio))) {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("invalid price limit (%s)", priceLimit.String())
	}

	// validation
	if tokenInMin.Denom != asset0 && tokenInMin.Denom != asset1 {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("tokenIn (%s) does not match any asset in pool", tokenInMin.Denom)
	}
	if tokenOutDenom != asset0 && tokenOutDenom != asset1 {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("tokenOutDenom (%s) does not match any asset in pool", tokenOutDenom)
	}
	if tokenInMin.Denom == tokenOutDenom {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("tokenIn (%s) cannot be the same as tokenOut (%s)", tokenInMin.Denom, tokenOutDenom)
	}

	// at first, we use the pool liquidity
	swapState := SwapState{
		amountSpecifiedRemaining: tokenAmountInAfterFee,
		amountCalculated:         sdk.ZeroDec(),
		sqrtPrice:                curSqrtPrice,
		tick:                     types.PriceToTick(curSqrtPrice.Power(2)),
		liquidity:                p.GetLiquidity(),
	}

	for swapState.amountSpecifiedRemaining.GT(sdk.ZeroDec()) && !swapState.sqrtPrice.Equal(sqrtPriceLimit) {
		sqrtPriceStart := swapState.sqrtPrice
		nextTick, ok := p.NextInitializedTick(ctx, poolTickKVStore, poolId, swapState.tick.Int64(), zeroForOne)
		if !ok {
			return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("there are no more ticks initialized to fill the swap")
		}

		nextSqrtPrice, err := types.TickToSqrtPrice(sdk.NewInt(nextTick))
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("could not convert next tick (%v) to nextSqrtPrice", sdk.NewInt(nextTick))
		}

		var sqrtPriceTarget sdk.Dec
		if zeroForOne && nextSqrtPrice.LT(sqrtPriceLimit) || !zeroForOne && nextSqrtPrice.GT(sqrtPriceLimit) {
			sqrtPriceTarget = sqrtPriceLimit
		} else {
			sqrtPriceTarget = nextSqrtPrice
		}

		sqrtPrice, amountIn, amountOut := types.ComputeSwapStep(
			swapState.sqrtPrice,
			sqrtPriceTarget,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
			zeroForOne,
		)
		swapState.sqrtPrice = sqrtPrice

		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.Sub(amountIn)
		swapState.amountCalculated = swapState.amountCalculated.Add(amountOut)

		// if we have moved to the next tick,
		if nextSqrtPrice.Equal(sqrtPrice) {
			liquidityDelta, err := p.crossTick(ctx, poolTickKVStore, p.GetId(), nextTick)

			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
			}
			if zeroForOne {
				liquidityDelta = liquidityDelta.Neg()
			}

			swapState.liquidity = swapState.liquidity.Add(liquidityDelta.ToDec())

			if swapState.liquidity.LTE(sdk.ZeroDec()) || swapState.liquidity.IsNil() {
				return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
			}
			if zeroForOne {
				swapState.tick = sdk.NewInt(nextTick - 1)
			} else {
				swapState.tick = sdk.NewInt(nextTick)
			}
		} else if !sqrtPriceStart.Equal(sqrtPrice) {
			swapState.tick = types.PriceToTick(sqrtPrice.Power(2))
		}
	}

	// coin amounts require int values
	// we truncate at last step to retain as much precision as possible
	amt0 := tokenAmountInAfterFee.Add(swapState.amountSpecifiedRemaining).TruncateInt()
	amt1 := swapState.amountCalculated.TruncateInt()

	tokenIn = sdk.NewCoin(tokenInMin.Denom, amt0)
	tokenOut = sdk.NewCoin(tokenOutDenom, amt1)

	return tokenIn, tokenOut, swapState.tick, swapState.liquidity, swapState.sqrtPrice, nil
}

// TODO: implement this
func (p *Pool) SwapOutAmtGivenIn(ctx sdk.Context, kvStore sdk.KVStore, tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec, priceLimit sdk.Dec, poolId uint64) (tokenOut sdk.Coin, err error) {
	tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, err := p.CalcOutAmtGivenIn(ctx, kvStore, tokenIn, tokenOutDenom, swapFee, priceLimit, poolId)
	if err != nil {
		return sdk.Coin{}, err
	}

	err = p.applySwap(ctx, tokenIn, tokenOut, poolId, newLiquidity, newCurrentTick, newSqrtPrice)
	if err != nil {
		return sdk.Coin{}, err
	}

	return tokenOut, nil
}

func (p Pool) CalcInAmtGivenOut(ctx sdk.Context,
	poolTickKVStore sdk.KVStore,
	tokenOutMin sdk.Coin,
	tokenInDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec,
	poolId uint64) (tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error) {
	tokenOutAmt := tokenOutMin.Amount.ToDec()

	asset0 := p.GetToken0()
	asset1 := p.GetToken1()
	zeroForOne := tokenOutMin.Denom == asset0

	// get current sqrt price from pool
	curSqrtPrice := p.GetCurrentSqrtPrice()

	// validation
	if tokenOutMin.Denom != asset0 && tokenOutMin.Denom != asset1 {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("tokenOut denom (%s) does not match any asset in pool", tokenOutMin.Denom)
	}
	if tokenInDenom != asset0 && tokenInDenom != asset1 {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("tokenInDenom (%s) does not match any asset in pool", tokenInDenom)
	}
	if tokenOutMin.Denom == tokenInDenom {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("tokenOut (%s) cannot be the same as tokenIn (%s)", tokenOutMin.Denom, tokenInDenom)
	}
	sqrtPriceLimit, err := priceLimit.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("issue calculating square root of price limit")
	}
	if (zeroForOne && (sqrtPriceLimit.GT(curSqrtPrice) || sqrtPriceLimit.LT(types.MinSqrtRatio))) ||
		(!zeroForOne && (sqrtPriceLimit.LT(curSqrtPrice) || sqrtPriceLimit.GT(types.MaxSqrtRatio))) {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("invalid price limit (%s)", priceLimit.String())
	}

	swapState := SwapState{
		amountSpecifiedRemaining: tokenOutAmt,
		amountCalculated:         sdk.ZeroDec(),
		sqrtPrice:                curSqrtPrice,
		tick:                     types.PriceToTick(curSqrtPrice.Power(2)),
		liquidity:                p.GetLiquidity(),
	}

	// TODO: This should be GT 0 but some instances have very small remainder
	// need to look into fixing this
	for swapState.amountSpecifiedRemaining.GT(sdk.NewDecWithPrec(1, 6)) {
		nextTick, ok := p.NextInitializedTick(ctx, poolTickKVStore, poolId, swapState.tick.Int64(), zeroForOne)

		// TODO: we can enable this error checking once we fix tick initialization
		if !ok {
			return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("there are no more ticks initialized to fill the swap")
		}
		nextSqrtPrice, err := types.TickToSqrtPrice(sdk.NewInt(nextTick))
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
		}

		// TODO: In and out get flipped based on if we are calculating for in or out, need to fix this
		sqrtPrice, amountIn, amountOut := types.ComputeSwapStep(
			swapState.sqrtPrice,
			nextSqrtPrice,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
			zeroForOne,
		)

		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.Sub(amountIn)
		swapState.amountCalculated = swapState.amountCalculated.Add(amountOut.Quo(sdk.OneDec().Sub(swapFee)))

		if swapState.sqrtPrice.Equal(sqrtPrice) {
			liquidityDelta, err := p.crossTick(ctx, poolTickKVStore, p.GetId(), nextTick)
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
			}
			if !zeroForOne {
				liquidityDelta = liquidityDelta.Neg()
			}
			swapState.liquidity = swapState.liquidity.Add(liquidityDelta.ToDec())
			if swapState.liquidity.LTE(sdk.ZeroDec()) || swapState.liquidity.IsNil() {
				return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, fmt.Errorf("no liquidity available, cannot swap")
			}
			if !zeroForOne {
				swapState.tick = sdk.NewInt(nextTick - 1)
			} else {
				swapState.tick = sdk.NewInt(nextTick)
			}
		} else {
			swapState.tick = types.PriceToTick(sqrtPrice.Power(2))
		}
	}

	return sdk.NewCoin(tokenInDenom, swapState.amountCalculated.RoundInt()), tokenOut, swapState.tick, swapState.liquidity, swapState.sqrtPrice, nil
}

func (p *Pool) SwapInAmtGivenOut(ctx sdk.Context, kvStore sdk.KVStore, tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec, priceLimit sdk.Dec, poolId uint64) (tokenOut sdk.Coin, err error) {
	tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice, err := p.CalcInAmtGivenOut(ctx, kvStore, tokenIn, tokenOutDenom, swapFee, priceLimit, poolId)
	if err != nil {
		return sdk.Coin{}, err
	}

	err = p.applySwap(ctx, tokenIn, tokenOut, poolId, newLiquidity, newCurrentTick, newSqrtPrice)
	if err != nil {
		return sdk.Coin{}, err
	}

	return tokenOut, nil
}

// ApplySwap.
func (p *Pool) applySwap(ctx sdk.Context, tokenIn sdk.Coin, tokenOut sdk.Coin, poolId uint64, newLiquidity sdk.Dec, newCurrentTick sdk.Int, newCurrentSqrtPrice sdk.Dec) error {
	// Fixed gas consumption per swap to prevent spam
	ctx.GasMeter().ConsumeGas(gammtypes.BalancerGasFeeForSwap, "cl pool swap computation")

	p.Liquidity = newLiquidity
	p.CurrentTick = newCurrentTick
	p.CurrentSqrtPrice = newCurrentSqrtPrice
	return nil
}
