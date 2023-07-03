package balancer_test

import (
	"errors"
	"fmt"
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/internal/test_helpers"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

var (
	defaultSpreadFactor       = sdk.MustNewDecFromStr("0.025")
	defaultZeroExitFee        = sdk.ZeroDec()
	defaultPoolId             = uint64(10)
	defaultBalancerPoolParams = balancer.PoolParams{
		SwapFee: defaultSpreadFactor,
		ExitFee: defaultZeroExitFee,
	}
	defaultFutureGovernor = ""
	defaultCurBlockTime   = time.Unix(1618700000, 0)
	//
	dummyPoolAssets = []balancer.PoolAsset{}
	wantErr         = true
	noErr           = false
)

// TestUpdateIntermediaryPoolAssetsLiquidity tests if `updateIntermediaryPoolAssetsLiquidity` returns poolAssetsByDenom map
// with the updated liquidity given by the parameter
func TestUpdateIntermediaryPoolAssetsLiquidity(t *testing.T) {
	const (
		uosmoValueOriginal = 1_000_000_000_000
		atomValueOriginal  = 123
		ionValueOriginal   = 657

		// Weight does not affect calculations so it is shared
		weight = 100
	)
	testCases := []struct {
		name         string
		newLiquidity sdk.Coins
		poolAssets   map[string]balancer.PoolAsset
		expectPass   bool
		err          error
	}{
		{
			name: "regular case with multiple pool assets and a subset of newLiquidity to update",
			newLiquidity: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 1_000),
				sdk.NewInt64Coin("atom", 2_000),
				sdk.NewInt64Coin("ion", 3_000)),
			poolAssets: map[string]balancer.PoolAsset{
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
			},
			expectPass: true,
		},
		{
			name:         "new liquidity has no coins",
			newLiquidity: sdk.NewCoins(),
			poolAssets: map[string]balancer.PoolAsset{
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
			},
			expectPass: true,
		},
		{
			name: "newLiquidity has a coin that poolAssets don't",
			newLiquidity: sdk.NewCoins(
				sdk.NewInt64Coin("juno", 1_000)),
			poolAssets: map[string]balancer.PoolAsset{
				"uosmo": {
					Token:  sdk.NewInt64Coin("uosmo", uosmoValueOriginal),
					Weight: sdk.NewInt(weight),
				},
			},
			expectPass: false,
			err:        fmt.Errorf(balancer.ErrMsgFormatFailedInterimLiquidityUpdate, "juno"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedPoolAssetsByDenom := map[string]balancer.PoolAsset{}
			for denom, asset := range tc.poolAssets {
				expectedValue := asset
				expectedValue.Token.Amount = expectedValue.Token.Amount.Add(tc.newLiquidity.AmountOf(denom))
				expectedPoolAssetsByDenom[denom] = expectedValue
			}

			err := balancer.UpdateIntermediaryPoolAssetsLiquidity(tc.newLiquidity, tc.poolAssets)

			if tc.expectPass {
				require.NoError(t, tc.err, "test: %v", tc.name)
				// make sure actual pool assets are properly updated
				require.Equal(t, expectedPoolAssetsByDenom, tc.poolAssets)
			} else {
				require.Error(t, tc.err, "test: %v", tc.name)
				require.Equal(t, tc.err, err)
				require.Equal(t, expectedPoolAssetsByDenom, tc.poolAssets)
			}
			return
		})
	}
}

func TestCalcSingleAssetJoin(t *testing.T) {
	for _, tc := range calcSingleAssetJoinTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.spreadFactor, sdk.MustNewDecFromStr("0"), tc.poolAssets...)

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
			poolAssetIn, err := pool.GetPoolAsset(poolAssetInDenom)
			require.NoError(t, err)

			// system under test
			sut := func() {
				shares, err := pool.CalcSingleAssetJoin(tokenIn, tc.spreadFactor, poolAssetIn, pool.GetTotalShares())

				if tc.expErr != nil {
					require.Error(t, err)
					require.ErrorAs(t, tc.expErr, &err)
					require.Equal(t, sdk.ZeroInt(), shares)
					return
				}

				require.NoError(t, err)
				assertExpectedSharesErrRatio(t, tc.expectShares, shares)
			}

			assertPoolStateNotModified(t, pool, func() {
				osmoassert.ConditionalPanic(t, tc.expectPanic, sut)
			})
		})
	}
}

