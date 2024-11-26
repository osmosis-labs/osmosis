package concentrated_liquidity_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	clmath "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var _ = suite.TestingSuite(nil)

// positionMeta defines the metadata of a position
// after its creation.
type positionMeta struct {
	positionId uint64
	lowerTick  int64
	upperTick  int64
	liquidity  osmomath.Dec
}

type secondPosition = apptesting.SecondConcentratedPosition

var (
	// Allow 0.01% margin of error.
	multiplicativeTolerance = osmomath.ErrTolerance{
		MultiplicativeTolerance: osmomath.MustNewDecFromStr("0.0001"),
	}
	// Allow additive margin equal of 1. It may stem from round up token in and rounding down
	// tokenOut in favor of the pool.
	oneAdditiveTolerance = osmomath.ErrTolerance{
		AdditiveTolerance: osmomath.OneDec(),
	}

	swapOutGivenInCases = map[string]apptesting.ConcentratedSwapTest{
		//  One price range
		//
		//          5000
		//  4545 -----|----- 5500
		"single position within one tick: usdc -> eth": {
			TokenIn:       sdk.NewCoin("usdc", osmomath.NewInt(42000000)),
			TokenOutDenom: "eth",
			PriceLimit:    osmomath.NewBigDec(5004),
			SpreadFactor:  osmomath.ZeroDec(),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// from math import *
			// from decimal import *

			// token_in = Decimal("42000000")
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000

			// rounding_direction = ROUND_DOWN # round delta down since we're swapping asset 1 in
			// sqrt_delta = (token_in / liq).quantize(precision, rounding=rounding_direction)
			// sqrt_next = sqrt_cur + sqrt_delta

			// token_out = floor(liq * (sqrt_next - sqrt_cur) / (sqrt_next * sqrt_cur))
			// token_in = ceil(liq * abs(sqrt_cur - sqrt_next))

			// print(sqrt_next)
			// print(token_in)
			// print(token_out)
			ExpectedTokenIn:  sdk.NewCoin("usdc", osmomath.NewInt(42000000)),
			ExpectedTokenOut: sdk.NewCoin("eth", osmomath.NewInt(8396)),
			ExpectedTick:     31003913,
			// Corresponds to sqrt_next in script above
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.738348247484497718514800000003600626"),
			// tick's accum coins stay same since crossing tick does not occur in this case
		},
		"single position within one tick: usdc -> eth, with zero price limit": {
			TokenIn:       sdk.NewCoin("usdc", osmomath.NewInt(42000000)),
			TokenOutDenom: "eth",
			PriceLimit:    osmomath.ZeroBigDec(),
			SpreadFactor:  osmomath.ZeroDec(),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// from math import *
			// from decimal import *

			// token_in = Decimal("42000000")
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000

			// rounding_direction = ROUND_DOWN # round delta down since we're swapping asset 1 in
			// sqrt_delta = (token_in / liq).quantize(precision, rounding=rounding_direction)
			// sqrt_next = sqrt_cur + sqrt_delta

			// token_out = floor(liq * (sqrt_next - sqrt_cur) / (sqrt_next * sqrt_cur))
			// token_in = ceil(liq * abs(sqrt_cur - sqrt_next))

			// print(sqrt_next)
			// print(token_in)
			// print(token_out)
			ExpectedTokenIn:  sdk.NewCoin("usdc", osmomath.NewInt(42000000)),
			ExpectedTokenOut: sdk.NewCoin("eth", osmomath.NewInt(8396)),
			ExpectedTick:     31003913,
			// Corresponds to sqrt_next in script above
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.738348247484497718514800000003600626"),
			// tick's accum coins stay same since crossing tick does not occur in this case
		},
		"single position within one tick: eth -> usdc": {
			TokenIn:       sdk.NewCoin("eth", osmomath.NewInt(13370)),
			TokenOutDenom: "usdc",
			PriceLimit:    osmomath.NewBigDec(4993),
			SpreadFactor:  osmomath.ZeroDec(),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// from math import *
			// from decimal import *

			// liquidity = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_in = 13370

			// sqrt_next = get_next_sqrt_price_from_amount0_in_round_up(liquidity, sqrt_cur, token_in)
			// token_out = round_sdk_prec_down(calc_amount_one_delta(liquidity, sqrt_cur, sqrt_next, False))
			// token_in = ceil(calc_amount_zero_delta(liquidity, sqrt_cur, sqrt_next, True))

			// # Summary
			// print(sqrt_next)
			// print(token_out)
			// print(token_in)
			ExpectedTokenIn:   sdk.NewCoin("eth", osmomath.NewInt(13370)),
			ExpectedTokenOut:  sdk.NewCoin("usdc", osmomath.NewInt(66808388)),
			ExpectedTick:      30993777,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.666663910857144332134093938182290274"),
		},
		"single position within one tick: eth -> usdc, with zero price limit": {
			TokenIn:       sdk.NewCoin("eth", osmomath.NewInt(13370)),
			TokenOutDenom: "usdc",
			PriceLimit:    osmomath.ZeroBigDec(),
			SpreadFactor:  osmomath.ZeroDec(),
			// from math import *
			// from decimal import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// liquidity = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_in = 13370

			// sqrt_next = (liquidity * sqrt_cur / (liquidity + token_in * sqrt_cur)).quantize(precision, rounding=ROUND_UP)

			// # CalcAmount0Delta rounded up
			// expectedTokenIn = ((liquidity * (sqrt_cur - sqrt_next)) / (sqrt_cur * sqrt_next)).quantize(Decimal('1'), rounding=ROUND_UP)
			// # CalcAmount1Delta rounded down
			// expectedTokenOut = (liquidity * (sqrt_cur - sqrt_next)).quantize(Decimal('1'), rounding=ROUND_DOWN)

			// # Summary
			// print(sqrt_next)
			// print(expectedTokenIn)
			// print(expectedTokenOut)
			ExpectedTokenIn:  sdk.NewCoin("eth", osmomath.NewInt(13370)),
			ExpectedTokenOut: sdk.NewCoin("usdc", osmomath.NewInt(66808388)),
			ExpectedTick:     30993777,
			// True value with arbitrary precision: 70.6666639108571443321...
			// Expected value due to our monotonic sqrt's >= true value guarantee: 70.666663910857144333
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.666663910857144332134093938182290274"),
		},
		//  Two equal price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//  4545 -----|----- 5500
		"two positions within one tick: usdc -> eth": {
			TokenIn:                  sdk.NewCoin("usdc", osmomath.NewInt(42000000)),
			TokenOutDenom:            "eth",
			PriceLimit:               osmomath.NewBigDec(5002),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: DefaultLowerPrice,
			SecondPositionUpperPrice: DefaultUpperPrice,
			// from math import *
			// from decimal import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// token_in = Decimal("42000000")
			// liq = 2 * Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000

			// rounding_direction = ROUND_DOWN # round delta down since we're swapping asset 1 in
			// sqrt_delta = (token_in / liq).quantize(precision, rounding=rounding_direction)
			// sqrt_next = sqrt_cur + sqrt_delta

			// token_out = floor(liq * (sqrt_next - sqrt_cur) / (sqrt_next * sqrt_cur))
			// token_in = ceil(liq * abs(sqrt_cur - sqrt_next))

			// print(sqrt_next)
			// print(token_in)
			// print(token_out)
			ExpectedTokenIn:  sdk.NewCoin("usdc", osmomath.NewInt(42000000)),
			ExpectedTokenOut: sdk.NewCoin("eth", osmomath.NewInt(8398)),
			ExpectedTick:     31001956,
			// Corresponds to sqrt_next in script above
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.724513183069625079757400000001800313"),
			// two positions with same liquidity entered
			PoolLiqAmount0: osmomath.NewInt(1000000).MulRaw(2),
			PoolLiqAmount1: osmomath.NewInt(5000000000).MulRaw(2),
		},
		"two positions within one tick: eth -> usdc": {
			TokenIn:                  sdk.NewCoin("eth", osmomath.NewInt(13370)),
			TokenOutDenom:            "usdc",
			PriceLimit:               osmomath.NewBigDec(4996),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: DefaultLowerPrice,
			SecondPositionUpperPrice: DefaultUpperPrice,
			// from math import *
			// from decimal import *
			// getcontext().prec = 60

			// liquidity = 2 * Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_in = 13370

			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// sqrt_next = (liquidity * sqrt_cur / (liquidity + token_in * sqrt_cur)).quantize(precision, rounding=ROUND_UP)

			// # CalcAmount0Delta rounded up
			// expectedTokenIn = ((liquidity * (sqrt_cur - sqrt_next)) / (sqrt_cur * sqrt_next)).quantize(Decimal('1'), rounding=ROUND_UP)
			// # CalcAmount1Delta rounded down
			// expectedTokenOut = (liquidity * (sqrt_cur - sqrt_next)).quantize(Decimal('1'), rounding=ROUND_DOWN)

			// # Summary
			// print(sqrt_next)
			// print(expectedTokenIn)
			// print(expectedTokenOut)
			ExpectedTokenIn:  sdk.NewCoin("eth", osmomath.NewInt(13370)),
			ExpectedTokenOut: sdk.NewCoin("usdc", osmomath.NewInt(66829187)),
			ExpectedTick:     30996887,
			// True value with arbitrary precision: 70.6886641634088363202...
			// Expected value due to our monotonic sqrt's >= true value guarantee: 70.688664163408836321
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.688664163408836320215015370847179540"),
			// two positions with same liquidity entered
			PoolLiqAmount0: osmomath.NewInt(1000000).MulRaw(2),
			PoolLiqAmount1: osmomath.NewInt(5000000000).MulRaw(2),
		},
		//  Consecutive price ranges

		//          5000
		//  4545 -----|----- 5500
		//             5500 ----------- 6250

		"two positions with consecutive price ranges: usdc -> eth": {
			TokenIn:                  sdk.NewCoin("usdc", osmomath.NewInt(10000000000)),
			TokenOutDenom:            "eth",
			PriceLimit:               osmomath.NewBigDec(6255),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(5500),
			SecondPositionUpperPrice: osmomath.NewDec(6250),
			// from math import *
			// from decimal import *
			// # Range 1: From 5000 to 5500
			// token_in = Decimal("10000000000")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("74.161984870956629488") # sqrt5500

			// token_out_1 = liq_1 * (sqrt_next_1 - sqrt_cur ) / (sqrt_next_1 * sqrt_cur)
			// token_in_1 = ceil(liq_1 * abs(sqrt_cur - sqrt_next_1 ))

			// token_in = token_in - token_in_1

			// # Range 2: from 5500 till end
			// # Using clmath.py scripts: get_liquidity_from_amounts(DefaultCurrSqrtPrice, sqrt5500, sqrt6250, DefaultPoolLiq0, DefaultPoolLiq1)
			// liq_2 = Decimal("1197767444.955508123483846019")

			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// rounding_direction = ROUND_DOWN # round delta down since we're swapping asset 1 in
			// sqrt_delta = (token_in / liq_2).quantize(precision, rounding=rounding_direction)
			// sqrt_next_2 = sqrt_next_1 + sqrt_delta

			// token_out_2 = liq_2 * (sqrt_next_2 - sqrt_next_1 ) / (sqrt_next_1 * sqrt_next_2)
			// token_in_2 = ceil(liq_2 * abs(sqrt_next_2 - sqrt_next_1 ))

			// # Summary:
			// token_in = token_in_1 + token_in_2
			// token_out = (token_out_1 + token_out_2).quantize(Decimal('1'), rounding=ROUND_DOWN)
			// print(sqrt_next_2)
			// print(token_in)
			// print(token_out)
			ExpectedTokenIn:  sdk.NewCoin("usdc", osmomath.NewInt(10000000000)),
			ExpectedTokenOut: sdk.NewCoin("eth", osmomath.NewInt(1820630)),
			ExpectedTick:     32105414,
			// Equivalent to sqrt_next_2 in the script above
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("78.137149196095607130096044752300452857"),
			//  second positions both have greater tick than the current tick, thus never initialized
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 315000, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(5500),
			NewUpperPrice: osmomath.NewDec(6250),
		},
		//  Consecutive price ranges
		//
		//                     5000
		//             4545 -----|----- 5500
		//  4000 ----------- 4545
		//
		"two positions with consecutive price ranges: eth -> usdc": {
			TokenIn:                  sdk.NewCoin("eth", osmomath.NewInt(2000000)),
			TokenOutDenom:            "usdc",
			PriceLimit:               osmomath.NewBigDec(3900),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(4000),
			SecondPositionUpperPrice: osmomath.NewDec(4545),
			ExpectedTokenIn:          sdk.NewCoin("eth", osmomath.NewInt(2000000)),
			ExpectedTokenOut:         sdk.NewCoin("usdc", osmomath.NewInt(9103422788)),
			// crosses one tick with spread reward growth outside
			ExpectedTick: 30095166,
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// token_in = Decimal("2000000")

			// # Swap step 1
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("67.416615162732695594")
			// token_out = round_sdk_prec_down(calc_amount_one_delta(liq, sqrt_cur, sqrt_next_1, False))
			// token_in = token_in - ceil(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next_1, True))

			// # Swap step 2
			// liq_2 = Decimal("1198735489.597250295669959397")
			// sqrt_next_2 = get_next_sqrt_price_from_amount0_in_round_up(liq_2, sqrt_next_1, token_in)
			// token_out = token_out + round_sdk_prec_down(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in = token_in - ceil(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, True))

			// print(sqrt_next_2)
			// print(token_out)
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("63.993489023323078692803734142129673908"),
			// crossing tick happens single time for each upper tick and lower tick.
			// Thus the tick's spread reward growth is DefaultSpreadRewardAccumCoins * 3 - DefaultSpreadRewardAccumCoins
			ExpectedLowerTickSpreadRewardGrowth: DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedUpperTickSpreadRewardGrowth: DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			//  second positions both have greater tick than the current tick, thus never initialized
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 300000, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 305450, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(4000),
			NewUpperPrice: osmomath.NewDec(4545),
		},
		//  Partially overlapping price ranges

		//          5000
		//  4545 -----|----- 5500
		//        5001 ----------- 6250
		//
		"two positions with partially overlapping price ranges: usdc -> eth": {
			TokenIn:                  sdk.NewCoin("usdc", osmomath.NewInt(10000000000)),
			TokenOutDenom:            "eth",
			PriceLimit:               osmomath.NewBigDec(6056),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(5001),
			SecondPositionUpperPrice: osmomath.NewDec(6250),
			ExpectedTokenIn:          sdk.NewCoin("usdc", osmomath.NewInt(10000000000)),
			ExpectedTokenOut:         sdk.NewCoin("eth", osmomath.NewInt(1864161)),
			ExpectedTick:             32055919,
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// from math import *
			// from decimal import *
			// # Range 1: From 5000 to 5500
			// token_in = Decimal("10000000000")

			// # Swap step 1
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("70.717748832948578243")
			// token_out = round_sdk_prec_down(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next_1, False))
			// token_in = token_in - ceil(calc_amount_one_delta(liq, sqrt_cur, sqrt_next_1, True))

			// # Swap step 2
			// liq_2 = Decimal("2188298432.357179144666797069")
			// sqrt_next_2 = Decimal("74.161984870956629488")
			// token_out = token_out + round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in = token_in - ceil(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, True))

			// # Swap step 3
			// liq_3 = Decimal("670416088.605668727039240782")
			// sqrt_next_3 = get_next_sqrt_price_from_amount1_in_round_down(liq_3, sqrt_next_2, token_in)

			// print(sqrt_next_3)
			// print(token_out)
			ExpectedSqrtPrice:                         osmomath.MustNewBigDecFromStr("77.819789636800169393792766394158739007"),
			ExpectedLowerTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins,
			ExpectedUpperTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins,
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 310010, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice:                             osmomath.NewDec(5001),
			NewUpperPrice:                             osmomath.NewDec(6250),
		},
		"two positions with partially overlapping price ranges, not utilizing full liquidity of second position: usdc -> eth": {
			TokenIn:       sdk.NewCoin("usdc", osmomath.NewInt(8500000000)),
			TokenOutDenom: "eth",
			PriceLimit:    osmomath.NewBigDec(6056),
			SpreadFactor:  osmomath.ZeroDec(),
			// from math import *
			// from decimal import *
			// getcontext().prec = 60
			// # Range 1: From 5000 to 5001
			// token_in = Decimal("8500000000")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("70.717748832948578243") # sqrt5001

			// token_out_1 = liq_1 * (sqrt_next_1 - sqrt_cur ) / (sqrt_next_1 * sqrt_cur)
			// token_in_1 = ceil(liq_1 * (sqrt_next_1 - sqrt_cur ))

			// token_in = token_in - token_in_1

			// # Range 2: from 5001 to 5500:
			// # Using clmath.py scripts: get_liquidity_from_amounts(DefaultCurrSqrtPrice, sqrt5001, sqrt6250, DefaultPoolLiq0, DefaultPoolLiq1)
			// second_pos_liq = Decimal("670416088.605668727039240782")
			// liq_2 = liq_1 + second_pos_liq
			// sqrt_next_2 = Decimal("74.161984870956629488") # sqrt5500

			// token_out_2 = liq_2 * (sqrt_next_2 - sqrt_next_1 ) / (sqrt_next_1 * sqrt_next_2)
			// token_in_2 = ceil(liq_2 * (sqrt_next_2 - sqrt_next_1 ))

			// token_in = token_in - token_in_2

			// # Range 3: from 5500 till end
			// liq_3 = second_pos_liq
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// rounding_direction = ROUND_DOWN # round delta down since we're swapping asset 1 in
			// sqrt_delta = (token_in / liq_3).quantize(precision, rounding=rounding_direction)
			// sqrt_next_3 = sqrt_next_2 + sqrt_delta

			// token_out_3 = liq_3 * (sqrt_next_3 - sqrt_next_2 ) / (sqrt_next_3 * sqrt_next_2)
			// token_in_3 = ceil(liq_3 * (sqrt_next_3 - sqrt_next_2 ))

			// # Summary:
			// token_in = token_in_1 + token_in_2 + token_in_3
			// token_out = (token_out_1 + token_out_2 + token_out_3).quantize(Decimal('1'), rounding=ROUND_DOWN)
			// print(sqrt_next_3)
			// print(token_in)
			// print(token_out)
			SecondPositionLowerPrice:                  osmomath.NewDec(5001),
			SecondPositionUpperPrice:                  osmomath.NewDec(6250),
			ExpectedTokenIn:                           sdk.NewCoin("usdc", osmomath.NewInt(8500000000)),
			ExpectedTokenOut:                          sdk.NewCoin("eth", osmomath.NewInt(1609138)),
			ExpectedLowerTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins,
			ExpectedUpperTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins,
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 310010, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedTick:                              31712695,
			// Corresponds to sqrt_next_3 in the script above
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("75.582373164412551492069079174313215667"),
			NewLowerPrice:     osmomath.NewDec(5001),
			NewUpperPrice:     osmomath.NewDec(6250),
		},
		//  Partially overlapping price ranges
		//
		//                5000
		//        4545 -----|----- 5500
		//  4000 ----------- 4999
		//
		"two positions with partially overlapping price ranges: eth -> usdc": {
			TokenIn:       sdk.NewCoin("eth", osmomath.NewInt(2000000)),
			TokenOutDenom: "usdc",
			PriceLimit:    osmomath.NewBigDec(4128),
			SpreadFactor:  osmomath.ZeroDec(),
			// from math import *
			// from decimal import *
			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// rounding_direction = ROUND_UP # round delta up since we're swapping asset 0 in

			// # Setup
			// token_in = Decimal("2000000")
			// liq_pos_1 = Decimal("1517882343.751510417627556287")
			// # Using clmath.py scripts: get_liquidity_from_amounts(DefaultCurrSqrtPrice, sqrt4000, sqrt4999, DefaultPoolLiq0, DefaultPoolLiq1)
			// liq_pos_2 = Decimal("670416215.718827443660400593")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000

			// # Swapping through range 5000 -> 4999
			// liq_0 = liq_pos_1
			// sqrt_next_0 = Decimal("70.703606697254136613") # sqrt4999
			// token_out_0 = liq_0 * abs(sqrt_cur - sqrt_next_0 )
			// token_in_0 = ceil(liq_0 * abs(sqrt_cur - sqrt_next_0 ) / (sqrt_next_0 * sqrt_cur))
			// token_in = token_in - token_in_0

			// # Swapping through range 4999 -> 4545
			// liq_1 = liq_pos_1 + liq_pos_2
			// sqrt_next_1 = Decimal("67.416615162732695594") # sqrt4545
			// token_out_1 = liq_1 * abs(sqrt_next_0 - sqrt_next_1 )
			// token_in_1 = ceil(liq_1 * abs(sqrt_next_0 - sqrt_next_1 ) / (sqrt_next_1 * sqrt_next_0))

			// token_in = token_in - token_in_1

			// # Swapping through range 4545 -> end
			// liq_2 = liq_pos_2
			// sqrt_next_2 = (liq_2 * sqrt_next_1 / (liq_2 + token_in * sqrt_next_1)).quantize(precision, rounding=rounding_direction)
			// token_out_2 = liq_2 * (sqrt_next_1 - sqrt_next_2 )
			// token_in_2 = ceil(liq_2 * (sqrt_next_1 - sqrt_next_2 ) / (sqrt_next_2 * sqrt_next_1))

			// # Summary:
			// token_in = token_in_0 + token_in_1 + token_in_2
			// token_out = (token_out_0 + token_out_1 + token_out_2).quantize(Decimal('1'), rounding=ROUND_DOWN)
			// print(sqrt_next_2)
			// print(token_in)
			// print(token_out)
			SecondPositionLowerPrice: osmomath.NewDec(4000),
			SecondPositionUpperPrice: osmomath.NewDec(4999),
			ExpectedTokenIn:          sdk.NewCoin("eth", osmomath.NewInt(2000000)),
			ExpectedTokenOut:         sdk.NewCoin("usdc", osmomath.NewInt(9321276930)),
			ExpectedTick:             30129083,
			// Corresponds to sqrt_next_2 in the script above
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("64.257943794993248953756640624575523292"),
			// Started from DefaultSpreadRewardAccumCoins * 3, crossed tick once, thus becoming
			// DefaultSpreadRewardAccumCoins * 3 - DefaultSpreadRewardAccumCoins = DefaultSpreadRewardAccumCoins * 2
			ExpectedLowerTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedUpperTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 300000, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 309990, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(4000),
			NewUpperPrice: osmomath.NewDec(4999),
		},
		//          		5000
		//  		4545 -----|----- 5500
		//  4000 ---------- 4999
		"two positions with partially overlapping price ranges, not utilizing full liquidity of second position: eth -> usdc": {
			TokenIn:                  sdk.NewCoin("eth", osmomath.NewInt(1800000)),
			TokenOutDenom:            "usdc",
			PriceLimit:               osmomath.NewBigDec(4128),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(4000),
			SecondPositionUpperPrice: osmomath.NewDec(4999),
			ExpectedTokenIn:          sdk.NewCoin("eth", osmomath.NewInt(1800000)),
			ExpectedTokenOut:         sdk.NewCoin("usdc", osmomath.NewInt(8479320318)),
			ExpectedTick:             30292059,
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// token_in = Decimal("1800000")

			// # Swap step 1
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("70.703606697254136613")
			// token_out = round_sdk_prec_down(calc_amount_one_delta(liq, sqrt_cur, sqrt_next_1, False))
			// token_in = token_in - ceil(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next_1, True))

			// # Swap step 2
			// liq_2 = Decimal("2188298559.470337861287956880")
			// sqrt_next_2 = Decimal("67.416615162732695594")
			// token_out = token_out + round_sdk_prec_down(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in = token_in - ceil(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, True))

			// # Swap step 3
			// liq_3 = Decimal("670416215.718827443660400593")
			// sqrt_next_3 = get_next_sqrt_price_from_amount0_in_round_up(liq_3, sqrt_next_2, token_in)
			// token_out = token_out + round_sdk_prec_down(calc_amount_one_delta(liq_3, sqrt_next_2, sqrt_next_3, False))
			// token_in = token_in - ceil(calc_amount_zero_delta(liq_3, sqrt_next_2, sqrt_next_3, True))

			// print(sqrt_next_3)
			// print(token_out)
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("65.513815285481060959469337552596846421"),
			// Started from DefaultSpreadRewardAccumCoins * 3, crossed tick once, thus becoming
			// DefaultSpreadRewardAccumCoins * 3 - DefaultSpreadRewardAccumCoins = DefaultSpreadRewardAccumCoins * 2
			ExpectedLowerTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedUpperTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 300000, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 309990, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(4000),
			NewUpperPrice: osmomath.NewDec(4999),
		},
		//  Sequential price ranges with a gap
		//
		//          5000
		//  4545 -----|----- 5500
		//              5501 ----------- 6250
		//
		"two sequential positions with a gap": {
			TokenIn:                  sdk.NewCoin("usdc", osmomath.NewInt(10000000000)),
			TokenOutDenom:            "eth",
			PriceLimit:               osmomath.NewBigDec(6106),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(5501),
			SecondPositionUpperPrice: osmomath.NewDec(6250),
			ExpectedTokenIn:          sdk.NewCoin("usdc", osmomath.NewInt(10000000000)),
			ExpectedTokenOut:         sdk.NewCoin("eth", osmomath.NewInt(1820545)),
			ExpectedTick:             32105555,
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// token_in = Decimal("10000000000")

			// # Swap step 1
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("74.161984870956629488")
			// token_out = round_sdk_prec_down(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next_1, False))
			// token_in = token_in - ceil(calc_amount_one_delta(liq, sqrt_cur, sqrt_next_1, True))

			// # Swap step 2
			// liq_2 = Decimal("0.000000000000000000")
			// sqrt_next_2 = Decimal("74.168726563154635304")
			// token_out = token_out + round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in = token_in - ceil(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, True))

			// # Swap step 3
			// liq_3 = Decimal("1199528406.187413669481596330")
			// sqrt_next_3 = get_next_sqrt_price_from_amount1_in_round_down(liq_3, sqrt_next_2, token_in)
			// token_out = token_out + round_sdk_prec_down(calc_amount_zero_delta(liq_3, sqrt_next_2, sqrt_next_3, False))
			// token_in = token_in - ceil(calc_amount_one_delta(liq_3, sqrt_next_2, sqrt_next_3, True))

			// print(sqrt_next_3)
			// print(token_out)
			ExpectedSqrtPrice:                         osmomath.MustNewBigDecFromStr("78.138055169663761658508234345605157554"),
			ExpectedLowerTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins,
			ExpectedUpperTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins,
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 315010, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice:                             osmomath.NewDec(5501),
			NewUpperPrice:                             osmomath.NewDec(6250),
		},
		// Slippage protection doesn't cause a failure but interrupts early.
		//          5000
		//  4545 ---!-|----- 5500
		"single position within one tick, trade completes but slippage protection interrupts trade early: eth -> usdc": {
			TokenIn:       sdk.NewCoin("eth", osmomath.NewInt(13370)),
			TokenOutDenom: "usdc",
			PriceLimit:    osmomath.NewBigDec(4994),
			SpreadFactor:  osmomath.ZeroDec(),
			// from math import *
			// from decimal import *
			//
			// liquidity = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_in = 13370
			//
			// # Exact since we hit price limit before next initialized tick
			// sqrt_next = Decimal("70.668238976219012614") # sqrt4994
			//
			// # CalcAmount0Delta rounded up
			// expectedTokenIn = ((liquidity * (sqrt_cur - sqrt_next)) / (sqrt_cur * sqrt_next)).quantize(Decimal('1'), rounding=ROUND_UP)
			// # CalcAmount1Delta rounded down
			// expectedTokenOut = (liquidity * (sqrt_cur - sqrt_next)).quantize(Decimal('1'), rounding=ROUND_DOWN)
			//
			// # Summary
			// print(sqrt_next)
			// print(expectedTokenIn)
			// print(expectedTokenOut)
			ExpectedTokenIn:  sdk.NewCoin("eth", osmomath.NewInt(12892)),
			ExpectedTokenOut: sdk.NewCoin("usdc", osmomath.NewInt(64417624)),
			ExpectedTick: func() int64 {
				tick, _ := math.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(sqrt4994), DefaultTickSpacing)
				return tick
			}(),
			// Since the next sqrt price is based on the price limit, we can calculate this directly.
			ExpectedSqrtPrice: osmomath.BigDecFromDec(osmomath.MustMonotonicSqrt(osmomath.NewDec(4994))),
		},
	}

	swapOutGivenInSpreadRewardCases = map[string]apptesting.ConcentratedSwapTest{
		//          5000
		//  4545 -----|----- 5500
		"spread factor 1 - single position within one tick: usdc -> eth (1% spread factor)": {
			TokenIn:       sdk.NewCoin("usdc", osmomath.NewInt(42000000)),
			TokenOutDenom: "eth",
			PriceLimit:    osmomath.NewBigDec(5004),
			SpreadFactor:  osmomath.MustNewDecFromStr("0.01"),
			// from math import *
			// from decimal import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')
			// token_in = Decimal("42000000")
			// spread_factor = Decimal("0.01")

			// # Swap step 1
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_in_after_fee = token_in * (1 - spread_factor)
			// sqrt_next = get_next_sqrt_price_from_amount1_in_round_down(liq, sqrt_cur, token_in_after_fee)
			// token_out = round_sdk_prec_down(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next, False))
			// token_in_after_fee = ceil(calc_amount_one_delta(liq, sqrt_cur, sqrt_next, True))
			// fee_charge_total = token_in - token_in_after_fee
			// fee_charge_total = fee_charge_total.quantize(eighteen_decimal_precision, ROUND_UP)
			// fee_charged_total_scaled = fee_charge_total * Decimal('1e27')
			// fee_amount_per_share = fee_charged_total_scaled / liq

			// print(sqrt_next)
			// print(token_in)
			// print(token_out)
			// print(fee_amount_per_share)
			ExpectedTokenIn:   sdk.NewCoin("usdc", osmomath.NewInt(42000000)),
			ExpectedTokenOut:  sdk.NewCoin("eth", osmomath.NewInt(8312)),
			ExpectedTick:      31003800,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.738071546196200264"),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("276701288297452775148000.000036006261209414"),
		},
		//          5000
		//  4545 -----|----- 5500
		//  4545 -----|----- 5500
		"spread factor 2 - two positions within one tick: eth -> usdc (3% spread factor) ": {
			TokenIn:                  sdk.NewCoin("eth", osmomath.NewInt(13370)),
			TokenOutDenom:            "usdc",
			PriceLimit:               osmomath.NewBigDec(4990),
			SpreadFactor:             osmomath.MustNewDecFromStr("0.03"),
			SecondPositionLowerPrice: DefaultLowerPrice,
			SecondPositionUpperPrice: DefaultUpperPrice,
			// from math import *
			// from decimal import *
			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')
			// token_in = Decimal("13370")
			// spread_factor = Decimal("0.03")

			// # Swap step 1
			// liq = 2 * Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_in_after_fee = token_in * (1 - spread_factor)
			// sqrt_next = get_next_sqrt_price_from_amount0_in_round_up(liq, sqrt_cur, token_in_after_fee)
			// token_in_after_fee = ceil(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next, True))
			// token_out = floor(liq * abs(sqrt_cur - sqrt_next))
			// fee_charge_total = token_in - token_in_after_fee
			// fee_charge_total = fee_charge_total.quantize(eighteen_decimal_precision, ROUND_UP)
			// fee_charged_total_scaled = fee_charge_total * Decimal('1e27')
			// fee_amount_per_share = fee_charged_total_scaled / liq

			// print(sqrt_next)
			// print(token_in)
			// print(token_out)
			// print(fee_amount_per_share)
			ExpectedTokenIn:   sdk.NewCoin("eth", osmomath.NewInt(13370)),
			ExpectedTokenOut:  sdk.NewCoin("usdc", osmomath.NewInt(64824917)),
			ExpectedTick:      30996900,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.689324382628080102"),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("132091924532474479564.700000017188703267"),
			// two positions with same liquidity entered
			PoolLiqAmount0: osmomath.NewInt(1000000).MulRaw(2),
			PoolLiqAmount1: osmomath.NewInt(5000000000).MulRaw(2),
		},
		//          		   5000
		//  		   4545 -----|----- 5500
		//  4000 ----------- 4545
		"spread factor 3 - two positions with consecutive price ranges: eth -> usdc (5% spread factor)": {
			TokenIn:                  sdk.NewCoin("eth", osmomath.NewInt(2000000)),
			TokenOutDenom:            "usdc",
			PriceLimit:               osmomath.NewBigDec(4094),
			SpreadFactor:             osmomath.MustNewDecFromStr("0.05"),
			SecondPositionLowerPrice: osmomath.NewDec(4000),
			SecondPositionUpperPrice: osmomath.NewDec(4545),
			// from math import *
			// from decimal import *
			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')
			// token_in = Decimal("2000000")
			// spread_factor = Decimal("0.05")

			// # Swap step 1
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("67.416615162732695594")
			// token_out = round_sdk_prec_down(calc_amount_one_delta(liq, sqrt_cur, sqrt_next_1, False))
			// expected_token_in_before_fee = ceil(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next_1, True))
			// expected_fee = expected_token_in_before_fee * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// expected_fee = expected_fee.quantize(eighteen_decimal_precision, ROUND_UP)
			// expected_token_in = expected_token_in_before_fee + expected_fee
			// expected_fee_scaled = expected_fee * Decimal('1e27')
			// fee_amount_per_share = expected_fee_scaled / liq
			// token_in = token_in - expected_token_in

			// # Swap step 2
			// token_in_remaining_after_fee = token_in * (1 - spread_factor)
			// liq_2 = Decimal("1198735489.597250295669959397")
			// sqrt_next_2 = get_next_sqrt_price_from_amount0_in_round_up(liq_2, sqrt_next_1, token_in_remaining_after_fee)
			// token_out = token_out + round_sdk_prec_down(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_after_fee_rounded_up = ceil(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, True))
			// fee_charge_total = token_in - token_in_after_fee_rounded_up
			// fee_charge_total_scaled = fee_charge_total * Decimal('1e27')
			// fee_amount_per_share_2 = fee_charge_total_scaled / liq_2

			// print(sqrt_next_2)
			// print(token_in)
			// print(token_out)
			// print(fee_amount_per_share + fee_amount_per_share_2)
			ExpectedTokenIn:  sdk.NewCoin("eth", osmomath.NewInt(2000000)),
			ExpectedTokenOut: sdk.NewCoin("usdc", osmomath.NewInt(8691708221)),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("73738597832046531336741.571424795459743274"),
			ExpectedTick:      30139200,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("64.336946417392457832"),
			NewLowerPrice:     osmomath.NewDec(4000),
			NewUpperPrice:     osmomath.NewDec(4545),
		},
		//          5000
		//  4545 -----|----- 5500
		//  	  5001 ----------- 6250
		"spread factor 4 - two positions with partially overlapping price ranges: usdc -> eth (10% spread factor)": {
			TokenIn:                  sdk.NewCoin("usdc", osmomath.NewInt(10000000000)),
			TokenOutDenom:            "eth",
			PriceLimit:               osmomath.NewBigDec(6056),
			SpreadFactor:             osmomath.MustNewDecFromStr("0.1"),
			SecondPositionLowerPrice: osmomath.NewDec(5001),
			SecondPositionUpperPrice: osmomath.NewDec(6250),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')
			// token_in = Decimal("10000000000")
			// spread_factor = Decimal("0.1")

			// # Swap step 1
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("70.717748832948578243")
			// token_out = round_sdk_prec_down(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next_1, False))
			// expected_token_in_before_fee = ceil(calc_amount_one_delta(liq, sqrt_cur, sqrt_next_1, True))
			// expected_fee = expected_token_in_before_fee * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// expected_fee = expected_fee.quantize(eighteen_decimal_precision, ROUND_UP)
			// expected_token_in_1 = expected_token_in_before_fee + expected_fee
			// expected_fee_scaled = expected_fee * Decimal('1e27')
			// fee_amount_per_share = expected_fee_scaled / liq
			// token_in = token_in - expected_token_in_1

			// # Swap step 2
			// liq_2 = liq + Decimal("670416088.605668727039240782")
			// sqrt_next_2 = Decimal("74.161984870956629487")
			// token_out = token_out + round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// expected_token_in_before_fee = ceil(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, True))
			// expected_fee = expected_token_in_before_fee * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// expected_fee = expected_fee.quantize(eighteen_decimal_precision, ROUND_UP)
			// expected_token_in_2 = expected_token_in_before_fee + expected_fee
			// expected_fee_scaled = expected_fee * Decimal('1e27')
			// fee_amount_per_share_2 = expected_fee_scaled / liq_2
			// token_in = token_in - expected_token_in_2

			// # Swap step 3
			// token_in_remaining_after_fee = token_in * (1 - spread_factor)
			// liq_3 = Decimal("670416088.605668727039240782")
			// sqrt_next_3 = get_next_sqrt_price_from_amount1_in_round_down(liq_3, sqrt_next_2, token_in_remaining_after_fee)
			// token_out = token_out + round_sdk_prec_down(calc_amount_zero_delta(liq_3, sqrt_next_2, sqrt_next_3, False))
			// token_in_after_fee_rounded_up = ceil(calc_amount_one_delta(liq_3, sqrt_next_2, sqrt_next_3, True))
			// fee_charge_total = token_in - token_in_after_fee_rounded_up
			// expected_token_in_3 = token_in_remaining_after_fee + fee_charge_total
			// fee_charge_total_scaled = fee_charge_total * Decimal('1e27')
			// fee_amount_per_share_3 = fee_charge_total_scaled / liq_3

			// print(sqrt_next_3)
			// print(expected_token_in_1 + expected_token_in_2 + expected_token_in_3)
			// print(token_out)
			// print(fee_amount_per_share + fee_amount_per_share_2 + fee_amount_per_share_3)
			ExpectedTokenIn:  sdk.NewCoin("usdc", osmomath.NewInt(10000000000)),
			ExpectedTokenOut: sdk.NewCoin("eth", osmomath.NewInt(1695807)),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("624166726347032851920318136.447360021110663239"),
			ExpectedTick:      31825900,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("76.328178655208424124"),
			NewLowerPrice:     osmomath.NewDec(5001),
			NewUpperPrice:     osmomath.NewDec(6250),
		},
		//          		5000
		//  		4545 -----|----- 5500
		// 4000 ----------- 4999
		"spread factor 5 - two positions with partially overlapping price ranges, not utilizing full liquidity of second position: eth -> usdc (0.5% spread factor)": {
			TokenIn:                  sdk.NewCoin("eth", osmomath.NewInt(1800000)),
			TokenOutDenom:            "usdc",
			PriceLimit:               osmomath.NewBigDec(4128),
			SpreadFactor:             osmomath.MustNewDecFromStr("0.005"),
			SecondPositionLowerPrice: osmomath.NewDec(4000),
			SecondPositionUpperPrice: osmomath.NewDec(4999),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')
			// token_in = Decimal("1800000")
			// spread_factor = Decimal("0.005")

			// # Swap step 1
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("70.703606697254136613")
			// token_out = round_sdk_prec_down(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// expected_token_in_before_fee = ceil(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, True))
			// expected_fee = expected_token_in_before_fee * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// expected_fee = expected_fee.quantize(eighteen_decimal_precision, ROUND_UP)
			// expected_token_in_1 = expected_token_in_before_fee + expected_fee
			// expected_fee_scaled = expected_fee * Decimal('1e27')
			// fee_amount_per_share = expected_fee_scaled / liq_1
			// token_in = token_in - expected_token_in_1

			// # Swap step 2
			// second_pos_liq = Decimal("670416215.718827443660400593")
			// liq_2 = liq_1 + second_pos_liq
			// sqrt_next_2 = Decimal("67.416615162732695594")
			// token_out = token_out + round_sdk_prec_down(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// expected_token_in_before_fee = ceil(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, True))
			// expected_fee = expected_token_in_before_fee * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// expected_fee = expected_fee.quantize(eighteen_decimal_precision, ROUND_UP)
			// expected_token_in_2 = expected_token_in_before_fee + expected_fee
			// expected_fee_scaled = expected_fee * Decimal('1e27')
			// fee_amount_per_share_2 = expected_fee_scaled / liq_2
			// token_in = token_in - expected_token_in_2

			// # Swap step 3
			// token_in_remaining_after_fee = token_in * (1 - spread_factor)
			// liq_3 = second_pos_liq
			// sqrt_next_3 = get_next_sqrt_price_from_amount0_in_round_up(liq_3, sqrt_next_2, token_in_remaining_after_fee)
			// token_out = token_out + round_sdk_prec_down(calc_amount_one_delta(liq_3, sqrt_next_2, sqrt_next_3, False))
			// token_in_after_fee_rounded_up = ceil(calc_amount_zero_delta(liq_3, sqrt_next_2, sqrt_next_3, True))
			// fee_charge_total = token_in - token_in_after_fee_rounded_up
			// expected_token_in_3 = token_in_remaining_after_fee + fee_charge_total
			// fee_charge_total_scaled = fee_charge_total * Decimal('1e27')
			// fee_amount_per_share_3 = fee_charge_total_scaled / liq_3

			// print(sqrt_next_3)
			// print(expected_token_in_1 + expected_token_in_2 + expected_token_in_3)
			// print(token_out)
			// print(fee_amount_per_share + fee_amount_per_share_2 + fee_amount_per_share_3)
			ExpectedTokenIn:  sdk.NewCoin("eth", osmomath.NewInt(1800000)),
			ExpectedTokenOut: sdk.NewCoin("usdc", osmomath.NewInt(8440657775)),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("5569829831409674130027.984038662045276604"),
			ExpectedTick:      30299600,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("65.571484748647169032"),
			NewLowerPrice:     osmomath.NewDec(4000),
			NewUpperPrice:     osmomath.NewDec(4999),
		},
		//          5000
		//  4545 -----|----- 5500
		// 			   5501 ----------- 6250
		"spread factor 6 - two sequential positions with a gap usdc -> eth (3% spread factor)": {
			TokenIn:                  sdk.NewCoin("usdc", osmomath.NewInt(10000000000)),
			TokenOutDenom:            "eth",
			PriceLimit:               osmomath.NewBigDec(6106),
			SecondPositionLowerPrice: osmomath.NewDec(5501),
			SecondPositionUpperPrice: osmomath.NewDec(6250),
			SpreadFactor:             osmomath.MustNewDecFromStr("0.03"),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')
			// token_in = Decimal("10000000000")
			// spread_factor = Decimal("0.03")

			// # Swap step 1
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("74.161984870956629488")
			// token_out = round_sdk_prec_down(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next_1, False))
			// expected_token_in_before_fee = ceil(calc_amount_one_delta(liq, sqrt_cur, sqrt_next_1, True))
			// expected_fee = expected_token_in_before_fee * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// expected_fee = expected_fee.quantize(eighteen_decimal_precision, ROUND_UP)
			// expected_token_in_1 = expected_token_in_before_fee + expected_fee
			// expected_fee_scaled = expected_fee * Decimal('1e27')
			// fee_amount_per_share = expected_fee_scaled / liq
			// token_in = token_in - expected_token_in_1

			// # Swap step 2
			// token_in_remaining_after_fee = token_in * (1 - spread_factor)
			// sqrt_next_1 = Decimal("74.168726563154635304")
			// liq_2 = Decimal("1199528406.187413669481596330")
			// sqrt_next_2 = get_next_sqrt_price_from_amount1_in_round_down(liq_2, sqrt_next_1, token_in_remaining_after_fee)
			// token_out = token_out + round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_after_fee_rounded_up = ceil(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, True))
			// fee_charge_total = token_in - token_in_after_fee_rounded_up
			// expected_token_in_2 = token_in_after_fee_rounded_up + fee_charge_total
			// fee_charge_total_scaled = fee_charge_total * Decimal('1e27')
			// fee_amount_per_share_2 = fee_charge_total_scaled / liq_2

			// print(sqrt_next_2)
			// print(expected_token_in_1 + expected_token_in_2)
			// print(token_out)
			// print(fee_amount_per_share + fee_amount_per_share_2)
			ExpectedTokenIn:  sdk.NewCoin("usdc", osmomath.NewInt(10000000000)),
			ExpectedTokenOut: sdk.NewCoin("eth", osmomath.NewInt(1771252)),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("221769187794051752204718133.920636386549115528"),
			ExpectedTick:      32066500,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("77.887956882326389372"),
			NewLowerPrice:     osmomath.NewDec(5501),
			NewUpperPrice:     osmomath.NewDec(6250),
		},
		//          5000
		//  4545 ---!-|----- 5500
		"spread factor 7: single position within one tick, trade completes but slippage protection interrupts trade early: eth -> usdc (1% spread factor)": {
			TokenIn:       sdk.NewCoin("eth", osmomath.NewInt(13370)),
			TokenOutDenom: "usdc",
			PriceLimit:    osmomath.NewBigDec(4994),
			SpreadFactor:  osmomath.MustNewDecFromStr("0.01"),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')
			// token_in = Decimal("13370")
			// spread_factor = Decimal("0.01")

			// # Swap step 1
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("70.668238976219012613")
			// token_out = round_sdk_prec_down(calc_amount_one_delta(liq, sqrt_cur, sqrt_next_1, False))
			// expected_token_in_before_fee = ceil(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next_1, True))
			// expected_fee = expected_token_in_before_fee * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// expected_fee = expected_fee.quantize(eighteen_decimal_precision, ROUND_UP)
			// expected_token_in = expected_token_in_before_fee + expected_fee
			// expected_fee_scaled = expected_fee * Decimal('1e27')
			// fee_amount_per_share = expected_fee_scaled / liq
			// token_in = token_in - expected_token_in

			// print(sqrt_next_1)
			// print(expected_token_in)
			// print(token_out)
			// print(fee_amount_per_share)
			ExpectedTokenIn:  sdk.NewCoin("eth", osmomath.NewInt(13023)),
			ExpectedTokenOut: sdk.NewCoin("usdc", osmomath.NewInt(64417624)),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("85792039652014466733.930997019691873157"),
			ExpectedTick: func() int64 {
				tick, _ := math.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(sqrt4994), DefaultTickSpacing)
				return tick
			}(),
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.668238976219012614"),
		},
	}

	swapOutGivenInErrorCases = map[string]apptesting.ConcentratedSwapTest{
		"single position within one tick, trade does not complete due to lack of liquidity: usdc -> eth": {
			TokenIn:       sdk.NewCoin("usdc", osmomath.NewInt(5300000000)),
			TokenOutDenom: "eth",
			PriceLimit:    osmomath.NewBigDec(6000),
			SpreadFactor:  osmomath.ZeroDec(),
			ExpectErr:     true,
		},
		"single position within one tick, trade does not complete due to lack of liquidity: eth -> usdc": {
			TokenIn:       sdk.NewCoin("eth", osmomath.NewInt(1100000)),
			TokenOutDenom: "usdc",
			PriceLimit:    osmomath.NewBigDec(4000),
			SpreadFactor:  osmomath.ZeroDec(),
			ExpectErr:     true,
		},
	}

	// Note: liquidity value for the default position is 1517882343.751510417627556287
	swapInGivenOutTestCases = map[string]apptesting.ConcentratedSwapTest{
		//  One price range
		//
		//          5000
		//  4545 -----|----- 5500
		"single position within one tick: eth (in) -> usdc (out) | zfo": {
			TokenOut:     sdk.NewCoin(USDC, osmomath.NewInt(42000000)),
			TokenInDenom: ETH,
			PriceLimit:   osmomath.NewBigDec(4993),
			SpreadFactor: osmomath.ZeroDec(),
			// from math import *
			// from decimal import *

			// import sys

			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_out = Decimal("42000000")

			// sqrt_next = get_next_sqrt_price_from_amount1_out_round_down(liq, sqrt_cur, token_out)
			// token_in = token_in = liq * (sqrt_cur - sqrt_next) / (sqrt_cur * sqrt_next)
			// print(sqrt_next)
			// print(token_in)
			ExpectedTokenOut:  sdk.NewCoin(USDC, osmomath.NewInt(42000000)),
			ExpectedTokenIn:   sdk.NewCoin(ETH, osmomath.NewInt(8404)),
			ExpectedTick:      30996087,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.683007989825007163485199999996399373"),
		},
		"single position within one tick: usdc (in) -> eth (out) ofz": {
			TokenOut:     sdk.NewCoin(ETH, osmomath.NewInt(13370)),
			TokenInDenom: USDC,
			PriceLimit:   osmomath.NewBigDec(5010),
			SpreadFactor: osmomath.ZeroDec(),
			// from math import *
			// from decimal import *
			// import sys
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *
			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_out = Decimal("13370")
			// sqrt_next = get_next_sqrt_price_from_amount0_out_round_up(liq, sqrt_cur, token_out)
			// token_in = ceil(liq * abs(sqrt_cur - sqrt_next))
			// print(sqrt_next)
			// print(token_in)
			ExpectedTokenOut:  sdk.NewCoin(ETH, osmomath.NewInt(13370)),
			ExpectedTokenIn:   sdk.NewCoin(USDC, osmomath.NewInt(66891663)),
			ExpectedTick:      31006234,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.754747188468900467378792612053774781"),
		},
		//  Two equal price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//  4545 -----|----- 5500
		"two positions within one tick: eth (in) -> usdc (out) | zfo": {
			TokenOut:                 sdk.NewCoin("usdc", osmomath.NewInt(66829187)),
			TokenInDenom:             "eth",
			PriceLimit:               osmomath.NewBigDec(4990),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: DefaultLowerPrice,
			SecondPositionUpperPrice: DefaultUpperPrice,
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// liq = Decimal("1517882343.751510417627556287") * 2
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_out = Decimal("66829187")

			// sqrt_next = get_next_sqrt_price_from_amount1_out_round_down(liq, sqrt_cur, token_out)
			// token_in = token_in = liq * (sqrt_cur - sqrt_next) / (sqrt_cur * sqrt_next)
			// print(sqrt_next)
			// print(token_in)
			ExpectedTokenOut: sdk.NewCoin("usdc", osmomath.NewInt(66829187)),
			ExpectedTokenIn:  sdk.NewCoin("eth", osmomath.NewInt(13370)),
			ExpectedTick:     30996887,
			// This value is the direct output of sqrt_next in the script above.
			// The precision is exact because we properly handle rounding behavior in intermediate steps.
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.688664163727643651554720661097135393"),
			// two positions with same liquidity entered
			PoolLiqAmount0: osmomath.NewInt(1000000).MulRaw(2),
			PoolLiqAmount1: osmomath.NewInt(5000000000).MulRaw(2),
		},
		"two positions within one tick: usdc (in) -> eth (out) | ofz": {
			TokenOut:                 sdk.NewCoin("eth", osmomath.NewInt(8398)),
			TokenInDenom:             "usdc",
			PriceLimit:               osmomath.NewBigDec(5020),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: DefaultLowerPrice,
			SecondPositionUpperPrice: DefaultUpperPrice,
			// from math import *
			// from decimal import *
			// import sys
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *
			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// liq = Decimal("1517882343.751510417627556287") * 2
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_out = Decimal("8398")
			// sqrt_next = get_next_sqrt_price_from_amount0_out_round_up(liq, sqrt_cur, token_out)
			// token_in = ceil(liq * abs(sqrt_cur - sqrt_next))
			// print(sqrt_next)
			// print(token_in)
			ExpectedTokenOut:  sdk.NewCoin("eth", osmomath.NewInt(8398)),
			ExpectedTokenIn:   sdk.NewCoin("usdc", osmomath.NewInt(41998216)),
			ExpectedTick:      31001956,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.724512595179305566327821490232558005"),
			// two positions with same liquidity entered
			PoolLiqAmount0: osmomath.NewInt(1000000).MulRaw(2),
			PoolLiqAmount1: osmomath.NewInt(5000000000).MulRaw(2),
		},
		//  Consecutive price ranges
		//
		//                     5000
		//             4545 -----|----- 5500
		//  4000 ----------- 4545
		"two positions with consecutive price ranges: eth (in) -> usdc (out) | zfo": {
			TokenOut:                 sdk.NewCoin("usdc", osmomath.NewInt(9103422788)),
			TokenInDenom:             "eth",
			PriceLimit:               osmomath.NewBigDec(3900),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(4000),
			SecondPositionUpperPrice: osmomath.NewDec(4545),
			// from math import *
			// from decimal import *

			// import sys

			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// # Swap step 1
			// token_out = Decimal("9103422788")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441")
			// sqrt_next_1 = Decimal("67.416615162732695594")

			// token_out_1 = round_sdk_prec_down(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, True))

			// token_out = token_out - token_out_1

			// # Swap step 2
			// liq_2 = Decimal("1198735489.597250295669959397")
			// sqrt_next_2 = get_next_sqrt_price_from_amount1_out_round_down(liq_2, sqrt_next_1, token_out)

			// token_out_2 = round_sdk_prec_down(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, True))

			// token_out = token_out - token_out_2

			// # Summary:
			// token_in = token_in_1 + token_in_2
			// print(sqrt_next_2)
			// print(token_in)
			// print(token_out_2)
			ExpectedTokenOut: sdk.NewCoin("usdc", osmomath.NewInt(9103422788)),
			ExpectedTokenIn:  sdk.NewCoin("eth", osmomath.NewInt(2000000)),
			ExpectedTick:     30095166,

			ExpectedSqrtPrice:                         osmomath.MustNewBigDecFromStr("63.993489023888951975210711246458277671"),
			ExpectedLowerTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedUpperTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 315000, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice:                             osmomath.NewDec(4000),
			NewUpperPrice:                             osmomath.NewDec(4545),
		},
		//  Consecutive price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//             5500 ----------- 6250
		//
		"two positions with consecutive price ranges: usdc (in) -> eth (out) | ofz": {
			TokenOut:                 sdk.NewCoin(ETH, osmomath.NewInt(1820630)),
			TokenInDenom:             USDC,
			PriceLimit:               osmomath.NewBigDec(6106),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(5500), // 315000
			SecondPositionUpperPrice: osmomath.NewDec(6250), // 322500
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// # Swap step 1
			// token_out = Decimal("1820630")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441")
			// sqrt_next_1 = Decimal("74.161984870956629488")

			// token_out_1 = round_sdk_prec_down(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, True))

			// token_out = token_out - token_out_1

			// # Swap step 2
			// liq_2 = Decimal("1197767444.955508123483846019")
			// sqrt_next_2 = get_next_sqrt_price_from_amount0_out_round_up(liq_2, sqrt_next_1, token_out)

			// token_out_2 = round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, True))

			// # Summary:
			// token_in = token_in_1 + token_in_2
			// spread_rewards_growth = spread_factor_1 / liq_1 + spread_factor_2 / liq_2
			// print(sqrt_next_2)
			// print(token_in)
			// print(token_out_2)
			ExpectedTokenOut:  sdk.NewCoin(ETH, osmomath.NewInt(1820630)),
			ExpectedTokenIn:   sdk.NewCoin(USDC, osmomath.NewInt(9999999570)),
			ExpectedTick:      32105414,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("78.137148837036751554352224945360339905"),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 315000, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(5500),
			NewUpperPrice: osmomath.NewDec(6250),
		},
		//  Partially overlapping price ranges
		//
		//                5000
		//        4545 -----|----- 5500
		//  4000 ----------- 4999
		//
		"two positions with partially overlapping price ranges: eth (in) -> usdc (out) | zfo": {
			TokenOut:                 sdk.NewCoin(USDC, osmomath.NewInt(9321276930)),
			TokenInDenom:             ETH,
			PriceLimit:               osmomath.NewBigDec(4128),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(4000),
			SecondPositionUpperPrice: osmomath.NewDec(4999),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// # Swap step 1
			// token_out = Decimal("9321276930")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441")
			// sqrt_next_1 = Decimal("70.703606697254136613")

			// token_out_1 = round_sdk_prec_down(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, True))

			// token_out = token_out - token_out_1

			// # Swap step 2
			// second_pos_liq = Decimal("670416215.718827443660400593")
			// liq_2 = liq_1 + second_pos_liq
			// sqrt_next_2 = Decimal("67.416615162732695594")

			// token_out_2 = round_sdk_prec_down(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, True))

			// token_out = token_out - token_out_2

			// # Swap step 3
			// liq_3 = second_pos_liq
			// sqrt_next_3 = get_next_sqrt_price_from_amount1_out_round_down(liq_3, sqrt_next_2, token_out)

			// token_out_3 = round_sdk_prec_down(calc_amount_one_delta(liq_3, sqrt_next_2, sqrt_next_3, False))
			// token_in_3 = ceil(calc_amount_zero_delta(liq_3, sqrt_next_2, sqrt_next_3, True))

			// # Summary:
			// token_in = token_in_1 + token_in_2
			// spread_rewards_growth = spread_factor_1 / liq_1 + spread_factor_2 / liq_2
			// print(sqrt_next_3)
			// print(token_in)
			// print(token_out_2)
			ExpectedTokenIn:   sdk.NewCoin("eth", osmomath.NewInt(2000000)),
			ExpectedTokenOut:  sdk.NewCoin("usdc", osmomath.NewInt(9321276930)),
			ExpectedTick:      30129083,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("64.257943796086567725876595411582357676"),
			// Started from DefaultSpreadRewardAccumCoins * 3, crossed tick once, thus becoming
			// DefaultSpreadRewardAccumCoins * 3 - DefaultSpreadRewardAccumCoins = DefaultSpreadRewardAccumCoins * 2
			ExpectedLowerTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedUpperTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 300000, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 309990, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(4000),
			NewUpperPrice: osmomath.NewDec(4999),
		},
		//          		5000
		//  		4545 -----|----- 5500
		//  4000 ---------- 4999
		"two positions with partially overlapping price ranges, not utilizing full liquidity of second position: eth (in) -> usdc (out) | zfo": {
			TokenOut:                 sdk.NewCoin(USDC, osmomath.NewInt(8479320318)),
			TokenInDenom:             ETH,
			PriceLimit:               osmomath.NewBigDec(4128),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(4000),
			SecondPositionUpperPrice: osmomath.NewDec(4999),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// # Swap step 1
			// token_out = Decimal("8479320318")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441")
			// sqrt_next_1 = Decimal("70.703606697254136613")

			// token_out_1 = round_sdk_prec_down(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, True))

			// token_out = token_out - token_out_1

			// # Swap step 2
			// second_pos_liq = Decimal("670416215.718827443660400593")
			// liq_2 = liq_1 + second_pos_liq
			// sqrt_next_2 = Decimal("67.416615162732695594")

			// token_out_2 = round_sdk_prec_down(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, True))

			// token_out = token_out - token_out_2

			// # Swap step 3
			// liq_3 = second_pos_liq
			// sqrt_next_3 = get_next_sqrt_price_from_amount1_out_round_down(liq_3, sqrt_next_2, token_out)

			// token_out_3 = round_sdk_prec_down(calc_amount_one_delta(liq_3, sqrt_next_2, sqrt_next_3, False))
			// token_in_3 = ceil(calc_amount_zero_delta(liq_3, sqrt_next_2, sqrt_next_3, True)))

			// # Summary:
			// token_in = token_in_1 + token_in_2
			// print(sqrt_next_3)
			// print(token_in)
			// print(token_out_2)
			ExpectedTokenIn:   sdk.NewCoin(ETH, osmomath.NewInt(1800000)),
			ExpectedTokenOut:  sdk.NewCoin(USDC, osmomath.NewInt(8479320318)),
			ExpectedTick:      30292059,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("65.513815286452064191403749708246274698"),
			// Started from DefaultSpreadRewardAccumCoins * 3, crossed tick once, thus becoming
			// DefaultSpreadRewardAccumCoins * 3 - DefaultSpreadRewardAccumCoins = DefaultSpreadRewardAccumCoins * 2
			ExpectedLowerTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedUpperTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 300000, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 309990, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(4000),
			NewUpperPrice: osmomath.NewDec(4999),
		},
		//  Partially overlapping price ranges

		//          5000
		//  4545 -----|----- 5500
		//        5001 ----------- 6250
		//
		"two positions with partially overlapping price ranges: usdc (in) -> eth (out) | ofz": {
			TokenOut:                 sdk.NewCoin(ETH, osmomath.NewInt(1864161)),
			TokenInDenom:             USDC,
			PriceLimit:               osmomath.NewBigDec(6056),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(5001),
			SecondPositionUpperPrice: osmomath.NewDec(6250),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// # Swap step 1
			// token_out = Decimal("1864161")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441")
			// sqrt_next_1 = Decimal("70.717748832948578243")

			// token_out_1 = round_sdk_prec_down(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, True))

			// token_out = token_out - token_out_1

			// # Swap step 2
			// second_pos_liq = Decimal("670416088.605668727039240782")
			// liq_2 = liq_1 + Decimal("670416088.605668727039240782")
			// sqrt_next_2 = Decimal("74.161984870956629488")

			// token_out_2 = round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, True))

			// token_out = token_out - token_out_2

			// # Swap step 3
			// liq_3 = second_pos_liq
			// sqrt_next_3 = get_next_sqrt_price_from_amount0_out_round_up(liq_3, sqrt_next_2, token_out)

			// token_out_3 = round_sdk_prec_down(calc_amount_zero_delta(liq_3, sqrt_next_2, sqrt_next_3, False))
			// token_in_3 = ceil(calc_amount_one_delta(liq_3, sqrt_next_2, sqrt_next_3, True))

			// # Summary:
			// token_in = token_in_1 + token_in_2
			// print(sqrt_next_3)
			// print(token_in)
			// print(token_out_2)
			ExpectedTokenIn:   sdk.NewCoin(USDC, osmomath.NewInt(9999994688)),
			ExpectedTokenOut:  sdk.NewCoin(ETH, osmomath.NewInt(1864161)),
			ExpectedTick:      32055918,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("77.819781711876553578435870496972242531"),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 310010, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(5001),
			NewUpperPrice: osmomath.NewDec(6250),
		},
		"two positions with partially overlapping price ranges, not utilizing full liquidity of second position: usdc (in) -> eth (out) | ofz": {
			TokenOut:                 sdk.NewCoin(ETH, osmomath.NewInt(1609138)),
			TokenInDenom:             USDC,
			PriceLimit:               osmomath.NewBigDec(6056),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(5001),
			SecondPositionUpperPrice: osmomath.NewDec(6250),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// # Swap step 1
			// token_out = Decimal("1609138")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441")
			// sqrt_next_1 = Decimal("70.717748832948578243")
			// spread_factor = Decimal("0.05")

			// token_out_1 = round_sdk_prec_down(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, True))
			// spread_factor_1 = token_in_1 *  spread_factor / (1 - spread_factor)

			// token_out = token_out - token_out_1

			// # Swap step 2
			// second_pos_liq = Decimal("670416088.605668727039240782")
			// liq_2 = liq_1 + Decimal("670416088.605668727039240782")
			// sqrt_next_2 = Decimal("74.161984870956629488")

			// token_out_2 = round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, True))
			// spread_factor_2 = token_in_2 *  spread_factor / (1 - spread_factor)

			// token_out = token_out - token_out_2

			// # Swap step 3
			// liq_3 = second_pos_liq
			// sqrt_next_3 = get_next_sqrt_price_from_amount0_out_round_up(liq_3, sqrt_next_2, token_out)

			// token_out_3 = round_sdk_prec_down(calc_amount_zero_delta(liq_3, sqrt_next_2, sqrt_next_3, False))
			// token_in_3 = ceil(calc_amount_one_delta(liq_3, sqrt_next_2, sqrt_next_3, True))
			// spread_factor_2 = token_in_3 *  spread_factor / (1 - spread_factor)

			// # Summary:
			// token_in = token_in_1 + token_in_2
			// spread_rewards_growth = spread_factor_1 / liq_1 + spread_factor_2 / liq_2
			// print(sqrt_next_3)
			// print(token_in)
			// print(spread_rewards_growth)
			// print(token_out_2)
			ExpectedTokenIn:  sdk.NewCoin(USDC, osmomath.NewInt(8499999458)),
			ExpectedTokenOut: sdk.NewCoin(ETH, osmomath.NewInt(1609138)),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 310010, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedTick:      31712695,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("75.582372355128594342857800328292876450"),
			NewLowerPrice:     osmomath.NewDec(5001),
			NewUpperPrice:     osmomath.NewDec(6250),
		},
		//  Sequential price ranges with a gap
		//
		//          5000
		//  4545 -----|----- 5500
		//              5501 ----------- 6250
		//
		"two sequential positions with a gap usdc (in) -> eth (out) | ofz": {
			TokenOut:                 sdk.NewCoin(ETH, osmomath.NewInt(1820545)),
			TokenInDenom:             USDC,
			PriceLimit:               osmomath.NewBigDec(6106),
			SpreadFactor:             osmomath.ZeroDec(),
			SecondPositionLowerPrice: osmomath.NewDec(5501), // 315010
			SecondPositionUpperPrice: osmomath.NewDec(6250), // 322500
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision

			// #Range 1: From 5000 to 5500
			// token_out = Decimal("1820545")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("74.161984870956629488") # sqrt5500

			// token_out_1 = round_sdk_prec_down(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, True))

			// token_out = token_out - token_out_1

			// # Range 2: from 5501 till end
			// liq_2 = Decimal("1199528406.187413669481596330")
			// sqrt_cur_2 = Decimal("74.168726563154635304") # sqrt5501
			// sqrt_next_2 = get_next_sqrt_price_from_amount0_out_round_up(liq_2, sqrt_cur_2, token_out)

			// token_out_2 = round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_cur_2, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_one_delta(liq_2, sqrt_cur_2, sqrt_next_2, True))

			// # Summary:
			// token_in = token_in_1 + token_in_2
			// print(sqrt_next_2)
			// print(token_in_2)
			// print(token_out_2)
			ExpectedTokenOut:  sdk.NewCoin(ETH, osmomath.NewInt(1820545)),
			ExpectedTokenIn:   sdk.NewCoin(USDC, osmomath.NewInt(9999994756)),
			ExpectedTick:      32105554,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("78.138050797173647031951910080474560428"),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 315010, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(5501),
			NewUpperPrice: osmomath.NewDec(6250),
		},
		// Slippage protection doesn't cause a failure but interrupts early.
		"single position within one tick, trade completes but slippage protection interrupts trade early: usdc (in) -> eth (out) | ofz": {
			TokenOut:     sdk.NewCoin(ETH, osmomath.NewInt(1820545)),
			TokenInDenom: USDC,
			PriceLimit:   osmomath.NewBigDec(5002),
			SpreadFactor: osmomath.ZeroDec(),
			// from math import *
			// from decimal import *
			// # Range 1: From 5000 to 5002
			// token_out = Decimal("1820545")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("70.724818840347693040") # sqrt5002

			// token_out_1 = liq_1 * (sqrt_next_1 - sqrt_cur ) / (sqrt_next_1 * sqrt_cur)
			// token_in_1 = ceil(liq_1 * (sqrt_next_1 - sqrt_cur ))

			// # Summary:
			// print(sqrt_next_1)
			// print(token_in_1)
			ExpectedTokenOut: sdk.NewCoin(ETH, osmomath.NewInt(4291)),
			ExpectedTokenIn:  sdk.NewCoin(USDC, osmomath.NewInt(21463952)),
			ExpectedTick:     31002000,
			// Since we know we're going up to the price limit, we can calculate the sqrt price exactly.
			ExpectedSqrtPrice: osmomath.BigDecFromDec(osmomath.MustMonotonicSqrt(osmomath.NewDec(5002))),
		},
	}

	swapInGivenOutSpreadRewardTestCases = map[string]apptesting.ConcentratedSwapTest{
		"spread factor 1: single position within one tick: eth (in) -> usdc (out) (1% spread factor) | zfo": {
			TokenOut:     sdk.NewCoin(USDC, osmomath.NewInt(42000000)),
			TokenInDenom: ETH,
			PriceLimit:   osmomath.NewBigDec(4993),
			SpreadFactor: osmomath.MustNewDecFromStr("0.01"),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')

			// token_out = Decimal("42000000")
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// spread_factor = Decimal("0.01")

			// sqrt_next = get_next_sqrt_price_from_amount1_out_round_down(liq, sqrt_cur, token_out)
			// token_in = ceil(calc_amount_zero_delta(liq, sqrt_cur, sqrt_next, True))
			// spread_factor = token_in * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor = spread_factor.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_scaled = spread_factor * Decimal('1e27')

			// # Summary:
			// token_in = ceil(token_in + spread_factor)
			// spread_rewards_growth_scaled = spread_factor_scaled / liq
			// print(sqrt_next)
			// print(token_in)
			// print(spread_rewards_growth_scaled)
			ExpectedTokenOut: sdk.NewCoin(USDC, osmomath.NewInt(42000000)),
			ExpectedTokenIn:  sdk.NewCoin(ETH, osmomath.NewInt(8489)),
			ExpectedTick:     30996087,
			// This value is the direct output of sqrt_next in the script above.
			// The precision is exact because we properly handle rounding behavior in intermediate steps.
			ExpectedSqrtPrice:                          osmomath.MustNewBigDecFromStr("70.683007989825007163485199999996399373"),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("55925868851654481727.579591913860572604"),
		},
		"spread factor 2: two positions within one tick: usdc (in) -> eth (out) (3% spread factor) | ofz": {
			TokenOut:                 sdk.NewCoin(ETH, osmomath.NewInt(8398)),
			TokenInDenom:             USDC,
			PriceLimit:               osmomath.NewBigDec(5020),
			SpreadFactor:             osmomath.MustNewDecFromStr("0.03"),
			SecondPositionLowerPrice: DefaultLowerPrice,
			SecondPositionUpperPrice: DefaultUpperPrice,
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')

			// token_out = Decimal("8398")
			// liq = Decimal("1517882343.751510417627556287") * 2
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// spread_factor = Decimal("0.03")

			// sqrt_next = get_next_sqrt_price_from_amount0_out_round_up(liq, sqrt_cur, token_out)
			// token_in = ceil(calc_amount_one_delta(liq, sqrt_cur, sqrt_next, True))
			// spread_factor = token_in * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor = spread_factor.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_scaled = spread_factor * Decimal('1e27')

			// # Summary:
			// token_in = ceil(token_in + spread_factor)
			// spread_rewards_growth = spread_factor_scaled / liq
			// print(sqrt_next)
			// print(token_in)
			// print(spread_rewards_growth)
			ExpectedTokenOut:  sdk.NewCoin(ETH, osmomath.NewInt(8398)),
			ExpectedTokenIn:   sdk.NewCoin(USDC, osmomath.NewInt(43297130)),
			ExpectedTick:      31001956,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.724512595179305566327821490232558005"),
			// two positions with same liquidity entered
			PoolLiqAmount0: osmomath.NewInt(1000000).MulRaw(2),
			PoolLiqAmount1: osmomath.NewInt(5000000000).MulRaw(2),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("427870415073442341437143.58456074189761953"),
		},
		"spread factor 3: two positions with consecutive price ranges: usdc (in) -> eth (out) (0.1% spread factor) | ofz": {
			TokenOut:                 sdk.NewCoin(ETH, osmomath.NewInt(1820630)),
			TokenInDenom:             USDC,
			PriceLimit:               osmomath.NewBigDec(6106),
			SpreadFactor:             osmomath.MustNewDecFromStr("0.001"),
			SecondPositionLowerPrice: osmomath.NewDec(5500), // 315000
			SecondPositionUpperPrice: osmomath.NewDec(6250), // 322500
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')

			// # Range 1: From 5000 to 5500
			// token_out = Decimal("1820630")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt5000 = Decimal("70.710678118654752441")
			// sqrt5500 = Decimal("74.161984870956629488")
			// sqrt_cur = sqrt5000
			// sqrt_next_1 = sqrt5500
			// spread_factor = Decimal("0.001")

			// token_out_1 = round_sdk_prec_down(calc_amount_zero_delta(liq_1, sqrt_next_1, sqrt_cur, False))
			// token_in_1 = ceil(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, True))
			// spread_factor_1 = token_in_1 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_1 = spread_factor_1.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_1_scaled = spread_factor_1 * Decimal('1e27')

			// token_out = token_out - token_out_1

			// # Range 2: from 5500 till end
			// # Using clmath.py scripts: get_liquidity_from_amounts(DefaultCurrSqrtPrice, sqrt5500, sqrt6250, DefaultPoolLiq0, DefaultPoolLiq1)
			// liq_2 = Decimal("1197767444.955508123483846019")
			// sqrt_next_2 = get_next_sqrt_price_from_amount0_out_round_up(liq_2, sqrt_next_1, token_out)

			// token_out_2 = round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, True))
			// spread_factor_2 = token_in_2 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_2 = spread_factor_2.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_2_scaled = spread_factor_2 * Decimal('1e27')

			// # Summary:
			// token_in = ceil(token_in_1 + spread_factor_1 + token_in_2 + spread_factor_2)
			// spread_rewards_growth = spread_factor_1_scaled / liq_1 + spread_factor_2_scaled / liq_2
			// print(sqrt_next_2)
			// print(token_in)
			// print(spread_rewards_growth)
			ExpectedTokenOut:  sdk.NewCoin(ETH, osmomath.NewInt(1820630)),
			ExpectedTokenIn:   sdk.NewCoin(USDC, osmomath.NewInt(10010009580)),
			ExpectedTick:      32105414,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("78.137148837036751554352224945360339905"),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 315000, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(5500),
			NewUpperPrice: osmomath.NewDec(6250),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("7433904623597259879249693.52783035536310194"),
		},
		"spread factor 4: two positions with partially overlapping price ranges: eth (in) -> usdc (out) (10% spread factor) | zfo": {
			TokenOut:                 sdk.NewCoin(USDC, osmomath.NewInt(9321276930)),
			TokenInDenom:             ETH,
			PriceLimit:               osmomath.NewBigDec(4128),
			SpreadFactor:             osmomath.MustNewDecFromStr("0.1"),
			SecondPositionLowerPrice: osmomath.NewDec(4000),
			SecondPositionUpperPrice: osmomath.NewDec(4999),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')

			// # Swap step 1
			// token_out = Decimal("9321276930")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441")
			// sqrt_next_1 = Decimal("70.703606697254136613")
			// spread_factor = Decimal("0.1")

			// token_out_1 = round_sdk_prec_down(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, True))
			// spread_factor_1 = token_in_1 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_1 = spread_factor_1.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_1_scaled = spread_factor_1 * Decimal('1e27')

			// token_out = token_out - token_out_1

			// # Swap step 2
			// second_pos_liq = Decimal("670416215.718827443660400593")
			// liq_2 = liq_1 + second_pos_liq
			// sqrt_next_2 = Decimal("67.416615162732695594")

			// token_out_2 = round_sdk_prec_down(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, True))
			// spread_factor_2 = token_in_2 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_2 = spread_factor_2.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_2_scaled = spread_factor_2 * Decimal('1e27')

			// token_out = token_out - token_out_2

			// # Swap step 3
			// liq_3 = second_pos_liq
			// sqrt_next_3 = get_next_sqrt_price_from_amount1_out_round_down(liq_3, sqrt_next_2, token_out)

			// token_out_3 = round_sdk_prec_down(calc_amount_one_delta(liq_3, sqrt_next_2, sqrt_next_3, False))
			// token_in_3 = ceil(calc_amount_zero_delta(liq_3, sqrt_next_2, sqrt_next_3, True))
			// spread_factor_3 = token_in_3 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_3 = spread_factor_3.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_3_scaled = spread_factor_3 * Decimal('1e27')

			// # Summary:
			// token_in = ceil(token_in_1 + spread_factor_1 + token_in_2 + spread_factor_2 + token_in_3 + spread_factor_3)
			// spread_rewards_growth = spread_factor_1_scaled / liq_1 + spread_factor_2_scaled / liq_2 + spread_factor_3_scaled / liq_3
			// print(sqrt_next_3)
			// print(token_in)
			// print(spread_rewards_growth)
			// print(token_out_1 + token_out_2 + token_out_3)
			ExpectedTokenIn:   sdk.NewCoin("eth", osmomath.NewInt(2222223)),
			ExpectedTokenOut:  sdk.NewCoin("usdc", osmomath.NewInt(9321276930)),
			ExpectedTick:      30129083,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("64.257943796086567725876595411582357676"),
			// Started from DefaultSpreadRewardAccumCoins * 3, crossed tick once, thus becoming
			// DefaultSpreadRewardAccumCoins * 3 - DefaultSpreadRewardAccumCoins = DefaultSpreadRewardAccumCoins * 2
			ExpectedLowerTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedUpperTickSpreadRewardGrowth:       DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 300000, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 309990, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(4000),
			NewUpperPrice: osmomath.NewDec(4999),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("157793641388331455795962.112193571559492696"),
		},
		"spread factor 5: two positions with partially overlapping price ranges, not utilizing full liquidity of second position: usdc (in) -> eth (out) (5% spread factor) | ofz": {
			TokenOut:                 sdk.NewCoin(ETH, osmomath.NewInt(1609138)),
			TokenInDenom:             USDC,
			PriceLimit:               osmomath.NewBigDec(6056),
			SpreadFactor:             osmomath.MustNewDecFromStr("0.05"),
			SecondPositionLowerPrice: osmomath.NewDec(5001),
			SecondPositionUpperPrice: osmomath.NewDec(6250),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')

			// # Swap step 1
			// token_out = Decimal("1609138")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441")
			// sqrt_next_1 = Decimal("70.717748832948578243")
			// spread_factor = Decimal("0.05")

			// token_out_1 = round_sdk_prec_down(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, True))
			// spread_factor_1 = token_in_1 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_1 = spread_factor_1.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_1_scaled = spread_factor_1 * Decimal('1e27')

			// token_out = token_out - token_out_1

			// # Swap step 2
			// second_pos_liq = Decimal("670416088.605668727039240782")
			// liq_2 = liq_1 + Decimal("670416088.605668727039240782")
			// sqrt_next_2 = Decimal("74.161984870956629488")

			// token_out_2 = round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_next_1, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_one_delta(liq_2, sqrt_next_1, sqrt_next_2, True))
			// spread_factor_2 = token_in_2 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_2 = spread_factor_2.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_2_scaled = spread_factor_2 * Decimal('1e27')

			// token_out = token_out - token_out_2

			// # Swap step 3
			// liq_3 = second_pos_liq
			// sqrt_next_3 = get_next_sqrt_price_from_amount0_out_round_up(liq_3, sqrt_next_2, token_out)

			// token_out_3 = round_sdk_prec_down(calc_amount_zero_delta(liq_3, sqrt_next_2, sqrt_next_3, False))
			// token_in_3 = ceil(calc_amount_one_delta(liq_3, sqrt_next_2, sqrt_next_3, True))
			// spread_factor_3 = token_in_3 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_3 = spread_factor_3.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_3_scaled = spread_factor_3 * Decimal('1e27')

			// # Summary:
			// token_in = ceil(token_in_1 + spread_factor_1 + token_in_2 + spread_factor_2 + token_in_3 + spread_factor_3)
			// spread_rewards_growth = spread_factor_1_scaled / liq_1 + spread_factor_2_scaled / liq_2 + spread_factor_3_scaled / liq_3
			// print(sqrt_next_3)
			// print(token_in)
			// print(spread_rewards_growth)
			// print(token_out_1 + token_out_2 + token_out_3)
			ExpectedTokenIn:  sdk.NewCoin(USDC, osmomath.NewInt(8947367851)),
			ExpectedTokenOut: sdk.NewCoin(ETH, osmomath.NewInt(1609138)),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 310010, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedTick:      31712695,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("75.582372355128594342857800328292876450"),
			NewLowerPrice:     osmomath.NewDec(5001),
			NewUpperPrice:     osmomath.NewDec(6250),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("256404959888119534778711315.330259108756531997"),
		},
		"spread factor 6: two sequential positions with a gap usdc (in) -> eth (out) (0.03% spread factor) | ofz": {
			TokenOut:                 sdk.NewCoin(ETH, osmomath.NewInt(1820545)),
			TokenInDenom:             USDC,
			PriceLimit:               osmomath.NewBigDec(6106),
			SpreadFactor:             osmomath.MustNewDecFromStr("0.0003"),
			SecondPositionLowerPrice: osmomath.NewDec(5501), // 315010
			SecondPositionUpperPrice: osmomath.NewDec(6250), // 322500
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')

			// # Swap step 1
			// token_out = Decimal("1820545")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441")
			// sqrt_next_1 = Decimal("74.161984870956629488")
			// spread_factor = Decimal("0.0003")

			// token_out_1 = round_sdk_prec_down(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, True))
			// spread_factor_1 = token_in_1 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_1 = spread_factor_1.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_1_scaled = spread_factor_1 * Decimal('1e27')

			// token_out = token_out - token_out_1

			// # Swap step 2
			// sqrt_cur_2 = Decimal("74.168726563154635304")
			// liq_2 = Decimal("1199528406.187413669481596330")
			// sqrt_next_2 = get_next_sqrt_price_from_amount0_out_round_up(liq_2, sqrt_cur_2, token_out)

			// token_out_2 = round_sdk_prec_down(calc_amount_zero_delta(liq_2, sqrt_cur_2, sqrt_next_2, False))
			// token_in_2 = ceil(calc_amount_one_delta(liq_2, sqrt_cur_2 , sqrt_next_2, True))
			// spread_factor_2 = token_in_2 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_2 = spread_factor_2.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_2_scaled = spread_factor_2 * Decimal('1e27')

			// # Summary:
			// token_in = ceil(token_in_1 + spread_factor_1 + token_in_2 + spread_factor_2)
			// spread_rewards_growth = spread_factor_1_scaled / liq_1 + spread_factor_2_scaled / liq_2
			// print(sqrt_next_2)
			// print(token_in)
			// print(spread_rewards_growth)
			// print(token_out_1 + token_out_2)
			ExpectedTokenOut:  sdk.NewCoin(ETH, osmomath.NewInt(1820545)),
			ExpectedTokenIn:   sdk.NewCoin(USDC, osmomath.NewInt(10002995655)),
			ExpectedTick:      32105554,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("78.138050797173647031951910080474560428"),
			ExpectedSecondLowerTickSpreadRewardGrowth: secondPosition{TickIndex: 315010, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: secondPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: cl.EmptyCoins},
			NewLowerPrice: osmomath.NewDec(5501),
			NewUpperPrice: osmomath.NewDec(6250),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("2226857353494147354437463.405839285844866989"),
		},
		"spread factor 7: single position within one tick, trade completes but slippage protection interrupts trade early: usdc (in) -> eth (out) (1% spread factor) | ofz": {
			TokenOut:     sdk.NewCoin(ETH, osmomath.NewInt(1820545)),
			TokenInDenom: USDC,
			PriceLimit:   osmomath.NewBigDec(5002),
			SpreadFactor: osmomath.MustNewDecFromStr("0.01"),
			// from math import *
			// from decimal import *

			// import sys

			// # import custom CL script
			// sys.path.insert(0, 'x/concentrated-liquidity/python')
			// from clmath import *

			// getcontext().prec = 60
			// precision = Decimal('1.000000000000000000000000000000000000') # 36 decimal precision
			// eighteen_decimal_precision = Decimal('1.000000000000000000')

			// # Range 1: From 5000 to 5002
			// token_out = Decimal("1820545")
			// liq_1 = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// sqrt_next_1 = Decimal("5002").sqrt()
			// spread_factor = Decimal("0.01")

			// token_out_1 =  round_sdk_prec_down(calc_amount_zero_delta(liq_1, sqrt_cur, sqrt_next_1, False))
			// token_in_1 = ceil(calc_amount_one_delta(liq_1, sqrt_cur, sqrt_next_1, True))
			// spread_factor_1 = token_in_1 * round_sdk_prec_up(spread_factor / (1 - spread_factor))
			// spread_factor_1 = spread_factor_1.quantize(eighteen_decimal_precision, ROUND_UP)
			// spread_factor_1_scaled = spread_factor_1 * Decimal('1e27')

			// # Summary:
			// token_in = ceil(token_in_1 + spread_factor_1)
			// spread_rewards_growth = spread_factor_1_scaled / liq_1
			// print(sqrt_next_1)
			// print(token_in)
			// print(spread_rewards_growth)
			// print(token_out_1)
			ExpectedTokenOut:  sdk.NewCoin(ETH, osmomath.NewInt(4291)),
			ExpectedTokenIn:   sdk.NewCoin(USDC, osmomath.NewInt(21680760)),
			ExpectedTick:      31002000,
			ExpectedSqrtPrice: osmomath.MustNewBigDecFromStr("70.724818840347693040"),
			ExpectedSpreadRewardGrowthAccumulatorValue: osmomath.MustNewDecFromStr("142835574082604345119662.712639063715501632"),
		},
	}

	swapInGivenOutErrorTestCases = map[string]apptesting.ConcentratedSwapTest{
		"single position within one tick, trade does not complete due to lack of liquidity: usdc -> eth ": {
			TokenOut:     sdk.NewCoin("usdc", osmomath.NewInt(5300000000)),
			TokenInDenom: "eth",
			PriceLimit:   osmomath.NewBigDec(6000),
			SpreadFactor: osmomath.ZeroDec(),
			ExpectErr:    true,
		},
		"single position within one tick, trade does not complete due to lack of liquidity: eth -> usdc ": {
			TokenOut:     sdk.NewCoin("eth", osmomath.NewInt(1100000)),
			TokenInDenom: "usdc",
			PriceLimit:   osmomath.NewBigDec(4000),
			SpreadFactor: osmomath.ZeroDec(),
			ExpectErr:    true,
		},
	}

	additiveSpreadRewardGrowthGlobalErrTolerance = osmomath.ErrTolerance{
		// 2 * 10^-18
		AdditiveTolerance: osmomath.SmallestDec().Mul(osmomath.NewDec(2)),
	}
)

