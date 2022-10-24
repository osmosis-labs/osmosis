package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// maintains the current swap state
type SwapState struct {
	amountSpecifiedRemaining sdk.Dec // remaining amount of tokens that need to be bought by the pool
	amountCalculated         sdk.Dec // amount out
	sqrtPrice                sdk.Dec // new current price when swap is done
	tick                     sdk.Int // new tick when swap is done
}

// tracks state of one iteration of an order filling
type StepState struct {
	sqrtPriceStart sdk.Dec // price iteration begins with
	nextTick       sdk.Int // next initialized tick that will provide liquidity for the swap
	sqrtPriceNext  sdk.Dec // price at the next tick
	amountIn       sdk.Dec
	amountOut      sdk.Dec
}

func (k Keeper) CreateNewConcentratedLiquidityPool(ctx sdk.Context, poolId uint64, denom0, denom1 string, currSqrtPrice, currTick sdk.Int) (Pool, error) {
	poolAddr := types.NewPoolAddress(poolId)
	denom0, denom1, err := k.orderInitialPoolDenoms(denom0, denom1)
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

// TODO: remove nolint
// nolint: unused
func priceToTick(price sdk.Dec) sdk.Int {
	logOfPrice := osmomath.BigDecFromSDKDec(price).ApproxLog2()
	logInt := osmomath.NewDecWithPrec(10001, 4)
	tick := logOfPrice.Quo(logInt.ApproxLog2())
	return tick.SDKDec().TruncateInt()
}

// TODO: remove nolint
// nolint: unused
func tickToPrice(tick sdk.Int) sdk.Dec {
	price := sdk.NewDecWithPrec(10001, 4).Power(tick.Uint64())
	return price
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

// this only works on a single directional trade, will implement bi directional trade in next milestone
func (k Keeper) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec, minPrice, maxPrice sdk.Dec, poolId uint64) (newTokenIn, tokenOut sdk.Coin, err error) {
	p := k.getPoolbyId(ctx, poolId)
	asset0 := p.Token0
	asset1 := p.Token1
	tokenAmountInAfterFee := tokenIn.Amount.ToDec().Mul(sdk.OneDec().Sub(swapFee))

	// get current sqrt price from pool
	curSqrtPrice := sdk.NewDecWithPrec(int64(p.CurrentSqrtPrice.Uint64()), 6)

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
	if minPrice.GTE(curSqrtPrice.Power(2)) {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("minPrice (%s) must be less than current price (%s)", minPrice, curSqrtPrice.Power(2))
	}
	if maxPrice.LTE(curSqrtPrice.Power(2)) {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("maxPrice (%s) must be greater than current price (%s)", maxPrice, curSqrtPrice.Power(2))
	}

	// sqrtPrice of upper and lower user defined price range
	sqrtPLowerTick, err := minPrice.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("issue calculating square root of minPrice")
	}
	sqrtPCurTick := curSqrtPrice
	sqrtPUpperTick, err := maxPrice.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("issue calculating square root of maxPrice")
	}

	// TODO: How do we remove/generalize this? I am stumped.
	amountETH := int64(1000000)
	amountUSDC := int64(5000000000)

	// find liquidity of assetA and assetB
	liq0 := liquidity0(amountETH, sqrtPCurTick, sqrtPUpperTick)
	liq1 := liquidity1(amountUSDC, sqrtPCurTick, sqrtPLowerTick)

	// utilize the smaller liquidity between assetA and assetB when performing the swap calculation
	liq := sdk.MinDec(liq0, liq1)

	swapState := SwapState{
		amountSpecifiedRemaining: tokenAmountInAfterFee,
		amountCalculated:         sdk.ZeroDec(),
		sqrtPrice:                sqrtPCurTick,
		tick:                     priceToTick(sqrtPCurTick.Power(2)),
	}
	// TODO: This should be GT 0 but some instances have very small remainder
	// need to look into fixing this
	for swapState.amountSpecifiedRemaining.GT(sdk.NewDecWithPrec(1, 6)) {
		stepState := StepState{}
		stepState.sqrtPriceStart = swapState.sqrtPrice
		lte := tokenIn.Denom == asset1
		nextTick, _ := k.NextInitializedTick(ctx, poolId, swapState.tick.Int64(), lte)
		stepState.nextTick = sdk.NewInt(nextTick)
		nextSqrtPrice, _ := k.tickToPrice(stepState.nextTick)
		stepState.sqrtPriceNext = nextSqrtPrice

		swapState.sqrtPrice, stepState.amountIn, stepState.amountOut = computeSwapStep(
			swapState.sqrtPrice,
			stepState.sqrtPriceNext,
			liq,
			swapState.amountSpecifiedRemaining,
			lte,
		)
		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.Sub(stepState.amountIn)
		swapState.amountCalculated = swapState.amountCalculated.Add(stepState.amountOut)
		swapState.tick = priceToTick(swapState.sqrtPrice.Power(2))
	}
	newTokenIn.Amount = tokenIn.Amount.Sub(swapState.amountSpecifiedRemaining.RoundInt())
	return sdk.NewCoin(tokenIn.Denom, newTokenIn.Amount), sdk.NewCoin(tokenOutDenom, swapState.amountCalculated.RoundInt()), nil
}

