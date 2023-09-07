package math_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
)

const (
	defaultTickSpacing = 100
)

var (
	// spot price - (10^(spot price exponent - 6 - 1))
	// Note we get spot price exponent by counting the number of digits in the max spot price and subtracting 1.
	closestPriceBelowMaxPriceDefaultTickSpacing = types.MaxSpotPrice.Sub(sdk.NewDec(10).PowerMut(uint64(len(types.MaxSpotPrice.TruncateInt().String()) - 1 - int(-types.ExponentAtPriceOne) - 1)))
	// min tick + 10 ^ -expoentAtPriceOne
	closestTickAboveMinPriceDefaultTickSpacing = osmomath.NewInt(types.MinInitializedTick).Add(osmomath.NewInt(10).ToDec().PowerInteger(uint64(types.ExponentAtPriceOne * -1)).TruncateInt())

	smallestBigDec = osmomath.SmallestDec()
	bigOneDec      = osmomath.OneDec()
	bigTenDec      = osmomath.NewBigDec(10)
)

// testing helper for price to tick round down spacing,
// state machine only implements sqrt price to tick round dow spacing.
func PriceToTickRoundDownSpacing(price osmomath.BigDec, tickSpacing uint64) (int64, error) {
	tickIndex, err := math.CalculatePriceToTick(price)
	if err != nil {
		return 0, err
	}

	tickIndex, err = math.RoundDownTickToSpacing(tickIndex, int64(tickSpacing))
	if err != nil {
		return 0, err
	}

	return tickIndex, nil
}

// use following equations to test testing vectors using sage
// geometricExponentIncrementDistanceInTicks(exponentAtPriceOne) = (9 * (10^(-exponentAtPriceOne)))
// geometricExponentDelta(tickIndex, exponentAtPriceOne)  = floor(tickIndex / geometricExponentIncrementDistanceInTicks(exponentAtPriceOne))
// exponentAtCurrentTick(tickIndex, exponentAtPriceOne) = exponentAtPriceOne + geometricExponentDelta(tickIndex, exponentAtPriceOne)
// currentAdditiveIncrementInTicks(tickIndex, exponentAtPriceOne) = pow(10, exponentAtCurrentTick(tickIndex, exponentAtPriceOne))
// numAdditiveTicks(tickIndex, exponentAtPriceOne) = tickIndex - (geometricExponentDelta(tickIndex, exponentAtPriceOne) * geometricExponentIncrementDistanceInTicks(exponentAtPriceOne)
// price(tickIndex, exponentAtPriceOne) = pow(10, geometricExponentDelta(tickIndex, exponentAtPriceOne)) +
// (numAdditiveTicks(tickIndex, exponentAtPriceOne) * currentAdditiveIncrementInTicks(tickIndex, exponentAtPriceOne))
func TestTickToSqrtPrice(t *testing.T) {
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
			tickIndex:     types.MinInitializedTick,
			expectedPrice: types.MinSpotPrice,
		},
		"error: tickIndex less than minimum": {
			tickIndex:     types.MinCurrentTick - 1,
			expectedError: types.TickIndexMinimumError{MinTick: types.MinCurrentTick},
		},
		"error: tickIndex greater than maximum": {
			tickIndex:     types.MaxTick + 1,
			expectedError: types.TickIndexMaximumError{MaxTick: types.MaxTick},
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
		t.Run(name, func(t *testing.T) {
			price, sqrtPrice, err := math.TickToSqrtPrice(tc.tickIndex)
			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			expectedSqrtPrice, err := osmomath.MonotonicSqrt(tc.expectedPrice)
			require.NoError(t, err)

			require.Equal(t, osmomath.BigDecFromSDKDec(tc.expectedPrice).String(), price.String())
			require.Equal(t, osmomath.BigDecFromSDKDec(expectedSqrtPrice).String(), sqrtPrice.String())
		})
	}
}

