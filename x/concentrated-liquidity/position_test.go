package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

var _ = suite.TestingSuite(nil)

func (s *KeeperTestSuite) TestInitOrUpdatePosition() {
	const validPoolId = 1
	type param struct {
		poolId         uint64
		lowerTick      int64
		upperTick      int64
		liquidityDelta sdk.Dec
		liquidityIn    sdk.Dec
	}

	tests := []struct {
		name              string
		param             param
		positionExists    bool
		expectedLiquidity sdk.Dec
		expectedErr       string
	}{
		{
			name: "Init position from -50 to 50 with DefaultLiquidityAmt liquidity",
			param: param{
				poolId:         validPoolId,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt,
			},
			positionExists:    false,
			expectedLiquidity: DefaultLiquidityAmt,
		},
		{
			name: "Update position from -50 to 50 that already contains DefaultLiquidityAmt liquidity with DefaultLiquidityAmt more liquidity",
			param: param{
				poolId:         validPoolId,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt,
			},
			positionExists:    true,
			expectedLiquidity: DefaultLiquidityAmt.Mul(sdk.NewDec(2)),
		},
		{
			name: "Init position for non-existing pool",
			param: param{
				poolId:         2,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt,
			},
			positionExists: false,
			expectedErr:    "cannot retrieve position for a non-existent pool",
		},
		{
			name: "Init position from -50 to 50 with negative DefaultLiquidityAmt liquidity",
			param: param{
				poolId:         validPoolId,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt.Neg(),
			},
			positionExists: false,
			expectedErr:    "liquidity cannot be negative",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a CL pool with poolId 1
			_, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, ETH, USDC, DefaultCurrSqrtPrice, sdk.NewInt(DefaultCurrTick))
			s.Require().NoError(err)

			// If positionExists set, initialize the specified position with defaultLiquidityAmt
			preexistingLiquidity := sdk.ZeroDec()
			if test.positionExists {
				err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, test.param.poolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick, test.param.liquidityDelta)
				s.Require().NoError(err)
				preexistingLiquidity = DefaultLiquidityAmt
			}

			// Get the position info for poolId 1
			positionInfo, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, 1, s.TestAccs[0], test.param.lowerTick, test.param.upperTick)
			if test.positionExists {
				// If we had a position before, ensure the position info displays proper liquidity
				s.Require().NoError(err)
				s.Require().Equal(preexistingLiquidity, positionInfo.Liquidity)
			} else {
				// If we did not have a position before, ensure getting the non-existent position returns an error
				s.Require().Error(err)
				s.Require().ErrorContains(err, "position not found")
			}

			// Initialize or update the position according to the test case
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, test.param.poolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick, test.param.liquidityDelta)
			if test.expectedErr != "" {
				s.Require().ErrorContains(err, test.expectedErr)
				return
			}
			s.Require().NoError(err)

			// Get the tick info for poolId 1
			positionInfo, err = s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, 1, s.TestAccs[0], test.param.lowerTick, test.param.upperTick)
			s.Require().NoError(err)

			// Check that the initialized or updated position matches our expectation
			s.Require().Equal(test.expectedLiquidity, positionInfo.Liquidity)
		})
	}
}