func init() {
	populateSwapTestFields(swapInGivenOutTestCases)
	populateSwapTestFields(swapOutGivenInCases)
	populateSwapTestFields(swapOutGivenInErrorCases)
	populateSwapTestFields(swapInGivenOutErrorTestCases)
}

func populateSwapTestFields(cases map[string]apptesting.ConcentratedSwapTest) {
	for k, v := range cases {
		if v.ExpectedLowerTickSpreadRewardGrowth == nil {
			v.ExpectedLowerTickSpreadRewardGrowth = DefaultSpreadRewardAccumCoins
		}
		if v.ExpectedUpperTickSpreadRewardGrowth == nil {
			v.ExpectedUpperTickSpreadRewardGrowth = DefaultSpreadRewardAccumCoins
		}
		cases[k] = v
	}
}

func (s *KeeperTestSuite) preparePoolAndDefaultPosition() types.ConcentratedPoolExtension {
	pool := s.PrepareConcentratedPool()
	s.SetupDefaultPosition(pool.GetId())
	return pool
}

func (s *KeeperTestSuite) preparePoolAndDefaultPositions(test apptesting.ConcentratedSwapTest) types.ConcentratedPoolExtension {
	pool := s.preparePoolAndDefaultPosition()
	s.SetupSecondPosition(test, pool)
	pool, _ = s.Clk.GetConcentratedPoolById(s.Ctx, pool.GetId())
	return pool
}

