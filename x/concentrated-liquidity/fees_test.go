package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	concentratedliquidity "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	clmodel "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

var (
	oneEth = sdk.NewDecCoin(ETH, sdk.OneInt())
)

func (s *KeeperTestSuite) TestInitializeFeeAccumulatorPosition() {
	defaultPoolId := uint64(1)
	defaultLiquidityDelta := sdk.MustNewDecFromStr("10.0")
	type initFeeAccumTest struct {
		setPoolAccumulator  bool
		setExistingPosition bool
		expectedPass        bool
	}
	tests := map[string]initFeeAccumTest{
		"existing accumulator, new position": {
			setPoolAccumulator:  true,
			setExistingPosition: false,
			expectedPass:        true,
		},
		"existing accumulator, try overriding existing position": {
			setPoolAccumulator:  true,
			setExistingPosition: true,
			expectedPass:        true,
		},
		"error: non-existing accumulator": {
			setPoolAccumulator: false,
			expectedPass:       false,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper
			if tc.setPoolAccumulator {
				err := clKeeper.CreateFeeAccumulator(s.Ctx, defaultPoolId)
				s.Require().NoError(err)
			}
			if tc.setExistingPosition {
				// initialize with default liquidity delta * 2 to see if sut correctly initializes
				err := clKeeper.InitializeFeeAccumulatorPosition(s.Ctx, defaultPoolId, s.TestAccs[0], defaultLiquidityDelta.Add(defaultLiquidityDelta))
				s.Require().NoError(err)
			}

			// system under test
			err := clKeeper.InitializeFeeAccumulatorPosition(s.Ctx, defaultPoolId, s.TestAccs[0], defaultLiquidityDelta)
			if tc.expectedPass {
				s.Require().NoError(err)

				// get fee accum and see if position size has been properly initialized
				poolFeeAccumulator, err := clKeeper.GetFeeAccumulator(s.Ctx, defaultPoolId)
				s.Require().NoError(err)
				positionSize, err := poolFeeAccumulator.GetPositionSize(string(s.TestAccs[0].String()))
				s.Require().NoError(err)
				// position should have been properly initialzied to liquidityDelta provided
				s.Require().Equal(positionSize, defaultLiquidityDelta)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetFeeGrowthOutside() {
	type feeGrowthOutsideTest struct {
		poolSetup           bool
		tickSetup           bool
		feeAccumulatorSetup bool
		expectedError       bool
	}

	tests := map[string]feeGrowthOutsideTest{
		// TODO: uncomment this once tickInfo feeGrowthOutside logic has been implemented
		"happy path": {
			poolSetup:           true,
			tickSetup:           true,
			feeAccumulatorSetup: true,
			expectedError:       false,
		},
		// "tick has not been initialized": {
		// 	poolSetup:           true,
		// 	tickSetup:           false,
		// 	feeAccumulatorSetup: true,
		// 	expectedError:       false,
		// },
		"error: pool has not been setup": {
			poolSetup:           false,
			tickSetup:           false,
			feeAccumulatorSetup: false,
			expectedError:       true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			defaultPoolId := uint64(1)
			defaultLiquidityForTick := sdk.MustNewDecFromStr("10.0")
			defaultUpperTickIndex := int64(5)
			defaultLowerTickIndex := int64(3)

			// if pool set up true, set up default pool
			if tc.poolSetup {
				s.PrepareConcentratedPool()
			}

			// if tick set up true, set upper and lower ticks to default values
			if tc.tickSetup {
				// first initialize upper tick
				err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(
					s.Ctx,
					defaultPoolId,
					defaultUpperTickIndex,
					defaultLiquidityForTick,
					true,
				)
				s.Require().NoError(err)

				// initialize lower tick
				err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(
					s.Ctx,
					defaultPoolId,
					defaultLowerTickIndex,
					defaultLiquidityForTick,
					true,
				)
				s.Require().NoError(err)
			}

			// system under test
			feeGrowthOutside, err := s.App.ConcentratedLiquidityKeeper.GetFeeGrowthOutside(s.Ctx, defaultPoolId, defaultLowerTickIndex, defaultUpperTickIndex)
			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// check if returned fee growth outside has correct value
				s.Require().Equal(feeGrowthOutside, sdk.DecCoins(nil))
			}
		})
	}
}

func (s *KeeperTestSuite) TestCalculateFeeGrowth() {
	defaultGeeFrowthGlobal := sdk.NewDecCoins(sdk.NewDecCoin("uosmo", sdk.NewInt(10)))
	defaultGeeFrowthOutside := sdk.NewDecCoins(sdk.NewDecCoin("uosmo", sdk.NewInt(3)))

	defaultSmallerTargetTick := int64(1)
	defaultCurrentTick := int64(2)
	defaultLargerTargetTick := int64(3)

	type calcFeeGrowthTest struct {
		isUpperTick                bool
		isCurrentTickGTETargetTick bool
		expectedFeeGrowth          sdk.DecCoins
	}

	tests := map[string]calcFeeGrowthTest{
		"current Tick is greater than the upper tick": {
			isUpperTick:                true,
			isCurrentTickGTETargetTick: false,
			expectedFeeGrowth:          defaultGeeFrowthOutside,
		},
		"current Tick is less than the upper tick": {
			isUpperTick:                true,
			isCurrentTickGTETargetTick: true,
			expectedFeeGrowth:          defaultGeeFrowthGlobal.Sub(defaultGeeFrowthOutside),
		},
		"current Tick is less than the lower tick": {
			isUpperTick:                false,
			isCurrentTickGTETargetTick: false,
			expectedFeeGrowth:          defaultGeeFrowthGlobal.Sub(defaultGeeFrowthOutside),
		},
		"current Tick is greater than the lower tick": {
			isUpperTick:                false,
			isCurrentTickGTETargetTick: true,
			expectedFeeGrowth:          defaultGeeFrowthOutside,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			var targetTick int64
			if tc.isCurrentTickGTETargetTick {
				targetTick = defaultSmallerTargetTick
			} else {
				targetTick = defaultLargerTargetTick
			}
			feeGrowth := concentratedliquidity.CalculateFeeGrowth(
				targetTick,
				defaultGeeFrowthOutside,
				defaultCurrentTick,
				defaultGeeFrowthGlobal,
				tc.isUpperTick,
			)
			s.Require().Equal(feeGrowth, tc.expectedFeeGrowth)
		})
	}

}

func (suite *KeeperTestSuite) TestGetInitialFeeGrowthOutsideForTick() {
	const (
		validPoolId = 1
	)

	var (
		initialPoolTick = math.PriceToTick(DefaultAmt1.ToDec().Quo(DefaultAmt0.ToDec())).Int64()
	)

	tests := map[string]struct {
		poolId                   uint64
		tick                     int64
		initialGlobalFeeGrowth   sdk.DecCoin
		shouldAvoidCreatingAccum bool

		expectedInitialFeeGrowthOutside sdk.DecCoins
		expectError                     error
	}{
		"current tick > tick -> fee growth global": {
			poolId:                 validPoolId,
			tick:                   initialPoolTick - 1,
			initialGlobalFeeGrowth: oneEth,

			expectedInitialFeeGrowthOutside: sdk.NewDecCoins(oneEth),
		},
		"current tick == tick -> fee growth global": {
			poolId:                 validPoolId,
			tick:                   initialPoolTick,
			initialGlobalFeeGrowth: oneEth,

			expectedInitialFeeGrowthOutside: sdk.NewDecCoins(oneEth),
		},
		"current tick < tick -> empty coins": {
			poolId:                 validPoolId,
			tick:                   initialPoolTick + 1,
			initialGlobalFeeGrowth: oneEth,

			expectedInitialFeeGrowthOutside: concentratedliquidity.EmptyCoins,
		},
		"pool does not exist": {
			poolId:                 validPoolId + 1,
			tick:                   initialPoolTick - 1,
			initialGlobalFeeGrowth: oneEth,

			expectError: types.PoolNotFoundError{PoolId: validPoolId + 1},
		},
		"accumulator does not exist": {
			poolId:                   validPoolId,
			tick:                     0,
			initialGlobalFeeGrowth:   oneEth,
			shouldAvoidCreatingAccum: true,

			expectError: accum.AccumDoesNotExistError{AccumName: concentratedliquidity.GetFeeAccumulatorName(validPoolId)},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			ctx := suite.Ctx
			clKeeper := suite.App.ConcentratedLiquidityKeeper

			pool, err := clmodel.NewConcentratedLiquidityPool(validPoolId, USDC, ETH, DefaultTickSpacing)
			suite.Require().NoError(err)

			err = clKeeper.SetPool(ctx, &pool)
			suite.Require().NoError(err)

			if !tc.shouldAvoidCreatingAccum {
				err = clKeeper.CreateFeeAccumulator(ctx, validPoolId)
				suite.Require().NoError(err)

				// Setup test position to make sure that tick is initialized
				suite.SetupPosition(validPoolId)

				err = clKeeper.ChargeFee(ctx, validPoolId, tc.initialGlobalFeeGrowth)
				suite.Require().NoError(err)
			}

			// System under test.
			initialFeeGrowthOutside, err := clKeeper.GetInitialFeeGrowthOutsideForTick(ctx, tc.poolId, tc.tick)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectError)
				return
			}
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedInitialFeeGrowthOutside, initialFeeGrowthOutside)
		})
	}
}

