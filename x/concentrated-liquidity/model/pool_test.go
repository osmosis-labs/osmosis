package model_test

import (
	fmt "fmt"
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
	clmath "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

const (
	ETH                = "eth"
	USDC               = "usdc"
	DAI                = "dai"
	DefaultValidPoolID = uint64(1)
	DefaultTickSpacing = uint64(1)
)

var (
	DefaultSpotPrice              = sdk.MustNewDecFromStr("0.2")
	DefaultReverseSpotPrice       = sdk.NewDec(1).Quo(DefaultSpotPrice)
	DefaultSqrtSpotPrice, _       = DefaultSpotPrice.ApproxSqrt()
	DefaultLiquidityAmt           = sdk.MustNewDecFromStr("1517882343.751510418088349649")
	DefaultCurrTick         int64 = 310000
	DefaultCurrPrice              = sdk.NewDec(5000)
	DefaultCurrSqrtPrice, _       = DefaultCurrPrice.ApproxSqrt() // 70.710678118654752440
	DefaultSpreadFactor           = sdk.MustNewDecFromStr("0.01")
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
	s.Require().Equal(poolString, "{\"address\":\"osmo19e2mf7cywkv7zaug6nk5f87d07fxrdgrladvymh2gwv5crvm3vnsuewhh7\",\"incentives_address\":\"osmo156gncm3w2hdvuxxaejue8nejxgdgsrvdf7jftntuhxnaarhxcuas4ywjxf\",\"spread_rewards_address\":\"osmo10t3u6ze74jn7et6rluuxyf9vr2arykewmhcx67svg6heuu0gte2syfudcv\",\"id\":1,\"current_tick_liquidity\":\"0.000000000000000000\",\"token0\":\"foo\",\"token1\":\"bar\",\"current_sqrt_price\":\"0.000000000000000000\",\"tick_spacing\":1,\"exponent_at_price_one\":-6,\"spread_factor\":\"0.010000000000000000\",\"last_liquidity_update\":\"0001-01-01T00:00:00Z\"}")
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
		expectedSpotPrice sdk.Dec
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
			expectedSpotPrice: sdk.ZeroDec(),
			expectedErr:       fmt.Errorf("quote asset denom (%s) is not in the pool", DAI),
		},
		{
			name: "Error: base asset denom does not exist in the pool",
			param: param{
				baseDenom:  DAI,
				quoteDenom: ETH,
			},
			expectedSpotPrice: sdk.ZeroDec(),
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
			spotPriceFromMethod, err := mock_pool.SpotPrice(sdk.Context{}, tc.param.baseDenom, tc.param.quoteDenom)

			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectedErr)
			} else {
				s.Require().NoError(err)

				// We use elipson due to sqrt approximation
				elipson := sdk.MustNewDecFromStr("0.0000000000000001")
				s.Require().True(spotPriceFromMethod.Sub(tc.expectedSpotPrice).Abs().LT(elipson))
			}
		})
	}
}

