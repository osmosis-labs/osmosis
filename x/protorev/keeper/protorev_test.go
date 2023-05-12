package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// TestGetTokenPairArbRoutes tests the GetTokenPairArbRoutes function.
func (suite *KeeperTestSuite) TestGetTokenPairArbRoutes() {
	// Tests that we can properly retrieve all of the routes that were set up
	for _, tokenPair := range suite.tokenPairArbRoutes {
		tokenPairArbRoutes, err := suite.App.ProtoRevKeeper.GetTokenPairArbRoutes(suite.Ctx, tokenPair.TokenIn, tokenPair.TokenOut)

		suite.Require().NoError(err)
		suite.Require().Equal(tokenPair, tokenPairArbRoutes)
	}

	// Testing to see if we will not find a route that does not exist
	_, err := suite.App.ProtoRevKeeper.GetTokenPairArbRoutes(suite.Ctx, "osmo", "abc")
	suite.Require().Error(err)
}

// TestGetAllTokenPairArbRoutes tests the GetAllTokenPairArbRoutes function.
func (suite *KeeperTestSuite) TestGetAllTokenPairArbRoutes() {
	// Tests that we can properly retrieve all of the routes that were set up
	tokenPairArbRoutes, err := suite.App.ProtoRevKeeper.GetAllTokenPairArbRoutes(suite.Ctx)

	suite.Require().NoError(err)

	suite.Require().Equal(len(suite.tokenPairArbRoutes), len(tokenPairArbRoutes))
	for _, tokenPair := range suite.tokenPairArbRoutes {
		suite.Require().Contains(tokenPairArbRoutes, tokenPair)
	}
}

// TestDeleteAllTokenPairArbRoutes tests the DeleteAllTokenPairArbRoutes function.
func (suite *KeeperTestSuite) TestDeleteAllTokenPairArbRoutes() {
	// Tests that we can properly retrieve all of the routes that were set up
	tokenPairArbRoutes, err := suite.App.ProtoRevKeeper.GetAllTokenPairArbRoutes(suite.Ctx)

	suite.Require().NoError(err)
	suite.Require().Equal(len(suite.tokenPairArbRoutes), len(tokenPairArbRoutes))
	for _, tokenPair := range suite.tokenPairArbRoutes {
		suite.Require().Contains(tokenPairArbRoutes, tokenPair)
	}

	// Delete all routes
	suite.App.ProtoRevKeeper.DeleteAllTokenPairArbRoutes(suite.Ctx)

	// Test after deletion
	tokenPairArbRoutes, err = suite.App.ProtoRevKeeper.GetAllTokenPairArbRoutes(suite.Ctx)

	suite.Require().NoError(err)
	suite.Require().Equal(0, len(tokenPairArbRoutes))
}

// TestGetAllBaseDenoms tests the GetAllBaseDenoms, SetBaseDenoms, and DeleteBaseDenoms functions.
func (suite *KeeperTestSuite) TestGetAllBaseDenoms() {
	// Should be initialized on genesis
	baseDenoms, err := suite.App.ProtoRevKeeper.GetAllBaseDenoms(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(baseDenoms))
	suite.Require().Equal(baseDenoms[0].Denom, types.OsmosisDenomination)
	suite.Require().Equal(baseDenoms[1].Denom, "Atom")
	suite.Require().Equal(baseDenoms[2].Denom, "test/3")

	// Should be able to delete all base denoms
	suite.App.ProtoRevKeeper.DeleteBaseDenoms(suite.Ctx)
	baseDenoms, err = suite.App.ProtoRevKeeper.GetAllBaseDenoms(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(baseDenoms))

	// Should be able to set the base denoms
	err = suite.App.ProtoRevKeeper.SetBaseDenoms(suite.Ctx, []types.BaseDenom{{Denom: "osmo"}, {Denom: "atom"}, {Denom: "weth"}})
	suite.Require().NoError(err)
	baseDenoms, err = suite.App.ProtoRevKeeper.GetAllBaseDenoms(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(baseDenoms))
	suite.Require().Equal(baseDenoms[0].Denom, "osmo")
	suite.Require().Equal(baseDenoms[1].Denom, "atom")
	suite.Require().Equal(baseDenoms[2].Denom, "weth")
}

