package stableswap

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/internal/cfmm_common"
	types "github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

// Simplified multi-asset CFMM is xy(x^2 + y^2 + w) = k,
// where w is the sum of the squares of the
// reserve assets (e.g. w = m^2 + n^2).
// When w = 0, this is equivalent to solidly's CFMM
// We use this version for calculations since the u
// term in the full CFMM is constant.
func cfmmConstantMultiNoV(xReserve, yReserve, wSumSquares osmomath.BigDec) osmomath.BigDec {
	return cfmmConstantMultiNoVY(xReserve, yReserve, wSumSquares).Mul(yReserve)
}

// returns x(x^2 + y^2 + w) = k
// For use in comparing values with the same y
func cfmmConstantMultiNoVY(xReserve, yReserve, wSumSquares osmomath.BigDec) osmomath.BigDec {
	if !xReserve.IsPositive() || !yReserve.IsPositive() || wSumSquares.IsNegative() {
		panic("invalid input: reserves must be positive")
	}

	x2 := xReserve.Mul(xReserve)
	y2 := yReserve.Mul(yReserve)
	return xReserve.Mul(x2.Add(y2).Add(wSumSquares))
}

// Solidly's CFMM is xy(x^2 + y^2) = k, and our multi-asset CFMM is xyz(x^2 + y^2 + w) = k
// So we want to solve for a given addition of `b` units of y into the pool,
// how many units `a` of x do we get out.
// So we solve the following expression for `a`:
// xy(x^2 + y^2 + w) = (x - a)(y + b)((x - a)^2 + (y + b)^2 + w)
// with w set to 0 for 2 asset pools
func solveCfmm(xReserve, yReserve osmomath.BigDec, remReserves []osmomath.BigDec, yIn osmomath.BigDec) osmomath.BigDec {
	wSumSquares := osmomath.ZeroDec()
	for _, assetReserve := range remReserves {
		wSumSquares = wSumSquares.Add(assetReserve.Mul(assetReserve))
	}
	return solveCFMMBinarySearchMulti(xReserve, yReserve, wSumSquares, yIn)
}

// $$k_{target} = \frac{x_0 y_0 (x_0^2 + y_0^2 + w)}{y_f} - (x_0 (y_f^2 + w) + x_0^3)$$
func targetKCalculator(x0, y0, w, yf osmomath.BigDec) osmomath.BigDec {
	// cfmmNoV(x0, y0, w) = x_0 y_0 (x_0^2 + y_0^2 + w)
	startK := cfmmConstantMultiNoV(x0, y0, w)
	// remove extra yf term
	yfRemoved := startK.Quo(yf)
	// removed constant term from expression
	// namely - (x_0 (y_f^2 + w) + x_0^3) = x_0(y_f^2 + w + x_0^2)
	// innerTerm = y_f^2 + w + x_0^2
	innerTerm := yf.Mul(yf).Add(w).Add((x0.Mul(x0)))
	constantTerm := innerTerm.Mul(x0)
	return yfRemoved.Sub(constantTerm)
}

// $$k_{iter}(x_f) = -x_{out}^3 + 3 x_0 x_{out}^2 - (y_f^2 + w + 3x_0^2)x_{out}$$
// where x_out = x_0 - x_f
func iterKCalculator(x0, w, yf osmomath.BigDec) func(osmomath.BigDec) osmomath.BigDec {
	// compute coefficients first
	cubicCoeff := osmomath.OneDec().Neg()
	quadraticCoeff := x0.MulInt64(3)
	linearCoeff := quadraticCoeff.Mul(x0).Add(w).Add(yf.Mul(yf)).Neg()
	return func(xf osmomath.BigDec) osmomath.BigDec {
		xOut := x0.Sub(xf)
		// horners method
		// ax^3 + bx^2 + cx = x(c + x(b + ax))
		res := cubicCoeff.Mul(xOut)
		res = res.Add(quadraticCoeff).Mul(xOut)
		res = res.Add(linearCoeff).Mul(xOut)
		return res
	}
}

var (
	zero = osmomath.ZeroDec()
	one  = osmomath.OneDec()
)

