package model_test

import (
	fmt "fmt"
	"math"
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	clmath "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

const (
	ETH                = "eth"
	USDC               = "usdc"
	DAI                = "dai"
	DefaultValidPoolID = uint64(1)
	DefaultTickSpacing = uint64(1)
)

var (
	DefaultSpotPrice        = osmomath.MustNewDecFromStr("0.2")
	DefaultReverseSpotPrice = osmomath.NewDec(1).Quo(DefaultSpotPrice)
	DefaultSqrtSpotPrice    = func() osmomath.BigDec {
		sqrtPrice, _ := osmomath.MonotonicSqrt(DefaultSpotPrice)
		return osmomath.BigDecFromDecMut(sqrtPrice)
	}()
	DefaultLiquidityAmt        = osmomath.MustNewDecFromStr("1517882343.751510418088349649")
	DefaultCurrTick      int64 = 310000
	DefaultCurrPrice           = osmomath.NewDec(5000)
	DefaultCurrSqrtPrice       = func() osmomath.BigDec {
		sqrtPrice, _ := osmomath.MonotonicSqrt(DefaultCurrPrice)
		return osmomath.BigDecFromDecMut(sqrtPrice)
	}() // 70.710678118654752440

	DefaultSpreadFactor = osmomath.MustNewDecFromStr("0.01")
)

type ConcentratedPoolTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestConcentratedPoolTestSuite(t *testing.T) {
	suite.Run(t, new(ConcentratedPoolTestSuite))
}

// TestGetAddress tests the GetAddress method of pool
func (s *ConcentratedPoolTestSuite) TestGetAddress() {

	tests := []struct {
		name          string
		expectedPanic bool
	}{
		{
			name: "Happy path",
		},
		{
			name:          "Unhappy path: wrong bech32 encoded address",
			expectedPanic: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Init suite for each test.
			s.Setup()

			address := s.TestAccs[0].String()

			// if the test case is expected to panic, we use wrong bech32 encoded address instead
			if tc.expectedPanic {
				address = "osmo15l7yueqf3tx4cvpt6njvj7zxmvuhkwyrr509e9"
			}
			mock_pool := model.Pool{
				Id:      1,
				Address: address,
			}

			// Check that the returned address is backward compatible
			osmoassert.ConditionalPanic(s.T(), tc.expectedPanic, func() {
				addr := mock_pool.GetAddress()
				s.Require().Equal(addr, s.TestAccs[0])
			})
		})
	}
}

// TestGetIncentivesAddress tests the GetIncentivesAddress method of pool
func (s *ConcentratedPoolTestSuite) TestGetIncentivesAddress() {

	tests := []struct {
		name          string
		expectedPanic bool
	}{
		{
			name: "Happy path",
		},
		{
			name:          "Unhappy path: wrong bech32 encoded address",
			expectedPanic: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a concentrated liquidity pool struct instance
			address := s.TestAccs[0].String()

			// if the test case is expected to panic, we use wrong bech32 encoded address instead
			if tc.expectedPanic {
				address = "osmo15l7yueqf3tx4cvpt6njvj7zxmvuhkwyrr509e9"
			}
			mock_pool := model.Pool{
				Id:                1,
				IncentivesAddress: address,
			}

			// Check that the returned address is backward compatible
			osmoassert.ConditionalPanic(s.T(), tc.expectedPanic, func() {
				addr := mock_pool.GetIncentivesAddress()
				s.Require().Equal(addr, s.TestAccs[0])
			})
		})
	}
}

// TestString tests if String method of the pool correctly json marshals the pool object
func (s *ConcentratedPoolTestSuite) TestString() {
	s.Setup()

	pool, err := model.NewConcentratedLiquidityPool(1, "foo", "bar", DefaultTickSpacing, DefaultSpreadFactor)
	s.Require().NoError(err)
	poolString := pool.String()
	s.Require().Equal(poolString, "{\"address\":\"osmo19e2mf7cywkv7zaug6nk5f87d07fxrdgrladvymh2gwv5crvm3vnsuewhh7\",\"incentives_address\":\"osmo156gncm3w2hdvuxxaejue8nejxgdgsrvdf7jftntuhxnaarhxcuas4ywjxf\",\"spread_rewards_address\":\"osmo10t3u6ze74jn7et6rluuxyf9vr2arykewmhcx67svg6heuu0gte2syfudcv\",\"id\":1,\"current_tick_liquidity\":\"0.000000000000000000\",\"token0\":\"foo\",\"token1\":\"bar\",\"current_sqrt_price\":\"0.000000000000000000000000000000000000\",\"tick_spacing\":1,\"exponent_at_price_one\":-6,\"spread_factor\":\"0.010000000000000000\",\"last_liquidity_update\":\"0001-01-01T00:00:00Z\"}")
}