// TestUpdateLiquidity tests the UpdateLiquidity method of the ConcentratedPoolTestSuite.
func (s *ConcentratedPoolTestSuite) TestUpdateLiquidity() {
	mock_pool := model.Pool{
		CurrentTickLiquidity: DefaultLiquidityAmt,
	}

	// Try updating the liquidity with a zero sdk.Dec value.
	mock_pool.UpdateLiquidity(sdk.ZeroDec())

	// Assert that the liquidity has not changed.
	s.Require().Equal(DefaultLiquidityAmt, mock_pool.CurrentTickLiquidity)

	// Try adding 10 to the pool liquidity.
	mock_pool.UpdateLiquidity(sdk.NewDec(10))

	// Assert that the liquidity has increased by 10.
	s.Require().Equal(DefaultLiquidityAmt.Add(sdk.NewDec(10)), mock_pool.CurrentTickLiquidity)
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

	negativeOne := sdk.NewDec(-1)
	tests := []struct {
		name             string
		currentLiquidity sdk.Dec
		currentTick      int64
		currentSqrtPrice sdk.Dec
		newLiquidity     sdk.Dec
		newTick          int64
		newSqrtPrice     sdk.Dec
		expectErr        error
	}{
		{
			name:             "positive liquidity and square root price",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      DefaultCurrTick,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt.Mul(sdk.NewDec(2)),
			newTick:          DefaultCurrTick * 2,
			newSqrtPrice:     DefaultCurrSqrtPrice.Mul(sdk.NewDec(2)),
			expectErr:        nil,
		},
		{
			name:             "negative liquidity",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      DefaultCurrTick,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     negativeOne,
			newTick:          DefaultCurrTick,
			newSqrtPrice:     DefaultCurrSqrtPrice,
			expectErr:        types.NegativeLiquidityError{Liquidity: negativeOne},
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
			name:             "upper tick is greater than max tick",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      1,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt,
			newTick:          math.MaxInt64,
			newSqrtPrice:     DefaultCurrSqrtPrice,
			expectErr: types.TickIndexNotWithinBoundariesError{
				MaxTick:    types.MaxTick,
				MinTick:    types.MinTick,
				ActualTick: math.MaxInt64,
			},
		},
		{
			name:             "lower tick is smaller than min tick",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      1,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt,
			newTick:          math.MinInt64,
			newSqrtPrice:     DefaultCurrSqrtPrice,
			expectErr: types.TickIndexNotWithinBoundariesError{
				MaxTick:    types.MaxTick,
				MinTick:    types.MinTick,
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
		spreadFactor sdk.Dec
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
				spreadFactor: sdk.ZeroDec(),
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
				spreadFactor: sdk.ZeroDec().Sub(sdk.SmallestDec()),
			},
			expectedErr: types.InvalidSpreadFactorError{ActualSpreadFactor: sdk.ZeroDec().Sub(sdk.SmallestDec())},
		},
		{
			name: "Error: spread factor == 1",
			param: param{
				poolId:       DefaultValidPoolID,
				denom0:       ETH,
				denom1:       USDC,
				tickSpacing:  DefaultTickSpacing,
				spreadFactor: sdk.OneDec(),
			},
			expectedErr: types.InvalidSpreadFactorError{ActualSpreadFactor: sdk.OneDec()},
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
		tickToSqrtPrice = func(tick int64) sdk.Dec {
			_, sqrtPrice, err := clmath.TickToSqrtPrice(tick)
			suite.Require().NoError(err)
			return sqrtPrice
		}

		defaultLiquidityDelta = sdk.NewDec(1000)

		lowerTick      = int64(-99)
		lowerSqrtPrice = tickToSqrtPrice(lowerTick)

		midtick      = int64(2)
		midSqrtPrice = tickToSqrtPrice(midtick)

		uppertick      = int64(74)
		upperSqrtPrice = tickToSqrtPrice(uppertick)
	)

	tests := map[string]struct {
		currentTick                 int64
		lowerTick                   int64
		upperTick                   int64
		liquidityDelta              sdk.Dec
		shouldTestRoundingInvariant bool
		expectError                 error

		expectedAmount0 sdk.Dec
		expectedAmount1 sdk.Dec
	}{
		"current in range, positive liquidity": {
			currentTick:                 midtick,
			lowerTick:                   lowerTick,
			upperTick:                   uppertick,
			liquidityDelta:              defaultLiquidityDelta,
			shouldTestRoundingInvariant: true,

			expectedAmount0: clmath.CalcAmount0Delta(defaultLiquidityDelta, midSqrtPrice, upperSqrtPrice, true),
			expectedAmount1: clmath.CalcAmount1Delta(defaultLiquidityDelta, midSqrtPrice, lowerSqrtPrice, true),
		},
		"current in range, negative liquidity": {
			currentTick:    midtick,
			lowerTick:      lowerTick,
			upperTick:      uppertick,
			liquidityDelta: defaultLiquidityDelta.Neg(),

			expectedAmount0: clmath.CalcAmount0Delta(defaultLiquidityDelta.Neg(), midSqrtPrice, upperSqrtPrice, false),
			expectedAmount1: clmath.CalcAmount1Delta(defaultLiquidityDelta.Neg(), midSqrtPrice, lowerSqrtPrice, false),
		},
		"current below range, positive liquidity": {
			currentTick:    lowerTick,
			lowerTick:      midtick,
			upperTick:      uppertick,
			liquidityDelta: defaultLiquidityDelta,

			expectedAmount0: clmath.CalcAmount0Delta(defaultLiquidityDelta, midSqrtPrice, upperSqrtPrice, true),
			expectedAmount1: sdk.ZeroDec(),
		},
		"current below range, negative liquidity": {
			currentTick:    lowerTick,
			lowerTick:      midtick,
			upperTick:      uppertick,
			liquidityDelta: defaultLiquidityDelta.Neg(),

			expectedAmount0: clmath.CalcAmount0Delta(defaultLiquidityDelta.Neg(), midSqrtPrice, upperSqrtPrice, false),
			expectedAmount1: sdk.ZeroDec(),
		},
		"current above range, positive liquidity": {
			currentTick:    uppertick,
			lowerTick:      lowerTick,
			upperTick:      midtick,
			liquidityDelta: defaultLiquidityDelta,

			expectedAmount0: sdk.ZeroDec(),
			expectedAmount1: clmath.CalcAmount1Delta(defaultLiquidityDelta, lowerSqrtPrice, midSqrtPrice, true),
		},
		"current above range, negative liquidity": {
			currentTick:    uppertick,
			lowerTick:      lowerTick,
			upperTick:      midtick,
			liquidityDelta: defaultLiquidityDelta.Neg(),

			expectedAmount0: sdk.ZeroDec(),
			expectedAmount1: clmath.CalcAmount1Delta(defaultLiquidityDelta.Neg(), tickToSqrtPrice(lowerTick), midSqrtPrice, false),
		},

		// errors
		"error: zero liqudiity": {
			currentTick:    midtick,
			lowerTick:      lowerTick,
			upperTick:      uppertick,
			liquidityDelta: sdk.ZeroDec(),

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
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.Setup()

			pool := model.Pool{
				CurrentTick: tc.currentTick,
			}
			_, pool.CurrentSqrtPrice, _ = clmath.TickToSqrtPrice(pool.CurrentTick)

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
				suite.Require().True(amt0Diff.GT(sdk.ZeroDec()) && amt0Diff.LT(sdk.OneDec()))
				suite.Require().True(amt1Diff.GT(sdk.ZeroDec()) && amt1Diff.LT(sdk.OneDec()))
			}
		})
	}
}

func (suite *ConcentratedPoolTestSuite) TestUpdateLiquidityIfActivePosition() {
	var (
		defaultLiquidityDelta = sdk.NewDec(1000)
		defaultLiquidityAmt   = sdk.NewDec(1000)

		lowerTick = int64(-99)
		midtick   = int64(2)
		uppertick = int64(74)
	)

	tests := map[string]struct {
		currentTick    int64
		lowerTick      int64
		upperTick      int64
		liquidityDelta sdk.Dec
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
			_, pool.CurrentSqrtPrice, _ = clmath.TickToSqrtPrice(pool.CurrentTick)

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