func TestTicksToSqrtPrice(t *testing.T) {
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
			lowerTickIndex:     sdk.NewInt(types.MinInitializedTick),
			upperTickIndex:     sdk.NewInt(types.MaxTick),
			expectedUpperPrice: types.MaxSpotPrice,
			expectedLowerPrice: types.MinSpotPrice,
		},
		"error: lowerTickIndex less than minimum": {
			lowerTickIndex: sdk.NewInt(types.MinCurrentTick - 1),
			upperTickIndex: sdk.NewInt(36073300),
			expectedError:  types.TickIndexMinimumError{MinTick: types.MinCurrentTick},
		},
		"error: upperTickIndex greater than maximum": {
			lowerTickIndex: sdk.NewInt(types.MinInitializedTick),
			upperTickIndex: sdk.NewInt(types.MaxTick + 1),
			expectedError:  types.TickIndexMaximumError{MaxTick: types.MaxTick},
		},
		"error: provided lower tick and upper tick are same": {
			lowerTickIndex: sdk.NewInt(types.MinInitializedTick),
			upperTickIndex: sdk.NewInt(types.MinInitializedTick),
			expectedError:  types.InvalidLowerUpperTickError{LowerTick: sdk.NewInt(types.MinInitializedTick).Int64(), UpperTick: sdk.NewInt(types.MinInitializedTick).Int64()},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			lowerSqrtPrice, upperSqrtPrice, err := math.TicksToSqrtPrice(tc.lowerTickIndex.Int64(), tc.upperTickIndex.Int64())
			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
				return
			}
			require.NoError(t, err)

			// convert test case's prices to sqrt price
			expectedLowerSqrtPrice, err := osmomath.MonotonicSqrt(tc.expectedLowerPrice)
			require.NoError(t, err)
			expectedUpperSqrtPrice, err := osmomath.MonotonicSqrt(tc.expectedUpperPrice)
			require.NoError(t, err)

			require.Equal(t, osmomath.BigDecFromSDKDec(expectedLowerSqrtPrice).String(), lowerSqrtPrice.String())
			require.Equal(t, osmomath.BigDecFromSDKDec(expectedUpperSqrtPrice).String(), upperSqrtPrice.String())
		})
	}
}

