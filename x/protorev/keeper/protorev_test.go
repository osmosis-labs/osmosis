package keeper_test

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
			exists:       true,
		},
		{
			description:  "Atom pool exists for denom Ethereum",
			denom:        "ethEreUm",
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

func (suite *KeeperTestSuite) TestDeleteAtomPool() {
	cases := []struct {
		description string
		denom       string
		exists      bool
	}{
		{
			description: "Atom pool exists for denom Akash",
			denom:       "akash",
			exists:      true,
		},
		{
			description: "Atom pool exists for denom juno",
			denom:       "juno",
			exists:      true,
		},
		{
			description: "Empty string returns error",
			denom:       "",
			exists:      false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			_, err := suite.App.ProtoRevKeeper.GetAtomPool(suite.Ctx, tc.denom)
			if tc.exists {
				suite.Require().NoError(err)

				suite.App.ProtoRevKeeper.DeleteAtomPool(suite.Ctx, tc.denom)

				_, err := suite.App.ProtoRevKeeper.GetAtomPool(suite.Ctx, tc.denom)
				suite.Require().Error(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

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

func (suite *KeeperTestSuite) TestDeleteOsmoPool() {
	cases := []struct {
		description string
		denom       string
		exists      bool
	}{
		{
			description: "Osmo pool exists for denom Akash",
			denom:       "akash",
			exists:      true,
		},
		{
			description: "Osmo pool exists for denom juno",
			denom:       "juno",
			exists:      true,
		},
		{
			description: "Empty string returns error",
			denom:       "",
			exists:      false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			_, err := suite.App.ProtoRevKeeper.GetOsmoPool(suite.Ctx, tc.denom)
			if tc.exists {
				suite.Require().NoError(err)

				suite.App.ProtoRevKeeper.DeleteOsmoPool(suite.Ctx, tc.denom)

				_, err := suite.App.ProtoRevKeeper.GetOsmoPool(suite.Ctx, tc.denom)
				suite.Require().Error(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetSearcherRoutes() {

	// Tests that we can properly retrieve all of the routes that were set up
	for _, searcherRoutes := range suite.searcherRoutes {
		routes, err := suite.App.ProtoRevKeeper.GetSearcherRoutes(suite.Ctx, searcherRoutes.TokenA, searcherRoutes.TokenB)

		suite.Require().NoError(err)
		suite.Require().Equal(searcherRoutes, *routes)
	}

	// Testing to see if we will not find a route that does not exist
	_, err := suite.App.ProtoRevKeeper.GetSearcherRoutes(suite.Ctx, "osmo", "abc")
	suite.Require().Error(err)
}
