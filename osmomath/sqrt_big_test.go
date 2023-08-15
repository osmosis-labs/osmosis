package osmomath

import (
	"math/big"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateRandomDecForEachBitlenBigDec(r *rand.Rand, numPerBitlen int) []BigDec {
	return generateRandomDecForEachBitlen[BigDec](r, numPerBitlen, NewDecFromBigIntWithPrec, Precision)
}

func TestSdkApproxSqrtVectors_BigDec(t *testing.T) {
	testCases := []struct {
		input    BigDec
		expected BigDec
	}{
		{OneDec(), OneDec()},                                                                    // 1.0 => 1.0
		{NewDecWithPrec(25, 2), NewDecWithPrec(5, 1)},                                           // 0.25 => 0.5
		{NewDecWithPrec(4, 2), NewDecWithPrec(2, 1)},                                            // 0.09 => 0.3
		{NewDecFromInt(NewInt(9)), NewDecFromInt(NewInt(3))},                                    // 9 => 3
		{NewDecFromInt(NewInt(2)), MustNewDecFromStr("1.414213562373095048801688724209698079")}, // 2 => 1.414213562373095048801688724209698079
		{smallestBigDec, NewDecWithPrec(1, 18)},                                                 // 10^-36 => 10^-18
		{smallestBigDec.MulInt64(3), NewDecWithPrec(1732050807568877294, 36)},                   // 3*10^-36 => sqrt(3)*10^-18
	}

	for i, tc := range testCases {
		res, err := MonotonicSqrtBigDec(tc.input)
		require.NoError(t, err)
		require.Equal(t, tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func testMonotonicityAroundBigDec(t *testing.T, x BigDec) {
	// test that sqrt(x) is monotonic around x
	// i.e. sqrt(x-1) <= sqrt(x) <= sqrt(x+1)
	sqrtX, err := MonotonicSqrtBigDec(x)
	require.NoError(t, err)
	sqrtXMinusOne, err := MonotonicSqrtBigDec(x.Sub(smallestBigDec))
	require.NoError(t, err)
	sqrtXPlusOne, err := MonotonicSqrtBigDec(x.Add(smallestBigDec))
	require.NoError(t, err)
	assert.True(t, sqrtXMinusOne.LTE(sqrtX), "sqrtXMinusOne: %s, sqrtX: %s", sqrtXMinusOne, sqrtX)
	assert.True(t, sqrtX.LTE(sqrtXPlusOne), "sqrtX: %s, sqrtXPlusOne: %s", sqrtX, sqrtXPlusOne)
}

func TestSqrtMonotinicity_BigDec(t *testing.T) {
	type testcase struct {
		smaller BigDec
		bigger  BigDec
	}
	testCases := []testcase{
		{MustNewDecFromStr("120.120060020005000000"), MustNewDecFromStr("120.120060020005000001")},
		{smallestBigDec, smallestBigDec.MulInt64(2)},
	}
	// create random test vectors for every bit-length
	r := rand.New(rand.NewSource(rand.Int63()))
	for i := 0; i < 255+sdk.DecimalPrecisionBits; i++ {
		upperbound := big.NewInt(1)
		upperbound.Lsh(upperbound, uint(i))
		for j := 0; j < 100; j++ {
			v := big.NewInt(0).Rand(r, upperbound)
			d := NewDecFromBigIntWithPrec(v, 36)
			testCases = append(testCases, testcase{d, d.Add(smallestBigDec)})
		}
	}
	for i := 0; i < 1024; i++ {
		d := NewDecWithPrec(int64(i), 18)
		testCases = append(testCases, testcase{d, d.Add(smallestBigDec)})
	}

	for _, i := range testCases {
		sqrtSmaller, err := MonotonicSqrtBigDec(i.smaller)
		require.NoError(t, err, "smaller: %s", i.smaller)
		sqrtBigger, err := MonotonicSqrtBigDec(i.bigger)
		require.NoError(t, err, "bigger: %s", i.bigger)
		assert.True(t, sqrtSmaller.LTE(sqrtBigger), "sqrtSmaller: %s, sqrtBigger: %s", sqrtSmaller, sqrtBigger)

		// separately sanity check that sqrt * sqrt >= input
		sqrtSmallerSquared := sqrtSmaller.Mul(sqrtSmaller)
		assert.True(t, sqrtSmallerSquared.GTE(i.smaller), "sqrt %s, sqrtSmallerSquared: %s, smaller: %s", sqrtSmaller, sqrtSmallerSquared, i.smaller)
	}
}

// Test that square(sqrt(x)) = x when x is a perfect square.
// We do this by sampling sqrt(v) from the set of numbers `a.b`, where a in [0, 2^128], b in [0, 10^9].
// and then setting x = sqrt(v)
// this is because this is the set of values whose squares are perfectly representable.
func TestPerfectSquares_BigDec(t *testing.T) {
	cases := []BigDec{
		NewBigDec(100),
	}
	r := rand.New(rand.NewSource(rand.Int63()))
	tenToMin9 := big.NewInt(1_000_000_000)
	for i := 0; i < 128; i++ {
		upperbound := big.NewInt(1)
		upperbound.Lsh(upperbound, uint(i))
		for j := 0; j < 100; j++ {
			v := big.NewInt(0).Rand(r, upperbound)
			dec := big.NewInt(0).Rand(r, tenToMin9)
			d := NewDecFromBigInt(v).Add(NewDecFromBigIntWithPrec(dec, 9))
			cases = append(cases, d.MulMut(d))
		}
	}

	for _, i := range cases {
		sqrt, err := MonotonicSqrtBigDec(i)
		require.NoError(t, err, "smaller: %s", i)
		assert.Equal(t, i, sqrt.MulMut(sqrt))
		if !i.IsZero() {
			testMonotonicityAroundBigDec(t, i)
		}
	}
}

func TestSqrtRounding_BigDec(t *testing.T) {
	testCases := []BigDec{
		MustNewDecFromStr("11662930532952632574132537947829685675668532938920838254939577167671385459971.396347723368091000"),
	}
	r := rand.New(rand.NewSource(rand.Int63()))
	testCases = append(testCases, generateRandomDecForEachBitlenBigDec(r, 10)...)
	for _, i := range testCases {
		sqrt, err := MonotonicSqrtBigDec(i)
		require.NoError(t, err, "smaller: %s", i)
		// Sanity check that sqrt * sqrt >= input
		sqrtSquared := sqrt.Mul(sqrt)
		assert.True(t, sqrtSquared.GTE(i), "sqrt %s, sqrtSquared: %s, original: %s", sqrt, sqrtSquared, i)
		// (aside) check that (sqrt - 1ulp)^2 <= input
		sqrtMin1 := sqrt.Sub(smallestBigDec)
		sqrtSquared = sqrtMin1.Mul(sqrtMin1)
		assert.True(t, sqrtSquared.LTE(i), "sqrtMin1ULP %s, sqrtSquared: %s, original: %s", sqrt, sqrtSquared, i)
	}
}

// benchmarks the new square root across bit-lengths, for comparison with the SDK square root.
func BenchmarkMonotonicSqrt_BigDec(b *testing.B) {
	r := rand.New(rand.NewSource(1))
	vectors := generateRandomDecForEachBitlenBigDec(r, 1)
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(vectors); j++ {
			a, _ := MonotonicSqrtBigDec(vectors[j])
			_ = a
		}
	}
}
