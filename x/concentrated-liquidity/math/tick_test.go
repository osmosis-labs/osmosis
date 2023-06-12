package math_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

const (
	defaultTickSpacing = 100
)

var (
	// spot price - (10^(spot price exponent - 6 - 1))
	// Note we get spot price exponent by counting the number of digits in the max spot price and subtracting 1.
	closestPriceBelowMaxPriceDefaultTickSpacing = types.MaxSpotPrice.Sub(sdk.NewDec(10).PowerMut(uint64(len(types.MaxSpotPrice.TruncateInt().String()) - 1 - int(-types.ExponentAtPriceOne) - 1)))
	// min tick + 10 ^ -expoentAtPriceOne
	closestTickAboveMinPriceDefaultTickSpacing = sdk.NewInt(types.MinTick).Add(sdk.NewInt(10).ToDec().Power(uint64(types.ExponentAtPriceOne * -1)).TruncateInt())
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
		tickIndex     int64
		expectedPrice sdk.Dec
		expectedError error
	}{
		"Ten billionths cent increments at the millionths place: 1": {
			tickIndex:     -51630100,
			expectedPrice: sdk.MustNewDecFromStr("0.0000033699"),
		},
		"Ten billionths cent increments at the millionths place: 2": {
			tickIndex:     -51630000,
			expectedPrice: sdk.MustNewDecFromStr("0.0000033700"),
		},
		"One millionths cent increments at the hundredths place: 1": {
			tickIndex:     -11999800,
			expectedPrice: sdk.MustNewDecFromStr("0.070002"),
		},
		"One millionths cent increments at the hundredths place: 2": {
			tickIndex:     -11999700,
			expectedPrice: sdk.MustNewDecFromStr("0.070003"),
		},
		"One hundred thousandth cent increments at the tenths place: 1": {
			tickIndex:     -999800,
			expectedPrice: sdk.MustNewDecFromStr("0.90002"),
		},
		"One hundred thousandth cent increments at the tenths place: 2": {
			tickIndex:     -999700,
			expectedPrice: sdk.MustNewDecFromStr("0.90003"),
		},
		"One ten thousandth cent increments at the ones place: 1": {
			tickIndex:     1000000,
			expectedPrice: sdk.MustNewDecFromStr("2"),
		},
		"One dollar increments at the ten thousands place: 2": {
			tickIndex:     1000100,
			expectedPrice: sdk.MustNewDecFromStr("2.0001"),
		},
		"One thousandth cent increments at the tens place: 1": {
			tickIndex:     9200100,
			expectedPrice: sdk.MustNewDecFromStr("12.001"),
		},
		"One thousandth cent increments at the tens place: 2": {
			tickIndex:     9200200,
			expectedPrice: sdk.MustNewDecFromStr("12.002"),
		},
		"One cent increments at the hundreds place: 1": {
			tickIndex:     18320100,
			expectedPrice: sdk.MustNewDecFromStr("132.01"),
		},
		"One cent increments at the hundreds place: 2": {
			tickIndex:     18320200,
			expectedPrice: sdk.MustNewDecFromStr("132.02"),
		},
		"Ten cent increments at the thousands place: 1": {
			tickIndex:     27732100,
			expectedPrice: sdk.MustNewDecFromStr("1732.10"),
		},
		"Ten cent increments at the thousands place: 2": {
			tickIndex:     27732200,
			expectedPrice: sdk.MustNewDecFromStr("1732.20"),
		},
		"Dollar increments at the ten thousands place: 1": {
			tickIndex:     36073200,
			expectedPrice: sdk.MustNewDecFromStr("10732"),
		},
		"Dollar increments at the ten thousands place: 2": {
			tickIndex:     36073300,
			expectedPrice: sdk.MustNewDecFromStr("10733"),
		},
		"Max tick and min k": {
			tickIndex:     342000000,
			expectedPrice: types.MaxSpotPrice,
		},
		"Min tick and max k": {
			tickIndex:     -162000000,
			expectedPrice: types.MinSpotPrice,
		},
		"error: tickIndex less than minimum": {
			tickIndex:     -162000000 - 1,
			expectedError: types.TickIndexMinimumError{MinTick: -162000000},
		},
		"error: tickIndex greater than maximum": {
			tickIndex:     342000000 + 1,
			expectedError: types.TickIndexMaximumError{MaxTick: 342000000},
		},
		"Gyen <> USD, tick -20594000 -> price 0.0074060": {
			tickIndex:     -20594000,
			expectedPrice: sdk.MustNewDecFromStr("0.007406000000000000"),
		},
		"Gyen <> USD, tick -20594000 + 100 -> price 0.0074061": {
			tickIndex:     -20593900,
			expectedPrice: sdk.MustNewDecFromStr("0.007406100000000000"),
		},
		"Spell <> USD, tick -29204000 -> price 0.00077960": {
			tickIndex:     -29204000,
			expectedPrice: sdk.MustNewDecFromStr("0.000779600000000000"),
		},
		"Spell <> USD, tick -29204000 + 100 -> price 0.00077961": {
			tickIndex:     -29203900,
			expectedPrice: sdk.MustNewDecFromStr("0.000779610000000000"),
		},
		"Atom <> Osmo, tick -12150000 -> price 0.068500": {
			tickIndex:     -12150000,
			expectedPrice: sdk.MustNewDecFromStr("0.068500000000000000"),
		},
		"Atom <> Osmo, tick -12150000 + 100 -> price 0.068501": {
			tickIndex:     -12149900,
			expectedPrice: sdk.MustNewDecFromStr("0.068501000000000000"),
		},
		"Boot <> Osmo, tick 64576000 -> price 25760000": {
			tickIndex:     64576000,
			expectedPrice: sdk.MustNewDecFromStr("25760000"),
		},
		"Boot <> Osmo, tick 64576000 + 100 -> price 25760000": {
			tickIndex:     64576100,
			expectedPrice: sdk.MustNewDecFromStr("25761000"),
		},
		"BTC <> USD, tick 38035200 -> price 30352": {
			tickIndex:     38035200,
			expectedPrice: sdk.MustNewDecFromStr("30352"),
		},
		"BTC <> USD, tick 38035200 + 100 -> price 30353": {
			tickIndex:     38035300,
			expectedPrice: sdk.MustNewDecFromStr("30353"),
		},
		"SHIB <> USD, tick -44821000 -> price 0.000011790": {
			tickIndex:     -44821000,
			expectedPrice: sdk.MustNewDecFromStr("0.00001179"),
		},
		"SHIB <> USD, tick -44821100 + 100 -> price 0.000011791": {
			tickIndex:     -44820900,
			expectedPrice: sdk.MustNewDecFromStr("0.000011791"),
		},
		"ETH <> BTC, tick -12104000 -> price 0.068960": {
			tickIndex:     -12104000,
			expectedPrice: sdk.MustNewDecFromStr("0.068960"),
		},
		"ETH <> BTC, tick -121044000 + 1 -> price 0.068961": {
			tickIndex:     -12103900,
			expectedPrice: sdk.MustNewDecFromStr("0.068961"),
		},
		"one tick spacing interval smaller than max sqrt price, max tick neg six - 100 -> one tick spacing interval smaller than max sqrt price": {
			tickIndex:     types.MaxTick - 100,
			expectedPrice: sdk.MustNewDecFromStr("99999000000000000000000000000000000000"),
		},
		"max sqrt price, max tick neg six -> max spot price": {
			tickIndex:     types.MaxTick,
			expectedPrice: types.MaxSpotPrice,
		},
	}

	for name, tc := range testCases {
		tc := tc
		suite.Run(name, func() {
			price, sqrtPrice, err := math.TickToSqrtPrice(tc.tickIndex)
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError.Error(), err.Error())
				return
			}
			suite.Require().NoError(err)
			expectedSqrtPrice, err := tc.expectedPrice.ApproxSqrt()
			suite.Require().NoError(err)

			suite.Require().Equal(tc.expectedPrice.String(), price.String())
			suite.Require().Equal(expectedSqrtPrice.String(), sqrtPrice.String())
		})
	}
}
func (suite *ConcentratedMathTestSuite) TestTicksToSqrtPrice() {
	testCases := map[string]struct {
		lowerTickIndex     sdk.Int
		upperTickIndex     sdk.Int
		expectedLowerPrice sdk.Dec
		expectedUpperPrice sdk.Dec
		expectedError      error
	}{
		"Ten billionths cent increments at the millionths place": {
			lowerTickIndex:     sdk.NewInt(-51630100),
			upperTickIndex:     sdk.NewInt(-51630000),
			expectedLowerPrice: sdk.MustNewDecFromStr("0.0000033699"),
			expectedUpperPrice: sdk.MustNewDecFromStr("0.0000033700"),
		},
		"One millionths cent increments at the hundredths place:": {
			lowerTickIndex:     sdk.NewInt(-11999800),
			upperTickIndex:     sdk.NewInt(-11999700),
			expectedLowerPrice: sdk.MustNewDecFromStr("0.070002"),
			expectedUpperPrice: sdk.MustNewDecFromStr("0.070003"),
		},
		"One hundred thousandth cent increments at the tenths place": {
			lowerTickIndex:     sdk.NewInt(-999800),
			upperTickIndex:     sdk.NewInt(-999700),
			expectedLowerPrice: sdk.MustNewDecFromStr("0.90002"),
			expectedUpperPrice: sdk.MustNewDecFromStr("0.90003"),
		},
		"Dollar increments at the ten thousands place": {
			lowerTickIndex:     sdk.NewInt(36073200),
			upperTickIndex:     sdk.NewInt(36073300),
			expectedLowerPrice: sdk.MustNewDecFromStr("10732"),
			expectedUpperPrice: sdk.MustNewDecFromStr("10733"),
		},
		"Max tick and min k": {
			lowerTickIndex:     sdk.NewInt(-162000000),
			upperTickIndex:     sdk.NewInt(342000000),
			expectedUpperPrice: types.MaxSpotPrice,
			expectedLowerPrice: types.MinSpotPrice,
		},
		"error: lowerTickIndex less than minimum": {
			lowerTickIndex: sdk.NewInt(-162000000 - 1),
			upperTickIndex: sdk.NewInt(36073300),
			expectedError:  types.TickIndexMinimumError{MinTick: -162000000},
		},
		"error: upperTickIndex greater than maximum": {
			lowerTickIndex: sdk.NewInt(-162000000),
			upperTickIndex: sdk.NewInt(342000000 + 1),
			expectedError:  types.TickIndexMaximumError{MaxTick: 342000000},
		},
		"error: provided lower tick and upper tick are same": {
			lowerTickIndex: sdk.NewInt(-162000000),
			upperTickIndex: sdk.NewInt(-162000000),
			expectedError:  types.InvalidLowerUpperTickError{LowerTick: sdk.NewInt(-162000000).Int64(), UpperTick: sdk.NewInt(-162000000).Int64()},
		},
	}

	for name, tc := range testCases {
		tc := tc
		suite.Run(name, func() {
			priceLower, priceUpper, lowerSqrtPrice, upperSqrtPrice, err := math.TicksToSqrtPrice(tc.lowerTickIndex.Int64(), tc.upperTickIndex.Int64())
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError.Error(), err.Error())
				return
			}
			suite.Require().NoError(err)

			// convert test case's prices to sqrt price
			expectedLowerSqrtPrice, err := tc.expectedLowerPrice.ApproxSqrt()
			suite.Require().NoError(err)
			expectedUpperSqrtPrice, err := tc.expectedUpperPrice.ApproxSqrt()
			suite.Require().NoError(err)

			suite.Require().Equal(tc.expectedLowerPrice.String(), priceLower.String())
			suite.Require().Equal(tc.expectedUpperPrice.String(), priceUpper.String())
			suite.Require().Equal(expectedLowerSqrtPrice.String(), lowerSqrtPrice.String())
			suite.Require().Equal(expectedUpperSqrtPrice.String(), upperSqrtPrice.String())
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestPriceToTick() {
	const (
		one = uint64(1)
	)

	testCases := map[string]struct {
		price        sdk.Dec
		tickExpected int64
	}{
		"BTC <> USD, tick 38035200 -> price 30352": {
			price:        sdk.MustNewDecFromStr("30352"),
			tickExpected: 38035200,
		},
		"BTC <> USD, tick 38035300 + 100 -> price 30353": {
			price:        sdk.MustNewDecFromStr("30353"),
			tickExpected: 38035300,
		},
		"SHIB <> USD, tick -44821000 -> price 0.000011790": {
			price:        sdk.MustNewDecFromStr("0.000011790"),
			tickExpected: -44821000,
		},
		"SHIB <> USD, tick -44820900 -> price 0.000011791": {
			price:        sdk.MustNewDecFromStr("0.000011791"),
			tickExpected: -44820900,
		},
		"ETH <> BTC, tick -12104000 -> price 0.068960": {
			price:        sdk.MustNewDecFromStr("0.068960"),
			tickExpected: -12104000,
		},
		"ETH <> BTC, tick -12104000 + 100 -> price 0.068961": {
			price:        sdk.MustNewDecFromStr("0.068961"),
			tickExpected: -12103900,
		},
		"max sqrt price -1, max neg tick six - 100 -> max tick neg six - 100": {
			price:        sdk.MustNewDecFromStr("99999000000000000000000000000000000000"),
			tickExpected: types.MaxTick - 100,
		},
		"max sqrt price, max tick neg six -> max spot price": {
			price:        types.MaxSqrtPrice.Power(2),
			tickExpected: types.MaxTick,
		},
		"Gyen <> USD, tick -20594000 -> price 0.0074060": {
			price:        sdk.MustNewDecFromStr("0.007406"),
			tickExpected: -20594000,
		},
		"Gyen <> USD, tick -20594000 + 100 -> price 0.0074061": {
			price:        sdk.MustNewDecFromStr("0.0074061"),
			tickExpected: -20593900,
		},
		"Spell <> USD, tick -29204000 -> price 0.00077960": {
			price:        sdk.MustNewDecFromStr("0.0007796"),
			tickExpected: -29204000,
		},
		"Spell <> USD, tick -29204000 + 100 -> price 0.00077961": {
			price:        sdk.MustNewDecFromStr("0.00077961"),
			tickExpected: -29203900,
		},
		"Atom <> Osmo, tick -12150000 -> price 0.068500": {
			price:        sdk.MustNewDecFromStr("0.0685"),
			tickExpected: -12150000,
		},
		"Atom <> Osmo, tick -12150000 + 100 -> price 0.068501": {
			price:        sdk.MustNewDecFromStr("0.068501"),
			tickExpected: -12149900,
		},
		"Boot <> Osmo, tick 64576000 -> price 25760000": {
			price:        sdk.MustNewDecFromStr("25760000"),
			tickExpected: 64576000,
		},
		"Boot <> Osmo, tick 64576000 + 100 -> price 25761000": {
			price:        sdk.MustNewDecFromStr("25761000"),
			tickExpected: 64576100,
		},
		"used price is 1": {
			price:        sdk.OneDec(),
			tickExpected: 0,
		},
	}
	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			tick, err := math.PriceToTick(tc.price)
			// With tick spacing of one, no rounding should occur.
			tickRoundDown, err1 := math.PriceToTickRoundDown(tc.price, one)

			suite.Require().NoError(err)
			suite.Require().NoError(err1)
			suite.Require().Equal(tc.tickExpected, tick)
			suite.Require().Equal(tc.tickExpected, tickRoundDown)
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestPriceToTickRoundDown() {
	testCases := map[string]struct {
		price        sdk.Dec
		tickSpacing  uint64
		tickExpected int64
	}{
		"tick spacing 100, price of 1": {
			price:        sdk.OneDec(),
			tickSpacing:  defaultTickSpacing,
			tickExpected: 0,
		},
		"tick spacing 100, price of 1.000030, tick 30 -> 0": {
			price:        sdk.MustNewDecFromStr("1.000030"),
			tickSpacing:  defaultTickSpacing,
			tickExpected: 0,
		},
		"tick spacing 100, price of 0.9999970, tick -30 -> -100": {
			price:        sdk.MustNewDecFromStr("0.9999970"),
			tickSpacing:  defaultTickSpacing,
			tickExpected: -100,
		},
		"tick spacing 50, price of 0.9999730, tick -270 -> -300": {
			price:        sdk.MustNewDecFromStr("0.9999730"),
			tickSpacing:  50,
			tickExpected: -300,
		},
		"tick spacing 100, MinSpotPrice, MinTick": {
			price:        types.MinSpotPrice,
			tickSpacing:  defaultTickSpacing,
			tickExpected: types.MinTick,
		},
		"tick spacing 100, Spot price one tick above min, one tick above min -> MinTick": {
			price:        types.MinSpotPrice.Add(sdk.SmallestDec()),
			tickSpacing:  defaultTickSpacing,
			tickExpected: closestTickAboveMinPriceDefaultTickSpacing.Int64(),
		},
		"tick spacing 100, Spot price one tick below max, one tick below max -> MaxTick - 1": {
			price:        closestPriceBelowMaxPriceDefaultTickSpacing,
			tickSpacing:  defaultTickSpacing,
			tickExpected: types.MaxTick - 100,
		},
		"tick spacing 100, Spot price 100_000_050 -> 72000000": {
			price:        sdk.NewDec(100_000_050),
			tickSpacing:  defaultTickSpacing,
			tickExpected: 72000000,
		},
		"tick spacing 100, Spot price 100_000_051 -> 72000100 (rounded up to tick spacing)": {
			price:        sdk.NewDec(100_000_051),
			tickSpacing:  defaultTickSpacing,
			tickExpected: 72000000,
		},
		"tick spacing 1, Spot price 100_000_051 -> 72000001 no tick spacing rounding": {
			price:        sdk.NewDec(100_000_051),
			tickSpacing:  1,
			tickExpected: 72000001,
		},
	}
	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			tick, err := math.PriceToTickRoundDown(tc.price, tc.tickSpacing)

			suite.Require().NoError(err)
			suite.Require().Equal(tc.tickExpected, tick)
		})
	}
}

