package stableswap

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	sdkrand "github.com/osmosis-labs/osmosis/v15/simulation/simtypes/random"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/internal/test_helpers"
	types "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
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
		"even 4-asset large pool (100M each), small input": {
			xReserve: osmomath.NewBigDec(100000000),
			yReserve: osmomath.NewBigDec(100000000),
			// represents a 4-asset pool with 100M in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100000000), osmomath.NewBigDec(100000000)},
			yIn:         osmomath.NewBigDec(100),
			expectPanic: false,
		},
		"even 4-asset pool (10B each post-scaled), small input": {
			xReserve: osmomath.NewBigDec(10000000000),
			yReserve: osmomath.NewBigDec(10000000000),
			// represents a 4-asset pool with 10B in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(10000000000), osmomath.NewBigDec(10000000000)},
			yIn:         osmomath.NewBigDec(100000000),
			expectPanic: false,
		},
		"even 10-asset pool (10B each post-scaled), small input": {
			xReserve: osmomath.NewBigDec(10_000_000_000),
			yReserve: osmomath.NewBigDec(10_000_000_000),
			// represents a 10-asset pool with 10B in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(10_000_000_000), osmomath.NewBigDec(10_000_000_000), osmomath.NewBigDec(10_000_000_000), osmomath.NewBigDec(10_000_000_000), osmomath.NewBigDec(10_000_000_000), osmomath.NewBigDec(10_000_000_000), osmomath.NewBigDec(10_000_000_000), osmomath.NewBigDec(10_000_000_000)},
			yIn:         osmomath.NewBigDec(100),
			expectPanic: false,
		},
		"even 10-asset pool (100B each post-scaled), large input": {
			xReserve: osmomath.NewBigDec(100_000_000_000),
			yReserve: osmomath.NewBigDec(100_000_000_000),
			// represents a 10-asset pool with 100B in each reserve
			remReserves: []osmomath.BigDec{osmomath.NewBigDec(100_000_000_000), osmomath.NewBigDec(100_000_000_000), osmomath.NewBigDec(100_000_000_000), osmomath.NewBigDec(100_000_000_000), osmomath.NewBigDec(100_000_000_000), osmomath.NewBigDec(100_000_000_000), osmomath.NewBigDec(100_000_000_000), osmomath.NewBigDec(100_000_000_000)},
			yIn:         osmomath.NewBigDec(10_000_000_000),
			expectPanic: false,
		},

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

func TestStableSwapTestSuite(t *testing.T) {
	suite.Run(t, new(StableSwapTestSuite))
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

func TestCFMMInvariantMultiAssetsDirect(t *testing.T) {
	kErrTolerance := osmomath.OneDec()

	tests := multiAssetCFMMTestCases

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				wSumSquares := calcWSumSquares(test.remReserves)

				// using multi-asset cfmm
				k2 := cfmmConstantMultiNoV(test.xReserve, test.yReserve, wSumSquares)
				xOut2 := solveCFMMMultiDirect(test.xReserve, test.yReserve, wSumSquares, test.yIn)
				k3 := cfmmConstantMultiNoV(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), wSumSquares)
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
				wSumSquares := calcWSumSquares(test.remReserves)

				// using multi-asset cfmm
				k2 := cfmmConstantMultiNoV(test.xReserve, test.yReserve, wSumSquares)
				xOut2 := solveCFMMBinarySearchMulti(test.xReserve, test.yReserve, wSumSquares, test.yIn)
				k3 := cfmmConstantMultiNoV(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), wSumSquares)
				osmomath.DecApproxEq(t, k2, k3, kErrTolerance)
			}

			osmoassert.ConditionalPanic(t, test.expectPanic, sut)
		})
	}
}

