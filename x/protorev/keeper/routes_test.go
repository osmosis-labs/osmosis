package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
)

type TestRoute struct {
	PoolId      uint64
	InputDenom  string
	OutputDenom string
}

func (suite *KeeperTestSuite) TestBuildRoutes() {
	cases := []struct {
		description string
		inputDenom  string
		outputDenom string
		poolID      uint64
		expected    [][]TestRoute
	}{
		{
			description: "Route exists for swap in Akash and swap out Atom",
			inputDenom:  "akash",
			outputDenom: types.AtomDenomination,
			poolID:      1,
			expected: [][]TestRoute{
				{
					{PoolId: 1, InputDenom: types.AtomDenomination, OutputDenom: "akash"},
					{PoolId: 14, InputDenom: "akash", OutputDenom: "bitcoin"},
					{PoolId: 4, InputDenom: "bitcoin", OutputDenom: types.AtomDenomination},
				},
				{
					{PoolId: 25, InputDenom: types.OsmosisDenomination, OutputDenom: types.AtomDenomination},
					{PoolId: 1, InputDenom: types.AtomDenomination, OutputDenom: "akash"},
					{PoolId: 7, InputDenom: "akash", OutputDenom: types.OsmosisDenomination},
				},
			},
		},
		{
			description: "Route exists for swap in Bitcoin and swap out Atom",
			inputDenom:  "bitcoin",
			outputDenom: types.AtomDenomination,
			poolID:      4,
			expected: [][]TestRoute{
				{
					{PoolId: 25, InputDenom: types.OsmosisDenomination, OutputDenom: types.AtomDenomination},
					{PoolId: 4, InputDenom: types.AtomDenomination, OutputDenom: "bitcoin"},
					{PoolId: 10, InputDenom: "bitcoin", OutputDenom: types.OsmosisDenomination},
				},
			},
		},
		{
			description: "Route exists for swap in Bitcoin and swap out ethereum",
			inputDenom:  "bitcoin",
			outputDenom: "ethereum",
			poolID:      19,
			expected: [][]TestRoute{
				{
					{PoolId: 9, InputDenom: types.OsmosisDenomination, OutputDenom: "ethereum"},
					{PoolId: 19, InputDenom: "ethereum", OutputDenom: "bitcoin"},
					{PoolId: 10, InputDenom: "bitcoin", OutputDenom: types.OsmosisDenomination},
				},
				{
					{PoolId: 3, InputDenom: types.AtomDenomination, OutputDenom: "ethereum"},
					{PoolId: 19, InputDenom: "ethereum", OutputDenom: "bitcoin"},
					{PoolId: 4, InputDenom: "bitcoin", OutputDenom: types.AtomDenomination},
				},
			},
		},
		{
			description: "No route exists for swap in osmo and swap out Atom",
			inputDenom:  types.OsmosisDenomination,
			outputDenom: types.AtomDenomination,
			poolID:      25,
			expected:    [][]TestRoute{},
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			routes := suite.App.ProtoRevKeeper.BuildRoutes(suite.Ctx, tc.inputDenom, tc.outputDenom, tc.poolID)

			suite.Require().Equal(len(tc.expected), len(routes))

			for routeIndex, route := range routes {
				for tradeIndex, trade := range route {
					suite.Require().Equal(tc.expected[routeIndex][tradeIndex].PoolId, trade.PoolId)
					suite.Require().Equal(tc.expected[routeIndex][tradeIndex].OutputDenom, trade.TokenOutDenom)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBuildAtomRoute() {
	cases := []struct {
		description   string
		swapIn        string
		swapOut       string
		poolId        uint64
		expectedRoute []TestRoute
		hasRoute      bool
	}{
		{
			description:   "Route exists for swap in Osmo and swap out Akash",
			swapIn:        types.OsmosisDenomination,
			swapOut:       "akash",
			poolId:        7,
			expectedRoute: []TestRoute{{1, types.AtomDenomination, "akash"}, {7, "akash", types.OsmosisDenomination}, {25, types.OsmosisDenomination, types.AtomDenomination}},
			hasRoute:      true,
		},
		{
			description:   "Route exists for swap in Akash and swap out Osmo",
			swapIn:        "akash",
			swapOut:       types.OsmosisDenomination,
			poolId:        7,
			expectedRoute: []TestRoute{{25, types.AtomDenomination, types.OsmosisDenomination}, {7, types.OsmosisDenomination, "akash"}, {1, "akash", types.AtomDenomination}},
			hasRoute:      true,
		},
		{
			description:   "Route does not exist for swap in Terra and swap out Osmo because the pool does not exist",
			swapIn:        "terra",
			swapOut:       types.OsmosisDenomination,
			poolId:        7,
			expectedRoute: []TestRoute{},
			hasRoute:      false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			route, err := suite.App.ProtoRevKeeper.BuildAtomRoute(suite.Ctx, tc.swapIn, tc.swapOut, tc.poolId)

			if tc.hasRoute {
				suite.Require().NoError(err)
				suite.Require().Equal(len(tc.expectedRoute), len(route.PoolIds()))

				for index, trade := range tc.expectedRoute {
					suite.Require().Equal(trade.PoolId, route[index].PoolId)
					suite.Require().Equal(trade.OutputDenom, route[index].TokenOutDenom)
				}
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
		expectedRoute []TestRoute
		hasRoute      bool
	}{
		{
			description:   "Route exists for swap in Atom and swap out Akash",
			swapIn:        types.AtomDenomination,
			swapOut:       "akash",
			poolId:        1,
			expectedRoute: []TestRoute{{7, types.OsmosisDenomination, "akash"}, {1, "akash", types.AtomDenomination}, {25, types.AtomDenomination, types.OsmosisDenomination}},
			hasRoute:      true,
		},
		{
			description:   "Route exists for swap in Akash and swap out Atom",
			swapIn:        "akash",
			swapOut:       types.AtomDenomination,
			poolId:        1,
			expectedRoute: []TestRoute{{25, types.OsmosisDenomination, types.AtomDenomination}, {1, types.AtomDenomination, "akash"}, {7, "akash", types.OsmosisDenomination}},
			hasRoute:      true,
		},
		{
			description:   "Route does not exist for swap in Terra and swap out Atom because the pool does not exist",
			swapIn:        "terra",
			swapOut:       types.AtomDenomination,
			poolId:        7,
			expectedRoute: []TestRoute{},
			hasRoute:      false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			route, err := suite.App.ProtoRevKeeper.BuildOsmoRoute(suite.Ctx, tc.swapIn, tc.swapOut, tc.poolId)

			if tc.hasRoute {
				suite.Require().NoError(err)
				suite.Require().Equal(len(tc.expectedRoute), len(route.PoolIds()))

				for index, trade := range tc.expectedRoute {
					suite.Require().Equal(trade.PoolId, route[index].PoolId)
					suite.Require().Equal(trade.OutputDenom, route[index].TokenOutDenom)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBuildTokenPairRoutes() {
	cases := []struct {
		description    string
		swapIn         string
		swapOut        string
		poolId         uint64
		expectedRoutes [][]TestRoute
		hasRoutes      bool
	}{
		{
			description:    "Route exists for swap in Atom and swap out Akash",
			swapIn:         "akash",
			swapOut:        types.AtomDenomination,
			poolId:         1,
			expectedRoutes: [][]TestRoute{{{1, types.AtomDenomination, "akash"}, {14, "akash", "bitcoin"}, {4, "bitcoin", types.AtomDenomination}}},
			hasRoutes:      true,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			routes, err := suite.App.ProtoRevKeeper.BuildTokenPairRoutes(suite.Ctx, tc.swapIn, tc.swapOut, tc.poolId)

			if tc.hasRoutes {
				suite.Require().NoError(err)
				suite.Require().Equal(len(tc.expectedRoutes), len(routes))

				for index, route := range routes {

					suite.Require().Equal(len(tc.expectedRoutes[index]), len(route.PoolIds()))

					for index, trade := range tc.expectedRoutes[index] {
						suite.Require().Equal(trade.PoolId, route[index].PoolId)
						suite.Require().Equal(trade.OutputDenom, route[index].TokenOutDenom)
					}
				}

			} else {
				suite.Require().Error(err)
			}
		})
	}
}
