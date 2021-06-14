package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Don't EVER change after initializing
// TODO: Analyze choice here
var powPrecision, _ = sdk.NewDecFromStr("0.00000001")

// Singletons
var zero sdk.Dec = sdk.ZeroDec()
var one_half sdk.Dec = sdk.MustNewDecFromStr("0.5")
var one sdk.Dec = sdk.OneDec()
var two sdk.Dec = sdk.MustNewDecFromStr("2")


/*********************************************************/

// AbsDifferenceWithSign returns | a - b |, (a - b).sign()
func AbsDifferenceWithSign(a, b sdk.Dec) (sdk.Dec, bool) {
	if a.GTE(b) {
		return a.Sub(b), false
	} else {
		return b.Sub(a), true
	}
}


// pow computes base^(exp)
// However since the exponent is not an integer, we must do an approximation algorithm.
// TODO: In the future, lets add some optimized routines for common exponents, e.g. for common wIn / wOut ratios
// Many simple exponents like 2:1 pools
func Pow(base sdk.Dec, exp sdk.Dec) sdk.Dec {
	// Exponentiation of a negative base with an arbitrary real exponent is not closed within the reals.
	// You can see this by recalling that `i = (-1)^(.5)`. We have to go to complex numbers to define this.
	// (And would have to implement complex logarithms)
	// We don't have a need for negative bases, so we don't include any such logic.
	if !base.IsPositive() {
		panic(fmt.Errorf("base must be greater than 0"))
	}
	// TODO: Remove this if we want to generalize the function,
	// we can adjust the algorithm in this setting.
	if base.GTE(two) {
		panic(fmt.Errorf("base must be lesser than two"))
	}


	// We will use an approximation algorithm to compute the power.
	// Since computing an integer power is easy, we split up the exponent into
	// an integer component and a fractional component.
	integer := exp.TruncateDec()
	fractional := exp.Sub(integer)

	integerPow := base.Power(uint64(integer.TruncateInt64()))

	if fractional.IsZero() {
		return integerPow
	}

	fractionalPow := PowApprox(base, fractional, powPrecision)

	return integerPow.Mul(fractionalPow)
}

// Contract: 0 < base < 2
func PowApprox(base sdk.Dec, exp sdk.Dec, precision sdk.Dec) sdk.Dec {
	if exp.IsZero() {
		return sdk.ZeroDec()
	}

	// Common case optimization
	// Optimize for it being equal to one-half
	if exp.Equal(one_half) {
		output, err := base.ApproxSqrt()
		if err != nil {
			panic(err)
		}
		return output
	}
	// TODO: Make an approx-equal function, and then check if exp * 3 = 1, and do a check accordingly

	a := exp
	x, xneg := AbsDifferenceWithSign(base, one)
	term := sdk.OneDec()
	sum := sdk.OneDec()
	negative := false


	// TODO: Document this computation via taylor expansion
	for i := 1; term.GTE(precision); i++ {
		bigK := sdk.OneDec().MulInt64(int64(i))
		c, cneg := AbsDifferenceWithSign(a, bigK.Sub(one))
		term = term.Mul(c.Mul(x))
		term = term.Quo(bigK)

		if term.IsZero() {
			break
		}
		if xneg {
			negative = !negative
		}

		if cneg {
			negative = !negative
		}

		if negative {
			sum = sum.Sub(term)
		} else {
			sum = sum.Add(term)
		}
	}
	return sum
}
