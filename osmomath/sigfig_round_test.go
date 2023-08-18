package osmomath

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSigFigRound(t *testing.T) {
	// sigfig = 8
	tenToSigFig := NewSDKDec(10).Power(8).TruncateInt()

	testCases := []struct {
		name           string
		decimal        SDKDec
		tenToSigFig    SDKInt
		expectedResult SDKDec
	}{
		{
			name:           "Zero decimal",
			decimal:        ZeroSDKDec(),
			tenToSigFig:    tenToSigFig,
			expectedResult: ZeroSDKDec(),
		},
		{
			name:           "Zero tenToSigFig",
			decimal:        MustNewSDKDecFromStr("2.123"),
			tenToSigFig:    ZeroSDKInt(),
			expectedResult: ZeroSDKDec(),
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
			decimal:        MustNewSDKDecFromStr("63.045"),
			tenToSigFig:    tenToSigFig,
			expectedResult: MustNewSDKDecFromStr("63.045"),
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
			decimal:        MustNewSDKDecFromStr("0.0867245957"),
			tenToSigFig:    tenToSigFig,
			expectedResult: MustNewSDKDecFromStr("0.086724596"),
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
			decimal:        MustNewSDKDecFromStr("0.086724"),
			tenToSigFig:    NewSDKInt(100),
			expectedResult: MustNewSDKDecFromStr("0.087"),
		},
		{
			name:           "minimum decimal is still kept",
			decimal:        NewSDKDecWithPrec(1, 18),
			tenToSigFig:    NewSDKInt(10),
			expectedResult: NewSDKDecWithPrec(1, 18),
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var actualResult SDKDec
			ConditionalPanic(t, tc.tenToSigFig.Equal(ZeroSDKInt()), func() {
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
