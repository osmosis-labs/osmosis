package osmomath

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAbsDifferenceWithSign(t *testing.T) {
	decA, err := NewSDKDecFromStr("3.2")
	require.NoError(t, err)
	decB, err := NewSDKDecFromStr("4.3432389")
	require.NoError(t, err)

	s, b := AbsDifferenceWithSign(decA, decB)
	require.True(t, b)

	expectedDec, err := NewSDKDecFromStr("1.1432389")
	require.NoError(t, err)
	require.Equal(t, expectedDec, s)
}

func TestPowApprox(t *testing.T) {
	testCases := []struct {
		base           SDKDec
		exp            SDKDec
		powPrecision   SDKDec
		expectedResult SDKDec
	}{
		{
			// medium base, small exp
			base:           MustNewSDKDecFromStr("0.8"),
			exp:            MustNewSDKDecFromStr("0.32"),
			powPrecision:   MustNewSDKDecFromStr("0.00000001"),
			expectedResult: MustNewSDKDecFromStr("0.93108385"),
		},
		{
			// zero exp
			base:           MustNewSDKDecFromStr("0.8"),
			exp:            ZeroSDKDec(),
			powPrecision:   MustNewSDKDecFromStr("0.00001"),
			expectedResult: OneSDKDec(),
		},
		{
			// zero base, this should panic
			base:           ZeroSDKDec(),
			exp:            OneSDKDec(),
			powPrecision:   MustNewSDKDecFromStr("0.00001"),
			expectedResult: ZeroSDKDec(),
		},
		{
			// large base, small exp
			base:           MustNewSDKDecFromStr("1.9999"),
			exp:            MustNewSDKDecFromStr("0.23"),
			powPrecision:   MustNewSDKDecFromStr("0.000000001"),
			expectedResult: MustNewSDKDecFromStr("1.172821461"),
		},
		{
			// large base, large integer exp
			base:           MustNewSDKDecFromStr("1.777"),
			exp:            MustNewSDKDecFromStr("20"),
			powPrecision:   MustNewSDKDecFromStr("0.000000000001"),
			expectedResult: MustNewSDKDecFromStr("98570.862372081602"),
		},
		{
			// medium base, large exp, high precision
			base:           MustNewSDKDecFromStr("1.556"),
			exp:            MustNewSDKDecFromStr("20.9123"),
			powPrecision:   MustNewSDKDecFromStr("0.0000000000000001"),
			expectedResult: MustNewSDKDecFromStr("10360.058421529811344618"),
		},
		{
			// high base, large exp, high precision
			base:           MustNewSDKDecFromStr("1.886"),
			exp:            MustNewSDKDecFromStr("31.9123"),
			powPrecision:   MustNewSDKDecFromStr("0.00000000000001"),
			expectedResult: MustNewSDKDecFromStr("621110716.84727942280335811"),
		},
		{
			// base equal one
			base:           MustNewSDKDecFromStr("1"),
			exp:            MustNewSDKDecFromStr("123"),
			powPrecision:   MustNewSDKDecFromStr("0.00000001"),
			expectedResult: OneSDKDec(),
		},
	}

	for i, tc := range testCases {
		var actualResult SDKDec
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
		base           SDKDec
		exp            SDKDec
		expectedResult SDKDec
	}{
		{
			// medium base, small exp
			base:           MustNewSDKDecFromStr("0.8"),
			exp:            MustNewSDKDecFromStr("0.32"),
			expectedResult: MustNewSDKDecFromStr("0.93108385"),
		},
		{
			// zero exp
			base:           MustNewSDKDecFromStr("0.8"),
			exp:            ZeroSDKDec(),
			expectedResult: OneSDKDec(),
		},
		{
			// zero base, this should panic
			base:           ZeroSDKDec(),
			exp:            OneSDKDec(),
			expectedResult: ZeroSDKDec(),
		},
		{
			// large base, small exp
			base:           MustNewSDKDecFromStr("1.9999"),
			exp:            MustNewSDKDecFromStr("0.23"),
			expectedResult: MustNewSDKDecFromStr("1.172821461"),
		},
		{
			// small base, large exp
			base:           MustNewSDKDecFromStr("0.0000123"),
			exp:            MustNewSDKDecFromStr("123"),
			expectedResult: ZeroSDKDec(),
		},
		{
			// large base, large exp
			base:           MustNewSDKDecFromStr("1.777"),
			exp:            MustNewSDKDecFromStr("20"),
			expectedResult: MustNewSDKDecFromStr("98570.862372081602"),
		},
		{
			// base equal one
			base:           MustNewSDKDecFromStr("1"),
			exp:            MustNewSDKDecFromStr("123"),
			expectedResult: OneSDKDec(),
		},
	}

	for i, tc := range testCases {
		var actualResult SDKDec
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
