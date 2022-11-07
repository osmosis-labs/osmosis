package stableswap

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/internal/cfmm_common"
	types "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

var (
	cubeRootTwo, _        = osmomath.NewBigDec(2).ApproxRoot(3)
	threeRootTwo, _       = osmomath.NewBigDec(3).ApproxRoot(2)
	cubeRootThree, _      = osmomath.NewBigDec(3).ApproxRoot(3)
	threeCubeRootTwo      = cubeRootTwo.MulInt64(3)
	cubeRootSixSquared, _ = (osmomath.NewBigDec(6).MulInt64(6)).ApproxRoot(3)
	twoCubeRootThree      = cubeRootThree.MulInt64(2)
	twentySevenRootTwo, _ = osmomath.NewBigDec(27).ApproxRoot(2)
)

// solidly CFMM is xy(x^2 + y^2) = k
func cfmmConstant(xReserve, yReserve osmomath.BigDec) osmomath.BigDec {
	if !xReserve.IsPositive() || !yReserve.IsPositive() {
		panic("invalid input: reserves must be positive")
	}
	xy := xReserve.Mul(yReserve)
	x2 := xReserve.Mul(xReserve)
	y2 := yReserve.Mul(yReserve)
	return xy.Mul(x2.Add(y2))
}

// multi-asset CFMM is xyv(x^2 + y^2 + w) = k,
// where u is the product of the reserves of assets
// outside of x and y (e.g. u = wz), and v is the sum
// of their squares (e.g. v = w^2 + z^2).
// When u = 1 and v = 0, this is equivalent to solidly's CFMM
// {TODO: Update this comment}
func cfmmConstantMultiNoV(xReserve, yReserve, vSumSquares osmomath.BigDec) osmomath.BigDec {
	if !xReserve.IsPositive() || !yReserve.IsPositive() || vSumSquares.IsNegative() {
		panic("invalid input: reserves must be positive")
	}

	xy := xReserve.Mul(yReserve)
	x2 := xReserve.Mul(xReserve)
	y2 := yReserve.Mul(yReserve)
	return xy.Mul(x2.Add(y2).Add(vSumSquares))
}

func cfmmConstantMulti(xReserve, yReserve, u, v osmomath.BigDec) osmomath.BigDec {
	if !u.IsPositive() {
		panic("invalid input: reserves must be positive")
	}
	return cfmmConstantMultiNoV(xReserve, yReserve, v).Mul(u)
}

// solidly CFMM is xy(x^2 + y^2) = k, and our multi-asset CFMM is xyz(x^2 + y^2 + w) = k
// So we want to solve for a given addition of `b` units of y into the pool,
// how many units `a` of x do we get out.
// So we solve the following expression for `a` in two-asset pools:
// xy(x^2 + y^2) = (x - a)(y + b)((x - a)^2 + (y + b)^2)
// and the following expression for `a` in multi-asset pools:
// xyz(x^2 + y^2 + w) = (x - a)(y + b)z((x - a)^2 + (y + b)^2 + w)
func solveCfmm(xReserve, yReserve osmomath.BigDec, remReserves []osmomath.BigDec, yIn osmomath.BigDec) osmomath.BigDec {
	wSumSquares := osmomath.ZeroDec()
	for _, assetReserve := range remReserves {
		wSumSquares = wSumSquares.Add(assetReserve.Mul(assetReserve))
	}
	return solveCFMMBinarySearchMulti(xReserve, yReserve, wSumSquares, yIn)
}

