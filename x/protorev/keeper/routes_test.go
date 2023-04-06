package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

type TestRoute struct {
	PoolId      uint64
	InputDenom  string
	OutputDenom string
}

// TestBuildRoutes tests the BuildRoutes function
func (suite *KeeperTestSuite) TestBuildRoutes() {
	cases := []struct {
		description    string
		inputDenom     string
		outputDenom    string
		poolID         uint64
		expectedRoutes [][]TestRoute
	}{
		{
			description: "Route exists for swap in Akash and swap out Atom",
			inputDenom:  "akash",
			outputDenom: "Atom",
			poolID:      1,
			expectedRoutes: [][]TestRoute{
				{
					{PoolId: 1, InputDenom: "Atom", OutputDenom: "akash"},
					{PoolId: 14, InputDenom: "akash", OutputDenom: "bitcoin"},
					{PoolId: 4, InputDenom: "bitcoin", OutputDenom: "Atom"},
				},
				{
					{PoolId: 25, InputDenom: types.OsmosisDenomination, OutputDenom: "Atom"},
					{PoolId: 1, InputDenom: "Atom", OutputDenom: "akash"},
					{PoolId: 7, InputDenom: "akash", OutputDenom: types.OsmosisDenomination},
				},
			},
		},
		{
			description: "Route exists for swap in Bitcoin and swap out Atom",
			inputDenom:  "bitcoin",
			outputDenom: "Atom",
			poolID:      4,
			expectedRoutes: [][]TestRoute{
				{
					{PoolId: 25, InputDenom: types.OsmosisDenomination, OutputDenom: "Atom"},
					{PoolId: 4, InputDenom: "Atom", OutputDenom: "bitcoin"},
					{PoolId: 10, InputDenom: "bitcoin", OutputDenom: types.OsmosisDenomination},
				},
			},
		},
		{
			description: "Route exists for swap in Bitcoin and swap out ethereum",
			inputDenom:  "bitcoin",
			outputDenom: "ethereum",
			poolID:      19,
			expectedRoutes: [][]TestRoute{
				{
					{PoolId: 9, InputDenom: types.OsmosisDenomination, OutputDenom: "ethereum"},
					{PoolId: 19, InputDenom: "ethereum", OutputDenom: "bitcoin"},
					{PoolId: 10, InputDenom: "bitcoin", OutputDenom: types.OsmosisDenomination},
				},
				{
					{PoolId: 3, InputDenom: "Atom", OutputDenom: "ethereum"},
					{PoolId: 19, InputDenom: "ethereum", OutputDenom: "bitcoin"},
					{PoolId: 4, InputDenom: "bitcoin", OutputDenom: "Atom"},
				},
			},
		},
		{
			description:    "No route exists for swap in osmo and swap out Atom",
			inputDenom:     types.OsmosisDenomination,
			outputDenom:    "Atom",
			poolID:         25,
			expectedRoutes: [][]TestRoute{},
		},
		{
			description: "Route exists for swap on stable pool",
			inputDenom:  "usdc",
			outputDenom: types.OsmosisDenomination,
			poolID:      29,
			expectedRoutes: [][]TestRoute{
				{
					{PoolId: 29, InputDenom: types.OsmosisDenomination, OutputDenom: "usdc"},
					{PoolId: 40, InputDenom: "usdc", OutputDenom: "busd"},
					{PoolId: 30, InputDenom: "busd", OutputDenom: types.OsmosisDenomination},
				},
			},
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			routes := suite.App.ProtoRevKeeper.BuildRoutes(suite.Ctx, tc.inputDenom, tc.outputDenom, tc.poolID)
			suite.Require().Equal(len(tc.expectedRoutes), len(routes))

			for routeIndex, route := range routes {
				for tradeIndex, poolID := range route.Route.PoolIds() {
					suite.Require().Equal(tc.expectedRoutes[routeIndex][tradeIndex].PoolId, poolID)
				}
			}
		})
	}
}

