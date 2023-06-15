package osmomath

import (
	"math/big"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Test that the guess for an initial square root value is always greater
// than the true square root value.
func TestInitialSqrtGuessGreaterThanTrueSqrt(t *testing.T) {
	cases := []sdk.Dec{
		sdk.SmallestDec(),
		sdk.MaxSortableDec,
		sdk.NewDecWithPrec(123456, 1),
		sdk.NewDecWithPrec(123456, 7),
	}
	for i := 0; i < 512; i++ {
		cases = append(cases, sdk.NewDec(int64(i)))
	}
	// create random test vectors for every bit-length
	r := rand.New(rand.NewSource(rand.Int63()))
	for i := 1; i < 255+sdk.DecimalPrecisionBits; i++ {
		upperbound := big.NewInt(1)
		upperbound.Lsh(upperbound, uint(i))
		for j := 0; j < 10; j++ {
			v := big.NewInt(0).Rand(r, upperbound)
			cases = append(cases, sdk.NewDecFromBigIntWithPrec(v, 18))
		}
	}
	for _, c := range cases {
		guess := getInitialSquareRootGuess(c)
		if guess.Mul(guess).LT(c) {
			t.Errorf("Guess %v is less than or equal to %v", guess, c)
		}
	}
}

func TestSdkApproxSqrtVectors(t *testing.T) {
	testCases := []struct {
		input    sdk.Dec
		expected sdk.Dec
	}{
		{sdk.OneDec(), sdk.OneDec()},                                                    // 1.0 => 1.0
		{sdk.NewDecWithPrec(25, 2), sdk.NewDecWithPrec(5, 1)},                           // 0.25 => 0.5
		{sdk.NewDecWithPrec(4, 2), sdk.NewDecWithPrec(2, 1)},                            // 0.09 => 0.3
		{sdk.NewDecFromInt(sdk.NewInt(9)), sdk.NewDecFromInt(sdk.NewInt(3))},            // 9 => 3
		{sdk.NewDecFromInt(sdk.NewInt(-9)), sdk.NewDecFromInt(sdk.NewInt(-3))},          // -9 => -3
		{sdk.NewDecFromInt(sdk.NewInt(2)), sdk.NewDecWithPrec(1414213562373095049, 18)}, // 2 => 1.414213562373095049
	}

	for i, tc := range testCases {
		res, err := tc.input.ApproxSqrt()
		require.NoError(t, err)
		require.Equal(t, tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func TestSqrtMonotinicity(t *testing.T) {
	testCases := []struct {
		smaller sdk.Dec
		bigger  sdk.Dec
	}{
		{sdk.MustNewDecFromStr("120.120060020005000000"), sdk.MustNewDecFromStr("120.120060020005000001")},
		{sdk.SmallestDec(), sdk.SmallestDec().MulInt64(2)},
	}

	for _, i := range testCases {
		sqrtSmaller, err := i.smaller.ApproxSqrt()
		require.NoError(t, err)
		sqrtBigger, err := i.bigger.ApproxSqrt()
		require.NoError(t, err)
		require.True(t, sqrtSmaller.LTE(sqrtBigger), "sqrtSmaller: %s, sqrtBigger: %s", sqrtSmaller, sqrtBigger)
	}
}
