package osmomath

import "fmt"

const uint64MaxInt = ^uint64(0)

var (
	// truncated at precision end.
	numeratorCoefficients13Param = []BigDec{
		MustNewDecFromStr("1.000000000000000000000044212244679434"),
		MustNewDecFromStr("0.352032455817400196452603772766844426"),
		MustNewDecFromStr("0.056507868883666405413116800969512484"),
		MustNewDecFromStr("0.005343900728213034434757419480319916"),
		MustNewDecFromStr("0.000317708814342353603087543715930732"),
		MustNewDecFromStr("0.000011429747507407623028722262874632"),
		MustNewDecFromStr("0.000000198381965651614980168744540366"),
	}
	// rounded up at precision end.
	denominatorCoefficients13Param = []BigDec{
		OneDec(),
		MustNewDecFromStr("0.341114724742545112949699755780593311").Neg(),
		MustNewDecFromStr("0.052724071627342653404436933178482287"),
		MustNewDecFromStr("0.004760950735524957576233524801866342").Neg(),
		MustNewDecFromStr("0.000267168475410566529819971616894193"),
		MustNewDecFromStr("0.000008923715368802211181557353097439").Neg(),
		MustNewDecFromStr("0.000000140277233177373698516010555917"),
	}

	uint64MaxBigInt = NewIntFromUint64(uint64MaxInt)
)

// exp2 takes 2 to the power of a given decimal exponent
// and returns the result.
// 2^{x + y} = 2^x * 2^y
func exp2(exponent BigDec) BigDec {
	isNegativeExponent := exponent.IsNegative()
	if isNegativeExponent {
		exponent = exponent.Neg()
	}

	integerExponentDec := exponent.TruncateDec()

	fractionalExponent := exponent.Sub(integerExponentDec)

	integerExponent := integerExponentDec.TruncateInt()

	integerResult := OneDec()
	// 2^(maxUint64Value + x) = 2^maxUint64Value * 2^2
	for integerExponent.GT(uint64MaxBigInt) {
		integerResult = integerResult.Mul(twoBigDec.PowerInteger(uint64MaxInt))
		integerExponent = integerExponent.Sub(uint64MaxBigInt)
	}

	integerResult = integerResult.Mul(twoBigDec.PowerInteger(integerExponent.Uint64()))

	fractionalResult := exp2ChebyshevRationalApprox(fractionalExponent)

	result := integerResult.Mul(fractionalResult)

	if isNegativeExponent {
		return OneDec().Quo(result)
	}

	return result
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

	h_x := numeratorCoefficients13Param[0].Clone()
	p_x := denominatorCoefficients13Param[0].Clone()
	x_exp_i := OneDec()
	for i := 1; i < len(numeratorCoefficients13Param); i++ {
		x_exp_i.MulMut(x)

		h_x = h_x.Add(numeratorCoefficients13Param[i].Mul(x_exp_i))
		p_x = p_x.Add(denominatorCoefficients13Param[i].Mul(x_exp_i))
	}

	return h_x.Quo(p_x)
}
