package swapstrategy_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
)

type StrategyTestSuite struct {
	apptesting.KeeperTestHelper
}

type position struct {
	lowerTick int64
	upperTick int64
}

const (
	defaultPoolId      = uint64(1)
	initialCurrentTick = int64(0)
	ETH                = "eth"
	USDC               = "usdc"
)

var (
	zero                    = sdk.NewDec(0)
	one                     = sdk.NewDec(1)
	two                     = sdk.NewDec(2)
	three                   = sdk.NewDec(3)
	four                    = sdk.NewDec(4)
	five                    = sdk.NewDec(5)
	sqrt5000                = sdk.MustNewDecFromStr("70.710678118654752440") // 5000
	defaultSqrtPriceLower   = sdk.MustNewDecFromStr("70.688664163408836321") // approx 4996.89
	defaultSqrtPriceUpper   = sqrt5000
	defaultAmountOne        = sdk.MustNewDecFromStr("66829187.967824033199646915")
	defaultAmountZero       = sdk.MustNewDecFromStr("13369.999999999998920002")
	defaultAmountZeroBigDec = osmomath.MustNewDecFromStr("13369.999999999998920003259839786649584880")
	defaultLiquidity        = sdk.MustNewDecFromStr("3035764687.503020836176699298")
	defaultSpreadReward     = sdk.MustNewDecFromStr("0.03")
	defaultTickSpacing      = uint64(100)
	defaultAmountReserves   = sdk.NewInt(1_000_000_000)
	DefaultCoins            = sdk.NewCoins(sdk.NewCoin(ETH, defaultAmountReserves), sdk.NewCoin(USDC, defaultAmountReserves))
	oneULPDec               = sdk.SmallestDec()
	oneULPBigDec            = osmomath.SmallestDec()
)

func TestStrategyTestSuite(t *testing.T) {
	suite.Run(t, new(StrategyTestSuite))
}

func (suite *StrategyTestSuite) SetupTest() {
	suite.Setup()
}

type tickIteratorTest struct {
	currentTick     int64
	preSetPositions []position
	tickSpacing     uint64

	expectIsValid  bool
	expectNextTick int64
	expectError    error
}

func (suite *StrategyTestSuite) runTickIteratorTest(strategy swapstrategy.SwapStrategy, tc tickIteratorTest) {
	pool := suite.PrepareCustomConcentratedPool(suite.TestAccs[0], ETH, USDC, tc.tickSpacing, sdk.ZeroDec())
	suite.setupPresetPositions(pool.GetId(), tc.preSetPositions)

	// refetch pool
	pool, err := suite.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(suite.Ctx, pool.GetId())
	suite.Require().NoError(err)

	currentTick := pool.GetCurrentTick()
	suite.Require().Equal(int64(0), currentTick)

	iter := strategy.InitializeNextTickIterator(suite.Ctx, defaultPoolId, currentTick)
	defer iter.Close()

	suite.Require().Equal(tc.expectIsValid, iter.Valid())
	if tc.expectIsValid {
		actualNextTick, err := types.TickIndexFromBytes(iter.Key())
		suite.Require().NoError(err)
		suite.Require().Equal(tc.expectNextTick, actualNextTick)
	}
}

func (suite *StrategyTestSuite) setupPresetPositions(poolId uint64, positions []position) {
	clMsgServer := cl.NewMsgServerImpl(suite.App.ConcentratedLiquidityKeeper)
	for _, pos := range positions {
		suite.FundAcc(suite.TestAccs[0], DefaultCoins.Add(DefaultCoins...))
		_, err := clMsgServer.CreatePosition(sdk.WrapSDKContext(suite.Ctx), &types.MsgCreatePosition{
			PoolId:          poolId,
			Sender:          suite.TestAccs[0].String(),
			LowerTick:       pos.lowerTick,
			UpperTick:       pos.upperTick,
			TokensProvided:  DefaultCoins.Add(sdk.NewCoin(USDC, sdk.OneInt())),
			TokenMinAmount0: sdk.ZeroInt(),
			TokenMinAmount1: sdk.ZeroInt(),
		})
		suite.Require().NoError(err)
	}
}

