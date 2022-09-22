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

// multiAssetCFMMTestCase defines a testcase for TestCFMMInvariantTwoAssets
// and TestCFMMInvariantTwoAssetsDirect
type multiAssetCFMMTestCase struct {
	xReserve    osmomath.BigDec
	yReserve    osmomath.BigDec
	uReserve    osmomath.BigDec
	wSumSquares osmomath.BigDec
	yIn         osmomath.BigDec
	expectPanic bool
}

var multiAssetCFMMTestCases = map[string]multiAssetCFMMTestCase{
	// sanity checks
	"even 3-asset small pool, small input": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(100),
		// represents a 3-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(100),
		wSumSquares: osmomath.NewBigDec(10000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: false,
	},
	"even 3-asset medium pool, small input": {
		xReserve: osmomath.NewBigDec(100000),
		yReserve: osmomath.NewBigDec(100000),
		// represents a 3-asset pool with 100,000 in each reserve
		uReserve:    osmomath.NewBigDec(100000),
		wSumSquares: osmomath.NewBigDec(10000000000),
		yIn:         osmomath.NewBigDec(100),
		expectPanic: false,
	},
	"even 4-asset small pool, small input": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(100),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(10000),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: false,
	},
	"even 4-asset medium pool, small input": {
		xReserve: osmomath.NewBigDec(100000),
		yReserve: osmomath.NewBigDec(100000),
		// represents a 4-asset pool with 100,000 in each reserve
		uReserve:    osmomath.NewBigDec(10000000000),
		wSumSquares: osmomath.NewBigDec(20000000000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: false,
	},
	/* TODO: increase BigDec precision (36 -> 72) to be able to accommodate this
	"even 4-asset large pool, small input": {
		xReserve: osmomath.NewBigDec(100000000),
		yReserve: osmomath.NewBigDec(100000000),
		// represents a 4-asset pool with 100M in each reserve
		uReserve: osmomath.NewBigDec(10000000000000000),
		wSumSquares: osmomath.NewBigDec(20000000000000000),
		yIn: osmomath.NewBigDec(100),
		expectPanic: false,
	},
	*/

	// uneven pools
	"uneven 3-asset pool, even swap assets as pool minority": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(100),
		// the asset not being swapped has 100,000 token reserves (swap assets in pool minority)
		uReserve:    osmomath.NewBigDec(100000),
		wSumSquares: osmomath.NewBigDec(10000000000),
		yIn:         osmomath.NewBigDec(10),
		expectPanic: false,
	},
	"uneven 3-asset pool, uneven swap assets as pool minority, y > x": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(200),
		// the asset not being swapped has 100,000 token reserves (swap assets in pool minority)
		uReserve:    osmomath.NewBigDec(100000),
		wSumSquares: osmomath.NewBigDec(10000000000),
		yIn:         osmomath.NewBigDec(10),
		expectPanic: false,
	},
	"uneven 3-asset pool, uneven swap assets as pool minority, x > y": {
		xReserve: osmomath.NewBigDec(200),
		yReserve: osmomath.NewBigDec(100),
		// the asset not being swapped has 100,000 token reserves (swap assets in pool minority)
		uReserve:    osmomath.NewBigDec(100000),
		wSumSquares: osmomath.NewBigDec(10000000000),
		yIn:         osmomath.NewBigDec(10),
		expectPanic: false,
	},
	"uneven 3-asset pool, no round numbers": {
		xReserve: osmomath.NewBigDec(1178349),
		yReserve: osmomath.NewBigDec(8329743),
		// the asset not being swapped has 329,847 token reserves (swap assets in pool minority)
		uReserve:    osmomath.NewBigDec(329847),
		wSumSquares: osmomath.NewBigDec(329847 * 329847),
		yIn:         osmomath.NewBigDec(10),
		expectPanic: false,
	},
	"uneven 4-asset pool, small input and swap assets in pool minority": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(100),
		// the assets not being swapped have 100,000 token reserves each (swap assets in pool minority)
		uReserve:    osmomath.NewBigDec(10000000000),
		wSumSquares: osmomath.NewBigDec(20000000000),
		yIn:         osmomath.NewBigDec(10),
		expectPanic: false,
	},
	"uneven 4-asset pool, even swap assets in pool majority": {
		xReserve: osmomath.NewBigDec(100000),
		yReserve: osmomath.NewBigDec(100000),
		// the assets not being swapped have 100 token reserves each (swap assets in pool majority)
		uReserve:    osmomath.NewBigDec(10000),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(10),
		expectPanic: false,
	},
	"uneven 4-asset pool, uneven swap assets in pool majority, y > x": {
		xReserve: osmomath.NewBigDec(100000),
		yReserve: osmomath.NewBigDec(200000),
		// the assets not being swapped have 100 token reserves each (swap assets in pool majority)
		uReserve:    osmomath.NewBigDec(10000),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(10),
		expectPanic: false,
	},
	"uneven 4-asset pool, uneven swap assets in pool majority, y < x": {
		xReserve: osmomath.NewBigDec(200000),
		yReserve: osmomath.NewBigDec(100000),
		// the assets not being swapped have 100 token reserves each (swap assets in pool majority)
		uReserve:    osmomath.NewBigDec(10000),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(10),
		expectPanic: false,
	},
	"uneven 4-asset pool, no round numbers": {
		xReserve: osmomath.NewBigDec(1178349),
		yReserve: osmomath.NewBigDec(8329743),
		// the assets not being swapped have 329,847 tokens and 4,372,897 respectively
		uReserve:    osmomath.NewBigDec(329847 * 4372897),
		wSumSquares: osmomath.NewBigDec((329847 * 329847) + (4372897 * 4372897)),
		yIn:         osmomath.NewBigDec(10),
		expectPanic: false,
	},

	// panic catching
	"negative xReserve": {
		xReserve: osmomath.NewBigDec(-100),
		yReserve: osmomath.NewBigDec(100),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(200),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: true,
	},
	"negative yReserve": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(-100),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(200),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: true,
	},
	"negative uReserve": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(100),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(-200),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: true,
	},
	"negative sumSquares": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(100),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(200),
		wSumSquares: osmomath.NewBigDec(-20000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: true,
	},
	"negative yIn": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(100),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(200),
		wSumSquares: osmomath.NewBigDec(-20000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: true,
	},
	"input greater than pool reserves (even 4-asset pool)": {
		xReserve:    osmomath.NewBigDec(100),
		yReserve:    osmomath.NewBigDec(100),
		uReserve:    osmomath.NewBigDec(200),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(1000),
		expectPanic: true,
	},

	// overflows
	"xReserve overflows in 4-asset pool": {
		xReserve: osmomath.NewDecFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(1024), nil), big.NewInt(1))),
		yReserve: osmomath.NewBigDec(100),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(200),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: true,
	},
	"yReserve overflows in 4-asset pool": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewDecFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(1024), nil), big.NewInt(1))),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(200),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: true,
	},
	"uReserve overflows in 4-asset pool": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(100),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewDecFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(1024), nil), big.NewInt(1))),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: true,
	},
	"wSumSquares overflows in 4-asset pool": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(100),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(200),
		wSumSquares: osmomath.NewDecFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(1024), nil), big.NewInt(1))),
		yIn:         osmomath.NewBigDec(1),
		expectPanic: true,
	},
	"yIn overflows in 4-asset pool": {
		xReserve: osmomath.NewBigDec(100),
		yReserve: osmomath.NewBigDec(100),
		// represents a 4-asset pool with 100 in each reserve
		uReserve:    osmomath.NewBigDec(200),
		wSumSquares: osmomath.NewBigDec(20000),
		yIn:         osmomath.NewDecFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(1024), nil), big.NewInt(1))),
		expectPanic: true,
	},
}