// solidly CFMM is xy(x^2 + y^2) = k
// So we want to solve for a given addition of `b` units of y into the pool,
// how many units `a` of x do we get out.
// Let y' = y + b
// we solve k = (x'y')(x'^2 + y^2) for x', using the following equation: https://www.wolframalpha.com/input?i2d=true&i=solve+for+y%5C%2844%29+x*y*%5C%2840%29Power%5Bx%2C2%5D%2BPower%5By%2C2%5D%5C%2841%29%3Dk
// which we simplify to be the following: https://www.desmos.com/calculator/bx5m5wpind
// Then we use that to derive the change in x as x_out = x' - x
//
// Since original reserves, y' and k are known and remain constant throughout the calculation,
// deriving x' and then finding x_out is equivalent to finding x_out directly.
func solveCfmmDirect(xReserve, yReserve, yIn osmomath.BigDec) osmomath.BigDec {
	if !xReserve.IsPositive() || !yReserve.IsPositive() || !yIn.IsPositive() {
		panic("invalid input: reserves and input must be positive")
	}

	if yIn.GT(yReserve) {
		panic("invalid input: cannot trade greater than reserve amount into CFMM")
	}

	// find k using existing reserves
	k := cfmmConstant(xReserve, yReserve)

	// find new yReserve after join
	y_new := yReserve.Add(yIn)

	// store powers to simplify calculations
	y2 := y_new.Mul(y_new)
	y3 := y2.Mul(y_new)
	y4 := y3.Mul(y_new)

	// We then solve for new xReserve using new yReserve and old k using a solver derived from xy(x^2 + y^2) = k
	// Full equation: x' = [((2^(1/3)) * ([y^2 * 9k) * ((sqrt(1 + ((2 / sqrt(27)) * (y^4 / k))^2)) + 1)]^(1/3)) / y')
	// 													 	- (2 * (3^(1/3)) * y^3 / ([y^2 * 9k) * ((sqrt(1 + ((2 / sqrt(27)) * (y^4 / k))^2)) + 1)]^(1/3)))
	// 						] / (6^(2/3))
	//
	// To simplify, we make the following abstractions:
	// 1. scaled_y4_quo_k = (2 / sqrt(27)) * (y^4 / k)
	// 2. sqrt_term = sqrt(1 + scaled_y4_quo_k2)
	// 3. common_factor = [y^2 * 9k) * (sqrt_term + 1)]^(1/3)
	// 4. term1 = (2^(1/3)) * common_factor / y'
	// 5. term2 = 2 * (3^(1/3)) * y^3 / common_factor
	//
	// With these, the final equation becomes: x' = (term1 - term2) / (6^(2/3))

	// let scaled_y4_quo_k = (2 / sqrt(27)) * (y^4 / k)
	scaled_y4_quo_k := (y4.Quo(k)).Mul(osmomath.NewBigDec(2).Quo(twentySevenRootTwo))
	scaled_y4_quo_k2 := scaled_y4_quo_k.Mul(scaled_y4_quo_k)

	// let sqrt_term = sqrt(1 + scaled_y4_quo_k2)
	sqrt_term, err := (osmomath.OneDec().Add(scaled_y4_quo_k2)).ApproxRoot(2)
	if err != nil {
		panic(err)
	}

	// let common_factor = [y^2 * 9k) * (sqrt_term + 1)]^(1/3)
	common_factor, err := (y2.MulInt64(9).Mul(k).Mul((sqrt_term.Add(osmomath.OneDec())))).ApproxRoot(3)
	if err != nil {
		panic(err)
	}

	// term1 = (2^(1/3)) * common_factor / y'
	term1 := cubeRootTwo.Mul(common_factor).Quo(y_new)
	// term2 = 2 * (3^(1/3)) * y^3 / common_factor
	term2 := twoCubeRootThree.Mul(y3).Quo(common_factor)

	// finally, x' = (term1 - term2) / (6^(2/3))
	x_new := (term1.Sub(term2)).Quo(cubeRootSixSquared)

	// find amount of x to output using initial and final xReserve values
	xOut := xReserve.Sub(x_new)

	if xOut.GTE(xReserve) {
		panic("invalid output: greater than full pool reserves")
	}

	return xOut
}