func (p Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (k Keeper) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec, minPrice, maxPrice sdk.Dec, poolId uint64) (tokenIn, newTokenOut sdk.Coin, err error) {
	tokenOutAmt := tokenOut.Amount.ToDec()
	p := k.getPoolbyId(ctx, poolId)
	asset0 := p.Token0
	asset1 := p.Token1

	// get current sqrt price from pool
	curSqrtPrice := sdk.NewDecWithPrec(int64(p.CurrentSqrtPrice.Uint64()), 6)

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
	if minPrice.GTE(curSqrtPrice.Power(2)) {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("minPrice (%s) must be less than current price (%s)", minPrice, curSqrtPrice.Power(2))
	}
	if maxPrice.LTE(curSqrtPrice.Power(2)) {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("maxPrice (%s) must be greater than current price (%s)", maxPrice, curSqrtPrice.Power(2))
	}

	// sqrtPrice of upper and lower user defined price range
	sqrtPLowerTick, err := minPrice.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("issue calculating square root of minPrice")
	}
	sqrtPCurTick := curSqrtPrice
	sqrtPUpperTick, err := maxPrice.ApproxSqrt()
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("issue calculating square root of maxPrice")
	}

	// TODO: How do we remove/generalize this? I am stumped.
	amountETH := int64(1000000)
	amountUSDC := int64(5000000000)

	// find liquidity of assetA and assetB
	liq0 := liquidity0(amountETH, sqrtPCurTick, sqrtPUpperTick)
	liq1 := liquidity1(amountUSDC, sqrtPCurTick, sqrtPLowerTick)

	// utilize the smaller liquidity between assetA and assetB when performing the swap calculation
	liq := sdk.MinDec(liq0, liq1)

	swapState := SwapState{
		amountSpecifiedRemaining: tokenOutAmt,
		amountCalculated:         sdk.ZeroDec(),
		sqrtPrice:                sqrtPCurTick,
		tick:                     priceToTick(sqrtPCurTick.Power(2)),
	}
	// TODO: This should be GT 0 but some instances have very small remainder
	// need to look into fixing this
	for swapState.amountSpecifiedRemaining.GT(sdk.NewDecWithPrec(1, 6)) {
		stepState := StepState{}
		stepState.sqrtPriceStart = swapState.sqrtPrice
		lte := tokenOut.Denom == asset1
		nextTick, _ := k.NextInitializedTick(ctx, poolId, swapState.tick.Int64(), lte)
		stepState.nextTick = sdk.NewInt(nextTick)
		nextSqrtPrice, _ := k.tickToPrice(stepState.nextTick)
		stepState.sqrtPriceNext = nextSqrtPrice
		// TODO: In and out get flipped based on if we are calculating for in or out, need to fix this
		swapState.sqrtPrice, stepState.amountIn, stepState.amountOut = computeSwapStep(
			swapState.sqrtPrice,
			stepState.sqrtPriceNext,
			liq,
			swapState.amountSpecifiedRemaining,
			lte,
		)
		stepState.amountOut = stepState.amountOut.Quo(sdk.OneDec().Sub(swapFee))
		swapState.amountSpecifiedRemaining = swapState.amountSpecifiedRemaining.Sub(stepState.amountIn)
		swapState.amountCalculated = swapState.amountCalculated.Add(stepState.amountOut)
		swapState.tick = priceToTick(swapState.sqrtPrice.Power(2))
	}
	return sdk.NewCoin(tokenInDenom, swapState.amountCalculated.RoundInt()), sdk.NewCoin(tokenOut.Denom, tokenOut.Amount), nil
}

func (k Keeper) orderInitialPoolDenoms(denom0, denom1 string) (string, string, error) {
	if denom0 == denom1 {
		return "", "", fmt.Errorf("cannot have the same asset in a single pool")
	}
	if denom0 > denom1 {
		denom1, denom0 = denom0, denom1
	}

	return denom0, denom1, nil
}
