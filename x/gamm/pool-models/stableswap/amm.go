package stableswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	cubeRootTwo, _   = sdk.NewDec(2).ApproxRoot(3)
	threeCubeRootTwo = cubeRootTwo.MulInt64(3)
)

// solidly CFMM is xy(x^2 + y^2) = k
func cfmmConstant(xReserve, yReserve sdk.Dec) sdk.Dec {
	xy := xReserve.Mul(yReserve)
	x2 := xReserve.Mul(xReserve)
	y2 := yReserve.Mul(yReserve)
	return xy.Mul(x2.Add(y2))
}

// solidly CFMM is xy(x^2 + y^2) = k
// So we want to solve for a given addition of `b` units of y into the pool,
// how many units `a` of x do we get out.
// So we solve the following expression for `a`
// xy(x^2 + y^2) = (x - a)(y + b)((x - a)^2 + (y + b)^2)
func solveCfmmDirectEquation(xReserve, yReserve, yIn sdk.Dec) sdk.Dec {
	if !yReserve.Add(yIn).IsPositive() {
		panic("invalid yReserve, yIn combo")
	}

	// use the following wolfram alpha link to solve the equation
	// https://www.wolframalpha.com/input?i=solve+for+a%2C+xy%28x%5E2+%2B+y%5E2%29+%3D+%28x+-+a%29%28y+%2B+b%29%28%28x+-+a%29%5E2+%2B+%28y+%2Bb%29%5E2%29+
	// This returns (copied from wolfram):
	// assuming (correctly) that b + y!=0
	// a = (-27 b^2 x^3 y - 27 b^2 x y^3 + sqrt((-27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^2 + 4 (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4)^3) - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^(1/3)/(3 2^(1/3) (b + y)) - (2^(1/3) (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4))/(3 (b + y) (-27 b^2 x^3 y - 27 b^2 x y^3 + sqrt((-27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^2 + 4 (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4)^3) - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^(1/3)) + (b x + x y)/(b + y) and b + y!=0
	// We simplify and separate out terms to get that its the following:
	// The key substitutions are that 3(b+y)^4 = 3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4
	// and -27 x y (b + y)^2 (x^2 + y^2) = -27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5
	// I added {} myself, making better distinctions between entirely distinct terms.
	// a = {(-27 x y (b + y)^2 (x^2 + y^2)
	//			+ sqrt(
	//				(-27 x y (b + y)^2 (x^2 + y^2))^2
	//				+ 108 ((b+y)^4)^3
	// 			)^(1/3)
	// 		  / (3 2^(1/3) (b + y))}
	//		- {(2^(1/3) (3 (b + y)^4))
	//		  /(3 (b + y)
	// 			(-27 x y (b + y)^2 (x^2 + y^2)
	// 				+ sqrt(
	//					(-27 x y (b + y)^2 (x^2 + y^2))^2
	//					+ 108 ((b+y)^4)^3)
	// 			)^(1/3))}
	//      + {(b x + x y)/(b + y)}
	// we further simplify, and call:
	// foo = (-27 x y (b + y)^2 (x^2 + y^2)
	// 			+ sqrt(
	//				(-27 x y (b + y)^2 (x^2 + y^2))^2
	//				+ 108 ((b+y)^4)^3)
	//		 )^(1/3)
	// Thus, a is then:
	// a = {foo / (3 2^(1/3) (b + y))}
	//		- {(3 * 2^(1/3) (b+y)^4)
	//		  /(3 (b + y) foo)}
	//      + {(b x + x y)/(b + y)}
	// Let:
	// term1 = {foo / (3 2^(1/3) (b + y))}
	// term2 = {(3 * 2^(1/3) (b+y)^4) /(3 (b + y) foo)} =  2^(1/3) (b+y)^3 / foo
	// term3 = {(b x + x y)/(b + y)}

	// prelude, compute all the xy cross terms. Consider keeping these precomputed in the struct,
	// and maybe in state.
	x := xReserve
	y := yReserve
	x2py2 := x.Mul(x).AddMut(y.Mul(y))

	xy := x.Mul(y)

	b := yIn

	bpy := b.Add(y)
	bpy2 := bpy.Mul(bpy)
	bpy3 := bpy2.Mul(bpy)
	bpy4 := bpy2.Mul(bpy2)

	// TODO: Come back and optimize alot of the calculations

	// Now we compute foo
	// foo = (-27 x y (b + y)^2 (x^2 + y^2)
	// 			+ sqrt(
	//				(-27 x y (b + y)^2 (x^2 + y^2))^2
	//				+ 108 ((b+y)^4)^3)
	//		 )^(1/3)
	// This has a y^12 term in it, which is unappealing, so we spend some energy reducing this max bitlen.
	// foo = (-27 x y (b + y)^2 (x^2 + y^2)
	// 			+ (b + y)^2 sqrt(
	//				729 (x y (x^2 + y^2))^2
	//				+ 108 (b+y)^8)
	//		 )^(1/3)
	// let e = x y (x^2 + y^2))
	// foo = (-27 (b + y)^2 e
	// 			+ (b + y)^2 sqrt(
	//				729 e^2 + 108 (b+y)^8)
	//		 )^(1/3)

	e := xy.Mul(x2py2) // xy(x^2 + y^2)

	// t1 = -27 (b + y)^2 e
	t1 := e.Mul(bpy2).MulInt64Mut(-27)

	// compute d = (b + y)^2 sqrt(729 e^2 + 108 (b+y)^8)
	bpy8 := bpy4.Mul(bpy4)
	sqrt_inner := e.MulMut(e).MulInt64Mut(729).AddMut(bpy8.MulInt64Mut(108)) // 729 e^2 + 108 (b+y)^8
	sqrt, err := sqrt_inner.ApproxSqrt()
	if err != nil {
		panic(err)
	}
	d := sqrt.MulMut(bpy2)

	// foo = (t1 + d)^(1/3)
	foo3 := t1.AddMut(d)
	foo, _ := foo3.ApproxRoot(3)

	// a = {foo / (3 2^(1/3) (b + y))}
	//		- {(2^(1/3) banana) / (3 (b + y) foo}
	//      + {(b x + x y)/(b + y)}

	// term1 := {foo / (3 2^(1/3) (b + y))}
	term1Denominator := threeCubeRootTwo.Mul(bpy)
	term1 := foo.Quo(term1Denominator)
	// term2 := {(2^(1/3) (b+y)^3) / (foo}
	term2 := cubeRootTwo.Mul(bpy3)
	term2 = term2.Quo(foo)
	// term3 := {(b x + x y)/(b + y)}
	term3Numerator := b.Mul(x).Add(xy)
	term3 := term3Numerator.Quo(bpy)

	a := term1.Sub(term2).Add(term3)
	return a
}

