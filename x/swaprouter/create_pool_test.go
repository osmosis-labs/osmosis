package swaprouter_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

func (suite *KeeperTestSuite) TestPoolCreationFee() {
	params := suite.App.SwapRouterKeeper.GetParams(suite.Ctx)

	// get raw pool creation fee(s) as DecCoins
	poolCreationFeeDecCoins := sdk.DecCoins{}
	for _, coin := range params.PoolCreationFee {
		poolCreationFeeDecCoins = poolCreationFeeDecCoins.Add(sdk.NewDecCoin(coin.Denom, coin.Amount))
	}

	tests := []struct {
		name            string
		poolCreationFee sdk.Coins
		msg             balancertypes.MsgCreateBalancerPool
		expectPass      bool
	}{
		{
			name:            "no pool creation fee for default asset pool",
			poolCreationFee: sdk.Coins{},
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: true,
		}, {
			name:            "nil pool creation fee on basic pool",
			poolCreationFee: nil,
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: true,
		}, {
			name:            "attempt pool creation without sufficient funds for fees",
			poolCreationFee: sdk.Coins{sdk.NewCoin("atom", sdk.NewInt(10000))},
			msg: balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, apptesting.DefaultPoolAssets, ""),
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()
		gammKeeper := suite.App.GAMMKeeper
		distributionKeeper := suite.App.DistrKeeper
		bankKeeper := suite.App.BankKeeper
		swaprouterKeeper := suite.App.SwapRouterKeeper

		// set pool creation fee
		swaprouterKeeper.SetParams(suite.Ctx, swaproutertypes.Params{
			PoolCreationFee: test.poolCreationFee,
		})

		// fund sender test account
		sender, err := sdk.AccAddressFromBech32(test.msg.Sender)
		suite.Require().NoError(err, "test: %v", test.name)
		suite.FundAcc(sender, apptesting.DefaultAcctFunds)

		// note starting balances for community fee pool and pool creator account
		feePoolBalBeforeNewPool := distributionKeeper.GetFeePoolCommunityCoins(suite.Ctx)
		senderBalBeforeNewPool := bankKeeper.GetAllBalances(suite.Ctx, sender)

		// attempt to create a pool with the given NewMsgCreateBalancerPool message
		poolId, err := gammKeeper.CreatePool(suite.Ctx, test.msg)

		if test.expectPass {
			suite.Require().NoError(err, "test: %v", test.name)

			// check to make sure new pool exists and has minted the correct number of pool shares
			pool, err := gammKeeper.GetPoolAndPoke(suite.Ctx, poolId)
			suite.Require().NoError(err, "test: %v", test.name)
			suite.Require().Equal(types.InitPoolSharesSupply.String(), pool.GetTotalShares().String(),
				fmt.Sprintf("share token should be minted as %s initially", types.InitPoolSharesSupply.String()),
			)

			// make sure pool creation fee is correctly sent to community pool
			feePool := distributionKeeper.GetFeePoolCommunityCoins(suite.Ctx)
			suite.Require().Equal(feePool, feePoolBalBeforeNewPool.Add(sdk.NewDecCoinsFromCoins(test.poolCreationFee...)...))
			// get expected tokens in new pool and corresponding pool shares
			expectedPoolTokens := sdk.Coins{}
			for _, asset := range test.msg.GetPoolAssets() {
				expectedPoolTokens = expectedPoolTokens.Add(asset.Token)
			}
			expectedPoolShares := sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), types.InitPoolSharesSupply)

			// make sure sender's balance is updated correctly
			senderBal := bankKeeper.GetAllBalances(suite.Ctx, sender)
			expectedSenderBal := senderBalBeforeNewPool.Sub(test.poolCreationFee).Sub(expectedPoolTokens).Add(expectedPoolShares)
			suite.Require().Equal(senderBal.String(), expectedSenderBal.String())

			// check pool's liquidity is correctly increased
			liquidity := gammKeeper.GetTotalLiquidity(suite.Ctx)
			suite.Require().Equal(expectedPoolTokens.String(), liquidity.String())
		} else {
			suite.Require().Error(err, "test: %v", test.name)
		}
	}
}