// TestSpotPrice tests the SpotPrice method of the ConcentratedPoolTestSuite.
func (s *ConcentratedPoolTestSuite) TestSpotPrice() {
	type param struct {
		baseDenom  string
		quoteDenom string
	}

	tests := []struct {
		name              string
		param             param
		expectedSpotPrice osmomath.Dec
		expectedErr       error
	}{
		{
			name: "Happy path",
			param: param{
				baseDenom:  ETH,
				quoteDenom: USDC,
			},
			expectedSpotPrice: DefaultSpotPrice,
		},
		{
			name: "Happy path: reverse spot price",
			param: param{
				baseDenom:  USDC,
				quoteDenom: ETH,
			},
			expectedSpotPrice: DefaultReverseSpotPrice,
		},
		{
			name: "Error: quote asset denom does not exist in the pool",
			param: param{
				baseDenom:  ETH,
				quoteDenom: DAI,
			},
			expectedSpotPrice: osmomath.ZeroDec(),
			expectedErr:       fmt.Errorf("quote asset denom (%s) is not in the pool", DAI),
		},
		{
			name: "Error: base asset denom does not exist in the pool",
			param: param{
				baseDenom:  DAI,
				quoteDenom: ETH,
			},
			expectedSpotPrice: osmomath.ZeroDec(),
			expectedErr:       fmt.Errorf("base asset denom (%s) is not in the pool", DAI),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a concentrated liquidity pool struct instance
			mock_pool := model.Pool{
				CurrentSqrtPrice: DefaultSqrtSpotPrice,
				Token0:           ETH,
				Token1:           USDC,
			}

			// Check the spot price of the mock pool using the SpotPrice method.
			spotPriceFromMethod, err := mock_pool.SpotPrice(sdk.Context{}, tc.param.quoteDenom, tc.param.baseDenom)

			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectedErr)
			} else {
				s.Require().NoError(err)

				// We use elipson due to sqrt approximation
				elipson := osmomath.MustNewDecFromStr("0.0000000000000001")
				// TODO: truncation is acceptable temporary
				// remove before https://github.com/osmosis-labs/osmosis/issues/5726 is complete
				s.Require().True(spotPriceFromMethod.Dec().Sub(tc.expectedSpotPrice).Abs().LT(elipson))
			}
		})
	}
}

// TestUpdateLiquidity tests the UpdateLiquidity method of the ConcentratedPoolTestSuite.
func (s *ConcentratedPoolTestSuite) TestUpdateLiquidity() {
	mock_pool := model.Pool{
		CurrentTickLiquidity: DefaultLiquidityAmt,
	}

	// Try updating the liquidity with a zero osmomath.Dec value.
	mock_pool.UpdateLiquidity(osmomath.ZeroDec())

	// Assert that the liquidity has not changed.
	s.Require().Equal(DefaultLiquidityAmt, mock_pool.CurrentTickLiquidity)

	// Try adding 10 to the pool liquidity.
	mock_pool.UpdateLiquidity(osmomath.NewDec(10))

	// Assert that the liquidity has increased by 10.
	s.Require().Equal(DefaultLiquidityAmt.Add(osmomath.NewDec(10)), mock_pool.CurrentTickLiquidity)
}

