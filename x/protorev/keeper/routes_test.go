package keeper_test

import (
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

type TestRoute struct {
	PoolId      uint64
	InputDenom  string
	OutputDenom string
}

// TestBuildRoutes tests the BuildRoutes function
func (suite *KeeperTestSuite) TestBuildRoutes() {
	cases := []struct {
		description        string
		inputDenom         string
		outputDenom        string
		poolID             uint64
		expected           [][]TestRoute
		expectedRouteCount uint64
		maxIterableRoutes  uint64
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
			expectedRouteCount: 12,
			maxIterableRoutes:  15,
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
			expectedRouteCount: 6,
			maxIterableRoutes:  15,
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
			expectedRouteCount: 12,
			maxIterableRoutes:  15,
		},
		{
			description:        "No route exists for swap in osmo and swap out Atom",
			inputDenom:         types.OsmosisDenomination,
			outputDenom:        types.AtomDenomination,
			poolID:             25,
			expected:           [][]TestRoute{},
			expectedRouteCount: 0,
			maxIterableRoutes:  15,
		},
		{
			description: "Route exists for swap on stable pool",
			inputDenom:  "usdc",
			outputDenom: types.OsmosisDenomination,
			poolID:      29,
			expected: [][]TestRoute{
				{
					{PoolId: 29, InputDenom: types.OsmosisDenomination, OutputDenom: "usdc"},
					{PoolId: 34, InputDenom: "usdc", OutputDenom: "busd"},
					{PoolId: 30, InputDenom: "busd", OutputDenom: types.OsmosisDenomination},
				},
			},
			expectedRouteCount: 7,
			maxIterableRoutes:  15,
		},
		{
			description:        "Route exists for swap on stable pool but not enough routes left to be explored",
			inputDenom:         "usdc",
			outputDenom:        types.OsmosisDenomination,
			poolID:             29,
			expected:           [][]TestRoute{},
			expectedRouteCount: 0,
			maxIterableRoutes:  3,
		},
		{
			description: "Two routes exist but only 1 route left to be explored (osmo route chosen)",
			inputDenom:  "bitcoin",
			outputDenom: "ethereum",
			poolID:      19,
			expected: [][]TestRoute{
				{
					{PoolId: 9, InputDenom: types.OsmosisDenomination, OutputDenom: "ethereum"},
					{PoolId: 19, InputDenom: "ethereum", OutputDenom: "bitcoin"},
					{PoolId: 10, InputDenom: "bitcoin", OutputDenom: types.OsmosisDenomination},
				},
			},
			expectedRouteCount: 6,
			maxIterableRoutes:  6,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()

			err := suite.App.ProtoRevKeeper.SetPoolWeights(suite.Ctx, types.PoolWeights{StableWeight: 3, BalancerWeight: 2, ConcentratedWeight: 1})
			suite.Require().NoError(err)

			routes := suite.App.ProtoRevKeeper.BuildRoutes(suite.Ctx, tc.inputDenom, tc.outputDenom, tc.poolID, &tc.maxIterableRoutes)

			suite.Require().Equal(len(tc.expected), len(routes))
			routeCount, err := suite.App.ProtoRevKeeper.GetRouteCountForBlock(suite.Ctx)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedRouteCount, routeCount)

			for routeIndex, route := range routes {
				for tradeIndex, trade := range route {
					suite.Require().Equal(tc.expected[routeIndex][tradeIndex].PoolId, trade.PoolId)
					suite.Require().Equal(tc.expected[routeIndex][tradeIndex].OutputDenom, trade.TokenOutDenom)
				}
			}
		})
	}
}