func (suite *StableSwapTestSuite) Test_StableSwap_CalculateAmountOutAndIn_InverseRelationship() {
	type testcase struct {
		denomOut       string
		initialCalcOut int64

		denomIn string

		poolLiquidity  sdk.Coins
		scalingFactors []uint64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := map[string]testcase{
		// two-asset pools
		"even pool": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000_000_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(1_000_000_000)),
			),
			scalingFactors: []uint64{1, 1},
		},
		"uneven pool (2:1)": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(500_000)),
			),
			scalingFactors: []uint64{1, 1},
		},
		"uneven pool (1_000_000:1)": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000_000_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(1_000)),
			),
			scalingFactors: []uint64{1, 1},
		},
		"uneven pool (1:1_000_000)": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(1_000_000_000)),
			),
			scalingFactors: []uint64{1, 1},
		},
		"even pool, uneven scaling factors": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000_000_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(1_000_000_000)),
			),
			scalingFactors: []uint64{1, 8},
		},
		"uneven pool, uneven scaling factors": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(500_000)),
			),
			scalingFactors: []uint64{1, 9},
		},

		// multi asset pools
		"even multi-asset pool": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(1_000_000)),
				sdk.NewCoin("foo", sdk.NewInt(1_000_000)),
			),
			scalingFactors: []uint64{1, 1, 1},
		},
		"uneven multi-asset pool (2:1:2)": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(500_000)),
				sdk.NewCoin("foo", sdk.NewInt(1_000_000)),
			),
			scalingFactors: []uint64{1, 1, 1},
		},
		"uneven multi-asset pool (1_000_000:1:1_000_000)": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(1_000)),
				sdk.NewCoin("foo", sdk.NewInt(1_000_000)),
			),
			scalingFactors: []uint64{1, 1, 1},
		},
		"uneven multi-asset pool (1:1_000_000:1_000_000)": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(1_000_000)),
				sdk.NewCoin("foo", sdk.NewInt(1_000_000)),
			),
			scalingFactors: []uint64{1, 1, 1},
		},
		"even multi-asset pool, uneven scaling factors": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(1_000_000)),
				sdk.NewCoin("foo", sdk.NewInt(1_000_000)),
			),
			scalingFactors: []uint64{5, 3, 9},
		},
		"uneven multi-asset pool (2:1:2), uneven scaling factors": {
			denomIn:        "ion",
			denomOut:       "uosmo",
			initialCalcOut: 100,

			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("ion", sdk.NewInt(1_000_000)),
				sdk.NewCoin("uosmo", sdk.NewInt(500_000)),
				sdk.NewCoin("foo", sdk.NewInt(1_000_000)),
			),
			scalingFactors: []uint64{100, 76, 33},
		},
	}
	// create randomized test cases
	r := rand.New(rand.NewSource(12345))
	coinMax := sdk.NewInt(10).ToDec().Power(30).TruncateInt()
	for c := 2; c < 5; c++ {
		for i := 0; i < 10; i++ {
			coins := sdk.NewCoins()
			scalingFactors := []uint64{}
			for j := 0; j < c; j++ {
				coins = coins.Add(sdkrand.RandExponentialCoin(r, sdk.NewCoin(fmt.Sprintf("token%d", j), coinMax)))
				sf := sdkrand.RandIntBetween(r, 1, 1<<60)
				scalingFactors = append(scalingFactors, uint64(sf))
			}
			initialCalcOut := sdkrand.RandIntBetween(r, 1, 1<<60)
			testcases[fmt.Sprintf("rand_case_%d_coins_%d", c, i)] = testcase{
				denomIn:        coins[0].Denom,
				denomOut:       coins[1].Denom,
				initialCalcOut: int64(initialCalcOut),
				poolLiquidity:  coins,
				scalingFactors: scalingFactors,
			}
		}
	}

	swapFeeCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(name string, tc testcase, swapFeeCase string) string {
		return fmt.Sprintf("%s: initialOut: %d, swapFee: %s",
			name,
			tc.initialCalcOut,
			swapFeeCase,
		)
	}

	for name, tc := range testcases {
		for _, swapFee := range swapFeeCases {
			suite.Run(getTestCaseName(name, tc, swapFee), func() {
				ctx := suite.CreateTestContext()

				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				suite.Require().NoError(err)

				exitFeeDec, err := sdk.NewDecFromStr("0")
				suite.Require().NoError(err)

				// TODO: add scaling factors into inverse relationship tests
				pool := createTestPool(suite.T(), tc.poolLiquidity, swapFeeDec, exitFeeDec, tc.scalingFactors)
				suite.Require().NotNil(pool)
				errTolerance := osmomath.ErrTolerance{
					AdditiveTolerance: sdk.Dec{}, MultiplicativeTolerance: sdk.NewDecWithPrec(1, 12)}
				test_helpers.TestCalculateAmountOutAndIn_InverseRelationship(suite.T(), ctx, pool, tc.denomIn, tc.denomOut, tc.initialCalcOut, swapFeeDec, errTolerance)
			})
		}
	}
}

