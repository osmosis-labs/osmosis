package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

// TestGetAtomPool tests the GetAtomPool function.
func (suite *KeeperTestSuite) TestGetAtomPool() {
	cases := []struct {
		description  string
		denom        string
		expectedPool uint64
		exists       bool
	}{
		{
			description:  "Atom pool exists for denom Akash",
			denom:        "akash",
			expectedPool: 1,
			exists:       true,
		},
		{
			description:  "Atom pool exists for denom juno",
			denom:        "juno",
			expectedPool: 2,
			exists:       true,
		},
		{
			description:  "Atom pool exists for denom juno with different casing",
			denom:        "JuNo",
			expectedPool: 2,
			exists:       false,
		},
		{
			description:  "Atom pool exists for denom Ethereum",
			denom:        "ethereum",
			expectedPool: 3,
			exists:       true,
		},
		{
			description:  "Atom pool does not exist for denom doge",
			denom:        "doge",
			expectedPool: 0,
			exists:       false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			pool, err := suite.App.ProtoRevKeeper.GetAtomPool(suite.Ctx, tc.denom)

			if tc.exists {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedPool, pool)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestDeleteAllAtomPools tests the DeleteAllAtomPools function.
func (suite *KeeperTestSuite) TestDeleteAllAtomPools() {
	suite.App.AppKeepers.ProtoRevKeeper.DeleteAllAtomPools(suite.Ctx)

	// Iterate through all of the pools and check if any paired with Atom exist
	for _, pool := range suite.pools {
		if otherDenom, match := types.CheckOsmoAtomDenomMatch(pool.Asset1, pool.Asset2, types.AtomDenomination); match {
			_, err := suite.App.AppKeepers.ProtoRevKeeper.GetAtomPool(suite.Ctx, otherDenom)
			suite.Require().Error(err)
		}
	}
}

// TestGetOsmoPool tests the GetOsmoPool function.
func (suite *KeeperTestSuite) TestGetOsmoPool() {
	cases := []struct {
		description  string
		denom        string
		expectedPool uint64
		exists       bool
	}{
		{
			description:  "Osmo pool exists for denom Akash",
			denom:        "akash",
			expectedPool: 7,
			exists:       true,
		},
		{
			description:  "Osmo pool exists for denom juno",
			denom:        "juno",
			expectedPool: 8,
			exists:       true,
		},
		{
			description:  "Empty string returns error",
			denom:        "",
			expectedPool: 0,
			exists:       false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			pool, err := suite.App.ProtoRevKeeper.GetOsmoPool(suite.Ctx, tc.denom)

			if tc.exists {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedPool, pool)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}

// TestDeleteAllOsmoPools tests the DeleteAllOsmoPools function.
func (suite *KeeperTestSuite) TestDeleteAllOsmoPools() {
	suite.App.AppKeepers.ProtoRevKeeper.DeleteAllOsmoPools(suite.Ctx)

	// Iterate through all of the pools and check if any paired with Osmo exist
	for _, pool := range suite.pools {
		if otherDenom, match := types.CheckOsmoAtomDenomMatch(pool.Asset1, pool.Asset2, types.OsmosisDenomination); match {
			_, err := suite.App.AppKeepers.ProtoRevKeeper.GetOsmoPool(suite.Ctx, otherDenom)
			suite.Require().Error(err)
		}
	}
}

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

// TestGetDaysSinceModuleGenesis tests the GetDaysSinceModuleGenesis and SetDaysSinceModuleGenesis functions.
func (suite *KeeperTestSuite) TestGetDaysSinceModuleGenesis() {
	// Should be initalized to 0 on genesis
	daysSinceGenesis, err := suite.App.AppKeepers.ProtoRevKeeper.GetDaysSinceModuleGenesis(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), daysSinceGenesis)

	// Should be able to set the days since genesis
	suite.App.AppKeepers.ProtoRevKeeper.SetDaysSinceModuleGenesis(suite.Ctx, 1)
	daysSinceGenesis, err = suite.App.AppKeepers.ProtoRevKeeper.GetDaysSinceModuleGenesis(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), daysSinceGenesis)
}

// TestGetDeveloperFees tests the GetDeveloperFees, SetDeveloperFees, and GetAllDeveloperFees functions.
func (suite *KeeperTestSuite) TestGetDeveloperFees() {
	// Should be initalized to [] on genesis
	fees := suite.App.AppKeepers.ProtoRevKeeper.GetAllDeveloperFees(suite.Ctx)
	suite.Require().Equal(0, len(fees))

	// Should be no osmo fees on genesis
	osmoFees, err := suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, types.OsmosisDenomination)
	suite.Require().Error(err)
	suite.Require().Equal(sdk.Coin{}, osmoFees)

	// Should be no atom fees on genesis
	atomFees, err := suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, types.AtomDenomination)
	suite.Require().Error(err)
	suite.Require().Equal(sdk.Coin{}, atomFees)

	// Should be able to set the fees
	err = suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperFees(suite.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)))
	suite.Require().NoError(err)
	err = suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperFees(suite.Ctx, sdk.NewCoin(types.AtomDenomination, sdk.NewInt(100)))
	suite.Require().NoError(err)

	// Should be able to get the fees
	osmoFees, err = suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, types.OsmosisDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)), osmoFees)
	atomFees, err = suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, types.AtomDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(types.AtomDenomination, sdk.NewInt(100)), atomFees)
	fees = suite.App.AppKeepers.ProtoRevKeeper.GetAllDeveloperFees(suite.Ctx)
	suite.Require().Equal(2, len(fees))
	suite.Require().Contains(fees, osmoFees)
	suite.Require().Contains(fees, atomFees)
}