func TestPriceToTick(t *testing.T) {
	const (
		one = uint64(1)
	)

	testCases := map[string]struct {
		price         osmomath.BigDec
		tickExpected  int64
		expectedError error
	}{
		"BTC <> USD, tick 38035200 -> price 30352": {
			price:        osmomath.MustNewDecFromStr("30352"),
			tickExpected: 38035200,
		},
		"BTC <> USD, tick 38035300 + 100 -> price 30353": {
			price:        osmomath.MustNewDecFromStr("30353"),
			tickExpected: 38035300,
		},
		"SHIB <> USD, tick -44821000 -> price 0.000011790": {
			price:        osmomath.MustNewDecFromStr("0.000011790"),
			tickExpected: -44821000,
		},
		"SHIB <> USD, tick -44820900 -> price 0.000011791": {
			price:        osmomath.MustNewDecFromStr("0.000011791"),
			tickExpected: -44820900,
		},
		"ETH <> BTC, tick -12104000 -> price 0.068960": {
			price:        osmomath.MustNewDecFromStr("0.068960"),
			tickExpected: -12104000,
		},
		"ETH <> BTC, tick -12104000 + 100 -> price 0.068961": {
			price:        osmomath.MustNewDecFromStr("0.068961"),
			tickExpected: -12103900,
		},
		"max sqrt price -1, max neg tick six - 100 -> max tick neg six - 100": {
			price:        osmomath.MustNewDecFromStr("99999000000000000000000000000000000000"),
			tickExpected: types.MaxTick - 100,
		},
		"max sqrt price, max tick neg six -> max spot price": {
			price:        osmomath.BigDecFromSDKDec(types.MaxSqrtPrice.Power(2)),
			tickExpected: types.MaxTick,
		},
		"Gyen <> USD, tick -20594000 -> price 0.0074060": {
			price:        osmomath.MustNewDecFromStr("0.007406"),
			tickExpected: -20594000,
		},
		"Gyen <> USD, tick -20594000 + 100 -> price 0.0074061": {
			price:        osmomath.MustNewDecFromStr("0.0074061"),
			tickExpected: -20593900,
		},
		"Spell <> USD, tick -29204000 -> price 0.00077960": {
			price:        osmomath.MustNewDecFromStr("0.0007796"),
			tickExpected: -29204000,
		},
		"Spell <> USD, tick -29204000 + 100 -> price 0.00077961": {
			price:        osmomath.MustNewDecFromStr("0.00077961"),
			tickExpected: -29203900,
		},
		"Atom <> Osmo, tick -12150000 -> price 0.068500": {
			price:        osmomath.MustNewDecFromStr("0.0685"),
			tickExpected: -12150000,
		},
		"Atom <> Osmo, tick -12150000 + 100 -> price 0.068501": {
			price:        osmomath.MustNewDecFromStr("0.068501"),
			tickExpected: -12149900,
		},
		"Boot <> Osmo, tick 64576000 -> price 25760000": {
			price:        osmomath.MustNewDecFromStr("25760000"),
			tickExpected: 64576000,
		},
		"Boot <> Osmo, tick 64576000 + 100 -> price 25761000": {
			price:        osmomath.MustNewDecFromStr("25761000"),
			tickExpected: 64576100,
		},
		"price is one Dec": {
			price:        osmomath.OneDec(),
			tickExpected: 0,
		},
		"price is negative decimal": {
			price:         osmomath.OneDec().Neg(),
			expectedError: fmt.Errorf("price must be greater than zero"),
		},
		"price is greater than max spot price": {
			price:         osmomath.BigDecFromSDKDec(types.MaxSpotPrice.Add(sdk.OneDec())),
			expectedError: types.PriceBoundError{ProvidedPrice: osmomath.BigDecFromSDKDec(types.MaxSpotPrice.Add(sdk.OneDec())), MinSpotPrice: types.MinSpotPriceV2, MaxSpotPrice: types.MaxSpotPrice},
		},
		"price is smaller than min spot price": {
			price:         types.MinSpotPriceV2.Quo(bigTenDec),
			expectedError: types.PriceBoundError{ProvidedPrice: types.MinSpotPriceV2.Quo(bigTenDec), MinSpotPrice: types.MinSpotPriceV2, MaxSpotPrice: types.MaxSpotPrice},
		},
	}
	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			// surpress error here, we only listen to errors from system under test.
			tick, _ := math.CalculatePriceToTick(tc.price)

			// With tick spacing of one, no rounding should occur.
			tickRoundDown, err := PriceToTickRoundDownSpacing(tc.price, one)
			if tc.expectedError != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectedError.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.tickExpected, tick)
			require.Equal(t, tc.tickExpected, tickRoundDown)
		})
	}
}