func TestCalcJoinSingleAssetTokensIn(t *testing.T) {
	testCases := []struct {
		name           string
		spreadFactor   sdk.Dec
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
			name:         "one token in - equal weights with zero spread factor",
			spreadFactor: sdk.MustNewDecFromStr("0"),
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
			name:         "two tokens in - equal weights with zero spread factor",
			spreadFactor: sdk.MustNewDecFromStr("0"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin("uatom", 50_000)),
			expectShares: sdk.NewInt(2_499_999_968_750 * 2),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with spreadFactorRatio added:
			// P_issued = P_supply * ((1 + (A_t * spreadFactorRatio  / B_t))^W_t - 1)
			//
			// 2_487_500_000_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.01) / 1e12))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	spreadFactorRatio = (1 - (1 - W_t) * spreadFactor)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2_487_500_000_000
			name:         "one token in - equal weights with spread factor of 0.01",
			spreadFactor: sdk.MustNewDecFromStr("0.01"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectShares: sdk.NewInt(2_487_500_000_000),
		},
		{
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with spreadFactorRatio added:
			// P_issued = P_supply * ((1 + (A_t * spreadFactorRatio  / B_t))^W_t - 1)
			//
			// 2_487_500_000_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.5) * 0.01) / 1e12))^0.5 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	spreadFactorRatio = (1 - (1 - W_t) * spreadFactor)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 2_487_500_000_000
			name:         "two tokens in - equal weights with spread factor of 0.01",
			spreadFactor: sdk.MustNewDecFromStr("0.01"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin("uatom", 50_000)),
			expectShares: sdk.NewInt(2_487_500_000_000 * 2),
		},
		{
			// For uosmo:
			//
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with spreadFactorRatio added:
			// P_issued = P_supply * ((1 + (A_t * spreadFactorRatio  / B_t))^W_t - 1)
			//
			// 2_072_912_400_000_000 = 1e20 * (( 1 + (50,000 * (1 - (1 - 0.83) * 0.03) / 2_000_000_000))^0.83 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 2_000_000_000
			//	W_t = normalized weight of deposited asset in pool = 500 / 500 + 100 = 0.83
			// 	spreadFactorRatio = (1 - (1 - W_t) * spreadFactor)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28500+%2F+%28500+%2B+100%29%29%29+*+0.03%29%2F2000000000%29%29%5E%28500+%2F+%28500+%2B+100%29%29+-+1%29
			// 	Simplified:  P_issued = 2_072_912_400_000_000
			//
			//
			// For uatom:
			//
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) with on page 10
			// with spreadFactorRatio added:
			// P_issued = P_supply * ((1 + (A_t * spreadFactorRatio  / B_t))^W_t - 1)
			//
			// 1_624_999_900_000 = 1e20 * (( 1 + (100_000 * (1 - (1 - 0.167) * 0.03) / 1e12))^0.167 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 1e20
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 100 / 500 + 100 = 0.167
			// 	spreadFactorRatio = (1 - (1 - W_t) * spreadFactor)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%28100000*%281+-+%281-%28100+%2F+%28500+%2B+100%29%29%29+*+0.03%29%2F1000000000000%29%29%5E%28100+%2F+%28500+%2B+100%29%29+-+1%29
			// 	Simplified:  P_issued = 1_624_999_900_000
			name:         "two varying tokens in, varying weights, with spread factor of 0.03",
			spreadFactor: sdk.MustNewDecFromStr("0.03"),
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
			spreadFactor: sdk.MustNewDecFromStr("0.03"),
			poolAssets:   oneTrillionEvenPoolAssets,
			tokensIn:     sdk.NewCoins(),
			expectShares: sdk.NewInt(0),
		},
		{
			name:         "one of the tokensIn asset does not exist in pool",
			spreadFactor: sdk.ZeroDec(),
			poolAssets:   oneTrillionEvenPoolAssets,
			// Second tokenIn does not exist.
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000), sdk.NewInt64Coin(doesNotExistDenom, 50_000)),
			expectShares: sdk.ZeroInt(),
			expErr:       fmt.Errorf(balancer.ErrMsgFormatNoPoolAssetFound, doesNotExistDenom),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.spreadFactor, sdk.ZeroDec(), tc.poolAssets...)

			poolAssetsByDenom, err := balancer.GetPoolAssetsByDenom(pool.GetAllPoolAssets())
			require.NoError(t, err)

			// estimate expected liquidity
			expectedNewLiquidity := sdk.NewCoins()
			for _, tokenIn := range tc.tokensIn {
				expectedNewLiquidity = expectedNewLiquidity.Add(tokenIn)
			}

			sut := func() {
				totalNumShares, totalNewLiquidity, err := pool.CalcJoinSingleAssetTokensIn(tc.tokensIn, pool.GetTotalShares(), poolAssetsByDenom, tc.spreadFactor)

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

			assertPoolStateNotModified(t, pool, sut)
		})
	}
}