// TestDeleteDeveloperFees tests the DeleteDeveloperFees function.
func (suite *KeeperTestSuite) TestDeleteDeveloperFees() {
	err := suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperFees(suite.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)))
	suite.Require().NoError(err)

	// Should be able to get the fees
	osmoFees, err := suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, types.OsmosisDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)), osmoFees)

	// Should be able to delete the fees
	suite.App.AppKeepers.ProtoRevKeeper.DeleteDeveloperFees(suite.Ctx, types.OsmosisDenomination)

	// Should be no osmo fees after deletion
	osmoFees, err = suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, types.OsmosisDenomination)
	suite.Require().Error(err)
	suite.Require().Equal(sdk.Coin{}, osmoFees)
}

// TestGetProtoRevEnabled tests the GetProtoRevEnabled and SetProtoRevEnabled functions.
func (suite *KeeperTestSuite) TestGetProtoRevEnabled() {
	// Should be initalized to true on genesis
	protoRevEnabled, err := suite.App.AppKeepers.ProtoRevKeeper.GetProtoRevEnabled(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(true, protoRevEnabled)

	// Should be able to set the protoRevEnabled
	suite.App.AppKeepers.ProtoRevKeeper.SetProtoRevEnabled(suite.Ctx, false)
	protoRevEnabled, err = suite.App.AppKeepers.ProtoRevKeeper.GetProtoRevEnabled(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(false, protoRevEnabled)
}

// TestGetAdminAccount tests the GetAdminAccount and SetAdminAccount functions.
func (suite *KeeperTestSuite) TestGetAdminAccount() {
	// Should be initalized (look at keeper_test.go)
	adminAccount, err := suite.App.AppKeepers.ProtoRevKeeper.GetAdminAccount(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.adminAccount, adminAccount)

	// Should be able to set the admin account
	suite.App.AppKeepers.ProtoRevKeeper.SetAdminAccount(suite.Ctx, suite.TestAccs[0])
	adminAccount, err = suite.App.AppKeepers.ProtoRevKeeper.GetAdminAccount(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.TestAccs[0], adminAccount)
}

// TestGetDeveloperAccount tests the GetDeveloperAccount and SetDeveloperAccount functions.
func (suite *KeeperTestSuite) TestGetDeveloperAccount() {
	// Should be null on genesis
	developerAccount, err := suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
	suite.Require().Error(err)
	suite.Require().Nil(developerAccount)

	// Should be able to set the developer account
	suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, suite.TestAccs[0])
	developerAccount, err = suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.TestAccs[0], developerAccount)
}

