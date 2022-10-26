package concentrated_liquidity

import (
	"errors"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func (k Keeper) CreateNewConcentratedLiquidityPool(ctx sdk.Context, poolId uint64, denom0, denom1 string, currSqrtPrice sdk.Dec, currTick sdk.Int) (Pool, error) {
	poolAddr := types.NewPoolAddress(poolId)
	denom0, denom1, err := cltypes.OrderInitialPoolDenoms(denom0, denom1)
	if err != nil {
		return Pool{}, err
	}
	pool := Pool{
		Address:          poolAddr.String(),
		Id:               poolId,
		CurrentSqrtPrice: currSqrtPrice,
		CurrentTick:      currTick,
		Token0:           denom0,
		Token1:           denom1,
	}

	k.setPoolById(ctx, poolId, pool)

	return pool, nil
}

// GetPool returns a pool with a given id.
func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (types.PoolI, error) {
	return nil, errors.New("not implemented")
}

func priceToTick(price sdk.Dec) sdk.Int {
	logOfPrice := osmomath.BigDecFromSDKDec(price).LogBase2()
	logInt := osmomath.NewDecWithPrec(10001, 4)
	tick := logOfPrice.Quo(logInt.LogBase2())
	return tick.SDKDec().TruncateInt()
}

func (p Pool) GetAddress() sdk.AccAddress {
	return sdk.AccAddress{}
}

func (p Pool) String() string {
	return ""
}

func (p Pool) GetId() uint64 {
	return 0
}
func (p Pool) GetSwapFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}
func (p Pool) GetExitFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}
func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}
func (p Pool) GetTotalShares() sdk.Int {
	return sdk.Int{}
}

func (p Pool) SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	return sdk.Dec{}, nil
}

type SwapState struct {
	amountSpecifiedRemaining sdk.Dec // remaining amount of tokens that need to be bought by the pool
	amountCalculated         sdk.Dec // amount out
	sqrtPrice                sdk.Dec // new current price when swap is done
	tick                     sdk.Int // new tick when swap is done
	liquidity                sdk.Dec
}

// this only works on a single directional trade, will implement bi directional trade in next milestone
func (k Keeper) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec, priceLimit sdk.Dec, poolId uint64) (newTokenIn, tokenOut sdk.Coin, err error) {
	p := k.getPoolbyId(ctx, poolId)
	asset0 := p.Token0
	asset1 := p.Token1
	tokenAmountInAfterFee := tokenIn.Amount.ToDec().Mul(sdk.OneDec().Sub(swapFee))

	zeroForOne := tokenIn.Denom == asset0
	sqrtPriceLimit, err := priceLimit.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("issue calculating square root of price limit")
	}
	if (zeroForOne && (sqrtPriceLimit.GT(p.CurrentSqrtPrice) || sqrtPriceLimit.LT(cltypes.MinSqrtRatio))) ||
		(!zeroForOne && (sqrtPriceLimit.LT(p.CurrentSqrtPrice) || sqrtPriceLimit.GT(cltypes.MaxSqrtRatio))) {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("invlaid price limit (%s)", priceLimit.String())
	}

	// get current sqrt price from pool
	curSqrtPrice := p.CurrentSqrtPrice

	// validation
	if tokenIn.Denom != asset0 && tokenIn.Denom != asset1 {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("tokenIn (%s) does not match any asset in pool", tokenIn.Denom)
	}
	if tokenOutDenom != asset0 && tokenOutDenom != asset1 {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("tokenOutDenom (%s) does not match any asset in pool", tokenOutDenom)
	}
	if tokenIn.Denom == tokenOutDenom {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("tokenIn (%s) cannot be the same as tokenOut (%s)", tokenIn.Denom, tokenOutDenom)
	}

	swapState := SwapState{
		amountSpecifiedRemaining: tokenAmountInAfterFee,
		amountCalculated:         sdk.ZeroDec(),
		sqrtPrice:                curSqrtPrice,
		tick:                     priceToTick(curSqrtPrice.Power(2)),
		// at first, we use the pool liquidity
		liquidity: p.Liquidity.ToDec(),
	}

	// TODO: This should be GT 0 but some instances have very small remainder
	// need to look into fixing this
	for swapState.amountSpecifiedRemaining.GT(sdk.NewDecWithPrec(1, 6)) && !swapState.sqrtPrice.Equal(priceLimit) {
		nextTick, _ := k.NextInitializedTick(ctx, poolId, swapState.tick.Int64(), zeroForOne)
		// TODO: we can enable this error checking once we fix tick initialization
		// if !ok {
		// 	return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("there are no more ticks initialized to fill the swap")
		// }
		nextSqrtPrice, err := k.tickToSqrtPrice(sdk.NewInt(nextTick))
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("could not convert next tick (%v) to nextSqrtPrice", sdk.NewInt(nextTick))
		}

		sqrtPrice, amountIn, amountOut := computeSwapStep(
			swapState.sqrtPrice,
			nextSqrtPrice,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
			zeroForOne,
		)
		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.Sub(amountIn)
		swapState.amountCalculated = swapState.amountCalculated.Add(amountOut)

		if nextSqrtPrice.Equal(sqrtPrice) {
			liquidityDelta, err := k.crossTick(ctx, p.Id, nextTick)
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, err
			}
			if zeroForOne {
				liquidityDelta = liquidityDelta.Neg()
			}
			swapState.liquidity = swapState.liquidity.Add(liquidityDelta.ToDec())
			if swapState.liquidity.LTE(sdk.ZeroDec()) || swapState.liquidity.IsNil() {
				return sdk.Coin{}, sdk.Coin{}, err
			}
			if zeroForOne {
				swapState.tick = sdk.NewInt(nextTick - 1)
			} else {
				swapState.tick = sdk.NewInt(nextTick)
			}
		} else {
			swapState.tick = priceToTick(sqrtPrice.Power(2))
		}
	}

	newTokenIn.Amount = tokenIn.Amount.Sub(swapState.amountSpecifiedRemaining.RoundInt())
	return sdk.NewCoin(tokenIn.Denom, newTokenIn.Amount), sdk.NewCoin(tokenOutDenom, swapState.amountCalculated.RoundInt()), nil
}