// TestGetPoolAssetsByDenom tests if `GetPoolAssetsByDenom` successfully creates a map of denom to pool asset
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
func (suite *BalancerTestSuite) TestBalancerCalculateAmountOutAndIn_InverseRelationship() {
	type testcase struct {
		denomOut         string
		initialPoolOut   int64
		initialWeightOut int64
		initialCalcOut   int64

		denomIn         string
		initialPoolIn   int64
		initialWeightIn int64
	}

	// For every test case in testcases, apply a spread factor in spreadFactorCases.
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

	spreadFactorCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(tc testcase, spreadFactorCase string) string {
		return fmt.Sprintf("tokenOutInitial: %d, tokenInInitial: %d, initialOut: %d, spreadFactor: %s",
			tc.initialPoolOut,
			tc.initialPoolIn,
			tc.initialCalcOut,
			spreadFactorCase,
		)
	}

	for _, tc := range testcases {
		for _, spreadFactor := range spreadFactorCases {
			suite.Run(getTestCaseName(tc, spreadFactor), func() {
				ctx := suite.CreateTestContext()

				poolAssetOut := balancer.PoolAsset{
					Token:  sdk.NewInt64Coin(tc.denomOut, tc.initialPoolOut),
					Weight: sdk.NewInt(tc.initialWeightOut),
				}

				poolAssetIn := balancer.PoolAsset{
					Token:  sdk.NewInt64Coin(tc.denomIn, tc.initialPoolIn),
					Weight: sdk.NewInt(tc.initialWeightIn),
				}

				spreadFactorDec, err := sdk.NewDecFromStr(spreadFactor)
				suite.Require().NoError(err)

				exitFeeDec, err := sdk.NewDecFromStr("0")
				suite.Require().NoError(err)

				pool := createTestPool(suite.T(), spreadFactorDec, exitFeeDec, poolAssetOut, poolAssetIn)
				suite.Require().NotNil(pool)

				errTolerance := osmomath.ErrTolerance{
					AdditiveTolerance: sdk.OneDec(), MultiplicativeTolerance: sdk.Dec{},
				}
				sut := func() {
					test_helpers.TestCalculateAmountOutAndIn_InverseRelationship(suite.T(), ctx, pool, poolAssetIn.Token.Denom, poolAssetOut.Token.Denom, tc.initialCalcOut, spreadFactorDec, errTolerance)
				}

				assertPoolStateNotModified(suite.T(), pool, sut)
			})
		}
	}
}

func TestCalcSingleAssetInAndOut_InverseRelationship(t *testing.T) {
	type testcase struct {
		initialPoolOut   int64
		initialWeightOut int64
		tokenOut         int64
		initialWeightIn  int64
	}

	// For every test case in testcases, apply a spread factor in spreadFactorCases.
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

	spreadFactorCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(tc testcase, spreadFactorDec string) string {
		return fmt.Sprintf("initialPoolOut: %d, initialCalcOut: %d, initialWeightOut: %d, initialWeightIn: %d, spreadFactor: %s",
			tc.initialPoolOut,
			tc.tokenOut,
			tc.initialWeightOut,
			tc.initialWeightIn,
			spreadFactorDec,
		)
	}

	for _, tc := range testcases {
		for _, spreadFactor := range spreadFactorCases {
			t.Run(getTestCaseName(tc, spreadFactor), func(t *testing.T) {
				spreadFactorDec, err := sdk.NewDecFromStr(spreadFactor)
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
					spreadFactorDec,
				)

				inverseCalcTokenOut := balancer.CalcSingleAssetInGivenPoolSharesOut(
					initialPoolBalanceOut.Add(initialCalcTokenOut).ToDec(),
					initialWeightOut.ToDec().Quo(initialWeightOut.Add(initialWeightIn).ToDec()),
					initialTotalShares.Add(actualSharesOut),
					actualSharesOut,
					spreadFactorDec,
				)

				tol := sdk.NewDec(1)
				osmoassert.DecApproxEq(t, initialCalcTokenOut.ToDec(), inverseCalcTokenOut, tol)
			})
		}
	}
}

