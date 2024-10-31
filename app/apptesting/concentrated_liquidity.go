package apptesting

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	clmath "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
)

type ConcentratedKeeperTestHelper struct {
	KeeperTestHelper
	Clk               *cl.Keeper
	AuthorizedUptimes []time.Duration
}

// Defines a concentrated liquidity swap test case.
type ConcentratedSwapTest struct {
	// Specific to swap out given in.
	TokenIn       sdk.Coin
	TokenOutDenom string

	// Specific to swap in given out.
	TokenOut     sdk.Coin
	TokenInDenom string

	// Shared.
	PriceLimit               osmomath.BigDec
	SpreadFactor             osmomath.Dec
	SecondPositionLowerPrice osmomath.Dec
	SecondPositionUpperPrice osmomath.Dec

	ExpectedTokenIn                            sdk.Coin
	ExpectedTokenOut                           sdk.Coin
	ExpectedTick                               int64
	ExpectedSqrtPrice                          osmomath.BigDec
	ExpectedLowerTickSpreadRewardGrowth        sdk.DecCoins
	ExpectedUpperTickSpreadRewardGrowth        sdk.DecCoins
	ExpectedSpreadRewardGrowthAccumulatorValue osmomath.Dec
	// since we use different values for the seondary position's tick, save (tick, expectedSpreadRewardGrowth) tuple
	ExpectedSecondLowerTickSpreadRewardGrowth SecondConcentratedPosition
	ExpectedSecondUpperTickSpreadRewardGrowth SecondConcentratedPosition

	NewLowerPrice  osmomath.Dec
	NewUpperPrice  osmomath.Dec
	PoolLiqAmount0 osmomath.Int
	PoolLiqAmount1 osmomath.Int
	ExpectErr      bool
}

type SecondConcentratedPosition struct {
	TickIndex                  int64
	ExpectedSpreadRewardGrowth sdk.DecCoins
}

var (
	WBTC                       = "ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F"
	DefaultTickSpacing         = uint64(100)
	DefaultLowerPrice          = osmomath.NewDec(4545)
	DefaultLowerTick           = int64(30545000)
	DefaultUpperPrice          = osmomath.NewDec(5500)
	DefaultUpperTick           = int64(31500000)
	DefaultCurrPrice           = osmomath.NewDec(5000)
	DefaultCurrTick      int64 = 31000000
	DefaultCurrSqrtPrice       = func() osmomath.BigDec {
		curSqrtPrice, _ := osmomath.MonotonicSqrt(DefaultCurrPrice) // 70.710678118654752440
		return osmomath.BigDecFromDec(curSqrtPrice)
	}()
	PerUnitLiqScalingFactor = osmomath.NewDec(1e15).MulMut(osmomath.NewDec(1e12))

	DefaultSpreadRewardAccumCoins = sdk.NewDecCoins(sdk.NewDecCoinFromDec("foo", osmomath.NewDec(50).MulTruncate(PerUnitLiqScalingFactor)))

	DefaultCoinAmount = osmomath.NewInt(1000000000000000000)

	// Default tokens and amounts
	ETH                 = "eth"
	DefaultAmt0         = osmomath.NewInt(1000000)
	DefaultAmt0Expected = osmomath.NewInt(998976)
	DefaultCoin0        = sdk.NewCoin(ETH, DefaultAmt0)
	USDC                = "usdc"
	DefaultAmt1         = osmomath.NewInt(5000000000)
	DefaultAmt1Expected = osmomath.NewInt(5000000000)
	DefaultCoin1        = sdk.NewCoin(USDC, DefaultAmt1)
	DefaultCoins        = sdk.NewCoins(DefaultCoin0, DefaultCoin1)

	// Both of the following liquidity values are calculated in x/concentrated-liquidity/python/swap_test.py
	DefaultLiquidityAmt   = osmomath.MustNewDecFromStr("1517882343.751510417627556287")
	FullRangeLiquidityAmt = osmomath.MustNewDecFromStr("70710678.118654752941000000")

	swapFundCoins = sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(10000000000000)), sdk.NewCoin("usdc", osmomath.NewInt(1000000000000)))

	roundingError = osmomath.OneInt()
	EmptyCoins    = sdk.DecCoins(nil)

	// Various sqrt estimates
	Sqrt4994 = osmomath.MustNewDecFromStr("70.668238976219012614")

	usdcChainDenom = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"

	// swap out given in test cases
	SwapOutGivenInCases = map[string]ConcentratedSwapTest{
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
			ExpectedSecondLowerTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 315000, ExpectedSpreadRewardGrowth: EmptyCoins},
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
			ExpectedSecondLowerTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 300000, ExpectedSpreadRewardGrowth: EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 305450, ExpectedSpreadRewardGrowth: EmptyCoins},
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
			ExpectedSecondLowerTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 310010, ExpectedSpreadRewardGrowth: EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: EmptyCoins},
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
			ExpectedSecondLowerTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 310010, ExpectedSpreadRewardGrowth: EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: EmptyCoins},
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
			ExpectedSecondLowerTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 300000, ExpectedSpreadRewardGrowth: EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 309990, ExpectedSpreadRewardGrowth: EmptyCoins},
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
			ExpectedSecondLowerTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 300000, ExpectedSpreadRewardGrowth: EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 309990, ExpectedSpreadRewardGrowth: EmptyCoins},
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
			ExpectedSecondLowerTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 315010, ExpectedSpreadRewardGrowth: EmptyCoins},
			ExpectedSecondUpperTickSpreadRewardGrowth: SecondConcentratedPosition{TickIndex: 322500, ExpectedSpreadRewardGrowth: EmptyCoins},
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
				tick, _ := clmath.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(Sqrt4994), DefaultTickSpacing)
				return tick
			}(),
			// Since the next sqrt price is based on the price limit, we can calculate this directly.
			ExpectedSqrtPrice: osmomath.BigDecFromDec(osmomath.MustMonotonicSqrt(osmomath.NewDec(4994))),
		},
	}
)