func (suite *KeeperTestSuite) TestChargeFee() {
	// setup once at the beginning.
	suite.SetupTest()

	ctx := suite.Ctx
	clKeeper := suite.App.ConcentratedLiquidityKeeper

	// create fee accumulators with ids 1 and 2 but not 3.
	err := clKeeper.CreateFeeAccumulator(ctx, 1)
	suite.Require().NoError(err)
	err = clKeeper.CreateFeeAccumulator(ctx, 2)
	suite.Require().NoError(err)

	tests := map[string]struct {
		poolId    uint64
		feeUpdate sdk.DecCoin

		expectedGlobalGrowth sdk.DecCoins
		expectError          error
	}{
		"pool id 1 - one eth": {
			poolId:    1,
			feeUpdate: oneEth,

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth),
		},
		"pool id 1 - 2 usdc": {
			poolId:    1,
			feeUpdate: sdk.NewDecCoin(USDC, sdk.NewInt(2)),

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth).Add(sdk.NewDecCoin(USDC, sdk.NewInt(2))),
		},
		"pool id 2 - 1 usdc": {
			poolId:    2,
			feeUpdate: oneEth,

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth),
		},
		"accumulator does not exist": {
			poolId:    3,
			feeUpdate: oneEth,

			expectError: accum.AccumDoesNotExistError{AccumName: concentratedliquidity.GetFeeAccumulatorName(3)},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			// System under test.
			err := clKeeper.ChargeFee(ctx, tc.poolId, tc.feeUpdate)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectError)
				return
			}
			suite.Require().NoError(err)

			feeAcumulator, err := clKeeper.GetFeeAccumulator(ctx, tc.poolId)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedGlobalGrowth, feeAcumulator.GetValue())
		})
	}
}