func TestPriceToTickRoundDown(t *testing.T) {
	testCases := map[string]struct {
		price        osmomath.BigDec
		tickSpacing  uint64
		tickExpected int64
	}{
		"tick spacing 100, price of 1": {
			price:        osmomath.OneDec(),
			tickSpacing:  defaultTickSpacing,
			tickExpected: 0,
		},
		"tick spacing 100, price of 1.000030, tick 30 -> 0": {
			price:        osmomath.MustNewDecFromStr("1.000030"),
			tickSpacing:  defaultTickSpacing,
			tickExpected: 0,
		},
		"tick spacing 100, price of 0.9999970, tick -30 -> -100": {
			price:        osmomath.MustNewDecFromStr("0.9999970"),
			tickSpacing:  defaultTickSpacing,
			tickExpected: -100,
		},
		"tick spacing 50, price of 0.9999730, tick -270 -> -300": {
			price:        osmomath.MustNewDecFromStr("0.9999730"),
			tickSpacing:  50,
			tickExpected: -300,
		},
		"tick spacing 100, MinSpotPrice, MinTick": {
			price:        types.MinSpotPriceBigDec,
			tickSpacing:  defaultTickSpacing,
			tickExpected: types.MinInitializedTick,
		},
		"tick spacing 100, Spot price one tick above min, one tick above min -> MinTick": {
			price:       osmomath.BigDecFromSDKDec(types.MinSpotPrice.Add(sdk.SmallestDec())),
			tickSpacing: defaultTickSpacing,
			// Since the tick should always be the closest tick below (and `smallestDec` isn't sufficient
			// to push us into the next tick), we expect MinTick to be returned here.
			tickExpected: types.MinInitializedTick,
		},
		"tick spacing 100, Spot price one tick below max, one tick below max -> MaxTick - 1": {
			price:        osmomath.BigDecFromSDKDec(closestPriceBelowMaxPriceDefaultTickSpacing),
			tickSpacing:  defaultTickSpacing,
			tickExpected: types.MaxTick - 100,
		},
		"tick spacing 100, Spot price 100_000_050 -> 72000000": {
			price:        osmomath.NewBigDec(100_000_050),
			tickSpacing:  defaultTickSpacing,
			tickExpected: 72000000,
		},
		"tick spacing 100, Spot price 100_000_051 -> 72000100 (rounded up to tick spacing)": {
			price:        osmomath.NewBigDec(100_000_051),
			tickSpacing:  defaultTickSpacing,
			tickExpected: 72000000,
		},
		"tick spacing 1, Spot price 100_000_051 -> 72000000 no tick spacing rounding": {
			price:        osmomath.NewBigDec(100_000_051),
			tickSpacing:  1,
			tickExpected: 72000000,
		},
		"tick spacing 1, Spot price 100_000_101 -> 72000001 no tick spacing rounding": {
			price:        osmomath.NewBigDec(100_000_101),
			tickSpacing:  1,
			tickExpected: 72000001,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tick, err := PriceToTickRoundDownSpacing(tc.price, tc.tickSpacing)

			require.NoError(t, err)
			require.Equal(t, tc.tickExpected, tick)
		})
	}
}

