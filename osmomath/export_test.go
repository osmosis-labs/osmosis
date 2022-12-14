package osmomath

var (
	MaxSupportedExponent = maxSupportedExponent
	EulersNumber         = eulersNumber
	TwoBigDec            = twoBigDec
)

func Exp2ChebyshevRationalApprox(exponent BigDec) BigDec {
	return exp2ChebyshevRationalApprox(exponent)
}