// multi-asset CFMM is xyu(x^2 + y^2 + w) = k
// As described in our spec, we can ignore the u term and simply solve within the bounds of k' = k / u
// since u remains constant throughout any independent operation this solver would be used for.
// We want to solve for a given addition of `b` units of y into the pool,
// how many units `a` of x do we get out.
// Let y' = y + b
// we solve k = (x'y')(x'^2 + y^2 + w) for x', using the following equation: https://www.wolframalpha.com/input?i2d=true&i=solve+for+y%5C%2844%29+x*y*%5C%2840%29Power%5Bx%2C2%5D+%2B+Power%5By%2C2%5D+%2B+w%5C%2841%29%3Dk
// which we simplify to be the following: https://www.desmos.com/calculator/zx2qslqndl
// Then we use that to derive the change in x as x_out = x' - x
//
// Since original reserves, y' and k are known and remain constant throughout the calculation,
// deriving x' and then finding x_out is equivalent to finding x_out directly.
func solveCFMMMultiDirect(xReserve, yReserve, wSumSquares, yIn osmomath.BigDec) osmomath.BigDec {
	if !xReserve.IsPositive() || !yReserve.IsPositive() || wSumSquares.IsNegative() || !yIn.IsPositive() {
		panic("invalid input: reserves and input must be positive")
	} else if yIn.GTE(yReserve) {
		panic("cannot input more than pool reserves")
	}

	// find k' using existing reserves (k' = k / v term)
	k := cfmmConstantMultiNoV(xReserve, yReserve, wSumSquares)
	k2 := k.Mul(k)

	// find new yReserve after join
	y_new := yReserve.Add(yIn)

	// store powers to simplify calculations
	y2 := y_new.Mul(y_new)
	y3 := y2.Mul(y_new)
	y4 := y3.Mul(y_new)

	// We then solve for new xReserve using new yReserve and old k using a solver derived from xy(x^2 + y^2 + w) = k
	// Full equation: x' = (sqrt(729 k^2 y^4 + 108 y^3 (w y + y^3)^3) + 27 k y^2)^(1/3) / (3 2^(1/3) y)
	// 								- (2^(1/3) (w y + y^3))/(sqrt(729 k^2 y^4 + 108 y^3 (w y + y^3)^3) + 27 k y^2)^(1/3)
	//
	//
	// To simplify, we make the following abstractions:
	// 1. sqrt_term = sqrt(729 k^2 y^4 + 108 y^3 (w y + y^3)^3)
	// 2. cube_root_term = (sqrt_term + 27 k y^2)^(1/3)
	// 3. term1 = cube_root_term / (3 2^(1/3) y)
	// 4. term2 = (2^(1/3) (w y + y^3)) / cube_root_term
	//
	// With these, the final equation becomes: x' = term1 - term2

	// let sqrt_term = sqrt(729 k^2 y^4 + 108 y^3 (w y + y^3)^3)
	wypy3 := (wSumSquares.Mul(y_new)).Add(y3)
	wypy3pow3 := wypy3.Mul(wypy3).Mul(wypy3)

	sqrt_term, err := ((k2.Mul(y4).MulInt64(729)).Add(y3.MulInt64(108).Mul(wypy3pow3))).ApproxRoot(2)
	if err != nil {
		panic(err)
	}

	// let cube_root_term = (sqrt_term + 27 k y^2)^(1/3)
	cube_root_term, err := (sqrt_term.Add(k.Mul(y2).MulInt64(27))).ApproxRoot(3)
	if err != nil {
		panic(err)
	}

	// let term1 = cube_root_term / (3 2^(1/3) y)
	term1 := cube_root_term.Quo(cubeRootTwo.MulInt64(3).Mul(y_new))

	// let term2 = cube_root_term * (2^(1/3) (w y + y^3))
	term2 := (cubeRootTwo.Mul(wypy3)).Quo(cube_root_term)

	// finally, let x' = term1 - term2
	x_new := term1.Sub(term2)

	// find amount of x to output using initial and final xReserve values
	xOut := xReserve.Sub(x_new)

	if xOut.GTE(xReserve) {
		panic("invalid output: greater than full pool reserves")
	}

	return xOut
}

