package keeper_test

// TestInitGenesis tests the initialization and export of the module's genesis state.
func (suite *KeeperTestSuite) TestInitGenesis() {
	// Export the genesis state
	exportedGenesis := suite.App.ProtoRevKeeper.ExportGenesis(suite.Ctx)

	// ------ Check that the exported genesis state matches the keeper test genesis state ------ //
	tokenPairArbRoutes, err := suite.App.ProtoRevKeeper.GetAllTokenPairArbRoutes(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(len(tokenPairArbRoutes), len(exportedGenesis.TokenPairArbRoutes))
	for _, route := range exportedGenesis.TokenPairArbRoutes {
		suite.Require().Contains(tokenPairArbRoutes, route)
	}

	baseDenoms, err := suite.App.ProtoRevKeeper.GetAllBaseDenoms(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(len(baseDenoms), len(exportedGenesis.BaseDenoms))
	for _, baseDenom := range exportedGenesis.BaseDenoms {
		suite.Require().Contains(baseDenoms, baseDenom)
	}

	params := suite.App.ProtoRevKeeper.GetParams(suite.Ctx)
	suite.Require().Equal(params, exportedGenesis.Params)

	poolWeights := suite.App.ProtoRevKeeper.GetPoolWeights(suite.Ctx)
	suite.Require().Equal(poolWeights, exportedGenesis.PoolWeights)

	daysSinceGenesis, err := suite.App.ProtoRevKeeper.GetDaysSinceModuleGenesis(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(daysSinceGenesis, exportedGenesis.DaysSinceModuleGenesis)

	developerFees, err := suite.App.ProtoRevKeeper.GetAllDeveloperFees(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(len(developerFees), len(exportedGenesis.DeveloperFees))
	for _, fee := range exportedGenesis.DeveloperFees {
		suite.Require().Contains(developerFees, fee)
	}

	developerAddress, _ := suite.App.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
	suite.Require().Equal(developerAddress.String(), exportedGenesis.DeveloperAddress)

	latestBlockHeight, err := suite.App.ProtoRevKeeper.GetLatestBlockHeight(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(latestBlockHeight, exportedGenesis.LatestBlockHeight)

	maxPoolPointsPerTx, err := suite.App.ProtoRevKeeper.GetMaxPointsPerTx(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(maxPoolPointsPerTx, exportedGenesis.MaxPoolPointsPerTx)

	maxPoolPointsPerBlock, err := suite.App.ProtoRevKeeper.GetMaxPointsPerBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(maxPoolPointsPerBlock, exportedGenesis.MaxPoolPointsPerBlock)

<<<<<<< HEAD
	pointCount, err := suite.App.ProtoRevKeeper.GetPointCountForBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(pointCount, exportedGenesis.PointCountForBlock)
=======
	pointCount, err := s.App.ProtoRevKeeper.GetPointCountForBlock(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(pointCount, exportedGenesis.PointCountForBlock)

	// Test protorev profits exported correctly
	profits := s.App.ProtoRevKeeper.GetAllProfits(s.Ctx)
	s.Require().Equal(len(profits), len(exportedGenesis.Profits))
	s.Require().Equal(profits, exportedGenesis.Profits)
>>>>>>> b29b2c66 (Jl/add protorev profits to genesis export (#5513))
}
