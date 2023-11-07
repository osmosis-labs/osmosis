package usecase_test

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/app/apptesting"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	routerusecase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

var (
	// Concentrated liquidity constants
	ETH    = apptesting.ETH
	USDC   = apptesting.USDC
	Denom0 = ETH
	Denom1 = USDC

	DefaultCurrentTick = apptesting.DefaultCurrTick

	DefaultAmt0 = apptesting.DefaultAmt0
	DefaultAmt1 = apptesting.DefaultAmt1

	DefaultCoin0 = apptesting.DefaultCoin0
	DefaultCoin1 = apptesting.DefaultCoin1

	DefaultLiquidityAmt = apptesting.DefaultLiquidityAmt

	// router specific variables
	defaultTickModel = &domain.TickModel{
		Ticks:            []domain.LiquidityDepthsWithRange{},
		CurrentTickIndex: 0,
		HasNoLiquidity:   false,
	}

	noTakerFee = osmomath.ZeroDec()
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
func (s *RouterTestSuite) TestCalculateTokenOutByTokenIn_Concentrated_SuccessChainVectors() {
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
			routablePool := routerusecase.NewRoutablePool(poolWrapper, tc.TokenOutDenom, noTakerFee)

			err = routablePool.Validate(osmomath.NewInt(100))
			s.Require().NoError(err)

			tokenOut, err := routablePool.CalculateTokenOutByTokenIn(tc.TokenIn)

			s.Require().NoError(err)
			s.Require().Equal(tc.ExpectedTokenOut.String(), tokenOut.String())
		})
	}
}

// This test cases focuses on testing error and edge cases for CL quote calculation out by token in.
func (s *RouterTestSuite) TestCalculateTokenOutByTokenIn_Concentrated_ErrorAndEdgeCases() {
	const (
		defaultCurrentTick = int64(0)
	)

	var (
		defaultSQSModel = domain.SQSPool{
			TotalValueLockedUSDC:  osmomath.NewInt(100),
			TotalValueLockedError: "",
			Balances:              sdk.Coins{},
			PoolDenoms:            []string{Denom0, Denom1},
		}
	)

	tests := map[string]struct {
		tokenIn       sdk.Coin
		tokenOutDenom string

		isWrongChainModel           bool
		tickModelOverwrite          *domain.TickModel
		isTickModelNil              bool
		shouldCreateDefaultPosition bool

		expectedTokenOut sdk.Coin
		expectError      error
	}{
		"error: chain model is not concentrated": {
			tokenIn:       DefaultCoin1,
			tokenOutDenom: Denom0,

			isWrongChainModel: true,

			expectError: domain.InvalidPoolTypeError{PoolType: int32(poolmanagertypes.Balancer)},
		},
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

			expectError: usecase.ConcentratedCurrentTickNotWithinBucketError{
				PoolId:             defaultPoolID,
				CurrentBucketIndex: -1,
				TotalBuckets:       0,
			},
		},
		"error: current bucket index is greater than or equal to total buckets": {
			tokenIn:       DefaultCoin1,
			tokenOutDenom: Denom0,

			tickModelOverwrite: defaultTickModel,

			expectError: usecase.ConcentratedCurrentTickNotWithinBucketError{
				PoolId:             defaultPoolID,
				CurrentBucketIndex: defaultCurrentTick,
				TotalBuckets:       defaultCurrentTick,
			},
		},
		"error: has no liquidity": {
			tokenIn:       DefaultCoin1,
			tokenOutDenom: Denom0,

			tickModelOverwrite: withHasNoLiquidity(defaultTickModel),

			expectError: usecase.ConcentratedNoLiquidityError{
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

			expectError: usecase.ConcentratedCurrentTickAndBucketMismatchError{
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

			expectError: usecase.ConcentratedZeroCurrentSqrtPriceError{PoolId: defaultPoolID},
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

			expectError: usecase.ConcentratedNotEnoughLiquidityToCompleteSwapError{
				PoolId:   defaultPoolID,
				AmountIn: DefaultCoin1.String(),
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			var (
				chainModel poolmanagertypes.PoolI
				tickModel  *domain.TickModel
				err        error
			)

			if tc.isWrongChainModel {
				balancerPoolID := s.PrepareBalancerPool()
				balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, balancerPoolID)
				s.Require().NoError(err)

				chainModel = balancerPool
			} else {
				chainModel = s.PrepareConcentratedPool()

				if tc.shouldCreateDefaultPosition {
					s.SetupDefaultPosition(chainModel.GetId())
				}

				// refetch the pool
				chainModel, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, chainModel.GetId())
				s.Require().NoError(err)

				if tc.tickModelOverwrite != nil {
					tickModel = tc.tickModelOverwrite

				} else if tc.isTickModelNil {
					// For clarity:
					tickModel = nil
				} else {
					// Get liquidity for full range
					ticks, currentTickIndex, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityForFullRange(s.Ctx, chainModel.GetId())
					s.Require().NoError(err)

					tickModel = &domain.TickModel{
						Ticks:            ticks,
						CurrentTickIndex: currentTickIndex,
						HasNoLiquidity:   false,
					}
				}
			}

			routablePool := routerusecase.RoutableConcentratedPoolImpl{
				PoolI: &domain.PoolWrapper{
					ChainModel: chainModel,
					TickModel:  tickModel,
					SQSModel:   defaultSQSModel,
				},
				TokenOutDenom: tc.tokenOutDenom,
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
