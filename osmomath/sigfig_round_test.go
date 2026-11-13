package osmomath

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSigFigRound(t *testing.T) {
	// sigfig = 8
	tenToSigFig := NewDec(10).Power(8).TruncateInt()

	testCases := []struct {
		name           string
		decimal        Dec
		tenToSigFig    Int
		expectedResult Dec
	}{
		{
			name:           "Zero decimal",
			decimal:        ZeroDec(),
			tenToSigFig:    tenToSigFig,
			expectedResult: ZeroDec(),
		},
		{
			name:           "Zero tenToSigFig",
			decimal:        MustNewDecFromStr("2.123"),
			tenToSigFig:    ZeroInt(),
			expectedResult: ZeroDec(),
		},
		// With input, decimal >= 0.1. We have:
		// 	- dTimesK = 63.045
		// 	- k = 0
		// Applying the formula, we have:
		//  - numerator = (dTimesK  * tenToSigFig).RoundInt() = 6304500000
		//  - denominator = tenToSigFig * 10^k = 100000000
		//  - result = numerator / denominator = 63
		{
			name:           "Big decimal, default tenToSigFig",
			decimal:        MustNewDecFromStr("63.045"),
			tenToSigFig:    tenToSigFig,
			expectedResult: MustNewDecFromStr("63.045"),
		},
		// With input, decimal < 0.1. We have:
		// 	- dTimesK = 0.86724
		// 	- k = 1
		// Applying the formula, we have:
		//  - numerator = (dTimesK  * tenToSigFig).RoundInt() = 86724596
		//  - denominator = tenToSigFig * 10^k = 1000000000
		//  - result = numerator / denominator = 0.086724
		{
			name:           "Small decimal, default tenToSigFig",
			decimal:        MustNewDecFromStr("0.0867245957"),
			tenToSigFig:    tenToSigFig,
			expectedResult: MustNewDecFromStr("0.086724596"),
		},
		// With input, decimal < 0.1. We have:
		// 	- dTimesK = 0.86724
		// 	- k = 1
		// Applying the formula, we have:
		//  - numerator = (dTimesK  * tenToSigFig).RoundInt() = 87
		//  - denominator = tenToSigFig * 10^k = 1000
		//  - result = numerator / denominator = 0.087
		{
			name:           "Small decimal, random tenToSigFig",
			decimal:        MustNewDecFromStr("0.086724"),
			tenToSigFig:    NewInt(100),
			expectedResult: MustNewDecFromStr("0.087"),
		},
		{
			name:           "minimum decimal is still kept",
			decimal:        NewDecWithPrec(1, 18),
			tenToSigFig:    NewInt(10),
			expectedResult: NewDecWithPrec(1, 18),
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var actualResult Dec
			ConditionalPanic(t, tc.tenToSigFig.Equal(ZeroInt()), func() {
				actualResult = SigFigRound(tc.decimal, tc.tenToSigFig)
				require.Equal(
					t,
					tc.expectedResult,
					actualResult,
					fmt.Sprintf("test %d failed: expected value & actual value should be equal", i),
				)
			})
		})

	}
}

// TestSigFigRound_DoesNotMutateInput verifies that SigFigRound leaves the
// caller's input Dec untouched. Inputs with 0 < d < 0.1 are normalized by a
// mutative multiply internally, so a missing clone would corrupt the caller's
// value by a factor of a power of ten.
func TestSigFigRound_DoesNotMutateInput(t *testing.T) {
	tenToSigFig := NewDec(10).Power(8).TruncateInt()

	testCases := []struct {
		name  string
		input Dec
	}{
		{
			name:  "small decimal below 0.1 is normalized internally",
			input: MustNewDecFromStr("0.0867245957"),
		},
		{
			name:  "very small decimal requires multiple normalization steps",
			input: MustNewDecFromStr("0.00000123"),
		},
		{
			name:  "decimal at or above 0.1 is not normalized",
			input: MustNewDecFromStr("63.045"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.input.Clone()

			SigFigRound(tc.input, tenToSigFig)

			require.Equal(t, expected, tc.input, "SigFigRound must not mutate its input")
		})
	}
}
