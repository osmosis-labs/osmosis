package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
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
		TokenOutDenom: types.AtomDenomination,
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
		PoolId:        34,
		TokenOutDenom: "busd",
	},
	poolmanagertypes.SwapAmountInRoute{
		PoolId:        30,
		TokenOutDenom: "uosmo",
	}}

func (suite *KeeperTestSuite) TestFindMaxProfitRoute() {

	type param struct {
		route          poolmanagertypes.SwapAmountInRoutes
		expectedAmtIn  sdk.Int
		expectedProfit sdk.Int
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{name: "Mainnet Arb Route - 2 Asset, Same Weights (Block: 5905150)",
			param: param{
				route:          routeTwoAssetSameWeight,
				expectedAmtIn:  sdk.NewInt(10000000),
				expectedProfit: sdk.NewInt(24848)},
			expectPass: true},
		{name: "Mainnet Arb Route - Multi Asset, Same Weights (Block: 6906570)",
			param: param{
				route:          routeMultiAssetSameWeight,
				expectedAmtIn:  sdk.NewInt(5000000),
				expectedProfit: sdk.NewInt(4538)},
			expectPass: true},
		{name: "Arb Route - Multi Asset, Same Weights - Pool 22 instead of 26 (Block: 6906570)",
			param: param{
				route:          routeMostProfitable,
				expectedAmtIn:  sdk.NewInt(520000000),
				expectedProfit: sdk.NewInt(67511675)},
			expectPass: true},
		{name: "Mainnet Arb Route - Multi Asset, Different Weights (Block: 6908256)",
			param: param{
				route:          routeDiffDenom,
				expectedAmtIn:  sdk.NewInt(4000000),
				expectedProfit: sdk.NewInt(5826)},
			expectPass: true},
		{name: "StableSwap Test Route",
			param: param{
				route:          routeStableSwap,
				expectedAmtIn:  sdk.NewInt(138000000),
				expectedProfit: sdk.NewInt(56585052)},
			expectPass: true},
		{name: "No Arbitrage Opportunity",
			param: param{
				route:          routeNoArb,
				expectedAmtIn:  sdk.Int{},
				expectedProfit: sdk.NewInt(0)},
			expectPass: true},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {

			amtIn, profit, err := suite.App.ProtoRevKeeper.FindMaxProfitForRoute(
				suite.Ctx,
				test.param.route,
				test.param.route[2].TokenOutDenom,
			)

			if test.expectPass {
				suite.Require().NoError(err)
				suite.Require().Equal(test.param.expectedAmtIn, amtIn.Amount)
				suite.Require().Equal(test.param.expectedProfit, profit)
			} else {
				suite.Require().Error(err)
			}
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
		name       string
		param      param
		arbDenom   string
		expectPass bool
	}{
		{
			name: "Mainnet Arb Route",
			param: param{
				route:          routeTwoAssetSameWeight,
				inputCoin:      sdk.NewCoin("uosmo", sdk.NewInt(10100000)),
				expectedProfit: sdk.NewInt(24852),
			},
			arbDenom:   types.OsmosisDenomination,
			expectPass: true,
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
			suite.Require().Equal(sdk.OneInt(), totalNumberOfTrades)
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
				expectedMaxProfitInputCoin: sdk.NewCoin(types.AtomDenomination, sdk.NewInt(4000000)),
				expectedOptimalRoute:       routeDiffDenom,
				arbDenom:                   types.AtomDenomination,
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		maxProfitInputCoin, maxProfitAmount, optimalRoute := suite.App.ProtoRevKeeper.IterateRoutes(suite.Ctx, test.params.routes)

		if test.expectPass {
			suite.Require().Equal(test.params.expectedMaxProfitAmount, maxProfitAmount)
			suite.Require().Equal(test.params.expectedMaxProfitInputCoin, maxProfitInputCoin)
			suite.Require().Equal(test.params.expectedOptimalRoute, optimalRoute)
		}
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
				inputCoin:           sdk.NewCoin(types.AtomDenomination, sdk.NewInt(100)),
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