// TestTickToSqrtPricePriceToTick_InverseRelationship tests that ensuring the inverse calculation
// between the following methods:
// 1) price -> tick, tick -> price yields expected
// 2) tick -> sqrt price, sqrt price -> tick yields expected
// TODO: Revisit this test, under the lens of bucket index.
func TestTickToSqrtPricePriceToTick_InverseRelationship(t *testing.T) {
	type testcase struct {
		price          osmomath.BigDec
		truncatedPrice osmomath.BigDec
		tickExpected   int64
	}
	testCases := map[string]testcase{
		"50000 to tick": {
			price:        osmomath.MustNewDecFromStr("50000"),
			tickExpected: 40000000,
		},
		"5.01 to tick": {
			price:        osmomath.MustNewDecFromStr("5.01"),
			tickExpected: 4010000,
		},
		"50000.01 to tick": {
			price:        osmomath.MustNewDecFromStr("50000.01"),
			tickExpected: 40000001,
		},
		"0.090001 to tick": {
			price:        osmomath.MustNewDecFromStr("0.090001"),
			tickExpected: -9999900,
		},
		"0.9998 to tick": {
			price:        osmomath.MustNewDecFromStr("0.9998"),
			tickExpected: -2000,
		},
		"53030 to tick": {
			price:        osmomath.MustNewDecFromStr("53030"),
			tickExpected: 40303000,
		},
		"max spot price": {
			price:        osmomath.BigDecFromSDKDec(types.MaxSpotPrice),
			tickExpected: types.MaxTick,
		},
		"max spot price - smallest price delta given exponent at price one of -6": {
			// 37 - 6 is calculated by counting the exponent of max spot price and subtracting exponent at price one
			price:        osmomath.BigDecFromSDKDec(types.MaxSpotPrice.Sub(sdk.NewDec(10).PowerMut(37 - 6))),
			tickExpected: types.MaxTick - 1, // still max
		},
		"min spot price": {
			price:        types.MinSpotPriceBigDec,
			tickExpected: types.MinInitializedTick,
		},
		"smallest + min price + tick": {
			price:        osmomath.MustNewDecFromStr("0.000000000001000001"),
			tickExpected: types.MinInitializedTick + 1,
		},
		"at price level of 0.01 - odd": {
			price:        osmomath.MustNewDecFromStr("0.012345670000000000"),
			tickExpected: -17765433,
		},
		"at price level of 0.01 - even": {
			price:        osmomath.MustNewDecFromStr("0.01234568000000000"),
			tickExpected: -17765432,
		},
		"at min price level of 0.01 - odd": {
			price:        osmomath.MustNewDecFromStr("0.000000000001234567"),
			tickExpected: -107765433,
		},
		"at min price level of 0.01 - even": {
			price:        osmomath.MustNewDecFromStr("0.000000000001234568"),
			tickExpected: -107765432,
		},
		"at price level of 1_000_000_000 - odd end": {
			price:        osmomath.MustNewDecFromStr("1234567000"),
			tickExpected: 81234567,
		},
		"at price level of 1_000_000_000 - in-between supported": {
			price:          osmomath.MustNewDecFromStr("1234567500"),
			tickExpected:   81234567,
			truncatedPrice: osmomath.MustNewDecFromStr("1234567000"),
		},
		"at price level of 1_000_000_000 - even end": {
			price:        osmomath.MustNewDecFromStr("1234568000"),
			tickExpected: 81234568,
		},
		"inverse testing with 1": {
			price:        osmomath.OneDec(),
			tickExpected: 0,
		},
	}
	var powTen int64 = 10
	for i := 1; i < 13; i++ {
		testCases[fmt.Sprintf("min spot price * 10^%d", i)] = testcase{
			price:        osmomath.BigDecFromSDKDec(types.MinSpotPrice.MulInt64(powTen)),
			tickExpected: types.MinInitializedTick + (int64(i) * 9e6),
		}
		powTen *= 10
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// 1. Compute tick from price.
			tickFromPrice, err := math.CalculatePriceToTick(tc.price)
			require.NoError(t, err)
			require.Equal(t, tc.tickExpected, tickFromPrice)

			// 2. Compute price from tick (inverse price)
			price, err := math.TickToPrice(tickFromPrice)
			require.NoError(t, err)

			// Make sure inverse price is correct.
			expectedPrice := tc.price
			if !tc.truncatedPrice.IsNil() {
				expectedPrice = tc.truncatedPrice
			}
			require.Equal(t, expectedPrice, price)

			// 3. Compute tick from inverse price (inverse tick)
			inverseTickFromPrice, err := math.CalculatePriceToTick(price)
			require.NoError(t, err)

			// Make sure original tick and inverse tick match.
			require.Equal(t, tickFromPrice, inverseTickFromPrice)

			// 4. Validate PriceToTick and TickToSqrtPrice functions
			_, sqrtPrice, err := math.TickToSqrtPrice(tickFromPrice)
			require.NoError(t, err)

			// TODO: investigate this separately
			// https://github.com/osmosis-labs/osmosis/issues/4925
			// require.Equal(t, expectedPrice.String(), priceFromSqrtPrice.String())

			// 5. Compute tick from sqrt price from the original tick.
			inverseTickFromSqrtPrice, err := math.CalculateSqrtPriceToTick(sqrtPrice)
			require.NoError(t, err)

			require.Equal(t, tickFromPrice, inverseTickFromSqrtPrice, "expected: %s, actual: %s", tickFromPrice, inverseTickFromSqrtPrice)
		})
	}
}

func TestPriceToTick_ErrorCases(t *testing.T) {
	testCases := map[string]struct {
		price osmomath.BigDec
	}{
		"use negative price": {
			price: osmomath.OneDec().Neg(),
		},
		"price is greater than max spot price": {
			price: osmomath.BigDecFromSDKDec(types.MaxSpotPrice.Add(sdk.OneDec())),
		},
		"price is less than min spot price": {
			price: osmomath.BigDecFromSDKDec(types.MinSpotPrice.Sub(sdk.OneDec())),
		},
	}
	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			tickFromPrice, err := math.CalculatePriceToTick(tc.price)
			require.Error(t, err)
			require.Equal(t, tickFromPrice, int64(0))
		})
	}
}
func TestTickToPrice_ErrorCases(t *testing.T) {
	testCases := map[string]struct {
		tickIndex int64
	}{
		"tick index is greater than max tick": {
			tickIndex: types.MaxTick + 1,
		},
		"tick index is less than min tick": {
			tickIndex: types.MinCurrentTick - 1,
		},
	}
	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			_, err := math.TickToPrice(tc.tickIndex)
			require.Error(t, err)
		})
	}
}