func (suite *StableSwapTestSuite) Test_StableSwap_Slippage_LiquidityRelation() {
	type testcase struct {
		initialLiquidity sdk.Coins
		scalingFactors   []uint64
	}
	testcases := map[string]testcase{
		"1:1 pool, 1:1 SF": {
			initialLiquidity: sdk.NewCoins(sdk.NewInt64Coin("bar", 10000), sdk.NewInt64Coin("foo", 10000)),
			scalingFactors:   []uint64{},
		},
		"10:1 pool, 1:1 SF": {
			initialLiquidity: sdk.NewCoins(sdk.NewInt64Coin("bar", 100000), sdk.NewInt64Coin("foo", 10000)),
			scalingFactors:   []uint64{},
		},
		"10:1 pool, 3:2 SF": {
			initialLiquidity: sdk.NewCoins(sdk.NewInt64Coin("bar", 100000), sdk.NewInt64Coin("foo", 10000)),
			scalingFactors:   []uint64{3, 2},
		},
	}
	swapFeeCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}
	for name, tc := range testcases {
		for _, swapFee := range swapFeeCases {
			createPoolFn := func(ctx sdk.Context, liq sdk.Coins) types.CFMMPoolI {
				return createTestPool(suite.T(), liq, sdk.MustNewDecFromStr(swapFee), sdk.ZeroDec(), tc.scalingFactors)
			}
			ctx := sdk.Context{}
			test_helpers.TestSlippageRelationWithLiquidityIncrease(name, suite.T(), ctx, createPoolFn, tc.initialLiquidity)
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

func TestCalcSingleAssetJoinShares(t *testing.T) {
	type testcase struct {
		tokenIn        sdk.Coin
		poolAssets     sdk.Coins
		scalingFactors []uint64
		swapFee        sdk.Dec
		expectedOut    sdk.Int
	}

	tests := map[string]testcase{
		// no swap fees
		"even two asset pool, no swap fee": {
			tokenIn:        sdk.NewCoin("foo", sdk.NewInt(100)),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectedOut:    sdk.NewInt(100),
		},
		"uneven two asset pool, no swap fee": {
			tokenIn:        sdk.NewCoin("foo", sdk.NewInt(100)),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectedOut:    sdk.NewInt(100),
		},
		"even 3-asset pool, no swap fee": {
			tokenIn:        sdk.NewCoin("asset/a", sdk.NewInt(1000)),
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectedOut:    sdk.NewInt(1000),
		},
		"uneven 3-asset pool, no swap fee": {
			tokenIn:        sdk.NewCoin("asset/a", sdk.NewInt(100)),
			poolAssets:     threeUnevenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectedOut:    sdk.NewInt(100),
		},

		// with swap fees
		"even two asset pool, default swap fee": {
			tokenIn:        sdk.NewCoin("foo", sdk.NewInt(100)),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        defaultSwapFee,
			expectedOut:    sdk.NewInt(100 - 3),
		},
		"uneven two asset pool, default swap fee": {
			tokenIn:        sdk.NewCoin("foo", sdk.NewInt(100)),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        defaultSwapFee,
			expectedOut:    sdk.NewInt(100 - 3),
		},
		"even 3-asset pool, default swap fee": {
			tokenIn:        sdk.NewCoin("asset/a", sdk.NewInt(100)),
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        defaultSwapFee,
			expectedOut:    sdk.NewInt(100 - 3),
		},
		"uneven 3-asset pool, default swap fee": {
			tokenIn:        sdk.NewCoin("asset/a", sdk.NewInt(100)),
			poolAssets:     threeUnevenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        defaultSwapFee,
			expectedOut:    sdk.NewInt(100 - 3),
		},
		"even 3-asset pool, 0.03 swap fee": {
			tokenIn:        sdk.NewCoin("asset/a", sdk.NewInt(100)),
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        sdk.MustNewDecFromStr("0.03"),
			expectedOut:    sdk.NewInt(100 - 3),
		},
		"uneven 3-asset pool, 0.03 swap fee": {
			tokenIn:        sdk.NewCoin("asset/a", sdk.NewInt(100)),
			poolAssets:     threeUnevenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        sdk.MustNewDecFromStr("0.03"),
			expectedOut:    sdk.NewInt(100 - 3),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			shares, err := p.calcSingleAssetJoinShares(tc.tokenIn, tc.swapFee)
			require.NoError(t, err, "test: %s", name)

			p.updatePoolForJoin(sdk.Coins{tc.tokenIn}, shares)
			exitTokens, err := p.ExitPool(ctx, shares, sdk.ZeroDec())
			require.NoError(t, err, "test: %s", name)

			// since each asset swap can have up to sdk.OneInt() error, our expected error bound is 1*numAssets
			correctnessThreshold := sdk.OneInt().Mul(sdk.NewInt(int64(len(p.PoolLiquidity))))

			tokenOutAmount, err := cfmm_common.SwapAllCoinsToSingleAsset(&p, ctx, exitTokens, tc.tokenIn.Denom, sdk.ZeroDec())
			require.True(t, tokenOutAmount.LTE(tc.tokenIn.Amount))
			require.True(t, tc.expectedOut.Sub(tokenOutAmount).Abs().LTE(correctnessThreshold))
		})
	}
}

func TestJoinPoolSharesInternal(t *testing.T) {
	tenPercentOfTwoPoolRaw := int64(1000000000 / 10)
	tenPercentOfTwoPoolCoins := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(int64(1000000000/10))), sdk.NewCoin("bar", sdk.NewInt(int64(1000000000/10))))
	twoAssetPlusTenPercent := twoEvenStablePoolAssets.Add(tenPercentOfTwoPoolCoins...)
	type testcase struct {
		tokensIn        sdk.Coins
		poolAssets      sdk.Coins
		scalingFactors  []uint64
		swapFee         sdk.Dec
		expNumShare     sdk.Int
		expTokensJoined sdk.Coins
		expPoolAssets   sdk.Coins
		expectPass      bool
	}

	tests := map[string]testcase{
		"even two asset pool, same tokenIn ratio": {
			tokensIn:        tenPercentOfTwoPoolCoins,
			poolAssets:      twoEvenStablePoolAssets,
			scalingFactors:  defaultTwoAssetScalingFactors,
			swapFee:         sdk.ZeroDec(),
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: tenPercentOfTwoPoolCoins,
			expPoolAssets:   twoAssetPlusTenPercent,
			expectPass:      true,
		},
		"even two asset pool, different tokenIn ratio with pool": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPoolRaw)), sdk.NewCoin("bar", sdk.NewInt(10+tenPercentOfTwoPoolRaw))),
			poolAssets:      twoEvenStablePoolAssets,
			scalingFactors:  defaultTwoAssetScalingFactors,
			swapFee:         sdk.ZeroDec(),
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPoolRaw)), sdk.NewCoin("bar", sdk.NewInt(tenPercentOfTwoPoolRaw))),
			expPoolAssets:   twoAssetPlusTenPercent,
			expectPass:      true,
		},
		"all-asset pool join attempt exceeds max scaled asset amount": {
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 1),
				sdk.NewInt64Coin("bar", 1),
			),
			poolAssets: sdk.NewCoins(
				sdk.NewCoin("foo", types.StableswapMaxScaledAmtPerAsset),
				sdk.NewCoin("bar", types.StableswapMaxScaledAmtPerAsset),
			),
			scalingFactors:  defaultTwoAssetScalingFactors,
			swapFee:         sdk.ZeroDec(),
			expNumShare:     sdk.ZeroInt(),
			expTokensJoined: sdk.Coins{},
			expPoolAssets: sdk.NewCoins(
				sdk.NewCoin("foo", types.StableswapMaxScaledAmtPerAsset),
				sdk.NewCoin("bar", types.StableswapMaxScaledAmtPerAsset),
			),
			expectPass: false,
		},
		"single-asset pool join exceeds hits max scaled asset amount": {
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 2),
			),
			poolAssets: sdk.NewCoins(
				sdk.NewCoin("foo", types.StableswapMaxScaledAmtPerAsset),
				sdk.NewCoin("bar", types.StableswapMaxScaledAmtPerAsset),
			),
			scalingFactors:  defaultTwoAssetScalingFactors,
			swapFee:         sdk.ZeroDec(),
			expNumShare:     sdk.ZeroInt(),
			expTokensJoined: sdk.Coins{},
			expPoolAssets: sdk.NewCoins(
				sdk.NewCoin("foo", types.StableswapMaxScaledAmtPerAsset),
				sdk.NewCoin("bar", types.StableswapMaxScaledAmtPerAsset),
			),
			expectPass: false,
		},
		"all-asset pool join attempt exactly hits max scaled asset amount": {
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 1),
				sdk.NewInt64Coin("bar", 1),
			),
			poolAssets: sdk.NewCoins(
				sdk.NewCoin("foo", types.StableswapMaxScaledAmtPerAsset.Sub(sdk.NewInt(1))),
				sdk.NewCoin("bar", types.StableswapMaxScaledAmtPerAsset.Sub(sdk.NewInt(1))),
			),
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expNumShare:    types.InitPoolSharesSupply.Quo(types.StableswapMaxScaledAmtPerAsset),
			expTokensJoined: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 1),
				sdk.NewInt64Coin("bar", 1),
			),
			expPoolAssets: sdk.NewCoins(
				sdk.NewCoin("foo", types.StableswapMaxScaledAmtPerAsset),
				sdk.NewCoin("bar", types.StableswapMaxScaledAmtPerAsset),
			),
			expectPass: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			shares, joinedLiquidity, err := p.joinPoolSharesInternal(ctx, tc.tokensIn, tc.swapFee)

			if tc.expectPass {
				require.Equal(t, tc.expNumShare, shares)
				require.Equal(t, tc.expTokensJoined, joinedLiquidity)
				require.Equal(t, tc.expPoolAssets, p.PoolLiquidity)
			}
			osmoassert.ConditionalError(t, !tc.expectPass, err)
		})
	}
}

