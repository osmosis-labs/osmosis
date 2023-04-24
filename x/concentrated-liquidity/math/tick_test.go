package math_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
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
		tickIndex     sdk.Int
		expectedPrice sdk.Dec
		expectedError error
	}{
		"Ten billionths cent increments at the millionths place: 1": {
			tickIndex:     sdk.NewInt(-51630100),
			expectedPrice: sdk.MustNewDecFromStr("0.0000033699"),
		},
		"Ten billionths cent increments at the millionths place: 2": {
			tickIndex:     sdk.NewInt(-51630000),
			expectedPrice: sdk.MustNewDecFromStr("0.0000033700"),
		},
		"One millionths cent increments at the hundredths place: 1": {
			tickIndex:     sdk.NewInt(-11999800),
			expectedPrice: sdk.MustNewDecFromStr("0.070002"),
		},
		"One millionths cent increments at the hundredths place: 2": {
			tickIndex:     sdk.NewInt(-11999700),
			expectedPrice: sdk.MustNewDecFromStr("0.070003"),
		},
		"One hundred thousandth cent increments at the tenths place: 1": {
			tickIndex:     sdk.NewInt(-999800),
			expectedPrice: sdk.MustNewDecFromStr("0.90002"),
		},
		"One hundred thousandth cent increments at the tenths place: 2": {
			tickIndex:     sdk.NewInt(-999700),
			expectedPrice: sdk.MustNewDecFromStr("0.90003"),
		},
		"One ten thousandth cent increments at the ones place: 1": {
			tickIndex:     sdk.NewInt(1000000),
			expectedPrice: sdk.MustNewDecFromStr("2"),
		},
		"One dollar increments at the ten thousands place: 2": {
			tickIndex:     sdk.NewInt(1000100),
			expectedPrice: sdk.MustNewDecFromStr("2.0001"),
		},
		"One thousandth cent increments at the tens place: 1": {
			tickIndex:     sdk.NewInt(9200100),
			expectedPrice: sdk.MustNewDecFromStr("12.001"),
		},
		"One thousandth cent increments at the tens place: 2": {
			tickIndex:     sdk.NewInt(9200200),
			expectedPrice: sdk.MustNewDecFromStr("12.002"),
		},
		"One cent increments at the hundreds place: 1": {
			tickIndex:     sdk.NewInt(18320100),
			expectedPrice: sdk.MustNewDecFromStr("132.01"),
		},
		"One cent increments at the hundreds place: 2": {
			tickIndex:     sdk.NewInt(18320200),
			expectedPrice: sdk.MustNewDecFromStr("132.02"),
		},
		"Ten cent increments at the thousands place: 1": {
			tickIndex:     sdk.NewInt(27732100),
			expectedPrice: sdk.MustNewDecFromStr("1732.10"),
		},
		"Ten cent increments at the thousands place: 2": {
			tickIndex:     sdk.NewInt(27732200),
			expectedPrice: sdk.MustNewDecFromStr("1732.20"),
		},
		"Dollar increments at the ten thousands place: 1": {
			tickIndex:     sdk.NewInt(36073200),
			expectedPrice: sdk.MustNewDecFromStr("10732"),
		},
		"Dollar increments at the ten thousands place: 2": {
			tickIndex:     sdk.NewInt(36073300),
			expectedPrice: sdk.MustNewDecFromStr("10733"),
		},
		"Max tick and min k": {
			tickIndex:     sdk.NewInt(342000000),
			expectedPrice: types.MaxSpotPrice,
		},
		"Min tick and max k": {
			tickIndex:     sdk.NewInt(-162000000),
			expectedPrice: types.MinSpotPrice,
		},
		"error: tickIndex less than minimum": {
			tickIndex:     sdk.NewInt(-162000000 - 1),
			expectedError: types.TickIndexMinimumError{MinTick: -162000000},
		},
		"error: tickIndex greater than maximum": {
			tickIndex:     sdk.NewInt(342000000 + 1),
			expectedError: types.TickIndexMaximumError{MaxTick: 342000000},
		},
		"Gyen <> USD, tick -20594000 -> price 0.0074060": {
			tickIndex:     sdk.NewInt(-20594000),
			expectedPrice: sdk.MustNewDecFromStr("0.007406000000000000"),
		},
		"Gyen <> USD, tick -20594000 + 100 -> price 0.0074061": {
			tickIndex:     sdk.NewInt(-20593900),
			expectedPrice: sdk.MustNewDecFromStr("0.007406100000000000"),
		},
		"Spell <> USD, tick -29204000 -> price 0.00077960": {
			tickIndex:     sdk.NewInt(-29204000),
			expectedPrice: sdk.MustNewDecFromStr("0.000779600000000000"),
		},
		"Spell <> USD, tick -29204000 + 100 -> price 0.00077961": {
			tickIndex:     sdk.NewInt(-29203900),
			expectedPrice: sdk.MustNewDecFromStr("0.000779610000000000"),
		},
		"Atom <> Osmo, tick -12150000 -> price 0.068500": {
			tickIndex:     sdk.NewInt(-12150000),
			expectedPrice: sdk.MustNewDecFromStr("0.068500000000000000"),
		},
		"Atom <> Osmo, tick -12150000 + 100 -> price 0.068501": {
			tickIndex:     sdk.NewInt(-12149900),
			expectedPrice: sdk.MustNewDecFromStr("0.068501000000000000"),
		},
		"Boot <> Osmo, tick 64576000 -> price 25760000": {
			tickIndex:     sdk.NewInt(64576000),
			expectedPrice: sdk.MustNewDecFromStr("25760000"),
		},
		"Boot <> Osmo, tick 64576000 + 100 -> price 25760000": {
			tickIndex:     sdk.NewInt(64576100),
			expectedPrice: sdk.MustNewDecFromStr("25761000"),
		},
		"BTC <> USD, tick 38035200 -> price 30352": {
			tickIndex:     sdk.NewInt(38035200),
			expectedPrice: sdk.MustNewDecFromStr("30352"),
		},
		"BTC <> USD, tick 38035200 + 100 -> price 30353": {
			tickIndex:     sdk.NewInt(38035300),
			expectedPrice: sdk.MustNewDecFromStr("30353"),
		},
		"SHIB <> USD, tick -44821000 -> price 0.000011790": {
			tickIndex:     sdk.NewInt(-44821000),
			expectedPrice: sdk.MustNewDecFromStr("0.00001179"),
		},
		"SHIB <> USD, tick -44821100 + 100 -> price 0.000011791": {
			tickIndex:     sdk.NewInt(-44820900),
			expectedPrice: sdk.MustNewDecFromStr("0.000011791"),
		},
		"ETH <> BTC, tick -12104000 -> price 0.068960": {
			tickIndex:     sdk.NewInt(-12104000),
			expectedPrice: sdk.MustNewDecFromStr("0.068960"),
		},
		"ETH <> BTC, tick -121044000 + 1 -> price 0.068961": {
			tickIndex:     sdk.NewInt(-12103900),
			expectedPrice: sdk.MustNewDecFromStr("0.068961"),
		},
		"one tick spacing interval smaller than max sqrt price, max tick neg six - 100 -> one tick spacing interval smaller than max sqrt price": {
			tickIndex:     sdk.NewInt(types.MaxTick).Sub(sdk.NewInt(100)),
			expectedPrice: sdk.MustNewDecFromStr("99999000000000000000000000000000000014"), // there is some excess here due to the math functions.
		},
		"max sqrt price, max tick neg six -> max spot price": {
			tickIndex:     sdk.NewInt(types.MaxTick),
			expectedPrice: types.MaxSpotPrice,
		},
	}

	for name, tc := range testCases {
		tc := tc
		suite.Run(name, func() {
			sqrtPrice, err := math.TickToSqrtPrice(tc.tickIndex)
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
		price         sdk.Dec
		tickExpected  string
		expectedError error
	}{
		"BTC <> USD, tick 38035200 -> price 30352": {
			price:        sdk.MustNewDecFromStr("30352"),
			tickExpected: "38035200",
		},
		"BTC <> USD, tick 38035300 + 100 -> price 30353": {
			price:        sdk.MustNewDecFromStr("30353"),
			tickExpected: "38035300",
		},
		"SHIB <> USD, tick -44821000 -> price 0.000011790": {
			price:        sdk.MustNewDecFromStr("0.000011790"),
			tickExpected: "-44821000",
		},
		"SHIB <> USD, tick -44820900 -> price 0.000011791": {
			price:        sdk.MustNewDecFromStr("0.000011791"),
			tickExpected: "-44820900",
		},
		"ETH <> BTC, tick -12104000 -> price 0.068960": {
			price:        sdk.MustNewDecFromStr("0.068960"),
			tickExpected: "-12104000",
		},
		"ETH <> BTC, tick -12104000 + 100 -> price 0.068961": {
			price:        sdk.MustNewDecFromStr("0.068961"),
			tickExpected: "-12103900",
		},
		"max sqrt price -1, max neg tick six - 100 -> max tick neg six - 100": {
			price:        sdk.MustNewDecFromStr("99999000000000000000000000000000000000"),
			tickExpected: sdk.NewInt(types.MaxTick - 100).String(),
		},
		"max sqrt price, max tick neg six -> max spot price": {

			price:        types.MaxSqrtPrice.Power(2),
			tickExpected: sdk.NewInt(types.MaxTick).String(),
		},
		"Gyen <> USD, tick -20594000 -> price 0.0074060": {
			price:        sdk.MustNewDecFromStr("0.007406"),
			tickExpected: "-20594000",
		},
		"Gyen <> USD, tick -20594000 + 100 -> price 0.0074061": {
			price:        sdk.MustNewDecFromStr("0.0074061"),
			tickExpected: "-20593900",
		},
		"Spell <> USD, tick -29204000 -> price 0.00077960": {
			price:        sdk.MustNewDecFromStr("0.0007796"),
			tickExpected: "-29204000",
		},
		"Spell <> USD, tick -29204000 + 100 -> price 0.00077961": {
			price:        sdk.MustNewDecFromStr("0.00077961"),
			tickExpected: "-29203900",
		},
		"Atom <> Osmo, tick -12150000 -> price 0.068500": {
			price:        sdk.MustNewDecFromStr("0.0685"),
			tickExpected: "-12150000",
		},
		"Atom <> Osmo, tick -12150000 + 100 -> price 0.068501": {
			price:        sdk.MustNewDecFromStr("0.068501"),
			tickExpected: "-12149900",
		},
		"Boot <> Osmo, tick 64576000 -> price 25760000": {
			price:        sdk.MustNewDecFromStr("25760000"),
			tickExpected: "64576000",
		},
		"Boot <> Osmo, tick 64576000 + 100 -> price 25761000": {
			price:        sdk.MustNewDecFromStr("25761000"),
			tickExpected: "64576100",
		},
	}
	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			tickSpacing := uint64(100)
			tick, err := math.PriceToTick(tc.price, tickSpacing)
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
		price        sdk.Dec
		tickExpected string
	}{
		"50000 to tick": {
			price:        sdk.MustNewDecFromStr("50000"),
			tickExpected: "40000000",
		},
		"5.01 to tick": {
			price:        sdk.MustNewDecFromStr("5.01"),
			tickExpected: "4010000",
		},
		"50000.01 to tick": {
			price:        sdk.MustNewDecFromStr("50000.01"),
			tickExpected: "40000001",
		},
		"0.090001 to tick": {
			price:        sdk.MustNewDecFromStr("0.090001"),
			tickExpected: "-9999900",
		},
		"0.9998 to tick": {
			price:        sdk.MustNewDecFromStr("0.9998"),
			tickExpected: "-2000",
		},
		"53030 to tick": {
			price:        sdk.MustNewDecFromStr("53030"),
			tickExpected: "40303000",
		},
		"max spot price": {
			price:        types.MaxSpotPrice,
			tickExpected: "342000000",
		},
		"min spot price": {
			price:        types.MinSpotPrice,
			tickExpected: "-162000000",
		},
	}
	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			tickSpacing := uint64(1)
			tick, err := math.PriceToTick(tc.price, tickSpacing)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.tickExpected, tick.String())

			sqrtPrice, err := math.TickToSqrtPrice(tick)
			price := sqrtPrice.Power(2)
			deltaPrice := tc.price.Sub(price).Abs()

			roundingTolerance := sdk.MustNewDecFromStr("0.0001")
			suite.Require().True(deltaPrice.LTE(roundingTolerance))
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestCalculatePriceToTick() {
	testCases := map[string]struct {
		price             sdk.Dec
		expectedTickIndex sdk.Int
	}{
		"Price greater than 1": {
			price:             sdk.MustNewDecFromStr("9.78"),
			expectedTickIndex: sdk.NewInt(8780000),
		},
		"Price less than 1": {
			price:             sdk.MustNewDecFromStr("0.71"),
			expectedTickIndex: sdk.NewInt(-2900000),
		},
	}
	for name, t := range testCases {
		suite.Run(name, func() {
			tickIndex := math.CalculatePriceToTick(t.price)
			suite.Require().Equal(t.expectedTickIndex.String(), tickIndex.String())
		})
	}
}
