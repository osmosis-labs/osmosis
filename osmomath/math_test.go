package osmomath

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestAbsDifferenceWithSign(t *testing.T) {
	decA, err := sdk.NewDecFromStr("3.2")
	require.NoError(t, err)
	decB, err := sdk.NewDecFromStr("4.3432389")
	require.NoError(t, err)

	s, b := AbsDifferenceWithSign(decA, decB)
	require.True(t, b)

	expectedDec, err := sdk.NewDecFromStr("1.1432389")
	require.NoError(t, err)
	require.Equal(t, expectedDec, s)
}

func TestPowApprox(t *testing.T) {
	testCases := []struct {
		base           sdk.Dec
		exp            sdk.Dec
		powPrecision   sdk.Dec
		expectedResult sdk.Dec
	}{
		{
			// medium base, small exp
			base:           sdk.MustNewDecFromStr("0.8"),
			exp:            sdk.MustNewDecFromStr("0.32"),
			powPrecision:   sdk.MustNewDecFromStr("0.00000001"),
			expectedResult: sdk.MustNewDecFromStr("0.93108385"),
		},
		{
			// zero exp
			base:           sdk.MustNewDecFromStr("0.8"),
			exp:            sdk.ZeroDec(),
			powPrecision:   sdk.MustNewDecFromStr("0.00001"),
			expectedResult: sdk.OneDec(),
		},
		{
			// large base, small exp
			base:           sdk.MustNewDecFromStr("1.9999"),
			exp:            sdk.MustNewDecFromStr("0.23"),
			powPrecision:   sdk.MustNewDecFromStr("0.000000001"),
			expectedResult: sdk.MustNewDecFromStr("1.172821461"),
		},
		{
			// large base, large integer exp
			base:           sdk.MustNewDecFromStr("1.777"),
			exp:            sdk.MustNewDecFromStr("20"),
			powPrecision:   sdk.MustNewDecFromStr("0.000000001"),
			expectedResult: sdk.MustNewDecFromStr("98570.862372081602"),
		},
		{
			// small base, large integer exp
			base:           sdk.MustNewDecFromStr("0.123"),
			exp:            sdk.MustNewDecFromStr("12"),
			powPrecision:   sdk.MustNewDecFromStr("0.00000001"),
			expectedResult: sdk.ZeroDec(),
		},
		{
			// base equal one
			base:           sdk.MustNewDecFromStr("1"),
			exp:            sdk.MustNewDecFromStr("123"),
			powPrecision:   sdk.MustNewDecFromStr("0.00000001"),
			expectedResult: sdk.OneDec(),
		},
	}

	for i, tc := range testCases {
		actualResult := PowApprox(tc.base, tc.exp, tc.powPrecision)
		require.True(
			t,
			tc.expectedResult.Sub(s).Abs().LTE(tc.powPrecision),
			fmt.Sprintf("test %d failed: expected value & actual value's difference should be less than precision", i),
		)
	}
}

func TestPow(t *testing.T) {
	testCases := []struct {
		base           sdk.Dec
		exp            sdk.Dec
		expectedResult sdk.Dec
	}{
		{
			// medium base, small exp
			base:           sdk.MustNewDecFromStr("0.8"),
			exp:            sdk.MustNewDecFromStr("0.32"),
			expectedResult: sdk.MustNewDecFromStr("0.93108385"),
		},
		{
			// zero exp
			base:           sdk.MustNewDecFromStr("0.8"),
			exp:            sdk.ZeroDec(),
			expectedResult: sdk.OneDec(),
		},
		{
			// large base, small exp
			base:           sdk.MustNewDecFromStr("1.9999"),
			exp:            sdk.MustNewDecFromStr("0.23"),
			expectedResult: sdk.MustNewDecFromStr("1.172821461"),
		},
		{
			// small base, large exp
			base:           sdk.MustNewDecFromStr("0.0000123"),
			exp:            sdk.MustNewDecFromStr("123"),
			expectedResult: sdk.ZeroDec(),
		},
		{
			// large base, large exp
			base:           sdk.MustNewDecFromStr("1.777"),
			exp:            sdk.MustNewDecFromStr("20"),
			expectedResult: sdk.MustNewDecFromStr("98570.862372081602"),
		},
		{
			// base equal one
			base:           sdk.MustNewDecFromStr("1"),
			exp:            sdk.MustNewDecFromStr("123"),
			expectedResult: sdk.OneDec(),
		},
	}

	for i, tc := range testCases {
		actualResult := Pow(tc.base, tc.exp)
		require.True(
			t,
			tc.expectedResult.Sub(s).Abs().LTE(powPrecision),
			fmt.Sprintf("test %d failed: expected value & actual value's difference should be less than precision", i),
		)
	}
}