func TestSingleAssetJoinSwapFeeRatio(t *testing.T) {
	largeInt, ok := sdk.NewIntFromString("123456789012345678")
	require.True(t, ok)
	type testcase struct {
		poolLiquidity  sdk.Coins
		scalingFactors []uint64
		tokenInDenom   string
		expectedRatio  sdk.Dec
	}
	tests := map[string]testcase{
		"godoc-example": {
			poolLiquidity:  sdk.NewCoins(sdk.NewInt64Coin("tokenA", 80), sdk.NewInt64Coin("tokenB", 20)),
			scalingFactors: []uint64{1, 1},
			tokenInDenom:   "tokenA",
			expectedRatio:  sdk.MustNewDecFromStr("0.2"),
		},
		"godoc-example-denom-rev": {
			poolLiquidity:  sdk.NewCoins(sdk.NewInt64Coin("tokenA", 80), sdk.NewInt64Coin("tokenB", 20)),
			scalingFactors: []uint64{1, 1},
			tokenInDenom:   "tokenB",
			expectedRatio:  sdk.MustNewDecFromStr("0.8"),
		},
		"80:20 -> 1:1 scaling factor": {
			poolLiquidity:  sdk.NewCoins(sdk.NewInt64Coin("tokenA", 80), sdk.NewInt64Coin("tokenB", 20)),
			scalingFactors: []uint64{80, 20},
			tokenInDenom:   "tokenA",
			expectedRatio:  sdk.MustNewDecFromStr("0.5"),
		},
		"80:20 -> 2:1 scaling factor": {
			poolLiquidity:  sdk.NewCoins(sdk.NewInt64Coin("tokenA", 80), sdk.NewInt64Coin("tokenB", 20)),
			scalingFactors: []uint64{40, 20},
			tokenInDenom:   "tokenA",
			expectedRatio:  sdk.MustNewDecFromStr("0.333333333333333334"),
		},
		"60:40:40, large numbers": {
			poolLiquidity: sdk.NewCoins(
				sdk.NewCoin("tokenA", largeInt.MulRaw(6)),
				sdk.NewCoin("tokenB", largeInt.MulRaw(4)),
				sdk.NewCoin("tokenC", largeInt.MulRaw(4))),
			scalingFactors: []uint64{1, 1, 1},
			tokenInDenom:   "tokenA",
			// 1 - (6 / 14) = 8/14 = 4/7 ~= 0.571428571428571429
			expectedRatio: sdk.MustNewDecFromStr("0.571428571428571429"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p := poolStructFromAssets(tc.poolLiquidity, tc.scalingFactors)

			ratio, err := p.singleAssetJoinSwapFeeRatio(tc.tokenInDenom)
			require.NoError(t, err)
			require.Equal(t, tc.expectedRatio, ratio)
		})
	}
}
