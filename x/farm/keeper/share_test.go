package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestDepositShareToFarm() {
	suite.Run("deposit to the non existing farm", func() {
		_, err := suite.app.FarmKeeper.DepositShareToFarm(suite.ctx, 1, acc1, sdk.NewInt(100))
		suite.Error(err)
	})

	suite.Run("deposit the negative share", func() {
		farm, err := suite.app.FarmKeeper.NewFarm(suite.ctx)
		suite.NoError(err)
		_, err = suite.app.FarmKeeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewInt(-100))
		suite.Error(err)
	})

	suite.Run("deposit the zero share", func() {
		farm, err := suite.app.FarmKeeper.NewFarm(suite.ctx)
		suite.NoError(err)
		_, err = suite.app.FarmKeeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewInt(0))
		suite.Error(err)
	})

	suite.Run("depositing should create new farmer", func() {
		farm, err := suite.app.FarmKeeper.NewFarm(suite.ctx)
		suite.NoError(err)
		_, err = suite.app.FarmKeeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewInt(100))
		suite.NoError(err)

		farmer, err := suite.app.FarmKeeper.GetFarmer(suite.ctx, farm.FarmId, acc1)
		suite.NoError(err)

		suite.Equal(farmer.Address, acc1.String())
		suite.Equal(farmer.FarmId, farm.FarmId)
		suite.Equal(farmer.Share.String(), "100")

		farm, err = suite.app.FarmKeeper.GetFarm(suite.ctx, farm.FarmId)
		suite.NoError(err)

		suite.Equal(farm.TotalShare.String(), "100")
	})

	suite.Run("depositing to existing farmer should increase the share", func() {
		farm, err := suite.app.FarmKeeper.NewFarm(suite.ctx)
		suite.NoError(err)
		_, err = suite.app.FarmKeeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewInt(100))
		suite.NoError(err)
		farmer, err := suite.app.FarmKeeper.GetFarmer(suite.ctx, farm.FarmId, acc1)
		suite.NoError(err)

		suite.Equal(farmer.Address, acc1.String())
		suite.Equal(farmer.FarmId, farm.FarmId)
		suite.Equal(farmer.Share.String(), "100")

		_, err = suite.app.FarmKeeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewInt(50))
		suite.NoError(err)
		farmer, err = suite.app.FarmKeeper.GetFarmer(suite.ctx, farm.FarmId, acc1)
		suite.NoError(err)

		suite.Equal(farmer.Address, acc1.String())
		suite.Equal(farmer.FarmId, farm.FarmId)
		suite.Equal(farmer.Share.String(), "150")

		farm, err = suite.app.FarmKeeper.GetFarm(suite.ctx, farm.FarmId)
		suite.NoError(err)

		suite.Equal(farm.TotalShare.String(), "150")
	})
}

func (suite *KeeperTestSuite) TestWithdrawShareFromFarm() {
	// Depoist 300 share to the farm to test the withdrawing the share
	prepare := func() uint64 {
		farm, err := suite.app.FarmKeeper.NewFarm(suite.ctx)
		suite.NoError(err)

		_, err = suite.app.FarmKeeper.DepositShareToFarm(suite.ctx, farm.FarmId, acc1, sdk.NewInt(300))
		suite.NoError(err)

		return farm.FarmId
	}

	suite.Run("withdraw share from the non existing farm", func() {
		_, err := suite.app.FarmKeeper.WithdrawShareFromFarm(suite.ctx, 1, acc1, sdk.NewInt(100))
		suite.Error(err)
	})

	suite.Run("withdraw the negative share", func() {
		farmId := prepare()

		_, err := suite.app.FarmKeeper.WithdrawShareFromFarm(suite.ctx, farmId, acc1, sdk.NewInt(-100))
		suite.Error(err)
	})

	suite.Run("withdraw the zero share", func() {
		farmId := prepare()

		_, err := suite.app.FarmKeeper.WithdrawShareFromFarm(suite.ctx, farmId, acc1, sdk.NewInt(0))
		suite.Error(err)
	})

	suite.Run("withdrawing the share from non existing farmer should fail", func() {
		farmId := prepare()

		_, err := suite.app.FarmKeeper.WithdrawShareFromFarm(suite.ctx, farmId, acc2, sdk.NewInt(100))
		suite.Error(err)
	})

	suite.Run("withdrawing the insufficient share should fail", func() {
		farmId := prepare()

		_, err := suite.app.FarmKeeper.WithdrawShareFromFarm(suite.ctx, farmId, acc1, sdk.NewInt(301))
		suite.Error(err)
	})

	suite.Run("withdrawing the share should decrease the share", func() {
		farmId := prepare()

		_, err := suite.app.FarmKeeper.WithdrawShareFromFarm(suite.ctx, farmId, acc1, sdk.NewInt(100))
		suite.Error(err)

		farmer, err := suite.app.FarmKeeper.GetFarmer(suite.ctx, farmId, acc1)
		suite.NoError(err)

		suite.Equal(farmer.Share.String(), "200")

		farm, err := suite.app.FarmKeeper.GetFarm(suite.ctx, farmId)
		suite.NoError(err)
		suite.Equal(farm.TotalShare.String(), "200")
	})
}
