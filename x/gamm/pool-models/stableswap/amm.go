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

// multi-asset CFMM is xyu(x^2 + y^2 + v) = k,
// where u is the product of the reserves of assets
// outside of x and y (e.g. u = wz), and v is the sum
// of their squares (e.g. v = w^2 + z^2).
// When u = 1 and v = 0, this is equivalent to solidly's CFMM
func cfmmConstantMulti(xReserve, yReserve, uReserve, vSumSquares osmomath.BigDec) osmomath.BigDec {
	if !xReserve.IsPositive() || !yReserve.IsPositive() || !uReserve.IsPositive() || vSumSquares.IsNegative() {
		panic("invalid input: reserves must be positive")
	}

	xyu := xReserve.Mul(yReserve.Mul(uReserve))
	x2 := xReserve.Mul(xReserve)
	y2 := yReserve.Mul(yReserve)
	return xyu.Mul(x2.Add(y2).Add(vSumSquares))
}

// solidly CFMM is xy(x^2 + y^2) = k, and our multi-asset CFMM is xyz(x^2 + y^2 + w) = k
// So we want to solve for a given addition of `b` units of y into the pool,
// how many units `a` of x do we get out.
// So we solve the following expression for `a` in two-asset pools:
// xy(x^2 + y^2) = (x - a)(y + b)((x - a)^2 + (y + b)^2)
// and the following expression for `a` in multi-asset pools:
// xyz(x^2 + y^2 + w) = (x - a)(y + b)z((x - a)^2 + (y + b)^2 + w)
func solveCfmm(xReserve, yReserve osmomath.BigDec, remReserves []osmomath.BigDec, yIn osmomath.BigDec) osmomath.BigDec {
	if len(remReserves) == 0 {
		return solveCFMMBinarySearch(cfmmConstant)(xReserve, yReserve, yIn)
	}
	uReserve := osmomath.OneDec()
	wSumSquares := osmomath.ZeroDec()
	for _, assetReserve := range remReserves {
		uReserve = uReserve.Mul(assetReserve)
		wSumSquares = wSumSquares.Add(assetReserve.Mul(assetReserve))
	}
	return solveCFMMBinarySearchMulti(cfmmConstantMulti)(xReserve, yReserve, uReserve, wSumSquares, yIn)
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

func approxDecEqual(a, b, tol osmomath.BigDec) bool {
	return (a.Sub(b).Abs()).LTE(tol)
}

var (
	twodec      = osmomath.MustNewDecFromStr("2.0")
	k_threshold = osmomath.NewDecWithPrec(1, 1) // Correct within a factor of 1 * 10^{-1}
)

// solveCFMMBinarySearch searches the correct dx using binary search over constant K.
// added for future extension
func solveCFMMBinarySearch(constantFunction func(osmomath.BigDec, osmomath.BigDec) osmomath.BigDec) func(osmomath.BigDec, osmomath.BigDec, osmomath.BigDec) osmomath.BigDec {
	return func(xReserve, yReserve, yIn osmomath.BigDec) osmomath.BigDec {
		if !xReserve.IsPositive() || !yReserve.IsPositive() || !yIn.IsPositive() {
			panic("invalid input: reserves and input must be positive")
		} else if yIn.GTE(yReserve) {
			panic("cannot input more than pool reserves")
		}
		k := constantFunction(xReserve, yReserve)
		yFinal := yReserve.Add(yIn)
		xLowEst := osmomath.ZeroDec()
		xHighEst := xReserve
		maxIterations := 256
		errTolerance := osmoutils.ErrTolerance{AdditiveTolerance: sdk.OneInt(), MultiplicativeTolerance: sdk.Dec{}}

		// create single-input CFMM to pass into binary search
		calc_x_est := func(xEst osmomath.BigDec) (osmomath.BigDec, error) {
			return constantFunction(xEst, yFinal), nil
		}

		x_est, err := osmoutils.BinarySearchBigDec(calc_x_est, xLowEst, xHighEst, k, errTolerance, maxIterations)
		if err != nil {
			panic(err)
		}

		xOut := xReserve.Sub(x_est)
		if xOut.GTE(xReserve) {
			panic("invalid output: greater than full pool reserves")
		}
		return xOut
	}
}

// solveCFMMBinarySearch searches the correct dx using binary search over constant K.
// added for future extension
func solveCFMMBinarySearchMulti(constantFunction func(osmomath.BigDec, osmomath.BigDec, osmomath.BigDec, osmomath.BigDec) osmomath.BigDec) func(osmomath.BigDec, osmomath.BigDec, osmomath.BigDec, osmomath.BigDec, osmomath.BigDec) osmomath.BigDec {
	return func(xReserve, yReserve, uReserve, wSumSquares, yIn osmomath.BigDec) osmomath.BigDec {
		if !xReserve.IsPositive() || !yReserve.IsPositive() || !uReserve.IsPositive() || wSumSquares.IsNegative() || !yIn.IsPositive() {
			panic("invalid input: reserves and input must be positive")
		} else if yIn.GTE(yReserve) {
			panic("cannot input more than pool reserves")
		}
		k := constantFunction(xReserve, yReserve, uReserve, wSumSquares)
		yFinal := yReserve.Add(yIn)
		xLowEst := osmomath.ZeroDec()
		xHighEst := xReserve
		maxIterations := 256
		errTolerance := osmoutils.ErrTolerance{AdditiveTolerance: sdk.OneInt(), MultiplicativeTolerance: sdk.Dec{}}

		// create single-input CFMM to pass into binary search
		calcXEst := func(xEst osmomath.BigDec) (osmomath.BigDec, error) {
			return constantFunction(xEst, yFinal, uReserve, wSumSquares), nil
		}

		xEst, err := osmoutils.BinarySearchBigDec(calcXEst, xLowEst, xHighEst, k, errTolerance, maxIterations)
		if err != nil {
			panic(err)
		}

		xOut := xReserve.Sub(xEst)
		if xOut.GTE(xReserve) {
			panic("invalid output: greater than full pool reserves")
		}
		return xOut
	}
}

func spotPrice(baseReserve, quoteReserve osmomath.BigDec) osmomath.BigDec {
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
	return solveCfmm(baseReserve, quoteReserve, []osmomath.BigDec{}, a)
}

// returns outAmt as a decimal
func (p *Pool) calcOutAmtGivenIn(tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec) (sdk.Dec, error) {
	reserves, err := p.getScaledPoolAmts(tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	tokenInSupply, tokenOutSupply := reserves[0], reserves[1]
	remReserves := osmomath.BigDecFromSDKDecSlice(reserves[2:])
	// We are solving for the amount of token out, hence x = tokenOutSupply, y = tokenInSupply
	cfmmOut := solveCfmm(osmomath.BigDecFromSDKDec(tokenOutSupply), osmomath.BigDecFromSDKDec(tokenInSupply), remReserves, osmomath.BigDecFromSDKDec(tokenIn.Amount.ToDec()))
	outAmt := p.getDescaledPoolAmt(tokenOutDenom, cfmmOut)
	return outAmt.SDKDec(), nil
}

// returns inAmt as a decimal
func (p *Pool) calcInAmtGivenOut(tokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec) (sdk.Dec, error) {
	reserves, err := p.getScaledPoolAmts(tokenInDenom, tokenOut.Denom)
	if err != nil {
		return sdk.Dec{}, err
	}
	tokenInSupply, tokenOutSupply := reserves[0], reserves[1]
	remReserves := osmomath.BigDecFromSDKDecSlice(reserves[2:])
	// We are solving for the amount of token in, cfmm(x,y) = cfmm(x + x_in, y - y_out)
	// x = tokenInSupply, y = tokenOutSupply, yIn = -tokenOutAmount
	cfmmIn := solveCfmm(osmomath.BigDecFromSDKDec(tokenInSupply), osmomath.BigDecFromSDKDec(tokenOutSupply), remReserves, osmomath.BigDecFromSDKDec(tokenOut.Amount.ToDec().Neg()))
	inAmt := p.getDescaledPoolAmt(tokenInDenom, cfmmIn.Neg())
	return inAmt.SDKDec(), nil
}

func (p *Pool) calcSingleAssetJoinShares(tokenIn sdk.Coin, swapFee sdk.Dec) (sdk.Int, error) {
	poolWithAddedLiquidityAndShares := func(newLiquidity sdk.Coin, newShares sdk.Int) types.PoolI {
		paCopy := p.Copy()
		paCopy.updatePoolForJoin(sdk.NewCoins(tokenIn), newShares)
		return &paCopy
	}
	// TODO: Correctly handle swap fee
	return cfmm_common.BinarySearchSingleAssetJoin(p, tokenIn, poolWithAddedLiquidityAndShares)
}

// We can mutate pa here
// TODO: some day switch this to a COW wrapped pa, for better perf
func (p *Pool) joinPoolSharesInternal(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	if len(tokensIn) == 1 {
		numShares, err = p.calcSingleAssetJoinShares(tokensIn[0], swapFee)
		newLiquidity = tokensIn
		return numShares, newLiquidity, err
	} else if len(tokensIn) != p.NumAssets() || !tokensIn.DenomsSubsetOf(p.GetTotalPoolLiquidity(ctx)) {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New(
			"stableswap pool only supports LP'ing with one asset, or all assets in pool")
	}

	// Add all exact coins we can (no swap). ctx arg doesn't matter for Stableswap
	numShares, remCoins, err := cfmm_common.MaximalExactRatioJoin(p, sdk.Context{}, tokensIn)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}
	p.updatePoolForJoin(tokensIn.Sub(remCoins), numShares)

	for _, coin := range remCoins {
		// TODO: Perhaps add a method to skip if this is too small.
		newShare, err := p.calcSingleAssetJoinShares(coin, swapFee)
		if err != nil {
			return sdk.ZeroInt(), sdk.NewCoins(), err
		}
		p.updatePoolForJoin(sdk.NewCoins(coin), newShare)
		numShares = numShares.Add(newShare)
	}

	return numShares, tokensIn, nil
}