// Expected is un-scaled
func testTotalWeight(t *testing.T, expected sdk.Int, pool balancer.Pool) {
	t.Helper()
	scaledExpected := expected.MulRaw(balancer.GuaranteedWeightPrecision)
	require.Equal(t,
		scaledExpected.String(),
		pool.GetTotalWeight().String())
}

// TODO: Refactor this into multiple tests
func TestBalancerPoolUpdatePoolAssetBalance(t *testing.T) {
	var poolId uint64 = 10

	initialAssets := []balancer.PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
	}

	pacc, err := balancer.NewBalancerPool(poolId, defaultBalancerPoolParams, initialAssets, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	_, err = pacc.GetPoolAsset("unknown")
	require.Error(t, err)
	_, err = pacc.GetPoolAsset("")
	require.Error(t, err)

	testTotalWeight(t, sdk.NewInt(300), pacc)

	// Break abstractions and start reasoning about the underlying internal representation's APIs.
	// TODO: This test actually just needs to be refactored to not be doing this, and just
	// create a different pool each time.

	err = pacc.SetInitialPoolAssets([]balancer.PoolAsset{{
		Weight: sdk.NewInt(-1),
		Token:  sdk.NewCoin("negativeWeight", sdk.NewInt(50000)),
	}})

	require.Error(t, err)

	err = pacc.SetInitialPoolAssets([]balancer.PoolAsset{{
		Weight: sdk.NewInt(0),
		Token:  sdk.NewCoin("zeroWeight", sdk.NewInt(50000)),
	}})
	require.Error(t, err)

	err = pacc.UpdatePoolAssetBalance(
		sdk.NewCoin("test1", sdk.NewInt(0)))
	require.Error(t, err)

	err = pacc.UpdatePoolAssetBalance(
		sdk.Coin{Denom: "test1", Amount: sdk.NewInt(-1)},
	)
	require.Error(t, err)

	err = pacc.UpdatePoolAssetBalance(
		sdk.NewCoin("test1", sdk.NewInt(1)))
	require.NoError(t, err)

	testTotalWeight(t, sdk.NewInt(300), pacc)

	PoolAsset, err := pacc.GetPoolAsset("test1")
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(1).String(), PoolAsset.Token.Amount.String())
}

func TestBalancerPoolAssetsWeightAndTokenBalance(t *testing.T) {
	// TODO: Add more cases
	// asset names should be i ascending order, starting from test1
	tests := []struct {
		assets    []balancer.PoolAsset
		shouldErr bool
	}{
		// weight 0
		{
			[]balancer.PoolAsset{
				{
					Weight: sdk.NewInt(0),
					Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
				},
			},
			wantErr,
		},
		// negative weight
		{
			[]balancer.PoolAsset{
				{
					Weight: sdk.NewInt(-1),
					Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
				},
			},
			wantErr,
		},
		// 0 token amount
		{
			[]balancer.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test1", sdk.NewInt(0)),
				},
			},
			wantErr,
		},
		// negative token amount
		{
			[]balancer.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token: sdk.Coin{
						Denom:  "test1",
						Amount: sdk.NewInt(-1),
					},
				},
			},
			wantErr,
		},
		// total weight 300
		{
			[]balancer.PoolAsset{
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
				},
			},
			noErr,
		},
		// two of the same token
		{
			[]balancer.PoolAsset{
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
				},
				{
					Weight: sdk.NewInt(300),
					Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test2", sdk.NewInt(10000)),
				},
			},
			wantErr,
		},
		// total weight 7300
		{
			[]balancer.PoolAsset{
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
				},
				{
					Weight: sdk.NewInt(7000),
					Token:  sdk.NewCoin("test3", sdk.NewInt(10000)),
				},
			},
			noErr,
		},
	}

	var poolId uint64 = 10

	for i, tc := range tests {
		pacc, err := balancer.NewBalancerPool(poolId, defaultBalancerPoolParams, tc.assets, defaultFutureGovernor, defaultCurBlockTime)
		if tc.shouldErr {
			require.Error(t, err, "unexpected lack of error, tc %v", i)
		} else {
			require.NoError(t, err, "unexpected error, tc %v", i)
			expectedTotalWeight := sdk.ZeroInt()
			for i, asset := range tc.assets {
				expectedTotalWeight = expectedTotalWeight.Add(asset.Weight)

				// Ensure pool assets are sorted
				require.Equal(t, "test"+fmt.Sprint(i+1), pacc.PoolAssets[i].Token.Denom)
			}
			testTotalWeight(t, expectedTotalWeight, pacc)
		}
	}
}

