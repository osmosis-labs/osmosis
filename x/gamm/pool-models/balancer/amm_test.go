package balancer_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// allowedErrRatio is the maximal multiplicative difference in either
// direction (positive or negative) that we accept to tolerate in
// unit tests for calcuating the number of shares to be returned by
// joining a pool. The comparison is done between Wolfram estimates and our AMM logic.
const allowedErrRatio = "0.0000001"

// This test sets up 2 asset pools, and then checks the spot price on them.
// It uses the pools spot price method, rather than the Gamm keepers spot price method.
func (suite *KeeperTestSuite) TestBalancerSpotPrice() {
	baseDenom := "uosmo"
	quoteDenom := "uion"

	tests := []struct {
		name                string
		baseDenomPoolInput  sdk.Coin
		quoteDenomPoolInput sdk.Coin
		expectError         bool
		expectedOutput      sdk.Dec
	}{
		{
			name:                "equal value",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 100),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("1"),
		},
		{
			name:                "1:2 ratio",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 200),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("0.500000000000000000"),
		},
		{
			name:                "2:1 ratio",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 200),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 100),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("2.000000000000000000"),
		},
		{
			name:                "rounding after sigfig ratio",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 220),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 115),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("1.913043480000000000"), // ans is 1.913043478260869565, rounded is 1.91304348
		},
		{
			name:                "check number of sig figs",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 300),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("0.333333330000000000"),
		},
		{
			name:                "check number of sig figs high sizes",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 343569192534),
			quoteDenomPoolInput: sdk.NewCoin(quoteDenom, sdk.MustNewDecFromStr("186633424395479094888742").TruncateInt()),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("0.000000000001840877"),
		},
	}

	for _, tc := range tests {
		suite.SetupTest()

		poolId := suite.PrepareUni2PoolWithAssets(
			tc.baseDenomPoolInput,
			tc.quoteDenomPoolInput,
		)

		pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
		suite.Require().NoError(err, "test: %s", tc.name)
		balancerPool, isPool := pool.(*balancer.Pool)
		suite.Require().True(isPool, "test: %s", tc.name)

		spotPrice, err := balancerPool.SpotPrice(
			suite.Ctx,
			tc.baseDenomPoolInput.Denom,
			tc.quoteDenomPoolInput.Denom)

		if tc.expectError {
			suite.Require().Error(err, "test: %s", tc.name)
		} else {
			suite.Require().NoError(err, "test: %s", tc.name)
			suite.Require().True(spotPrice.Equal(tc.expectedOutput),
				"test: %s\nSpot price wrong, got %s, expected %s\n", tc.name,
				spotPrice, tc.expectedOutput)
		}
	}
}

// TestCalculateAmountOutAndIn_InverseRelationship tests that the same amount of token is guaranteed upon
// sequential operation of CalcInAmtGivenOut and CalcOutAmtGivenIn.
func TestCalculateAmountOutAndIn_InverseRelationship(t *testing.T) {
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
				ctx := createTestContext(t)

				poolAssetOut := balancer.PoolAsset{
					Token:  sdk.NewInt64Coin(tc.denomOut, tc.initialPoolOut),
					Weight: sdk.NewInt(tc.initialWeightOut),
				}

				poolAssetIn := balancer.PoolAsset{
					Token:  sdk.NewInt64Coin(tc.denomIn, tc.initialPoolIn),
					Weight: sdk.NewInt(tc.initialWeightIn),
				}

				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				require.NoError(t, err)

				exitFeeDec, err := sdk.NewDecFromStr("0")
				require.NoError(t, err)

				pool := createTestPool(t, swapFeeDec, exitFeeDec, poolAssetOut, poolAssetIn)
				require.NotNil(t, pool)

				initialOut := sdk.NewInt64Coin(poolAssetOut.Token.Denom, tc.initialCalcOut)
				initialOutCoins := sdk.NewCoins(initialOut)

				actualTokenIn, err := pool.CalcInAmtGivenOut(ctx, initialOutCoins, poolAssetIn.Token.Denom, swapFeeDec)
				require.NoError(t, err)

				inverseTokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(actualTokenIn), poolAssetOut.Token.Denom, swapFeeDec)
				require.NoError(t, err)

				require.Equal(t, initialOut.Denom, inverseTokenOut.Denom)

				expected := initialOut.Amount.ToDec()
				actual := inverseTokenOut.Amount.ToDec()

				// allow a rounding error of up to 1 for this relation
				tol := sdk.NewDec(1)
				require.True(osmoutils.DecApproxEq(t, expected, actual, tol))
			})
		}
	}
}

