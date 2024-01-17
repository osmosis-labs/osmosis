package keeper_test

// TestInitGenesis tests the initialization and export of the module's genesis state.
func (s *KeeperTestSuite) TestInitGenesis() {
	// Export the genesis state
	exportedGenesis := s.App.ProtoRevKeeper.ExportGenesis(s.Ctx)

	// ------ Check that the exported genesis state matches the keeper test genesis state ------ //
	tokenPairArbRoutes, err := s.App.ProtoRevKeeper.GetAllTokenPairArbRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(len(tokenPairArbRoutes), len(exportedGenesis.TokenPairArbRoutes))
	for _, route := range exportedGenesis.TokenPairArbRoutes {
		s.Require().Contains(tokenPairArbRoutes, route)
	}

	baseDenoms, err := s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(len(baseDenoms), len(exportedGenesis.BaseDenoms))
	for _, baseDenom := range exportedGenesis.BaseDenoms {
		s.Require().Contains(baseDenoms, baseDenom)
	}

	params := s.App.ProtoRevKeeper.GetParams(s.Ctx)
	s.Require().Equal(params, exportedGenesis.Params)

	poolInfo := s.App.ProtoRevKeeper.GetInfoByPoolType(s.Ctx)
	s.Require().Equal(poolInfo, exportedGenesis.InfoByPoolType)

	daysSinceGenesis, err := s.App.ProtoRevKeeper.GetDaysSinceModuleGenesis(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(daysSinceGenesis, exportedGenesis.DaysSinceModuleGenesis)

	developerFees, err := s.App.ProtoRevKeeper.GetAllDeveloperFees(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(len(developerFees), len(exportedGenesis.DeveloperFees))
	for _, fee := range exportedGenesis.DeveloperFees {
		s.Require().Contains(developerFees, fee)
	}

	developerAddress, _ := s.App.ProtoRevKeeper.GetDeveloperAccount(s.Ctx)
	s.Require().Equal(developerAddress.String(), exportedGenesis.DeveloperAddress)

	latestBlockHeight, err := s.App.ProtoRevKeeper.GetLatestBlockHeight(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(latestBlockHeight, exportedGenesis.LatestBlockHeight)

	maxPoolPointsPerTx, err := s.App.ProtoRevKeeper.GetMaxPointsPerTx(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(maxPoolPointsPerTx, exportedGenesis.MaxPoolPointsPerTx)

	maxPoolPointsPerBlock, err := s.App.ProtoRevKeeper.GetMaxPointsPerBlock(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(maxPoolPointsPerBlock, exportedGenesis.MaxPoolPointsPerBlock)

	pointCount, err := s.App.ProtoRevKeeper.GetPointCountForBlock(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(pointCount, exportedGenesis.PointCountForBlock)

	// Test protorev profits exported correctly
	profits := s.App.ProtoRevKeeper.GetAllProfits(s.Ctx)
	s.Require().Equal(len(profits), len(exportedGenesis.Profits))
	s.Require().Equal(profits, exportedGenesis.Profits)

	cyclicArbProfit := s.App.ProtoRevKeeper.GetCyclicArbProfitTrackerValue(s.Ctx)
	s.Require().Equal(cyclicArbProfit, exportedGenesis.CyclicArbTracker.CyclicArb)

	cyclicArbProfitAccountingHeight := s.App.ProtoRevKeeper.GetCyclicArbProfitTrackerStartHeight(s.Ctx)
	s.Require().Equal(cyclicArbProfitAccountingHeight, exportedGenesis.CyclicArbTracker.HeightAccountingStartsFrom)
}
