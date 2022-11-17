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

func (suite *KeeperTestSuite) TestGetSearcherRoutes() {
	routes := []*types.Route{
		{
			Pools: []uint64{0, 1, 2},
		},
		{
			Pools: []uint64{1, 2, 3},
		},
		{
			Pools: []uint64{2, 3, 4},
		},
		{
			Pools: []uint64{3, 4, 5},
		},
		{
			Pools: []uint64{4, 5, 6},
		},
	}

	cases := []struct {
		description    string
		tokenA         string
		tokenB         string
		searcherRoutes types.SearcherRoutes
		exists         bool
	}{
		{
			description: "Route exists for denom Akash",
			tokenA:      "Akash",
			tokenB:      "atom",
			searcherRoutes: types.SearcherRoutes{
				TokenA: "AKASH",
				TokenB: "ATOM",
				Routes: routes,
			},
			exists: true,
		},
	}

	// Insert route data
	suite.SetUpSearcherRoutes()

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			searcherRoutes, err := suite.App.ProtoRevKeeper.GetSearcherRoutes(suite.Ctx, tc.tokenA, tc.tokenB)

			if tc.exists {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.searcherRoutes.TokenA, searcherRoutes.TokenA)
				suite.Require().Equal(tc.searcherRoutes.TokenB, searcherRoutes.TokenB)
				suite.Require().Equal(tc.searcherRoutes.Routes, searcherRoutes.Routes)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