func TestCalcSingleAssetInAndOut_InverseRelationship(t *testing.T) {
	type testcase struct {
		initialPoolOut   int64
		initialPoolIn    int64
		initialWeightOut int64
		tokenOut         int64
		initialWeightIn  int64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := []testcase{
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 50,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         50,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  50,
		},
		{
			initialPoolOut:   1_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   2_351_333,
			tokenOut:         7,
			initialWeightOut: 148,
			initialWeightIn:  57,
		},
		{
			initialPoolOut:   1_000,
			tokenOut:         25,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000,
			tokenOut:         26,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
	}

	swapFeeCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(tc testcase, swapFeeCase string) string {
		return fmt.Sprintf("initialPoolOut: %d, initialCalcOut: %d, initialWeightOut: %d, initialWeightIn: %d, swapFee: %s",
			tc.initialPoolOut,
			tc.tokenOut,
			tc.initialWeightOut,
			tc.initialWeightIn,
			swapFeeCase,
		)
	}

	for _, tc := range testcases {
		for _, swapFee := range swapFeeCases {
			t.Run(getTestCaseName(tc, swapFee), func(t *testing.T) {
				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				require.NoError(t, err)

				initialPoolBalanceOut := sdk.NewInt(tc.initialPoolOut)

				initialWeightOut := sdk.NewInt(tc.initialWeightOut)
				initialWeightIn := sdk.NewInt(tc.initialWeightIn)

				initialTotalShares := types.InitPoolSharesSupply.ToDec()
				initialCalcTokenOut := sdk.NewInt(tc.tokenOut)

				actualSharesOut := balancer.CalcPoolSharesOutGivenSingleAssetIn(
					initialPoolBalanceOut.ToDec(),
					initialWeightOut.ToDec().Quo(initialWeightOut.Add(initialWeightIn).ToDec()),
					initialTotalShares,
					initialCalcTokenOut.ToDec(),
					swapFeeDec,
				)

				inverseCalcTokenOut := balancer.CalcSingleAssetInGivenPoolSharesOut(
					initialPoolBalanceOut.Add(initialCalcTokenOut).ToDec(),
					initialWeightOut.ToDec().Quo(initialWeightOut.Add(initialWeightIn).ToDec()),
					initialTotalShares.Add(actualSharesOut),
					actualSharesOut,
					swapFeeDec,
				)

				tol := sdk.NewDec(1)
				require.True(osmoutils.DecApproxEq(t, initialCalcTokenOut.ToDec(), inverseCalcTokenOut, tol))
			})
		}
	}
}

func TestCalcJoinPoolShares(t *testing.T) {
	testCases := []struct {
		name         string
		swapFee      sdk.Dec
		poolAssets   []balancer.PoolAsset
		tokensIn     sdk.Coins
		expectErr    bool
		expectShares sdk.Int
		expectLiq    sdk.Coins
	}{
		{
			name:    "equal weights with zero swap fee",
			swapFee: sdk.MustNewDecFromStr("0"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectErr:    false,
			expectShares: sdk.NewInt(2499999968800),
			expectLiq:    sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		},
		{
			name:    "swap equal weights with zero swap fee",
			swapFee: sdk.MustNewDecFromStr("0"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 25_000),
			),
			expectErr: false,
// Raises liquidity perfectly by 25_000 / 1_000_000_000_000. Initial number of pool shares = 100 * 10**18 = 10**20
// Expected increase = liquidity_increase_ratio * initial number of pool shares = (25_000 / 1_000_000_000_000) * 10**20 = 2500000000000.0 = 2.5 * 10**12
			expectShares: sdk.NewInt(2.5e12),
			expectLiq: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 25_000),
			),
		},
		{
			name:    "swap equal weights with 0.001 swap fee",
			swapFee: sdk.MustNewDecFromStr("0.001"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 25_000),
			),
			expectErr: false,
			// TODO: Verify expectShares value using manual calculation using the
			// following as reference:
			// https://github.com/osmosis-labs/osmosis/blob/6125f81d1ad95c73e03d86c3d5e96ceef3f8c12f/x/gamm/pool-models/balancer/amm_test.go#L348-L357
			expectShares: sdk.NewInt(2500000000000),
			expectLiq: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 25_000),
			),
		},
		{
			name:    "equal weights with 0.001 swap fee",
			swapFee: sdk.MustNewDecFromStr("0.001"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectErr:    false,
			expectShares: sdk.NewInt(2498749968800),
			expectLiq:    sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		},
		{
			name:    "equal weights with 0.1 swap fee",
			swapFee: sdk.MustNewDecFromStr("0.1"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectErr:    false,
			expectShares: sdk.NewInt(2374999971800),
			expectLiq:    sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		},
		{
			name:    "equal weights with 0.99 swap fee",
			swapFee: sdk.MustNewDecFromStr("0.99"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectErr:    false,
			expectShares: sdk.NewInt(1262499992100),
			expectLiq:    sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.swapFee, sdk.MustNewDecFromStr("0"), tc.poolAssets...)

			shares, liquidity, err := pool.CalcJoinPoolShares(sdk.Context{}, tc.tokensIn, tc.swapFee)
			if tc.expectErr {
				require.Error(t, err)
				require.Equal(t, sdk.ZeroInt(), shares)
				require.Equal(t, sdk.NewCoins(), liquidity)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectShares, shares)
				require.Equal(t, tc.expectLiq, liquidity)
			}
		})
	}
}