func TestCalculatePriceToTick(t *testing.T) {
	testCases := map[string]struct {
		price             osmomath.BigDec
		expectedTickIndex int64
	}{
		"Price greater than 1": {
			price:             osmomath.MustNewDecFromStr("9.78"),
			expectedTickIndex: 8780000,
		},
		"Price less than 1": {
			price:             osmomath.MustNewDecFromStr("0.71"),
			expectedTickIndex: -2900000,
		},
		"100_000_000 -> 72000000": {
			price:             osmomath.NewBigDec(100_000_000),
			expectedTickIndex: 72000000,
		},
		"100_000_050 -> 72000000": {
			price:             osmomath.NewBigDec(100_000_050),
			expectedTickIndex: 72000000,
		},
		"100_000_051 -> 72000000": {
			price:             osmomath.NewBigDec(100_000_051),
			expectedTickIndex: 72000000,
		},
		"100_000_100 -> 72000001": {
			price:             osmomath.NewBigDec(100_000_100),
			expectedTickIndex: 72000001,
		},
		"MinSpotPrice V1 -> MinInitializedTick": {
			price:             types.MinSpotPriceBigDec,
			expectedTickIndex: types.MinInitializedTick,
		},
		"MinSpotPrice V1 - 10^-19 -> MinCurrentTick": {
			price:             types.MinSpotPriceBigDec.Sub(osmomath.NewDecWithPrec(1, 19)),
			expectedTickIndex: types.MinCurrentTick,
		},
		"MinSpotPrice V2 -> MinInitializedTick V2": {
			price:             types.MinSpotPriceV2,
			expectedTickIndex: types.MinInitializedTickV2,
		},
		"between MinSpotPrice V2 + 1 ULP -> MinInitializedTick V2 + 1": {
			price:             types.MinSpotPriceV2.Add(smallestBigDec),
			expectedTickIndex: types.MinInitializedTickV2 + 1,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tickIndex, err := math.CalculatePriceToTick(tc.price)
			require.NoError(t, err)
			require.Equal(t, tc.expectedTickIndex, tickIndex)

			// Only run tests on the BigDec version on range [MinCurrentTickV2, MinCurrentTick]
			if tc.price.LT(types.MinSpotPriceBigDec) {
				return
			}

			tickIndex, err = math.CalculatePriceToTick(tc.price)
			require.NoError(t, err)
			require.Equal(t, tc.expectedTickIndex, tickIndex)
		})
	}
}

func TestPowTenInternal(t *testing.T) {
	testCases := map[string]struct {
		exponent             int64
		expectedPowTenResult sdk.Dec
	}{
		"Power by 5": {
			exponent:             5,
			expectedPowTenResult: sdk.NewDec(100000),
		},
		"Power by 0": {
			exponent:             0,
			expectedPowTenResult: sdk.NewDec(1),
		},
		"Power by -5": {
			exponent:             -5,
			expectedPowTenResult: sdk.MustNewDecFromStr("0.00001"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			powTenResult := math.PowTenInternal(tc.exponent)
			require.Equal(t, tc.expectedPowTenResult, powTenResult)
		})
	}
}

