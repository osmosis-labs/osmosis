package stableswap

import (
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/internal/test_helpers"
)

// CFMMTestCase defines a testcase for stableswap pools
type CFMMTestCase struct {
	xReserve    osmomath.BigDec
	yReserve    osmomath.BigDec
	remReserves []osmomath.BigDec
	yIn         osmomath.BigDec
	expectPanic bool
}

var (
	overflowDec           = osmomath.NewDecFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(1024), nil), big.NewInt(1)))
	twoAssetCFMMTestCases = map[string]CFMMTestCase{
		// sanity checks
		"small pool small input": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: false,
		},
		"small pool large input": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(99),
			expectPanic: false,
		},
		"medium pool medium join": {
			xReserve:    osmomath.NewBigDec(100000),
			yReserve:    osmomath.NewBigDec(100000),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(10000),
			expectPanic: false,
		},
		"large pool medium join": {
			xReserve:    osmomath.NewBigDec(10000000),
			yReserve:    osmomath.NewBigDec(10000000),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(10000),
			expectPanic: false,
		},
		"large pool large join": {
			xReserve:    osmomath.NewBigDec(10000000),
			yReserve:    osmomath.NewBigDec(10000000),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1000000),
			expectPanic: false,
		},
		"very large pool medium join": {
			xReserve:    osmomath.NewBigDec(1000000000),
			yReserve:    osmomath.NewBigDec(1000000000),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(100000),
			expectPanic: false,
		},
		"billion token pool hundred million token join": {
			xReserve:    osmomath.NewBigDec(1000000000),
			yReserve:    osmomath.NewBigDec(1000000000),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(100000000),
			expectPanic: false,
		},

		// uneven reserves
		"xReserve double yReserve (small)": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    osmomath.NewBigDec(50),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: false,
		},
		"yReserve double xReserve (small)": {
			xReserve:    osmomath.NewBigDec(50),
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: false,
		},
		"xReserve double yReserve (large)": {
			xReserve:    osmomath.NewBigDec(13789470),
			yReserve:    osmomath.NewBigDec(59087324),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1047829),
			expectPanic: false,
		},
		"yReserve double xReserve (large)": {
			xReserve:    osmomath.NewBigDec(50000000),
			yReserve:    osmomath.NewBigDec(100000000),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1000000),
			expectPanic: false,
		},
		"uneven medium pool medium join": {
			xReserve:    osmomath.NewBigDec(123456),
			yReserve:    osmomath.NewBigDec(434245),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(23314),
			expectPanic: false,
		},
		"uneven large pool medium join": {
			xReserve:    osmomath.NewBigDec(11023432),
			yReserve:    osmomath.NewBigDec(17432897),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(89734),
			expectPanic: false,
		},
		"uneven large pool large join": {
			xReserve:    osmomath.NewBigDec(38987364),
			yReserve:    osmomath.NewBigDec(52893462),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(9819874),
			expectPanic: false,
		},
		"uneven very large pool medium join": {
			xReserve:    osmomath.NewBigDec(1473891748),
			yReserve:    osmomath.NewBigDec(7438971234),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(100000),
			expectPanic: false,
		},
		"uneven billion token pool billion token join": {
			xReserve:    osmomath.NewBigDec(2678238934),
			yReserve:    osmomath.NewBigDec(1573917894),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(5378748),
			expectPanic: false,
		},

		// panic catching
		"yIn greater than pool reserves": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1000),
			expectPanic: true,
		},
		"xReserve negative": {
			xReserve:    osmomath.NewBigDec(-100),
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"yReserve negative": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    osmomath.NewBigDec(-100),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"yIn negative": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(-1),
			expectPanic: true,
		},

		// overflows
		"xReserve near max bitlen": {
			xReserve:    overflowDec,
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"yReserve near max bitlen": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    overflowDec,
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"both assets near max bitlen": {
			xReserve:    overflowDec,
			yReserve:    overflowDec,
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
	}

	multiAssetCFMMTestCases = map[string]CFMMTestCase{
		// sanity checks
		"even 3-asset small pool, small input": {
			xReserve: osmomath.NewBigDec(100),
			yReserve: osmomath.NewBigDec(100),
			// represents a 3-asset pool with 100 in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: false,
		},
		"even 3-asset medium pool, small input": {
			xReserve: osmomath.NewBigDec(100000),
			yReserve: osmomath.NewBigDec(100000),
			// represents a 3-asset pool with 100,000 in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100000)},
			yIn:         osmomath.NewBigDec(100),
			expectPanic: false,
		},
		"even 4-asset small pool, small input": {
			xReserve: osmomath.NewBigDec(100),
			yReserve: osmomath.NewBigDec(100),
			// represents a 4-asset pool with 100 in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: false,
		},
		"even 4-asset medium pool, small input": {
			xReserve: osmomath.NewBigDec(100000),
			yReserve: osmomath.NewBigDec(100000),
			// represents a 4-asset pool with 100,000 in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100000), osmomath.NewBigDec(100000)},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: false,
		},
		/* TODO: increase BigDec precision (36 -> 72) to be able to accommodate this
		"even 4-asset large pool, small input": {
			xReserve: osmomath.NewBigDec(100000000),
			yReserve: osmomath.NewBigDec(100000000),
			// represents a 4-asset pool with 100M in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100000000), osmomath.NewBigDec(100000000)},
			yIn: osmomath.NewBigDec(100),
			expectPanic: false,
		},
		*/

		// uneven pools
		"uneven 3-asset pool, even swap assets as pool minority": {
			xReserve: osmomath.NewBigDec(100),
			yReserve: osmomath.NewBigDec(100),
			// the asset not being swapped has 100,000 token reserves (swap assets in pool minority)
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100000)},
			yIn:         osmomath.NewBigDec(10),
			expectPanic: false,
		},
		"uneven 3-asset pool, uneven swap assets as pool minority, y > x": {
			xReserve: osmomath.NewBigDec(100),
			yReserve: osmomath.NewBigDec(200),
			// the asset not being swapped has 100,000 token reserves (swap assets in pool minority)
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100000)},
			yIn:         osmomath.NewBigDec(10),
			expectPanic: false,
		},
		"uneven 3-asset pool, uneven swap assets as pool minority, x > y": {
			xReserve: osmomath.NewBigDec(200),
			yReserve: osmomath.NewBigDec(100),
			// the asset not being swapped has 100,000 token reserves (swap assets in pool minority)
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100000)},
			yIn:         osmomath.NewBigDec(10),
			expectPanic: false,
		},
		"uneven 3-asset pool, no round numbers": {
			xReserve: osmomath.NewBigDec(1178349),
			yReserve: osmomath.NewBigDec(8329743),
			// the asset not being swapped has 329,847 token reserves (swap assets in pool minority)
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(329847)},
			yIn:         osmomath.NewBigDec(10),
			expectPanic: false,
		},
		"uneven 4-asset pool, small input and swap assets in pool minority": {
			xReserve: osmomath.NewBigDec(100),
			yReserve: osmomath.NewBigDec(100),
			// the assets not being swapped have 100,000 token reserves each (swap assets in pool minority)
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100000), osmomath.NewBigDec(100000)},
			yIn:         osmomath.NewBigDec(10),
			expectPanic: false,
		},
		"uneven 4-asset pool, even swap assets in pool majority": {
			xReserve: osmomath.NewBigDec(100000),
			yReserve: osmomath.NewBigDec(100000),
			// the assets not being swapped have 100 token reserves each (swap assets in pool majority)
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(10),
			expectPanic: false,
		},
		"uneven 4-asset pool, uneven swap assets in pool majority, y > x": {
			xReserve: osmomath.NewBigDec(100000),
			yReserve: osmomath.NewBigDec(200000),
			// the assets not being swapped have 100 token reserves each (swap assets in pool majority)
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(10),
			expectPanic: false,
		},
		"uneven 4-asset pool, uneven swap assets in pool majority, y < x": {
			xReserve: osmomath.NewBigDec(200000),
			yReserve: osmomath.NewBigDec(100000),
			// the assets not being swapped have 100 token reserves each (swap assets in pool majority)
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(10),
			expectPanic: false,
		},
		"uneven 4-asset pool, no round numbers": {
			xReserve: osmomath.NewBigDec(1178349),
			yReserve: osmomath.NewBigDec(8329743),
			// the assets not being swapped have 329,847 tokens and 4,372,897 respectively
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(329847), osmomath.NewBigDec(4372897)},
			yIn:         osmomath.NewBigDec(10),
			expectPanic: false,
		},

		// panic catching
		"negative xReserve": {
			xReserve: osmomath.NewBigDec(-100),
			yReserve: osmomath.NewBigDec(100),
			// represents a 4-asset pool with 100 in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"negative yReserve": {
			xReserve: osmomath.NewBigDec(100),
			yReserve: osmomath.NewBigDec(-100),
			// represents a 4-asset pool with 100 in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"negative remReserve": {
			xReserve: osmomath.NewBigDec(100),
			yReserve: osmomath.NewBigDec(100),
			// represents a 4-asset pool with 100 in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(-100), osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"negative yIn": {
			xReserve: osmomath.NewBigDec(100),
			yReserve: osmomath.NewBigDec(100),
			// represents a 4-asset pool with 100 in each reserve
			remReserves: []osmomath.BigDec{},
			yIn:         osmomath.NewBigDec(-1),
			expectPanic: true,
		},
		"input greater than pool reserves (even 4-asset pool)": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(1000),
			expectPanic: true,
		},

		// overflows
		"xReserve overflows in 4-asset pool": {
			xReserve:    overflowDec,
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"yReserve overflows in 4-asset pool": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    overflowDec,
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), osmomath.NewBigDec(100)},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"remReserve overflows in 3-asset pool": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{overflowDec},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"remReserve overflows in 4-asset pool": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), overflowDec},
			yIn:         osmomath.NewBigDec(1),
			expectPanic: true,
		},
		"yIn overflows in 4-asset pool": {
			xReserve:    osmomath.NewBigDec(100),
			yReserve:    osmomath.NewBigDec(100),
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100), osmomath.NewBigDec(100)},
			yIn:         overflowDec,
			expectPanic: true,
		},
	}
)

