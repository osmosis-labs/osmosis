package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func (suite *KeeperTestSuite) TestCleanupPool() {

	defaultCoins := sdk.NewCoins(
		sdk.NewCoin("foo", sdk.NewInt(1000)),
		sdk.NewCoin("bar", sdk.NewInt(1000)),
		sdk.NewCoin("baz", sdk.NewInt(1000)),
	)

	tests := []struct {
		name           string
		createPoolFund sdk.Coins
		joinPoolFund   sdk.Coins
		expectedfail   bool
	}{
		{
			name:           "create pool with default coins, join pool with default coins",
			createPoolFund: defaultCoins,
			joinPoolFund:   defaultCoins,
		},
		{
			name:           "create pool with default coins, join pool larger coins",
			createPoolFund: defaultCoins,
			joinPoolFund:   defaultCoins.Add(defaultCoins...),
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		suite.Run(test.name, func() {
			joinPoolAcc1 := suite.TestAccs[1]
			joinPoolAcc2 := suite.TestAccs[2]

			// suite.TestAccs[0] gets funded and joins pool
			poolId := suite.PrepareBalancerPoolWithCoins(test.createPoolFund...)

			for _, acc := range []sdk.AccAddress{joinPoolAcc1, joinPoolAcc2} {
				suite.FundAcc(acc, test.joinPoolFund)

				_, _, err := suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, acc, poolId, types.OneShare.MulRaw(100), test.joinPoolFund)
				suite.NoError(err)
			}

			err := suite.App.GAMMKeeper.CleanupPools(suite.Ctx, []uint64{poolId})
			suite.Require().NoError(err)

			// double check that pool is deleted
			_, err = suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
			suite.Require().Error(err)

			// check that the balances are refunded
			suite.Require().Equal(suite.App.BankKeeper.GetAllBalances(suite.Ctx, suite.TestAccs[0]), test.createPoolFund)
			suite.Require().Equal(suite.App.BankKeeper.GetAllBalances(suite.Ctx, joinPoolAcc1), test.joinPoolFund)
			suite.Require().Equal(suite.App.BankKeeper.GetAllBalances(suite.Ctx, joinPoolAcc2), test.joinPoolFund)
		})
	}

}