// TODO: Figure out what parts of this test, if any, make sense.
func TestGetBalancerPoolAssets(t *testing.T) {
	// Adds []PoolAssets, one after another
	// if the addition doesn't error, adds the weight of the pool assets to a running total,
	// and ensures the pool's total weight is equal to the expected.
	// This also ensures that the pool assets remain sorted within the pool.
	// Furthermore, it ensures that GetPoolAsset succeeds for everything in the pool,
	// and fails for things not in it.
	denomNotInPool := "xyzCoin"

	assets := []balancer.PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
		},
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test3", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test4", sdk.NewInt(10000)),
		},
	}

	// TODO: We need way more robust test cases here, and should table drive these cases
	pacc, err := balancer.NewBalancerPool(defaultPoolId, defaultBalancerPoolParams, assets, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	// Hardcoded GetPoolAssets tests.
	assets, err = pacc.GetPoolAssets("test1", "test2")
	require.NoError(t, err)
	require.Equal(t, 2, len(assets))

	assets, err = pacc.GetPoolAssets("test1", "test2", "test3", "test4")
	require.NoError(t, err)
	require.Equal(t, 4, len(assets))

	_, err = pacc.GetPoolAssets("test1", "test5")
	require.Error(t, err)
	_, err = pacc.GetPoolAssets(denomNotInPool)
	require.Error(t, err)

	assets, err = pacc.GetPoolAssets()
	require.NoError(t, err)
	require.Equal(t, 0, len(assets))
}