func (s *ConcentratedPoolTestSuite) TestIsCurrentTickInRange() {
	s.Setup()
	currentTick := DefaultCurrTick

	tests := []struct {
		name           string
		lowerTick      int64
		upperTick      int64
		expectedResult bool
	}{
		{
			"given lower tick tick is within range of pool tick",
			DefaultCurrTick - 1,
			DefaultCurrTick + 1,
			true,
		},
		{
			"range only includes current tick",
			DefaultCurrTick,
			DefaultCurrTick + 1,
			true,
		},
		{
			"current tick is on upper tick",
			DefaultCurrTick - 3,
			DefaultCurrTick,
			false,
		},
		{
			"lower tick and upper tick are equal to pool tick",
			DefaultCurrTick,
			DefaultCurrTick,
			false,
		},
		{
			"only lower tick is equal to the pool tick",
			DefaultCurrTick,
			DefaultCurrTick + 3,
			true,
		},
		{
			"only upper tick is equal to the pool tick",
			DefaultCurrTick - 3,
			DefaultCurrTick,
			false,
		},
		{
			"lower tick is greater then pool tick",
			DefaultCurrTick + 1,
			DefaultCurrTick + 3,
			false,
		},
		{
			"upper tick is lower then pool tick",
			DefaultCurrTick - 3,
			DefaultCurrTick - 1,
			false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Create a concentrated liquidity pool struct instance
			mock_pool := model.Pool{
				CurrentTick: currentTick,
			}

			// System under test
			iscurrentTickInRange := mock_pool.IsCurrentTickInRange(tc.lowerTick, tc.upperTick)
			if tc.expectedResult {
				s.Require().True(iscurrentTickInRange)
			} else {
				s.Require().False(iscurrentTickInRange)
			}
		})
	}
}

func (s *ConcentratedPoolTestSuite) TestApplySwap() {
	// Set up the test suite.
	s.Setup()

	var (
		negativeOne    = osmomath.NewBigDec(-1)
		negativeOneDec = osmomath.OneDec().Neg()
	)

	tests := []struct {
		name             string
		currentLiquidity osmomath.Dec
		currentTick      int64
		currentSqrtPrice osmomath.BigDec
		newLiquidity     osmomath.Dec
		newTick          int64
		newSqrtPrice     osmomath.BigDec
		expectErr        error
	}{
		{
			name:             "positive liquidity and square root price",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      DefaultCurrTick,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt.MulInt64(2),
			newTick:          DefaultCurrTick * 2,
			newSqrtPrice:     DefaultCurrSqrtPrice.MulInt64(2),
			expectErr:        nil,
		},
		{
			name:             "negative liquidity",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      DefaultCurrTick,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     negativeOneDec,
			newTick:          DefaultCurrTick,
			newSqrtPrice:     DefaultCurrSqrtPrice,
			expectErr:        types.NegativeLiquidityError{Liquidity: negativeOneDec},
		},
		{
			name:             "negative square root price",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      DefaultCurrTick,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt,
			newTick:          DefaultCurrTick,
			newSqrtPrice:     negativeOne,
			expectErr:        types.SqrtPriceNegativeError{ProvidedSqrtPrice: negativeOne},
		},
		{
			name:             "new tick is equal to max tick",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      DefaultCurrTick,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt,
			newTick:          types.MaxTick,
			newSqrtPrice:     DefaultCurrSqrtPrice,
			expectErr:        nil,
		},
		{
			name:             "new tick is equal to min initialized tick",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      DefaultCurrTick,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt,
			newTick:          types.MinInitializedTick,
			newSqrtPrice:     DefaultCurrSqrtPrice,
			expectErr:        nil,
		},
		{
			name:             "new tick is equal to min current tick",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      DefaultCurrTick,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt,
			newTick:          types.MinCurrentTick,
			newSqrtPrice:     DefaultCurrSqrtPrice,
			expectErr:        nil,
		},
		{
			name:             "error: upper tick is greater than max tick",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      1,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt,
			newTick:          math.MaxInt64,
			newSqrtPrice:     DefaultCurrSqrtPrice,
			expectErr: types.TickIndexNotWithinBoundariesError{
				MaxTick:    types.MaxTick,
				MinTick:    types.MinCurrentTick,
				ActualTick: math.MaxInt64,
			},
		},
		{
			name:             "error: lower tick is smaller than min tick",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      1,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt,
			newTick:          math.MinInt64,
			newSqrtPrice:     DefaultCurrSqrtPrice,
			expectErr: types.TickIndexNotWithinBoundariesError{
				MaxTick:    types.MaxTick,
				MinTick:    types.MinCurrentTick,
				ActualTick: math.MinInt64,
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Create a concentrated liquidity pool struct instance
			mock_pool := model.Pool{
				ExponentAtPriceOne:   types.ExponentAtPriceOne,
				CurrentTickLiquidity: tt.currentLiquidity,
				CurrentTick:          tt.currentTick,
				CurrentSqrtPrice:     tt.currentSqrtPrice,
			}

			// Apply the new values to the mock pool using the ApplySwap method.
			err := mock_pool.ApplySwap(tt.newLiquidity, tt.newTick, tt.newSqrtPrice)

			if tt.expectErr != nil {
				s.Require().ErrorIs(tt.expectErr, err)
				return
			}

			// Assert that the values in the mock pool have been updated.
			s.Require().Equal(tt.newLiquidity, mock_pool.CurrentTickLiquidity)
			s.Require().Equal(tt.newTick, mock_pool.CurrentTick)
			s.Require().Equal(tt.newSqrtPrice, mock_pool.CurrentSqrtPrice)
		})
	}
}

