package model_test

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
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
	DefaultSpotPrice          = sdk.MustNewDecFromStr("0.2")
	DefaultReverseSpotPrice   = sdk.NewDec(1).Quo(DefaultSpotPrice)
	DefaultSqrtSpotPrice, _   = DefaultSpotPrice.ApproxSqrt()
	DefaultLiquidityAmt       = sdk.MustNewDecFromStr("1517882343.751510418088349649")
	DefaultCurrTick           = sdk.NewInt(310000)
	DefaultCurrPrice          = sdk.NewDec(5000)
	DefaultCurrSqrtPrice, _   = DefaultCurrPrice.ApproxSqrt() // 70.710678118654752440
	DefaultExponentAtPriceOne = sdk.NewInt(-4)
	DefaultSwapFee            = sdk.MustNewDecFromStr("0.01")
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
		Liquidity: DefaultLiquidityAmt,
	}

	// Try updating the liquidity with a zero sdk.Dec value.
	mock_pool.UpdateLiquidity(sdk.ZeroDec())

	// Assert that the liquidity has not changed.
	s.Require().Equal(DefaultLiquidityAmt, mock_pool.Liquidity)

	// Try adding 10 to the pool liquidity.
	mock_pool.UpdateLiquidity(sdk.NewDec(10))

	// Assert that the liquidity has increased by 10.
	s.Require().Equal(DefaultLiquidityAmt.Add(sdk.NewDec(10)), mock_pool.Liquidity)
}

func (s *ConcentratedPoolTestSuite) TestApplySwap() {
	// Set up the test suite.
	s.Setup()

	// Create a concentrated liquidity pool struct instance
	mock_pool := model.Pool{
		Liquidity:        DefaultLiquidityAmt,
		CurrentTick:      DefaultCurrTick,
		CurrentSqrtPrice: DefaultCurrSqrtPrice,
	}

	// Create new values for liquidity, current tick, and current square root price.
	newLiquidity := DefaultLiquidityAmt.Mul(sdk.NewDec(2))
	newCurrTick := DefaultCurrTick.Mul(sdk.NewInt(2))
	newCurrSqrtPrice := DefaultCurrSqrtPrice.Mul(sdk.NewDec(2))

	// Apply the new values to the mock pool using the ApplySwap method.
	mock_pool.ApplySwap(newLiquidity, newCurrTick, newCurrSqrtPrice)

	// Assert that the values in the mock pool have been updated.
	s.Require().Equal(mock_pool.Liquidity, newLiquidity)
	s.Require().Equal(mock_pool.CurrentTick, newCurrTick)
	s.Require().Equal(mock_pool.CurrentSqrtPrice, newCurrSqrtPrice)
}

// TestNewConcentratedLiquidityPool is a test suite that tests the NewConcentratedLiquidityPool function.
func (s *ConcentratedPoolTestSuite) TestNewConcentratedLiquidityPool() {
	type param struct {
		poolId         uint64
		denom0         string
		denom1         string
		tickSpacing    uint64
		precisionValue sdk.Int
		swapFee        sdk.Dec
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
				poolId:         DefaultValidPoolID,
				denom0:         ETH,
				denom1:         USDC,
				tickSpacing:    DefaultTickSpacing,
				precisionValue: DefaultExponentAtPriceOne,
				swapFee:        DefaultSwapFee,
			},
			expectedPoolId:      DefaultValidPoolID,
			expectedDenom0:      ETH,
			expectedDenom1:      USDC,
			expectedTickSpacing: DefaultTickSpacing,
		},
		{
			name: "Non lexicographical order of denoms should get reordered",
			param: param{
				poolId:         DefaultValidPoolID,
				denom0:         USDC,
				denom1:         ETH,
				tickSpacing:    DefaultTickSpacing,
				precisionValue: DefaultExponentAtPriceOne,
				swapFee:        sdk.ZeroDec(),
			},
			expectedPoolId:      DefaultValidPoolID,
			expectedDenom0:      ETH,
			expectedDenom1:      USDC,
			expectedTickSpacing: DefaultTickSpacing,
		},
		{
			name: "Error: precisionValue greater than maximum",
			param: param{
				poolId:         DefaultValidPoolID,
				denom0:         ETH,
				denom1:         USDC,
				tickSpacing:    DefaultTickSpacing,
				precisionValue: types.ExponentAtPriceOneMax.Add(sdk.OneInt()),
				swapFee:        DefaultSwapFee,
			},
			expectedErr: types.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: types.ExponentAtPriceOneMax.Add(sdk.OneInt()), PrecisionValueAtPriceOneMin: types.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: types.ExponentAtPriceOneMax},
		},
		{
			name: "Error: precisionValue less than minimum",
			param: param{
				poolId:         DefaultValidPoolID,
				denom0:         ETH,
				denom1:         USDC,
				tickSpacing:    DefaultTickSpacing,
				precisionValue: types.ExponentAtPriceOneMin.Sub(sdk.OneInt()),
				swapFee:        DefaultSwapFee,
			},
			expectedErr: types.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: types.ExponentAtPriceOneMin.Sub(sdk.OneInt()), PrecisionValueAtPriceOneMin: types.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: types.ExponentAtPriceOneMax},
		},
		{
			name: "Error: same denom not allowed",
			param: param{
				poolId:         DefaultValidPoolID,
				denom0:         USDC,
				denom1:         USDC,
				tickSpacing:    DefaultTickSpacing,
				precisionValue: DefaultExponentAtPriceOne,
				swapFee:        DefaultSwapFee,
			},
			expectedErr: fmt.Errorf("cannot have the same asset in a single pool"),
		},
		{
			name: "Error: negative swap fee",
			param: param{
				poolId:         DefaultValidPoolID,
				denom0:         ETH,
				denom1:         USDC,
				tickSpacing:    DefaultTickSpacing,
				precisionValue: DefaultExponentAtPriceOne,
				swapFee:        sdk.ZeroDec().Sub(sdk.SmallestDec()),
			},
			expectedErr: types.InvalidSwapFeeError{ActualFee: sdk.ZeroDec().Sub(sdk.SmallestDec())},
		},
		{
			name: "Error: swap fee == 1",
			param: param{
				poolId:         DefaultValidPoolID,
				denom0:         ETH,
				denom1:         USDC,
				tickSpacing:    DefaultTickSpacing,
				precisionValue: DefaultExponentAtPriceOne,
				swapFee:        sdk.OneDec(),
			},
			expectedErr: types.InvalidSwapFeeError{ActualFee: sdk.OneDec()},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Call NewConcentratedLiquidityPool with the parameters from the current test.
			pool, err := model.NewConcentratedLiquidityPool(test.param.poolId, test.param.denom0, test.param.denom1, test.param.tickSpacing, test.param.precisionValue, test.param.swapFee)

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
