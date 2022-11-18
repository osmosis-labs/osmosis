package keeper_test

import "github.com/osmosis-labs/osmosis/v12/x/protorev/types"

func (suite *KeeperTestSuite) TestBuildAtomRoute() {
	cases := []struct {
		description   string
		swapIn        string
		swapOut       string
		poolId        uint64
		expectedRoute []uint64
		hasRoute      bool
	}{
		{
			description:   "Route exists for swap in Osmo and swap out Akash",
			swapIn:        types.OsmosisDenomination,
			swapOut:       "akash",
			poolId:        7,
			expectedRoute: []uint64{1, 7, 6},
			hasRoute:      true,
		},
		{
			description:   "Route exists for swap in Akash and swap out Osmo",
			swapIn:        "akash",
			swapOut:       types.OsmosisDenomination,
			poolId:        7,
			expectedRoute: []uint64{6, 7, 1},
			hasRoute:      true,
		},
		{
			description:   "Route exists for swap in Terra and swap out Osmo (no mapping pool)",
			swapIn:        "terra",
			swapOut:       types.OsmosisDenomination,
			poolId:        7,
			expectedRoute: []uint64{},
			hasRoute:      false,
		},
		{
			description:   "Route exists for swap in Akash and swap out Atom (invalid route)",
			swapIn:        "terra",
			swapOut:       types.OsmosisDenomination,
			poolId:        7,
			expectedRoute: []uint64{},
			hasRoute:      false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			route, err := suite.App.ProtoRevKeeper.BuildAtomRoute(suite.Ctx, tc.swapIn, tc.swapOut, tc.poolId)

			if tc.hasRoute {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedRoute, route)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBuildOsmoRoute() {
	cases := []struct {
		description   string
		swapIn        string
		swapOut       string
		poolId        uint64
		expectedRoute []uint64
		hasRoute      bool
	}{
		{
			description:   "Route exists for swap in Atom and swap out Akash",
			swapIn:        types.AtomDenomination,
			swapOut:       "akash",
			poolId:        1,
			expectedRoute: []uint64{7, 1, 6},
			hasRoute:      true,
		},
		{
			description:   "Route exists for swap in Akash and swap out Atom",
			swapIn:        "akash",
			swapOut:       types.AtomDenomination,
			poolId:        1,
			expectedRoute: []uint64{6, 1, 7},
			hasRoute:      true,
		},
		{
			description:   "Route exists for swap in Terra and swap out Atom (no mapping pool)",
			swapIn:        "terra",
			swapOut:       types.OsmosisDenomination,
			poolId:        7,
			expectedRoute: []uint64{},
			hasRoute:      false,
		},
		{
			description:   "Route exists for swap in Akash and swap out Atom (invalid route)",
			swapIn:        "terra",
			swapOut:       types.OsmosisDenomination,
			poolId:        7,
			expectedRoute: []uint64{},
			hasRoute:      false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			route, err := suite.App.ProtoRevKeeper.BuildOsmoRoute(suite.Ctx, tc.swapIn, tc.swapOut, tc.poolId)

			if tc.hasRoute {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedRoute, route)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBuildSearchersRoutes() {
	cases := []struct {
		description    string
		swapIn         string
		swapOut        string
		poolId         uint64
		expectedRoutes [][]uint64
		hasRoutes      bool
	}{
		{
			description:    "Route exists for swap in Atom and swap out Akash",
			swapIn:         types.AtomDenomination,
			swapOut:        "akash",
			poolId:         1,
			expectedRoutes: [][]uint64{{1, 14, 4}, {1, 13, 3}},
			hasRoutes:      true,
		},
		{
			description:    "Route exists for swap in Atom and swap out Akash",
			swapIn:         types.OsmosisDenomination,
			swapOut:        "juno",
			poolId:         8,
			expectedRoutes: [][]uint64{{7, 12, 8}},
			hasRoutes:      true,
		},
		{
			description:    "Route exists for swap in Atom and swap out Akash",
			swapIn:         types.OsmosisDenomination,
			swapOut:        "terra",
			poolId:         800,
			expectedRoutes: [][]uint64{},
			hasRoutes:      false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			routes, err := suite.App.ProtoRevKeeper.BuildSearcherRoutes(suite.Ctx, tc.swapIn, tc.swapOut, tc.poolId)

			if tc.hasRoutes {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedRoutes, routes)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
