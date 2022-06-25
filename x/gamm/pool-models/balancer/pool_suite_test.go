package balancer_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	v10 "github.com/osmosis-labs/osmosis/v7/app/upgrades/v10"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

const (
	// allowedErrRatio is the maximal multiplicative difference in either
	// direction (positive or negative) that we accept to tolerate in
	// unit tests for calcuating the number of shares to be returned by
	// joining a pool. The comparison is done between Wolfram estimates and our AMM logic.
	allowedErrRatio = "0.0000001"
	// doesNotExistDenom denom name assummed to be used in test cases where the provided
	// denom does not exist in pool
	doesNotExistDenom = "doesnotexist"
)

var (
	oneTrillion          = sdk.NewInt(1e12)
	defaultOsmoPoolAsset = balancer.PoolAsset{
		Token:  sdk.NewCoin("uosmo", oneTrillion),
		Weight: sdk.NewInt(100),
	}
	defaultAtomPoolAsset = balancer.PoolAsset{
		Token:  sdk.NewCoin("uatom", oneTrillion),
		Weight: sdk.NewInt(100),
	}
	oneTrillionEvenPoolAssets = []balancer.PoolAsset{
		defaultOsmoPoolAsset,
		defaultAtomPoolAsset,
	}
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
	// be post-bug
	suite.Ctx = suite.Ctx.WithBlockHeight(v10.ForkHeight)
}

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

		sut := func() {
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
		assertPoolStateNotModified(suite.T(), balancerPool, sut)
	}
}
