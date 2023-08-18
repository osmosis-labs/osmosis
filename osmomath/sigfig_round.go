package osmomath

var pointOne = OneSDKDec().QuoInt64(10)

// SigFigRound rounds to a specified significant figure.
func SigFigRound(d SDKDec, tenToSigFig SDKInt) SDKDec {
	if d.IsZero() {
		return d
	}
	// for d > .1, we do round(d * 10^sigfig) / 10^sigfig
	// for k, where 10^k*d > .1 && 10^{k-1}*d < .1, we do:
	// (round(10^k * d * 10^sigfig) / (10^sigfig * 10^k)
	// take note of floor div, vs normal div
	k := uint64(0)
	dTimesK := d
	for ; dTimesK.LT(pointOne); k += 1 {
		dTimesK.MulInt64Mut(10)
	}
	// d * 10^k * 10^sigfig
	dkSigFig := dTimesK.MulInt(tenToSigFig)
	numerator := dkSigFig.RoundInt().ToDec()

	tenToK := NewSDKInt(10).ToDec().Power(k)
	denominator := tenToSigFig.Mul(tenToK.TruncateInt())
	return numerator.QuoInt(denominator)
}