func TestLBPParamsEmptyStartTime(t *testing.T) {
	// Test that when the start time is empty, the pool
	// sets its start time to be the first start time it is called on
	defaultDuration := 100 * time.Second

	initialPoolAssets := []balancer.PoolAsset{
		{
			Weight: sdk.NewInt(1),
			Token:  sdk.NewCoin("asset1", sdk.NewInt(1000)),
		},
		{
			Weight: sdk.NewInt(1),
			Token:  sdk.NewCoin("asset2", sdk.NewInt(1000)),
		},
	}

	params := balancer.SmoothWeightChangeParams{
		Duration: defaultDuration,
		TargetPoolWeights: []balancer.PoolAsset{
			{
				Weight: sdk.NewInt(1),
				Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
			},
			{
				Weight: sdk.NewInt(2),
				Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
			},
		},
	}

	pacc, err := balancer.NewBalancerPool(defaultPoolId, balancer.PoolParams{
		SmoothWeightChangeParams: &params,
		SwapFee:                  defaultSpreadFactor,
		ExitFee:                  defaultZeroExitFee,
	}, initialPoolAssets, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	// Consistency check that SmoothWeightChangeParams params are set
	require.NotNil(t, pacc.PoolParams.SmoothWeightChangeParams)
	// Ensure that the start time got set
	require.Equal(t, pacc.PoolParams.SmoothWeightChangeParams.StartTime, defaultCurBlockTime)
}

func TestBalancerPoolPokeTokenWeights(t *testing.T) {
	// Set default date
	defaultStartTime := time.Unix(1618703511, 0)
	defaultStartTimeUnix := defaultStartTime.Unix()
	defaultDuration := 100 * time.Second
	floatGuaranteedPrecision := float64(balancer.GuaranteedWeightPrecision)

	// testCases don't need to be ordered by time. but the blockTime should be
	// less than the end time of the SmoothWeightChange. Testing past the end time
	// is already handled.
	type testCase struct {
		blockTime       time.Time
		expectedWeights []sdk.Int
	}

	// Tests how the pool weights get updated via PokeTokenWeights at different block times.
	// The framework underneath will automatically add tests for times before the start time,
	// at the start time, at the end time, and after the end time. It is up to the test writer to
	// test the behavior at times in-between.
	tests := []struct {
		// We take the initial weights from here
		params balancer.SmoothWeightChangeParams
		cases  []testCase
	}{
		{
			// 1:1 pool, between asset1 and asset2
			// transitioning to a 1:2 pool
			params: balancer.SmoothWeightChangeParams{
				StartTime: defaultStartTime,
				Duration:  defaultDuration,
				InitialPoolWeights: []balancer.PoolAsset{
					{
						Weight: sdk.NewInt(1),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(1),
						Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
					},
				},
				TargetPoolWeights: []balancer.PoolAsset{
					{
						Weight: sdk.NewInt(1),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(2),
						Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
					},
				},
			},
			cases: []testCase{
				{
					// Halfway through at 50 seconds elapsed
					blockTime: time.Unix(defaultStartTimeUnix+50, 0),
					expectedWeights: []sdk.Int{
						sdk.NewInt(1 * balancer.GuaranteedWeightPrecision),
						// Halfway between 1 & 2
						sdk.NewInt(3 * balancer.GuaranteedWeightPrecision / 2),
					},
				},
				{
					// Quarter way through at 25 seconds elapsed
					blockTime: time.Unix(defaultStartTimeUnix+25, 0),
					expectedWeights: []sdk.Int{
						sdk.NewInt(1 * balancer.GuaranteedWeightPrecision),
						// Quarter way between 1 & 2 = 1.25
						sdk.NewInt(int64(1.25 * floatGuaranteedPrecision)),
					},
				},
			},
		},
		{
			// 2:2 pool, between asset1 and asset2
			// transitioning to a 4:1 pool
			params: balancer.SmoothWeightChangeParams{
				StartTime: defaultStartTime,
				Duration:  defaultDuration,
				InitialPoolWeights: []balancer.PoolAsset{
					{
						Weight: sdk.NewInt(2),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(2),
						Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
					},
				},
				TargetPoolWeights: []balancer.PoolAsset{
					{
						Weight: sdk.NewInt(4),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(1),
						Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
					},
				},
			},
			cases: []testCase{
				{
					// Halfway through at 50 seconds elapsed
					blockTime: time.Unix(defaultStartTimeUnix+50, 0),
					expectedWeights: []sdk.Int{
						// Halfway between 2 & 4
						sdk.NewInt(6 * balancer.GuaranteedWeightPrecision / 2),
						// Halfway between 1 & 2
						sdk.NewInt(3 * balancer.GuaranteedWeightPrecision / 2),
					},
				},
				{
					// Quarter way through at 25 seconds elapsed
					blockTime: time.Unix(defaultStartTimeUnix+25, 0),
					expectedWeights: []sdk.Int{
						// Quarter way between 2 & 4 = 2.5
						sdk.NewInt(int64(2.5 * floatGuaranteedPrecision)),
						// Quarter way between 2 & 1 = 1.75
						sdk.NewInt(int64(1.75 * floatGuaranteedPrecision)),
					},
				},
			},
		},
	}

	// Add test cases at a time before the start, the start, the end, and a time after the end.
	addDefaultCases := func(params balancer.SmoothWeightChangeParams, cases []testCase) []testCase {
		// Set times one second before the start, and one second after the end
		timeBeforeWeightChangeStart := time.Unix(params.StartTime.Unix()-1, 0)
		timeAtWeightChangeEnd := params.StartTime.Add(params.Duration)
		timeAfterWeightChangeEnd := time.Unix(timeAtWeightChangeEnd.Unix()+1, 0)
		initialWeights := make([]sdk.Int, len(params.InitialPoolWeights))
		finalWeights := make([]sdk.Int, len(params.TargetPoolWeights))
		for i, v := range params.InitialPoolWeights {
			initialWeights[i] = v.Weight.MulRaw(balancer.GuaranteedWeightPrecision)
		}
		for i, v := range params.TargetPoolWeights {
			// Doesn't need to be scaled, due to this being done already in param initialization,
			// and because params is only shallow copied
			finalWeights[i] = v.Weight
		}
		// Set the test cases for times before the start, and the start
		updatedCases := []testCase{
			{
				blockTime:       timeBeforeWeightChangeStart,
				expectedWeights: initialWeights,
			},
			{
				blockTime:       params.StartTime,
				expectedWeights: initialWeights,
			},
		}
		// Append the provided cases
		updatedCases = append(updatedCases, cases...)
		finalCases := []testCase{
			{
				blockTime:       timeAtWeightChangeEnd,
				expectedWeights: finalWeights,
			},
			{
				blockTime:       timeAfterWeightChangeEnd,
				expectedWeights: finalWeights,
			},
		}
		// Append the final cases
		updatedCases = append(updatedCases, finalCases...)
		return updatedCases
	}

	for poolId, tc := range tests {
		paramsCopy := tc.params
		// First we create the initial pool assets we will use
		initialPoolAssets := make([]balancer.PoolAsset, len(paramsCopy.InitialPoolWeights))
		for i, asset := range paramsCopy.InitialPoolWeights {
			assetCopy := balancer.PoolAsset{
				Weight: asset.Weight,
				Token:  sdk.NewInt64Coin(asset.Token.Denom, 10000),
			}
			initialPoolAssets[i] = assetCopy
		}
		// Initialize the pool
		pacc, err := balancer.NewBalancerPool(uint64(poolId), balancer.PoolParams{
			SwapFee:                  defaultSpreadFactor,
			ExitFee:                  defaultZeroExitFee,
			SmoothWeightChangeParams: &tc.params,
		}, initialPoolAssets, defaultFutureGovernor, defaultCurBlockTime)
		require.NoError(t, err, "poolId %v", poolId)

		// Consistency check that SmoothWeightChangeParams params are set
		require.NotNil(t, pacc.PoolParams.SmoothWeightChangeParams)

		testCases := addDefaultCases(paramsCopy, tc.cases)
		for caseNum, testCase := range testCases {
			pacc.PokePool(testCase.blockTime)

			totalWeight := sdk.ZeroInt()

			for assetNum, asset := range pacc.GetAllPoolAssets() {
				require.Equal(t, testCase.expectedWeights[assetNum], asset.Weight,
					"Didn't get the expected weights, poolId %v, caseNumber %v, assetNumber %v",
					poolId, caseNum, assetNum)

				totalWeight = totalWeight.Add(asset.Weight)
			}

			require.Equal(t, totalWeight, pacc.GetTotalWeight())
		}
		// Should have been deleted by the last test case of after PokeTokenWeights pokes past end time.
		require.Nil(t, pacc.PoolParams.SmoothWeightChangeParams)
	}
}

// This test (currently trivially) checks to make sure that `IsActive` returns true for balancer pools.
// This is mainly to make sure that if IsActive is ever used as an emergency switch, it is not accidentally left off for any (or all) pools.
func TestIsActive(t *testing.T) {
	tests := map[string]struct {
		expectedIsActive bool
	}{
		"IsActive is true": {
			expectedIsActive: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}

			// Initialize a pool
			pool, err := balancer.NewBalancerPool(defaultPoolId, defaultBalancerPoolParams, dummyPoolAssets, defaultFutureGovernor, defaultCurBlockTime)
			require.NoError(t, err, "test %v", name)

			isActive := pool.IsActive(ctx)
			require.Equal(t, tc.expectedIsActive, isActive)
		})
	}
}

