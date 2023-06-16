package osmomath

import (
	"math/big"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
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
		{sdk.NewDecFromInt(sdk.NewInt(2)), sdk.NewDecWithPrec(1414213562373095049, 18)}, // 2 => 1.414213562373095049
		{smallestDec, sdk.NewDecWithPrec(1, 9)},                                         // 10^-18 => 10^-9
		{smallestDec.MulInt64(3), sdk.NewDecWithPrec(1732050808, 18)},                   // 3*10^-18 => sqrt(3)*10^-9
	}

	for i, tc := range testCases {
		res, err := MonotonicSqrt(tc.input)
		require.NoError(t, err)
		require.Equal(t, tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func TestSqrtMonotinicity(t *testing.T) {
	type testcase struct {
		smaller sdk.Dec
		bigger  sdk.Dec
	}
	testCases := []testcase{
		{sdk.MustNewDecFromStr("120.120060020005000000"), sdk.MustNewDecFromStr("120.120060020005000001")},
		{sdk.SmallestDec(), sdk.SmallestDec().MulInt64(2)},
	}
	// create random test vectors for every bit-length
	r := rand.New(rand.NewSource(rand.Int63()))
	differences := []sdk.Dec{}
	for i := 0; i < 5; i++ {
		differences = append(differences, smallestDec.MulInt64(1<<i))
	}
	for i := 1; i < 255+sdk.DecimalPrecisionBits; i++ {
		upperbound := big.NewInt(1)
		upperbound.Lsh(upperbound, uint(i))
		for j := 0; j < 100; j++ {
			v := big.NewInt(0).Rand(r, upperbound)
			d := sdk.NewDecFromBigIntWithPrec(v, 18)
			testCases = append(testCases, testcase{d, d.Add(differences[j%len(differences)])})
		}
	}
	for i := 0; i < 1024; i++ {
		d := sdk.NewDecWithPrec(int64(i), 18)
		testCases = append(testCases, testcase{d, d.Add(smallestDec)})
	}

	for _, i := range testCases {
		sqrtSmaller, err := MonotonicSqrt(i.smaller)
		require.NoError(t, err, "smaller: %s", i.smaller)
		sqrtBigger, err := MonotonicSqrt(i.bigger)
		require.NoError(t, err, "bigger: %s", i.bigger)
		assert.True(t, sqrtSmaller.LTE(sqrtBigger), "sqrtSmaller: %s, sqrtBigger: %s", sqrtSmaller, sqrtBigger)

		// separately sanity check that sqrt * sqrt >= input
		sqrtSmallerSquared := sqrtSmaller.Mul(sqrtSmaller)
		assert.True(t, sqrtSmallerSquared.GTE(i.smaller), "sqrt %s, sqrtSmallerSquared: %s, smaller: %s", sqrtSmaller, sqrtSmallerSquared, i.smaller)
	}
}

func TestSqrtRounding(t *testing.T) {
	testCases := []sdk.Dec{
		sdk.MustNewDecFromStr("11662930532952632574132537947829685675668532938920838254939577167671385459971.396347723368091000"),
	}
	for _, i := range testCases {
		sqrt, err := MonotonicSqrt(i)
		require.NoError(t, err, "smaller: %s", i)
		// separately sanity check that sqrt * sqrt >= input
		sqrtSquared := sqrt.Mul(sqrt)
		assert.True(t, sqrtSquared.GTE(i), "sqrt %s, sqrtSquared: %s, original: %s", sqrt, sqrtSquared, i)
	}
}