// TestNewConcentratedLiquidityPool is a test suite that tests the NewConcentratedLiquidityPool function.
func (s *ConcentratedPoolTestSuite) TestNewConcentratedLiquidityPool() {
	type param struct {
		poolId       uint64
		denom0       string
		denom1       string
		tickSpacing  uint64
		spreadFactor osmomath.Dec
	}

	tests := []struct {
		name                string
		param               param
		expectedPoolId      uint64
		expectedDenom0      string
		expectedDenom1      string
		expectedTickSpacing uint64
		expectedErr         error
	}{
		{
			name: "Happy path",
			param: param{
				poolId:       DefaultValidPoolID,
				denom0:       ETH,
				denom1:       USDC,
				tickSpacing:  DefaultTickSpacing,
				spreadFactor: DefaultSpreadFactor,
			},
			expectedPoolId:      DefaultValidPoolID,
			expectedDenom0:      ETH,
			expectedDenom1:      USDC,
			expectedTickSpacing: DefaultTickSpacing,
		},
		{
			name: "Non lexicographical order of denoms should not get reordered",
			param: param{
				poolId:       DefaultValidPoolID,
				denom0:       USDC,
				denom1:       ETH,
				tickSpacing:  DefaultTickSpacing,
				spreadFactor: osmomath.ZeroDec(),
			},
			expectedPoolId:      DefaultValidPoolID,
			expectedDenom0:      USDC,
			expectedDenom1:      ETH,
			expectedTickSpacing: DefaultTickSpacing,
		},

		{
			name: "Error: same denom not allowed",
			param: param{
				poolId:       DefaultValidPoolID,
				denom0:       USDC,
				denom1:       USDC,
				tickSpacing:  DefaultTickSpacing,
				spreadFactor: DefaultSpreadFactor,
			},
			expectedErr: types.MatchingDenomError{Denom: USDC},
		},
		{
			name: "Error: negative spread factor",
			param: param{
				poolId:       DefaultValidPoolID,
				denom0:       ETH,
				denom1:       USDC,
				tickSpacing:  DefaultTickSpacing,
				spreadFactor: osmomath.ZeroDec().Sub(osmomath.SmallestDec()),
			},
			expectedErr: types.InvalidSpreadFactorError{ActualSpreadFactor: osmomath.ZeroDec().Sub(osmomath.SmallestDec())},
		},
		{
			name: "Error: spread factor == 1",
			param: param{
				poolId:       DefaultValidPoolID,
				denom0:       ETH,
				denom1:       USDC,
				tickSpacing:  DefaultTickSpacing,
				spreadFactor: osmomath.OneDec(),
			},
			expectedErr: types.InvalidSpreadFactorError{ActualSpreadFactor: osmomath.OneDec()},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Call NewConcentratedLiquidityPool with the parameters from the current test.
			pool, err := model.NewConcentratedLiquidityPool(test.param.poolId, test.param.denom0, test.param.denom1, test.param.tickSpacing, test.param.spreadFactor)

			if test.expectedErr != nil {
				// If the test is expected to produce an error, check if it does.
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedErr.Error())
			} else {
				// If the test is not expected to produce an error, check if it doesn't.
				s.Require().NoError(err)

				// Check if the values of the returned pool match the expected values.
				s.Require().Equal(test.expectedPoolId, pool.Id)
				s.Require().Equal(test.expectedDenom0, pool.Token0)
				s.Require().Equal(test.expectedDenom1, pool.Token1)
				s.Require().Equal(test.expectedTickSpacing, pool.TickSpacing)
				s.Require().Equal(test.param.spreadFactor, pool.SpreadFactor)
			}
		})
	}
}