// TestGetPoolForDenomPair tests the GetPoolForDenomPair, SetPoolForDenomPair, and DeleteAllPoolsForBaseDenom functions.
func (suite *KeeperTestSuite) TestGetPoolForDenomPair() {
	// Should be able to set a pool for a denom pair
	suite.App.ProtoRevKeeper.SetPoolForDenomPair(suite.Ctx, "Atom", types.OsmosisDenomination, 1000)
	pool, err := suite.App.ProtoRevKeeper.GetPoolForDenomPair(suite.Ctx, "Atom", types.OsmosisDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1000), pool)

	// Should be able to add another pool for a denom pair
	suite.App.ProtoRevKeeper.SetPoolForDenomPair(suite.Ctx, "Atom", "weth", 2000)
	pool, err = suite.App.ProtoRevKeeper.GetPoolForDenomPair(suite.Ctx, "Atom", "weth")
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(2000), pool)

	suite.App.ProtoRevKeeper.SetPoolForDenomPair(suite.Ctx, types.OsmosisDenomination, "Atom", 3000)
	pool, err = suite.App.ProtoRevKeeper.GetPoolForDenomPair(suite.Ctx, types.OsmosisDenomination, "Atom")
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(3000), pool)

	// Should be able to delete all pools for a base denom
	suite.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(suite.Ctx, "Atom")
	_, err = suite.App.ProtoRevKeeper.GetPoolForDenomPair(suite.Ctx, "Atom", types.OsmosisDenomination)
	suite.Require().Error(err)
	_, err = suite.App.ProtoRevKeeper.GetPoolForDenomPair(suite.Ctx, "Atom", "weth")
	suite.Require().Error(err)

	// Other denoms should still exist
	pool, err = suite.App.ProtoRevKeeper.GetPoolForDenomPair(suite.Ctx, types.OsmosisDenomination, "Atom")
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(3000), pool)
}

// TestGetDaysSinceModuleGenesis tests the GetDaysSinceModuleGenesis and SetDaysSinceModuleGenesis functions.
func (suite *KeeperTestSuite) TestGetDaysSinceModuleGenesis() {
	// Should be initialized to 0 on genesis
	daysSinceGenesis, err := suite.App.ProtoRevKeeper.GetDaysSinceModuleGenesis(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), daysSinceGenesis)

	// Should be able to set the days since genesis
	suite.App.ProtoRevKeeper.SetDaysSinceModuleGenesis(suite.Ctx, 1)
	daysSinceGenesis, err = suite.App.ProtoRevKeeper.GetDaysSinceModuleGenesis(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), daysSinceGenesis)
}

// TestGetDeveloperFees tests the GetDeveloperFees, SetDeveloperFees, and GetAllDeveloperFees functions.
func (suite *KeeperTestSuite) TestGetDeveloperFees() {
	// Should be initialized to [] on genesis
	fees, err := suite.App.ProtoRevKeeper.GetAllDeveloperFees(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(fees))

	// Should be no osmo fees on genesis
	osmoFees, err := suite.App.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, types.OsmosisDenomination)
	suite.Require().Error(err)
	suite.Require().Equal(sdk.Coin{}, osmoFees)

	// Should be no atom fees on genesis
	atomFees, err := suite.App.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, "Atom")
	suite.Require().Error(err)
	suite.Require().Equal(sdk.Coin{}, atomFees)

	// Should be able to set the fees
	err = suite.App.ProtoRevKeeper.SetDeveloperFees(suite.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)))
	suite.Require().NoError(err)
	err = suite.App.ProtoRevKeeper.SetDeveloperFees(suite.Ctx, sdk.NewCoin("Atom", sdk.NewInt(100)))
	suite.Require().NoError(err)
	err = suite.App.ProtoRevKeeper.SetDeveloperFees(suite.Ctx, sdk.NewCoin("weth", sdk.NewInt(100)))
	suite.Require().NoError(err)

	// Should be able to get the fees
	osmoFees, err = suite.App.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, types.OsmosisDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)), osmoFees)
	atomFees, err = suite.App.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, "Atom")
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin("Atom", sdk.NewInt(100)), atomFees)
	wethFees, err := suite.App.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, "weth")
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin("weth", sdk.NewInt(100)), wethFees)

	fees, err = suite.App.ProtoRevKeeper.GetAllDeveloperFees(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(fees))
	suite.Require().Contains(fees, osmoFees)
	suite.Require().Contains(fees, atomFees)
}