// (a - b) must be within tolerance
func approxDecEqual(a, b, tol sdk.Dec) bool {
	diff := a.Sub(b).Abs()
	return diff.LTE(tol)
}

var (
	twodec = sdk.MustNewDecFromStr("2.0")
	// TODO: Come back and analyze this threshold
	threshold = sdk.MustNewDecFromStr("0.00001")
)

// solveCFMMBinarySearch searches the correct dx using binary search over constant K.
// added for future extension
func solveCFMMBinarySearch(constantFunction func(sdk.Dec, sdk.Dec) sdk.Dec) func(sdk.Dec, sdk.Dec, sdk.Dec) sdk.Dec {
	return func(xReserve, yReserve, yIn sdk.Dec) sdk.Dec {
		if !yReserve.Add(yIn).IsPositive() {
			panic("invalid yReserve, yIn combo")
		}

		k := constantFunction(xReserve, yReserve)
		yf := yReserve.Add(yIn)
		x_low_est := sdk.ZeroDec()
		x_high_est := xReserve
		x_est := (x_high_est.Add(x_low_est)).Quo(twodec)
		cur_k := constantFunction(x_est, yf)
		for !approxDecEqual(cur_k, k, threshold) { // cap max iteration to 256
			if cur_k.GT(k) {
				x_high_est = x_est
			} else if cur_k.LT(k) {
				x_low_est = x_est
			}
			x_est = (x_high_est.Add(x_low_est)).Quo(twodec)
			cur_k = constantFunction(x_est, yf)
		}
		return xReserve.Sub(x_est)
	}
}

func solveCfmm(xReserve, yReserve, yIn sdk.Dec) sdk.Dec {
	return solveCFMMBinarySearch(cfmmConstant)(xReserve, yReserve, yIn)
}

//nolint:unused
func spotPrice(baseReserve, quoteReserve sdk.Dec) sdk.Dec {
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
	a := sdk.OneDec()
	// no need to divide by a, since a = 1.
	return solveCfmm(baseReserve, quoteReserve, a)
}

// returns outAmt as a decimal
func (pa *Pool) calcOutAmtGivenIn(tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec) (sdk.Dec, error) {
	reserves, err := pa.getPoolAmts(tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	tokenInSupply := reserves[0].ToDec()
	tokenOutSupply := reserves[1].ToDec()
	// We are solving for the amount of token out, hence x = tokenOutSupply, y = tokenInSupply
	outAmt := solveCfmm(tokenOutSupply, tokenInSupply, tokenIn.Amount.ToDec())
	return outAmt, nil
}

// returns inAmt as a decimal
func (pa *Pool) calcInAmtGivenOut(tokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec) (sdk.Dec, error) {
	reserves, err := pa.getPoolAmts(tokenInDenom, tokenOut.Denom)
	if err != nil {
		return sdk.Dec{}, err
	}
	tokenInSupply := reserves[0].ToDec()
	tokenOutSupply := reserves[1].ToDec()
	// We are solving for the amount of token in, cfmm(x,y) = cfmm(x + x_in, y - y_out)
	// x = tokenInSupply, y = tokenOutSupply, yIn = -tokenOutAmount
	inAmtRaw := solveCfmm(tokenInSupply, tokenOutSupply, tokenOut.Amount.ToDec().Neg())
	inAmt := inAmtRaw.NegMut()
	return inAmt, nil
}
