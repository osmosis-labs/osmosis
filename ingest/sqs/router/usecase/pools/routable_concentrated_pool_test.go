package pools_test

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/app/apptesting"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/pools"
	concentratedmodel "github.com/osmosis-labs/osmosis/v21/x/concentrated-liquidity/model"
)

func deepCopyTickModel(tickModel *domain.TickModel) *domain.TickModel {
	ticks := make([]domain.LiquidityDepthsWithRange, len(tickModel.Ticks))
	copy(ticks, tickModel.Ticks)
	return &domain.TickModel{
		Ticks:            ticks,
		CurrentTickIndex: tickModel.CurrentTickIndex,
		HasNoLiquidity:   tickModel.HasNoLiquidity,
	}
}

func withHasNoLiquidity(tickModel *domain.TickModel) *domain.TickModel {
	tickModel = deepCopyTickModel(tickModel)
	tickModel.HasNoLiquidity = true
	return tickModel
}

func withCurrentTickIndex(tickModel *domain.TickModel, currentTickIndex int64) *domain.TickModel {
	tickModel = deepCopyTickModel(tickModel)
	tickModel.CurrentTickIndex = currentTickIndex
	return tickModel
}

func withTicks(tickModel *domain.TickModel, ticks []domain.LiquidityDepthsWithRange) *domain.TickModel {
	tickModel = deepCopyTickModel(tickModel)
	tickModel.Ticks = ticks
	return tickModel
}

// Tests the CalculateTokenOutByTokenIn method of the RoutableConcentratedPoolImpl struct
// when the pool is concentrated.
//
// It uses the same success test cases as the chain logic.
// The error cases are tested in a separate fixture because the edge cases are different..
func (s *RoutablePoolTestSuite) TestCalculateTokenOutByTokenIn_Concentrated_SuccessChainVectors() {
	tests := apptesting.SwapOutGivenInCases

	for name, tc := range tests {
		s.Run(name, func() {
			// Note: router quote tests do not have the concept of slippage protection.
			// These quotes are used to derive the slippage protection amount.
			// So we skip these tests.
			if strings.Contains(name, "slippage protection") {
				s.T().Skip("no slippage protection in router quote tests")
			}

			s.SetupAndFundSwapTest()
			concentratedPool := s.PreparePoolWithCustSpread(tc.SpreadFactor)
			// add default position
			s.SetupDefaultPosition(concentratedPool.GetId())
			s.SetupSecondPosition(tc, concentratedPool)

			// Refetch the pool
			concentratedPool, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, concentratedPool.GetId())
			s.Require().NoError(err)

			// Get liquidity for full range
			ticks, currentTickIndex, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityForFullRange(s.Ctx, concentratedPool.GetId())
			s.Require().NoError(err)

			poolWrapper := &domain.PoolWrapper{
				ChainModel: concentratedPool,
				TickModel: &domain.TickModel{
					Ticks:            ticks,
					CurrentTickIndex: currentTickIndex,
					HasNoLiquidity:   false,
				},
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC:  osmomath.NewInt(100),
					TotalValueLockedError: "",
					Balances:              sdk.Coins{},
					PoolDenoms:            []string{"foo", "bar"},
				},
			}
			routablePool := pools.NewRoutablePool(poolWrapper, tc.TokenOutDenom, noTakerFee)

			tokenOut, err := routablePool.CalculateTokenOutByTokenIn(tc.TokenIn)

			s.Require().NoError(err)
			s.Require().Equal(tc.ExpectedTokenOut.String(), tokenOut.String())
		})
	}
}

