package stableswap

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/internal/test_helpers"
)

type StableSwapTestSuite struct {
	test_helpers.CfmmCommonTestSuite
}

func TestCFMMInvariantTwoAssets(t *testing.T) {
	kErrTolerance := sdk.OneDec()

	tests := map[string]struct {
		xReserve    sdk.Dec
		yReserve    sdk.Dec
		yIn         sdk.Dec
		expectPanic bool
	}{
		"small pool small input": {
			sdk.NewDec(100),
			sdk.NewDec(100),
			sdk.NewDec(1),
			false,
		},
		"small pool large input": {
			sdk.NewDec(100),
			sdk.NewDec(100),
			sdk.NewDec(1000),
			false,
		},
		// This test fails due to a bug in our original solver
		// {
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(10000),
		// },

		// panic catching
		"xReserve negative": {
			sdk.NewDec(-100),
			sdk.NewDec(100),
			sdk.NewDec(1),
			true,
		},
		"yReserve negative": {
			sdk.NewDec(100),
			sdk.NewDec(-100),
			sdk.NewDec(1),
			true,
		},
		"yIn negative": {
			sdk.NewDec(100),
			sdk.NewDec(100),
			sdk.NewDec(-1),
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				// using two-asset cfmm
				k0 := cfmmConstant(test.xReserve, test.yReserve)
				xOut := solveCfmm(test.xReserve, test.yReserve, test.yIn)

				k1 := cfmmConstant(test.xReserve.Sub(xOut), test.yReserve.Add(test.yIn))
				osmoassert.DecApproxEq(t, k0, k1, kErrTolerance)

				// using multi-asset cfmm (should be equivalent with u = 1, w = 0)
				k2 := cfmmConstantMulti(test.xReserve, test.yReserve, sdk.OneDec(), sdk.ZeroDec())
				osmoassert.DecApproxEq(t, k2, k0, kErrTolerance)
				xOut2 := solveCfmmMulti(test.xReserve, test.yReserve, sdk.ZeroDec(), test.yIn)
				k3 := cfmmConstantMulti(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), sdk.OneDec(), sdk.ZeroDec())
				osmoassert.DecApproxEq(t, k2, k3, kErrTolerance)
			}

			osmoassert.ConditionalPanic(t, test.expectPanic, sut)
		})
	}
}

func TestCFMMInvariantMultiAssets(t *testing.T) {
	kErrTolerance := sdk.OneDec()

	tests := map[string]struct {
		xReserve    sdk.Dec
		yReserve    sdk.Dec
		uReserve    sdk.Dec
		wSumSquares sdk.Dec
		yIn         sdk.Dec
		expectPanic bool
	}{
		"4-asset pool, small input": {
			sdk.NewDec(100),
			sdk.NewDec(100),
			// represents a 4-asset pool with 100 in each reserve
			sdk.NewDec(200),
			sdk.NewDec(20000),
			sdk.NewDec(1),
			false,
		},
		"4-asset pool, large input": {
			sdk.NewDec(100),
			sdk.NewDec(100),
			sdk.NewDec(200),
			sdk.NewDec(20000),
			sdk.NewDec(1000),
			false,
		},
		// This test fails due to a bug in our original solver
		// "large pool, large input": {
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(10000),
		// },

		// panic catching
		"negative xReserve": {
			sdk.NewDec(-100),
			sdk.NewDec(100),
			// represents a 4-asset pool with 100 in each reserve
			sdk.NewDec(200),
			sdk.NewDec(20000),
			sdk.NewDec(1),
			true,
		},
		"negative yReserve": {
			sdk.NewDec(100),
			sdk.NewDec(-100),
			// represents a 4-asset pool with 100 in each reserve
			sdk.NewDec(200),
			sdk.NewDec(20000),
			sdk.NewDec(1),
			true,
		},
		"negative uReserve": {
			sdk.NewDec(100),
			sdk.NewDec(100),
			// represents a 4-asset pool with 100 in each reserve
			sdk.NewDec(-200),
			sdk.NewDec(20000),
			sdk.NewDec(1),
			true,
		},
		"negative sumSquares": {
			sdk.NewDec(100),
			sdk.NewDec(100),
			// represents a 4-asset pool with 100 in each reserve
			sdk.NewDec(200),
			sdk.NewDec(-20000),
			sdk.NewDec(1),
			true,
		},
		"negative yIn": {
			sdk.NewDec(100),
			sdk.NewDec(100),
			// represents a 4-asset pool with 100 in each reserve
			sdk.NewDec(200),
			sdk.NewDec(-20000),
			sdk.NewDec(1),
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				// using multi-asset cfmm
				k2 := cfmmConstantMulti(test.xReserve, test.yReserve, test.uReserve, test.wSumSquares)
				xOut2 := solveCfmmMulti(test.xReserve, test.yReserve, test.wSumSquares, test.yIn)
				k3 := cfmmConstantMulti(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), test.uReserve, test.wSumSquares)
				osmoassert.DecApproxEq(t, k2, k3, kErrTolerance)
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
