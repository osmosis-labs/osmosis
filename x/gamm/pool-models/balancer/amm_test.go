package balancer_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
)

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

func TestCalculateAmountOutAndIn_InverseRelationship_ZeroSwapFee(t *testing.T) {
	type testcase struct {
		denomOut         string
		initialPoolOut   int64
		initialWeightOut int64
		initialCalcOut   int64

		denomIn         string
		initialPoolIn   int64
		initialWeightIn int64
	}

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
	}

	getTestCaseName := func(tc testcase) string {
		return fmt.Sprintf("tokenOutInitial: %d, tokenInInitial: %d, initialOut: %d",
			tc.initialPoolOut,
			tc.initialPoolIn,
			tc.initialCalcOut,
		)
	}

	for _, tc := range testcases {
		t.Run(getTestCaseName(tc), func(t *testing.T) {
			ctx := createTestContext(t)

			poolAssetOut := balancer.PoolAsset{
				Token:  sdk.NewInt64Coin(tc.denomOut, tc.initialPoolOut),
				Weight: sdk.NewInt(tc.initialWeightOut),
			}

			poolAssetIn := balancer.PoolAsset{
				Token:  sdk.NewInt64Coin(tc.denomIn, tc.initialPoolIn),
				Weight: sdk.NewInt(tc.initialWeightIn),
			}

			pool := createTestPool(t, []balancer.PoolAsset{
				poolAssetOut,
				poolAssetIn,
			},
				"0",
				"0",
			)
			require.NotNil(t, pool)

			initialOut := sdk.NewInt64Coin(poolAssetOut.Token.Denom, tc.initialCalcOut)
			initialOutCoins := sdk.NewCoins(initialOut)

			actualTokenIn, err := pool.CalcInAmtGivenOut(ctx, initialOutCoins, poolAssetIn.Token.Denom, sdk.ZeroDec())
			require.NoError(t, err)

			inverseTokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(sdk.NewInt64Coin(poolAssetIn.Token.Denom, actualTokenIn.Amount.TruncateInt64())), poolAssetOut.Token.Denom, sdk.ZeroDec())
			require.NoError(t, err)

			require.Equal(t, initialOut.Denom, inverseTokenOut.Denom)

			expected := initialOut.Amount.ToDec()
			actual := inverseTokenOut.Amount.RoundInt().ToDec() // must round to be able to compare with expected.

			require.Equal(t, expected, actual)
		})
	}
}
