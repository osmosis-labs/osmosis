package math_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

func (suite *ConcentratedMathTestSuite) TestTickToPrice() {
	testCases := map[string]struct {
		tickIndex     sdk.Int
		kAtPriceOne   sdk.Int
		expectedPrice string
		expectedError error
	}{
		"One dollar increments at the ten thousands place: 1": {
			tickIndex:     sdk.NewInt(400000),
			kAtPriceOne:   sdk.NewInt(-4),
			expectedPrice: "50000.000000000000000000",
		},
		"One dollar increments at the ten thousands place: 2": {
			tickIndex:     sdk.NewInt(400001),
			kAtPriceOne:   sdk.NewInt(-4),
			expectedPrice: "50001.000000000000000000",
		},
		"Ten cent increments at the ten thousands place: 1": {
			tickIndex:     sdk.NewInt(4000000),
			kAtPriceOne:   sdk.NewInt(-5),
			expectedPrice: "50000.000000000000000000",
		},
		"Ten cent increments at the ten thousands place: 2": {
			tickIndex:     sdk.NewInt(4000001),
			kAtPriceOne:   sdk.NewInt(-5),
			expectedPrice: "50000.100000000000000000",
		},
		"One cent increments at the ten thousands place: 1": {
			tickIndex:     sdk.NewInt(40000000),
			kAtPriceOne:   sdk.NewInt(-6),
			expectedPrice: "50000.000000000000000000",
		},
		"One cent increments at the ten thousands place: 2": {
			tickIndex:     sdk.NewInt(40000001),
			kAtPriceOne:   sdk.NewInt(-6),
			expectedPrice: "50000.010000000000000000",
		},
		"One cent increments at the ones place: 1": {
			tickIndex:     sdk.NewInt(400),
			kAtPriceOne:   sdk.NewInt(-2),
			expectedPrice: "5.000000000000000000",
		},
		"One cent increments at the ones place: 2": {
			tickIndex:     sdk.NewInt(401),
			kAtPriceOne:   sdk.NewInt(-2),
			expectedPrice: "5.010000000000000000",
		},
		"Ten cent increments at the tens place: 1": {
			tickIndex:     sdk.NewInt(1300),
			kAtPriceOne:   sdk.NewInt(-2),
			expectedPrice: "50.000000000000000000",
		},
		"Ten cent increments at the ones place: 2": {
			tickIndex:     sdk.NewInt(1301),
			kAtPriceOne:   sdk.NewInt(-2),
			expectedPrice: "50.100000000000000000",
		},
		"One cent increments at the tenths place: 1": {
			tickIndex:     sdk.NewInt(-2),
			kAtPriceOne:   sdk.NewInt(-2),
			expectedPrice: "0.980000000000000000",
		},
		"One tenth of a cent increments at the hundredths place: 1": {
			tickIndex:     sdk.NewInt(-2),
			kAtPriceOne:   sdk.NewInt(-3),
			expectedPrice: "0.998000000000000000",
		},
		"One hundredths of a cent increments at the thousandths place: 1": {
			tickIndex:     sdk.NewInt(-2),
			kAtPriceOne:   sdk.NewInt(-4),
			expectedPrice: "0.999800000000000000",
		},
		"One ten millionth of a cent increments at the hundred millionths place: 1": {
			tickIndex:     sdk.NewInt(-2),
			kAtPriceOne:   sdk.NewInt(-8),
			expectedPrice: "0.999999980000000000",
		},
		"One ten millionth of a cent increments at the hundred millionths place: 2": {
			tickIndex:     sdk.NewInt(-99999111),
			kAtPriceOne:   sdk.NewInt(-8),
			expectedPrice: "0.090000889000000000",
		},
		"More variety of numbers in each place": {
			tickIndex:     sdk.NewInt(4030301),
			kAtPriceOne:   sdk.NewInt(-5),
			expectedPrice: "53030.100000000000000000",
		},
		"Max tick and min k": {
			tickIndex:     sdk.NewInt(1000),
			kAtPriceOne:   sdk.NewInt(-1),
			expectedPrice: "200000000000.000000000000000000",
		},
		"error: tickIndex less than minimum": {
			tickIndex:     sdk.NewInt(-163),
			kAtPriceOne:   sdk.NewInt(-1),
			expectedError: fmt.Errorf("tickIndex must be greater than or equal to %s", "-162"),
		},
		"error: tickIndex greater than maximum": {
			tickIndex:     sdk.NewInt(1001),
			kAtPriceOne:   sdk.NewInt(-1),
			expectedError: fmt.Errorf("tickIndex must be less than or equal to %s", "1000"),
		},
		"error: kAtPriceOne less than minimum": {
			tickIndex:     sdk.NewInt(100),
			kAtPriceOne:   types.PrecisionValueAtPriceOneMin.Sub(sdk.OneInt()),
			expectedError: fmt.Errorf("kAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax),
		},
		"error: kAtPriceOne greater than maximum": {
			tickIndex:     sdk.NewInt(100),
			kAtPriceOne:   types.PrecisionValueAtPriceOneMax.Add(sdk.OneInt()),
			expectedError: fmt.Errorf("kAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax),
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			sqrtPrice, err := math.TickToPrice(tc.tickIndex, tc.kAtPriceOne)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError.Error(), err.Error())
				return
			}
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedPrice, sqrtPrice.String())

		})
	}
}

