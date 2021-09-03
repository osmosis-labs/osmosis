package keeper_test

func (suite *KeeperTestSuite) TestNativeHrpLifeCycle() {
	suite.SetupTest()

	// check genesis native hrp
	nativeHrp, err := suite.app.Bech32IBCKeeper.GetNativeHrp(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(nativeHrp, "uosmo")

	// check update of native hrp correctly
	err = suite.app.Bech32IBCKeeper.SetNativeHrp(suite.ctx, "osmo")
	suite.Require().NoError(err)

	nativeHrp, err = suite.app.Bech32IBCKeeper.GetNativeHrp(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(nativeHrp, "osmo")

	// error for uppercase in denom
	err = suite.app.Bech32IBCKeeper.SetNativeHrp(suite.ctx, "OSMO")
	suite.Require().Error(err)
}

// TODO: test ValidateHrpIbcRecord
// TODO: test GetHrpSourceChannel
// TODO: test GetHrpIbcRecord
// TODO: test setHrpIbcRecord
// TODO: test GetHrpIbcRecords
// TODO: test SetHrpIbcRecords
