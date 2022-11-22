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
		"happy path 1": {
			tickIndex:         sdk.NewInt(85176),
			sqrtPriceExpected: "70.710004849206351867", // 70.710004849206120647
			// https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B85176%2C2%5D%5D
		},
		"happy path 2": {
			tickIndex:         sdk.NewInt(86129),
			sqrtPriceExpected: "74.160724590951092256", // 74.160724590950847046
			// https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B86129%2C2%5D%5D
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