func TestCalcSingleAssetJoin(t *testing.T) {
	testCases := []struct {
		name         string
		swapFee      sdk.Dec
		poolAssets   []balancer.PoolAsset
		tokenIn      sdk.Coin
		expectShares sdk.Int
	}{
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
			//
			// 2_499_999_968_750 = 100 * 10^18 * (( 1 + (50,000 / 1_000_000_000_000))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100000000000000000000*%28%281+%2B+%2850000%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2,499,999,968,750
			name:    "equal weights with zero swap fee",
			swapFee: sdk.MustNewDecFromStr("0"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokenIn:      sdk.NewInt64Coin("uosmo", 50_000),
			expectShares: sdk.NewInt(2_499_999_968_750),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 2_487_500_000_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.01) / 1_000_000_000_000))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2_487_500_000_000
			name:    "equal weights with swap fee of 0.01",
			swapFee: sdk.MustNewDecFromStr("0.01"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokenIn:      sdk.NewInt64Coin("uosmo", 50_000),
			expectShares: sdk.NewInt(2_487_500_000_000),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
			//
			// 4_159_722_200_000 = 100 * 10^18 * (( 1 + (50,000 / 1_000_000_000_000))^0.83 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 500 / (500 + 100) approx = 0.83
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28500+%2F+%28100+%2B+500%29%29%29+*+0%29%2F1000000000000%29%29%5E%28500+%2F+%28100+%2B+500%29%29+-+1%29
			// 	Simplified:  P_issued = 4_159_722_200_000
			name:    "token in weight is greater than the other token, with zero swap fee",
			swapFee: sdk.MustNewDecFromStr("0"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(500),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokenIn:      sdk.NewInt64Coin("uosmo", 50_000),
			expectShares: sdk.NewInt(4_166_666_649_306),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
			//
			// 4_159_722_200_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.83) * 0.01) / 1_000_000_000_000))^0.83 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 500 / (500 + 100) approx = 0.83
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28500+%2F+%28100+%2B+500%29%29%29+*+0.01%29%2F1000000000000%29%29%5E%28500+%2F+%28100+%2B+500%29%29+-+1%29
			// 	Simplified:  P_issued = 4_159_722_200_000
			name:    "token in weight is greater than the other token, with non-zero swap fee",
			swapFee: sdk.MustNewDecFromStr("0.01"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(500),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokenIn:      sdk.NewInt64Coin("uosmo", 50_000),
			expectShares: sdk.NewInt(4_159_722_200_000),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
			//
			// 833_333_315_972 = 100 * 10^18 * (( 1 + (50,000 / 1_000_000_000_000))^0.167 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 200 / (200 + 1000) approx = 0.167
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28200+%2F+%28200+%2B+1000%29%29%29+*+0%29%2F1000000000000%29%29%5E%28200+%2F+%28200+%2B+1000%29%29+-+1%29
			// 	Simplified:  P_issued = 833_333_315_972
			name:    "token in weight is smaller than the other token, with zero swap fee",
			swapFee: sdk.MustNewDecFromStr("0"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(200),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(1000),
				},
			},
			tokenIn:      sdk.NewInt64Coin("uosmo", 50_000),
			expectShares: sdk.NewInt(833_333_315_972),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
			//
			// 819_444_430_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.167) * 0.02) / 1_000_000_000_000))^0.167 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 200 / (200 + 1000) approx = 0.167
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28200+%2F+%28200+%2B+1000%29%29%29+*+0.02%29%2F1000000000000%29%29%5E%28200+%2F+%28200+%2B+1000%29%29+-+1%29
			// 	Simplified:  P_issued = 819_444_430_000
			name:    "token in weight is smaller than the other token, with non-zero swap fee",
			swapFee: sdk.MustNewDecFromStr("0.02"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(200),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(1000),
				},
			},
			tokenIn:      sdk.NewInt64Coin("uosmo", 50_000),
			expectShares: sdk.NewInt(819_444_430_000),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.swapFee, sdk.MustNewDecFromStr("0"), tc.poolAssets...)

			balancerPool, ok := pool.(*balancer.Pool)
			require.True(t, ok)

			// find pool asset in pool
			// must be in pool since weights get scaled in Balancer pool
			// constructor
			poolAssetIn, err := balancerPool.GetPoolAsset(tc.tokenIn.Denom)
			require.NoError(t, err)

			shares, err := balancerPool.CalcSingleAssetJoin(tc.tokenIn, tc.swapFee, poolAssetIn, pool.GetTotalShares())
			// It is impossible to set up a test case with error here so we omit it.
			require.NoError(t, err)
			assertExpectedSharesErrRatio(t, tc.expectShares, shares)
		})
	}
}

