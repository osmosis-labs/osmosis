package keeper_test

func (suite *KeeperTestSuite) TestGaugeReferencesManagement() {
	key1 := []byte{0x11}
	key2 := []byte{0x12}

	suite.SetupTest()
	suite.app.IncentivesKeeper.AddGaugeRefByKey(suite.ctx, key1, 1)
	suite.app.IncentivesKeeper.AddGaugeRefByKey(suite.ctx, key2, 1)
	suite.app.IncentivesKeeper.AddGaugeRefByKey(suite.ctx, key1, 2)
	suite.app.IncentivesKeeper.AddGaugeRefByKey(suite.ctx, key2, 2)
	suite.app.IncentivesKeeper.AddGaugeRefByKey(suite.ctx, key2, 3)

	gaugeRefs1 := suite.app.IncentivesKeeper.GetGaugeRefs(suite.ctx, key1)
	suite.Require().Equal(len(gaugeRefs1), 2)
	gaugeRefs2 := suite.app.IncentivesKeeper.GetGaugeRefs(suite.ctx, key2)
	suite.Require().Equal(len(gaugeRefs2), 3)

	suite.app.IncentivesKeeper.DeleteGaugeRefByKey(suite.ctx, key2, 1)
	gaugeRefs3 := suite.app.IncentivesKeeper.GetGaugeRefs(suite.ctx, key2)
	suite.Require().Equal(len(gaugeRefs3), 2)
}