func deriveUpperLowerXFinalReserveBounds(xReserve, yReserve, wSumSquares, yFinal osmomath.BigDec) (
	xFinalLowerbound, xFinalUpperbound osmomath.BigDec,
) {
	xFinalLowerbound, xFinalUpperbound = xReserve, xReserve

	k0 := cfmmConstantMultiNoV(xReserve, yFinal, wSumSquares)
	k := cfmmConstantMultiNoV(xReserve, yReserve, wSumSquares)
	// fmt.Println(k0, k)
	if k0.Equal(zero) || k.Equal(zero) {
		panic("k should never be zero")
	}
	kRatio := k0.Quo(k)
	if kRatio.LT(one) {
		// k_0 < k. Need to find an upperbound. Worst case assume a linear relationship, gives an upperbound
		// TODO: In the future, we can derive better bounds via reasoning about coefficients in the cubic
		// These are quite close when we are in the "stable" part of the curve though.
		xFinalUpperbound = xReserve.Quo(kRatio).Ceil()
	} else if kRatio.GT(one) {
		// need to find a lowerbound. We could use a cubic relation, but for now we just set it to 0.
		xFinalLowerbound = osmomath.ZeroDec()
	}
	// else
	// k remains unchanged.
	// So we keep bounds equal to each other
	return xFinalLowerbound, xFinalUpperbound
}

// solveCFMMBinarySearch searches the correct dx using binary search over constant K.
func solveCFMMBinarySearchMulti(xReserve, yReserve, wSumSquares, yIn osmomath.BigDec) osmomath.BigDec {
	if !xReserve.IsPositive() || !yReserve.IsPositive() || wSumSquares.IsNegative() {
		panic("invalid input: reserves and input must be positive")
	} else if yIn.Abs().GTE(yReserve) {
		panic("cannot input more than pool reserves")
	}
	// fmt.Printf("solve cfmm xreserve %v, yreserve %v, w %v, yin %v\n", xReserve, yReserve, wSumSquares, yIn)
	yFinal := yReserve.Add(yIn)
	xLowEst, xHighEst := deriveUpperLowerXFinalReserveBounds(xReserve, yReserve, wSumSquares, yFinal)
	targetK := targetKCalculator(xReserve, yReserve, wSumSquares, yFinal)
	iterKCalc := iterKCalculator(xReserve, wSumSquares, yFinal)
	maxIterations := 256

	// we use a geometric error tolerance that guarantees approximately 10^-12 precision on outputs
	errTolerance := osmomath.ErrTolerance{AdditiveTolerance: sdk.Dec{}, MultiplicativeTolerance: sdk.NewDecWithPrec(1, 12)}

	// if yIn is positive, we want to under-estimate the amount of xOut.
	// This means, we want x_out to be rounded down, as x_out = x_init - x_final, for x_init > x_final.
	// Thus we round-up x_final, to make it greater (and therefore ) x_out smaller.
	// If yIn is negative, the amount of xOut will also be negative (representing that we must add tokens into the pool)
	// this means x_out = x_init - x_final, for x_init < x_final.
	// we want to over_estimate |x_out|, which means rounding x_out down as its a negative quantity.
	// This means rounding x_final up, to give us a larger negative.
	// Therefore we always round up.
	roundingDirection := osmomath.RoundUp
	errTolerance.RoundingDir = roundingDirection

	xEst, err := osmomath.BinarySearchBigDec(iterKCalc, xLowEst, xHighEst, targetK, errTolerance, maxIterations)
	if err != nil {
		panic(err)
	}

	xOut := xReserve.Sub(xEst)
	// fmt.Printf("xOut %v\n", xOut)

	// We check the absolute value of the output against the xReserve amount to ensure that:
	// 1. Swaps cannot more than double the input token's pool supply
	// 2. Swaps cannot output more than the output token's pool supply
	if xOut.Abs().GTE(xReserve) {
		panic("invalid output: greater than full pool reserves")
	}
	return xOut
}

