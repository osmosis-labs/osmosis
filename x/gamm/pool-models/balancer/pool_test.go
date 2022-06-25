package balancer_test

import (
	"errors"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// TestUpdateIntermediaryPoolAssetsLiquidity tests if `updateIntermediaryPoolAssetsLiquidity` returns poolAssetsByDenom map
// with the updated liquidity given by the parameter
func TestUpdateIntermediaryPoolAssetsLiquidity(t *testing.T) {
	testCases := []struct {
		name string

		// returns newLiquidity, originalPoolAssetsByDenom, expectedPoolAssetsByDenom
		setup func() (sdk.Coins, map[string]balancer.PoolAsset, map[string]balancer.PoolAsset)

		err error
	}{
		{
			name: "regular case with multiple pool assets and a subset of newLiquidity to update",

			setup: func() (sdk.Coins, map[string]balancer.PoolAsset, map[string]balancer.PoolAsset) {
				const (
					uosmoValueOriginal = 1_000_000_000_000
					atomValueOriginal  = 123
					ionValueOriginal   = 657

					// Weight does not affect calculations so it is shared
					weight = 100
				)

				newLiquidity := sdk.NewCoins(
					sdk.NewInt64Coin("uosmo", 1_000),
					sdk.NewInt64Coin("atom", 2_000),
					sdk.NewInt64Coin("ion", 3_000))

				originalPoolAssetsByDenom := map[string]balancer.PoolAsset{
					"uosmo": {
						Token:  sdk.NewInt64Coin("uosmo", uosmoValueOriginal),
						Weight: sdk.NewInt(weight),
					},
					"atom": {
						Token:  sdk.NewInt64Coin("atom", atomValueOriginal),
						Weight: sdk.NewInt(weight),
					},
					"ion": {
						Token:  sdk.NewInt64Coin("ion", ionValueOriginal),
						Weight: sdk.NewInt(weight),
					},
				}

				expectedPoolAssetsByDenom := map[string]balancer.PoolAsset{}
				for k, v := range originalPoolAssetsByDenom {
					expectedValue := balancer.PoolAsset{Token: v.Token, Weight: v.Weight}
					expectedValue.Token.Amount = expectedValue.Token.Amount.Add(newLiquidity.AmountOf(k))
					expectedPoolAssetsByDenom[k] = expectedValue
				}

				return newLiquidity, originalPoolAssetsByDenom, expectedPoolAssetsByDenom
			},
		},
		{
			name: "new liquidity has no coins",

			setup: func() (sdk.Coins, map[string]balancer.PoolAsset, map[string]balancer.PoolAsset) {
				const (
					uosmoValueOriginal = 1_000_000_000_000
					atomValueOriginal  = 123
					ionValueOriginal   = 657

					// Weight does not affect calculations so it is shared
					weight = 100
				)

				newLiquidity := sdk.NewCoins()

				originalPoolAssetsByDenom := map[string]balancer.PoolAsset{
					"uosmo": {
						Token:  sdk.NewInt64Coin("uosmo", uosmoValueOriginal),
						Weight: sdk.NewInt(weight),
					},
					"atom": {
						Token:  sdk.NewInt64Coin("atom", atomValueOriginal),
						Weight: sdk.NewInt(weight),
					},
					"ion": {
						Token:  sdk.NewInt64Coin("ion", ionValueOriginal),
						Weight: sdk.NewInt(weight),
					},
				}

				return newLiquidity, originalPoolAssetsByDenom, originalPoolAssetsByDenom
			},
		},
		{
			name: "newLiquidity has a coin that poolAssets don't",

			setup: func() (sdk.Coins, map[string]balancer.PoolAsset, map[string]balancer.PoolAsset) {
				const (
					uosmoValueOriginal = 1_000_000_000_000

					// Weight does not affect calculations so it is shared
					weight = 100
				)

				newLiquidity := sdk.NewCoins(
					sdk.NewInt64Coin("juno", 1_000))

				originalPoolAssetsByDenom := map[string]balancer.PoolAsset{
					"uosmo": {
						Token:  sdk.NewInt64Coin("uosmo", uosmoValueOriginal),
						Weight: sdk.NewInt(weight),
					},
				}

				return newLiquidity, originalPoolAssetsByDenom, originalPoolAssetsByDenom
			},

			err: fmt.Errorf(balancer.ErrMsgFormatFailedInterimLiquidityUpdate, "juno"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newLiquidity, originalPoolAssetsByDenom, expectedPoolAssetsByDenom := tc.setup()

			err := balancer.UpdateIntermediaryPoolAssetsLiquidity(newLiquidity, originalPoolAssetsByDenom)

			require.Equal(t, tc.err, err)

			if tc.err != nil {
				return
			}

			require.Equal(t, expectedPoolAssetsByDenom, originalPoolAssetsByDenom)
		})
	}
}

func TestCalcSingleAssetJoin(t *testing.T) {
	for _, tc := range calcSingleAssetJoinTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.swapFee, sdk.MustNewDecFromStr("0"), tc.poolAssets...)

			balancerPool, ok := pool.(*balancer.Pool)
			require.True(t, ok)

			tokenIn := tc.tokensIn[0]

			poolAssetInDenom := tokenIn.Denom
			// when testing a case with tokenIn that does not exist in pool, we just want
			// to provide any pool asset.
			if tc.expErr != nil && errors.Is(tc.expErr, types.ErrDenomNotFoundInPool) {
				poolAssetInDenom = tc.poolAssets[0].Token.Denom
			}

			// find pool asset in pool
			// must be in pool since weights get scaled in Balancer pool
			// constructor
			poolAssetIn, err := balancerPool.GetPoolAsset(poolAssetInDenom)
			require.NoError(t, err)

			// system under test
			sut := func() {
				shares, err := balancerPool.CalcSingleAssetJoin(tokenIn, tc.swapFee, poolAssetIn, pool.GetTotalShares())

				if tc.expErr != nil {
					require.Error(t, err)
					require.ErrorAs(t, tc.expErr, &err)
					require.Equal(t, sdk.ZeroInt(), shares)
					return
				}

				require.NoError(t, err)
				assertExpectedSharesErrRatio(t, tc.expectShares, shares)
			}

			assertPoolStateNotModified(t, balancerPool, func() {
				assertPanic(t, tc.expectPanic, sut)
			})
		})
	}
}