// PrepareConcentratedPool sets up an eth usdc concentrated liquidity pool with a tick spacing of 100,
// no liquidity and zero spread factor.
func (s *KeeperTestHelper) PrepareConcentratedPool() types.ConcentratedPoolExtension {
	return s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, osmomath.ZeroDec())
}

// PrepareConcentratedPoolWithCoins sets up a concentrated liquidity pool with custom denoms.
func (s *KeeperTestHelper) PrepareConcentratedPoolWithCoins(denom1, denom2 string) types.ConcentratedPoolExtension {
	return s.PrepareCustomConcentratedPool(s.TestAccs[0], denom1, denom2, DefaultTickSpacing, osmomath.ZeroDec())
}

// PrepareConcentratedPoolWithCoinsAndFullRangePosition sets up a concentrated liquidity pool with custom denoms.
// It also creates a full range position.
func (s *KeeperTestHelper) PrepareConcentratedPoolWithCoinsAndFullRangePosition(denom1, denom2 string) types.ConcentratedPoolExtension {
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], denom1, denom2, DefaultTickSpacing, osmomath.ZeroDec())
	fundCoins := sdk.NewCoins(sdk.NewCoin(denom1, DefaultCoinAmount), sdk.NewCoin(denom2, DefaultCoinAmount))
	s.FundAcc(s.TestAccs[0], fundCoins)
	s.CreateFullRangePosition(clPool, fundCoins)
	return clPool
}

// createConcentratedPoolsFromCoinsWithSpreadFactor creates CL pools from given sets of coins and respective swap fees.
// Where element 1 of the input corresponds to the first pool created, element 2 to the second pool created etc.
func (s *KeeperTestHelper) CreateConcentratedPoolsAndFullRangePositionWithSpreadFactor(poolDenoms [][]string, spreadFactor []osmomath.Dec) {
	for i, curPoolDenoms := range poolDenoms {
		s.Require().Equal(2, len(curPoolDenoms))
		var curSpreadFactor osmomath.Dec
		if len(spreadFactor) > i {
			curSpreadFactor = spreadFactor[i]
		} else {
			curSpreadFactor = osmomath.ZeroDec()
		}

		clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], curPoolDenoms[0], curPoolDenoms[1], DefaultTickSpacing, curSpreadFactor)
		fundCoins := sdk.NewCoins(sdk.NewCoin(curPoolDenoms[0], DefaultCoinAmount), sdk.NewCoin(curPoolDenoms[1], DefaultCoinAmount))
		s.FundAcc(s.TestAccs[0], fundCoins)
		s.CreateFullRangePosition(clPool, fundCoins)
	}
}