func (p Pool) spotPrice(quoteDenom, baseDenom string) (spotPrice sdk.Dec, err error) {
	// Define f_{y -> x}(a) as the function that outputs the amount of tokens X you'd get by
	// trading "a" units of Y against the pool, assuming 0 spread factor, at the current liquidity.
	// The spot price of the pool is then lim a -> 0, f_{y -> x}(a) / a
	// For uniswap f_{y -> x}(a) = x - xy/(y + a),
	// The spot price equation of y in terms of x is X_SUPPLY/Y_SUPPLY.
	// You can work out that it follows from the above relation!
	//
	// Now we have to work this out for the much more complex CFMM xy(x^2 + y^2).
	// Or we can sidestep this, by just picking a small value a, and computing f_{y -> x}(a) / a,
	// and accept the precision error.

	// We arbitrarily choose a = 1, and anticipate that this is a small value at the scale of
	// xReserve & yReserve.
	a := sdk.OneInt()

	res, err := p.calcOutAmtGivenIn(sdk.NewCoin(baseDenom, a), quoteDenom, sdk.ZeroDec())
	// fmt.Println("spot price res", res)
	return res, err
}

func oneMinus(spreadFactor sdk.Dec) osmomath.BigDec {
	return osmomath.BigDecFromSDKDec(sdk.OneDec().Sub(spreadFactor))
}

// calcOutAmtGivenIn calculate amount of specified denom to output from a pool in sdk.Dec given the input `tokenIn`
func (p Pool) calcOutAmtGivenIn(tokenIn sdk.Coin, tokenOutDenom string, spreadFactor sdk.Dec) (sdk.Dec, error) {
	// round liquidity down, and round token in down
	reserves, err := p.scaledSortedPoolReserves(tokenIn.Denom, tokenOutDenom, osmomath.RoundDown)
	if err != nil {
		return sdk.Dec{}, err
	}
	tokenInSupply, tokenOutSupply, remReserves := reserves[0], reserves[1], reserves[2:]
	tokenInDec, err := p.scaleCoin(tokenIn, osmomath.RoundDown)
	if err != nil {
		return sdk.Dec{}, err
	}

	// amm input = tokenIn * (1 - spread factor)
	ammIn := tokenInDec.Mul(oneMinus(spreadFactor))
	// We are solving for the amount of token out, hence x = tokenOutSupply, y = tokenInSupply
	// fmt.Printf("outSupply %s, inSupply %s, remReservs %s, ammIn %s\n ", tokenOutSupply, tokenInSupply, remReserves, ammIn)
	cfmmOut := solveCfmm(tokenOutSupply, tokenInSupply, remReserves, ammIn)
	// fmt.Println("cfmmout ", cfmmOut)
	outAmt := p.getDescaledPoolAmt(tokenOutDenom, cfmmOut)
	return outAmt, nil
}

// calcInAmtGivenOut calculates exact input amount given the desired output and return as a decimal
func (p *Pool) calcInAmtGivenOut(tokenOut sdk.Coin, tokenInDenom string, spreadFactor sdk.Dec) (sdk.Dec, error) {
	// round liquidity down, and round token out up
	reserves, err := p.scaledSortedPoolReserves(tokenInDenom, tokenOut.Denom, osmomath.RoundDown)
	if err != nil {
		return sdk.Dec{}, err
	}
	tokenInSupply, tokenOutSupply, remReserves := reserves[0], reserves[1], reserves[2:]
	tokenOutAmount, err := p.scaleCoin(tokenOut, osmomath.RoundUp)
	if err != nil {
		return sdk.Dec{}, err
	}

	// We are solving for the amount of token in, cfmm(x,y) = cfmm(x + x_in, y - y_out)
	// x = tokenInSupply, y = tokenOutSupply, yIn = -tokenOutAmount
	cfmmIn := solveCfmm(tokenInSupply, tokenOutSupply, remReserves, tokenOutAmount.Neg())
	// returned cfmmIn is negative, representing we need to add this many tokens to pool.
	// We invert that negative here.
	cfmmIn = cfmmIn.Neg()
	// divide by (1 - spread factor) to force a corresponding increase in input asset
	inAmt := cfmmIn.QuoRoundUp(oneMinus(spreadFactor))
	inCoinAmt := p.getDescaledPoolAmt(tokenInDenom, inAmt)
	return inCoinAmt, nil
}