func approxDecEqual(a, b, tol osmomath.BigDec) bool {
	return (a.Sub(b).Abs()).LTE(tol)
}

var (
	twodec      = osmomath.MustNewDecFromStr("2.0")
	k_threshold = osmomath.NewDecWithPrec(1, 1) // Correct within a factor of 1 * 10^{-1}
)

// solveCFMMBinarySearch searches the correct dx using binary search over constant K.
func solveCFMMBinarySearchMulti(xReserve, yReserve, wSumSquares, yIn osmomath.BigDec) osmomath.BigDec {
	if !xReserve.IsPositive() || !yReserve.IsPositive() || wSumSquares.IsNegative() {
		panic("invalid input: reserves and input must be positive")
	} else if yIn.Abs().GTE(yReserve) {
		panic("cannot input more than pool reserves")
	}
	yFinal := yReserve.Add(yIn)
	xLowEst, xHighEst := xReserve, xReserve
	k0 := cfmmConstantMultiNoV(xReserve, yFinal, wSumSquares)
	k := cfmmConstantMultiNoV(xReserve, yReserve, wSumSquares)
	if k0.Equal(osmomath.ZeroDec()) || k.Equal(osmomath.ZeroDec()) {
		panic("k should never be zero")
	}
	kRatio := k0.Quo(k)

	if kRatio.LT(osmomath.OneDec()) {
		// k_0 < k. Need to find an upperbound. Worst case assume a linear relationship, gives an upperbound
		// TODO: In the future, we can derive better bounds via reasoning about coefficients in the cubic
		// These are quite close when we are in the "stable" part of the curve though.
		xHighEst = xReserve.Quo(kRatio).Ceil()
	} else if kRatio.GT(osmomath.OneDec()) {
		// need to find a lowerbound. We could use a cubic relation, but for now we just set it to 0.
		xLowEst = osmomath.ZeroDec()
	} else {
		// k remains unchanged, so xOut = 0
		return osmomath.ZeroDec()
	}

	maxIterations := 256

	// we use a geometric error tolerance that guarantees approximately 10^-12 precision on outputs
	errTolerance := osmoutils.ErrTolerance{AdditiveTolerance: sdk.Int{}, MultiplicativeTolerance: sdk.NewDecWithPrec(1, 12)}

	// create single-input CFMM to pass into binary search
	computeFromEst := func(xEst osmomath.BigDec) (osmomath.BigDec, error) {
		return cfmmConstantMultiNoV(xEst, yFinal, wSumSquares), nil
	}

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

	xEst, err := osmoutils.BinarySearchBigDec(computeFromEst, xLowEst, xHighEst, k, errTolerance, maxIterations)
	if err != nil {
		panic(err)
	}

	xOut := xReserve.Sub(xEst)
	// We check the absolute value of the output against the xReserve amount to ensure that:
	// 1. Swaps cannot more than double the input token's pool supply
	// 2. Swaps cannot output more than the output token's pool supply
	if xOut.Abs().GTE(xReserve) {
		panic("invalid output: greater than full pool reserves")
	}
	return xOut
}

func (p Pool) spotPrice(baseDenom, quoteDenom string) (sdk.Dec, error) {
	roundMode := osmomath.RoundBankers // TODO:
	reserves, err := p.scaledSortedPoolReserves(baseDenom, quoteDenom, roundMode)
	if err != nil {
		return sdk.Dec{}, err
	}
	baseReserve, quoteReserve, remReserves := reserves[0], reserves[1], reserves[2:]
	// y = baseAsset, x = quoteAsset
	// Define f_{y -> x}(a) as the function that outputs the amount of tokens X you'd get by
	// trading "a" units of Y against the pool, assuming 0 swap fee, at the current liquidity.
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
	a := osmomath.OneDec()
	// no need to divide by a, since a = 1.
	bigDec := solveCfmm(baseReserve, quoteReserve, remReserves, a)
	return bigDec.SDKDec(), nil
}