// TestComputeSwapState_Inverse validates that out given in and in given out compute swap steps
// for one strategy produce the same results as the other given inverse inputs.
// That is, if we swap in A of token0 and expect to get out B of token 1,
// we should be able to get A of token 0 in when swapping out for B of token 1.
// Note that the expected values in this test are
// computed with x/concentrated-liquidity/python/clmath.py
func (suite *StrategyTestSuite) TestComputeSwapState_Inverse() {
	var (
		errToleranceOne = osmomath.ErrTolerance{
			AdditiveTolerance: sdk.OneDec(),
			RoundingDir:       osmomath.RoundUp,
		}

		errToleranceSmall = osmomath.ErrTolerance{
			AdditiveTolerance: sdk.NewDecFromIntWithPrec(sdk.OneInt(), 5),
		}
	)

	testCases := map[string]struct {
		sqrtPriceCurrent sdk.Dec
		sqrtPriceTarget  sdk.Dec
		liquidity        sdk.Dec
		amountIn         sdk.Dec
		amountOut        sdk.Dec
		zeroForOne       bool
		spreadFactor     sdk.Dec

		expectedSqrtPriceNextOutGivenIn osmomath.BigDec
		expectedSqrtPriceNextInGivenOut osmomath.BigDec
		expectedAmountIn                sdk.Dec
		expectedAmountOut               sdk.Dec
	}{
		"1: one_for_zero__not_equal_target__no_spread_reward": {
			sqrtPriceCurrent: sqrt5000,                                       // 5000
			sqrtPriceTarget:  sdk.MustNewDecFromStr("70.724818840347693039"), // 5002
			liquidity:        sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountIn:         sdk.NewDec(42000000),
			amountOut:        sdk.NewDec(8398),
			zeroForOne:       false,
			spreadFactor:     sdk.ZeroDec(),

			// get_next_sqrt_price_from_amount1_in_round_down(liquidity, sqrtPriceCurrent, tokenIn)
			expectedSqrtPriceNextOutGivenIn: osmomath.MustNewDecFromStr("70.724513183069625078753200000000838853"), // approx 5001.96

			// tokenOut = round_sdk_prec_down(calc_amount_zero_delta(liquidity, Decimal('70.724513183069625078753200000000838853'), sqrtPriceCurrent, False))
			// get_next_sqrt_price_from_amount0_out_round_up(liquidity, sqrtPriceCurrent, tokenOut)
			expectedSqrtPriceNextInGivenOut: osmomath.MustNewDecFromStr("70.724513183069625078753199315615320286"), // approx 5001.96

			expectedAmountIn:  sdk.NewDec(42000000),
			expectedAmountOut: sdk.NewDec(8398),
		},
		"2: zero_for_one__not_equal_target_no_spread_reward": {
			sqrtPriceCurrent: sqrt5000,                                       // 5000
			sqrtPriceTarget:  sdk.MustNewDecFromStr("70.682388188289167342"), // 4996
			liquidity:        sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountIn:         sdk.NewDec(13370),
			amountOut:        sdk.NewDec(66829187),
			zeroForOne:       true,
			spreadFactor:     sdk.ZeroDec(),

			// get_next_sqrt_price_from_amount0_in_round_up(liquidity, sqrtPriceCurrent, tokenIn)
			expectedSqrtPriceNextOutGivenIn: osmomath.MustNewDecFromStr("70.688664163408836319222318760848762802"), // approx 4996.89

			// tokenOut = round_sdk_prec_down(calc_amount_one_delta(liquidity, Decimal('70.688664163408836319222318760848762802'), sqrtPriceCurrent, False))
			// get_next_sqrt_price_from_amount1_out_round_down(liquidity, sqrtPriceCurrent, tokenOut)
			expectedSqrtPriceNextInGivenOut: osmomath.MustNewDecFromStr("70.688664163408836319222318761064639455"), // approx 4996.89

			expectedAmountIn:  sdk.NewDec(13370),
			expectedAmountOut: sdk.NewDec(66829187),
		},
		"3: one_for_zero__equal_target__no_spread_reward": {
			sqrtPriceCurrent: sqrt5000,                                       // 5000
			sqrtPriceTarget:  sdk.MustNewDecFromStr("70.724513183069625078"), // approx 5001.96
			liquidity:        sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountIn:         sdk.NewDec(42000000),
			amountOut:        sdk.NewDec(8398),
			spreadFactor:     sdk.ZeroDec(),

			zeroForOne: false,
			// same as target
			expectedSqrtPriceNextOutGivenIn: osmomath.MustNewDecFromStr("70.724513183069625078"), // approx 5001.96

			// tokenOut = round_sdk_prec_down(calc_amount_zero_delta(liquidity, Decimal('70.724513183069625078'), sqrtPriceCurrent, False))
			// get_next_sqrt_price_from_amount0_out_round_up(liquidity, sqrtPriceCurrent, tokenOut)
			expectedSqrtPriceNextInGivenOut: osmomath.MustNewDecFromStr("70.724513183069625077999998811165066229"), // approx 5001.96

			expectedAmountIn:  sdk.NewDec(42000000),
			expectedAmountOut: sdk.NewDec(8398),
		},
		"4: zero_for_one__equal_target__no_spread_reward": {
			sqrtPriceCurrent: sqrt5000,                                       // 5000
			sqrtPriceTarget:  sdk.MustNewDecFromStr("70.688664163408836320"), // approx 4996.89
			liquidity:        sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountIn:         sdk.NewDec(13370),
			amountOut:        sdk.NewDec(66829187),
			zeroForOne:       true,
			spreadFactor:     sdk.ZeroDec(),

			// same as target
			expectedSqrtPriceNextOutGivenIn: osmomath.MustNewDecFromStr("70.688664163408836320"), // approx 4996.89

			// tokenOut = round_sdk_prec_down(calc_amount_one_delta(liquidity, Decimal('70.688664163408836320'), sqrtPriceCurrent, False))
			// get_next_sqrt_price_from_amount1_out_round_down(liquidity, sqrtPriceCurrent, tokenOut)
			expectedSqrtPriceNextInGivenOut: osmomath.MustNewDecFromStr("70.688664163408836320000000000232703515"), // approx 4996.89

			expectedAmountIn:  sdk.NewDec(13370),
			expectedAmountOut: sdk.NewDec(66829187),
		},
	}

	for name, tc := range testCases {
		tc := tc
		suite.Run(name, func() {
			sut := swapstrategy.New(tc.zeroForOne, sdk.ZeroDec(), suite.App.GetKey(types.ModuleName), sdk.ZeroDec())
			sqrtPriceNextOutGivenIn, amountInOutGivenIn, amountOutOutGivenIn, _ := sut.ComputeSwapWithinBucketOutGivenIn(osmomath.BigDecFromSDKDec(tc.sqrtPriceCurrent), tc.sqrtPriceTarget, tc.liquidity, tc.amountIn)
			suite.Require().Equal(tc.expectedSqrtPriceNextOutGivenIn.String(), sqrtPriceNextOutGivenIn.String())

			fmt.Println("amountOutOutGivenIn", amountOutOutGivenIn)

			sqrtPriceNextInGivenOut, amountOutInGivenOut, amountInInGivenOut, _ := sut.ComputeSwapWithinBucketInGivenOut(osmomath.BigDecFromSDKDec(tc.sqrtPriceCurrent), tc.sqrtPriceTarget, tc.liquidity, amountOutOutGivenIn)

			suite.Require().Equal(tc.expectedSqrtPriceNextInGivenOut.String(), sqrtPriceNextInGivenOut.String())

			// Tolerance of 1 with rounding up because we round up for in given out.
			// This is to ensure that inflow into the pool is rounded in favor of the pool.
			suite.Require().Equal(0, errToleranceOne.CompareBigDec(
				osmomath.BigDecFromSDKDec(amountInOutGivenIn),
				osmomath.BigDecFromSDKDec(amountInInGivenOut)),
				fmt.Sprintf("amount in out given in: %s, amount in in given out: %s", amountInOutGivenIn, amountInInGivenOut))

			// These should be approximately equal. The difference stems from minor roundings and truncations in the intermediary calculations.
			suite.Require().Equal(0, errToleranceSmall.CompareBigDec(
				osmomath.BigDecFromSDKDec(amountOutOutGivenIn),
				osmomath.BigDecFromSDKDec(amountOutInGivenOut)),
				fmt.Sprintf("amount out out given in: %s, amount out in given out: %s", amountOutOutGivenIn, amountOutInGivenOut))
		})
	}
}

func (suite *StrategyTestSuite) TestGetPriceLimit() {
	tests := map[string]struct {
		zeroForOne bool
		expected   sdk.Dec
	}{
		"zero for one -> min": {
			zeroForOne: true,
			expected:   types.MinSpotPrice,
		},
		"one for zero -> max": {
			zeroForOne: false,
			expected:   types.MaxSpotPrice,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			priceLimit := swapstrategy.GetPriceLimit(tc.zeroForOne)
			suite.Require().Equal(tc.expected, priceLimit)
		})
	}
}