// This test cases focuses on testing error and edge cases for CL quote calculation out by token in.
func (s *RoutablePoolTestSuite) TestCalculateTokenOutByTokenIn_Concentrated_ErrorAndEdgeCases() {
	const (
		defaultCurrentTick = int64(0)
	)

	tests := map[string]struct {
		tokenIn       sdk.Coin
		tokenOutDenom string

		tickModelOverwrite          *domain.TickModel
		isTickModelNil              bool
		shouldCreateDefaultPosition bool

		expectedTokenOut sdk.Coin
		expectError      error
	}{
		"error: failed to get tick model": {
			tokenIn:       DefaultCoin1,
			tokenOutDenom: Denom0,

			isTickModelNil: true,

			expectError: domain.ConcentratedPoolNoTickModelError{
				PoolId: defaultPoolID,
			},
		},
		"error: current bucket index is negative": {
			tokenIn:       DefaultCoin1,
			tokenOutDenom: Denom0,

			tickModelOverwrite: withCurrentTickIndex(defaultTickModel, -1),

			expectError: domain.ConcentratedCurrentTickNotWithinBucketError{
				PoolId:             defaultPoolID,
				CurrentBucketIndex: -1,
				TotalBuckets:       0,
			},
		},
		"error: current bucket index is greater than or equal to total buckets": {
			tokenIn:       DefaultCoin1,
			tokenOutDenom: Denom0,

			tickModelOverwrite: defaultTickModel,

			expectError: domain.ConcentratedCurrentTickNotWithinBucketError{
				PoolId:             defaultPoolID,
				CurrentBucketIndex: defaultCurrentTick,
				TotalBuckets:       defaultCurrentTick,
			},
		},
		"error: has no liquidity": {
			tokenIn:       DefaultCoin1,
			tokenOutDenom: Denom0,

			tickModelOverwrite: withHasNoLiquidity(defaultTickModel),

			expectError: domain.ConcentratedNoLiquidityError{
				PoolId: defaultPoolID,
			},
		},
		"error: current tick is not within current bucket": {
			tokenIn:       DefaultCoin1,
			tokenOutDenom: Denom0,

			tickModelOverwrite: withTicks(defaultTickModel, []domain.LiquidityDepthsWithRange{
				{
					LowerTick:       defaultCurrentTick - 2,
					UpperTick:       defaultCurrentTick - 1,
					LiquidityAmount: DefaultLiquidityAmt,
				},
			}),

			expectError: domain.ConcentratedCurrentTickAndBucketMismatchError{
				CurrentTick: defaultCurrentTick,
				LowerTick:   defaultCurrentTick - 2,
				UpperTick:   defaultCurrentTick - 1,
			},
		},
		"error: zero current sqrt price": {
			tokenIn:       DefaultCoin1,
			tokenOutDenom: Denom0,

			tickModelOverwrite: &domain.TickModel{
				Ticks: []domain.LiquidityDepthsWithRange{
					{
						LowerTick:       defaultCurrentTick,
						UpperTick:       defaultCurrentTick + 1,
						LiquidityAmount: DefaultLiquidityAmt,
					},
				},
				CurrentTickIndex: defaultCurrentTick,

				// Note that despite setting HasNoLiquidity to false,
				// the pool is in invalid state. We expect that the ingester
				// will not allow this to happen.
				HasNoLiquidity: false,
			},

			expectError: domain.ConcentratedZeroCurrentSqrtPriceError{PoolId: defaultPoolID},
		},
		"error: not enough liquidity to complete swap": {
			tokenIn:       DefaultCoin1,
			tokenOutDenom: Denom0,

			shouldCreateDefaultPosition: true,

			tickModelOverwrite: withTicks(defaultTickModel, []domain.LiquidityDepthsWithRange{
				{
					LowerTick:       DefaultCurrentTick,
					UpperTick:       DefaultCurrentTick + 1,
					LiquidityAmount: DefaultLiquidityAmt,
				},
			}),

			expectError: domain.ConcentratedNotEnoughLiquidityToCompleteSwapError{
				PoolId:   defaultPoolID,
				AmountIn: DefaultCoin1.String(),
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			var (
				tickModel *domain.TickModel
				err       error
			)

			pool := s.PrepareConcentratedPool()
			concentratedPool, ok := pool.(*concentratedmodel.Pool)
			s.Require().True(ok)

			if tc.shouldCreateDefaultPosition {
				s.SetupDefaultPosition(concentratedPool.Id)
			}

			// refetch the pool
			pool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, concentratedPool.Id)
			s.Require().NoError(err)
			concentratedPool, ok = pool.(*concentratedmodel.Pool)
			s.Require().True(ok)

			if tc.tickModelOverwrite != nil {
				tickModel = tc.tickModelOverwrite

			} else if tc.isTickModelNil {
				// For clarity:
				tickModel = nil
			} else {
				// Get liquidity for full range
				ticks, currentTickIndex, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityForFullRange(s.Ctx, concentratedPool.Id)
				s.Require().NoError(err)

				tickModel = &domain.TickModel{
					Ticks:            ticks,
					CurrentTickIndex: currentTickIndex,
					HasNoLiquidity:   false,
				}
			}

			routablePool := pools.RoutableConcentratedPoolImpl{
				ChainPool:     concentratedPool,
				TickModel:     tickModel,
				TokenOutDenom: tc.tokenOutDenom,
				TakerFee:      osmomath.ZeroDec(),
			}

			tokenOut, err := routablePool.CalculateTokenOutByTokenIn(tc.tokenIn)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			s.Require().Equal(tc.expectedTokenOut.String(), tokenOut.String())
		})
	}
}