type StableSwapTestSuite struct {
	test_helpers.CfmmCommonTestSuite
}

func TestCFMMInvariantTwoAssets(t *testing.T) {
	kErrTolerance := osmomath.OneDec()

	tests := twoAssetCFMMTestCases

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				// ensure there are only two assets
				require.True(t, len(test.remReserves) == 0)

				// using two-asset cfmm
				k0 := cfmmConstant(test.xReserve, test.yReserve)
				xOut := solveCfmm(test.xReserve, test.yReserve, test.remReserves, test.yIn)

				k1 := cfmmConstant(test.xReserve.Sub(xOut), test.yReserve.Add(test.yIn))
				osmomath.DecApproxEq(t, k0, k1, kErrTolerance)
			}

			osmoassert.ConditionalPanic(t, test.expectPanic, sut)
		})
	}
}

func TestCFMMInvariantTwoAssetsBinarySearch(t *testing.T) {
	kErrTolerance := osmomath.OneDec()

	tests := twoAssetCFMMTestCases

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				// using two-asset binary search cfmm solver
				k0 := cfmmConstant(test.xReserve, test.yReserve)
				xOut := solveCFMMBinarySearch(cfmmConstant)(test.xReserve, test.yReserve, test.yIn)

				k1 := cfmmConstant(test.xReserve.Sub(xOut), test.yReserve.Add(test.yIn))
				osmomath.DecApproxEq(t, k0, k1, kErrTolerance)
			}

			osmoassert.ConditionalPanic(t, test.expectPanic, sut)
		})
	}
}