func TestRandomizedJoinPoolExitPoolInvariants(t *testing.T) {
	type testCase struct {
		initialTokensDenomIn  int64
		initialTokensDenomOut int64

		percentRatio int64

		numShares sdk.Int
	}

	const (
		denomOut = "denomOut"
		denomIn  = "denomIn"
	)

	now := time.Now().Unix()
	rng := rand.NewSource(now)
	t.Logf("Using random source of %d\n", now)

	// generate test case with randomized initial assets and join/exit ratio
	newCase := func() (tc *testCase) {
		tc = new(testCase)
		tc.initialTokensDenomIn = rng.Int63() % 1_000_000
		tc.initialTokensDenomOut = rng.Int63() % 1_000_000

		// 1%~100% of initial assets
		tc.percentRatio = rng.Int63()%100 + 1

		return tc
	}

	swapFeeDec, err := sdk.NewDecFromStr("0")
	require.NoError(t, err)

	exitFeeDec, err := sdk.NewDecFromStr("0")
	require.NoError(t, err)

	// create pool with randomized initial token amounts
	// and randomized ratio of join/exit
	createPool := func(tc *testCase) (pool *balancer.Pool) {
		poolAssetOut := balancer.PoolAsset{
			Token:  sdk.NewInt64Coin(denomOut, tc.initialTokensDenomOut),
			Weight: sdk.NewInt(5),
		}

		poolAssetIn := balancer.PoolAsset{
			Token:  sdk.NewInt64Coin(denomIn, tc.initialTokensDenomIn),
			Weight: sdk.NewInt(5),
		}

		pool = createTestPool(t, swapFeeDec, exitFeeDec, poolAssetOut, poolAssetIn).(*balancer.Pool)
		require.NotNil(t, pool)

		return pool
	}

	// joins with predetermined ratio
	joinPool := func(pool types.PoolI, tc *testCase) {
		tokensIn := sdk.Coins{
			sdk.NewInt64Coin(denomIn, tc.initialTokensDenomIn*tc.percentRatio/100),
			sdk.NewInt64Coin(denomOut, tc.initialTokensDenomOut*tc.percentRatio/100),
		}
		numShares, err := pool.JoinPool(sdk.Context{}, tokensIn, swapFeeDec)
		require.NoError(t, err)
		tc.numShares = numShares
	}

	// exits for same amount of shares minted
	exitPool := func(pool types.PoolI, tc *testCase) {
		_, err := pool.ExitPool(sdk.Context{}, tc.numShares, exitFeeDec)
		require.NoError(t, err)
	}

	invariantJoinExitInversePreserve := func(
		beforeCoins, afterCoins sdk.Coins,
		beforeShares, afterShares sdk.Int,
	) {
		// test token amount has been preserved
		require.True(t,
			!beforeCoins.IsAnyGT(afterCoins),
			"Coins has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
			beforeCoins, afterCoins,
		)
		// test share amount has been preserved
		require.True(t,
			beforeShares.Equal(afterShares),
			"Shares has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
			beforeShares, afterShares,
		)
	}

	testPoolInvariants := func() {
		tc := newCase()
		pool := createPool(tc)
		originalCoins, originalShares := pool.GetTotalPoolLiquidity(sdk.Context{}), pool.GetTotalShares()
		joinPool(pool, tc)
		exitPool(pool, tc)
		invariantJoinExitInversePreserve(
			originalCoins, pool.GetTotalPoolLiquidity(sdk.Context{}),
			originalShares, pool.GetTotalShares(),
		)
	}

	for i := 0; i < 1000; i++ {
		testPoolInvariants()
	}
}

