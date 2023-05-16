package model_test

import (
	fmt "fmt"
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	clmath "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

const (
	ETH                = "eth"
	USDC               = "usdc"
	DAI                = "dai"
	DefaultValidPoolID = uint64(1)
	DefaultTickSpacing = uint64(1)
)

var (
	DefaultSpotPrice        = sdk.MustNewDecFromStr("0.2")
	DefaultReverseSpotPrice = sdk.NewDec(1).Quo(DefaultSpotPrice)
	DefaultSqrtSpotPrice, _ = DefaultSpotPrice.ApproxSqrt()
	DefaultLiquidityAmt     = sdk.MustNewDecFromStr("1517882343.751510418088349649")
	DefaultCurrTick         = sdk.NewInt(310000)
	DefaultCurrPrice        = sdk.NewDec(5000)
	DefaultCurrSqrtPrice, _ = DefaultCurrPrice.ApproxSqrt() // 70.710678118654752440
	DefaultSwapFee          = sdk.MustNewDecFromStr("0.01")
)

type ConcentratedPoolTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestConcentratedPoolTestSuite(t *testing.T) {
	suite.Run(t, new(ConcentratedPoolTestSuite))
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
			DefaultCurrTick.Int64() - 1,
			DefaultCurrTick.Int64() + 1,
			true,
		},
		{
			"lower tick and upper tick are equal to pool tick",
			DefaultCurrTick.Int64(),
			DefaultCurrTick.Int64(),
			true,
		},
		{
			"lower tick is greater then pool tick",
			DefaultCurrTick.Int64() + 1,
			DefaultCurrTick.Int64() + 3,
			false,
		},
		{
			"upper tick is lower then pool tick",
			DefaultCurrTick.Int64() - 3,
			DefaultCurrTick.Int64() - 1,
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
		currentTick      sdk.Int
		currentSqrtPrice sdk.Dec
		newLiquidity     sdk.Dec
		newTick          sdk.Int
		newSqrtPrice     sdk.Dec
		expectErr        error
	}{
		{
			name:             "positive liquidity and square root price",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      DefaultCurrTick,
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt.Mul(sdk.NewDec(2)),
			newTick:          DefaultCurrTick.Mul(sdk.NewInt(2)),
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
			name:             "upper tick too big",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      sdk.NewInt(1),
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt,
			newTick:          sdk.NewInt(math.MaxInt64),
			newSqrtPrice:     DefaultCurrSqrtPrice,
			expectErr: types.TickIndexNotWithinBoundariesError{
				MaxTick:    types.MaxTick,
				MinTick:    types.MinTick,
				ActualTick: math.MaxInt64,
			},
		},
		{
			name:             "lower tick too small",
			currentLiquidity: DefaultLiquidityAmt,
			currentTick:      sdk.NewInt(1),
			currentSqrtPrice: DefaultCurrSqrtPrice,
			newLiquidity:     DefaultLiquidityAmt,
			newTick:          sdk.NewInt(math.MinInt64),
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
		poolId      uint64
		denom0      string
		denom1      string
		tickSpacing uint64
		swapFee     sdk.Dec
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
				poolId:      DefaultValidPoolID,
				denom0:      ETH,
				denom1:      USDC,
				tickSpacing: DefaultTickSpacing,
				swapFee:     DefaultSwapFee,
			},
			expectedPoolId:      DefaultValidPoolID,
			expectedDenom0:      ETH,
			expectedDenom1:      USDC,
			expectedTickSpacing: DefaultTickSpacing,
		},
		{
			name: "Non lexicographical order of denoms should not get reordered",
			param: param{
				poolId:      DefaultValidPoolID,
				denom0:      USDC,
				denom1:      ETH,
				tickSpacing: DefaultTickSpacing,
				swapFee:     sdk.ZeroDec(),
			},
			expectedPoolId:      DefaultValidPoolID,
			expectedDenom0:      USDC,
			expectedDenom1:      ETH,
			expectedTickSpacing: DefaultTickSpacing,
		},

		{
			name: "Error: same denom not allowed",
			param: param{
				poolId:      DefaultValidPoolID,
				denom0:      USDC,
				denom1:      USDC,
				tickSpacing: DefaultTickSpacing,
				swapFee:     DefaultSwapFee,
			},
			expectedErr: types.MatchingDenomError{Denom: USDC},
		},
		{
			name: "Error: negative swap fee",
			param: param{
				poolId:      DefaultValidPoolID,
				denom0:      ETH,
				denom1:      USDC,
				tickSpacing: DefaultTickSpacing,
				swapFee:     sdk.ZeroDec().Sub(sdk.SmallestDec()),
			},
			expectedErr: types.InvalidSwapFeeError{ActualFee: sdk.ZeroDec().Sub(sdk.SmallestDec())},
		},
		{
			name: "Error: swap fee == 1",
			param: param{
				poolId:      DefaultValidPoolID,
				denom0:      ETH,
				denom1:      USDC,
				tickSpacing: DefaultTickSpacing,
				swapFee:     sdk.OneDec(),
			},
			expectedErr: types.InvalidSwapFeeError{ActualFee: sdk.OneDec()},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Call NewConcentratedLiquidityPool with the parameters from the current test.
			pool, err := model.NewConcentratedLiquidityPool(test.param.poolId, test.param.denom0, test.param.denom1, test.param.tickSpacing, test.param.swapFee)

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
				s.Require().Equal(test.param.swapFee, pool.SwapFee)
			}
		})
	}
}

func (suite *ConcentratedPoolTestSuite) TestCalcActualAmounts() {
	var (
		tickToSqrtPrice = func(tick int64) sdk.Dec {
			sqrtPrice, err := clmath.TickToSqrtPrice(sdk.NewInt(tick))
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
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.Setup()

			pool := model.Pool{
				CurrentTick: sdk.NewInt(tc.currentTick),
			}
			pool.CurrentSqrtPrice, _ = clmath.TickToSqrtPrice(pool.CurrentTick)

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
				CurrentTick:          sdk.NewInt(tc.currentTick),
				CurrentTickLiquidity: defaultLiquidityAmt,
			}
			pool.CurrentSqrtPrice, _ = clmath.TickToSqrtPrice(pool.CurrentTick)

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