func TestCFMMInvariantTwoAssetsDirect(t *testing.T) {
	kErrTolerance := osmomath.OneDec()

	tests := twoAssetCFMMTestCases

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				// using two-asset cfmm
				k0 := cfmmConstant(test.xReserve, test.yReserve)
				xOut := solveCfmmDirect(test.xReserve, test.yReserve, test.yIn)

				k1 := cfmmConstant(test.xReserve.Sub(xOut), test.yReserve.Add(test.yIn))
				osmomath.DecApproxEq(t, k0, k1, kErrTolerance)
			}

			osmoassert.ConditionalPanic(t, test.expectPanic, sut)
		})
	}
}

func TestCFMMInvariantMultiAssets(t *testing.T) {
	kErrTolerance := osmomath.OneDec()

	tests := multiAssetCFMMTestCases

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				uReserve := calcUReserve(test.remReserves)
				wSumSquares := calcWSumSquares(test.remReserves)

				// using multi-asset cfmm
				k2 := cfmmConstantMulti(test.xReserve, test.yReserve, uReserve, wSumSquares)
				xOut2 := solveCfmm(test.xReserve, test.yReserve, test.remReserves, test.yIn)
				k3 := cfmmConstantMulti(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), uReserve, wSumSquares)
				osmomath.DecApproxEq(t, k2, k3, kErrTolerance)
			}

			osmoassert.ConditionalPanic(t, test.expectPanic, sut)
		})
	}
}

