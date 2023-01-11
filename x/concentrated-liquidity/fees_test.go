package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	concentratedliquidity "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/internal/math"
	clmodel "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

// fields used to identify a fee position.
type positionIdentifiers struct {
	poolId    uint64
	owner     sdk.AccAddress
	lowerTick int64
	upperTick int64
}

var (
	oneEth = sdk.NewDecCoin(ETH, sdk.OneInt())
)

func (s *KeeperTestSuite) TestInitializeFeeAccumulatorPosition() {
	// Setup is done once so that we test
	// the relationship between test cases.
	// For example, that positions with non-zero liquidity
	// cannot be overriden.
	s.SetupTest()

	var (
		defaultPoolId         = uint64(1)
		defaultLiquidityDelta = sdk.MustNewDecFromStr("10.0")
		defaultPositionId     = positionIdentifiers{
			defaultPoolId,
			s.TestAccs[0],
			DefaultLowerTick,
			DefaultUpperTick,
		}
	)

	withOwner := func(posId positionIdentifiers, owner sdk.AccAddress) positionIdentifiers {
		posId.owner = owner
		return posId
	}

	withUpperTick := func(posId positionIdentifiers, upperTick int64) positionIdentifiers {
		posId.upperTick = upperTick
		return posId
	}

	withLowerTick := func(posId positionIdentifiers, lowerTick int64) positionIdentifiers {
		posId.lowerTick = lowerTick
		return posId
	}

	clKeeper := s.App.ConcentratedLiquidityKeeper

	err := clKeeper.CreateFeeAccumulator(s.Ctx, defaultPoolId)
	s.Require().NoError(err)

	type initFeeAccumTest struct {
		name           string
		liquidityDelta sdk.Dec
		positionId     positionIdentifiers

		expectedPass bool
	}
	tests := []initFeeAccumTest{
		{
			name:           "new zero liqudidity position",
			liquidityDelta: sdk.ZeroDec(),
			// Gets overwritten by the next test case
			positionId:   defaultPositionId,
			expectedPass: true,
		},
		{
			name: "new non-zero liqudidity position",
			// Does not get overwritten by the next test case
			// because contains non-zero liquidity.
			liquidityDelta: defaultLiquidityDelta,
			positionId:     defaultPositionId,
			expectedPass:   true,
		},
		{
			name:           "overriding existing position with non-zero liqudity - error",
			liquidityDelta: defaultLiquidityDelta.Add(sdk.OneDec()),
			positionId:     defaultPositionId,
			expectedPass:   false,
		},
		{
			name:           "error: non-existing accumulator",
			liquidityDelta: defaultLiquidityDelta,
			positionId: positionIdentifiers{
				defaultPoolId + 1, // non-existing pool
				s.TestAccs[0],
				DefaultLowerTick,
				DefaultUpperTick,
			},
			expectedPass: false,
		},
		{
			name:           "existing accumulator, different owner - different position",
			liquidityDelta: defaultLiquidityDelta,
			positionId:     withOwner(defaultPositionId, s.TestAccs[1]),
			expectedPass:   true,
		},
		{
			name:           "existing accumulator, different upper tick - different position",
			liquidityDelta: defaultLiquidityDelta,
			positionId:     withUpperTick(defaultPositionId, DefaultUpperTick+1),
			expectedPass:   true,
		},
		{
			name:           "existing accumulator, different lower tick - different position",
			liquidityDelta: defaultLiquidityDelta,
			positionId:     withLowerTick(defaultPositionId, DefaultUpperTick+1),
			expectedPass:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// system under test
			err := clKeeper.InitializeFeeAccumulatorPosition(s.Ctx, tc.positionId.poolId, tc.positionId.owner, tc.positionId.lowerTick, tc.positionId.upperTick, tc.liquidityDelta)
			if tc.expectedPass {
				s.Require().NoError(err)

				// get fee accum and see if position size has been properly initialized
				poolFeeAccumulator, err := clKeeper.GetFeeAccumulator(s.Ctx, defaultPoolId)
				s.Require().NoError(err)

				positionKey := concentratedliquidity.FormatPositionAccumulatorKey(tc.positionId.poolId, tc.positionId.owner, tc.positionId.lowerTick, tc.positionId.upperTick)

				positionSize, err := poolFeeAccumulator.GetPositionSize(positionKey)
				s.Require().NoError(err)
				// position should have been properly initialzied to liquidityDelta provided
				s.Require().Equal(positionSize, tc.liquidityDelta)
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