// TestBuildHighestLiquidityRoute tests the BuildHighestLiquidityRoute function
func (suite *KeeperTestSuite) TestBuildHighestLiquidityRoute() {
	cases := []struct {
		description        string
		swapDenom          string
		swapIn             string
		swapOut            string
		poolId             uint64
		expectedRoute      []TestRoute
		hasRoute           bool
		expectedRouteCount uint64
	}{
		{
			description:        "Route exists for swap in Atom and swap out Akash",
			swapDenom:          types.OsmosisDenomination,
			swapIn:             types.AtomDenomination,
			swapOut:            "akash",
			poolId:             1,
			expectedRoute:      []TestRoute{{7, types.OsmosisDenomination, "akash"}, {1, "akash", types.AtomDenomination}, {25, types.AtomDenomination, types.OsmosisDenomination}},
			hasRoute:           true,
			expectedRouteCount: 6,
		},
		{
			description:        "Route exists for swap in Akash and swap out Atom",
			swapDenom:          types.OsmosisDenomination,
			swapIn:             "akash",
			swapOut:            types.AtomDenomination,
			poolId:             1,
			expectedRoute:      []TestRoute{{25, types.OsmosisDenomination, types.AtomDenomination}, {1, types.AtomDenomination, "akash"}, {7, "akash", types.OsmosisDenomination}},
			hasRoute:           true,
			expectedRouteCount: 6,
		},
		{
			description:        "Route does not exist for swap in Terra and swap out Atom because the pool does not exist",
			swapDenom:          types.OsmosisDenomination,
			swapIn:             "terra",
			swapOut:            types.AtomDenomination,
			poolId:             7,
			expectedRoute:      []TestRoute{},
			hasRoute:           false,
			expectedRouteCount: 0,
		},
		{
			description:        "Route exists for swap in Osmo and swap out Akash",
			swapDenom:          types.AtomDenomination,
			swapIn:             types.OsmosisDenomination,
			swapOut:            "akash",
			poolId:             7,
			expectedRoute:      []TestRoute{{1, types.AtomDenomination, "akash"}, {7, "akash", types.OsmosisDenomination}, {25, types.OsmosisDenomination, types.AtomDenomination}},
			hasRoute:           true,
			expectedRouteCount: 6,
		},
		{
			description:        "Route exists for swap in Akash and swap out Osmo",
			swapDenom:          types.AtomDenomination,
			swapIn:             "akash",
			swapOut:            types.OsmosisDenomination,
			poolId:             7,
			expectedRoute:      []TestRoute{{25, types.AtomDenomination, types.OsmosisDenomination}, {7, types.OsmosisDenomination, "akash"}, {1, "akash", types.AtomDenomination}},
			hasRoute:           true,
			expectedRouteCount: 6,
		},
		{
			description:        "Route does not exist for swap in Terra and swap out Osmo because the pool does not exist",
			swapDenom:          types.AtomDenomination,
			swapIn:             "terra",
			swapOut:            types.OsmosisDenomination,
			poolId:             7,
			expectedRoute:      []TestRoute{},
			hasRoute:           false,
			expectedRouteCount: 0,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()

			numberIterable, err := suite.App.ProtoRevKeeper.CalcNumberOfIterableRoutes(suite.Ctx)
			suite.Require().NoError(err)
			before := *numberIterable

			route, buildErr := suite.App.ProtoRevKeeper.BuildHighestLiquidityRoute(suite.Ctx, tc.swapDenom, tc.swapIn, tc.swapOut, tc.poolId, numberIterable)
			routeCount, err := suite.App.ProtoRevKeeper.GetRouteCountForBlock(suite.Ctx)
			suite.Require().NoError(err)

			suite.Require().Equal(tc.expectedRouteCount, routeCount)
			suite.Require().Equal(*numberIterable+routeCount, before)

			if tc.hasRoute {
				suite.Require().NoError(buildErr)
				suite.Require().Equal(len(tc.expectedRoute), len(route.PoolIds()))

				for index, trade := range tc.expectedRoute {
					suite.Require().Equal(trade.PoolId, route[index].PoolId)
					suite.Require().Equal(trade.OutputDenom, route[index].TokenOutDenom)
				}
			} else {
				suite.Require().Error(buildErr)
			}
		})
	}
}

