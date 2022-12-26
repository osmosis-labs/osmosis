package math_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
)

func (suite *ConcentratedMathTestSuite) TestTickToSqrtPrice() {
	testCases := map[string]struct {
		tickIndex         sdk.Int
		sqrtPriceExpected string
	}{
		"positive tick index 1": {
			tickIndex:         sdk.NewInt(85176),
			sqrtPriceExpected: "70.710004849206351867", // 70.710004849206120647
			// https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B85176%2C2%5D%5D
		},
		"positive tick index 2": {
			tickIndex:         sdk.NewInt(86129),
			sqrtPriceExpected: "74.160724590951092256", // 74.160724590950847046
			// https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B86129%2C2%5D%5D
		},
		"negative tick index 1": {
			tickIndex:         sdk.NewInt(-85176),
			sqrtPriceExpected: "0.014142270278902748", // 0.0141422702789027917
			// https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B-85176%2C2%5D%5D
		},
		"negative tick index 2": {
			tickIndex:         sdk.NewInt(-86129),
			sqrtPriceExpected: "0.013484226394978060", // 0.0134842263949780882
			// https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B-86129%2C2%5D%5D
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			sqrtPrice, err := math.TickToSqrtPrice(tc.tickIndex)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.sqrtPriceExpected, sqrtPrice.String())

		})
	}
}

func (suite *ConcentratedMathTestSuite) TestPriceToTick() {
	testCases := []struct {
		name         string
		price        sdk.Dec
		tickExpected string
	}{
		{
			"happy path",
			sdk.NewDec(5000),
			"85176",
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			tick := math.PriceToTick(tc.price)
			suite.Require().Equal(tc.tickExpected, tick.String())
		})
	}
}