func TestCalcJoinPoolNoSwapShares(t *testing.T) {
	balancerPoolAssets := []balancer.PoolAsset{
		{Token: sdk.NewInt64Coin("foo", 100), Weight: sdk.NewIntFromUint64(5)},
		{Token: sdk.NewInt64Coin("bar", 100), Weight: sdk.NewIntFromUint64(5)},
	}

	balancerThreePoolAssets := []balancer.PoolAsset{
		{Token: sdk.NewInt64Coin("foo", 100), Weight: sdk.NewIntFromUint64(5)},
		{Token: sdk.NewInt64Coin("bar", 100), Weight: sdk.NewIntFromUint64(5)},
		{Token: sdk.NewInt64Coin("baz", 100), Weight: sdk.NewIntFromUint64(5)},
	}

	tests := map[string]struct {
		tokensIn        sdk.Coins
		poolAssets      []balancer.PoolAsset
		expNumShare     sdk.Int
		expTokensJoined sdk.Coins
		expPoolAssets   []balancer.PoolAsset
		expectPass      bool
	}{
		"two asset pool, same tokenIn ratio": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10)), sdk.NewCoin("bar", sdk.NewInt(10))),
			poolAssets:      balancerPoolAssets,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10)), sdk.NewCoin("bar", sdk.NewInt(10))),
			expPoolAssets:   balancerPoolAssets,
			expectPass:      true,
		},
		"two asset pool, different tokenIn ratio with pool": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10)), sdk.NewCoin("bar", sdk.NewInt(11))),
			poolAssets:      balancerPoolAssets,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10)), sdk.NewCoin("bar", sdk.NewInt(10))),
			expPoolAssets:   balancerPoolAssets,
			expectPass:      true,
		},
		"three asset pool, same tokenIn ratio": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10)), sdk.NewCoin("bar", sdk.NewInt(10)), sdk.NewCoin("baz", sdk.NewInt(10))),
			poolAssets:      balancerThreePoolAssets,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10)), sdk.NewCoin("bar", sdk.NewInt(10)), sdk.NewCoin("baz", sdk.NewInt(10))),
			expPoolAssets:   balancerThreePoolAssets,
			expectPass:      true,
		},
		"three asset pool, different tokenIn ratio with pool": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10)), sdk.NewCoin("bar", sdk.NewInt(10)), sdk.NewCoin("baz", sdk.NewInt(11))),
			poolAssets:      balancerThreePoolAssets,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10)), sdk.NewCoin("bar", sdk.NewInt(10)), sdk.NewCoin("baz", sdk.NewInt(10))),
			expPoolAssets:   balancerThreePoolAssets,
			expectPass:      true,
		},
		"two asset pool, no-swap join attempt with one asset": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10))),
			poolAssets:      balancerPoolAssets,
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   balancerPoolAssets,
			expectPass:      false,
		},
		"two asset pool, no-swap join attempt with one valid and one invalid asset": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10)), sdk.NewCoin("baz", sdk.NewInt(10))),
			poolAssets:      balancerPoolAssets,
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   balancerPoolAssets,
			expectPass:      false,
		},
		"two asset pool, no-swap join attempt with two invalid assets": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("baz", sdk.NewInt(10)), sdk.NewCoin("qux", sdk.NewInt(10))),
			poolAssets:      balancerPoolAssets,
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   balancerPoolAssets,
			expectPass:      false,
		},
		"three asset pool, no-swap join attempt with an invalid asset": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10)), sdk.NewCoin("bar", sdk.NewInt(10)), sdk.NewCoin("qux", sdk.NewInt(10))),
			poolAssets:      balancerThreePoolAssets,
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   balancerThreePoolAssets,
			expectPass:      false,
		},
		"single asset pool, no-swap join attempt with one asset": {
			tokensIn: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(sdk.MaxSortableDec.TruncateInt64()))),
			poolAssets: []balancer.PoolAsset{
				{Token: sdk.NewCoin("foo", sdk.NewInt(1)), Weight: sdk.NewIntFromUint64(1)},
			},
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets: []balancer.PoolAsset{
				{Token: sdk.NewCoin("foo", sdk.NewInt(1)), Weight: sdk.NewIntFromUint64(1)},
			},
			expectPass: false,
		},
		"duplicate asset pool, no-swap join attempt with duplicate assets": {
			tokensIn: sdk.Coins{sdk.NewCoin("foo", sdk.NewInt(1)), sdk.NewCoin("foo", sdk.NewInt(1))},
			poolAssets: []balancer.PoolAsset{
				{Token: sdk.NewCoin("foo", sdk.NewInt(100)), Weight: sdk.NewIntFromUint64(1)},
				{Token: sdk.NewCoin("foo", sdk.NewInt(100)), Weight: sdk.NewIntFromUint64(1)},
			},
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets: []balancer.PoolAsset{
				{Token: sdk.NewCoin("foo", sdk.NewInt(100)), Weight: sdk.NewIntFromUint64(1)},
				{Token: sdk.NewCoin("foo", sdk.NewInt(100)), Weight: sdk.NewIntFromUint64(1)},
			},
			expectPass: false,
		},
		"attempt joining pool with no assets in it": {
			tokensIn:        sdk.Coins{sdk.NewCoin("foo", sdk.NewInt(1)), sdk.NewCoin("foo", sdk.NewInt(1))},
			poolAssets:      []balancer.PoolAsset{},
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   []balancer.PoolAsset{},
			expectPass:      false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			balancerPool := balancer.Pool{
				Address:            poolmanagertypes.NewPoolAddress(defaultPoolId).String(),
				Id:                 defaultPoolId,
				PoolParams:         balancer.PoolParams{SwapFee: defaultSpreadFactor, ExitFee: defaultZeroExitFee},
				PoolAssets:         test.poolAssets,
				FuturePoolGovernor: defaultFutureGovernor,
				TotalShares:        sdk.NewCoin(types.GetPoolShareDenom(defaultPoolId), types.InitPoolSharesSupply),
			}

			numShare, tokensJoined, err := balancerPool.CalcJoinPoolNoSwapShares(ctx, test.tokensIn, balancerPool.GetSpreadFactor(ctx))

			if test.expectPass {
				require.NoError(t, err)
				require.Equal(t, test.expPoolAssets, balancerPool.PoolAssets)
				require.Equal(t, test.expNumShare, numShare)
				require.Equal(t, test.expTokensJoined, tokensJoined)
			} else {
				require.Error(t, err)
				require.Equal(t, test.expPoolAssets, balancerPool.PoolAssets)
				require.Equal(t, test.expNumShare, numShare)
				require.Equal(t, test.expTokensJoined, tokensJoined)
			}
		})
	}
}
