package math_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// use following equations to test testing vectors using sage
// geometricExponentIncrementDistanceInTicks(exponentAtPriceOne) = (9 * (10^(-exponentAtPriceOne)))
// geometricExponentDelta(tickIndex, exponentAtPriceOne)  = floor(tickIndex / geometricExponentIncrementDistanceInTicks(exponentAtPriceOne))
// exponentAtCurrentTick(tickIndex, exponentAtPriceOne) = exponentAtPriceOne + geometricExponentDelta(tickIndex, exponentAtPriceOne)
// currentAdditiveIncrementInTicks(tickIndex, exponentAtPriceOne) = pow(10, exponentAtCurrentTick(tickIndex, exponentAtPriceOne))
// numAdditiveTicks(tickIndex, exponentAtPriceOne) = tickIndex - (geometricExponentDelta(tickIndex, exponentAtPriceOne) * geometricExponentIncrementDistanceInTicks(exponentAtPriceOne)
// price(tickIndex, exponentAtPriceOne) = pow(10, geometricExponentDelta(tickIndex, exponentAtPriceOne)) +
// (numAdditiveTicks(tickIndex, exponentAtPriceOne) * currentAdditiveIncrementInTicks(tickIndex, exponentAtPriceOne))
func (suite *ConcentratedMathTestSuite) TestTickToSqrtPrice() {
	testCases := map[string]struct {
		tickIndex          sdk.Int
		exponentAtPriceOne sdk.Int
		expectedPrice      sdk.Dec
		expectedError      error
	}{
		"One dollar increments at the ten thousands place: 1": {
			tickIndex:          sdk.NewInt(400000),
			exponentAtPriceOne: sdk.NewInt(-4),
			expectedPrice:      sdk.NewDec(50000),
		},
		"One dollar increments at the ten thousands place: 2": {
			tickIndex:          sdk.NewInt(400001),
			exponentAtPriceOne: sdk.NewInt(-4),
			expectedPrice:      sdk.NewDec(50001),
		},
		"Ten cent increments at the ten thousands place: 1": {
			tickIndex:          sdk.NewInt(4000000),
			exponentAtPriceOne: sdk.NewInt(-5),
			expectedPrice:      sdk.NewDec(50000),
		},
		"Ten cent increments at the ten thousands place: 2": {
			tickIndex:          sdk.NewInt(4000001),
			exponentAtPriceOne: sdk.NewInt(-5),
			expectedPrice:      sdk.MustNewDecFromStr("50000.1"),
		},
		"One cent increments at the ten thousands place: 1": {
			tickIndex:          sdk.NewInt(40000000),
			exponentAtPriceOne: sdk.NewInt(-6),
			expectedPrice:      sdk.NewDec(50000),
		},
		"One cent increments at the ten thousands place: 2": {
			tickIndex:          sdk.NewInt(40000001),
			exponentAtPriceOne: sdk.NewInt(-6),
			expectedPrice:      sdk.MustNewDecFromStr("50000.01"),
		},
		"One cent increments at the ones place: 1": {
			tickIndex:          sdk.NewInt(400),
			exponentAtPriceOne: sdk.NewInt(-2),
			expectedPrice:      sdk.NewDec(5),
		},
		"One cent increments at the ones place: 2": {
			tickIndex:          sdk.NewInt(401),
			exponentAtPriceOne: sdk.NewInt(-2),
			expectedPrice:      sdk.MustNewDecFromStr("5.01"),
		},
		"Ten cent increments at the tens place: 1": {
			tickIndex:          sdk.NewInt(1300),
			exponentAtPriceOne: sdk.NewInt(-2),
			expectedPrice:      sdk.NewDec(50),
		},
		"Ten cent increments at the ones place: 2": {
			tickIndex:          sdk.NewInt(1301),
			exponentAtPriceOne: sdk.NewInt(-2),
			expectedPrice:      sdk.MustNewDecFromStr("50.1"),
		},
		"One cent increments at the tenths place: 1": {
			tickIndex:          sdk.NewInt(-2),
			exponentAtPriceOne: sdk.NewInt(-1),
			expectedPrice:      sdk.MustNewDecFromStr("0.98"),
		},
		"One tenth of a cent increments at the hundredths place: 1": {
			tickIndex:          sdk.NewInt(-2),
			exponentAtPriceOne: sdk.NewInt(-2),
			expectedPrice:      sdk.MustNewDecFromStr("0.998"),
		},
		"One hundredths of a cent increments at the thousandths place: 1": {
			tickIndex:          sdk.NewInt(-2),
			exponentAtPriceOne: sdk.NewInt(-3),
			expectedPrice:      sdk.MustNewDecFromStr("0.9998"),
		},
		"One ten millionth of a cent increments at the hundred millionths place: 1": {
			tickIndex:          sdk.NewInt(-2),
			exponentAtPriceOne: sdk.NewInt(-7),
			expectedPrice:      sdk.MustNewDecFromStr("0.99999998"),
		},
		"One ten millionth of a cent increments at the hundred millionths place: 2": {
			tickIndex:          sdk.NewInt(-99999111),
			exponentAtPriceOne: sdk.NewInt(-7),
			expectedPrice:      sdk.MustNewDecFromStr("0.090000889"),
		},
		"More variety of numbers in each place": {
			tickIndex:          sdk.NewInt(4030301),
			exponentAtPriceOne: sdk.NewInt(-5),
			expectedPrice:      sdk.MustNewDecFromStr("53030.1"),
		},
		"Max tick and min k": {
			tickIndex:          sdk.NewInt(3420),
			exponentAtPriceOne: sdk.NewInt(-1),
			expectedPrice:      types.MaxSpotPrice,
		},
		"Min tick and max k": {
			tickIndex:          sdk.NewInt(-162000000000000),
			exponentAtPriceOne: sdk.NewInt(-12),
			expectedPrice:      types.MinSpotPrice,
		},
		"error: tickIndex less than minimum": {
			tickIndex:          sdk.NewInt(DefaultMinTick - 1),
			exponentAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:      types.TickIndexMinimumError{MinTick: DefaultMinTick},
		},
		"error: tickIndex greater than maximum": {
			tickIndex:          sdk.NewInt(DefaultMaxTick + 1),
			exponentAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:      types.TickIndexMaximumError{MaxTick: DefaultMaxTick},
		},
		"error: exponentAtPriceOne less than minimum": {
			tickIndex:          sdk.NewInt(100),
			exponentAtPriceOne: types.ExponentAtPriceOneMin.Sub(sdk.OneInt()),
			expectedError:      types.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: types.ExponentAtPriceOneMin.Sub(sdk.OneInt()), PrecisionValueAtPriceOneMin: types.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: types.ExponentAtPriceOneMax},
		},
		"error: exponentAtPriceOne greater than maximum": {
			tickIndex:          sdk.NewInt(100),
			exponentAtPriceOne: types.ExponentAtPriceOneMax.Add(sdk.OneInt()),
			expectedError:      types.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: types.ExponentAtPriceOneMax.Add(sdk.OneInt()), PrecisionValueAtPriceOneMin: types.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: types.ExponentAtPriceOneMax},
		},
		"random": {
			tickIndex:          sdk.NewInt(-9111000000),
			exponentAtPriceOne: sdk.NewInt(-8),
			expectedPrice:      sdk.MustNewDecFromStr("0.000000000088900000"),
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			sqrtPrice, err := math.TickToSqrtPrice(tc.tickIndex, tc.exponentAtPriceOne)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError.Error(), err.Error())
				return
			}
			suite.Require().NoError(err)
			expectedSqrtPrice, err := tc.expectedPrice.ApproxSqrt()
			suite.Require().NoError(err)
			suite.Require().Equal(expectedSqrtPrice.String(), sqrtPrice.String())

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
			price:              sdk.MustNewDecFromStr("5.01"),
			exponentAtPriceOne: sdk.NewInt(-2),
			tickExpected:       "401",
		},
		"50000.01 to tick with -6 k at price one": {
			price:              sdk.MustNewDecFromStr("50000.01"),
			exponentAtPriceOne: sdk.NewInt(-6),
			tickExpected:       "40000001",
		},
		"0.090000889 to tick with -8 k at price one": {
			price:              sdk.MustNewDecFromStr("0.090000889"),
			exponentAtPriceOne: sdk.NewInt(-8),
			tickExpected:       "-999991110",
		},
		"0.9998 to tick with -4 k at price one": {
			price:              sdk.MustNewDecFromStr("0.9998"),
			exponentAtPriceOne: sdk.NewInt(-4),
			tickExpected:       "-20",
		},
		"53030.10 to tick with -5 k at price one": {
			price:              sdk.MustNewDecFromStr("53030.1"),
			exponentAtPriceOne: sdk.NewInt(-5),
			tickExpected:       "4030301",
		},
		"max spot price and minimum exponentAtPriceOne": {
			price:              types.MaxSpotPrice,
			exponentAtPriceOne: sdk.NewInt(-1),
			tickExpected:       "3420",
		},
		"min spot price and minimum exponentAtPriceOne": {
			price:              types.MinSpotPrice,
			exponentAtPriceOne: sdk.NewInt(-1),
			tickExpected:       "-1620",
		},
		"error: max spot price plus one and minimum exponentAtPriceOne": {
			price:              types.MaxSpotPrice.Add(sdk.OneDec()),
			exponentAtPriceOne: sdk.NewInt(-1),
			expectedError:      types.PriceBoundError{ProvidedPrice: types.MaxSpotPrice.Add(sdk.OneDec()), MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice},
		},
		"error: price must be positive": {
			price:              sdk.NewDec(-1),
			exponentAtPriceOne: sdk.NewInt(-6),
			expectedError:      fmt.Errorf("price must be greater than zero"),
		},
		"error: exponentAtPriceOne less than minimum": {
			price:              sdk.NewDec(50000),
			exponentAtPriceOne: types.ExponentAtPriceOneMin.Sub(sdk.OneInt()),
			expectedError:      types.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: types.ExponentAtPriceOneMin.Sub(sdk.OneInt()), PrecisionValueAtPriceOneMin: types.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: types.ExponentAtPriceOneMax},
		},
		"error: exponentAtPriceOne greater than maximum": {
			price:              sdk.NewDec(50000),
			exponentAtPriceOne: types.ExponentAtPriceOneMax.Add(sdk.OneInt()),
			expectedError:      types.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: types.ExponentAtPriceOneMax.Add(sdk.OneInt()), PrecisionValueAtPriceOneMin: types.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: types.ExponentAtPriceOneMax},
		},
		"random": {
			price:              sdk.MustNewDecFromStr("0.0000000000889"),
			exponentAtPriceOne: sdk.NewInt(-8),
			tickExpected:       "-9111000000",
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

// TestTickToSqrtPricePriceToTick_InverseRelationship tests that ensuring the inverse calculation
// between the two methods: tick to square root price to power of 2 and price to tick
func (suite *ConcentratedMathTestSuite) TestTickToSqrtPricePriceToTick_InverseRelationship() {
	testCases := map[string]struct {
		price              sdk.Dec
		exponentAtPriceOne sdk.Int
		tickExpected       string
	}{
		"50000 to tick with -4 k at price one": {
			price:              sdk.NewDec(50000),
			exponentAtPriceOne: sdk.NewInt(-4),
			tickExpected:       "400000",
		},
		"5.01 to tick with -2 k at price one": {
			price:              sdk.MustNewDecFromStr("5.01"),
			exponentAtPriceOne: sdk.NewInt(-2),
			tickExpected:       "401",
		},
		"50000.01 to tick with -6 k at price one": {
			price:              sdk.MustNewDecFromStr("50000.01"),
			exponentAtPriceOne: sdk.NewInt(-6),
			tickExpected:       "40000001",
		},
		"0.090000889 to tick with -8 k at price one": {
			price:              sdk.MustNewDecFromStr("0.090000889"),
			exponentAtPriceOne: sdk.NewInt(-8),
			tickExpected:       "-999991110",
		},
		"0.9998 to tick with -4 k at price one": {
			price:              sdk.MustNewDecFromStr("0.9998"),
			exponentAtPriceOne: sdk.NewInt(-4),
			tickExpected:       "-20",
		},
		"53030.10 to tick with -5 k at price one": {
			price:              sdk.MustNewDecFromStr("53030.1"),
			exponentAtPriceOne: sdk.NewInt(-5),
			tickExpected:       "4030301",
		},
		"max spot price and minimum exponentAtPriceOne": {
			price:              types.MaxSpotPrice,
			exponentAtPriceOne: sdk.NewInt(-1),
			tickExpected:       "3420",
		},
		"min spot price and minimum exponentAtPriceOne": {
			price:              types.MinSpotPrice,
			exponentAtPriceOne: sdk.NewInt(-1),
			tickExpected:       "-1620",
		},
		"random": {
			price:              sdk.MustNewDecFromStr("0.0000000000889"),
			exponentAtPriceOne: sdk.NewInt(-8),
			tickExpected:       "-9111000000",
		},
	}
	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			tick, err := math.PriceToTick(tc.price, tc.exponentAtPriceOne)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.tickExpected, tick.String())

			sqrtPrice, err := math.TickToSqrtPrice(tick, tc.exponentAtPriceOne)
			price := sqrtPrice.Power(2)
			deltaPrice := tc.price.Sub(price).Abs()

			roundingTolerance := sdk.MustNewDecFromStr("0.0001")
			suite.Require().True(deltaPrice.LTE(roundingTolerance))
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestCalculatePriceAndTicksPassed() {
	testCases := map[string]struct {
		price                            sdk.Dec
		exponentAtPriceOne               sdk.Int
		expectedCurrentPrice             sdk.Dec
		expectedTicksPassed              sdk.Int
		expectedAdditiveIncrementInTicks osmomath.BigDec
	}{
		"Price greater than 1": {
			price:                            sdk.MustNewDecFromStr("9.78"),
			exponentAtPriceOne:               sdk.NewInt(-5),
			expectedCurrentPrice:             sdk.NewDec(10),
			expectedTicksPassed:              sdk.NewInt(900000),
			expectedAdditiveIncrementInTicks: osmomath.MustNewDecFromStr("0.00001"),
		},
		"Price less than 1": {
			price:                            sdk.MustNewDecFromStr("0.71"),
			exponentAtPriceOne:               sdk.NewInt(-6),
			expectedCurrentPrice:             sdk.MustNewDecFromStr("0.1"),
			expectedTicksPassed:              sdk.NewInt(-9000000),
			expectedAdditiveIncrementInTicks: osmomath.MustNewDecFromStr("0.0000001"),
		},
	}
	for name, tt := range testCases {
		suite.Run(name, func() {
			currentPrice, ticksPassed, currentAdditiveIncrementInTicks := math.CalculatePriceAndTicksPassed(tt.price, tt.exponentAtPriceOne)
			suite.Require().Equal(tt.expectedCurrentPrice.String(), currentPrice.String())
			suite.Require().Equal(tt.expectedTicksPassed.String(), ticksPassed.String())
			suite.Require().Equal(tt.expectedAdditiveIncrementInTicks.String(), currentAdditiveIncrementInTicks.String())
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestGetMinAndMaxTicksFromExponentAtPriceOneInternal() {
	testCases := map[string]struct {
		price              sdk.Dec
		exponentAtPriceOne sdk.Int
		expectedMinTick    int64
		expectedMaxTick    int64
	}{
		"exponentAtPriceOne = -1": {
			exponentAtPriceOne: sdk.NewInt(-1),
			expectedMinTick:    types.MinTickNegOne,
			expectedMaxTick:    types.MaxTickNegOne,
		},
		"exponentAtPriceOne = -6": {
			exponentAtPriceOne: sdk.NewInt(-6),
			expectedMinTick:    types.MinTickNegSix,
			expectedMaxTick:    types.MaxTickNegSix,
		},
		"exponentAtPriceOne = -12": {
			exponentAtPriceOne: sdk.NewInt(-12),
			expectedMinTick:    types.MinTickNegTwelve,
			expectedMaxTick:    types.MaxTickNegTwelve,
		},
		"exponentAtPriceOne = -13 (non pre-computed value)": {
			exponentAtPriceOne: sdk.NewInt(-13),
			expectedMinTick: func() int64 {
				minTick, _ := math.ComputeMinAndMaxTicksFromExponentAtPriceOneInternal(sdk.NewInt(-13))
				return minTick
			}(),
			expectedMaxTick: func() int64 {
				_, maxTick := math.ComputeMinAndMaxTicksFromExponentAtPriceOneInternal(sdk.NewInt(-13))
				return maxTick
			}(),
		},
	}
	for name, tt := range testCases {
		suite.Run(name, func() {
			minTick, maxTick := math.GetMinAndMaxTicksFromExponentAtPriceOneInternal(tt.exponentAtPriceOne)
			suite.Require().Equal(tt.expectedMinTick, minTick)
			suite.Require().Equal(tt.expectedMaxTick, maxTick)
		})
	}
}
