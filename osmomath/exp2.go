package osmomath

import "fmt"

var (
	// Truncated at precision end.
	// See scripts/approximations/main.py exponent_approximation_choice function for details.
	numeratorCoefficients13Param = []BigDec{
		MustNewDecFromStr("1.000000000000000000000044212244679434"),
		MustNewDecFromStr("0.352032455817400196452603772766844426"),
		MustNewDecFromStr("0.056507868883666405413116800969512484"),
		MustNewDecFromStr("0.005343900728213034434757419480319916"),
		MustNewDecFromStr("0.000317708814342353603087543715930732"),
		MustNewDecFromStr("0.000011429747507407623028722262874632"),
		MustNewDecFromStr("0.000000198381965651614980168744540366"),
	}

	// Rounded up at precision end.
	// See scripts/approximations/main.py exponent_approximation_choice function for details.
	denominatorCoefficients13Param = []BigDec{
		OneDec(),
		MustNewDecFromStr("0.341114724742545112949699755780593311").Neg(),
		MustNewDecFromStr("0.052724071627342653404436933178482287"),
		MustNewDecFromStr("0.004760950735524957576233524801866342").Neg(),
		MustNewDecFromStr("0.000267168475410566529819971616894193"),
		MustNewDecFromStr("0.000008923715368802211181557353097439").Neg(),
		MustNewDecFromStr("0.000000140277233177373698516010555916"),
	}

	// maxSupportedExponent = 2^10. The value is chosen by benchmarking
	// when the underlying internal functions overflow.
	// If needed in the future, Exp2 can be reimplemented to allow for greater exponents.
	maxSupportedExponent = MustNewDecFromStr("2").PowerInteger(9)
)

// Exp2 takes 2 to the power of a given non-negative decimal exponent
// and returns the result.
// The computation is performed by using th following property:
// 2^decimal_exp = 2^{integer_exp + fractional_exp} = 2^integer_exp * 2^fractional_exp
// The max supported exponent is defined by the global maxSupportedExponent.
// If a greater exponent is given, the function panics.
// Panics if the exponent is negative.
// The answer is correct up to a factor of 10^-18.
// Meaning, result = result * k for k in [1 - 10^(-18), 1 + 10^(-18)]
// Note: our Python script plots show accuracy up to a factor of 10^22.
// However, in Go tests we only test up to 10^18. Therefore, this is the guarantee.
func Exp2(exponent BigDec) BigDec {
	if exponent.Abs().GT(maxSupportedExponent) {
		panic(fmt.Sprintf("integer exponent %s is too large, max (%s)", exponent, maxSupportedExponent))
	}
	if exponent.IsNegative() {
		panic(fmt.Sprintf("negative exponent %s is not supported", exponent))
	}

	integerExponent := exponent.TruncateDec()

	fractionalExponent := exponent.Sub(integerExponent)
	fractionalResult := exp2ChebyshevRationalApprox(fractionalExponent)

	// Left bit shift is equivalent to multiplying by 2^integerExponent.
	fractionalResult.i = fractionalResult.i.Lsh(fractionalResult.i, uint(integerExponent.TruncateInt().Uint64()))

	return fractionalResult
}

// exp2ChebyshevRationalApprox takes 2 to the power of a given decimal exponent.
// The result is approximated by a 13 parameter Chebyshev rational approximation.
// f(x) = h(x) / p(x) (7, 7) terms. We set the first term of p(x) to 1.
// As a result, this ends up being 7 + 6 = 13 parameters.
// The numerator coefficients are truncated at precision end. The denominator
// coefficients are rounded up at precision end.
// See scripts/approximations/README.md for details of the scripts used
// to compute the coefficients.
// CONTRACT: exponent must be in the range [0, 1], panics if not.
// The answer is correct up to a factor of 10^-18.
// Meaning, result = result * k for k in [1 - 10^(-18), 1 + 10^(-18)]
// Note: our Python script plots show accuracy up to a factor of 10^22.
// However, in Go tests we only test up to 10^18. Therefore, this is the guarantee.
func exp2ChebyshevRationalApprox(x BigDec) BigDec {
	if x.LT(ZeroDec()) || x.GT(OneDec()) {
		panic(fmt.Sprintf("exponent must be in the range [0, 1], got %s", x))
	}
	if x.IsZero() {
		return OneDec()
	}
	if x.Equal(OneDec()) {
		return twoBigDec
	}

	h_x := numeratorCoefficients13Param[0].Clone()
	p_x := denominatorCoefficients13Param[0].Clone()
	x_exp_i := OneDec()
	for i := 1; i < len(numeratorCoefficients13Param); i++ {
		x_exp_i.MulMut(x)

		h_x.AddMut(numeratorCoefficients13Param[i].Mul(x_exp_i))
		p_x.AddMut(denominatorCoefficients13Param[i].Mul(x_exp_i))
	}

	return h_x.QuoMut(p_x)
}
