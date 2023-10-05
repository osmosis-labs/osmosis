package osmomath

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAbsDifferenceWithSign(t *testing.T) {
	decA, err := NewDecFromStr("3.2")
	require.NoError(t, err)
	decB, err := NewDecFromStr("4.3432389")
	require.NoError(t, err)

	s, b := AbsDifferenceWithSign(decA, decB)
	require.True(t, b)

	expectedDec, err := NewDecFromStr("1.1432389")
	require.NoError(t, err)
	require.Equal(t, expectedDec, s)
}

func TestPowApprox(t *testing.T) {
	testCases := []struct {
		base           Dec
		exp            Dec
		powPrecision   Dec
		expectedResult Dec
	}{
		{
			// medium base, small exp
			base:           MustNewDecFromStr("0.8"),
			exp:            MustNewDecFromStr("0.32"),
			powPrecision:   MustNewDecFromStr("0.00000001"),
			expectedResult: MustNewDecFromStr("0.93108385"),
		},
		{
			// zero exp
			base:           MustNewDecFromStr("0.8"),
			exp:            ZeroDec(),
			powPrecision:   MustNewDecFromStr("0.00001"),
			expectedResult: OneDec(),
		},
		{
			// zero base, this should panic
			base:           ZeroDec(),
			exp:            OneDec(),
			powPrecision:   MustNewDecFromStr("0.00001"),
			expectedResult: ZeroDec(),
		},
		{
			// large base, small exp
			base:           MustNewDecFromStr("1.9999"),
			exp:            MustNewDecFromStr("0.23"),
			powPrecision:   MustNewDecFromStr("0.000000001"),
			expectedResult: MustNewDecFromStr("1.172821461"),
		},
		{
			// large base, large integer exp
			base:           MustNewDecFromStr("1.777"),
			exp:            MustNewDecFromStr("20"),
			powPrecision:   MustNewDecFromStr("0.000000000001"),
			expectedResult: MustNewDecFromStr("98570.862372081602"),
		},
		{
			// medium base, large exp, high precision
			base:           MustNewDecFromStr("1.556"),
			exp:            MustNewDecFromStr("20.9123"),
			powPrecision:   MustNewDecFromStr("0.0000000000000001"),
			expectedResult: MustNewDecFromStr("10360.058421529811344618"),
		},
		{
			// high base, large exp, high precision
			base:           MustNewDecFromStr("1.886"),
			exp:            MustNewDecFromStr("31.9123"),
			powPrecision:   MustNewDecFromStr("0.00000000000001"),
			expectedResult: MustNewDecFromStr("621110716.84727942280335811"),
		},
		{
			// base equal one
			base:           MustNewDecFromStr("1"),
			exp:            MustNewDecFromStr("123"),
			powPrecision:   MustNewDecFromStr("0.00000001"),
			expectedResult: OneDec(),
		},
	}

	for i, tc := range testCases {
		var actualResult Dec
		ConditionalPanic(t, tc.base.IsZero(), func() {
			fmt.Println(tc.base)
			actualResult = PowApprox(tc.base, tc.exp, tc.powPrecision)
			require.True(
				t,
				tc.expectedResult.Sub(actualResult).Abs().LTE(tc.powPrecision),
				fmt.Sprintf("test %d failed: expected value & actual value's difference should be less than precision", i),
			)
		})
	}
}

func TestPow(t *testing.T) {
	testCases := []struct {
		base           Dec
		exp            Dec
		expectedResult Dec
	}{
		{
			// medium base, small exp
			base:           MustNewDecFromStr("0.8"),
			exp:            MustNewDecFromStr("0.32"),
			expectedResult: MustNewDecFromStr("0.93108385"),
		},
		{
			// zero exp
			base:           MustNewDecFromStr("0.8"),
			exp:            ZeroDec(),
			expectedResult: OneDec(),
		},
		{
			// zero base, this should panic
			base:           ZeroDec(),
			exp:            OneDec(),
			expectedResult: ZeroDec(),
		},
		{
			// large base, small exp
			base:           MustNewDecFromStr("1.9999"),
			exp:            MustNewDecFromStr("0.23"),
			expectedResult: MustNewDecFromStr("1.172821461"),
		},
		{
			// small base, large exp
			base:           MustNewDecFromStr("0.0000123"),
			exp:            MustNewDecFromStr("123"),
			expectedResult: ZeroDec(),
		},
		{
			// large base, large exp
			base:           MustNewDecFromStr("1.777"),
			exp:            MustNewDecFromStr("20"),
			expectedResult: MustNewDecFromStr("98570.862372081602"),
		},
		{
			// base equal one
			base:           MustNewDecFromStr("1"),
			exp:            MustNewDecFromStr("123"),
			expectedResult: OneDec(),
		},
	}

	for i, tc := range testCases {
		var actualResult Dec
		ConditionalPanic(t, tc.base.IsZero(), func() {
			actualResult = Pow(tc.base, tc.exp)
			require.True(
				t,
				tc.expectedResult.Sub(actualResult).Abs().LTE(powPrecision),
				fmt.Sprintf("test %d failed: expected value & actual value's difference should be less than precision", i),
			)
		})
	}
}