func (suite *ConcentratedMathTestSuite) TestPriceToTick() {
	testCases := map[string]struct {
		price         sdk.Dec
		kAtPriceOne   sdk.Int
		tickExpected  string
		expectedError error
	}{
		"50000 to tick with -4 k at price one": {
			price:        sdk.NewDec(50000),
			kAtPriceOne:  sdk.NewInt(-4),
			tickExpected: "400000",
		},
		"5.01 to tick with -2 k at price one": {
			price:        sdk.MustNewDecFromStr("5.010000000000000000"),
			kAtPriceOne:  sdk.NewInt(-2),
			tickExpected: "401",
		},
		"50000.01 to tick with -6 k at price one": {
			price:        sdk.MustNewDecFromStr("50000.010000000000000000"),
			kAtPriceOne:  sdk.NewInt(-6),
			tickExpected: "40000001",
		},
		"0.090000889 to tick with -8 k at price one": {
			price:        sdk.MustNewDecFromStr("0.090000889000000000"),
			kAtPriceOne:  sdk.NewInt(-8),
			tickExpected: "-99999111",
		},
		"0.9998 to tick with -4 k at price one": {
			price:        sdk.MustNewDecFromStr("0.999800000000000000"),
			kAtPriceOne:  sdk.NewInt(-4),
			tickExpected: "-2",
		},
		"53030.10 to tick with -5 k at price one": {
			price:        sdk.MustNewDecFromStr("53030.100000000000000000"),
			kAtPriceOne:  sdk.NewInt(-5),
			tickExpected: "4030301",
		},
		"max spot price and minimum kAtPriceOne": {
			price:        sdk.MustNewDecFromStr("200000000000"),
			kAtPriceOne:  sdk.NewInt(-1),
			tickExpected: "1000",
		},
		"error: price must be positive": {
			price:         sdk.NewDec(-1),
			kAtPriceOne:   sdk.NewInt(-6),
			expectedError: fmt.Errorf("price must be greater than zero"),
		},
		"error: resulting tickIndex too large": {
			price:         sdk.NewDec(200000000001),
			kAtPriceOne:   sdk.NewInt(-6),
			expectedError: fmt.Errorf("tickIndex must be less than or equal to %s", "100000000"),
		},
		"error: kAtPriceOne less than minimum": {
			price:         sdk.NewDec(50000),
			kAtPriceOne:   types.PrecisionValueAtPriceOneMin.Sub(sdk.OneInt()),
			expectedError: fmt.Errorf("kAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax),
		},
		"error: kAtPriceOne greater than maximum": {
			price:         sdk.NewDec(50000),
			kAtPriceOne:   types.PrecisionValueAtPriceOneMax.Add(sdk.OneInt()),
			expectedError: fmt.Errorf("kAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax),
		},
	}
	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			tick, err := math.PriceToTick(tc.price, tc.kAtPriceOne)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError.Error(), err.Error())
				return
			}
			suite.Require().NoError(err)
			suite.Require().Equal(tc.tickExpected, tick.String())
		})
	}
}
