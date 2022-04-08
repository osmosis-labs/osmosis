package stableswap

import sdk "github.com/cosmos/cosmos-sdk/types"

var cubeRootTwo, _ = sdk.NewDec(2).ApproxRoot(3)

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
// foo = (-27 b^2 x^3 y - 27 b^2 x y^3
//			+ discriminant
// 			- 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^(1/3)
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
// discriminant = sqrt(
//				(-27 b^2 x^3 y - 27 b^2 x y^3 - 54 b x^3 y^2 - 54 b x y^4 - 27 x^3 y^3 - 27 x y^5)^2
//				+ 4 (banana)^3)
func solveCfmm() {

}

//nolint:unused
func spotPrice(assetXAmount sdk.Int, assetYAmount sdk.Int) sdk.Dec {
	// Define f_{y -> x}(a) as the function that outputs the amount of tokens X you'd get by
	// trading "a" units of Y against the pool, assuming 0 swap fee, at the current liquidity.
	// The spot price of the pool is then lim a -> 0, f_{y -> x}(a) / a
	// For uniswap f_{y -> x}(a) = x - xy/(y + a),
	// The spot price equation of y in terms of x is X_SUPPLY/Y_SUPPLY.
	// You can work out that it follows from the above relation!
	//
	// Now we have to work this out for the much more complex CFMM xy(x^2 + y^2).

	return sdk.ZeroDec()
}
