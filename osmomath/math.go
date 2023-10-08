package osmomath

import (
	"fmt"
)

// Don't EVER change after initializing
// TODO: Analyze choice here.
var powPrecision, _ = NewDecFromStr("0.00000001")

const powIterationLimit = 150_000

var (
	one_half Dec = MustNewDecFromStr("0.5")
	one      Dec = OneDec()
	two      Dec = MustNewDecFromStr("2")

	// https://www.wolframalpha.com/input?i=2.718281828459045235360287471352662498&assumption=%22ClashPrefs%22+-%3E+%7B%22Math%22%7D
	// nolint: unused
	eulersNumber = MustNewBigDecFromStr("2.718281828459045235360287471352662498")
)

// Returns the internal "power precision".
// All fractional exponentiation in osmosis is expected to be accurate up to
// powPrecision.
// *technically* the error term can be greater than this powPrecision,
// but for small bases this bound applies. See comments in the PowApprox function
// for more detail.
func GetPowPrecision() Dec {
	return powPrecision.Clone()
}

/*********************************************************/

// AbsDifferenceWithSign returns | a - b |, (a - b).sign()
// a is mutated and returned.
func AbsDifferenceWithSign(a, b Dec) (Dec, bool) {
	if a.GTE(b) {
		return a.SubMut(b), false
	} else {
		return a.NegMut().AddMut(b), true
	}
}

// func largeBasePow(base Dec, exp Dec) Dec {
// 	// pow requires the base to be <= 2
// }

// Pow computes base^(exp)
// However since the exponent is not an integer, we must do an approximation algorithm.
// TODO: In the future, lets add some optimized routines for common exponents, e.g. for common wIn / wOut ratios
// Many simple exponents like 2:1 pools.
func Pow(base Dec, exp Dec) Dec {
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

// Contract: 0 < base <= 2
// 0 <= exp < 1.
func PowApprox(originalBase Dec, exp Dec, precision Dec) Dec {
	if !originalBase.IsPositive() {
		panic(fmt.Errorf("base must be greater than 0"))
	}

	if exp.IsZero() {
		return OneDec()
	}

	// Common case optimization
	// Optimize for it being equal to one-half
	if exp.Equal(one_half) {
		output, err := originalBase.ApproxSqrt()
		if err != nil {
			panic(err)
		}
		return output
	}
	// TODO: Make an approx-equal function, and then check if exp * 3 = 1, and do a check accordingly

	// We compute this via taking the maclaurin series of (1 + x)^a
	// where x = base - 1.
	// The maclaurin series of (1 + x)^a = sum_{k=0}^{infty} binom(a, k) x^k
	// Binom(a, k) takes the natural continuation on the first parameter, namely that
	// Binom(a, k) = N/D, where D = k!, and N = a(a-1)(a-2)...(a-k+1)
	// Next we show that the absolute value of each term is less than the last term.
	// Note that the change in term n's value vs term n + 1 is a multiplicative factor of
	// v_n = x(a - n) / (n+1)
	// So if |v_n| < 1, we know that each term has a lesser impact on the result than the last.
	// For our bounds on |x| < 1, |a| < 1,
	// it suffices to see for what n is |v_n| < 1,
	// in the worst parameterization of x = 1, a = -1.
	// v_n = |(-1 + epsilon - n) / (n+1)|
	// So |v_n| is always less than 1, as n ranges over the integers.
	//
	// Note that term_n of the expansion is 1 * prod_{i=0}^{n-1} v_i
	// The error if we stop the expansion at term_n is:
	// error_n = sum_{k=n+1}^{infty} term_k
	// At this point we further restrict a >= 0, so 0 <= a < 1.
	// Now we take the _INCORRECT_ assumption that if term_n < p, then
	// error_n < p.
	// This assumption is obviously wrong.
	// However our usages of this function don't use the full domain.
	// With a > 0, |x| << 1, and p sufficiently low, perhaps this actually is true.

	// TODO: Check with our parameterization
	// TODO: If theres a bug, balancer is also wrong here :thonk:

	base := originalBase.Clone()
	x, xneg := AbsDifferenceWithSign(base, one)
	term := OneDec()
	sum := OneDec()
	negative := false

	a := exp.Clone()
	bigK := NewDec(0)
	// TODO: Document this computation via taylor expansion
	for i := int64(1); term.GTE(precision); i++ {
		// At each iteration, we need two values, i and i-1.
		// To avoid expensive big.Int allocation, we reuse bigK variable.
		// On this line, bigK == i-1.
		c, cneg := AbsDifferenceWithSign(a, bigK)
		// On this line, bigK == i.
		bigK.SetInt64(i)
		term.MulMut(c).MulMut(x).QuoMut(bigK)

		// a is mutated on absDifferenceWithSign, reset
		a.Set(exp)

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
			sum.SubMut(term)
		} else {
			sum.AddMut(term)
		}

		if i == powIterationLimit {
			panic(fmt.Errorf("failed to reach precision within %d iterations, best guess: %s for %s^%s", powIterationLimit, sum, originalBase, exp))
		}
	}
	return sum
}