func makeTests[T any](tests ...map[string]T) map[string]T {
	length := 0
	for i := range tests {
		length += len(tests[i])
	}
	retTests := make(map[string]T, length)
	for _, tt := range tests {
		for name, test := range tt {
			retTests[name] = test
		}
	}
	return retTests
}

func (s *KeeperTestSuite) assertPoolNotModified(poolBeforeCalc types.ConcentratedPoolExtension) {
	poolAfterCalc, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolBeforeCalc.GetId())
	s.Require().NoError(err)
	s.Require().Equal(poolBeforeCalc.GetCurrentSqrtPrice(), poolAfterCalc.GetCurrentSqrtPrice())
	s.Require().Equal(poolBeforeCalc.GetCurrentTick(), poolAfterCalc.GetCurrentTick())
	s.Require().Equal(poolBeforeCalc.GetLiquidity(), poolAfterCalc.GetLiquidity())
	s.Require().Equal(poolBeforeCalc.GetTickSpacing(), poolAfterCalc.GetTickSpacing())
}

func (s *KeeperTestSuite) assertSpreadRewardAccum(test apptesting.ConcentratedSwapTest, poolId uint64) {
	spreadRewardAccum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, poolId)
	s.Require().NoError(err)

	spreadRewardAccumValue := spreadRewardAccum.GetValue()
	if test.ExpectedSpreadRewardGrowthAccumulatorValue.IsNil() {
		s.Require().Equal(0, spreadRewardAccumValue.Len())
		return
	}
	amountOfDenom := test.TokenIn.Denom
	if amountOfDenom == "" {
		amountOfDenom = test.TokenInDenom
	}
	spreadRewardVal := spreadRewardAccumValue.AmountOf(amountOfDenom)
	s.Require().Equal(1, spreadRewardAccumValue.Len(), "spread reward accumulator should only have one denom, was (%s)", spreadRewardAccumValue)
	osmoassert.Equal(
		s.T(),
		additiveSpreadRewardGrowthGlobalErrTolerance,
		osmomath.BigDecFromDec(test.ExpectedSpreadRewardGrowthAccumulatorValue),
		osmomath.BigDecFromDec(spreadRewardVal),
	)
}

