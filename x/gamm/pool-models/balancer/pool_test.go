package balancer_test

import (
	"errors"
	"fmt"
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var (
	defaultSwapFee            = sdk.MustNewDecFromStr("0.025")
	defaultExitFee            = sdk.MustNewDecFromStr("0.025")
	defaultPoolId             = uint64(10)
	defaultBalancerPoolParams = balancer.PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
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
				err = validateExpectedSharesErrRatio(t, tc.expectShares, shares)
				if tc.expectToleranceError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}

			assertPoolStateNotModified(t, balancerPool, func() {
				osmoutils.ConditionalPanic(t, tc.expectPanic, sut)
			})
		})
	}
}

func TestCalcJoinSingleAssetTokensIn(t *testing.T) {
	testCases := []calcJoinSharesTestCase{
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
			// Compares expected shares and actual shares to make sure the multiplicativeTolerance ratio below is maintained
			// Expected shares were determined based on the amount of tokensIn, the higher the tokens the more the shares
			// |a - b| / min(a, b) <= allowedErrRatio
			// 0.199999996 = abs(2_999_999_955_000-2_499_999_968_800)/2_499_999_968_800
			// 0.199999996 <= 0.0000001 (NOT WITHIN ACCEPTED RATIO)
			//
			// where:
			// a = expected shares,
			// b = actual shares
			// allowedErrRatio = 0.0000001
			// 	Full solution: https://www.wolframalpha.com/input?i=abs%282999999955000-2499999968800%29%2F2499999968800
			name:                 "Error Tolerance: One token in, expected > actual",
			swapFee:              sdk.MustNewDecFromStr("0"),
			poolAssets:           oneTrillionEvenPoolAssets,
			tokensIn:             sdk.NewCoins(sdk.NewInt64Coin("uosmo", 60_000)),
			expectShares:         sdk.NewInt(2_599_999_968_750), // ACTUAL: 2_999_999_955_000
			expectToleranceError: true,
		},
		{
			// Compares expected shares and actual shares to make sure the multiplicativeTolerance ratio below is maintained
			// Expected shares were determined based on the amount of tokensIn, the higher the tokens the more the shares
			// |a - b| / min(a, b) <= allowedErrRatio
			// 0.1538461677 = abs(5_999_999_999_999-5_199_999_937_500)/5_199_999_937_500
			// 0.1538461677 <= 0.0000001 (NOT WITHIN ACCEPTED RATIO)
			//
			// where:
			// a = expected shares,
			// b = actual shares
			// allowedErrRatio = 0.0000001
			// Full Solution: https://www.wolframalpha.com/input?i=abs%285999999999999-5199999937500%29%2F5199999937500
			name:                 "Error Tolerance: Two token in, expected > actual",
			swapFee:              sdk.MustNewDecFromStr("0"),
			poolAssets:           oneTrillionEvenPoolAssets,
			tokensIn:             sdk.NewCoins(sdk.NewInt64Coin("uosmo", 60_000), sdk.NewInt64Coin("uatom", 60_000)),
			expectShares:         sdk.NewInt(2_599_999_968_750 * 2), // ACTUAL: 5_999_999_999_999
			expectToleranceError: true,
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

				// If the expectShares is 0, no need to do error tolerance check
				if tc.expectShares.Int64() == 0 {
					require.Equal(t, tc.expectShares, totalNumShares)
					return
				}
				fmt.Println(tc.name, tc.expectShares, totalNumShares)
				err = validateExpectedSharesErrRatio(t, tc.expectShares, totalNumShares)
				if tc.expectToleranceError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
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
func (suite *BalancerTestSuite) TestBalancerCalculateAmountOutAndIn_InverseRelationship(t *testing.T) {
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

				sut := func() {
					suite.TestCalculateAmountOutAndIn_InverseRelationship(ctx, pool, poolAssetIn.Token.Denom, poolAssetOut.Token.Denom, tc.initialCalcOut, swapFeeDec)
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

// Expected is un-scaled
func testTotalWeight(t *testing.T, expected sdk.Int, pool balancer.Pool) {
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
		SwapFee:                  defaultSwapFee,
		ExitFee:                  defaultExitFee,
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

	for poolNum, tc := range tests {
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
		pacc, err := balancer.NewBalancerPool(uint64(poolNum), balancer.PoolParams{
			SwapFee:                  defaultSwapFee,
			ExitFee:                  defaultExitFee,
			SmoothWeightChangeParams: &tc.params,
		}, initialPoolAssets, defaultFutureGovernor, defaultCurBlockTime)
		require.NoError(t, err, "poolNumber %v", poolNum)

		// Consistency check that SmoothWeightChangeParams params are set
		require.NotNil(t, pacc.PoolParams.SmoothWeightChangeParams)

		testCases := addDefaultCases(paramsCopy, tc.cases)
		for caseNum, testCase := range testCases {
			pacc.PokePool(testCase.blockTime)

			totalWeight := sdk.ZeroInt()

			for assetNum, asset := range pacc.GetAllPoolAssets() {
				require.Equal(t, testCase.expectedWeights[assetNum], asset.Weight,
					"Didn't get the expected weights, poolNumber %v, caseNumber %v, assetNumber %v",
					poolNum, caseNum, assetNum)

				totalWeight = totalWeight.Add(asset.Weight)
			}

			require.Equal(t, totalWeight, pacc.GetTotalWeight())
		}
		// Should have been deleted by the last test case of after PokeTokenWeights pokes past end time.
		require.Nil(t, pacc.PoolParams.SmoothWeightChangeParams)
	}
}