func TestCFMMInvariantMultiAssetsBinarySearch(t *testing.T) {
	kErrTolerance := osmomath.OneDec()

	tests := multiAssetCFMMTestCases

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				uReserve := calcUReserve(test.remReserves)
				wSumSquares := calcWSumSquares(test.remReserves)

				// using multi-asset cfmm
				k2 := cfmmConstantMulti(test.xReserve, test.yReserve, uReserve, wSumSquares)
				xOut2 := solveCFMMBinarySearchMulti(cfmmConstantMulti)(test.xReserve, test.yReserve, uReserve, wSumSquares, test.yIn)
				k3 := cfmmConstantMulti(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), uReserve, wSumSquares)
				osmomath.DecApproxEq(t, k2, k3, kErrTolerance)
			}

			osmoassert.ConditionalPanic(t, test.expectPanic, sut)
		})
	}
}

func (suite *StableSwapTestSuite) Test_StableSwap_CalculateAmountOutAndIn_InverseRelationship(t *testing.T) {
	type testcase struct {
		denomOut         string
		initialPoolOut   int64
		initialWeightOut int64
		initialCalcOut   int64

		denomIn         string
		initialPoolIn   int64
		initialWeightIn int64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := []testcase{
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000_000_000_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 200,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 50,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000_000,
			initialWeightOut: 200,
			initialCalcOut:   100000,

			denomIn:         "ion",
			initialPoolIn:   1_000_000_000,
			initialWeightIn: 50,
		},
	}

	swapFeeCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(tc testcase, swapFeeCase string) string {
		return fmt.Sprintf("tokenOutInitial: %d, tokenInInitial: %d, initialOut: %d, swapFee: %s",
			tc.initialPoolOut,
			tc.initialPoolIn,
			tc.initialCalcOut,
			swapFeeCase,
		)
	}

	for _, tc := range testcases {
		for _, swapFee := range swapFeeCases {
			t.Run(getTestCaseName(tc, swapFee), func(t *testing.T) {
				ctx := suite.CreateTestContext()

				poolLiquidityIn := sdk.NewInt64Coin(tc.denomOut, tc.initialPoolOut)
				poolLiquidityOut := sdk.NewInt64Coin(tc.denomIn, tc.initialPoolIn)
				poolLiquidity := sdk.NewCoins(poolLiquidityIn, poolLiquidityOut)

				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				require.NoError(t, err)

				exitFeeDec, err := sdk.NewDecFromStr("0")
				require.NoError(t, err)

				pool := createTestPool(t, poolLiquidity, swapFeeDec, exitFeeDec)
				require.NotNil(t, pool)

				suite.TestCalculateAmountOutAndIn_InverseRelationship(ctx, pool, poolLiquidityIn.Denom, poolLiquidityOut.Denom, tc.initialCalcOut, swapFeeDec)
			})
		}
	}
}

func calcUReserve(remReserves []osmomath.BigDec) osmomath.BigDec {
	uReserve := osmomath.OneDec()
	for _, assetReserve := range remReserves {
		uReserve = uReserve.Mul(assetReserve)
	}
	return uReserve
}

func calcWSumSquares(remReserves []osmomath.BigDec) osmomath.BigDec {
	wSumSquares := osmomath.ZeroDec()
	for _, assetReserve := range remReserves {
		wSumSquares = wSumSquares.Add(assetReserve.Mul(assetReserve))
	}
	return wSumSquares
}
