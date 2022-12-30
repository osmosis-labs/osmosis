package keeper_test

import "github.com/osmosis-labs/osmosis/v13/x/protorev/types"

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
	suite.Require().Equal(suite.tokenPairArbRoutes, tokenPairArbRoutes)
}

// TestDeleteAllTokenPairArbRoutes tests the DeleteAllTokenPairArbRoutes function.
func (suite *KeeperTestSuite) TestDeleteAllTokenPairArbRoutes() {
	// Tests that we can properly retrieve all of the routes that were set up
	tokenPairArbRoutes, err := suite.App.ProtoRevKeeper.GetAllTokenPairArbRoutes(suite.Ctx)

	suite.Require().NoError(err)
	suite.Require().Equal(len(suite.tokenPairArbRoutes), len(tokenPairArbRoutes))
	suite.Require().Equal(suite.tokenPairArbRoutes, tokenPairArbRoutes)

	// Delete all routes
	suite.App.ProtoRevKeeper.DeleteAllTokenPairArbRoutes(suite.Ctx)

	// Test after deletion
	tokenPairArbRoutes, err = suite.App.ProtoRevKeeper.GetAllTokenPairArbRoutes(suite.Ctx)

	suite.Require().NoError(err)
	suite.Require().Equal(0, len(tokenPairArbRoutes))
}