// calcSingleAssetJoinShares calculates the number of LP shares that
// should be granted given the passed in single-token input (non-mutative)
func (p *Pool) calcSingleAssetJoinShares(tokenIn sdk.Coin, spreadFactor sdk.Dec) (sdk.Int, error) {
	poolWithAddedLiquidityAndShares := func(newLiquidity sdk.Coin, newShares sdk.Int) types.CFMMPoolI {
		paCopy := p.Copy()
		paCopy.updatePoolForJoin(sdk.NewCoins(newLiquidity), newShares)
		return &paCopy
	}

	// We apply the spread factor by multiplying by:
	// 1) getting what % of the input the spread factor should apply to
	// 2) multiplying that by spread factor
	// 3) oneMinusSpreadFactor := (1 - spread_factor * spread_factor_applicable_percent)
	// 4) Multiplying token in by one minus spread factor.
	spreadFactorApplicableRatio, err := p.singleAssetJoinSpreadFactorRatio(tokenIn.Denom)
	if err != nil {
		return sdk.Int{}, err
	}
	oneMinusSpreadFactor := sdk.OneDec().Sub(spreadFactor.Mul(spreadFactorApplicableRatio))
	tokenInAmtAfterFee := tokenIn.Amount.ToDec().Mul(oneMinusSpreadFactor).TruncateInt()

	return cfmm_common.BinarySearchSingleAssetJoin(p, sdk.NewCoin(tokenIn.Denom, tokenInAmtAfterFee), poolWithAddedLiquidityAndShares)
}

// returns the ratio of input asset liquidity, to total liquidity in pool, post-scaling.
// We use this as the portion of input liquidity to apply a spread factor too, for single asset joins.
// So if a pool is currently comprised of 80% of asset A, and 20% of asset B (post-scaling),
// and we input asset A, this function will return 20%.
// Note that this will over-estimate spread factor for single asset joins slightly,
// as in the swapping process into the pool, the A to B ratio would decrease the relative supply of B.
func (p *Pool) singleAssetJoinSpreadFactorRatio(tokenInDenom string) (sdk.Dec, error) {
	// get a second denom in pool
	tokenOut := p.PoolLiquidity[0]
	if tokenOut.Denom == tokenInDenom {
		tokenOut = p.PoolLiquidity[1]
	}
	// We round bankers scaled liquidity, since we care about the ratio of liquidity.
	scaledLiquidity, err := p.scaledSortedPoolReserves(tokenInDenom, tokenOut.Denom, osmomath.RoundDown)
	if err != nil {
		return sdk.Dec{}, err
	}

	totalLiquidityDenominator := osmomath.ZeroDec()
	for _, amount := range scaledLiquidity {
		totalLiquidityDenominator = totalLiquidityDenominator.Add(amount)
	}
	ratioOfInputAssetLiquidityToTotalLiquidity := scaledLiquidity[0].Quo(totalLiquidityDenominator)
	// SDKDec() rounds down (as it truncates), therefore 1 - term is rounded up, as desired.
	nonInternalAssetRatio := sdk.OneDec().Sub(ratioOfInputAssetLiquidityToTotalLiquidity.SDKDec())
	return nonInternalAssetRatio, nil
}

// Route a pool join attempt to either a single-asset join or all-asset join (mutates pool state)
// Eventually, we intend to switch this to a COW wrapped pa for better performance
func (p *Pool) joinPoolSharesInternal(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor sdk.Dec) (numShares sdk.Int, tokensJoined sdk.Coins, err error) {
	if !tokensIn.DenomsSubsetOf(p.GetTotalPoolLiquidity(ctx)) {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("attempted joining pool with assets that do not exist in pool")
	}

	if len(tokensIn) == 1 && tokensIn[0].Amount.GT(sdk.OneInt()) {
		numShares, err = p.calcSingleAssetJoinShares(tokensIn[0], spreadFactor)
		if err != nil {
			return sdk.ZeroInt(), sdk.NewCoins(), err
		}

		tokensJoined = tokensIn
	} else if len(tokensIn) != p.NumAssets() {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New(
			"stableswap pool only supports LP'ing with one asset, or all assets in pool")
	} else {
		// Add all exact coins we can (no swap). ctx arg doesn't matter for Stableswap
		var remCoins sdk.Coins
		numShares, remCoins, err = cfmm_common.MaximalExactRatioJoin(p, sdk.Context{}, tokensIn)
		if err != nil {
			return sdk.ZeroInt(), sdk.NewCoins(), err
		}

		tokensJoined = tokensIn.Sub(remCoins)
	}

	p.updatePoolForJoin(tokensJoined, numShares)

	if err = validatePoolLiquidity(p.PoolLiquidity, p.ScalingFactors); err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}

	return numShares, tokensJoined, nil
}