// TestBuildHighestLiquidityRoute tests the BuildHighestLiquidityRoute function
func (suite *KeeperTestSuite) TestBuildHighestLiquidityRoute() {
	cases := []struct {
		description              string
		swapDenom                string
		swapIn                   string
		swapOut                  string
		poolId                   uint64
		expectedRoute            []TestRoute
		hasRoute                 bool
		expectedRoutePointPoints uint64
	}{
		{
			description: "Route exists for swap in Atom and swap out Akash",
			swapDenom:   types.OsmosisDenomination,
			swapIn:      "Atom",
			swapOut:     "akash",
			poolId:      1,
			expectedRoute: []TestRoute{
				{7, types.OsmosisDenomination, "akash"},
				{1, "akash", "Atom"},
				{25, "Atom", types.OsmosisDenomination},
			},
			hasRoute:                 true,
			expectedRoutePointPoints: 6,
		},
		{
			description: "Route exists for swap in Akash and swap out Atom",
			swapDenom:   types.OsmosisDenomination,
			swapIn:      "akash",
			swapOut:     "Atom",
			poolId:      1,
			expectedRoute: []TestRoute{
				{25, types.OsmosisDenomination, "Atom"},
				{1, "Atom", "akash"},
				{7, "akash", types.OsmosisDenomination},
			},
			hasRoute:                 true,
			expectedRoutePointPoints: 6,
		},
		{
			description:              "Route does not exist for swap in Terra and swap out Atom because the pool does not exist",
			swapDenom:                types.OsmosisDenomination,
			swapIn:                   "terra",
			swapOut:                  "Atom",
			poolId:                   7,
			expectedRoute:            []TestRoute{},
			hasRoute:                 false,
			expectedRoutePointPoints: 0,
		},
		{
			description: "Route exists for swap in Osmo and swap out Akash",
			swapDenom:   "Atom",
			swapIn:      types.OsmosisDenomination,
			swapOut:     "akash",
			poolId:      7,
			expectedRoute: []TestRoute{
				{1, "Atom", "akash"},
				{7, "akash", types.OsmosisDenomination},
				{25, types.OsmosisDenomination, "Atom"},
			},
			hasRoute:                 true,
			expectedRoutePointPoints: 6,
		},
		{
			description: "Route exists for swap in Akash and swap out Osmo",
			swapDenom:   "Atom",
			swapIn:      "akash",
			swapOut:     types.OsmosisDenomination,
			poolId:      7,
			expectedRoute: []TestRoute{
				{25, "Atom", types.OsmosisDenomination},
				{7, types.OsmosisDenomination, "akash"},
				{1, "akash", "Atom"},
			},
			hasRoute:                 true,
			expectedRoutePointPoints: 6,
		},
		{
			description:              "Route does not exist for swap in Terra and swap out Osmo because the pool does not exist",
			swapDenom:                "Atom",
			swapIn:                   "terra",
			swapOut:                  types.OsmosisDenomination,
			poolId:                   7,
			expectedRoute:            []TestRoute{},
			hasRoute:                 false,
			expectedRoutePointPoints: 0,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.App.ProtoRevKeeper.SetPoolWeights(suite.Ctx, types.PoolWeights{
				StableWeight:       5,
				BalancerWeight:     2,
				ConcentratedWeight: 2,
			})

			baseDenom := types.BaseDenom{
				Denom:    tc.swapDenom,
				StepSize: sdk.NewInt(1_000_000),
			}
			routeMetaData, err := suite.App.ProtoRevKeeper.BuildHighestLiquidityRoute(suite.Ctx, baseDenom, tc.swapIn, tc.swapOut, tc.poolId)

			if tc.hasRoute {
				suite.Require().NoError(err)
				suite.Require().Equal(len(tc.expectedRoute), len(routeMetaData.Route.PoolIds()))

				for index, trade := range tc.expectedRoute {
					suite.Require().Equal(trade.PoolId, routeMetaData.Route.PoolIds()[index])
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestBuildHotRoutes tests the BuildHotRoutes function
func (suite *KeeperTestSuite) TestBuildHotRoutes() {
	cases := []struct {
		description             string
		swapIn                  string
		swapOut                 string
		poolId                  uint64
		expectedRoutes          [][]TestRoute
		expectedStepSize        []sdk.Int
		expectedRoutePoolPoints []uint64
		hasRoutes               bool
	}{
		{
			description: "Route exists for swap in Atom and swap out Akash",
			swapIn:      "akash",
			swapOut:     "Atom",
			poolId:      1,
			expectedRoutes: [][]TestRoute{
				{
					{1, "Atom", "akash"},
					{14, "akash", "bitcoin"},
					{4, "bitcoin", "Atom"},
				},
			},
			expectedStepSize:        []sdk.Int{sdk.NewInt(1_000_000)},
			expectedRoutePoolPoints: []uint64{6},
			hasRoutes:               true,
		},
		{
			description: "Route exists for a four pool route",
			swapIn:      "Atom",
			swapOut:     "test/2",
			poolId:      10,
			expectedRoutes: [][]TestRoute{
				{
					{34, "Atom", "test/1"},
					{35, "test/1", types.OsmosisDenomination},
					{36, types.OsmosisDenomination, "test/2"},
					{10, "test/2", "Atom"},
				},
			},
			expectedStepSize:        []sdk.Int{sdk.NewInt(1_000_000)},
			expectedRoutePoolPoints: []uint64{8},
			hasRoutes:               true,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.App.ProtoRevKeeper.SetPoolWeights(suite.Ctx, types.PoolWeights{
				StableWeight:       5,
				BalancerWeight:     2,
				ConcentratedWeight: 2,
			})

			routes, err := suite.App.ProtoRevKeeper.BuildHotRoutes(suite.Ctx, tc.swapIn, tc.swapOut, tc.poolId)

			if tc.hasRoutes {
				suite.Require().NoError(err)
				suite.Require().Equal(len(tc.expectedRoutes), len(routes))

				for routeIndex, routeMetaData := range routes {
					expectedHops := len(tc.expectedRoutes[routeIndex])
					suite.Require().Equal(expectedHops, len(routeMetaData.Route.PoolIds()))

					expectedStepSize := tc.expectedStepSize[routeIndex]
					suite.Require().Equal(expectedStepSize, routeMetaData.StepSize)

					expectedPoolPoints := tc.expectedRoutePoolPoints[routeIndex]
					suite.Require().Equal(expectedPoolPoints, routeMetaData.PoolPoints)

					expectedRoutes := tc.expectedRoutes[routeIndex]

					for tradeIndex, trade := range expectedRoutes {
						suite.Require().Equal(trade.PoolId, routeMetaData.Route.PoolIds()[tradeIndex])
					}
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestCalculateRoutePoolPoints tests the CalculateRoutePoolPoints function
func (suite *KeeperTestSuite) TestCalculateRoutePoolPoints() {
	cases := []struct {
		description             string
		route                   poolmanagertypes.SwapAmountInRoutes
		expectedRoutePoolPoints uint64
		expectedPass            bool
	}{
		{
			description:             "Valid route containing only balancer pools",
			route:                   []poolmanagertypes.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: ""}, {PoolId: 2, TokenOutDenom: ""}, {PoolId: 3, TokenOutDenom: ""}},
			expectedRoutePoolPoints: 6,
			expectedPass:            true,
		},
		{
			description:             "Valid route containing only balancer pools and equal number of pool points",
			route:                   []poolmanagertypes.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: ""}, {PoolId: 2, TokenOutDenom: ""}, {PoolId: 3, TokenOutDenom: ""}},
			expectedRoutePoolPoints: 6,
			expectedPass:            true,
		},
		{
			description:             "Valid route containing only stable swap pools",
			route:                   []poolmanagertypes.SwapAmountInRoute{{PoolId: 40, TokenOutDenom: ""}, {PoolId: 40, TokenOutDenom: ""}, {PoolId: 40, TokenOutDenom: ""}},
			expectedRoutePoolPoints: 9,
			expectedPass:            true,
		},
		{
			description:             "Valid route with more than 3 hops",
			route:                   []poolmanagertypes.SwapAmountInRoute{{PoolId: 40, TokenOutDenom: ""}, {PoolId: 40, TokenOutDenom: ""}, {PoolId: 40, TokenOutDenom: ""}, {PoolId: 1, TokenOutDenom: ""}},
			expectedRoutePoolPoints: 11,
			expectedPass:            true,
		},
		{
			description:             "Invalid route with more than 3 hops",
			route:                   []poolmanagertypes.SwapAmountInRoute{{PoolId: 4000, TokenOutDenom: ""}, {PoolId: 40, TokenOutDenom: ""}, {PoolId: 40, TokenOutDenom: ""}, {PoolId: 1, TokenOutDenom: ""}},
			expectedRoutePoolPoints: 11,
			expectedPass:            false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()
			suite.App.ProtoRevKeeper.SetPoolWeights(suite.Ctx, types.PoolWeights{StableWeight: 3, BalancerWeight: 2, ConcentratedWeight: 1})

			routePoolPoints, err := suite.App.ProtoRevKeeper.CalculateRoutePoolPoints(suite.Ctx, tc.route)
			if tc.expectedPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedRoutePoolPoints, routePoolPoints)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}