// TestGetMaxRoutesPerTx tests the GetMaxRoutesPerTx and SetMaxRoutesPerTx functions.
func (suite *KeeperTestSuite) TestGetMaxRoutesPerTx() {
	// Should be initalized on genesis
	maxRoutes, err := suite.App.AppKeepers.ProtoRevKeeper.GetMaxRoutesPerTx(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(6), maxRoutes)

	// Should be able to set the maxRoutes
	suite.App.AppKeepers.ProtoRevKeeper.SetMaxRoutesPerTx(suite.Ctx, 4)
	maxRoutes, err = suite.App.AppKeepers.ProtoRevKeeper.GetMaxRoutesPerTx(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(4), maxRoutes)

	// Can only initalize between 1 and types.MaxIterableRoutesPerTx
	err = suite.App.AppKeepers.ProtoRevKeeper.SetMaxRoutesPerTx(suite.Ctx, 0)
	suite.Require().Error(err)
	err = suite.App.AppKeepers.ProtoRevKeeper.SetMaxRoutesPerTx(suite.Ctx, types.MaxIterableRoutesPerTx+1)
	suite.Require().Error(err)
}

// TestGetRouteCountForBlock tests the GetRouteCountForBlock, IncrementRouteCountForBlock and SetRouteCountForBlock functions.
func (suite *KeeperTestSuite) TestGetRouteCountForBlock() {
	// Should be initalized to 0 on genesis
	routeCount, err := suite.App.AppKeepers.ProtoRevKeeper.GetRouteCountForBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), routeCount)

	// Should be able to set the route count
	suite.App.AppKeepers.ProtoRevKeeper.SetRouteCountForBlock(suite.Ctx, 4)
	routeCount, err = suite.App.AppKeepers.ProtoRevKeeper.GetRouteCountForBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(4), routeCount)

	// Should be able to increment the route count
	err = suite.App.AppKeepers.ProtoRevKeeper.IncrementRouteCountForBlock(suite.Ctx, 10)
	suite.Require().NoError(err)
	routeCount, err = suite.App.AppKeepers.ProtoRevKeeper.GetRouteCountForBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(14), routeCount)
}

// TestGetLatestBlockHeight tests the GetLatestBlockHeight and SetLatestBlockHeight functions.
func (suite *KeeperTestSuite) TestGetLatestBlockHeight() {
	// Should be initalized to 0 on genesis
	blockHeight, err := suite.App.AppKeepers.ProtoRevKeeper.GetLatestBlockHeight(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), blockHeight)

	// Should be able to set the blockHeight
	suite.App.AppKeepers.ProtoRevKeeper.SetLatestBlockHeight(suite.Ctx, 4)
	blockHeight, err = suite.App.AppKeepers.ProtoRevKeeper.GetLatestBlockHeight(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(4), blockHeight)
}

// TestGetMaxRoutesPerBlock tests the GetMaxRoutesPerBlock and SetMaxRoutesPerBlock functions.
func (suite *KeeperTestSuite) TestGetMaxRoutesPerBlock() {
	// Should be initalized to 20 on genesis
	maxRoutes, err := suite.App.AppKeepers.ProtoRevKeeper.GetMaxRoutesPerBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(100), maxRoutes)

	// Should be able to set the maxRoutes
	suite.App.AppKeepers.ProtoRevKeeper.SetMaxRoutesPerBlock(suite.Ctx, 4)
	maxRoutes, err = suite.App.AppKeepers.ProtoRevKeeper.GetMaxRoutesPerBlock(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(4), maxRoutes)

	// Can only initalize between 1 and types.MaxIterableRoutesPerBlock
	err = suite.App.AppKeepers.ProtoRevKeeper.SetMaxRoutesPerBlock(suite.Ctx, 0)
	suite.Require().Error(err)
	err = suite.App.AppKeepers.ProtoRevKeeper.SetMaxRoutesPerBlock(suite.Ctx, types.MaxIterableRoutesPerBlock+1)
	suite.Require().Error(err)
}
