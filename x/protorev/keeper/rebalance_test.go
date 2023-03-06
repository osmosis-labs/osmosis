package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v15/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// Mainnet Arb Route - 2 Asset, Same Weights (Block: 5905150)
// expectedAmtIn:  sdk.NewInt(10100000),
// expectedProfit: sdk.NewInt(24852)
var routeTwoAssetSameWeight = poolmanagertypes.SwapAmountInRoutes{
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        22,
		TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        23,
		TokenOutDenom: "ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        24,
		TokenOutDenom: "uosmo",
	}}

// Mainnet Arb Route - Multi Asset, Same Weights (Block: 6906570)
// expectedAmtIn:  sdk.NewInt(4800000),
// expectedProfit: sdk.NewInt(4547)
var routeMultiAssetSameWeight = poolmanagertypes.SwapAmountInRoutes{
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        26,
		TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        28,
		TokenOutDenom: "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        27,
		TokenOutDenom: "uosmo",
	}}

// Arb Route - Multi Asset, Same Weights - Pool 22 instead of 26 (Block: 6906570)
// expectedAmtIn:  sdk.NewInt(519700000),
// expectedProfit: sdk.NewInt(67511701)
var routeMostProfitable = poolmanagertypes.SwapAmountInRoutes{
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        22,
		TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        28,
		TokenOutDenom: "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        27,
		TokenOutDenom: "uosmo",
	}}

// Mainnet Arb Route - Multi Asset, Different Weights (Block: 6908256)
// expectedAmtIn:  sdk.NewInt(4100000),
// expectedProfit: sdk.NewInt(5826)
var routeDiffDenom = poolmanagertypes.SwapAmountInRoutes{
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        31,
		TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        32,
		TokenOutDenom: "ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        33,
		TokenOutDenom: "Atom",
	}}

// No Arbitrage Opportunity
// expectedAmtIn:  sdk.NewInt(0),
// expectedProfit: sdk.NewInt(0)
var routeNoArb = poolmanagertypes.SwapAmountInRoutes{
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        7,
		TokenOutDenom: "akash",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        12,
		TokenOutDenom: "juno",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        8,
		TokenOutDenom: "uosmo",
	}}

// StableSwap Test Route
// expectedAmtIn:  sdk.NewInt(137600000),
// expectedProfit: sdk.NewInt(56585438)
var routeStableSwap = poolmanagertypes.SwapAmountInRoutes{
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        29,
		TokenOutDenom: "usdc",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        40,
		TokenOutDenom: "busd",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        30,
		TokenOutDenom: "uosmo",
	}}

// Four Pool Test Route (Mainnet Block: 1855422)
// expectedAmtIn:  sdk.NewInt(1_147_000_000)
// expectedProfit: sdk.NewInt(15_761_405)
var fourPoolRoute = poolmanagertypes.SwapAmountInRoutes{
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        34,
		TokenOutDenom: "test/1",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        35,
		TokenOutDenom: types.OsmosisDenomination,
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        36,
		TokenOutDenom: "test/2",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        37,
		TokenOutDenom: "Atom",
	},
}

// Two Pool Test Route (Mainnet Block: 6_300_675)
// expectedAmtIn:  sdk.NewInt(989_000_000)
// expectedProfit: sdk.NewInt(218_149_058)
var twoPoolRoute = poolmanagertypes.SwapAmountInRoutes{
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        38,
		TokenOutDenom: types.OsmosisDenomination,
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        39,
		TokenOutDenom: "test/3",
	},
}

// Tests the binary search range extends to the correct amount
var extendedRangeRoute = poolmanagertypes.SwapAmountInRoutes{
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        42,
		TokenOutDenom: "usdy",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        43,
		TokenOutDenom: "usdx",
	},
}

// EstimateMultiHopSwap Panic catching test
var panicRoute = poolmanagertypes.SwapAmountInRoutes{
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        44,
		TokenOutDenom: "usdy",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        45,
		TokenOutDenom: "usdx",
	},
}