func (p Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (k Keeper) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec, priceLimit sdk.Dec, poolId uint64) (tokenIn, newTokenOut sdk.Coin, err error) {
	tokenOutAmt := tokenOut.Amount.ToDec()
	p := k.getPoolbyId(ctx, poolId)
	asset0 := p.Token0
	asset1 := p.Token1
	zeroForOne := tokenOut.Denom == asset0

	sqrtPriceLimit, err := priceLimit.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("issue calculating square root of price limit")
	}
	if (zeroForOne && (sqrtPriceLimit.GT(p.CurrentSqrtPrice) || sqrtPriceLimit.LT(cltypes.MinSqrtRatio))) ||
		(!zeroForOne && (sqrtPriceLimit.LT(p.CurrentSqrtPrice) || sqrtPriceLimit.GT(cltypes.MaxSqrtRatio))) {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("invlaid price limit (%s)", priceLimit.String())
	}

	// get current sqrt price from pool
	curSqrtPrice := p.CurrentSqrtPrice

	// validation
	if tokenOut.Denom != asset0 && tokenOut.Denom != asset1 {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("tokenOut denom (%s) does not match any asset in pool", tokenOut.Denom)
	}
	if tokenInDenom != asset0 && tokenInDenom != asset1 {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("tokenInDenom (%s) does not match any asset in pool", tokenInDenom)
	}
	if tokenOut.Denom == tokenInDenom {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("tokenOut (%s) cannot be the same as tokenIn (%s)", tokenOut.Denom, tokenInDenom)
	}

	swapState := SwapState{
		amountSpecifiedRemaining: tokenOutAmt,
		amountCalculated:         sdk.ZeroDec(),
		sqrtPrice:                curSqrtPrice,
		tick:                     priceToTick(curSqrtPrice.Power(2)),
		// at first, we use the pool liquidity
		liquidity: p.Liquidity.ToDec(),
	}

	// TODO: This should be GT 0 but some instances have very small remainder
	// need to look into fixing this
	for swapState.amountSpecifiedRemaining.GT(sdk.NewDecWithPrec(1, 6)) && !swapState.sqrtPrice.Equal(priceLimit) {
		nextTick, _ := k.NextInitializedTick(ctx, poolId, swapState.tick.Int64(), zeroForOne)
		// TODO: we can enable this error checking once we fix tick initialization
		// if !ok {
		// 	return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("there are no more ticks initialized to fill the swap")
		// }
		nextSqrtPrice, err := k.tickToSqrtPrice(sdk.NewInt(nextTick))
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, err
		}

		// TODO: In and out get flipped based on if we are calculating for in or out, need to fix this
		sqrtPrice, amountIn, amountOut := computeSwapStep(
			swapState.sqrtPrice,
			nextSqrtPrice,
			swapState.liquidity,
			swapState.amountSpecifiedRemaining,
			zeroForOne,
		)

		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.Sub(amountIn)
		swapState.amountCalculated = swapState.amountCalculated.Add(amountOut.Quo(sdk.OneDec().Sub(swapFee)))

		if swapState.sqrtPrice.Equal(sqrtPrice) {
			liquidityDelta, err := k.crossTick(ctx, p.Id, nextTick)
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, err
			}
			if !zeroForOne {
				liquidityDelta = liquidityDelta.Neg()
			}
			swapState.liquidity = swapState.liquidity.Add(liquidityDelta.ToDec())
			if swapState.liquidity.LTE(sdk.ZeroDec()) || swapState.liquidity.IsNil() {
				return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("no liquidity available, cannot swap")
			}
			if !zeroForOne {
				swapState.tick = sdk.NewInt(nextTick - 1)
			} else {
				swapState.tick = sdk.NewInt(nextTick)
			}
		} else {
			swapState.tick = priceToTick(sqrtPrice.Power(2))
		}
	}
	return sdk.NewCoin(tokenInDenom, swapState.amountCalculated.RoundInt()), sdk.NewCoin(tokenOut.Denom, tokenOut.Amount), nil
}
