package gov_test

import (
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
	testCases := []struct {
		name      string
		args      types.SetSwapFeeProposal
		expectErr bool
	}{
		{
			"happy path flow",
			types.SetSwapFeeProposal{
				Title:       "tittle",
				Description: "des",
				Content: types.SetSwapFeeContent{
					PoolId:  1,
					SwapFee: sdk.MustNewDecFromStr("0.03"),
				},
			},
			false,
		},
		{
			"invalid pool id",
			types.SetSwapFeeProposal{
				Title:       "tittle",
				Description: "des",
				Content: types.SetSwapFeeContent{
					PoolId:  0,
					SwapFee: sdk.MustNewDecFromStr("0.03"),
				},
			},
			true,
		},
		{
			"pool not found",
			types.SetSwapFeeProposal{
				Title:       "tittle",
				Description: "des",
				Content: types.SetSwapFeeContent{
					PoolId:  2,
					SwapFee: sdk.MustNewDecFromStr("0.03"),
				},
			},
			true,
		},
		{
			"swap fee negative",
			types.SetSwapFeeProposal{
				Title:       "tittle",
				Description: "des",
				Content: types.SetSwapFeeContent{
					PoolId:  1,
					SwapFee: sdk.MustNewDecFromStr("-0.03"),
				},
			},
			true,
		},
		{
			"swap fee too much",
			types.SetSwapFeeProposal{
				Title:       "tittle",
				Description: "des",
				Content: types.SetSwapFeeContent{
					PoolId:  1,
					SwapFee: sdk.MustNewDecFromStr("1.1"),
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
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

			err = gov.HandleSetSwapFeeProposal(suite.ctx, *suite.app.GAMMKeeper, &tc.args)

			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				pool, err := suite.app.GAMMKeeper.GetPoolAndPoke(suite.ctx, poolId)
				suite.Require().NoError(err)
				suite.Require().Equal(pool.GetSwapFee(suite.ctx), tc.args.Content.SwapFee)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestHandleSetExitFeeProposal() {
	testCases := []struct {
		name      string
		args      types.SetExitFeeProposal
		expectErr bool
	}{
		{
			"happy path flow",
			types.SetExitFeeProposal{
				Title:       "tittle",
				Description: "des",
				Content: types.SetExitFeeContent{
					PoolId:  1,
					ExitFee: sdk.MustNewDecFromStr("0.03"),
				},
			},
			false,
		},
		{
			"invalid pool id",
			types.SetExitFeeProposal{
				Title:       "tittle",
				Description: "des",
				Content: types.SetExitFeeContent{
					PoolId:  0,
					ExitFee: sdk.MustNewDecFromStr("0.03"),
				},
			},
			true,
		},
		{
			"pool not found",
			types.SetExitFeeProposal{
				Title:       "tittle",
				Description: "des",
				Content: types.SetExitFeeContent{
					PoolId:  2,
					ExitFee: sdk.MustNewDecFromStr("0.03"),
				},
			},
			true,
		},
		{
			"exit fee negative",
			types.SetExitFeeProposal{
				Title:       "tittle",
				Description: "des",
				Content: types.SetExitFeeContent{
					PoolId:  1,
					ExitFee: sdk.MustNewDecFromStr("-0.03"),
				},
			},
			true,
		},
		{
			"exit fee too much",
			types.SetExitFeeProposal{
				Title:       "tittle",
				Description: "des",
				Content: types.SetExitFeeContent{
					PoolId:  1,
					ExitFee: sdk.MustNewDecFromStr("1.1"),
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
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

			err = gov.HandleSetExitFeeProposal(suite.ctx, *suite.app.GAMMKeeper, &tc.args)

			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				pool, err := suite.app.GAMMKeeper.GetPoolAndPoke(suite.ctx, poolId)
				suite.Require().NoError(err)
				suite.Require().Equal(pool.GetExitFee(suite.ctx), tc.args.Content.ExitFee)
			}
		})
	}
}