func (suite *KeeperTestSuite) TestFindMaxProfitRoute() {

	type param struct {
		route           poolmanagertypes.SwapAmountInRoutes
		expectedAmtIn   sdk.Int
		expectedProfit  sdk.Int
		routePoolPoints uint64
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "Mainnet Arb Route - 2 Asset, Same Weights (Block: 5905150)",
			param: param{
				route:           routeTwoAssetSameWeight,
				expectedAmtIn:   sdk.NewInt(10000000),
				expectedProfit:  sdk.NewInt(24848),
				routePoolPoints: 6,
			},
			expectPass: true,
		},
		{
			name: "Mainnet Arb Route - Multi Asset, Same Weights (Block: 6906570)",
			param: param{
				route:           routeMultiAssetSameWeight,
				expectedAmtIn:   sdk.NewInt(5000000),
				expectedProfit:  sdk.NewInt(4538),
				routePoolPoints: 6,
			},
			expectPass: true,
		},
		{
			name: "Arb Route - Multi Asset, Same Weights - Pool 22 instead of 26 (Block: 6906570)",
			param: param{
				route:           routeMostProfitable,
				expectedAmtIn:   sdk.NewInt(520000000),
				expectedProfit:  sdk.NewInt(67511675),
				routePoolPoints: 6,
			},
			expectPass: true,
		},
		{
			name: "Mainnet Arb Route - Multi Asset, Different Weights (Block: 6908256)",
			param: param{
				route:           routeDiffDenom,
				expectedAmtIn:   sdk.NewInt(4000000),
				expectedProfit:  sdk.NewInt(5826),
				routePoolPoints: 6,
			},
			expectPass: true,
		},
		{
			name: "StableSwap Test Route",
			param: param{
				route:           routeStableSwap,
				expectedAmtIn:   sdk.NewInt(138000000),
				expectedProfit:  sdk.NewInt(56585052),
				routePoolPoints: 9,
			},
			expectPass: true,
		},
		{
			name: "No Arbitrage Opportunity",
			param: param{
				route:           routeNoArb,
				expectedAmtIn:   sdk.Int{},
				expectedProfit:  sdk.NewInt(0),
				routePoolPoints: 0,
			},
			expectPass: true,
		},
		{
			name: "Four Pool Test Route",
			param: param{
				route:           fourPoolRoute,
				expectedAmtIn:   sdk.NewInt(1_147_000_000),
				expectedProfit:  sdk.NewInt(15_761_405),
				routePoolPoints: 8,
			},
			expectPass: true,
		},
		{
			name: "Two Pool Test Route",
			param: param{
				route:           twoPoolRoute,
				expectedAmtIn:   sdk.NewInt(989_000_000),
				expectedProfit:  sdk.NewInt(218_149_058),
				routePoolPoints: 4,
			},
			expectPass: true,
		},
		{
			name: "Extended Range Test Route",
			param: param{
				route:           extendedRangeRoute,
				expectedAmtIn:   sdk.NewInt(131_072_000_000),
				expectedProfit:  sdk.NewInt(20_900_656_975),
				routePoolPoints: 10,
			},
			expectPass: true,
		},
		{
			name: "Panic Route",
			param: param{
				route:           panicRoute,
				expectedAmtIn:   sdk.NewInt(0),
				expectedProfit:  sdk.NewInt(0),
				routePoolPoints: 0,
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			// init the route
			remainingPoolPoints := uint64(1000)
			route := protorevtypes.RouteMetaData{
				Route:      test.param.route,
				PoolPoints: test.param.routePoolPoints,
				StepSize:   sdk.NewInt(1_000_000),
			}

			amtIn, profit, err := suite.App.ProtoRevKeeper.FindMaxProfitForRoute(
				suite.Ctx,
				route,
				&remainingPoolPoints,
			)

			if test.expectPass {
				suite.Require().NoError(err)
				suite.Require().Equal(test.param.expectedAmtIn, amtIn.Amount)
				suite.Require().Equal(test.param.expectedProfit, profit)
			} else {
				suite.Require().Error(err)
			}

			// check that the remaining pool points is correct
			suite.Require().Equal(uint64(1000), remainingPoolPoints+test.param.routePoolPoints)
		})
	}
}

