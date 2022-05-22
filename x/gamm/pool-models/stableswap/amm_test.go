package stableswap

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/internal/test_helpers"
)

type StableSwapTestSuite struct {
	test_helpers.CfmmCommonTestSuite
}

// Replace with https://github.com/cosmos/cosmos-sdk/blob/master/types/decimal.go#L892-L895
// once our SDK branch is up to date with it
func decApproxEq(t *testing.T, exp sdk.Dec, actual sdk.Dec, errTolerance sdk.Dec) {
	// We want |exp - actual| < errTolerance
	diff := exp.Sub(actual).Abs()
	require.True(t, diff.LTE(errTolerance), "expected %s, got %s, maximum errTolerance %s", exp, actual, errTolerance)
}

func TestCFMMInvariantTwoAssets(t *testing.T) {
	kErrTolerance := sdk.OneDec()

	tests := []struct {
		xReserve sdk.Dec
		yReserve sdk.Dec
		yIn      sdk.Dec
	}{
		{
			sdk.NewDec(100),
			sdk.NewDec(100),
			sdk.NewDec(1),
		},
		{
			sdk.NewDec(100),
			sdk.NewDec(100),
			sdk.NewDec(1000),
		},
		// {
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(10000),
		// },
	}

	for _, test := range tests {
		// using two-asset cfmm
		k0 := cfmmConstant(test.xReserve, test.yReserve)
		xOut := solveCfmm(test.xReserve, test.yReserve, test.yIn)

		k1 := cfmmConstant(test.xReserve.Sub(xOut), test.yReserve.Add(test.yIn))
		decApproxEq(t, k0, k1, kErrTolerance)

		// using multi-asset cfmm (should be equivalent with u = 1, w = 0)
		k2 := cfmmConstantMulti(test.xReserve, test.yReserve, sdk.OneDec(), sdk.ZeroDec())
		decApproxEq(t, k2, k0, kErrTolerance)
		xOut2 := solveCfmmMulti(test.xReserve, test.yReserve, sdk.ZeroDec(), test.yIn)
		fmt.Println(xOut2)
		k3 := cfmmConstantMulti(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), sdk.OneDec(), sdk.ZeroDec())
		decApproxEq(t, k2, k3, kErrTolerance)
	}
}

func TestCFMMInvariantMultiAssets(t *testing.T) {
	kErrTolerance := sdk.OneDec()

	tests := []struct {
		xReserve    sdk.Dec
		yReserve    sdk.Dec
		uReserve    sdk.Dec
		wSumSquares sdk.Dec
		yIn         sdk.Dec
	}{
		{
			sdk.NewDec(100),
			sdk.NewDec(100),
			// represents a 4-asset pool with 100 in each reserve
			sdk.NewDec(200),
			sdk.NewDec(20000),
			sdk.NewDec(1),
		},
		{
			sdk.NewDec(100),
			sdk.NewDec(100),
			sdk.NewDec(200),
			sdk.NewDec(20000),
			sdk.NewDec(1000),
		},
		// {
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(10000),
		// },
	}

	for _, test := range tests {
		// using multi-asset cfmm
		k2 := cfmmConstantMulti(test.xReserve, test.yReserve, test.uReserve, test.wSumSquares)
		xOut2 := solveCfmmMulti(test.xReserve, test.yReserve, test.wSumSquares, test.yIn)
		fmt.Println(xOut2)
		k3 := cfmmConstantMulti(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), test.uReserve, test.wSumSquares)
		decApproxEq(t, k2, k3, kErrTolerance)
	}
}

func (suite *StableSwapTestSuite) Test_StableSwap_CalculateAmountOutAndIn_InverseRelationship(t *testing.T) {
	type testcase struct {
		denomOut         string
		initialPoolOut   int64
		scalingFactorOut int64
		initialCalcOut   int64

		denomIn         string
		initialPoolIn   int64
		scalingFactorIn int64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := []testcase{
		{
			denomOut:         usdcDenom,
			initialPoolOut:   1_000_000_000_000,
			scalingFactorOut: 100,
			initialCalcOut:   100,

			denomIn:         usdbDenom,
			initialPoolIn:   1_000_000_000_000,
			scalingFactorIn: 100,
		},
		{
			denomOut:         usdcDenom,
			initialPoolOut:   1_000,
			scalingFactorOut: 100,
			initialCalcOut:   100,

			denomIn:         usdbDenom,
			initialPoolIn:   1_000_000,
			scalingFactorIn: 100,
		},
		{
			denomOut:         usdcDenom,
			initialPoolOut:   1_000,
			scalingFactorOut: 100,
			initialCalcOut:   100,

			denomIn:         usdbDenom,
			initialPoolIn:   1_000_000,
			scalingFactorIn: 100,
		},
		{
			denomOut:         usdcDenom,
			initialPoolOut:   1_000,
			scalingFactorOut: 200,
			initialCalcOut:   100,

			denomIn:         usdbDenom,
			initialPoolIn:   1_000_000,
			scalingFactorIn: 50,
		},
		{
			denomOut:         usdcDenom,
			initialPoolOut:   1_000_000,
			scalingFactorOut: 200,
			initialCalcOut:   100000,

			denomIn:         usdbDenom,
			initialPoolIn:   1_000_000_000,
			scalingFactorIn: 50,
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
				ctx := createTestContext(t)

				poolAssetOut := stableswap.PoolAsset{
					Token:         sdk.NewCoin(tc.denomOut, tc.initialPoolOut),
					ScalingFactor: sdk.NewInt(tc.scalingFactorOut),
				}

				poolAssetIn := stableswap.PoolAsset{
					Token:         sdk.NewCoin(tc.denomIn, tc.initialPoolIn),
					ScalingFactor: sdk.NewInt(tc.scalingFactorIn),
				}

				//poolLiquidity := sdk.NewCoins(poolLiquidityIn, poolLiquidityOut)

				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				require.NoError(t, err)

				exitFeeDec, err := sdk.NewDecFromStr("0")
				require.NoError(t, err)

				pool := createTestPool(t, []stableswap.PoolAsset{
					poolAssetOut,
					poolAssetIn,
				},
					swapFeeDec,
					exitFeeDec,
				)

				require.NotNil(t, pool)

				suite.TestCalculateAmountOutAndIn_InverseRelationship(ctx, pool, poolAssetIn.Token.Denom, poolAssetOut.Token.Denom, tc.initialCalcOut, swapFeeDec)
			})
		}
	}
}
