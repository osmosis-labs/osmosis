package osmomath

import "fmt"

var (
	// truncated at precision end.
	numeratorCoefficients = []BigDec{
		MustNewDecFromStr("1.000000000000000000000044212244679434"),
		MustNewDecFromStr("0.352032455817400196452603772766844426"),
		MustNewDecFromStr("0.056507868883666405413116800969512484"),
		MustNewDecFromStr("0.005343900728213034434757419480319916"),
		MustNewDecFromStr("0.000317708814342353603087543715930732"),
		MustNewDecFromStr("0.000011429747507407623028722262874632"),
		MustNewDecFromStr("0.000000198381965651614980168744540366"),
	}
	// rounded up at precision end.
	denominatorCoefficients = []BigDec{
		OneDec(),
		MustNewDecFromStr("0.341114724742545112949699755780593311").Neg(),
		MustNewDecFromStr("0.052724071627342653404436933178482287"),
		MustNewDecFromStr("0.004760950735524957576233524801866342").Neg(),
		MustNewDecFromStr("0.000267168475410566529819971616894193"),
		MustNewDecFromStr("0.000008923715368802211181557353097439").Neg(),
		MustNewDecFromStr("0.000000140277233177373698516010555917"),
	}
)

// Exp2 takes 2 to the power of a given decimal exponent
// and returns the result.
// 2^{x + y} = 2^x * 2^y
func Exp2(exponent BigDec) BigDec {

	integerExponent := exponent.TruncateDec()

	fractionalExponent := exponent.Sub(integerExponent)

	integerResult := twoBigDec.Power(integerExponent.TruncateInt().Uint64())

	fractionalResult := exp2ChebyshevRationalApprox(fractionalExponent)

	return integerResult.Mul(fractionalResult)
}

// exp2ChebyshevRationalApprox takes 2 to the power of a given decimal exponent.
// The result is approximated by a 13 parameter Chebyshev rational approximation.
// f(x) = h(x) / p(x) (7, 7) terms. We set the first term of p(x) to 1.
// As a result, this ends up being 7 + 6 = 13 parameters.
// See scripts/approximations/README.md for details of the scripts used
// to compute the coefficients.
// CONTRACT: exponent must be in the range [0, 1], panics if not.
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

	h_x := numeratorCoefficients[0].Clone()
	p_x := denominatorCoefficients[0].Clone()
	x_exp_i := OneDec()
	for i := 1; i < len(numeratorCoefficients); i++ {
		x_exp_i.MulMut(x)

		h_x = h_x.Add(numeratorCoefficients[i].Mul(x_exp_i))
		p_x = p_x.Add(denominatorCoefficients[i].Mul(x_exp_i))
	}

	return h_x.Quo(p_x)
}