func (s *KeeperTestSuite) TestCollectFees() {
	tests := map[string]struct {
		// setup parameters.
		initialLiquidity          sdk.Dec
		lowerTickFeeGrowthOutside sdk.DecCoins
		upperTickFeeGrowthOutside sdk.DecCoins
		globalFeeGrowth           sdk.DecCoins
		currentTick               int64
		isInvalidPoolIdGiven      bool

		// inputs parameters.
		owner     sdk.AccAddress
		lowerTick int64
		upperTick int64

		// expectations.
		expectedFeesClaimed sdk.Coins
		expectedError       error
	}{
		// imagine single swap over entire position
		// crossing left > right and stopping above upper tick
		// In this case, only the upper tick accumulator must have
		// been updated when crossed.
		// Since we track fees accrued below a tick, upper tick is updated
		// while lower tick is zero.
		"single swap left -> right: 2 ticks, one share - current price > upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:     s.TestAccs[0],
			lowerTick: 0,
			upperTick: 1,

			currentTick: 2,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares - current price > upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:     s.TestAccs[0],
			lowerTick: 0,
			upperTick: 2,

			currentTick: 3,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(20))),
		},
		"single swap left -> right: 2 ticks, one share - current price == upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:     s.TestAccs[0],
			lowerTick: 0,
			upperTick: 1,

			currentTick: 1,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},

		// imagine single swap over entire position
		// crossing right > left and stopping at lower tick
		// In this case, all fees must have been accrued inside the tick
		// Since we track fees accrued below a tick, both upper and lower position
		// ticks are zero
		"single swap right -> left: 2 ticks, one share - current price == lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator is not updated yet.
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:     s.TestAccs[0],
			lowerTick: 0,
			upperTick: 1,

			currentTick: 0,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share - current price < lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator updated when crossed.
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:     s.TestAccs[0],
			lowerTick: 0,
			upperTick: 1,

			currentTick: -1,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},

		"invalid pool id given": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:     s.TestAccs[0],
			lowerTick: 0,
			upperTick: 1,

			currentTick: 2,

			isInvalidPoolIdGiven: true,
			expectedError:        cltypes.PoolNotFoundError{PoolId: 2},

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			validPool := s.PrepareConcentratedPool()
			validPoolId := validPool.GetId()

			s.FundAcc(validPool.GetAddress(), tc.expectedFeesClaimed)

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx

			err := clKeeper.InitializeFeeAccumulatorPosition(ctx, validPoolId, tc.owner, tc.initialLiquidity)
			s.Require().NoError(err)

			s.initializeTick(ctx, tc.lowerTick, tc.initialLiquidity, tc.lowerTickFeeGrowthOutside, false)

			s.initializeTick(ctx, tc.upperTick, tc.initialLiquidity, tc.upperTickFeeGrowthOutside, true)

			validPool.SetCurrentTick(sdk.NewInt(tc.currentTick))
			clKeeper.SetPool(ctx, validPool)

			err = clKeeper.ChargeFee(ctx, validPoolId, tc.globalFeeGrowth[0])
			s.Require().NoError(err)

			poolBalanceBeforeCollect := s.App.BankKeeper.GetBalance(ctx, validPool.GetAddress(), ETH)
			ownerBalancerBeforeCollect := s.App.BankKeeper.GetBalance(ctx, tc.owner, ETH)

			sutPoolId := validPoolId
			if tc.isInvalidPoolIdGiven {
				sutPoolId = sutPoolId + 1
			}

			// System under test
			actualFeesClaimed, err := clKeeper.CollectFees(ctx, sutPoolId, tc.owner, tc.lowerTick, tc.upperTick)

			// Assertions.

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectedError)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedFeesClaimed, actualFeesClaimed)

			poolBalanceAfterCollect := s.App.BankKeeper.GetBalance(ctx, validPool.GetAddress(), ETH)
			ownerBalancerAfterCollect := s.App.BankKeeper.GetBalance(ctx, tc.owner, ETH)

			expectedETHAmount := tc.expectedFeesClaimed.AmountOf(ETH)
			s.Require().Equal(expectedETHAmount, poolBalanceBeforeCollect.Sub(poolBalanceAfterCollect).Amount)
			s.Require().Equal(expectedETHAmount, ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect).Amount)
		})
	}
}

func (s *KeeperTestSuite) initializeTick(ctx sdk.Context, tickIndex int64, initialLiquidity sdk.Dec, feeGrowthOutside sdk.DecCoins, isLower bool) {
	err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(ctx, validPoolId, tickIndex, initialLiquidity, isLower)
	s.Require().NoError(err)

	tickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, validPoolId, tickIndex)
	s.Require().NoError(err)

	tickInfo.FeeGrowthOutside = feeGrowthOutside

	s.App.ConcentratedLiquidityKeeper.SetTickInfo(ctx, validPoolId, tickIndex, tickInfo)
}