// TestDeleteDeveloperFees tests the DeleteDeveloperFees function.
func (suite *KeeperTestSuite) TestDeleteDeveloperFees() {
	err := suite.App.ProtoRevKeeper.SetDeveloperFees(suite.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)))
	suite.Require().NoError(err)

	// Should be able to get the fees
	osmoFees, err := suite.App.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, types.OsmosisDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)), osmoFees)

	// Should be able to delete the fees
	suite.App.ProtoRevKeeper.DeleteDeveloperFees(suite.Ctx, types.OsmosisDenomination)

	// Should be no osmo fees after deletion
	osmoFees, err = suite.App.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, types.OsmosisDenomination)
	suite.Require().Error(err)
	suite.Require().Equal(sdk.Coin{}, osmoFees)
}

// TestGetProtoRevEnabled tests the GetProtoRevEnabled and SetProtoRevEnabled functions.
func (suite *KeeperTestSuite) TestGetProtoRevEnabled() {
	// Should be initialized to true on genesis
	protoRevEnabled := suite.App.ProtoRevKeeper.GetProtoRevEnabled(suite.Ctx)
	suite.Require().Equal(true, protoRevEnabled)

	// Should be able to set the protoRevEnabled
	suite.App.ProtoRevKeeper.SetProtoRevEnabled(suite.Ctx, false)
	protoRevEnabled = suite.App.ProtoRevKeeper.GetProtoRevEnabled(suite.Ctx)
	suite.Require().Equal(false, protoRevEnabled)
}

// TestGetAdminAccount tests the GetAdminAccount and SetAdminAccount functions.
func (suite *KeeperTestSuite) TestGetAdminAccount() {
	// Should be initialized (look at keeper_test.go)
	adminAccount := suite.App.ProtoRevKeeper.GetAdminAccount(suite.Ctx)
	suite.Require().Equal(suite.adminAccount, adminAccount)

	// Should be able to set the admin account
	suite.App.ProtoRevKeeper.SetAdminAccount(suite.Ctx, suite.TestAccs[0])
	adminAccount = suite.App.ProtoRevKeeper.GetAdminAccount(suite.Ctx)
	suite.Require().Equal(suite.TestAccs[0], adminAccount)
}

// TestGetDeveloperAccount tests the GetDeveloperAccount and SetDeveloperAccount functions.
func (suite *KeeperTestSuite) TestGetDeveloperAccount() {
	// Should be null on genesis
	developerAccount, err := suite.App.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
	suite.Require().Error(err)
	suite.Require().Nil(developerAccount)

	// Should be able to set the developer account
	suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, suite.TestAccs[0])
	developerAccount, err = suite.App.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.TestAccs[0], developerAccount)
}

// TestGetMaxPointsPerTx tests the GetMaxPointsPerTx and SetMaxPointsPerTx functions.
func (suite *KeeperTestSuite) TestGetMaxPointsPerTx() {
	// Should be initialized on genesis
	maxPoints, err := suite.App.ProtoRevKeeper.GetMaxPointsPerTx(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(18), maxPoints)

	// Should be able to set the max points per tx
	err = suite.App.ProtoRevKeeper.SetMaxPointsPerTx(suite.Ctx, 4)
	suite.Require().NoError(err)
	maxPoints, err = suite.App.ProtoRevKeeper.GetMaxPointsPerTx(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(4), maxPoints)

	// Can only be set between 1 and types.MaxPoolPointsPerTx
	err = suite.App.ProtoRevKeeper.SetMaxPointsPerTx(suite.Ctx, 0)
	suite.Require().Error(err)
	err = suite.App.ProtoRevKeeper.SetMaxPointsPerTx(suite.Ctx, types.MaxPoolPointsPerTx+1)
	suite.Require().Error(err)
}

