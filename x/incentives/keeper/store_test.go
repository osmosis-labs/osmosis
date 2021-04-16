package keeper_test

func (suite *KeeperTestSuite) TestPotReferencesManagement() {

	key1 := []byte{0x11}
	key2 := []byte{0x12}

	suite.SetupTest()
	suite.app.IncentivesKeeper.AddPotRefByKey(suite.ctx, key1, 1)
	suite.app.IncentivesKeeper.AddPotRefByKey(suite.ctx, key2, 1)
	suite.app.IncentivesKeeper.AddPotRefByKey(suite.ctx, key1, 2)
	suite.app.IncentivesKeeper.AddPotRefByKey(suite.ctx, key2, 2)
	suite.app.IncentivesKeeper.AddPotRefByKey(suite.ctx, key2, 3)

	potRefs1 := suite.app.IncentivesKeeper.GetPotRefs(suite.ctx, key1)
	suite.Require().Equal(len(potRefs1), 2)
	potRefs2 := suite.app.IncentivesKeeper.GetPotRefs(suite.ctx, key2)
	suite.Require().Equal(len(potRefs2), 3)

	suite.app.IncentivesKeeper.DeletePotRefByKey(suite.ctx, key2, 1)
	potRefs3 := suite.app.IncentivesKeeper.GetPotRefs(suite.ctx, key2)
	suite.Require().Equal(len(potRefs3), 2)
}
