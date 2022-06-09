package gov_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper/gov"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var (
	defaultSwapFee    = sdk.MustNewDecFromStr("0.025")
	defaultExitFee    = sdk.MustNewDecFromStr("0.025")
	defaultPoolParams = balancer.PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	defaultFutureGovernor = ""

	// pool assets
	defaultFooAsset = balancertypes.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}
	defaultBarAsset = balancertypes.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}
	defaultPoolAssets           = []balancertypes.PoolAsset{defaultFooAsset, defaultBarAsset}
	defaultAcctFunds  sdk.Coins = sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewCoin("foo", sdk.NewInt(10000000)),
		sdk.NewCoin("bar", sdk.NewInt(10000000)),
		sdk.NewCoin("baz", sdk.NewInt(10000000)),
	)
)

func (suite *KeeperTestSuite) TestHandleSetSwapFeeProposal() {
	tests := []struct {
		fn func(poolId uint64)
	}{
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				err := gov.HandleSetSwapFeeProposal(suite.ctx, *keeper, &types.SetSwapFeeProposal{
					Title: "tittle",
					Description: "des",
					Content: types.SetSwapFeeContent{
						PoolId: 1,
						SwapFee: sdk.MustNewDecFromStr("1.1"),
					},
				})

				suite.Require().Error(err, types.ErrTooMuchSwapFee)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				err := gov.HandleSetSwapFeeProposal(suite.ctx, *keeper, &types.SetSwapFeeProposal{
					Title: "tittle",
					Description: "des",
					Content: types.SetSwapFeeContent{
						PoolId: 1,
						SwapFee: sdk.MustNewDecFromStr("-0.01"),
					},
				})

				suite.Require().Error(err, types.ErrNegativeSwapFee)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				err := gov.HandleSetSwapFeeProposal(suite.ctx, *keeper, &types.SetSwapFeeProposal{
					Title: "tittle",
					Description: "des",
					Content: types.SetSwapFeeContent{
						PoolId: 1,
						SwapFee: sdk.MustNewDecFromStr("0.03"),
					},
				})

				suite.Require().NoError(err)
				pool, err := keeper.GetPool(suite.ctx, poolId)
				suite.Require().NoError(err)
				suite.Require().Equal(pool.GetSwapFee(suite.ctx), sdk.MustNewDecFromStr("0.03"))
			},
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range suite.TestAccs {
			suite.FundAcc(acc, defaultAcctFunds)
		}

		// Create the pool at first
		msg := balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, defaultPoolAssets, defaultFutureGovernor)
		poolId, err := suite.app.GAMMKeeper.CreatePool(suite.ctx, msg)
		suite.Require().NoError(err)

		test.fn(poolId)
	}
}

func (suite *KeeperTestSuite) TestHandleSetExitFeeProposal() {
	tests := []struct {
		fn func(poolId uint64)
	}{
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				err := gov.HandleSetExitFeeProposal(suite.ctx, *keeper, &types.SetExitFeeProposal{
					Title: "tittle",
					Description: "des",
					Content: types.SetExitFeeContent{
						PoolId: 1,
						ExitFee: sdk.MustNewDecFromStr("1.1"),
					},
				})

				suite.Require().Error(err, types.ErrTooMuchExitFee)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				err := gov.HandleSetExitFeeProposal(suite.ctx, *keeper, &types.SetExitFeeProposal{
					Title: "tittle",
					Description: "des",
					Content: types.SetExitFeeContent{
						PoolId: 1,
						ExitFee: sdk.MustNewDecFromStr("-0.01"),
					},
				})

				suite.Require().Error(err, types.ErrNegativeSwapFee)
			},
		},
		{
			fn: func(poolId uint64) {
				keeper := suite.app.GAMMKeeper

				err := gov.HandleSetExitFeeProposal(suite.ctx, *keeper, &types.SetExitFeeProposal{
					Title: "tittle",
					Description: "des",
					Content: types.SetExitFeeContent{
						PoolId: 1,
						ExitFee: sdk.MustNewDecFromStr("0.03"),
					},
				})

				suite.Require().NoError(err)
				pool, err := keeper.GetPool(suite.ctx, poolId)
				suite.Require().NoError(err)
				suite.Require().Equal(pool.GetExitFee(suite.ctx), sdk.MustNewDecFromStr("0.03"))
			},
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range suite.TestAccs {
			suite.FundAcc(acc, defaultAcctFunds)
		}

		// Create the pool at first
		msg := balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, defaultPoolAssets, defaultFutureGovernor)
		poolId, err := suite.app.GAMMKeeper.CreatePool(suite.ctx, msg)
		suite.Require().NoError(err)

		test.fn(poolId)
	}
}