// TestGetPointCountForBlock tests the GetPointCountForBlock, IncrementPointCountForBlock and SetPointCountForBlock functions.
func (suite *KeeperTestSuite) TestGetPointCountForBlock() {
	// Should be initialized to 0 on genesis
	pointCount, err := suite.App.ProtoRevKeeper.GetPointCountForBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), pointCount)

	// Should be able to set the point count
	suite.App.ProtoRevKeeper.SetPointCountForBlock(suite.Ctx, 4)
	pointCount, err = suite.App.ProtoRevKeeper.GetPointCountForBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(4), pointCount)

	// Should be able to increment the point count
	err = suite.App.ProtoRevKeeper.IncrementPointCountForBlock(suite.Ctx, 10)
	suite.Require().NoError(err)
	pointCount, err = suite.App.ProtoRevKeeper.GetPointCountForBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(14), pointCount)
}

// TestGetLatestBlockHeight tests the GetLatestBlockHeight and SetLatestBlockHeight functions.
func (suite *KeeperTestSuite) TestGetLatestBlockHeight() {
	// Should be initialized on genesis
	blockHeight, err := suite.App.ProtoRevKeeper.GetLatestBlockHeight(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), blockHeight)

	// Should be able to set the blockHeight
	suite.App.ProtoRevKeeper.SetLatestBlockHeight(suite.Ctx, 4)
	blockHeight, err = suite.App.ProtoRevKeeper.GetLatestBlockHeight(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(4), blockHeight)
}

// TestGetMaxPointsPerBlock tests the GetMaxPointsPerBlock and SetMaxPointsPerBlock functions.
func (suite *KeeperTestSuite) TestGetMaxPointsPerBlock() {
	// Should be initialized on genesis
	maxPoints, err := suite.App.ProtoRevKeeper.GetMaxPointsPerBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(100), maxPoints)

	// Should be able to set the max points per block
	err = suite.App.ProtoRevKeeper.SetMaxPointsPerBlock(suite.Ctx, 4)
	suite.Require().NoError(err)
	maxPoints, err = suite.App.ProtoRevKeeper.GetMaxPointsPerBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(4), maxPoints)

	// Can only initialize between 1 and types.MaxPoolPointsPerBlock
	err = suite.App.ProtoRevKeeper.SetMaxPointsPerBlock(suite.Ctx, 0)
	suite.Require().Error(err)
	err = suite.App.ProtoRevKeeper.SetMaxPointsPerBlock(suite.Ctx, types.MaxPoolPointsPerBlock+1)
	suite.Require().Error(err)
}

// TestGetPoolWeights tests the GetPoolWeights and SetPoolWeights functions.
func (suite *KeeperTestSuite) TestGetPoolWeights() {
	// Should be initialized on genesis
	poolWeights := suite.App.ProtoRevKeeper.GetPoolWeights(suite.Ctx)
	suite.Require().Equal(types.PoolWeights{StableWeight: 5, BalancerWeight: 2, ConcentratedWeight: 2}, poolWeights)

	// Should be able to set the PoolWeights
	newRouteWeights := types.PoolWeights{
		StableWeight:       10,
		BalancerWeight:     2,
		ConcentratedWeight: 22,
	}

	suite.App.ProtoRevKeeper.SetPoolWeights(suite.Ctx, newRouteWeights)

	poolWeights = suite.App.ProtoRevKeeper.GetPoolWeights(suite.Ctx)
	suite.Require().Equal(newRouteWeights, poolWeights)
}
