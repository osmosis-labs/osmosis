package math_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
)

func (suite *ConcentratedMathTestSuite) TestTickToPrice() {
	testCases := map[string]struct {
		tickIndex     sdk.Int
		kAtPriceOne   sdk.Int
		expectedPrice string
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
			expectedPrice: "0.000008890000000000",
		},
		"More variety of numbers in each place": {
			tickIndex:     sdk.NewInt(4030301),
			kAtPriceOne:   sdk.NewInt(-5),
			expectedPrice: "53030.100000000000000000",
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			sqrtPrice, err := math.TickToPrice(tc.tickIndex, tc.kAtPriceOne)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedPrice, sqrtPrice.String())

		})
	}
}

func (suite *ConcentratedMathTestSuite) TestPriceToTick() {
	testCases := []struct {
		name         string
		price        sdk.Dec
		kAtPriceOne  sdk.Int
		tickExpected string
	}{
		{
			"50000 to tick with -4 k at price one",
			sdk.NewDec(50000),
			sdk.NewInt(-4),
			"400000",
		},
		{
			"5.01 to tick with -2 k at price one",
			sdk.MustNewDecFromStr("5.010000000000000000"),
			sdk.NewInt(-2),
			"401",
		},
		{
			"50000.01 to tick with -6 k at price one",
			sdk.MustNewDecFromStr("50000.010000000000000000"),
			sdk.NewInt(-6),
			"40000001",
		},
		{
			"0.00000889 to tick with -8 k at price one",
			sdk.MustNewDecFromStr("0.000008890000000000"),
			sdk.NewInt(-8),
			"-99999111",
		},
		{
			"0.9998 to tick with -4 k at price one",
			sdk.MustNewDecFromStr("0.999800000000000000"),
			sdk.NewInt(-4),
			"-2",
		},
		{
			"53030.10 to tick with -5 k at price one",
			sdk.MustNewDecFromStr("53030.100000000000000000"),
			sdk.NewInt(-5),
			"4030301",
		},
	}
	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			tick, err := math.PriceToTick(tc.price, tc.kAtPriceOne)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.tickExpected, tick.String())
		})
	}
}