func (suite *ConcentratedPoolTestSuite) TestCalcActualAmounts() {
	var (
		tickToSqrtPrice = func(tick int64) osmomath.BigDec {
			sqrtPrice, err := clmath.TickToSqrtPrice(tick)
			suite.Require().NoError(err)
			return sqrtPrice
		}

		defaultLiquidityDelta = osmomath.NewDec(1000)

		lowerTick            = int64(-99)
		lowerSqrtPriceBigDec = tickToSqrtPrice(lowerTick)

		midtick            = int64(2)
		midSqrtPriceBigDec = tickToSqrtPrice(midtick)

		uppertick            = int64(74)
		upperSqrtPriceBigDec = tickToSqrtPrice(uppertick)
	)

	tests := map[string]struct {
		currentTick                 int64
		lowerTick                   int64
		upperTick                   int64
		liquidityDelta              osmomath.Dec
		shouldTestRoundingInvariant bool
		expectError                 error

		expectedAmount0 osmomath.Dec
		expectedAmount1 osmomath.Dec
	}{
		"current in range, positive liquidity": {
			currentTick:                 midtick,
			lowerTick:                   lowerTick,
			upperTick:                   uppertick,
			liquidityDelta:              defaultLiquidityDelta,
			shouldTestRoundingInvariant: true,

			expectedAmount0: clmath.CalcAmount0Delta(defaultLiquidityDelta, midSqrtPriceBigDec, upperSqrtPriceBigDec, true).Dec(),
			expectedAmount1: clmath.CalcAmount1Delta(defaultLiquidityDelta, midSqrtPriceBigDec, lowerSqrtPriceBigDec, true).Dec(),
		},
		"current in range, negative liquidity": {
			currentTick:    midtick,
			lowerTick:      lowerTick,
			upperTick:      uppertick,
			liquidityDelta: defaultLiquidityDelta.Neg(),

			expectedAmount0: clmath.CalcAmount0Delta(defaultLiquidityDelta.Neg(), midSqrtPriceBigDec, upperSqrtPriceBigDec, false).Dec(),
			expectedAmount1: clmath.CalcAmount1Delta(defaultLiquidityDelta.Neg(), midSqrtPriceBigDec, lowerSqrtPriceBigDec, false).Dec(),
		},
		"current below range, positive liquidity": {
			currentTick:    lowerTick,
			lowerTick:      midtick,
			upperTick:      uppertick,
			liquidityDelta: defaultLiquidityDelta,

			expectedAmount0: clmath.CalcAmount0Delta(defaultLiquidityDelta, midSqrtPriceBigDec, upperSqrtPriceBigDec, true).Dec(),
			expectedAmount1: osmomath.ZeroDec(),
		},
		"current below range, negative liquidity": {
			currentTick:    lowerTick,
			lowerTick:      midtick,
			upperTick:      uppertick,
			liquidityDelta: defaultLiquidityDelta.Neg(),

			expectedAmount0: clmath.CalcAmount0Delta(defaultLiquidityDelta.Neg(), midSqrtPriceBigDec, upperSqrtPriceBigDec, false).Dec(),
			expectedAmount1: osmomath.ZeroDec(),
		},
		"current above range, positive liquidity": {
			currentTick:    uppertick,
			lowerTick:      lowerTick,
			upperTick:      midtick,
			liquidityDelta: defaultLiquidityDelta,

			expectedAmount0: osmomath.ZeroDec(),
			expectedAmount1: clmath.CalcAmount1Delta(defaultLiquidityDelta, lowerSqrtPriceBigDec, midSqrtPriceBigDec, true).Dec(),
		},
		"current above range, negative liquidity": {
			currentTick:    uppertick,
			lowerTick:      lowerTick,
			upperTick:      midtick,
			liquidityDelta: defaultLiquidityDelta.Neg(),

			expectedAmount0: osmomath.ZeroDec(),
			expectedAmount1: clmath.CalcAmount1Delta(defaultLiquidityDelta.Neg(), lowerSqrtPriceBigDec, midSqrtPriceBigDec, false).Dec(),
		},

		// errors
		"error: zero liqudiity": {
			currentTick:    midtick,
			lowerTick:      lowerTick,
			upperTick:      uppertick,
			liquidityDelta: osmomath.ZeroDec(),

			expectError: types.ErrZeroLiquidity,
		},
		"error: lower tick equals upper tick": {
			currentTick:    lowerTick,
			lowerTick:      lowerTick,
			upperTick:      lowerTick,
			liquidityDelta: defaultLiquidityDelta,

			expectError: types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: lowerTick},
		},
		"error: lower tick is greater than upper tick": {
			currentTick:    lowerTick,
			lowerTick:      lowerTick + 1,
			upperTick:      lowerTick,
			liquidityDelta: defaultLiquidityDelta,

			expectError: types.InvalidLowerUpperTickError{LowerTick: lowerTick + 1, UpperTick: lowerTick},
		},
		"error: lower tick is equal to upper tick": {
			currentTick:    lowerTick,
			lowerTick:      lowerTick,
			upperTick:      lowerTick,
			liquidityDelta: defaultLiquidityDelta,

			expectError: types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: lowerTick},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.Setup()

			pool := model.Pool{
				CurrentTick: tc.currentTick,
			}
			currenTicktSqrtPrice, _ := clmath.TickToSqrtPrice(pool.CurrentTick)
			pool.CurrentSqrtPrice = currenTicktSqrtPrice

			actualAmount0, actualAmount1, err := pool.CalcActualAmounts(suite.Ctx, tc.lowerTick, tc.upperTick, tc.liquidityDelta)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectError)
				return
			}
			suite.Require().NoError(err)

			suite.Require().Equal(tc.expectedAmount0, actualAmount0)
			suite.Require().Equal(tc.expectedAmount1, actualAmount1)

			// Note: to test rounding invariants around positive and negative liquidity.
			if tc.shouldTestRoundingInvariant {
				actualAmount0Neg, actualAmount1Neg, err := pool.CalcActualAmounts(suite.Ctx, tc.lowerTick, tc.upperTick, tc.liquidityDelta.Neg())
				suite.Require().NoError(err)

				amt0Diff := actualAmount0.Sub(actualAmount0Neg.Neg())
				amt1Diff := actualAmount1.Sub(actualAmount1Neg.Neg())

				// Difference is between 0 and 1 due to positive liquidity rounding up and negative liquidity performing math normally.
				suite.Require().True(amt0Diff.IsPositive() && amt0Diff.LT(osmomath.OneDec()))
				suite.Require().True(amt1Diff.IsPositive() && amt1Diff.LT(osmomath.OneDec()))
			}
		})
	}
}