// createConcentratedPoolsFromCoins creates CL pools from given sets of coins (with zero swap fees).
// Where element 1 of the input corresponds to the first pool created, element 2 to the second pool created etc.
func (s *KeeperTestHelper) CreateConcentratedPoolsAndFullRangePosition(poolDenoms [][]string) {
	s.CreateConcentratedPoolsAndFullRangePositionWithSpreadFactor(poolDenoms, []osmomath.Dec{osmomath.ZeroDec()})
}

// PrepareConcentratedPoolWithCoinsAndLockedFullRangePosition sets up a concentrated liquidity pool with custom denoms.
// It also creates a full range position and locks it for 14 days.
func (s *KeeperTestHelper) PrepareConcentratedPoolWithCoinsAndLockedFullRangePosition(denom1, denom2 string) (types.ConcentratedPoolExtension, uint64, uint64) {
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], denom1, denom2, DefaultTickSpacing, osmomath.ZeroDec())
	fundCoins := sdk.NewCoins(sdk.NewCoin(denom1, DefaultCoinAmount), sdk.NewCoin(denom2, DefaultCoinAmount))
	s.FundAcc(s.TestAccs[0], fundCoins)
	positionData, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), s.TestAccs[0], fundCoins, time.Hour*24*14)
	s.Require().NoError(err)
	clPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, clPool.GetId())
	s.Require().NoError(err)
	return clPool, concentratedLockId, positionData.ID
}

// PrepareCustomConcentratedPool sets up a concentrated liquidity pool with the custom parameters.
func (s *KeeperTestHelper) PrepareCustomConcentratedPool(owner sdk.AccAddress, denom0, denom1 string, tickSpacing uint64, spreadFactor osmomath.Dec) types.ConcentratedPoolExtension {
	// Mint some assets to the account.
	s.FundAcc(s.TestAccs[0], s.App.PoolManagerKeeper.GetParams(s.Ctx).PoolCreationFee)

	// Create a concentrated pool via the poolmanager
	poolID, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, clmodel.NewMsgCreateConcentratedPool(owner, denom0, denom1, tickSpacing, spreadFactor))
	s.Require().NoError(err)

	// Retrieve the poolInterface via the poolID
	poolI, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolID)
	s.Require().NoError(err)

	// Type cast the PoolInterface to a ConcentratedPoolExtension
	pool, ok := poolI.(types.ConcentratedPoolExtension)
	s.Require().True(ok)

	return pool
}

// PrepareMultipleConcentratedPools returns X cl pool's with X being provided by the user.
func (s *KeeperTestHelper) PrepareMultipleConcentratedPools(poolsToCreate uint16) []uint64 {
	var poolIds []uint64
	for i := uint16(0); i < poolsToCreate; i++ {
		pool := s.PrepareConcentratedPool()
		poolIds = append(poolIds, pool.GetId())
	}

	return poolIds
}

// CreateFullRangePosition creates a full range position and returns position id and the liquidity created.
func (s *KeeperTestHelper) CreateFullRangePosition(pool types.ConcentratedPoolExtension, coins sdk.Coins) (uint64, osmomath.Dec) {
	s.FundAcc(s.TestAccs[0], coins)
	positionData, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, pool.GetId(), s.TestAccs[0], coins)
	s.Require().NoError(err)
	return positionData.ID, positionData.Liquidity
}

// WithdrawFullRangePosition withdraws given liquidity from a position specified by id.
func (s *KeeperTestHelper) WithdrawFullRangePosition(pool types.ConcentratedPoolExtension, positionId uint64, liquidityToRemove osmomath.Dec) {
	clMsgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)

	_, err := clMsgServer.WithdrawPosition(s.Ctx, &types.MsgWithdrawPosition{
		PositionId:      positionId,
		LiquidityAmount: liquidityToRemove,
		Sender:          s.TestAccs[0].String(),
	})
	s.Require().NoError(err)
}

