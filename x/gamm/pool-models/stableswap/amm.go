package stableswap

import sdk "github.com/cosmos/cosmos-sdk/types"

var cubeRootTwo, _ = sdk.NewDec(2).ApproxRoot(3)
var threeCubeRootTwo = cubeRootTwo.MulInt64(3)

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
// use the following wolfram alpha link
// https://www.wolframalpha.com/input?i=solve+for+a%2C+xy%28x%5E2+%2B+y%5E2%29+%3D+%28x+-+a%29%28y+%2B+b%29%28%28x+-+a%29%5E2+%2B+%28y+%2Bb%29%5E2%29+
// This returns:
// copied from wolfram:
// assuming (correctly) that b + y!=0
// a = (-27 b^2 x^3 y - 27 b^2 x y^3 + sqrt((-27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^2 + 4 (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4)^3) - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^(1/3)/(3 2^(1/3) (b + y)) - (2^(1/3) (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4))/(3 (b + y) (-27 b^2 x^3 y - 27 b^2 x y^3 + sqrt((-27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^2 + 4 (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4)^3) - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^(1/3)) + (b x + x y)/(b + y) and b + y!=0
// Dev separating out terms:
// I added {} myself, making better distinctions between entirely distinct terms. (parenthesis don't solve this, because divisions)
// a = {(-27 b^2 x^3 y - 27 b^2 x y^3
//			+ sqrt(
//				(-27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^2
//				+ 4 (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4)^3)
// 			- 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^(1/3)
// 		  / (3 2^(1/3) (b + y))}
//		- {(2^(1/3) (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4))
//		  /(3 (b + y)
// 			(-27 b^2 x^3 y - 27 b^2 x y^3
// 				+ sqrt(
//					(-27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^2
//					+ 4 (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4)^3)
// 				- 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^(1/3))}
//      + {(b x + x y)/(b + y)}
// Then notice that the two sqrt terms and surrounding items in the expression are the same.
// So we replace them with a single term:
// discriminant = sqrt(
//				(-27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^2
//				+ 4 (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4)^3)
// We further do common term elimination, by writing:
// apple = -27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5
// Then we can simplify one of the main terms as
// foo = (discriminant + apple)^(1/3)
// Thus, a is then:
// a = {foo / (3 2^(1/3) (b + y))}
//		- {(2^(1/3) (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4))
//		  /(3 (b + y) foo}
//      + {(b x + x y)/(b + y)}
// Furthermore, it becomes clearer in this expression, that
// (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4) is a sub-term of the discriminant.
// thus we call it banana.
// a = {foo / (3 2^(1/3) (b + y))}
//		- {(2^(1/3) banana)
//		  /(3 (b + y) foo}
//      + {(b x + x y)/(b + y)}
// discriminant = sqrt((apple)^2 + 4 (banana)^3)
func solveCfmm(xReserve, yReserve, yIn sdk.Dec) sdk.Dec {
	if !yReserve.Add(yIn).IsPositive() {
		panic("invalid yReserve, yIn combo")
	}
	// prelude, compute all the xy cross terms. Consider keeping these precomputed in the struct,
	// and maybe in state.
	x := xReserve
	x2 := x.Mul(x)
	y := yReserve
	y2 := y.Mul(y)

	xy := x.Mul(y)

	b := yIn

	bpy := b.Add(y)
	bpy2 := bpy.Mul(bpy)

	// TODO: Once we have correctness tests, can come back and optimize alot of the calculations

	// banana = (3 b^4 + 12 b^3 y + 18 b^2 y^2 + 12 b y^3 + 3 y^4)
	// banana = 3 (b + y)^4
	banana := bpy2.Mul(bpy2) // (b + y)^4
	banana = banana.MulInt64Mut(3)

	// apple = -27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5
	// e = -apple/27 = b^2 x^3 y + b^2 x y^3 + 2 b x^3 y^2 + 2 b x y^4 + x^3 y^3 + x y^5
	// e = x y (b + y)^2 (x^2 + y^2)
	// apple = -27 e

	e := xy
	e = e.Mul(bpy2)
	x2py2 := x2.Add(y2)
	e = e.MulMut(x2py2)
	apple := e.MulInt64(-27) // apple = - 27e

	// d = discriminant = sqrt((apple)^2 + 4 (banana)^3)
	// d2 = (apple)^2 + 4 (banana)^3
	// d2 = e^2 + 4 banana^3
	d2 := apple.Mul(apple)
	d2 = d2.AddMut(banana.Power(3).MulInt64(4))
	d, _ := d2.ApproxSqrt()

	// foo = (discriminant + apple)^(1/3)
	foo3 := d.Add(apple)
	foo, _ := foo3.ApproxRoot(3)

	// a = {foo / (3 2^(1/3) (b + y))}
	//		- {(2^(1/3) banana) / (3 (b + y) foo}
	//      + {(b x + x y)/(b + y)}

	// term1 := {foo / (3 2^(1/3) (b + y))}
	term1Denominator := threeCubeRootTwo.Mul(bpy)
	term1 := foo.Quo(term1Denominator)
	// term2 := {(2^(1/3) banana) / (3 (b + y) foo}
	term2 := cubeRootTwo.Mul(banana)
	term2 = term2.Quo(bpy.Mul(foo).MulInt64Mut(3))
	// term3 := {(b x + x y)/(b + y)}
	term3Numerator := b.Mul(x).Add(xy)
	term3 := term3Numerator.Quo(bpy)

	a := term1.Sub(term2).Add(term3)
	return a
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
