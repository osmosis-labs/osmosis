package osmomath

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmoassert"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func TestSigFigRound(t *testing.T) {
	// sigfig = 100000000
	sigfig := gammtypes.SpotPriceSigFigs

	testCases := []struct {
		name           string
		decimal        sdk.Dec
		sigfig         sdk.Int
		expectedResult sdk.Dec
	}{
		{
			name:           "Zero decimal",
			decimal:        sdk.ZeroDec(),
			sigfig:         sigfig,
			expectedResult: sdk.ZeroDec(),
		},
		{
			name:           "Zero sigfig",
			decimal:        sdk.MustNewDecFromStr("2.123"),
			sigfig:         sdk.ZeroInt(),
			expectedResult: sdk.ZeroDec(),
		},
		// With input, decimal >= 0.1. We have:
		// 	- dTimesK = 63.045
		// 	- k = 0
		// Applying the formula, we have:
		//  - numerator = (dTimesK  * sigfig).RoundInt() = 6304500000
		//  - denominator = sigfig * 10^k = 100000000
		//  - result = numerator / denominator = 63
		{
			name:           "Big decimal, default sigfig",
			decimal:        sdk.MustNewDecFromStr("63.045"),
			sigfig:         sigfig,
			expectedResult: sdk.MustNewDecFromStr("63.045"),
		},
		// With input, decimal < 0.1. We have:
		// 	- dTimesK = 0.86724
		// 	- k = 1
		// Applying the formula, we have:
		//  - numerator = (dTimesK  * sigfig).RoundInt() = 86724596
		//  - denominator = sigfig * 10^k = 1000000000
		//  - result = numerator / denominator = 0.086724
		{
			name:           "Small decimal, default sigfig",
			decimal:        sdk.MustNewDecFromStr("0.0867245957"),
			sigfig:         sigfig,
			expectedResult: sdk.MustNewDecFromStr("0.086724596"),
		},
		// With input, decimal < 0.1. We have:
		// 	- dTimesK = 0.86724
		// 	- k = 1
		// Applying the formula, we have:
		//  - numerator = (dTimesK  * sigfig).RoundInt() = 87
		//  - denominator = sigfig * 10^k = 1000
		//  - result = numerator / denominator = 0.087
		{
			name:           "Small decimal, random sigfig",
			decimal:        sdk.MustNewDecFromStr("0.086724"),
			sigfig:         sdk.NewInt(100),
			expectedResult: sdk.MustNewDecFromStr("0.087"),
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var actualResult sdk.Dec
			osmoassert.ConditionalPanic(t, tc.sigfig.Equal(sdk.ZeroInt()), func() {
				actualResult = SigFigRound(tc.decimal, tc.sigfig)
				fmt.Println(sdk.NewDec(10).Power(8).TruncateInt())
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