func TestSqrtPriceToTickRoundDownSpacing(t *testing.T) {
	sdkULP := osmomath.BigDecFromSDKDec(sdk.SmallestDec())

	// Compute reference values that need to be satisfied
	_, sqp1, err := math.TickToSqrtPrice(1)
	require.NoError(t, err)
	_, sqp99, err := math.TickToSqrtPrice(99)
	require.NoError(t, err)
	_, sqp100, err := math.TickToSqrtPrice(100)
	require.NoError(t, err)
	_, sqpn100, err := math.TickToSqrtPrice(-100)
	require.NoError(t, err)
	_, sqpn101, err := math.TickToSqrtPrice(-101)
	require.NoError(t, err)
	_, sqpMaxTickSubOne, err := math.TickToSqrtPrice(types.MaxTick - 1)
	require.NoError(t, err)
	_, sqpMinTickPlusOne, err := math.TickToSqrtPrice(types.MinInitializedTick + 1)
	require.NoError(t, err)
	_, sqpMinTickPlusTwo, err := math.TickToSqrtPrice(types.MinInitializedTick + 2)
	require.NoError(t, err)

	testCases := map[string]struct {
		sqrtPrice    osmomath.BigDec
		tickSpacing  uint64
		tickExpected int64
	}{
		"sqrt price of 1 (tick spacing 1)": {
			sqrtPrice:    osmomath.OneDec(),
			tickSpacing:  1,
			tickExpected: 0,
		},
		"sqrt price exactly on boundary of next tick (tick spacing 1)": {
			sqrtPrice:    sqp1,
			tickSpacing:  1,
			tickExpected: 1,
		},
		"sqrt price one ULP below boundary of next tick (tick spacing 1)": {
			sqrtPrice:    sqp1.Sub(sdkULP),
			tickSpacing:  1,
			tickExpected: 0,
		},
		"sqrt price corresponding to bucket 99 (tick spacing 100)": {
			sqrtPrice:    sqp99,
			tickSpacing:  defaultTickSpacing,
			tickExpected: 0,
		},
		"sqrt price exactly on bucket 100 (tick spacing 100)": {
			sqrtPrice:    sqp100,
			tickSpacing:  defaultTickSpacing,
			tickExpected: 100,
		},
		"sqrt price one ULP below bucket 100 (tick spacing 100)": {
			sqrtPrice:    sqp100.Sub(sdkULP),
			tickSpacing:  defaultTickSpacing,
			tickExpected: 0,
		},
		"sqrt price exactly on bucket -100 (tick spacing 100)": {
			sqrtPrice:    sqpn100,
			tickSpacing:  defaultTickSpacing,
			tickExpected: -100,
		},
		"sqrt price one ULP below bucket -100 (tick spacing 100)": {
			sqrtPrice:    sqpn100.Sub(sdkULP),
			tickSpacing:  defaultTickSpacing,
			tickExpected: -200,
		},
		"sqrt price exactly on tick -101 (tick spacing 100)": {
			sqrtPrice:    sqpn101,
			tickSpacing:  defaultTickSpacing,
			tickExpected: -200,
		},
		"sqrt price exactly equal to max sqrt price": {
			sqrtPrice:    osmomath.BigDecFromSDKDec(types.MaxSqrtPrice),
			tickSpacing:  defaultTickSpacing,
			tickExpected: types.MaxTick,
		},
		"sqrt price exactly equal to min sqrt price": {
			sqrtPrice:    osmomath.BigDecFromSDKDec(types.MinSqrtPrice),
			tickSpacing:  defaultTickSpacing,
			tickExpected: types.MinInitializedTick,
		},
		"sqrt price equal to max sqrt price minus one ULP": {
			sqrtPrice:    osmomath.BigDecFromSDKDec(types.MaxSqrtPrice).Sub(sdkULP),
			tickSpacing:  defaultTickSpacing,
			tickExpected: types.MaxTick - defaultTickSpacing,
		},
		"sqrt price corresponds exactly to max tick - 1 (tick spacing 1)": {
			sqrtPrice:    sqpMaxTickSubOne,
			tickSpacing:  1,
			tickExpected: types.MaxTick - 1,
		},
		"sqrt price one ULP below max tick - 1 (tick spacing 1)": {
			sqrtPrice:    sqpMaxTickSubOne.Sub(sdkULP),
			tickSpacing:  1,
			tickExpected: types.MaxTick - 2,
		},
		"sqrt price one ULP below max tick - 1 (tick spacing 100)": {
			sqrtPrice:    sqpMaxTickSubOne.Sub(sdkULP),
			tickSpacing:  defaultTickSpacing,
			tickExpected: types.MaxTick - defaultTickSpacing,
		},
		"sqrt price corresponds exactly to min tick + 1 (tick spacing 1)": {
			sqrtPrice:    sqpMinTickPlusOne,
			tickSpacing:  1,
			tickExpected: types.MinInitializedTick + 1,
		},
		"sqrt price corresponds exactly to min tick + 1 minus 1 ULP (tick spacing 1)": {
			// Calculated using TickToSqrtPrice(types.MinInitializedTick + 1) - 1 ULP
			sqrtPrice:    sqpMinTickPlusOne.Sub(sdkULP),
			tickSpacing:  1,
			tickExpected: types.MinInitializedTick,
		},
		"sqrt price corresponds exactly to min tick + 1 plus 1 ULP (tick spacing 1)": {
			// Calculated using TickToSqrtPrice(types.MinInitializedTick + 1) + 1 ULP
			sqrtPrice:    sqpMinTickPlusOne.Add(sdkULP),
			tickSpacing:  1,
			tickExpected: types.MinInitializedTick + 1,
		},
		"sqrt price corresponds exactly to min tick + 2 (tick spacing 1)": {
			sqrtPrice:    sqpMinTickPlusTwo,
			tickSpacing:  1,
			tickExpected: types.MinInitializedTick + 2,
		},
		"sqrt price corresponds exactly to min tick + 2 plus 1 ULP (tick spacing 1)": {
			// Calculated using TickToSqrtPrice(types.MinInitializedTick + 2) + 1 ULP
			sqrtPrice:    sqpMinTickPlusTwo.Add(sdkULP),
			tickSpacing:  1,
			tickExpected: types.MinInitializedTick + 2,
		},
		"sqrt price corresponds exactly to min tick + 2 minus 1 ULP (tick spacing 1)": {
			// Calculated using TickToSqrtPrice(types.MinInitializedTick + 2) - 1 ULP
			sqrtPrice:    sqpMinTickPlusTwo.Sub(sdkULP),
			tickSpacing:  1,
			tickExpected: types.MinInitializedTick + 1,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tickIndex, err := math.SqrtPriceToTickRoundDownSpacing(tc.sqrtPrice, tc.tickSpacing)
			require.NoError(t, err)
			require.Equal(t, tc.tickExpected, tickIndex)

			// Ensure returned bucket properly encapsulates given sqrt price, skipping the upper bound
			// check if we're on the max tick
			_, inverseSqrtPrice, err := math.TickToSqrtPrice(tickIndex)
			require.NoError(t, err)
			require.True(t, inverseSqrtPrice.LTE(tc.sqrtPrice))

			if tc.tickExpected != types.MaxTick {
				_, inverseSqrtPriceTickAbove, err := math.TickToSqrtPrice(tickIndex + int64(tc.tickSpacing))
				require.NoError(t, err)
				require.True(t, inverseSqrtPriceTickAbove.GT(tc.sqrtPrice))
			}
		})
	}
}

// Computes sqrt price to tick near the min spot price V1 bound (10^-12)
// This case is important because it helped catch non-monotonicity when
// BigDec price with Dec sqrt function was used.
// To work around this issue, the price is truncated to 18 decimals
// in the original price range of [10^-12, 10^38] and 18 decimal TickToSqrt is used,
// helping maintain backwards compatibility.
//
// In the future, for price range [10^-30, 10^-12), we will use non-truncated BigDec
// with 36 decimal TickToSqrt function.
func TestCalculateSqrtPriceToTick_NearMinSpotPriceV1Bound(t *testing.T) {
	sqrtPrice := osmomath.MustNewDecFromStr("0.000001000049998750999999999999999999")
	_, err := math.CalculateSqrtPriceToTick(sqrtPrice)
	require.NoError(t, err)
}
