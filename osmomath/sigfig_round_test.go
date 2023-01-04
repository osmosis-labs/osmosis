package osmomath

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestSigFigRound(t *testing.T) {
	// sigfig = 8
	tenToSigFig := sdk.NewDec(10).Power(8).TruncateInt()

	testCases := []struct {
		name           string
		decimal        sdk.Dec
		tenToSigFig    sdk.Int
		expectedResult sdk.Dec
	}{
		{
			name:           "Zero decimal",
			decimal:        sdk.ZeroDec(),
			tenToSigFig:    tenToSigFig,
			expectedResult: sdk.ZeroDec(),
		},
		{
			name:           "Zero tenToSigFig",
			decimal:        sdk.MustNewDecFromStr("2.123"),
			tenToSigFig:    sdk.ZeroInt(),
			expectedResult: sdk.ZeroDec(),
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
			decimal:        sdk.MustNewDecFromStr("63.045"),
			tenToSigFig:    tenToSigFig,
			expectedResult: sdk.MustNewDecFromStr("63.045"),
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
			decimal:        sdk.MustNewDecFromStr("0.0867245957"),
			tenToSigFig:    tenToSigFig,
			expectedResult: sdk.MustNewDecFromStr("0.086724596"),
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
			decimal:        sdk.MustNewDecFromStr("0.086724"),
			tenToSigFig:    sdk.NewInt(100),
			expectedResult: sdk.MustNewDecFromStr("0.087"),
		},
		{
			name:           "minimum decimal is still kept",
			decimal:        sdk.NewDecWithPrec(1, 18),
			tenToSigFig:    sdk.NewInt(10),
			expectedResult: sdk.NewDecWithPrec(1, 18),
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var actualResult sdk.Dec
			ConditionalPanic(t, tc.tenToSigFig.Equal(sdk.ZeroInt()), func() {
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
