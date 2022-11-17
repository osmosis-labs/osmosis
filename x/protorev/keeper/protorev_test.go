package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"
)

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
			expectedPool: 6,
			exists:       true,
		},
		{
			description:  "Atom pool exists for denom juno",
			denom:        "juno",
			expectedPool: 7,
			exists:       true,
		},
		{
			description:  "Atom pool exists for denom juno with different casing",
			denom:        "JuNo",
			expectedPool: 7,
			exists:       true,
		},
		{
			description:  "Atom pool exists for denom Ethereum",
			denom:        "ethEreUm",
			expectedPool: 8,
			exists:       true,
		},
		{
			description:  "Atom pool does not exist for denom doge",
			denom:        "doge",
			expectedPool: 0,
			exists:       false,
		},
	}

	// Insert the atom pool data
	suite.SetUpAtomPools()

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

	// Insert the atom pool data
	suite.SetUpAtomPools()

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
			expectedPool: 1,
			exists:       true,
		},
		{
			description:  "Osmo pool exists for denom juno",
			denom:        "juno",
			expectedPool: 2,
			exists:       true,
		},
		{
			description:  "Empty string returns error",
			denom:        "",
			expectedPool: 0,
			exists:       false,
		},
	}

	// Insert the atom pool data
	suite.SetUpOsmoPools()

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

	// Insert the atom pool data
	suite.SetUpOsmoPools()

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

func (suite *KeeperTestSuite) TestGetRoute() {
	cases := []struct {
		description string
		denom1      string
		denom2      string
		route       types.Route
		exists      bool
	}{
		{
			description: "Route exists for denom Akash",
			denom1:      "atom",
			denom2:      "akash",
			route: types.Route{
				ArbDenom:     "ATOM",
				SwapInDenom:  "AKASH",
				SwapOutDenom: "ATOM",
				Pools:        []uint64{0, 1, 2, 3, 4},
			},
			exists: true,
		},
		{
			description: "Route exists for denom juno",
			denom1:      "osmo",
			denom2:      "juno",
			route: types.Route{
				ArbDenom:     "OSMO",
				SwapInDenom:  "OSMO",
				SwapOutDenom: "JUNO",
				Pools:        []uint64{0, 1, 2, 3, 4},
			},
			exists: true,
		},
		{
			description: "Empty string returns error",
			denom1:      "",
			denom2:      "",
			route:       types.Route{},
			exists:      false,
		},
		{
			description: "No matching route",
			denom1:      "bussincoin",
			denom2:      "skip",
			route:       types.Route{},
			exists:      false,
		},
	}

	// Insert route data
	suite.SetUpRoutes()

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			route, err := suite.App.ProtoRevKeeper.GetRoute(suite.Ctx, tc.denom1, tc.denom2)

			if tc.exists {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.route, *route)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