func (s *KeeperTestSuite) assertZeroSpreadRewards(poolId uint64) {
	spreadRewardAccum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, poolId)
	s.Require().NoError(err)
	s.Require().Equal(0, spreadRewardAccum.GetValue().Len())
}

func (s *KeeperTestSuite) getExpectedLiquidity(test apptesting.ConcentratedSwapTest, pool types.ConcentratedPoolExtension) osmomath.Dec {
	if test.NewLowerPrice.IsNil() && test.NewUpperPrice.IsNil() {
		test.NewLowerPrice = DefaultLowerPrice
		test.NewUpperPrice = DefaultUpperPrice
	}

	newLowerTick, newUpperTick := s.lowerUpperPricesToTick(test.NewLowerPrice, test.NewUpperPrice, pool.GetTickSpacing())

	lowerSqrtPrice, err := math.TickToSqrtPrice(newLowerTick)
	s.Require().NoError(err)
	upperSqrtPrice, err := math.TickToSqrtPrice(newUpperTick)
	s.Require().NoError(err)

	if test.PoolLiqAmount0.IsNil() && test.PoolLiqAmount1.IsNil() {
		test.PoolLiqAmount0 = DefaultAmt0
		test.PoolLiqAmount1 = DefaultAmt1
	}

	expectedLiquidity := math.GetLiquidityFromAmounts(DefaultCurrSqrtPrice, lowerSqrtPrice, upperSqrtPrice, test.PoolLiqAmount0, test.PoolLiqAmount1)
	return expectedLiquidity
}

