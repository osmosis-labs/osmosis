package osmomath

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	MaxSupportedExponent = maxSupportedExponent
	EulersNumber         = eulersNumber
	TwoBigDec            = twoBigDec
)

// 2^128 - 1, needs to be the same as gammtypes.MaxSpotPrice
// but we can't directly import that due to import cycles.
// Hence we use the same var name, in hopes that if any change there happens,
// this is caught via a CTRL+F
var MaxSpotPrice = sdk.NewDec(2).Power(128).Sub(sdk.OneDec())

// ConditionalPanic checks if expectPanic is true, asserts that sut (system under test)
// panics. If expectPanic is false, asserts that sut does not panic.
// returns true if sut panics and false it it does not
func ConditionalPanic(t *testing.T, expectPanic bool, sut func()) {
	if expectPanic {
		require.Panics(t, sut)
		return
	}
	require.NotPanics(t, sut)
}

func Exp2ChebyshevRationalApprox(exponent BigDec) BigDec {
	return exp2ChebyshevRationalApprox(exponent)
}