// TestBuildHotRoutes tests the BuildHotRoutes function
func (suite *KeeperTestSuite) TestBuildHotRoutes() {
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
			numberIterable, err := suite.App.ProtoRevKeeper.CalcNumberOfIterableRoutes(suite.Ctx)
			suite.Require().NoError(err)

			routes, err := suite.App.ProtoRevKeeper.BuildHotRoutes(suite.Ctx, tc.swapIn, tc.swapOut, tc.poolId, numberIterable)

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

// TestCheckAndUpdateRouteState tests the CheckAndUpdateRouteState function
func (suite *KeeperTestSuite) TestCheckAndUpdateRouteState() {
	cases := []struct {
		description                     string
		route                           poolmanagertypes.SwapAmountInRoutes
		maxIterableRoutes               uint64
		expectedRemainingIterableRoutes uint64
		expectedPass                    bool
	}{
		{
			description:                     "Valid route containing only balancer pools",
			route:                           []poolmanagertypes.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: ""}, {PoolId: 2, TokenOutDenom: ""}, {PoolId: 3, TokenOutDenom: ""}},
			maxIterableRoutes:               10,
			expectedRemainingIterableRoutes: 4,
			expectedPass:                    true,
		},
		{
			description:                     "Valid route containing only balancer pools but not enough iterable routes",
			route:                           []poolmanagertypes.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: ""}, {PoolId: 2, TokenOutDenom: ""}, {PoolId: 3, TokenOutDenom: ""}},
			maxIterableRoutes:               2,
			expectedRemainingIterableRoutes: 2,
			expectedPass:                    false,
		},
		{
			description:                     "Valid route containing only balancer pools and equal number of iterable routes",
			route:                           []poolmanagertypes.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: ""}, {PoolId: 2, TokenOutDenom: ""}, {PoolId: 3, TokenOutDenom: ""}},
			maxIterableRoutes:               6,
			expectedRemainingIterableRoutes: 0,
			expectedPass:                    true,
		},
		{
			description:                     "Valid route containing only stable swap pools",
			route:                           []poolmanagertypes.SwapAmountInRoute{{PoolId: 34, TokenOutDenom: ""}, {PoolId: 34, TokenOutDenom: ""}, {PoolId: 34, TokenOutDenom: ""}},
			maxIterableRoutes:               10,
			expectedRemainingIterableRoutes: 1,
			expectedPass:                    true,
		},
		{
			description:                     "Valid route with more than 3 hops",
			route:                           []poolmanagertypes.SwapAmountInRoute{{PoolId: 34, TokenOutDenom: ""}, {PoolId: 34, TokenOutDenom: ""}, {PoolId: 34, TokenOutDenom: ""}, {PoolId: 1, TokenOutDenom: ""}},
			maxIterableRoutes:               12,
			expectedRemainingIterableRoutes: 1,
			expectedPass:                    true,
		},
		{
			description:                     "Valid route with more than 3 hops",
			route:                           []poolmanagertypes.SwapAmountInRoute{{PoolId: 34, TokenOutDenom: ""}, {PoolId: 34, TokenOutDenom: ""}, {PoolId: 34, TokenOutDenom: ""}, {PoolId: 1, TokenOutDenom: ""}, {PoolId: 2, TokenOutDenom: ""}, {PoolId: 3, TokenOutDenom: ""}},
			maxIterableRoutes:               12,
			expectedRemainingIterableRoutes: 12,
			expectedPass:                    false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()

			err := suite.App.ProtoRevKeeper.SetPoolWeights(suite.Ctx, types.PoolWeights{StableWeight: 3, BalancerWeight: 2, ConcentratedWeight: 1})
			suite.Require().NoError(err)

			var maxIterable *uint64 = &tc.maxIterableRoutes

			err = suite.App.ProtoRevKeeper.CheckAndUpdateRouteState(suite.Ctx, tc.route, maxIterable)
			if tc.expectedPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			suite.Require().Equal(tc.expectedRemainingIterableRoutes, tc.maxIterableRoutes)
		})
	}
}

// TestCalcNumberOfIterableRoutes tests the CalcNumberOfIterableRoutes function.
func (suite *KeeperTestSuite) TestCalcNumberOfIterableRoutes() {
	cases := []struct {
		description        string
		maxRoutesPerTx     uint64
		maxRoutesPerBlock  uint64
		currentRouteCount  uint64
		expectedRouteCount uint64
	}{
		{
			description:        "Max routes per tx is 10 and max routes per block is 100",
			maxRoutesPerTx:     10,
			maxRoutesPerBlock:  100,
			currentRouteCount:  0,
			expectedRouteCount: 10,
		},
		{
			description:        "Max routes per tx is 10, max routes per block is 100, and current route count is 90",
			maxRoutesPerTx:     10,
			maxRoutesPerBlock:  100,
			currentRouteCount:  90,
			expectedRouteCount: 10,
		},
		{
			description:        "Max routes per tx is 10, max routes per block is 100, and current route count is 100",
			maxRoutesPerTx:     10,
			maxRoutesPerBlock:  100,
			currentRouteCount:  100,
			expectedRouteCount: 0,
		},
		{
			description:        "Max routes per tx is 10, max routes per block is 100, and current route count is 95",
			maxRoutesPerTx:     10,
			maxRoutesPerBlock:  100,
			currentRouteCount:  95,
			expectedRouteCount: 5,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()

			suite.App.ProtoRevKeeper.SetMaxRoutesPerTx(suite.Ctx, tc.maxRoutesPerTx)
			suite.App.ProtoRevKeeper.SetMaxRoutesPerBlock(suite.Ctx, tc.maxRoutesPerBlock)
			suite.App.ProtoRevKeeper.SetRouteCountForBlock(suite.Ctx, tc.currentRouteCount)

			numberIterable, err := suite.App.ProtoRevKeeper.CalcNumberOfIterableRoutes(suite.Ctx)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedRouteCount, *numberIterable)
		})
	}
}