func oneMinus(swapFee sdk.Dec) osmomath.BigDec {
	return osmomath.BigDecFromSDKDec(sdk.OneDec().Sub(swapFee))
}

// returns outAmt as a decimal
func (p Pool) calcOutAmtGivenIn(tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec) (sdk.Dec, error) {
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

	// amm input = tokenIn * (1 - swap fee)
	ammIn := tokenInDec.Mul(oneMinus(swapFee))
	// We are solving for the amount of token out, hence x = tokenOutSupply, y = tokenInSupply
	cfmmOut := solveCfmm(tokenOutSupply, tokenInSupply, remReserves, ammIn)
	outAmt := p.getDescaledPoolAmt(tokenOutDenom, cfmmOut)
	return outAmt, nil
}

// returns inAmt as a decimal
func (p *Pool) calcInAmtGivenOut(tokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec) (sdk.Dec, error) {
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
	// divide by (1 - swapfee) to force a corresponding increase in input asset
	inAmt := cfmmIn.QuoRoundUp(oneMinus(swapFee))
	inCoinAmt := p.getDescaledPoolAmt(tokenInDenom, inAmt)
	return inCoinAmt, nil
}

func (p *Pool) calcSingleAssetJoinShares(tokenIn sdk.Coin, swapFee sdk.Dec) (sdk.Int, error) {
	poolWithAddedLiquidityAndShares := func(newLiquidity sdk.Coin, newShares sdk.Int) types.PoolI {
		paCopy := p.Copy()
		paCopy.updatePoolForJoin(sdk.NewCoins(tokenIn), newShares)
		return &paCopy
	}

	// We apply the swap fee by multiplying by (1 - swapFee) and then truncating to int
	oneMinusSwapFee := sdk.OneDec().Sub(swapFee)
	tokenInAmtAfterFee := tokenIn.Amount.ToDec().Mul(oneMinusSwapFee).TruncateInt()

	return cfmm_common.BinarySearchSingleAssetJoin(p, sdk.NewCoin(tokenIn.Denom, tokenInAmtAfterFee), poolWithAddedLiquidityAndShares)
}

// We can mutate pa here
// TODO: some day switch this to a COW wrapped pa, for better perf
func (p *Pool) joinPoolSharesInternal(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	if !tokensIn.DenomsSubsetOf(p.GetTotalPoolLiquidity(ctx)) {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("attempted joining pool with assets that do not exist in pool")
	}
	if len(tokensIn) == 1 && tokensIn[0].Amount.GT(sdk.OneInt()) {
		numShares, err = p.calcSingleAssetJoinShares(tokensIn[0], swapFee)
		if err != nil {
			return sdk.ZeroInt(), sdk.NewCoins(), err
		}

		newLiquidity = tokensIn

		p.updatePoolForJoin(newLiquidity, numShares)

		if err = validatePoolLiquidity(p.PoolLiquidity, p.ScalingFactors); err != nil {
			return sdk.ZeroInt(), sdk.NewCoins(), err
		}

		return numShares, newLiquidity, err
	} else if len(tokensIn) != p.NumAssets() {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New(
			"stableswap pool only supports LP'ing with one asset, or all assets in pool")
	}

	// Add all exact coins we can (no swap). ctx arg doesn't matter for Stableswap
	numShares, remCoins, err := cfmm_common.MaximalExactRatioJoin(p, sdk.Context{}, tokensIn)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}
	p.updatePoolForJoin(tokensIn.Sub(remCoins), numShares)

	tokensJoined := tokensIn.Sub(remCoins)

	if err = validatePoolLiquidity(p.PoolLiquidity, p.ScalingFactors); err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}

	return numShares, tokensJoined, nil
}