func (suite *ConcentratedPoolTestSuite) TestUpdateLiquidityIfActivePosition() {
	var (
		defaultLiquidityDelta = osmomath.NewDec(1000)
		defaultLiquidityAmt   = osmomath.NewDec(1000)

		lowerTick = int64(-99)
		midtick   = int64(2)
		uppertick = int64(74)
	)

	tests := map[string]struct {
		currentTick    int64
		lowerTick      int64
		upperTick      int64
		liquidityDelta osmomath.Dec
		expectError    error
	}{
		"current in range, positive liquidity": {
			currentTick:    midtick,
			lowerTick:      lowerTick,
			upperTick:      uppertick,
			liquidityDelta: defaultLiquidityDelta,
		},
		"current in range, negative liquidity": {
			currentTick:    midtick,
			lowerTick:      lowerTick,
			upperTick:      uppertick,
			liquidityDelta: defaultLiquidityDelta.Neg(),
		},
		"current below range, positive liquidity": {
			currentTick:    lowerTick,
			lowerTick:      midtick,
			upperTick:      uppertick,
			liquidityDelta: defaultLiquidityDelta,
		},
		"current below range, negative liquidity": {
			currentTick:    lowerTick,
			lowerTick:      midtick,
			upperTick:      uppertick,
			liquidityDelta: defaultLiquidityDelta.Neg(),
		},
		"current above range, positive liquidity": {
			currentTick:    uppertick,
			lowerTick:      lowerTick,
			upperTick:      midtick,
			liquidityDelta: defaultLiquidityDelta,
		},
		"current above range, negative liquidity": {
			currentTick:    uppertick,
			lowerTick:      lowerTick,
			upperTick:      midtick,
			liquidityDelta: defaultLiquidityDelta.Neg(),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.Setup()

			pool := model.Pool{
				CurrentTick:          tc.currentTick,
				CurrentTickLiquidity: defaultLiquidityAmt,
			}
			currenTicktSqrtPrice, _ := clmath.TickToSqrtPrice(pool.CurrentTick)
			pool.CurrentSqrtPrice = currenTicktSqrtPrice

			wasUpdated := pool.UpdateLiquidityIfActivePosition(suite.Ctx, tc.lowerTick, tc.upperTick, tc.liquidityDelta)
			if tc.lowerTick <= tc.currentTick && tc.currentTick <= tc.upperTick {
				suite.Require().True(wasUpdated)
				expectedCurrentTickLiquidity := defaultLiquidityAmt.Add(tc.liquidityDelta)
				suite.Require().Equal(expectedCurrentTickLiquidity, pool.CurrentTickLiquidity)
			} else {
				suite.Require().False(wasUpdated)
				suite.Require().Equal(defaultLiquidityAmt, pool.CurrentTickLiquidity)
			}
		})
	}
}

