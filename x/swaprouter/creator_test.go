package swaprouter_test

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clmodel "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// TestCreatePool tests that all possible pools are created correctly.
func (suite *KeeperTestSuite) TestCreatePool() {

	validBalancerPoolMsg := balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.NewPoolParams(sdk.ZeroDec(), sdk.ZeroDec(), nil), []balancer.PoolAsset{
		{
			Token:  sdk.NewCoin(foo, defaultInitPoolAmount),
			Weight: sdk.NewInt(1),
		},
		{
			Token:  sdk.NewCoin(bar, defaultInitPoolAmount),
			Weight: sdk.NewInt(1),
		},
	}, "")

	validConcentratedPoolMsg := clmodel.NewMsgCreateConcentratedPool(suite.TestAccs[0], foo, bar, 1, defaultPrecisionValue)

	tests := []struct {
		name               string
		creatorFundAmount  sdk.Coins
		msg                types.CreatePoolMsg
		expectedModuleType reflect.Type
		expectError        bool
	}{
		{
			name:               "first balancer pool - success",
			creatorFundAmount:  sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount.Mul(sdk.NewInt(2))), sdk.NewCoin(bar, defaultInitPoolAmount.Mul(sdk.NewInt(2)))),
			msg:                validBalancerPoolMsg,
			expectedModuleType: gammKeeperType,
		},
		{
			name:               "second balancer pool - success",
			creatorFundAmount:  sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount.Mul(sdk.NewInt(2))), sdk.NewCoin(bar, defaultInitPoolAmount.Mul(sdk.NewInt(2)))),
			msg:                validBalancerPoolMsg,
			expectedModuleType: gammKeeperType,
		},
		{
			name:               "concentrated pool - success",
			creatorFundAmount:  sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount.Mul(sdk.NewInt(2))), sdk.NewCoin(bar, defaultInitPoolAmount.Mul(sdk.NewInt(2)))),
			msg:                validConcentratedPoolMsg,
			expectedModuleType: concentratedKeeperType,
		},
		// TODO: add stableswap test
		// TODO: cover errors and edge cases
	}

	for i, tc := range tests {
		suite.Run(tc.name, func() {
			tc := tc

			swaprouterKeeper := suite.App.SwapRouterKeeper
			ctx := suite.Ctx

			poolCreationFee := swaprouterKeeper.GetParams(suite.Ctx).PoolCreationFee
			suite.FundAcc(suite.TestAccs[0], append(tc.creatorFundAmount, poolCreationFee...))

			poolId, err := swaprouterKeeper.CreatePool(ctx, tc.msg)

			if tc.expectError {
				suite.Require().Error(err)
				return
			}

			// Validate pool.
			suite.Require().NoError(err)
			suite.Require().Equal(uint64(i+1), poolId)

			// Validate that mapping pool id -> module type has been persisted.
			swapModule, err := swaprouterKeeper.GetSwapModule(ctx, poolId)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedModuleType, reflect.TypeOf(swapModule))
		})
	}
}
