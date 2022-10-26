package swaprouter_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

// TestCreatePool tests that all possible pools are created correctly.
func (suite *KeeperTestSuite) TestCreatePool() {

	tests := []struct {
		name              string
		creatorFundAmount sdk.Coins
		msg               types.CreatePoolMsg

		expectedPoolId uint64
		expectError    bool
	}{
		{
			name:              "balancer pool - success",
			creatorFundAmount: sdk.NewCoins(sdk.NewCoin(denomA, defaultInitPoolAmount.Mul(sdk.NewInt(2))), sdk.NewCoin(denomB, defaultInitPoolAmount.Mul(sdk.NewInt(2)))),
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.NewPoolParams(sdk.ZeroDec(), sdk.ZeroDec(), nil), []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin(denomA, defaultInitPoolAmount),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(denomB, defaultInitPoolAmount),
					Weight: sdk.NewInt(1),
				},
			}, ""),

			expectedPoolId: 0,
		},
		// TODO: add stableswap test
		// TODO: add concentrated-liquidity rest
		// TODO: cover errors and edge cases
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc := tc

			swaprouterKeeper := suite.App.SwapRouterKeeper

			poolCreationFee := swaprouterKeeper.GetParams(suite.Ctx).PoolCreationFee
			suite.FundAcc(suite.TestAccs[0], append(tc.creatorFundAmount, poolCreationFee...))

			poolId, err := swaprouterKeeper.CreatePool(suite.Ctx, tc.msg)

			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(uint64(1), poolId)
			}
		})
	}
}
