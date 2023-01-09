package math_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

func (suite *ConcentratedMathTestSuite) TestTickToPrice() {
	testCases := map[string]struct {
		tickIndex          sdk.Int
		exponentAtPriceOne sdk.Int
		expectedPrice      string
		expectedError      error
	}{
		"One dollar increments at the ten thousands place: 1": {
			tickIndex:          sdk.NewInt(400000),
			exponentAtPriceOne: sdk.NewInt(-4),
			expectedPrice:      "50000.000000000000000000",
		},
		"One dollar increments at the ten thousands place: 2": {
			tickIndex:          sdk.NewInt(400001),
			exponentAtPriceOne: sdk.NewInt(-4),
			expectedPrice:      "50001.000000000000000000",
		},
		"Ten cent increments at the ten thousands place: 1": {
			tickIndex:          sdk.NewInt(4000000),
			exponentAtPriceOne: sdk.NewInt(-5),
			expectedPrice:      "50000.000000000000000000",
		},
		"Ten cent increments at the ten thousands place: 2": {
			tickIndex:          sdk.NewInt(4000001),
			exponentAtPriceOne: sdk.NewInt(-5),
			expectedPrice:      "50000.100000000000000000",
		},
		"One cent increments at the ten thousands place: 1": {
			tickIndex:          sdk.NewInt(40000000),
			exponentAtPriceOne: sdk.NewInt(-6),
			expectedPrice:      "50000.000000000000000000",
		},
		"One cent increments at the ten thousands place: 2": {
			tickIndex:          sdk.NewInt(40000001),
			exponentAtPriceOne: sdk.NewInt(-6),
			expectedPrice:      "50000.010000000000000000",
		},
		"One cent increments at the ones place: 1": {
			tickIndex:          sdk.NewInt(400),
			exponentAtPriceOne: sdk.NewInt(-2),
			expectedPrice:      "5.000000000000000000",
		},
		"One cent increments at the ones place: 2": {
			tickIndex:          sdk.NewInt(401),
			exponentAtPriceOne: sdk.NewInt(-2),
			expectedPrice:      "5.010000000000000000",
		},
		"Ten cent increments at the tens place: 1": {
			tickIndex:          sdk.NewInt(1300),
			exponentAtPriceOne: sdk.NewInt(-2),
			expectedPrice:      "50.000000000000000000",
		},
		"Ten cent increments at the ones place: 2": {
			tickIndex:          sdk.NewInt(1301),
			exponentAtPriceOne: sdk.NewInt(-2),
			expectedPrice:      "50.100000000000000000",
		},
		"One cent increments at the tenths place: 1": {
			tickIndex:          sdk.NewInt(-2),
			exponentAtPriceOne: sdk.NewInt(-2),
			expectedPrice:      "0.980000000000000000",
		},
		"One tenth of a cent increments at the hundredths place: 1": {
			tickIndex:          sdk.NewInt(-2),
			exponentAtPriceOne: sdk.NewInt(-3),
			expectedPrice:      "0.998000000000000000",
		},
		"One hundredths of a cent increments at the thousandths place: 1": {
			tickIndex:          sdk.NewInt(-2),
			exponentAtPriceOne: sdk.NewInt(-4),
			expectedPrice:      "0.999800000000000000",
		},
		"One ten millionth of a cent increments at the hundred millionths place: 1": {
			tickIndex:          sdk.NewInt(-2),
			exponentAtPriceOne: sdk.NewInt(-8),
			expectedPrice:      "0.999999980000000000",
		},
		"One ten millionth of a cent increments at the hundred millionths place: 2": {
			tickIndex:          sdk.NewInt(-99999111),
			exponentAtPriceOne: sdk.NewInt(-8),
			expectedPrice:      "0.090000889000000000",
		},
		"More variety of numbers in each place": {
			tickIndex:          sdk.NewInt(4030301),
			exponentAtPriceOne: sdk.NewInt(-5),
			expectedPrice:      "53030.100000000000000000",
		},
		"Max tick and min k": {
			tickIndex:          sdk.NewInt(1000),
			exponentAtPriceOne: sdk.NewInt(-1),
			expectedPrice:      "200000000000.000000000000000000",
		},
		"error: tickIndex less than minimum": {
			tickIndex:          sdk.NewInt(DefaultMinTick - 1),
			exponentAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:      fmt.Errorf("tickIndex must be greater than or equal to %d", DefaultMinTick),
		},
		"error: tickIndex greater than maximum": {
			tickIndex:          sdk.NewInt(DefaultMaxTick + 1),
			exponentAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:      fmt.Errorf("tickIndex must be less than or equal to %d", DefaultMaxTick),
		},
		"error: exponentAtPriceOne less than minimum": {
			tickIndex:          sdk.NewInt(100),
			exponentAtPriceOne: types.PrecisionValueAtPriceOneMin.Sub(sdk.OneInt()),
			expectedError:      fmt.Errorf("exponentAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax),
		},
		"error: exponentAtPriceOne greater than maximum": {
			tickIndex:          sdk.NewInt(100),
			exponentAtPriceOne: types.PrecisionValueAtPriceOneMax.Add(sdk.OneInt()),
			expectedError:      fmt.Errorf("exponentAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax),
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			sqrtPrice, err := math.TickToPrice(tc.tickIndex, tc.exponentAtPriceOne)
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
		price              sdk.Dec
		exponentAtPriceOne sdk.Int
		tickExpected       string
		expectedError      error
	}{
		"50000 to tick with -4 k at price one": {
			price:              sdk.NewDec(50000),
			exponentAtPriceOne: sdk.NewInt(-4),
			tickExpected:       "400000",
		},
		"5.01 to tick with -2 k at price one": {
			price:              sdk.MustNewDecFromStr("5.010000000000000000"),
			exponentAtPriceOne: sdk.NewInt(-2),
			tickExpected:       "401",
		},
		"50000.01 to tick with -6 k at price one": {
			price:              sdk.MustNewDecFromStr("50000.010000000000000000"),
			exponentAtPriceOne: sdk.NewInt(-6),
			tickExpected:       "40000001",
		},
		"0.090000889 to tick with -8 k at price one": {
			price:              sdk.MustNewDecFromStr("0.090000889000000000"),
			exponentAtPriceOne: sdk.NewInt(-8),
			tickExpected:       "-99999111",
		},
		"0.9998 to tick with -4 k at price one": {
			price:              sdk.MustNewDecFromStr("0.999800000000000000"),
			exponentAtPriceOne: sdk.NewInt(-4),
			tickExpected:       "-2",
		},
		"53030.10 to tick with -5 k at price one": {
			price:              sdk.MustNewDecFromStr("53030.100000000000000000"),
			exponentAtPriceOne: sdk.NewInt(-5),
			tickExpected:       "4030301",
		},
		"max spot price and minimum exponentAtPriceOne": {
			price:              MaxSpotPrice,
			exponentAtPriceOne: sdk.NewInt(-1),
			tickExpected:       "3420",
		},
		"error: price must be positive": {
			price:              sdk.NewDec(-1),
			exponentAtPriceOne: sdk.NewInt(-6),
			expectedError:      fmt.Errorf("price must be greater than zero"),
		},
		"error: resulting tickIndex too large": {
			price:              MaxSpotPrice.Mul(sdk.NewDec(2)),
			exponentAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:      fmt.Errorf("tickIndex must be less than or equal to %d", DefaultMaxTick),
		},
		"error: exponentAtPriceOne less than minimum": {
			price:              sdk.NewDec(50000),
			exponentAtPriceOne: types.PrecisionValueAtPriceOneMin.Sub(sdk.OneInt()),
			expectedError:      fmt.Errorf("exponentAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax),
		},
		"error: exponentAtPriceOne greater than maximum": {
			price:              sdk.NewDec(50000),
			exponentAtPriceOne: types.PrecisionValueAtPriceOneMax.Add(sdk.OneInt()),
			expectedError:      fmt.Errorf("exponentAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax),
		},
	}
	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			tick, err := math.PriceToTick(tc.price, tc.exponentAtPriceOne)
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