func (suite *ConcentratedPoolTestSuite) TestPoolSetMethods() {
	var (
		newCurrentTick      = DefaultCurrTick
		newCurrentSqrtPrice = DefaultCurrSqrtPrice
		newTickSpacing      = DefaultTickSpacing
	)

	tests := map[string]struct {
		currentTick              int64
		currentSqrtPrice         osmomath.BigDec
		tickSpacing              uint64
		lastLiquidityUpdateDelta time.Duration
	}{
		"happy path": {
			currentTick:              newCurrentTick,
			currentSqrtPrice:         newCurrentSqrtPrice,
			tickSpacing:              newTickSpacing,
			lastLiquidityUpdateDelta: time.Hour,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.Setup()

			currentBlockTime := suite.Ctx.BlockTime()

			// Create the pool and check that the initial values are not equal to the new values we will set.
			clPool := suite.PrepareConcentratedPool()
			suite.Require().NotEqual(tc.currentTick, clPool.GetCurrentTick())
			suite.Require().NotEqual(tc.currentSqrtPrice, clPool.GetCurrentSqrtPrice())
			suite.Require().NotEqual(tc.tickSpacing, clPool.GetTickSpacing())
			suite.Require().NotEqual(currentBlockTime.Add(tc.lastLiquidityUpdateDelta), clPool.GetLastLiquidityUpdate())

			// Run the setters.
			clPool.SetCurrentTick(tc.currentTick)
			clPool.SetCurrentSqrtPrice(tc.currentSqrtPrice)
			clPool.SetTickSpacing(tc.tickSpacing)
			clPool.SetLastLiquidityUpdate(currentBlockTime.Add(tc.lastLiquidityUpdateDelta))

			// Check that the values are now equal to the new values.
			suite.Require().Equal(tc.currentTick, clPool.GetCurrentTick())
			suite.Require().Equal(tc.currentSqrtPrice, clPool.GetCurrentSqrtPrice())
			suite.Require().Equal(tc.tickSpacing, clPool.GetTickSpacing())
			suite.Require().Equal(currentBlockTime.Add(tc.lastLiquidityUpdateDelta), clPool.GetLastLiquidityUpdate())
		})
	}
}

// Test that the right denoms are returned.
func (s *ConcentratedPoolTestSuite) TestGetPoolDenoms() {
	s.Setup()

	const (
		expectedDenom1 = "bar"
		expectedDenom2 = "foo"
	)

	pool := s.PrepareConcentratedPoolWithCoins(expectedDenom1, expectedDenom2)

	denoms := pool.GetPoolDenoms(s.Ctx)
	s.Require().Equal(2, len(denoms))
	s.Require().Equal(expectedDenom1, denoms[0])
	s.Require().Equal(expectedDenom2, denoms[1])
}
