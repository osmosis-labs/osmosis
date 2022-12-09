package osmomath_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v13/osmomath"
)

func TestExp2ChebyshevRationalApprox(t *testing.T) {

	tests := map[string]struct {
		exponent       osmomath.BigDec
		expectedResult osmomath.BigDec
		expectPanic    bool
	}{
		"example test": {
			exponent: osmomath.MustNewDecFromStr("0.5"),
			// https://www.wolframalpha.com/input?i=2%5E0.5+37+digits
			expectedResult: osmomath.MustNewDecFromStr("1.414213562373095048801688724209698079"),
		},

		// TODO:
		// exponent < 0
		// exponent > 1
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			osmoassert.ConditionalPanic(t, tc.expectPanic, func() {

				result := osmomath.Exp2ChebyshevRationalApprox(tc.exponent)

				require.Equal(t, tc.expectedResult, result)
			})
		})
	}
}