func (s *KeeperTestSuite) lowerUpperPricesToTick(lowerPrice, upperPrice osmomath.Dec, tickSpacing uint64) (int64, int64) {
	lowerSqrtPrice := osmomath.MustMonotonicSqrt(lowerPrice)
	newLowerTick, err := clmath.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(lowerSqrtPrice), tickSpacing)
	s.Require().NoError(err)
	upperSqrtPrice := osmomath.MustMonotonicSqrt(upperPrice)
	newUpperTick, err := clmath.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(upperSqrtPrice), tickSpacing)
	s.Require().NoError(err)
	return newLowerTick, newUpperTick
}

func (s *KeeperTestSuite) TestComputeAndSwapOutAmtGivenIn() {
	// add error cases as well
	tests := makeTests(swapOutGivenInCases, swapOutGivenInErrorCases)

	for name, test := range tests {
		test := test
		s.Run(name, func() {
			s.SetupAndFundSwapTest()
			pool := s.PreparePoolWithCustSpread(test.SpreadFactor)
			// add default position
			s.SetupDefaultPosition(pool.GetId())
			s.SetupSecondPosition(test, pool)

			poolBeforeCalc, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			// Refetch the pool
			pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			// perform compute
			cacheCtx, _ := s.Ctx.CacheContext()
			swapResult, poolUpdates, err := s.App.ConcentratedLiquidityKeeper.ComputeOutAmtGivenIn(
				cacheCtx,
				pool.GetId(),
				test.TokenIn, test.TokenOutDenom,
				test.SpreadFactor, test.PriceLimit, true)

			if test.ExpectErr {
				s.Require().Error(err)
			} else {
				s.testSwapResult(test, pool, swapResult.AmountIn, swapResult.AmountOut, poolUpdates, err)

				expectedSpreadFactors := swapResult.AmountIn.ToLegacyDec().Mul(pool.GetSpreadFactor(s.Ctx)).TruncateInt()
				s.Require().Equal(expectedSpreadFactors.String(), swapResult.SpreadRewards.TruncateInt().String())

				// check that the pool has not been modified after performing calc
				s.assertPoolNotModified(poolBeforeCalc)
			}

			// perform swap
			tokenIn, tokenOut, poolUpdates, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(
				s.Ctx, s.TestAccs[0], pool,
				test.TokenIn, test.TokenOutDenom,
				test.SpreadFactor, test.PriceLimit,
			)

			if test.ExpectErr {
				s.Require().Error(err)
			} else {
				s.testSwapResult(test, pool, tokenIn.Amount, tokenOut.Amount, poolUpdates, err)
				s.assertSpreadRewardAccum(test, pool.GetId())
			}
		})
	}
}

func (s *KeeperTestSuite) TestSwap_NoPositions() {
	s.SetupTest()
	pool := s.PrepareConcentratedPool()
	// perform swap
	_, _, _, err := s.App.ConcentratedLiquidityKeeper.SwapInAmtGivenOut(
		s.Ctx, s.TestAccs[0], pool,
		DefaultCoin0, DefaultCoin1.Denom,
		osmomath.ZeroDec(), osmomath.ZeroBigDec(),
	)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.NoSpotPriceWhenNoLiquidityError{PoolId: pool.GetId()})

	_, _, _, err = s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(
		s.Ctx, s.TestAccs[0], pool,
		DefaultCoin0, DefaultCoin1.Denom,
		osmomath.ZeroDec(), osmomath.ZeroBigDec(),
	)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.NoSpotPriceWhenNoLiquidityError{PoolId: pool.GetId()})
}

func (s *KeeperTestSuite) TestSwapOutAmtGivenIn_TickUpdates() {
	tests := makeTests(swapOutGivenInCases)
	for name, test := range tests {
		test := test
		s.Run(name, func() {
			s.SetupAndFundSwapTest()

			// Create default CL pool
			pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, osmomath.ZeroDec())

			// manually update spread factor accumulator for the pool
			spreadFactorAccum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, 1)
			s.Require().NoError(err)
			spreadFactorAccum.AddToAccumulator(DefaultSpreadRewardAccumCoins)

			// add default position
			s.SetupDefaultPosition(pool.GetId())
			s.SetupSecondPosition(test, pool)

			// add 2*DefaultSpreadRewardAccumCoins to spread factor accumulator, now spread factor accumulator has 3*DefaultSpreadRewardAccumCoins as its value
			spreadFactorAccum, err = s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, 1)
			s.Require().NoError(err)
			spreadFactorAccum.AddToAccumulator(DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)))

			// perform swap
			_, _, _, err = s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(
				s.Ctx, s.TestAccs[0], pool,
				test.TokenIn, test.TokenOutDenom,
				test.SpreadFactor, test.PriceLimit)

			s.Require().NoError(err)

			// check lower tick and upper tick spread reward growth
			lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, pool.GetId(), DefaultLowerTick)
			s.Require().NoError(err)
			s.Require().Equal(test.ExpectedLowerTickSpreadRewardGrowth, lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal)

			upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, pool.GetId(), DefaultLowerTick)
			s.Require().NoError(err)
			s.Require().Equal(test.ExpectedUpperTickSpreadRewardGrowth, upperTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal)

			if test.ExpectedSecondLowerTickSpreadRewardGrowth.ExpectedSpreadRewardGrowth != nil {
				newTickIndex := test.ExpectedSecondLowerTickSpreadRewardGrowth.TickIndex
				expectedSpreadRewardGrowth := test.ExpectedSecondLowerTickSpreadRewardGrowth.ExpectedSpreadRewardGrowth

				newLowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, pool.GetId(), newTickIndex)
				s.Require().NoError(err)
				s.Require().Equal(expectedSpreadRewardGrowth, newLowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal)
			}

			if test.ExpectedSecondUpperTickSpreadRewardGrowth.ExpectedSpreadRewardGrowth != nil {
				newTickIndex := test.ExpectedSecondUpperTickSpreadRewardGrowth.TickIndex
				expectedSpreadRewardGrowth := test.ExpectedSecondUpperTickSpreadRewardGrowth.ExpectedSpreadRewardGrowth

				newLowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, pool.GetId(), newTickIndex)
				s.Require().NoError(err)
				s.Require().Equal(expectedSpreadRewardGrowth, newLowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal)
			}
		})
	}
}
func (s *KeeperTestSuite) testSwapResult(test apptesting.ConcentratedSwapTest, pool types.ConcentratedPoolExtension, amountIn, amountOut osmomath.Int, poolUpdates cl.PoolUpdates, err error) {
	s.Require().NoError(err)

	// check that tokenIn, tokenOut, tick, and sqrtPrice from CalcOut are all what we expected
	s.Require().Equal(test.ExpectedSqrtPrice, poolUpdates.NewSqrtPrice, "resultant sqrt price not equal to expected sqrt price")
	s.Require().Equal(test.ExpectedTokenOut.Amount.String(), amountOut.String())
	s.Require().Equal(test.ExpectedTokenIn.Amount.String(), amountIn.String())
	s.Require().Equal(test.ExpectedTick, poolUpdates.NewCurrentTick)

	expectedLiquidity := s.getExpectedLiquidity(test, pool)
	s.Require().Equal(expectedLiquidity.String(), poolUpdates.NewLiquidity.String())
}

