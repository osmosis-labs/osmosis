package osmomath

import sdk "github.com/cosmos/cosmos-sdk/types"

var pointOne = sdk.OneDec().QuoInt64(10)

func SigFigRound(d sdk.Dec, tenToSigFig sdk.Int) sdk.Dec {
	// for d > .1, we do round(d * 10^sigfig) / 10^sigfig
	// for k, where 10^k*d > 1 && 10^{k-1}*d < .1, we do:
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
	denominator := tenToSigFig
	if k != 0 {
		tenToK := sdk.NewInt(10).ToDec().Power(k)
		denominator = denominator.Mul(tenToK.TruncateInt())
	}
	return numerator.QuoInt(denominator)
}