func (suite *KeeperTestSuite) TestExecuteTrade() {

	type param struct {
		route          poolmanagertypes.SwapAmountInRoutes
		inputCoin      sdk.Coin
		expectedProfit sdk.Int
	}

	tests := []struct {
		name                string
		param               param
		arbDenom            string
		expectPass          bool
		expectedNumOfTrades sdk.Int
	}{
		{
			name: "Mainnet Arb Route",
			param: param{
				route:          routeTwoAssetSameWeight,
				inputCoin:      sdk.NewCoin("uosmo", sdk.NewInt(10100000)),
				expectedProfit: sdk.NewInt(24852),
			},
			arbDenom:            types.OsmosisDenomination,
			expectPass:          true,
			expectedNumOfTrades: sdk.NewInt(1),
		},
		{
			name: "No arbitrage opportunity - expect error at multihopswap due to profitability invariant",
			param: param{
				route:          routeNoArb,
				inputCoin:      sdk.NewCoin("uosmo", sdk.NewInt(1000000)),
				expectedProfit: sdk.NewInt(0),
			},
			arbDenom:   types.OsmosisDenomination,
			expectPass: false,
		},
		{
			name: "0 input amount - expect error at multihopswap due to amount needing to be positive",
			param: param{
				route:          routeNoArb,
				inputCoin:      sdk.NewCoin("uosmo", sdk.NewInt(0)),
				expectedProfit: sdk.NewInt(0),
			},
			arbDenom:   types.OsmosisDenomination,
			expectPass: false,
		},
		{
			name: "4-Pool Route Arb",
			param: param{
				route:          fourPoolRoute,
				inputCoin:      sdk.NewCoin("Atom", sdk.NewInt(1_147_000_000)),
				expectedProfit: sdk.NewInt(15_761_405),
			},
			arbDenom:            "Atom",
			expectPass:          true,
			expectedNumOfTrades: sdk.NewInt(2),
		},
		{
			name: "2-Pool Route Arb",
			param: param{
				route:          twoPoolRoute,
				inputCoin:      sdk.NewCoin("test/3", sdk.NewInt(989_000_000)),
				expectedProfit: sdk.NewInt(218_149_058),
			},
			arbDenom:            "test/3",
			expectPass:          true,
			expectedNumOfTrades: sdk.NewInt(3),
		},
	}

	for _, test := range tests {

		err := suite.App.ProtoRevKeeper.ExecuteTrade(
			suite.Ctx,
			test.param.route,
			test.param.inputCoin,
		)

		if test.expectPass {
			suite.Require().NoError(err)

			// Check the protorev statistics
			numberOfTrades, err := suite.App.ProtoRevKeeper.GetTradesByRoute(suite.Ctx, test.param.route.PoolIds())
			suite.Require().NoError(err)
			suite.Require().Equal(sdk.OneInt(), numberOfTrades)

			routeProfit, err := suite.App.ProtoRevKeeper.GetProfitsByRoute(suite.Ctx, test.param.route.PoolIds(), test.arbDenom)
			suite.Require().NoError(err)
			suite.Require().Equal(test.param.expectedProfit, routeProfit.Amount)

			profit, err := suite.App.ProtoRevKeeper.GetProfitsByDenom(suite.Ctx, test.arbDenom)
			suite.Require().NoError(err)
			suite.Require().Equal(test.param.expectedProfit, profit.Amount)

			totalNumberOfTrades, err := suite.App.ProtoRevKeeper.GetNumberOfTrades(suite.Ctx)
			suite.Require().NoError(err)
			suite.Require().Equal(test.expectedNumOfTrades, totalNumberOfTrades)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestIterateRoutes() {
	type paramm struct {
		routes                     []poolmanagertypes.SwapAmountInRoutes
		expectedMaxProfitAmount    sdk.Int
		expectedMaxProfitInputCoin sdk.Coin
		expectedOptimalRoute       poolmanagertypes.SwapAmountInRoutes

		arbDenom string
	}

	tests := []struct {
		name       string
		params     paramm
		expectPass bool
	}{
		{name: "Single Route Test",
			params: paramm{
				routes:                     []poolmanagertypes.SwapAmountInRoutes{routeTwoAssetSameWeight},
				expectedMaxProfitAmount:    sdk.NewInt(24848),
				expectedMaxProfitInputCoin: sdk.NewCoin("uosmo", sdk.NewInt(10000000)),
				expectedOptimalRoute:       routeTwoAssetSameWeight,
				arbDenom:                   types.OsmosisDenomination,
			},
			expectPass: true,
		},
		{name: "Two routes with same arb denom test - more profitable route second",
			params: paramm{
				routes:                     []poolmanagertypes.SwapAmountInRoutes{routeMultiAssetSameWeight, routeTwoAssetSameWeight},
				expectedMaxProfitAmount:    sdk.NewInt(24848),
				expectedMaxProfitInputCoin: sdk.NewCoin("uosmo", sdk.NewInt(10000000)),
				expectedOptimalRoute:       routeTwoAssetSameWeight,
				arbDenom:                   types.OsmosisDenomination,
			},
			expectPass: true,
		},
		{name: "Three routes with same arb denom test - most profitable route first",
			params: paramm{
				routes:                     []poolmanagertypes.SwapAmountInRoutes{routeMostProfitable, routeMultiAssetSameWeight, routeTwoAssetSameWeight},
				expectedMaxProfitAmount:    sdk.NewInt(67511675),
				expectedMaxProfitInputCoin: sdk.NewCoin("uosmo", sdk.NewInt(520000000)),
				expectedOptimalRoute:       routeMostProfitable,
				arbDenom:                   types.OsmosisDenomination,
			},
			expectPass: true,
		},
		{name: "Two routes, different arb denoms test - more profitable route second",
			params: paramm{
				routes:                     []poolmanagertypes.SwapAmountInRoutes{routeNoArb, routeDiffDenom},
				expectedMaxProfitAmount:    sdk.NewInt(4880),
				expectedMaxProfitInputCoin: sdk.NewCoin("Atom", sdk.NewInt(4000000)),
				expectedOptimalRoute:       routeDiffDenom,
				arbDenom:                   "Atom",
			},
			expectPass: true,
		},
		{name: "Four-pool route test",
			params: paramm{
				routes:                     []poolmanagertypes.SwapAmountInRoutes{fourPoolRoute},
				expectedMaxProfitAmount:    sdk.NewInt(13_202_729),
				expectedMaxProfitInputCoin: sdk.NewCoin("Atom", sdk.NewInt(1_147_000_000)),
				expectedOptimalRoute:       fourPoolRoute,
				arbDenom:                   "Atom",
			},
			expectPass: true,
		},
		{name: "Two-pool route test",
			params: paramm{
				routes:                     []poolmanagertypes.SwapAmountInRoutes{twoPoolRoute},
				expectedMaxProfitAmount:    sdk.NewInt(198_653_535),
				expectedMaxProfitInputCoin: sdk.NewCoin("test/3", sdk.NewInt(989_000_000)),
				expectedOptimalRoute:       twoPoolRoute,
				arbDenom:                   "test/3",
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			routes := make([]protorevtypes.RouteMetaData, len(test.params.routes))
			for i, route := range test.params.routes {
				routes[i] = protorevtypes.RouteMetaData{
					Route:      route,
					PoolPoints: 0,
					StepSize:   sdk.NewInt(1_000_000),
				}
			}
			// Set a high default pool points so that all routes are considered
			remainingPoolPoints := uint64(40)

			maxProfitInputCoin, maxProfitAmount, optimalRoute := suite.App.ProtoRevKeeper.IterateRoutes(suite.Ctx, routes, &remainingPoolPoints)
			if test.expectPass {
				suite.Require().Equal(test.params.expectedMaxProfitAmount, maxProfitAmount)
				suite.Require().Equal(test.params.expectedMaxProfitInputCoin, maxProfitInputCoin)
				suite.Require().Equal(test.params.expectedOptimalRoute, optimalRoute)
			}
		})
	}
}

// Test logic that compares proftability of routes with different assets
func (suite *KeeperTestSuite) TestConvertProfits() {
	type param struct {
		inputCoin           sdk.Coin
		profit              sdk.Int
		expectedUosmoProfit sdk.Int
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{name: "Convert atom to uosmo",
			param: param{
				inputCoin:           sdk.NewCoin("Atom", sdk.NewInt(100)),
				profit:              sdk.NewInt(10),
				expectedUosmoProfit: sdk.NewInt(8),
			},
			expectPass: true,
		},
		{name: "Convert juno to uosmo (random denom)",
			param: param{
				inputCoin:           sdk.NewCoin("juno", sdk.NewInt(100)),
				profit:              sdk.NewInt(10),
				expectedUosmoProfit: sdk.NewInt(9),
			},
			expectPass: true,
		},
		{name: "Convert denom without pool to uosmo",
			param: param{
				inputCoin:           sdk.NewCoin("random", sdk.NewInt(100)),
				profit:              sdk.NewInt(10),
				expectedUosmoProfit: sdk.NewInt(10),
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		profit, err := suite.App.ProtoRevKeeper.ConvertProfits(suite.Ctx, test.param.inputCoin, test.param.profit)

		if test.expectPass {
			suite.Require().NoError(err)
			suite.Require().Equal(test.param.expectedUosmoProfit, profit)
		} else {
			suite.Require().Error(err)
		}
	}
}

// TestRemainingPoolPointsForTx tests the RemainingPoolPointsForTx function.
func (suite *KeeperTestSuite) TestRemainingPoolPointsForTx() {
	cases := []struct {
		description        string
		maxRoutesPerTx     uint64
		maxRoutesPerBlock  uint64
		currentRouteCount  uint64
		expectedPointCount uint64
	}{
		{
			description:        "Max pool points per tx is 10 and max pool points per block is 100",
			maxRoutesPerTx:     10,
			maxRoutesPerBlock:  100,
			currentRouteCount:  0,
			expectedPointCount: 10,
		},
		{
			description:        "Max pool points per tx is 10, max pool points per block is 100, and current point count is 90",
			maxRoutesPerTx:     10,
			maxRoutesPerBlock:  100,
			currentRouteCount:  90,
			expectedPointCount: 10,
		},
		{
			description:        "Max pool points per tx is 10, max pool points per block is 100, and current point count is 100",
			maxRoutesPerTx:     10,
			maxRoutesPerBlock:  100,
			currentRouteCount:  100,
			expectedPointCount: 0,
		},
		{
			description:        "Max pool points per tx is 10, max pool points per block is 100, and current point count is 95",
			maxRoutesPerTx:     10,
			maxRoutesPerBlock:  100,
			currentRouteCount:  95,
			expectedPointCount: 5,
		},
		{
			description:        "Checking overflow",
			maxRoutesPerTx:     10,
			maxRoutesPerBlock:  100,
			currentRouteCount:  105,
			expectedPointCount: 0,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()

			suite.App.ProtoRevKeeper.SetMaxPointsPerTx(suite.Ctx, tc.maxRoutesPerTx)
			suite.App.ProtoRevKeeper.SetMaxPointsPerBlock(suite.Ctx, tc.maxRoutesPerBlock)
			suite.App.ProtoRevKeeper.SetPointCountForBlock(suite.Ctx, tc.currentRouteCount)

			points, err := suite.App.ProtoRevKeeper.RemainingPoolPointsForTx(suite.Ctx)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedPointCount, points)
		})
	}
}