func (s *KeeperTestSuite) TestComputeAndSwapInAmtGivenOut() {
	// add error cases as well
	tests := makeTests(swapInGivenOutTestCases, swapInGivenOutSpreadRewardTestCases, swapInGivenOutErrorTestCases)
	for name, test := range tests {
		test := test
		s.Run(name, func() {
			s.SetupAndFundSwapTest()
			pool := s.PreparePoolWithCustSpread(test.SpreadFactor)
			// add default position
			s.SetupDefaultPosition(pool.GetId())
			s.SetupSecondPosition(test, pool)

			// perform compute
			cacheCtx, _ := s.Ctx.CacheContext()
			swapResult, poolUpdates, err := s.App.ConcentratedLiquidityKeeper.ComputeInAmtGivenOut(
				cacheCtx,
				test.TokenOut, test.TokenInDenom,
				test.SpreadFactor, test.PriceLimit, pool.GetId(), true)
			if test.ExpectErr {
				s.Require().Error(err)
			} else {
				s.testSwapResult(test, pool, swapResult.AmountIn, swapResult.AmountOut, poolUpdates, err)

				expectedSpreadRewards := swapResult.AmountIn.ToLegacyDec().Mul(pool.GetSpreadFactor(s.Ctx)).TruncateInt()
				s.Require().Equal(expectedSpreadRewards.String(), swapResult.SpreadRewards.TruncateInt().String())
			}

			// perform swap
			tokenIn, tokenOut, poolUpdates, err := s.App.ConcentratedLiquidityKeeper.SwapInAmtGivenOut(
				s.Ctx, s.TestAccs[0], pool,
				test.TokenOut, test.TokenInDenom,
				test.SpreadFactor, test.PriceLimit)
			if test.ExpectErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			// check that tokenIn, tokenOut, tick, and sqrtPrice from SwapOut are all what we expected
			s.testSwapResult(test, pool, tokenIn.Amount, tokenOut.Amount, poolUpdates, err)

			// Check variables on pool were set correctly
			// - ensure the pool's currentTick and currentSqrtPrice was updated due to calling a mutative method
			s.Require().Equal(test.ExpectedTick, pool.GetCurrentTick())
			// check that liquidity is what we expected
			expectedLiquidity := s.getExpectedLiquidity(test, pool)
			s.Require().Equal(expectedLiquidity.String(), pool.GetLiquidity().String())

			if test.SpreadFactor.IsZero() {
				s.assertZeroSpreadRewards(pool.GetId())
				return
			}
			s.assertSpreadRewardAccum(test, pool.GetId())
		})
	}
}

func (s *KeeperTestSuite) TestSwapInAmtGivenOut_TickUpdates() {
	tests := makeTests(swapInGivenOutTestCases)
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupAndFundSwapTest()

			// Create default CL pool
			pool := s.PrepareConcentratedPool()

			// manually update spread factor accumulator for the pool
			spreadFactorAccum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, 1)
			s.Require().NoError(err)
			spreadFactorAccum.AddToAccumulator(DefaultSpreadRewardAccumCoins)

			// add default position
			s.SetupDefaultPosition(pool.GetId())
			s.SetupSecondPosition(test, pool)

			// add 2*DefaultSpreadRewardAccumCoins to spread factor accumulator, now spread factor accumulator has 3*DefaultSpreadRewardAccumCoins as its value
			spreadFactorAccum, err = s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, 1)
			s.Require().NoError(err)
			spreadFactorAccum.AddToAccumulator(DefaultSpreadRewardAccumCoins.MulDec(osmomath.NewDec(2)))

			// perform swap
			_, _, _, err = s.App.ConcentratedLiquidityKeeper.SwapInAmtGivenOut(
				s.Ctx, s.TestAccs[0], pool,
				test.TokenOut, test.TokenInDenom,
				test.SpreadFactor, test.PriceLimit)
			s.Require().NoError(err)

			// check lower tick and upper tick spread reward growth
			lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, pool.GetId(), DefaultLowerTick)
			s.Require().NoError(err)
			s.Require().Equal(test.ExpectedLowerTickSpreadRewardGrowth, lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal)

			upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, pool.GetId(), DefaultLowerTick)
			s.Require().NoError(err)
			s.Require().Equal(test.ExpectedUpperTickSpreadRewardGrowth, upperTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal)

			if test.ExpectedSecondLowerTickSpreadRewardGrowth.ExpectedSpreadRewardGrowth != nil {
				newTickIndex := test.ExpectedSecondLowerTickSpreadRewardGrowth.TickIndex
				expectedSpreadRewardGrowth := test.ExpectedSecondLowerTickSpreadRewardGrowth.ExpectedSpreadRewardGrowth

				newLowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, pool.GetId(), newTickIndex)
				s.Require().NoError(err)
				s.Require().Equal(expectedSpreadRewardGrowth, newLowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal)
			}

			if test.ExpectedSecondUpperTickSpreadRewardGrowth.ExpectedSpreadRewardGrowth != nil {
				newTickIndex := test.ExpectedSecondUpperTickSpreadRewardGrowth.TickIndex
				expectedSpreadRewardGrowth := test.ExpectedSecondUpperTickSpreadRewardGrowth.ExpectedSpreadRewardGrowth

				newLowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, pool.GetId(), newTickIndex)
				s.Require().NoError(err)
				s.Require().Equal(expectedSpreadRewardGrowth, newLowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSwapExactAmountIn() {
	type param struct {
		tokenIn           sdk.Coin
		tokenOutDenom     string
		underFundBy       osmomath.Int
		tokenOutMinAmount osmomath.Int
		expectedTokenOut  osmomath.Int
	}

	// liquidity and sqrtPriceCurrent for all tests are:
	// liquidity = 1517882343.751510417627556287
	// sqrtPriceCurrent = 70.710678118654752441 # sqrt5000
	tests := []struct {
		name        string
		param       param
		expectedErr error
	}{
		{
			name: "Proper swap usdc > eth",
			// from math import *
			// from decimal import *
			//
			// liquidity = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_in = 42000000
			//
			// precision = Decimal('1.000000000000000000') # 18 decimal precision
			// sqrt_delta = (token_in / liquidity).quantize(precision, rounding=ROUND_DOWN)
			// sqrt_next = sqrt_cur + sqrt_delta
			//
			// # Round token in up to nearest integer and token out down to nearest integer
			// expectedTokenIn = (liquidity * (sqrt_next - sqrt_cur)).quantize(Decimal('1'), rounding=ROUND_UP)
			// expectedTokenOut = (liquidity * (sqrt_next - sqrt_cur) / (sqrt_next * sqrt_cur)).quantize(Decimal('1'), rounding=ROUND_DOWN)
			//
			// # Summary
			// print(sqrt_next)
			// print(expectedTokenIn)
			// print(expectedTokenOut)
			param: param{
				tokenIn:           sdk.NewCoin(USDC, osmomath.NewInt(42000000)),
				tokenOutDenom:     ETH,
				tokenOutMinAmount: types.MinSpotPrice.RoundInt(),
				expectedTokenOut:  osmomath.NewInt(8396),
			},
		},
		{
			name: "Proper swap eth > usdc",
			// from math import *
			// from decimal import *
			//
			// liquidity = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_in = 13370
			//
			// precision = Decimal('1.000000000000000000') # 18 decimal precision
			// sqrt_next = liquidity * sqrt_cur / (liquidity + token_in * sqrt_cur)
			//
			// # Round token out down to nearest integer
			// expectedTokenOut = (liquidity * (sqrt_cur - sqrt_next)).quantize(Decimal('1'), rounding=ROUND_DOWN)
			//
			// # Summary
			// print(sqrt_next)
			// print(expectedTokenIn)
			// print(expectedTokenOut)
			param: param{
				tokenIn:           sdk.NewCoin(ETH, osmomath.NewInt(13370)),
				tokenOutDenom:     USDC,
				tokenOutMinAmount: types.MinSpotPrice.RoundInt(),
				expectedTokenOut:  osmomath.NewInt(66808388),
			},
		},
		{
			name: "out is lesser than min amount",
			param: param{
				tokenIn:           sdk.NewCoin(USDC, osmomath.NewInt(42000000)),
				tokenOutDenom:     ETH,
				tokenOutMinAmount: osmomath.NewInt(8397),
			},
			expectedErr: &types.AmountLessThanMinError{TokenAmount: osmomath.NewInt(8396), TokenMin: osmomath.NewInt(8397)},
		},
		{
			name: "in and out denom are same",
			param: param{
				tokenIn:           sdk.NewCoin(ETH, osmomath.NewInt(13370)),
				tokenOutDenom:     ETH,
				tokenOutMinAmount: types.MinSpotPrice.RoundInt(),
			},
			expectedErr: &types.DenomDuplicatedError{TokenInDenom: ETH, TokenOutDenom: ETH},
		},
		{
			name: "unknown in denom",
			param: param{
				tokenIn:           sdk.NewCoin("etha", osmomath.NewInt(13370)),
				tokenOutDenom:     ETH,
				tokenOutMinAmount: types.MinSpotPrice.RoundInt(),
			},
			expectedErr: &types.TokenInDenomNotInPoolError{TokenInDenom: "etha"},
		},
		{
			name: "unknown out denom",
			param: param{
				tokenIn:           sdk.NewCoin(ETH, osmomath.NewInt(13370)),
				tokenOutDenom:     "etha",
				tokenOutMinAmount: types.MinSpotPrice.RoundInt(),
			},
			expectedErr: &types.TokenOutDenomNotInPoolError{TokenOutDenom: "etha"},
		},
		{
			name: "insufficient user balance",
			param: param{
				tokenIn:           sdk.NewCoin(USDC, osmomath.NewInt(42000000)),
				tokenOutDenom:     ETH,
				tokenOutMinAmount: types.MinSpotPrice.RoundInt(),
				expectedTokenOut:  osmomath.NewInt(8396),
				underFundBy:       osmomath.OneInt(),
			},
			expectedErr: &types.InsufficientUserBalanceError{},
		},
		{
			name: "calculates zero due to small amount in",
			param: param{
				tokenIn:           sdk.NewCoin(USDC, osmomath.NewInt(1)),
				tokenOutDenom:     ETH,
				tokenOutMinAmount: osmomath.OneInt(),
			},
			expectedErr: &types.InvalidAmountCalculatedError{},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			pool := s.preparePoolAndDefaultPosition()

			// Check the test case to see if we are swapping asset0 for asset1 or vice versa
			asset0 := pool.GetToken0()
			zeroForOne := test.param.tokenIn.Denom == asset0

			// Set mock listener to make sure that is is called when desired.
			s.setListenerMockOnConcentratedLiquidityKeeper()

			// The logic below is to trigger a specific error branch
			// where user does not have enough funds.
			underFundBy := osmomath.ZeroInt()
			if !test.param.underFundBy.IsNil() {
				underFundBy = test.param.underFundBy
			}

			// Fund the account with token in.
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(test.param.tokenIn.SubAmount(underFundBy)))

			// Retrieve pool post position set up
			pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			// Note spot price and gas used prior to swap
			spotPriceBefore := pool.GetCurrentSqrtPrice().PowerInteger(2)
			prevGasConsumed := s.Ctx.GasMeter().GasConsumed()

			// Execute the swap directed in the test case
			tokenOutAmount, err := s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool.(poolmanagertypes.PoolI), test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount, DefaultZeroSpreadFactor)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, test.expectedErr)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.param.expectedTokenOut.String(), tokenOutAmount.String())

			gasConsumedForSwap := s.Ctx.GasMeter().GasConsumed() - prevGasConsumed

			// Check that we consume enough gas that a CL pool swap warrants
			// We consume `types.GasFeeForSwap` directly, so the extra I/O operation mean we end up consuming more.
			s.Require().Greater(gasConsumedForSwap, uint64(types.ConcentratedGasFeeForSwap))

			// Assert events
			s.AssertEventEmitted(s.Ctx, types.TypeEvtTokenSwapped, 1)

			// Retrieve pool again post swap
			pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			spotPriceAfter := pool.GetCurrentSqrtPrice().PowerInteger(2)

			// Ratio of the token out should be between the before spot price and after spot price.
			tradeAvgPrice := tokenOutAmount.ToLegacyDec().Quo(test.param.tokenIn.Amount.ToLegacyDec())

			if zeroForOne {
				s.Require().True(tradeAvgPrice.LT(spotPriceBefore.Dec()))
				s.Require().True(tradeAvgPrice.GT(spotPriceAfter.Dec()))
			} else {
				tradeAvgPrice = osmomath.OneDec().Quo(tradeAvgPrice)
				s.Require().True(tradeAvgPrice.GT(spotPriceBefore.Dec()))
				s.Require().True(tradeAvgPrice.LT(spotPriceAfter.Dec()))
			}

			// Validate that listeners were called the desired number of times
			s.validateListenerCallCount(0, 0, 0, 1)
		})
	}
}

func (s *KeeperTestSuite) TestSwapExactAmountOut() {
	// this is used for the test case with price impact protection
	// to ensure that the balances always have enough funds to cover
	// the swap and trigger the desired error branch
	differenceFromMax := osmomath.OneInt()

	type param struct {
		tokenOut         sdk.Coin
		tokenInDenom     string
		tokenInMaxAmount osmomath.Int
		expectedTokenIn  osmomath.Int
	}

	tests := []struct {
		name        string
		param       param
		expectedErr error
	}{
		{
			name: "Proper swap eth > usdc",

			param: param{
				tokenOut:         sdk.NewCoin(USDC, osmomath.NewInt(42000000)),
				tokenInDenom:     ETH,
				tokenInMaxAmount: types.MaxSpotPrice.RoundInt(),
				// from math import *
				// from decimal import *
				// liq = Decimal("1517882343.751510417627556287")
				// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
				// sqrt_next = sqrt_cur - token_out / liq
				// token_in = math.ceil(liq * (sqrt_cur - sqrt_next) / (sqrt_cur * sqrt_next))
				// print(token_in)
				expectedTokenIn: osmomath.NewInt(8404),
			},
		},
		{
			name: "Proper swap usdc > eth",
			// from math import *
			// from decimal import *
			// liq = Decimal("1517882343.751510417627556287")
			// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
			// token_out = Decimal("13370")
			// sqrt_next = liq * sqrt_cur / (liq - token_out * sqrt_cur)
			// token_in = ceil(liq * abs(sqrt_cur - sqrt_next))
			// print(token_in)
			param: param{
				tokenOut:         sdk.NewCoin(ETH, osmomath.NewInt(13370)),
				tokenInDenom:     USDC,
				tokenInMaxAmount: types.MaxSpotPrice.RoundInt(),
				expectedTokenIn:  osmomath.NewInt(66891663),
			},
		},
		{
			name: "out is more than max amount",
			param: param{
				tokenOut:         sdk.NewCoin(ETH, osmomath.NewInt(13370)),
				tokenInDenom:     USDC,
				tokenInMaxAmount: osmomath.NewInt(66891663).Sub(differenceFromMax),
				expectedTokenIn:  osmomath.NewInt(66891663),
			},
			expectedErr: &types.AmountGreaterThanMaxError{TokenAmount: osmomath.NewInt(66891663), TokenMax: osmomath.NewInt(66891663).Sub(differenceFromMax)},
		},
		{
			name: "insufficient user balance",
			param: param{
				tokenOut:         sdk.NewCoin(ETH, osmomath.NewInt(13370)),
				tokenInDenom:     USDC,
				tokenInMaxAmount: osmomath.NewInt(66891663).Sub(differenceFromMax.Mul(osmomath.NewInt(2))),
				expectedTokenIn:  osmomath.NewInt(66891663),
			},
			expectedErr: &types.InsufficientUserBalanceError{},
		},
		{
			name: "in and out denom are same",
			param: param{
				tokenOut:         sdk.NewCoin(ETH, osmomath.NewInt(13370)),
				tokenInDenom:     ETH,
				tokenInMaxAmount: types.MaxSpotPrice.RoundInt(),
			},
			expectedErr: &types.DenomDuplicatedError{TokenInDenom: ETH, TokenOutDenom: ETH},
		},
		{
			name: "unknown out denom",
			param: param{
				tokenOut:         sdk.NewCoin("etha", osmomath.NewInt(13370)),
				tokenInDenom:     ETH,
				tokenInMaxAmount: types.MaxSpotPrice.RoundInt(),
			},
			expectedErr: &types.TokenOutDenomNotInPoolError{TokenOutDenom: "etha"},
		},
		{
			name: "unknown in denom",
			param: param{
				tokenOut:         sdk.NewCoin(ETH, osmomath.NewInt(13370)),
				tokenInDenom:     "etha",
				tokenInMaxAmount: types.MaxSpotPrice.RoundInt(),
			},
			expectedErr: &types.TokenInDenomNotInPoolError{TokenInDenom: "etha"},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			pool := s.preparePoolAndDefaultPosition()

			// Check the test case to see if we are swapping asset0 for asset1 or vice versa
			asset1 := pool.GetToken1()
			zeroForOne := test.param.tokenOut.Denom == asset1

			// Set mock listener to make sure that is is called when desired.
			s.setListenerMockOnConcentratedLiquidityKeeper()

			// Fund the account with token in.
			// We add differenceFromMax for the test case with price impact protection
			// to ensure that the balances always have enough funds to cover
			// the swap and trigger the desired error branch
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(test.param.tokenInDenom, test.param.tokenInMaxAmount.Add(differenceFromMax))))

			// Retrieve pool post position set up
			pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			// Note spot price and gas used prior to swap
			spotPriceBefore := pool.GetCurrentSqrtPrice().PowerInteger(2)
			prevGasConsumed := s.Ctx.GasMeter().GasConsumed()

			// Execute the swap directed in the test case
			tokenIn, err := s.App.ConcentratedLiquidityKeeper.SwapExactAmountOut(s.Ctx, s.TestAccs[0], pool.(poolmanagertypes.PoolI), test.param.tokenInDenom, test.param.tokenInMaxAmount, test.param.tokenOut, DefaultZeroSpreadFactor)

			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, test.expectedErr)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.param.expectedTokenIn.String(), tokenIn.String())

			gasConsumedForSwap := s.Ctx.GasMeter().GasConsumed() - prevGasConsumed
			// Check that we consume enough gas that a CL pool swap warrants
			// We consume `types.GasFeeForSwap` directly, so the extra I/O operation mean we end up consuming more.
			s.Require().Greater(gasConsumedForSwap, uint64(types.ConcentratedGasFeeForSwap))

			// Assert events
			s.AssertEventEmitted(s.Ctx, types.TypeEvtTokenSwapped, 1)

			// Retrieve pool again post swap
			pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			spotPriceAfter := pool.GetCurrentSqrtPrice().PowerInteger(2)

			// Ratio of the token out should be between the before spot price and after spot price.
			tradeAvgPrice := tokenIn.ToLegacyDec().Quo(test.param.tokenOut.Amount.ToLegacyDec())

			if zeroForOne {
				// token in is token zero, token out is token one
				tradeAvgPrice = osmomath.OneDec().Quo(tradeAvgPrice)
				s.Require().True(tradeAvgPrice.LT(spotPriceBefore.Dec()), fmt.Sprintf("tradeAvgPrice: %s, spotPriceBefore: %s", tradeAvgPrice, spotPriceBefore))
				s.Require().True(tradeAvgPrice.GT(spotPriceAfter.Dec()), fmt.Sprintf("tradeAvgPrice: %s, spotPriceAfter: %s", tradeAvgPrice, spotPriceAfter))
			} else {
				// token in is token one, token out is token zero
				s.Require().True(tradeAvgPrice.GT(spotPriceBefore.Dec()), fmt.Sprintf("tradeAvgPrice: %s, spotPriceBefore: %s", tradeAvgPrice, spotPriceBefore))
				s.Require().True(tradeAvgPrice.LT(spotPriceAfter.Dec()), fmt.Sprintf("tradeAvgPrice: %s, spotPriceAfter: %s", tradeAvgPrice, spotPriceAfter))
			}

			// Validate that listeners were called the desired number of times
			s.validateListenerCallCount(0, 0, 0, 1)
		})
	}
}

