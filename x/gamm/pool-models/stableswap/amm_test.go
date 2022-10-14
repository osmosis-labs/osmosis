package stableswap

import (
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/internal/cfmm_common"
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
		initialPoolOut int64
		initialCalcOut int64

		denomIn       string
		initialPoolIn int64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := []testcase{
		{
			denomOut:       "uosmo",
			initialPoolOut: 1_000_000_000,
			initialCalcOut: 100,

			denomIn:       "ion",
			initialPoolIn: 1_000_000_000,
		},
		{
			denomOut:       "uosmo",
			initialPoolOut: 500_000,
			initialCalcOut: 100,

			denomIn:       "ion",
			initialPoolIn: 1_000_000,
		},
		{
			denomOut:       "uosmo",
			initialPoolOut: 1_000,
			initialCalcOut: 100,

			denomIn:       "ion",
			initialPoolIn: 1_000_000_000,
		},
		{
			denomOut:       "uosmo",
			initialPoolOut: 1_000_000_000,
			initialCalcOut: 100,

			denomIn:       "ion",
			initialPoolIn: 1_000,
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
			suite.Run(getTestCaseName(tc, swapFee), func() {
				ctx := suite.CreateTestContext()

				poolLiquidityIn := sdk.NewInt64Coin(tc.denomOut, tc.initialPoolOut)
				poolLiquidityOut := sdk.NewInt64Coin(tc.denomIn, tc.initialPoolIn)
				poolLiquidity := sdk.NewCoins(poolLiquidityIn, poolLiquidityOut)

				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				suite.Require().NoError(err)

				exitFeeDec, err := sdk.NewDecFromStr("0")
				suite.Require().NoError(err)

				// TODO: add scaling factors into inverse relationship tests
				pool := createTestPool(suite.T(), poolLiquidity, swapFeeDec, exitFeeDec, []uint64{1, 1})
				suite.Require().NotNil(pool)
				test_helpers.TestCalculateAmountOutAndIn_InverseRelationship(suite.T(), ctx, pool, poolLiquidityIn.Denom, poolLiquidityOut.Denom, tc.initialCalcOut, swapFeeDec)
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

		// TODO: increase BigDec precision further to be able to accommodate 5-asset pool tests
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			shares, err := p.calcSingleAssetJoinShares(tc.tokenIn, tc.swapFee)
			require.NoError(t, err, "test: %s", name)

			p.updatePoolLiquidityForExit(sdk.Coins{tc.tokenIn})
			exitTokens, err := p.ExitPool(ctx, shares, sdk.ZeroDec())
			require.NoError(t, err, "test: %s", name)

			// since each asset swap can have up to sdk.OneInt() error, our expected error bound is 1*numAssets
			correctnessThreshold := sdk.OneInt().Mul(sdk.NewInt(int64(len(p.PoolLiquidity))))

			tokenOutAmount, err := cfmm_common.SwapAllCoinsToSingleAsset(&p, ctx, exitTokens, tc.tokenIn.Denom)
			require.True(t, tokenOutAmount.LTE(tc.tokenIn.Amount))
			require.True(t, tc.expectedOut.Sub(tokenOutAmount).Abs().LTE(correctnessThreshold))
		})
	}
}
