package keeper_test

// TestInitGenesis tests the initialization and export of the module's genesis state.
func (suite *KeeperTestSuite) TestInitGenesis() {
	// Export the genesis state
	exportedGenesis := suite.App.ProtoRevKeeper.ExportGenesis(suite.Ctx)

	// ------ Check that the exported genesis state matches the default genesis state ------ //
	// Check that the tokenPairArbRoutes match from what is initialized in keeper_test
	suite.Require().Equal(len(suite.tokenPairArbRoutes), len(exportedGenesis.TokenPairArbRoutes))
	for _, route := range exportedGenesis.TokenPairArbRoutes {
		suite.Require().Contains(suite.tokenPairArbRoutes, route)
	}

	// Check that the base denoms match from what is initialized in keeper_test
	baseDenoms, err := suite.App.ProtoRevKeeper.GetAllBaseDenoms(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(len(baseDenoms), len(exportedGenesis.BaseDenoms))
	for _, baseDenom := range exportedGenesis.BaseDenoms {
		suite.Require().Contains(baseDenoms, baseDenom)
	}

	// Check that the module parameters match from what is initialized in keeper_test
	params := suite.App.ProtoRevKeeper.GetParams(suite.Ctx)
	suite.Require().Equal(params, exportedGenesis.Params)

	// Check that the pool weights match from what is initialized in keeper_test
	poolWeights := suite.App.ProtoRevKeeper.GetPoolWeights(suite.Ctx)
	suite.Require().Equal(poolWeights, exportedGenesis.PoolWeights)

	// Check that the number of days since module genesis match from what is initialized in keeper_test
	daysSinceGenesis, err := suite.App.ProtoRevKeeper.GetDaysSinceModuleGenesis(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(exportedGenesis.DaysSinceModuleGenesis, daysSinceGenesis)

	// Check that the developer fees match from what is initialized in keeper_test
	developerFees, err := suite.App.ProtoRevKeeper.GetAllDeveloperFees(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(len(developerFees), len(exportedGenesis.DeveloperFees))
	for _, fee := range exportedGenesis.DeveloperFees {
		suite.Require().Contains(developerFees, fee)
	}

	// Check that the developer address matches from what is initialized in keeper_test
	developerAddress, _ := suite.App.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
	suite.Require().Equal(developerAddress.String(), exportedGenesis.DeveloperAddress)

	// Check that the latest block height matches from what is initialized in keeper_test
	latestBlockHeight, err := suite.App.ProtoRevKeeper.GetLatestBlockHeight(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(latestBlockHeight, exportedGenesis.LatestBlockHeight)

	// Check that the max pool points per tx matches from what is initialized in keeper_test
	maxPoolPointsPerTx, err := suite.App.ProtoRevKeeper.GetMaxPointsPerTx(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(maxPoolPointsPerTx, exportedGenesis.MaxPoolPointsPerTx)

	// Check that the max pool points per block matches from what is initialized in keeper_test
	maxPoolPointsPerBlock, err := suite.App.ProtoRevKeeper.GetMaxPointsPerBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(maxPoolPointsPerBlock, exportedGenesis.MaxPoolPointsPerBlock)

	// Check that the point count for tx matches from what is initialized in keeper_test
	pointCount, err := suite.App.ProtoRevKeeper.GetPointCountForBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(exportedGenesis.PointCountForBlock, pointCount)
}