// TestComputeOutAmtGivenIn tests that ComputeOutAmtGivenIn successfully performs state changes as expected.
// We expect to only change spread factor accum state, since pool state change is not handled by ComputeOutAmtGivenIn.
func (s *KeeperTestSuite) TestComputeOutAmtGivenIn() {
	// we only use spread factor cases here since write Ctx only takes effect in the spread factor accumulator
	tests := make(map[string]apptesting.ConcentratedSwapTest, len(swapOutGivenInSpreadRewardCases))

	for name, test := range swapOutGivenInSpreadRewardCases {
		tests[name] = test
	}

	for name, test := range tests {
		test := test
		s.Run(name, func() {
			s.SetupAndFundSwapTest()
			poolBeforeCalc := s.preparePoolAndDefaultPositions(test)

			// perform calc
			_, _, err := s.App.ConcentratedLiquidityKeeper.ComputeOutAmtGivenIn(
				s.Ctx,
				poolBeforeCalc.GetId(),
				test.TokenIn, test.TokenOutDenom,
				test.SpreadFactor, test.PriceLimit, true)
			s.Require().NoError(err)

			// check that the pool has not been modified after performing calc
			s.assertPoolNotModified(poolBeforeCalc)
			s.assertSpreadRewardAccum(test, poolBeforeCalc.GetId())
		})
	}
}

// TestCalcOutAmtGivenIn_NonMutative tests that CalcOutAmtGivenIn is non-mutative.
func (s *KeeperTestSuite) TestCalcOutAmtGivenIn_NonMutative() {
	// we only use spread reward cases here since write Ctx only takes effect in the spread reward accumulator
	tests := makeTests(swapOutGivenInSpreadRewardCases)
	for name, test := range tests {
		test := test
		s.Run(name, func() {
			s.SetupAndFundSwapTest()
			poolBeforeCalc := s.preparePoolAndDefaultPositions(test)

			// perform calc
			_, err := s.App.ConcentratedLiquidityKeeper.CalcOutAmtGivenIn(
				s.Ctx,
				poolBeforeCalc,
				test.TokenIn, test.TokenOutDenom,
				test.SpreadFactor)
			s.Require().NoError(err)

			// check that the pool has not been modified after performing calc
			s.assertPoolNotModified(poolBeforeCalc)
			s.assertZeroSpreadRewards(poolBeforeCalc.GetId())
		})
	}
}

func (s *KeeperTestSuite) SetupSecondPosition(test apptesting.ConcentratedSwapTest, pool types.ConcentratedPoolExtension) {
	if !test.SecondPositionLowerPrice.IsNil() {
		newLowerTick, newUpperTick := s.LowerUpperPricesToTick(test.SecondPositionLowerPrice, test.SecondPositionUpperPrice, pool.GetTickSpacing())

		_, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), s.TestAccs[1], DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), newLowerTick, newUpperTick)
		s.Require().NoError(err)
	}
}

// TestCalcInAmtGivenOut_NonMutative tests that CalcInAmtGivenOut is non-mutative.
func (s *KeeperTestSuite) TestCalcInAmtGivenOut_NonMutative() {
	// we only use spread reward cases here since write Ctx only takes effect in the spread reward accumulator
	tests := makeTests(swapOutGivenInSpreadRewardCases)

	for name, test := range tests {
		test := test
		s.Run(name, func() {
			s.SetupAndFundSwapTest()
			poolBeforeCalc := s.preparePoolAndDefaultPositions(test)

			// perform calc
			_, err := s.App.ConcentratedLiquidityKeeper.CalcOutAmtGivenIn(
				s.Ctx,
				poolBeforeCalc,
				test.TokenIn, test.TokenOutDenom,
				test.SpreadFactor)
			s.Require().NoError(err)

			// check that the pool has not been modified after performing calc
			s.assertPoolNotModified(poolBeforeCalc)
			s.assertZeroSpreadRewards(poolBeforeCalc.GetId())
		})
	}
}

// TestComputeInAmtGivenOut tests that ComputeInAmtGivenOut successfully performs state changes as expected.
func (s *KeeperTestSuite) TestComputeInAmtGivenOut() {
	// we only use spread reward cases here since write Ctx only takes effect in the spread reward accumulator
	tests := makeTests(swapInGivenOutSpreadRewardTestCases)

	for name, test := range tests {
		test := test
		s.Run(name, func() {
			s.SetupAndFundSwapTest()
			poolBeforeCalc := s.preparePoolAndDefaultPositions(test)

			// perform calc
			_, _, err := s.App.ConcentratedLiquidityKeeper.ComputeInAmtGivenOut(
				s.Ctx,
				test.TokenOut, test.TokenInDenom,
				test.SpreadFactor, test.PriceLimit, poolBeforeCalc.GetId(), true)
			s.Require().NoError(err)

			// check that the pool has not been modified after performing calc
			s.assertPoolNotModified(poolBeforeCalc)
			s.assertSpreadRewardAccum(test, poolBeforeCalc.GetId())
		})
	}
}

func (s *KeeperTestSuite) TestInverseRelationshipSwapOutAmtGivenIn() {
	tests := swapOutGivenInCases

	for name, test := range tests {
		s.Run(name, func() {
			s.SetupAndFundSwapTest()
			poolBefore := s.preparePoolAndDefaultPositions(test)
			userBalanceBeforeSwap := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalanceBeforeSwap := s.App.BankKeeper.GetAllBalances(s.Ctx, poolBefore.GetAddress())

			// system under test
			firstTokenIn, firstTokenOut, _, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(
				s.Ctx, s.TestAccs[0], poolBefore,
				test.TokenIn, test.TokenOutDenom,
				DefaultZeroSpreadFactor, test.PriceLimit)
			s.Require().NoError(err)

			secondTokenIn, secondTokenOut, _, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(
				s.Ctx, s.TestAccs[0], poolBefore,
				firstTokenOut, firstTokenIn.Denom,
				DefaultZeroSpreadFactor, osmomath.ZeroBigDec(),
			)
			s.Require().NoError(err)

			// Run invariants on pool state, balances, and swap outputs.
			s.inverseRelationshipInvariants(firstTokenIn, firstTokenOut, secondTokenIn, secondTokenOut, poolBefore, userBalanceBeforeSwap, poolBalanceBeforeSwap, true)
		})
	}
}

func (s *KeeperTestSuite) TestUpdateSpreadRewardGrowthGlobal() {
	ten := osmomath.NewDec(10)

	tests := map[string]struct {
		liquidity                        osmomath.Dec
		spreadRewardChargeTotal          osmomath.Dec
		expectedSpreadRewardGrowthGlobal osmomath.Dec
	}{
		"zero liquidity -> no-op": {
			liquidity:                        osmomath.ZeroDec(),
			spreadRewardChargeTotal:          ten,
			expectedSpreadRewardGrowthGlobal: osmomath.ZeroDec(),
		},
		"non-zero liquidity -> updated": {
			liquidity:               ten,
			spreadRewardChargeTotal: ten,
			// 10 / 10 = 1
			expectedSpreadRewardGrowthGlobal: osmomath.OneDec(),
		},
		"rounding test: boundary spread reward growth": {
			liquidity:               ten.Add(ten).Mul(osmomath.NewDec(1e18)),
			spreadRewardChargeTotal: ten,
			// 10 / (20 * 10^18) = 5 * 10^-19, which we expect to truncate and leave 0.
			expectedSpreadRewardGrowthGlobal: osmomath.ZeroDec(),
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()

			// Setup.
			swapState := cl.SwapState{}
			swapState.SetLiquidity(tc.liquidity)
			swapState.SetGlobalSpreadRewardGrowthPerUnitLiquidity(osmomath.ZeroDec())
			swapState.SetGlobalSpreadRewardGrowth(osmomath.ZeroDec())

			// System under test.
			// TODO: Dont hardcode one here
			swapState.UpdateSpreadRewardGrowthGlobal(tc.spreadRewardChargeTotal, osmomath.OneDec())

			// Assertion.
			s.Require().Equal(tc.expectedSpreadRewardGrowthGlobal, swapState.GetGlobalSpreadRewardGrowthPerUnitLiquidity())
		})
	}
}

func (s *KeeperTestSuite) TestInverseRelationshipSwapInAmtGivenOut() {
	tests := swapInGivenOutTestCases

	for name, test := range tests {
		s.Run(name, func() {
			s.SetupAndFundSwapTest()
			poolBefore := s.preparePoolAndDefaultPositions(test)
			userBalanceBeforeSwap := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			poolBalanceBeforeSwap := s.App.BankKeeper.GetAllBalances(s.Ctx, poolBefore.GetAddress())

			// system under test
			firstTokenIn, firstTokenOut, _, err := s.App.ConcentratedLiquidityKeeper.SwapInAmtGivenOut(
				s.Ctx, s.TestAccs[0], poolBefore,
				test.TokenOut, test.TokenInDenom,
				DefaultZeroSpreadFactor, test.PriceLimit)
			s.Require().NoError(err)

			secondTokenIn, secondTokenOut, _, err := s.App.ConcentratedLiquidityKeeper.SwapInAmtGivenOut(
				s.Ctx, s.TestAccs[0], poolBefore,
				firstTokenIn, firstTokenOut.Denom,
				DefaultZeroSpreadFactor, osmomath.ZeroBigDec(),
			)
			s.Require().NoError(err)

			// Run invariants on pool state, balances, and swap outputs.
			s.inverseRelationshipInvariants(firstTokenIn, firstTokenOut, secondTokenIn, secondTokenOut, poolBefore, userBalanceBeforeSwap, poolBalanceBeforeSwap, false)
		})
	}
}

func (s *KeeperTestSuite) TestUpdatePoolForSwap() {
	var (
		oneHundredETH         = sdk.NewCoin(ETH, osmomath.NewInt(100_000_000))
		oneHundredUSDC        = sdk.NewCoin(USDC, osmomath.NewInt(100_000_000))
		defaultInitialBalance = sdk.NewCoins(oneHundredETH, oneHundredUSDC)
	)

	tests := map[string]struct {
		senderInitialBalance sdk.Coins
		poolInitialBalance   sdk.Coins
		tokenIn              sdk.Coin
		tokenOut             sdk.Coin
		spreadFactor         osmomath.Dec
		newCurrentTick       int64
		newLiquidity         osmomath.Dec
		newSqrtPrice         osmomath.BigDec
		expectError          error
	}{
		"success case": {
			senderInitialBalance: defaultInitialBalance,
			poolInitialBalance:   defaultInitialBalance,
			tokenIn:              oneHundredETH,
			tokenOut:             oneHundredUSDC,
			spreadFactor:         osmomath.MustNewDecFromStr("0.003"), // 0.3%
			newCurrentTick:       2,
			newLiquidity:         osmomath.NewDec(2),
			newSqrtPrice:         osmomath.NewBigDec(2),
		},
		"success case with different/uneven numbers": {
			senderInitialBalance: defaultInitialBalance.Add(defaultInitialBalance...),
			poolInitialBalance:   defaultInitialBalance,
			tokenIn:              oneHundredETH.Add(oneHundredETH),
			tokenOut:             oneHundredUSDC,
			spreadFactor:         osmomath.MustNewDecFromStr("0.002"), // 0.2%
			newCurrentTick:       8,
			newLiquidity:         osmomath.NewDec(37),
			newSqrtPrice:         osmomath.NewBigDec(91),
		},
		"sender does not have enough balance": {
			senderInitialBalance: defaultInitialBalance,
			poolInitialBalance:   defaultInitialBalance,
			tokenIn:              oneHundredETH.Add(oneHundredETH),
			tokenOut:             oneHundredUSDC,
			spreadFactor:         osmomath.MustNewDecFromStr("0.003"),
			newCurrentTick:       2,
			newLiquidity:         osmomath.NewDec(2),
			newSqrtPrice:         osmomath.NewBigDec(2),
			expectError:          types.InsufficientUserBalanceError{},
		},
		"pool does not have enough balance": {
			senderInitialBalance: defaultInitialBalance,
			poolInitialBalance:   defaultInitialBalance,
			tokenIn:              oneHundredETH,
			tokenOut:             oneHundredUSDC.Add(oneHundredUSDC),
			spreadFactor:         osmomath.MustNewDecFromStr("0.003"),
			newCurrentTick:       2,
			newLiquidity:         osmomath.NewDec(2),
			newSqrtPrice:         osmomath.NewBigDec(2),
			expectError:          types.InsufficientPoolBalanceError{},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			pool := s.PreparePoolWithCustSpread(tc.spreadFactor)

			s.FundAcc(pool.GetAddress(), tc.poolInitialBalance)
			// Create account with empty balance and fund with initial balance
			sender := apptesting.CreateRandomAccounts(1)[0]
			s.FundAcc(sender, tc.senderInitialBalance)

			// Default pool values are initialized to one.
			err := pool.ApplySwap(osmomath.OneDec(), 1, osmomath.OneBigDec())
			s.Require().NoError(err)

			// Write default pool to state.
			err = s.Clk.SetPool(s.Ctx, pool)
			s.Require().NoError(err)

			// Set mock listener to make sure that is is called when desired.
			s.setListenerMockOnConcentratedLiquidityKeeper()

			expectedSpreadFactors := tc.tokenIn.Amount.ToLegacyDec().Mul(pool.GetSpreadFactor(s.Ctx)).Ceil()
			expectedSpreadFactorsCoins := sdk.NewCoins(sdk.NewCoin(tc.tokenIn.Denom, expectedSpreadFactors.TruncateInt()))
			swapDetails := cl.SwapDetails{sender, tc.tokenIn, tc.tokenOut}
			poolUpdates := cl.PoolUpdates{tc.newCurrentTick, tc.newLiquidity, tc.newSqrtPrice}
			err = s.Clk.UpdatePoolForSwap(s.Ctx, pool, swapDetails, poolUpdates, expectedSpreadFactors)

			// Test that pool is updated
			poolAfterUpdate, err2 := s.Clk.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err2)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectError)

				// Test that pool is not updated
				s.Require().Equal(pool.String(), poolAfterUpdate.String())
				return
			}
			s.Require().NoError(err)

			s.Require().Equal(tc.newCurrentTick, poolAfterUpdate.GetCurrentTick())
			s.Require().Equal(tc.newLiquidity, poolAfterUpdate.GetLiquidity())
			s.Require().Equal(tc.newSqrtPrice, poolAfterUpdate.GetCurrentSqrtPrice())

			// Estimate expected final balances from inputs.
			expectedSenderFinalBalance := tc.senderInitialBalance.Sub(sdk.NewCoins(tc.tokenIn)...).Add(tc.tokenOut)
			expectedPoolFinalBalance := tc.poolInitialBalance.Add(tc.tokenIn).Sub(sdk.NewCoins(tc.tokenOut)...).Sub(expectedSpreadFactorsCoins...)

			// Test that token out is sent from pool to sender.
			senderBalanceAfterSwap := s.App.BankKeeper.GetAllBalances(s.Ctx, sender)
			s.Require().Equal(expectedSenderFinalBalance.String(), senderBalanceAfterSwap.String())

			// Test that token in is sent from sender to pool.
			poolBalanceAfterSwap := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
			s.Require().Equal(expectedPoolFinalBalance.String(), poolBalanceAfterSwap.String())

			spreadFactorBalanceAfterSwap := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetSpreadRewardsAddress())
			s.Require().Equal(expectedSpreadFactorsCoins.String(), spreadFactorBalanceAfterSwap.String())

			// Validate that listeners were called the desired number of times
			s.validateListenerCallCount(0, 0, 0, 1)
		})
	}
}

func (s *KeeperTestSuite) inverseRelationshipInvariants(firstTokenIn, firstTokenOut, secondTokenIn, secondTokenOut sdk.Coin, poolBefore poolmanagertypes.PoolI, userBalanceBeforeSwap sdk.Coins, poolBalanceBeforeSwap sdk.Coins, outGivenIn bool) {
	pool, ok := poolBefore.(types.ConcentratedPoolExtension)
	s.Require().True(ok)

	liquidityBefore, err := s.App.ConcentratedLiquidityKeeper.GetTotalPoolLiquidity(s.Ctx, pool.GetId())
	s.Require().NoError(err)

	// The output of the first swap should be exactly the same as the input of the second swap.
	// The input of the first swap should be within a margin of error of the output of the second swap.
	if outGivenIn {
		s.Require().Equal(firstTokenOut, secondTokenIn)
		s.validateAmountsWithTolerance(firstTokenIn.Amount, secondTokenOut.Amount)
	} else {
		s.Require().Equal(firstTokenIn, secondTokenOut)
		s.validateAmountsWithTolerance(firstTokenOut.Amount, secondTokenIn.Amount)
	}

	// Assure that pool state came back to original state
	poolAfter, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolBefore.GetId())
	s.Require().NoError(err)

	liquidityAfter, err := s.App.ConcentratedLiquidityKeeper.GetTotalPoolLiquidity(s.Ctx, pool.GetId())
	s.Require().NoError(err)

	// After both swaps, the pool should have the same total shares and total liquidity.
	s.Require().Equal(liquidityBefore, liquidityAfter)

	// Within a margin of error, the spot price should be the same before and after the swap
	oldSpotPrice, err := poolBefore.SpotPrice(s.Ctx, pool.GetToken1(), pool.GetToken0())
	s.Require().NoError(err)
	newSpotPrice, err := poolAfter.SpotPrice(s.Ctx, pool.GetToken1(), pool.GetToken0())
	s.Require().NoError(err)
	multiplicativeTolerance = osmomath.ErrTolerance{
		MultiplicativeTolerance: osmomath.MustNewDecFromStr("0.001"),
	}
	// Note: spot price truncation is made because the test was created before we changed in from Dec to BigDec
	// As a result, it is acceptable to truncate for test correctness.
	osmoassert.Equal(s.T(), multiplicativeTolerance, oldSpotPrice.Dec().RoundInt(), newSpotPrice.Dec().RoundInt())

	// Assure that user balance now as it was before both swaps.
	// TODO: Come back to this choice after deciding if we are using BigDec for swaps
	// https://github.com/osmosis-labs/osmosis/issues/4475
	userBalanceAfterSwap := s.App.BankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
	poolBalanceAfterSwap := s.App.BankKeeper.GetAllBalances(s.Ctx, poolBefore.GetAddress())
	for _, coin := range userBalanceBeforeSwap {
		beforeSwap := userBalanceBeforeSwap.AmountOf(coin.Denom)
		afterSwap := userBalanceAfterSwap.AmountOf(coin.Denom)
		osmoassert.Equal(s.T(), multiplicativeTolerance, beforeSwap, afterSwap)
	}
	for _, coin := range poolBalanceBeforeSwap {
		beforeSwap := poolBalanceBeforeSwap.AmountOf(coin.Denom)
		afterSwap := poolBalanceAfterSwap.AmountOf(coin.Denom)
		osmoassert.Equal(s.T(), multiplicativeTolerance, beforeSwap, afterSwap)
	}
}

