package keeper_test

import sdk "github.com/cosmos/cosmos-sdk/types"

func (suite *KeeperTestSuite) TestSimpleReward() {
	suite.prepareAccounts()

	keeper := suite.app.FarmKeeper

	farm, err := keeper.NewFarm(suite.ctx)
	suite.NoError(err)

	rewards, err := keeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewInt(1))
	suite.NoError(err)
	suite.Equal(0, len(rewards))
	suite.Equal("0", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "foo").Amount.String())

	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
	suite.NoError(err)

	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc1)
	suite.NoError(err)
	suite.Equal("1000foo", rewards.String())
	suite.Equal("1000", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "foo").Amount.String())

	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
	suite.NoError(err)
	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
	suite.NoError(err)

	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc1)
	suite.NoError(err)
	suite.Equal("2000foo", rewards.String())
	suite.Equal("3000", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "foo").Amount.String())
}

func (suite *KeeperTestSuite) TestSimpleReward2() {
	suite.prepareAccounts()

	keeper := suite.app.FarmKeeper

	farm, err := keeper.NewFarm(suite.ctx)
	suite.NoError(err)

	rewards, err := keeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewInt(1))
	suite.NoError(err)
	suite.Equal(0, len(rewards))
	suite.Equal("0", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "foo").Amount.String())

	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
	suite.NoError(err)

	// Until this, acc1 has the 1000foo as rewards.

	rewards, err = keeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc2, sdk.NewInt(2))
	suite.NoError(err)
	suite.Equal(0, len(rewards))
	suite.Equal("0", suite.app.BankKeeper.GetBalance(suite.ctx, acc2, "foo").Amount.String())

	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000))))
	suite.NoError(err)
	err = keeper.AllocateAssetsFromAccountToFarm(suite.ctx, farm.FarmId, allocatorAcc, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(2000))))
	suite.NoError(err)

	// Until this, acc1 has the 2000foo as rewards. And, acc2 has the 2000foo as rewards.
	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc1)
	suite.NoError(err)
	// But has small difference...
	suite.Equal("1999foo", rewards.String())
	suite.Equal("1999", suite.app.BankKeeper.GetBalance(suite.ctx, acc1, "foo").Amount.String())
	rewards, err = keeper.WithdrawRewardsFromFarm(suite.ctx, farm.FarmId, acc2)
	suite.NoError(err)
	// But has small difference...
	suite.Equal("1999foo", rewards.String())
	suite.Equal("1999", suite.app.BankKeeper.GetBalance(suite.ctx, acc2, "foo").Amount.String())
}