func TestCalcJoinSingleAssetTokensIn(t *testing.T) {
	testCases := []struct {
		name           string
		swapFee        sdk.Dec
		poolAssets     []balancer.PoolAsset
		tokensIn       sdk.Coins
		expectShares   sdk.Int
		expectLiqudity sdk.Coins
		expErr         error
	}{
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
			//
			// 2_499_999_968_750 = 1e20 * (( 1 + (50,000 / 1e12))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100000000000000000000*%28%281+%2B+%2850000%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2,499,999,968,750
			name:         "one token in - equal weights with zero swap fee",
			swapFee:      sdk.MustNewDecFromStr("0"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectShares: sdk.NewInt(2_499_999_968_750),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t / B_t))^W_t - 1)
			//
			// 2_499_999_968_750 = 1e20 * (( 1 + (50,000 / 1e12))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100000000000000000000*%28%281+%2B+%2850000%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2,499,999,968,750
			name:         "two tokens in - equal weights with zero swap fee",
			swapFee:      sdk.MustNewDecFromStr("0"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin("uatom", 50_000)),
			expectShares: sdk.NewInt(2_499_999_968_750 * 2),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with swapFeeRatio added:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 2_487_500_000_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.01) / 1e12))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2_487_500_000_000
			name:         "one token in - equal weights with swap fee of 0.01",
			swapFee:      sdk.MustNewDecFromStr("0.01"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectShares: sdk.NewInt(2_487_500_000_000),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with swapFeeRatio added:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 2_487_500_000_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.01) / 1e12))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2_487_500_000_000
			name:         "two tokens in - equal weights with swap fee of 0.01",
			swapFee:      sdk.MustNewDecFromStr("0.01"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin("uatom", 50_000)),
			expectShares: sdk.NewInt(2_487_500_000_000 * 2),
		},
		{
			// For uosmo:
			//
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with swapFeeRatio added:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 2_072_912_400_000_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.83) * 0.03) / 2_000_000_000))^0.83 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 2_000_000_000
			//	W_t = normalized weight of deposited asset in pool = 500 / 500 + 100 = 0.83
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28500+%2F+%28500+%2B+100%29%29%29+*+0.03%29%2F2000000000%29%29%5E%28500+%2F+%28500+%2B+100%29%29+-+1%29
			// 	Simplified:  P_issued = 2_072_912_400_000_000
			//
			//
			// For uatom:
			//
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with swapFeeRatio added:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 1_624_999_900_000 = 1e20 * (( 1 + (100_000 * (1 - (1 - 0.167) * 0.03) / 1e12))^0.167 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 100 / 500 + 100 = 0.167
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%28100000*%281+-+%281-%28100+%2F+%28500+%2B+100%29%29%29+*+0.03%29%2F1000000000000%29%29%5E%28100+%2F+%28500+%2B+100%29%29+-+1%29
			// 	Simplified:  P_issued = 1_624_999_900_000
			name:    "two varying tokens in, varying weights, with swap fee of 0.03",
			swapFee: sdk.MustNewDecFromStr("0.03"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 2_000_000_000),
					Weight: sdk.NewInt(500),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1e12),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin("uatom", 100_000)),
			expectShares: sdk.NewInt(2_072_912_400_000_000 + 1_624_999_900_000),
		},
		{
			name:         "no tokens in",
			swapFee:      sdk.MustNewDecFromStr("0.03"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(),
			expectShares: sdk.NewInt(0),
		},
		{
			name:       "one of the tokensIn asset does not exist in pool",
			swapFee:    sdk.ZeroDec(),
			poolAssets: oneTrillionEvenPoolAssets,
			// Second tokenIn does not exist.
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin(doesNotExistDenom, 50_000)),
			expectShares: sdk.ZeroInt(),
			expErr:       fmt.Errorf(balancer.ErrMsgFormatNoPoolAssetFound, doesNotExistDenom),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.swapFee, sdk.ZeroDec(), tc.poolAssets...)

			balancerPool, ok := pool.(*balancer.Pool)
			require.True(t, ok)

			poolAssetsByDenom, err := balancer.GetPoolAssetsByDenom(balancerPool.GetAllPoolAssets())
			require.NoError(t, err)

			// estimate expected liquidity
			expectedNewLiquidity := sdk.NewCoins()
			for _, tokenIn := range tc.tokensIn {
				expectedNewLiquidity = expectedNewLiquidity.Add(tokenIn)
			}

			sut := func() {
				totalNumShares, totalNewLiquidity, err := balancerPool.CalcJoinSingleAssetTokensIn(tc.tokensIn, pool.GetTotalShares(), poolAssetsByDenom, tc.swapFee)

				if tc.expErr != nil {
					require.Error(t, err)
					require.ErrorAs(t, tc.expErr, &err)
					require.Equal(t, sdk.ZeroInt(), totalNumShares)
					require.Equal(t, sdk.Coins{}, totalNewLiquidity)
					return
				}

				require.NoError(t, err)

				require.Equal(t, expectedNewLiquidity, totalNewLiquidity)

				if tc.expectShares.Int64() == 0 {
					require.Equal(t, tc.expectShares, totalNumShares)
					return
				}

				assertExpectedSharesErrRatio(t, tc.expectShares, totalNumShares)
			}

			assertPoolStateNotModified(t, balancerPool, sut)
		})
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
					Token:  sdk.NewInt64Coin("uosmo", 1e12),
					Weight: sdk.NewInt(100),
				},
			},
			expectedPoolAssetsByDenom: map[string]balancer.PoolAsset{
				"uosmo": {
					Token:  sdk.NewInt64Coin("uosmo", 1e12),
					Weight: sdk.NewInt(100),
				},
			},
		},
		{
			name: "two pool assets",
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1e12),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("atom", 123),
					Weight: sdk.NewInt(400),
				},
			},
			expectedPoolAssetsByDenom: map[string]balancer.PoolAsset{
				"uosmo": {
					Token:  sdk.NewInt64Coin("uosmo", 1e12),
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
					Token:  sdk.NewInt64Coin("uosmo", 1e12),
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

				sut := func() {
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
				}

				balancerPool, ok := pool.(*balancer.Pool)
				require.True(t, ok)

				assertPoolStateNotModified(t, balancerPool, sut)
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