// SetupConcentratedLiquidityDenomsAndPoolCreation sets up the default authorized quote denoms.
// Additionally, enables permissionless pool creation.
// This is to overwrite the default params set in concentrated liquidity genesis to account for the test cases that
// used various denoms before the authorized quote denoms were introduced.
func (s *KeeperTestHelper) SetupConcentratedLiquidityDenomsAndPoolCreation() {
	// modify authorized quote denoms to include test denoms.
	defaultParams := types.DefaultParams()
	defaultParams.IsPermissionlessPoolCreationEnabled = true
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, defaultParams)

	authorizedQuoteDenoms := append(poolmanagertypes.DefaultParams().AuthorizedQuoteDenoms, ETH, USDC, BAR, BAZ, FOO, appparams.BaseCoinUnit, STAKE, WBTC, usdcChainDenom)
	s.App.PoolManagerKeeper.SetParam(s.Ctx, poolmanagertypes.KeyAuthorizedQuoteDenoms, authorizedQuoteDenoms)
}

func (s *ConcentratedKeeperTestHelper) SetupTest() {
	s.Reset()
	s.setupClGeneral()
}

func (s *ConcentratedKeeperTestHelper) SetupAndFundSwapTest() {
	s.SetupTest()
	s.FundAcc(s.TestAccs[0], swapFundCoins)
	s.FundAcc(s.TestAccs[1], swapFundCoins)
}

func (s *ConcentratedKeeperTestHelper) PreparePoolWithCustSpread(spread osmomath.Dec) types.ConcentratedPoolExtension {
	clParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	clParams.AuthorizedSpreadFactors = append(clParams.AuthorizedSpreadFactors, spread)
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, clParams)
	return s.PrepareCustomConcentratedPool(s.TestAccs[0], "eth", "usdc", DefaultTickSpacing, spread)
}

func (s *ConcentratedKeeperTestHelper) SetupDefaultPosition(poolId uint64) {
	s.SetupPosition(poolId, s.TestAccs[0], DefaultCoins, DefaultLowerTick, DefaultUpperTick, false)
}

func (s *ConcentratedKeeperTestHelper) SetupPosition(poolId uint64, owner sdk.AccAddress, providedCoins sdk.Coins, lowerTick, upperTick int64, addRoundingError bool) (osmomath.Dec, uint64) {
	roundingErrorCoins := sdk.NewCoins()
	if addRoundingError {
		roundingErrorCoins = sdk.NewCoins(sdk.NewCoin(ETH, roundingError), sdk.NewCoin(USDC, roundingError))
	}

	s.FundAcc(owner, providedCoins.Add(roundingErrorCoins...))
	positionData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, owner, providedCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), lowerTick, upperTick)
	s.Require().NoError(err)
	liquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, positionData.ID)
	s.Require().NoError(err)
	return liquidity, positionData.ID
}

func (s *ConcentratedKeeperTestHelper) SetupSecondPosition(test ConcentratedSwapTest, pool types.ConcentratedPoolExtension) {
	if !test.SecondPositionLowerPrice.IsNil() {
		newLowerTick, newUpperTick := s.LowerUpperPricesToTick(test.SecondPositionLowerPrice, test.SecondPositionUpperPrice, pool.GetTickSpacing())

		_, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), s.TestAccs[1], DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), newLowerTick, newUpperTick)
		s.Require().NoError(err)
	}
}

func (s *ConcentratedKeeperTestHelper) LowerUpperPricesToTick(lowerPrice, upperPrice osmomath.Dec, tickSpacing uint64) (int64, int64) {
	lowerSqrtPrice := osmomath.MustMonotonicSqrt(lowerPrice)
	newLowerTick, err := clmath.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(lowerSqrtPrice), tickSpacing)
	s.Require().NoError(err)
	upperSqrtPrice := osmomath.MustMonotonicSqrt(upperPrice)
	newUpperTick, err := clmath.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(upperSqrtPrice), tickSpacing)
	s.Require().NoError(err)
	return newLowerTick, newUpperTick
}

func (s *ConcentratedKeeperTestHelper) setupClGeneral() {
	s.Clk = s.App.ConcentratedLiquidityKeeper

	if s.AuthorizedUptimes != nil {
		clParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
		clParams.AuthorizedUptimes = s.AuthorizedUptimes
		s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, clParams)
	}
}