// TestGetPoolAssetsByDenom tests if `GetPoolAssetsByDenom` succesfully creates a map of denom to pool asset
// given pool asset as parameter
func TestGetPoolAssetsByDenom(t *testing.T) {
	testCases := []struct {
		name                      string
		poolAssets                []balancer.PoolAsset
		expectedPoolAssetsByDenom map[string]balancer.PoolAsset

		err error
	}{
		{
			name:                      "zero pool assets",
			poolAssets:                []balancer.PoolAsset{},
			expectedPoolAssetsByDenom: make(map[string]balancer.PoolAsset),
		},
		{
			name: "one pool asset",
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			expectedPoolAssetsByDenom: map[string]balancer.PoolAsset{
				"uosmo": {
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
		},
		{
			name: "two pool assets",
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("atom", 123),
					Weight: sdk.NewInt(400),
				},
			},
			expectedPoolAssetsByDenom: map[string]balancer.PoolAsset{
				"uosmo": {
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				"atom": {
					Token:  sdk.NewInt64Coin("atom", 123),
					Weight: sdk.NewInt(400),
				},
			},
		},
		{
			name: "duplicate pool assets",
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uosmo", 123),
					Weight: sdk.NewInt(400),
				},
			},
			err: fmt.Errorf(balancer.ErrMsgFormatRepeatingPoolAssetsNotAllowed, "uosmo"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualPoolAssetsByDenom, err := balancer.GetPoolAssetsByDenom(tc.poolAssets)

			require.Equal(t, tc.err, err)

			if tc.err != nil {
				return
			}

			require.Equal(t, tc.expectedPoolAssetsByDenom, actualPoolAssetsByDenom)
		})
	}
}

func assertExpectedSharesErrRatio(t *testing.T, expectedShares, actualShares sdk.Int) {
	allowedErrRatioDec, err := sdk.NewDecFromStr(allowedErrRatio)
	require.NoError(t, err)

	errTolerance := osmoutils.ErrTolerance{
		MultiplicativeTolerance: allowedErrRatioDec,
	}

	require.Equal(
		t,
		0,
		errTolerance.Compare(expectedShares, actualShares),
		fmt.Sprintf("expectedShares: %d, actualShares: %d", expectedShares.Int64(), actualShares.Int64()))
}