// validateAmountsWithTolerance validates the given amounts a and b, allowing
// a negligible multiplicative error and an additive error of 1.
func (s *KeeperTestSuite) validateAmountsWithTolerance(amountA osmomath.Int, amountB osmomath.Int) {
	multCompare := multiplicativeTolerance.Compare(amountA, amountB)
	if multCompare != 0 {
		// If the multiplicative comparison fails, try again with additive tolerance of one.
		// This may occur for small amounts where the multiplicative tolerance ends up being
		// too restrictive for the rounding difference of just 1. E.g. 100 vs 101 does not satisfy the
		// 0.01% multiplciative margin of error but it is acceptable due to expected rounding epsilon.
		osmoassert.Equal(s.T(), oneAdditiveTolerance, amountA, amountB)
	} else {
		s.Require().Equal(0, multCompare, "amountA: %s, amountB: %s", amountA, amountB)
	}
}

func (s *KeeperTestSuite) TestFunctionalSwaps() {
	positions := Positions{
		numSwaps:       5,
		numAccounts:    5,
		numFullRange:   4,
		numNarrowRange: 3,
		numConsecutive: 2,
		numOverlapping: 1,
	}
	// Init s.
	s.SetupTest()

	// Determine amount of ETH and USDC to swap per swap.
	// These values were chosen as to not deplete the entire liquidity, but enough to move the price considerably.
	swapCoin0 := sdk.NewCoin(ETH, DefaultAmt0.Quo(osmomath.NewInt(int64(positions.numSwaps))))
	swapCoin1 := sdk.NewCoin(USDC, DefaultAmt1.Quo(osmomath.NewInt(int64(positions.numSwaps))))

	// Default setup only creates 3 accounts, but we need 5 for this test.
	s.TestAccs = apptesting.CreateRandomAccounts(positions.numAccounts)

	// Create a default CL pool, but with a 0.3 percent swap spread factor.
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, osmomath.MustNewDecFromStr("0.003"))

	positionIds := make([][]uint64, 4)
	// Setup full range position across all four accounts
	for i := 0; i < positions.numFullRange; i++ {
		positionId := s.SetupFullRangePositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[0] = append(positionIds[0], positionId)
	}

	// Setup narrow range position across three of four accounts
	for i := 0; i < positions.numNarrowRange; i++ {
		positionId := s.SetupDefaultPositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[1] = append(positionIds[1], positionId)
	}

	// Setup consecutive range position (in relation to narrow range position) across two of four accounts
	for i := 0; i < positions.numConsecutive; i++ {
		positionId := s.SetupConsecutiveRangePositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[2] = append(positionIds[2], positionId)
	}

	// Setup overlapping range position (in relation to narrow range position) on one of four accounts
	for i := 0; i < positions.numOverlapping; i++ {
		positionId := s.SetupOverlappingRangePositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[3] = append(positionIds[3], positionId)
	}

	// Depiction of the pool before any swaps
	//
	//  0 -----------------------------|-------------------------------------------- ∞
	//                   4545 ---------|-------- 5500
	//                                 |    5500 --------------- 6250
	//         4000 ----------------- 4999
	//                                 |
	//                              5000

	// Swap multiple times USDC for ETH, therefore increasing the spot price
	_, _, totalTokenIn, totalTokenOut := s.swapAndTrackXTimesInARow(clPool.GetId(), swapCoin1, ETH, types.MaxSpotPriceBigDec, positions.numSwaps)
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
	s.Require().NoError(err)

	// Depiction of the pool after the swaps (from 5000 to 5146), increasing the spot price
	//                                   >
	//  0 -----------------------------|--|----------------------------------------- ∞
	//                   4545 ---------|--|----- 5500
	//                                 |  | 5500 --------------- 6250
	//         4000 ----------------- 4999|
	//                                 |  |
	//                              5000 > 5146
	//
	// from math import *
	// from decimal import *
	// liq = Decimal("4836489743.729150266025048947")
	// sqrt_cur = Decimal("70.710678118654752441") # sqrt5000
	// token_in = Decimal("5000000000")
	// spread_factor = Decimal("0.003")
	// token_in_after_spread_factors = token_in * (Decimal("1") - spread_factor)
	// sqrt_next = sqrt_cur + token_in_after_spread_factors / liq
	// token_out = liq * (sqrt_next - sqrt_cur) / (sqrt_cur * sqrt_next)

	// # Summary:
	// print(sqrt_next) # 71.74138432587113364823838192
	// print(token_out) # 982676.1324268988579833395181

	// Get expected values from the calculations above
	expectedSqrtPrice := osmomath.MustNewBigDecFromStr("71.74138432587113364823838192")
	actualSqrtPrice := clPool.GetCurrentSqrtPrice()
	expectedTokenIn := swapCoin1.Amount.Mul(osmomath.NewInt(int64(positions.numSwaps)))
	expectedTokenOut := osmomath.NewInt(982676)

	// Compare the expected and actual values with a multiplicative tolerance of 0.0001%
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedSqrtPrice, actualSqrtPrice)
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedTokenIn, totalTokenIn.Amount)
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedTokenOut, totalTokenOut.Amount)

	// Withdraw all full range positions
	for _, positionId := range positionIds[0] {
		position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
		s.Require().NoError(err)
		owner, err := sdk.AccAddressFromBech32(position.Address)
		s.Require().NoError(err)
		_, _, err = s.App.ConcentratedLiquidityKeeper.WithdrawPosition(s.Ctx, owner, positionId, position.Liquidity)
		s.Require().NoError(err)
	}

	// Swap multiple times ETH for USDC, therefore decreasing the spot price
	_, _, totalTokenIn, totalTokenOut = s.swapAndTrackXTimesInARow(clPool.GetId(), swapCoin0, USDC, types.MinSpotPriceBigDec, positions.numSwaps)
	clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
	s.Require().NoError(err)

	// Depiction of the pool after the swaps (from 5146 to 4990), decreasing the spot price
	//								   <
	//                   4545 -----|------|----- 5500
	//                             |      | 5500 --------------- 6250
	//         4000 ---------------|- 4999|
	//                             |      |
	//                          4990   <  5146
	// from math import *
	// from decimal import *
	// # Range 1: From 5146 to 4999
	// token_in = Decimal("1000000")
	// spread_factor = Decimal("0.003")
	// token_in_after_spread_factors = token_in - (token_in * spread_factor)
	// liq_1 = Decimal("4553647031.254531254265048947")
	// sqrt_cur = Decimal("71.741384325871133645")
	// sqrt_next_1 = Decimal("4999").sqrt()

	// token_out_1 = liq_1 * (sqrt_cur - sqrt_next_1)
	// token_in_1 = ceil(liq_1 * (sqrt_cur - sqrt_next_1) / (sqrt_next_1 * sqrt_cur))

	// token_in = token_in_after_spread_factors - token_in_1

	// # Range 2: from 4999 till end
	// liq_2 = Decimal("5224063246.973358697925449540")
	// sqrt_next_2 = liq_2 / ((liq_2 / sqrt_next_1) + token_in)
	// token_out_2 = liq_2 * (sqrt_next_1 - sqrt_next_2)
	// token_in_2 = ceil(liq_2 * (sqrt_next_1 - sqrt_next_2) /
	// 				  (sqrt_next_2 * sqrt_next_1))
	// token_out = token_out_1 + token_out_2

	// # Summary:
	// print(sqrt_next_2)  # 70.64112736841825140176332377
	// print(token_out)    # 5052068983.121266708067570832

	// Get expected values from the calculations above
	expectedSqrtPrice = osmomath.MustNewBigDecFromStr("70.64112736841825140176332377")
	actualSqrtPrice = clPool.GetCurrentSqrtPrice()
	expectedTokenIn = swapCoin0.Amount.Mul(osmomath.NewInt(int64(positions.numSwaps)))
	expectedTokenOut = osmomath.NewInt(5052068983)

	// Compare the expected and actual values with a multiplicative tolerance of 0.0001%
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedSqrtPrice, actualSqrtPrice)
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedTokenIn, totalTokenIn.Amount)
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedTokenOut, totalTokenOut.Amount)

	// Withdraw all narrow range positions
	for _, positionId := range positionIds[1] {
		position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
		s.Require().NoError(err)
		owner, err := sdk.AccAddressFromBech32(position.Address)
		s.Require().NoError(err)
		_, _, err = s.App.ConcentratedLiquidityKeeper.WithdrawPosition(s.Ctx, owner, positionId, position.Liquidity)
		s.Require().NoError(err)
	}

	// Swap multiple times USDC for ETH, therefore increasing the spot price
	_, _, totalTokenIn, totalTokenOut = s.swapAndTrackXTimesInARow(clPool.GetId(), swapCoin1, ETH, types.MaxSpotPriceBigDec, positions.numSwaps)
	clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
	s.Require().NoError(err)

	// Depiction of the pool after the swaps (from 4990 to 5810), increasing the spot price
	//								      >
	//                             |        5500 -|------------- 6250
	//         4000 ---------------|- 4999        |
	//                             |              |
	//                          4990      >       5810
	// from math import *
	// from decimal import *
	// # Range 1: From 4990.16... to 4999
	// token_in = Decimal("5000000000")
	// spread_factor = Decimal("0.003")
	// token_in_after_spread_factors = token_in - (token_in * spread_factor)
	// liq_1 = Decimal("670416215.718827443660400593")
	// sqrt_cur = Decimal("70.641127368418251403")
	// sqrt_next_1 = Decimal("4999").sqrt()

	// token_out_1 = liq_1 * (sqrt_next_1 - sqrt_cur) / (sqrt_next_1 * sqrt_cur)
	// token_in_1 = ceil(liq_1 * abs(sqrt_cur - sqrt_next_1))

	// token_in = token_in_after_spread_factors - token_in_1

	// # Range 2: from 5500 till end
	// sqrt_next_1 = Decimal("74.161984870956629488") # sqrt5500
	// liq_2 = Decimal("2395534889.911016246446002272")
	// sqrt_next_2 = sqrt_next_1 + token_in / liq_2

	// token_out_2 = liq_2 * (sqrt_next_2 - sqrt_next_1) / (sqrt_next_1 * sqrt_next_2)
	// token_in_2 = ceil(liq_2 * abs(sqrt_next_2 - sqrt_next_1))
	// token_out = token_out_1 + token_out_2

	// # Summary:
	// print(sqrt_next_2)  # 76.22545423006231767390422658
	// print(token_out)    # 882804.6589413517320313885494

	// Get expected values from the calculations above
	expectedSqrtPrice = osmomath.MustNewBigDecFromStr("76.22545423006231767390422658")
	actualSqrtPrice = clPool.GetCurrentSqrtPrice()
	expectedTokenIn = swapCoin1.Amount.Mul(osmomath.NewInt(int64(positions.numSwaps)))
	expectedTokenOut = osmomath.NewInt(882804)

	// Compare the expected and actual values with a multiplicative tolerance of 0.0001%
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedSqrtPrice, actualSqrtPrice)
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedTokenIn, totalTokenIn.Amount)
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedTokenOut, totalTokenOut.Amount)

	// Withdraw all consecutive range position (in relation to narrow range position)
	for _, positionId := range positionIds[2] {
		position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
		s.Require().NoError(err)
		owner, err := sdk.AccAddressFromBech32(position.Address)
		s.Require().NoError(err)
		_, _, err = s.App.ConcentratedLiquidityKeeper.WithdrawPosition(s.Ctx, owner, positionId, position.Liquidity)
		s.Require().NoError(err)
	}

	// Swap multiple times ETH for USDC, therefore decreasing the spot price
	_, _, totalTokenIn, totalTokenOut = s.swapAndTrackXTimesInARow(clPool.GetId(), swapCoin0, USDC, types.MinSpotPriceBigDec, positions.numSwaps)
	clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
	s.Require().NoError(err)

	// Depiction of the pool after the swaps (from 5810 to 4093), decreasing the spot price
	//                               <
	//         4000 -|--------------- 4999          |
	//				 |							    |
	//            4093		         <	            5810
	//
	// from math import *
	// from decimal import *
	// liq = Decimal("670416215.718827443660400593")
	// sqrt_cur = Decimal("4999").sqrt()
	// token_in = Decimal("1000000")
	// spread_factor = Decimal("0.003")
	// token_in_after_spread_factors = token_in * (Decimal("1") - spread_factor)
	// sqrt_next = liq / ((liq / sqrt_cur) + token_in_after_spread_factors)
	// token_out = liq * (sqrt_cur - sqrt_next)

	// # Summary:
	// print(sqrt_next)  # 63.97671895942244949922335999
	// print(token_out)  # 4509814620.762503497903902725

	// Get expected values from the calculations above
	expectedSqrtPrice = osmomath.MustNewBigDecFromStr("63.97671895942244949922335999")
	actualSqrtPrice = clPool.GetCurrentSqrtPrice()
	expectedTokenIn = swapCoin0.Amount.Mul(osmomath.NewInt(int64(positions.numSwaps)))
	expectedTokenOut = osmomath.NewInt(4509814620)

	// Compare the expected and actual values with a multiplicative tolerance of 0.0001%
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedSqrtPrice, actualSqrtPrice)
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedTokenIn, totalTokenIn.Amount)
	osmoassert.Equal(s.T(), multiplicativeTolerance, expectedTokenOut, totalTokenOut.Amount)
}

// TestInfiniteSwapLoop_OutGivenIn demonstrates a case where an infinite loop can be triggered in swap logic if no
// swap limit and other constraints are applied.
func (s *KeeperTestSuite) TestInfiniteSwapLoop_OutGivenIn() {
	s.SetupTest()
	pool := s.PrepareConcentratedPool()

	testAccs := apptesting.CreateRandomAccounts(2)
	positionOwner := testAccs[0]

	// Create position near min tick
	s.FundAcc(positionOwner, DefaultRangeTestParams.baseAssets.Add(DefaultRangeTestParams.baseAssets...))
	_, err := s.Clk.CreatePosition(s.Ctx, pool.GetId(), positionOwner, DefaultRangeTestParams.baseAssets, osmomath.ZeroInt(), osmomath.ZeroInt(), -108000000, -107999900)
	s.Require().NoError(err)

	// Swap small amount to get current tick to position above, triggering the problematic function/branch (CalcAmount0Delta)
	swapAddress := testAccs[1]
	swapEthFunded := sdk.NewCoin(ETH, osmomath.Int(osmomath.MustNewDecFromStr("10000000000000000000000000000000000000000")))
	swapUSDCFunded := sdk.NewCoin(USDC, osmomath.Int(osmomath.MustNewDecFromStr("10000")))
	s.FundAcc(swapAddress, sdk.NewCoins(swapEthFunded, swapUSDCFunded))
	_, tokenOut, _, err := s.Clk.SwapInAmtGivenOut(s.Ctx, swapAddress, pool, sdk.NewCoin(USDC, osmomath.NewInt(10000)), ETH, pool.GetSpreadFactor(s.Ctx), osmomath.ZeroBigDec())
	s.Require().NoError(err)

	// Swap back in the amount that was swapped out to test the inverse relationship
	_, _, _, err = s.Clk.SwapOutAmtGivenIn(s.Ctx, swapAddress, pool, tokenOut, ETH, pool.GetSpreadFactor(s.Ctx), osmomath.ZeroBigDec())
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestComputeMaxInAmtGivenMaxTicksCrossed() {
	tests := []struct {
		name            string
		tokenInDenom    string
		tokenOutDenom   string
		maxTicksCrossed uint64
		expectedError   error
	}{
		{
			name:            "happy path, ETH in, max ticks equal to number of initialized ticks in swap direction",
			tokenInDenom:    ETH,
			tokenOutDenom:   USDC,
			maxTicksCrossed: 3,
		},
		{
			name:            "happy path, USDC in, max ticks equal to number of initialized ticks in swap direction",
			tokenInDenom:    USDC,
			tokenOutDenom:   ETH,
			maxTicksCrossed: 3,
		},
		{
			name:            "ETH in, max ticks less than number of initialized ticks in swap direction",
			tokenInDenom:    ETH,
			tokenOutDenom:   USDC,
			maxTicksCrossed: 2,
		},
		{
			name:            "USDC in, max ticks less than number of initialized ticks in swap direction",
			tokenInDenom:    USDC,
			tokenOutDenom:   ETH,
			maxTicksCrossed: 2,
		},
		{
			name:            "ETH in, max ticks greater than number of initialized ticks in swap direction",
			tokenInDenom:    ETH,
			tokenOutDenom:   USDC,
			maxTicksCrossed: 4,
		},
		{
			name:            "USDC in, max ticks greater than number of initialized ticks in swap direction",
			tokenInDenom:    USDC,
			tokenOutDenom:   ETH,
			maxTicksCrossed: 4,
		},
		{
			name:            "error: tokenInDenom not in pool",
			tokenInDenom:    "BTC",
			tokenOutDenom:   ETH,
			maxTicksCrossed: 4,
			expectedError:   types.TokenInDenomNotInPoolError{TokenInDenom: "BTC"},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			clPool := s.PrepareConcentratedPool()
			expectedResultingTokenOutAmount := osmomath.ZeroInt()

			// Create positions and calculate expected resulting tokens
			positions := []struct {
				lowerTick, upperTick int64
				maxTicks             uint64
			}{
				{DefaultLowerTick, DefaultUpperTick, 0},                 // Surrounding the current price
				{DefaultLowerTick - 10000, DefaultLowerTick, 1},         // Below the position surrounding the current price
				{DefaultLowerTick - 20000, DefaultLowerTick - 10000, 2}, // Below the position below the position surrounding the current price
				{DefaultUpperTick, DefaultUpperTick + 10000, 1},         // Above the position surrounding the current price
				{DefaultUpperTick + 10000, DefaultUpperTick + 20000, 2}, // Above the position above the position surrounding the current price
			}

			// Create positions and determine how much token out we should expect given the maxTicksCrossed provided.
			for _, pos := range positions {
				amt0, amt1 := s.createPositionAndFundAcc(clPool, pos.lowerTick, pos.upperTick)
				expectedResultingTokenOutAmount = s.calculateExpectedTokens(test.tokenInDenom, test.maxTicksCrossed, pos.maxTicks, amt0, amt1, expectedResultingTokenOutAmount)
			}

			// System Under Test
			_, resultingTokenOut, err := s.App.ConcentratedLiquidityKeeper.ComputeMaxInAmtGivenMaxTicksCrossed(s.Ctx, clPool.GetId(), test.tokenInDenom, test.maxTicksCrossed)

			if test.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedError.Error())
			} else {
				s.Require().NoError(err)

				errTolerance := osmomath.ErrTolerance{AdditiveTolerance: osmomath.NewDec(int64(test.maxTicksCrossed))}
				osmoassert.Equal(s.T(), errTolerance, expectedResultingTokenOutAmount, resultingTokenOut.Amount)
			}
		})
	}
}

func (s *KeeperTestSuite) createPositionAndFundAcc(clPool types.ConcentratedPoolExtension, lowerTick, upperTick int64) (amt0, amt1 osmomath.Int) {
	s.FundAcc(s.TestAccs[0], DefaultCoins)
	positionData, _ := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, clPool.GetId(), s.TestAccs[0], DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), lowerTick, upperTick)
	amt0 = positionData.Amount0
	amt1 = positionData.Amount1
	return
}

func (s *KeeperTestSuite) calculateExpectedTokens(tokenInDenom string, testMaxTicks, positionMaxTicks uint64, amt0, amt1, currentTotal osmomath.Int) osmomath.Int {
	if tokenInDenom == ETH && testMaxTicks > positionMaxTicks {
		return currentTotal.Add(amt1)
	} else if tokenInDenom == USDC && testMaxTicks > positionMaxTicks {
		return currentTotal.Add(amt0)
	}
	return currentTotal
}