// TestTickToSqrtPricePriceToTick_InverseRelationship tests that ensuring the inverse calculation
// between the two methods: tick to square root price to power of 2 and price to tick
func (suite *ConcentratedMathTestSuite) TestTickToSqrtPricePriceToTick_InverseRelationship() {
	testCases := map[string]struct {
		price          sdk.Dec
		truncatedPrice sdk.Dec
		tickExpected   int64
	}{
		"50000 to tick": {
			price:        sdk.MustNewDecFromStr("50000"),
			tickExpected: 40000000,
		},
		"5.01 to tick": {
			price:        sdk.MustNewDecFromStr("5.01"),
			tickExpected: 4010000,
		},
		"50000.01 to tick": {
			price:        sdk.MustNewDecFromStr("50000.01"),
			tickExpected: 40000001,
		},
		"0.090001 to tick": {
			price:        sdk.MustNewDecFromStr("0.090001"),
			tickExpected: -9999900,
		},
		"0.9998 to tick": {
			price:        sdk.MustNewDecFromStr("0.9998"),
			tickExpected: -2000,
		},
		"53030 to tick": {
			price:        sdk.MustNewDecFromStr("53030"),
			tickExpected: 40303000,
		},
		"max spot price": {
			price:        types.MaxSpotPrice,
			tickExpected: types.MaxTick,
		},
		"max spot price - smallest price delta given exponent at price one of -6": {
			// 37 - 6 is calculated by counting the exponent of max spot price and subtracting exponent at price one
			price:        types.MaxSpotPrice.Sub(sdk.NewDec(10).PowerMut(37 - 6)),
			tickExpected: types.MaxTick - 1, // still max
		},
		"min spot price": {
			price:        types.MinSpotPrice,
			tickExpected: -162000000,
		},
		"smallest + min price increment": {
			price:        sdk.MustNewDecFromStr("0.000000000000000002"),
			tickExpected: -161000000,
		},
		"min price increment 10^1": {
			price:        sdk.MustNewDecFromStr("0.000000000000000009"),
			tickExpected: -154000000,
		},
		"smallest + min price increment 10^1": {
			price:        sdk.MustNewDecFromStr("0.000000000000000010"),
			tickExpected: -153000000,
		},
		"smallest + min price increment * 10^2": {
			price:        sdk.MustNewDecFromStr("0.000000000000000100"),
			tickExpected: -144000000,
		},
		"smallest + min price increment * 10^3": {
			price:        sdk.MustNewDecFromStr("0.000000000000001000"),
			tickExpected: -135000000,
		},
		"smallest + min price increment * 10^4": {
			price:        sdk.MustNewDecFromStr("0.000000000000010000"),
			tickExpected: -126000000,
		},
		"smallest + min price * increment 10^5": {
			price:        sdk.MustNewDecFromStr("0.000000000000100000"),
			tickExpected: -117000000,
		},
		"smallest + min price * increment 10^6": {
			price:        sdk.MustNewDecFromStr("0.000000000001000000"),
			tickExpected: -108000000,
		},
		"smallest + min price * increment 10^6 + tick": {
			price:        sdk.MustNewDecFromStr("0.000000000001000001"),
			tickExpected: -107999999,
		},
		"smallest + min price * increment 10^17": {
			price:        sdk.MustNewDecFromStr("0.100000000000000000"),
			tickExpected: -9000000,
		},
		"smallest + min price * increment 10^18": {
			price:        sdk.MustNewDecFromStr("1.000000000000000000"),
			tickExpected: 0,
		},
		"at price level of 0.01 - odd": {
			price:        sdk.MustNewDecFromStr("0.012345670000000000"),
			tickExpected: -17765433,
		},
		"at price level of 0.01 - even": {
			price:        sdk.MustNewDecFromStr("0.01234568000000000"),
			tickExpected: -17765432,
		},
		"at min price level of 0.01 - odd": {
			price:        sdk.MustNewDecFromStr("0.000000000001234567"),
			tickExpected: -107765433,
		},
		"at min price level of 0.01 - even": {
			price:        sdk.MustNewDecFromStr("0.000000000001234568"),
			tickExpected: -107765432,
		},
		"at price level of 1_000_000_000 - odd end": {
			price:        sdk.MustNewDecFromStr("1234567000"),
			tickExpected: 81234567,
		},
		"at price level of 1_000_000_000 - in-between supported": {
			price:          sdk.MustNewDecFromStr("1234567500"),
			tickExpected:   81234568,
			truncatedPrice: sdk.MustNewDecFromStr("1234568000"),
		},
		"at price level of 1_000_000_000 - even end": {
			price:        sdk.MustNewDecFromStr("1234568000"),
			tickExpected: 81234568,
		},
		"inverse testing with 1": {
			price:        sdk.OneDec(),
			tickExpected: 0,
		},
	}
	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			tickSpacing := uint64(1)

			// 1. Compute tick from price.
			tickFromPrice, err := math.PriceToTickRoundDown(tc.price, tickSpacing)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.tickExpected, tickFromPrice)

			// 2. Compute price from tick (inverse price)
			price, err := math.TickToPrice(tickFromPrice)
			suite.Require().NoError(err)

			// Make sure inverse price is correct.
			expectedPrice := tc.price
			if !tc.truncatedPrice.IsNil() {
				expectedPrice = tc.truncatedPrice
			}
			suite.Require().Equal(expectedPrice, price)

			// 3. Compute tick from inverse price (inverse tick)
			inverseTickFromPrice, err := math.PriceToTickRoundDown(price, tickSpacing)
			suite.Require().NoError(err)

			// Make sure original tick and inverse tick match.
			suite.Require().Equal(tickFromPrice, inverseTickFromPrice)

			// 4. Validate PriceToTick and TickToSqrtPrice functions
			_, sqrtPrice, err := math.TickToSqrtPrice(tickFromPrice)
			suite.Require().NoError(err)

			priceFromSqrtPrice := sqrtPrice.Power(2)

			// TODO: investigate this separately
			// https://github.com/osmosis-labs/osmosis/issues/4925
			// suite.Require().Equal(expectedPrice.String(), priceFromSqrtPrice.String())

			// 5. Compute tick from sqrt price from the original tick.
			inverseTickFromSqrtPrice, err := math.PriceToTickRoundDown(priceFromSqrtPrice, tickSpacing)
			suite.Require().NoError(err)

			suite.Require().Equal(tickFromPrice, inverseTickFromSqrtPrice, "expected: %s, actual: %s", tickFromPrice, inverseTickFromSqrtPrice)
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestPriceToTick_ErrorCases() {
	testCases := map[string]struct {
		price sdk.Dec
	}{
		"use negative price": {
			price: sdk.OneDec().Neg(),
		},
		"price is greater than max spot price": {
			price: types.MaxSpotPrice.Add(sdk.OneDec()),
		},
		"price is less than min spot price": {
			price: types.MinSpotPrice.Sub(sdk.OneDec()),
		},
	}
	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			tickFromPrice, err := math.PriceToTick(tc.price)
			suite.Require().Error(err)
			suite.Require().Equal(tickFromPrice, int64(0))
		})
	}
}
func (suite *ConcentratedMathTestSuite) TestTickToPrice_ErrorCases() {
	testCases := map[string]struct {
		tickIndex int64
	}{
		"tick index is greater than max tick": {
			tickIndex: types.MaxTick + 1,
		},
		"tick index is less than min tick": {
			tickIndex: types.MinTick - 1,
		},
	}
	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			_, err := math.TickToPrice(tc.tickIndex)
			suite.Require().Error(err)
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestCalculatePriceToTick() {
	testCases := map[string]struct {
		price             sdk.Dec
		expectedTickIndex int64
	}{
		"Price greater than 1": {
			price:             sdk.MustNewDecFromStr("9.78"),
			expectedTickIndex: 8780000,
		},
		"Price less than 1": {
			price:             sdk.MustNewDecFromStr("0.71"),
			expectedTickIndex: -2900000,
		},
		"100_000_000 -> 72000000": {
			price:             sdk.NewDec(100_000_000),
			expectedTickIndex: 72000000,
		},
		"100_000_050 -> 72000000": {
			price:             sdk.NewDec(100_000_050),
			expectedTickIndex: 72000000,
		},
		"100_000_051 -> 72000001": {
			price:             sdk.NewDec(100_000_051),
			expectedTickIndex: 72000001,
		},
		"100_000_100 -> 72000001": {
			price:             sdk.NewDec(100_000_100),
			expectedTickIndex: 72000001,
		},
	}
	for name, t := range testCases {
		suite.Run(name, func() {
			tickIndex := math.CalculatePriceToTick(t.price)
			suite.Require().Equal(t.expectedTickIndex, tickIndex)
		})
	}
}