type StableSwapTestSuite struct {
	test_helpers.CfmmCommonTestSuite
}

func TestCFMMInvariantTwoAssets(t *testing.T) {
	kErrTolerance := osmomath.OneDec()

	tests := map[string]struct {
		xReserve    osmomath.BigDec
		yReserve    osmomath.BigDec
		yIn         osmomath.BigDec
		expectPanic bool
	}{
		"small pool small input": {
			osmomath.NewBigDec(100),
			osmomath.NewBigDec(100),
			osmomath.NewBigDec(1),
			false,
		},
		"small pool large input": {
			osmomath.NewBigDec(100),
			osmomath.NewBigDec(100),
			osmomath.NewBigDec(1000),
			true,
		},
		"large pool large input": {
			osmomath.NewBigDec(1000000000),
			osmomath.NewBigDec(1000000000),
			osmomath.NewBigDec(1000),
			false,
		},

		// panic catching
		"xReserve negative": {
			osmomath.NewBigDec(-100),
			osmomath.NewBigDec(100),
			osmomath.NewBigDec(1),
			true,
		},
		"yReserve negative": {
			osmomath.NewBigDec(100),
			osmomath.NewBigDec(-100),
			osmomath.NewBigDec(1),
			true,
		},
		"yIn negative": {
			osmomath.NewBigDec(100),
			osmomath.NewBigDec(100),
			osmomath.NewBigDec(-1),
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				// using two-asset cfmm
				k0 := cfmmConstant(test.xReserve, test.yReserve)
				xOut := solveCFMMBinarySearch(cfmmConstant)(test.xReserve, test.yReserve, test.yIn)

				k1 := cfmmConstant(test.xReserve.Sub(xOut), test.yReserve.Add(test.yIn))
				osmomath.DecApproxEq(t, k0, k1, kErrTolerance)

				// using multi-asset cfmm (should be equivalent with u = 1, w = 0)
				k2 := cfmmConstantMulti(test.xReserve, test.yReserve, osmomath.OneDec(), osmomath.ZeroDec())
				osmomath.DecApproxEq(t, k2, k0, kErrTolerance)
				xOut2 := solveCFMMBinarySearchMulti(cfmmConstantMulti)(test.xReserve, test.yReserve, osmomath.OneDec(), osmomath.ZeroDec(), test.yIn)
				k3 := cfmmConstantMulti(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), osmomath.OneDec(), osmomath.ZeroDec())
				osmomath.DecApproxEq(t, k2, k3, kErrTolerance)
			}

			osmoassert.ConditionalPanic(t, test.expectPanic, sut)
		})
	}
}

func TestCFMMInvariantMultiAssets(t *testing.T) {
	kErrTolerance := osmomath.OneDec()

	// TODO: switch solveCfmmMulti to binary search and replace this with test case suite
	tests := map[string]multiAssetCFMMTestCase{}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				// using multi-asset cfmm
				k2 := cfmmConstantMulti(test.xReserve, test.yReserve, test.uReserve, test.wSumSquares)
				xOut2 := solveCfmmMulti(test.xReserve, test.yReserve, test.wSumSquares, test.yIn)
				k3 := cfmmConstantMulti(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), test.uReserve, test.wSumSquares)
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
				// using multi-asset cfmm
				k2 := cfmmConstantMulti(test.xReserve, test.yReserve, test.uReserve, test.wSumSquares)
				xOut2 := solveCFMMBinarySearchMulti(cfmmConstantMulti)(test.xReserve, test.yReserve, test.uReserve, test.wSumSquares, test.yIn)
				k3 := cfmmConstantMulti(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), test.uReserve, test.wSumSquares)
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
