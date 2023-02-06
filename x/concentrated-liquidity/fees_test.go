package concentrated_liquidity_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	cl "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/internal/math"
	clmodel "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

// fields used to identify a fee position.
type positionIdentifiers struct {
	poolId      uint64
	owner       sdk.AccAddress
	lowerTick   int64
	upperTick   int64
	frozenUntil time.Time
}

var (
	oneEth      = sdk.NewDecCoin(ETH, sdk.OneInt())
	oneEthCoins = sdk.NewDecCoins(oneEth)
)

func (s *KeeperTestSuite) TestInitializeFeeAccumulatorPosition() {
	// Setup is done once so that we test
	// the relationship between test cases.
	// For example, that positions with non-zero liquidity
	// cannot be overriden.
	s.SetupTest()
	defaultFrozenUntil := s.Ctx.BlockTime().Add(DefaultFreezeDuration)
	defaultAccount := s.TestAccs[0]

	var (
		defaultPoolId     = uint64(1)
		defaultPositionId = positionIdentifiers{
			defaultPoolId,
			defaultAccount,
			DefaultLowerTick,
			DefaultUpperTick,
			defaultFrozenUntil,
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
		name       string
		positionId positionIdentifiers

		expectedPass bool
	}
	tests := []initFeeAccumTest{
		{
			name:         "first position",
			positionId:   defaultPositionId,
			expectedPass: true,
		},
		{
			name:         "second position",
			positionId:   withLowerTick(defaultPositionId, DefaultLowerTick+1),
			expectedPass: true,
		},
		{
			name:       "overriding first position - error",
			positionId: defaultPositionId,
			// Does not get overwritten by the next test case.
			expectedPass: false,
		},
		{
			name:       "overriding second position - error",
			positionId: withLowerTick(defaultPositionId, DefaultLowerTick+1),
			// Does not get overwritten by the next test case.
			expectedPass: false,
		},
		{
			name: "error: non-existing accumulator (wrong pool)",
			positionId: positionIdentifiers{
				defaultPoolId + 1, // non-existing pool
				defaultAccount,
				DefaultLowerTick,
				DefaultUpperTick,
				defaultFrozenUntil,
			},
			expectedPass: false,
		},
		{
			name:         "existing accumulator, different owner - different position",
			positionId:   withOwner(defaultPositionId, s.TestAccs[1]),
			expectedPass: true,
		},
		{
			name:         "existing accumulator, different upper tick - different position",
			positionId:   withUpperTick(defaultPositionId, DefaultUpperTick+1),
			expectedPass: true,
		},
		{
			name:         "existing accumulator, different lower tick - different position",
			positionId:   withLowerTick(defaultPositionId, DefaultUpperTick+1),
			expectedPass: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// system under test
			err := clKeeper.InitializeFeeAccumulatorPosition(s.Ctx, tc.positionId.poolId, tc.positionId.owner, tc.positionId.lowerTick, tc.positionId.upperTick)
			if tc.expectedPass {
				s.Require().NoError(err)

				// get fee accum and see if position size has been properly initialized
				poolFeeAccumulator, err := clKeeper.GetFeeAccumulator(s.Ctx, defaultPoolId)
				s.Require().NoError(err)

				positionKey := cl.FormatPositionAccumulatorKey(tc.positionId.poolId, tc.positionId.owner, tc.positionId.lowerTick, tc.positionId.upperTick)

				positionSize, err := poolFeeAccumulator.GetPositionSize(positionKey)
				s.Require().NoError(err)
				// position should have been properly initialized to zero
				s.Require().Equal(positionSize, sdk.ZeroDec())
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetFeeGrowthOutside() {
	type feeGrowthOutsideTest struct {
		poolSetup bool

		lowerTick                 int64
		upperTick                 int64
		currentTick               int64
		lowerTickFeeGrowthOutside sdk.DecCoins
		upperTickFeeGrowthOutside sdk.DecCoins
		globalFeeGrowth           sdk.DecCoin

		expectedFeeGrowthOutside sdk.DecCoins
		invalidTick              bool
		expectedError            bool
	}

	defaultAccumCoins := sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10)))
	defaultPoolId := uint64(1)
	defaultInitialLiquidity := sdk.OneDec()

	defaultUpperTickIndex := int64(5)
	defaultLowerTickIndex := int64(3)

	tests := map[string]feeGrowthOutsideTest{
		// imagine single swap over entire position
		// crossing left > right and stopping above upper tick
		// In this case, only the upper tick accumulator must have
		// been updated when crossed.
		// Since we track fees accrued below a tick, upper tick is updated
		// while lower tick is zero.
		"single swap left -> right: 2 ticks, one share - current tick > upper tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               2,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap left -> right: 3 ticks, two shares - current tick > upper tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 2,
			currentTick:               3,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap left -> right: 2 ticks, one share - current tick == lower tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               0,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap left -> right: 2 ticks, one share - current tick < lower tick": {
			poolSetup:                 true,
			lowerTick:                 1,
			upperTick:                 2,
			currentTick:               0,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap left -> right: 2 ticks, one share - current tick == upper tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               1,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		// imagine single swap over entire position
		// crossing right > left and stopping at lower tick
		// In this case, all fees must have been accrued inside the tick
		// Since we track fees accrued below a tick, both upper and lower position
		// ticks are zero
		"single swap right -> left: 2 ticks, one share - current tick == lower tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               0,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap right -> left: 2 ticks, one share - current tick == upper tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               1,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap right -> left: 2 ticks, one share - current tick < lower tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               -1,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap right -> left: 2 ticks, one share - current tick > lower tick": {
			poolSetup:                 true,
			lowerTick:                 -1,
			upperTick:                 1,
			currentTick:               0,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap right -> left: 2 ticks, one share - current tick > upper tick": {
			poolSetup:                 true,
			lowerTick:                 -1,
			upperTick:                 1,
			currentTick:               2,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"error: pool has not been setup": {
			poolSetup:     false,
			expectedError: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// if pool set up true, set up default pool
			var pool types.ConcentratedPoolExtension
			if tc.poolSetup {
				pool = s.PrepareConcentratedPool()

				s.initializeTick(s.Ctx, tc.lowerTick, defaultInitialLiquidity, tc.lowerTickFeeGrowthOutside, false)
				s.initializeTick(s.Ctx, tc.upperTick, defaultInitialLiquidity, tc.upperTickFeeGrowthOutside, true)
				pool.SetCurrentTick(sdk.NewInt(tc.currentTick))
				s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
				err := s.App.ConcentratedLiquidityKeeper.ChargeFee(s.Ctx, validPoolId, tc.globalFeeGrowth)
				s.Require().NoError(err)
			}

			// system under test
			feeGrowthOutside, err := s.App.ConcentratedLiquidityKeeper.GetFeeGrowthOutside(s.Ctx, defaultPoolId, defaultLowerTickIndex, defaultUpperTickIndex)
			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// check if returned fee growth outside has correct value
				s.Require().Equal(feeGrowthOutside, tc.expectedFeeGrowthOutside)
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
			feeGrowth := cl.CalculateFeeGrowth(
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

	initialPoolTickInt, err := math.PriceToTick(DefaultAmt1.ToDec().Quo(DefaultAmt0.ToDec()), DefaultExponentAtPriceOne)
	initialPoolTick := initialPoolTickInt.Int64()
	suite.Require().NoError(err)

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

			expectedInitialFeeGrowthOutside: cl.EmptyCoins,
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

			expectError: accum.AccumDoesNotExistError{AccumName: cl.GetFeeAccumulatorName(validPoolId)},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			ctx := suite.Ctx
			clKeeper := suite.App.ConcentratedLiquidityKeeper

			pool, err := clmodel.NewConcentratedLiquidityPool(validPoolId, USDC, ETH, DefaultTickSpacing, DefaultExponentAtPriceOne, DefaultZeroSwapFee)
			suite.Require().NoError(err)

			err = clKeeper.SetPool(ctx, &pool)
			suite.Require().NoError(err)

			if !tc.shouldAvoidCreatingAccum {
				err = clKeeper.CreateFeeAccumulator(ctx, validPoolId)
				suite.Require().NoError(err)

				// Setup test position to make sure that tick is initialized
				// We also set up uptime accums to ensure position creation works as intended
				err = clKeeper.CreateUptimeAccumulators(ctx, validPoolId)
				suite.Require().NoError(err)
				suite.SetupDefaultPosition(validPoolId)

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

	tests := []struct {
		name      string
		poolId    uint64
		feeUpdate sdk.DecCoin

		expectedGlobalGrowth sdk.DecCoins
		expectError          error
	}{
		{
			name:      "pool id 1 - one eth",
			poolId:    1,
			feeUpdate: oneEth,

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth),
		},
		{
			name:      "pool id 1 - 2 usdc",
			poolId:    1,
			feeUpdate: sdk.NewDecCoin(USDC, sdk.NewInt(2)),

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth).Add(sdk.NewDecCoin(USDC, sdk.NewInt(2))),
		},
		{
			name:      "pool id 2 - 1 usdc",
			poolId:    2,
			feeUpdate: oneEth,

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth),
		},
		{
			name:      "accumulator does not exist",
			poolId:    3,
			feeUpdate: oneEth,

			expectError: accum.AccumDoesNotExistError{AccumName: cl.GetFeeAccumulatorName(3)},
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
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
	ownerWithValidPosition := s.TestAccs[0]
	defaultFrozenUntil := s.Ctx.BlockTime().Add(DefaultFreezeDuration)

	tests := map[string]struct {
		// setup parameters.
		initialLiquidity          sdk.Dec
		lowerTickFeeGrowthOutside sdk.DecCoins
		upperTickFeeGrowthOutside sdk.DecCoins
		globalFeeGrowth           sdk.DecCoins
		currentTick               int64
		isInvalidPoolIdGiven      bool

		// inputs parameters.
		owner       sdk.AccAddress
		lowerTick   int64
		upperTick   int64
		frozenUntil time.Time

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
		"single swap left -> right: 2 ticks, one share, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:       ownerWithValidPosition,
			lowerTick:   0,
			upperTick:   1,
			frozenUntil: defaultFrozenUntil,

			currentTick: 2,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick > upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:       ownerWithValidPosition,
			lowerTick:   0,
			upperTick:   2,
			frozenUntil: defaultFrozenUntil,

			currentTick: 3,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(20))),
		},
		"single swap left -> right: 2 ticks, one share, current tick == upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:       ownerWithValidPosition,
			lowerTick:   0,
			upperTick:   1,
			frozenUntil: defaultFrozenUntil,

			currentTick: 1,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},

		// imagine single swap over entire position
		// crossing right -> left and stopping at lower tick
		// In this case, all fees must have been accrued inside the tick
		// Since we track fees accrued below a tick, both upper and lower position
		// ticks are zero
		"single swap right -> left: 2 ticks, one share, current tick == lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator is not updated yet.
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:       ownerWithValidPosition,
			lowerTick:   0,
			upperTick:   1,
			frozenUntil: defaultFrozenUntil,

			currentTick: 0,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator updated when crossed.
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:       ownerWithValidPosition,
			lowerTick:   0,
			upperTick:   1,
			frozenUntil: defaultFrozenUntil,

			currentTick: -1,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},

		// imagine swap occurring outside of the position
		// As a result, lower and upper ticks are not updated.
		"swap occurs above the position, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			// none are updated.
			lowerTickFeeGrowthOutside: cl.EmptyCoins,
			upperTickFeeGrowthOutside: cl.EmptyCoins,

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:       ownerWithValidPosition,
			lowerTick:   0,
			upperTick:   1,
			frozenUntil: defaultFrozenUntil,

			currentTick: 5,

			expectedFeesClaimed: sdk.NewCoins(),
		},
		"swap occurs below the position, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			// none are updated.
			lowerTickFeeGrowthOutside: cl.EmptyCoins,
			upperTickFeeGrowthOutside: cl.EmptyCoins,

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:       ownerWithValidPosition,
			lowerTick:   -10,
			upperTick:   -4,
			frozenUntil: defaultFrozenUntil,

			currentTick: -13,

			expectedFeesClaimed: sdk.NewCoins(),
		},

		// error cases.

		"accumulator does not exist": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:       ownerWithValidPosition,
			lowerTick:   0,
			upperTick:   1,
			frozenUntil: defaultFrozenUntil,

			currentTick: 2,

			isInvalidPoolIdGiven: true,
			expectedError:        accum.AccumDoesNotExistError{AccumName: "fee/2"},
		},
		"position does not exist": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:       s.TestAccs[1], // different owner from the one who initialized the position.
			lowerTick:   0,
			upperTick:   1,
			frozenUntil: defaultFrozenUntil,

			currentTick: 2,

			expectedError: cltypes.PositionNotFoundError{PoolId: 1, LowerTick: 0, UpperTick: 1},
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

			s.initializeFeeAccumulatorPositionWithLiquidity(ctx, validPoolId, ownerWithValidPosition, tc.lowerTick, tc.upperTick, tc.initialLiquidity)

			s.initializeTick(ctx, tc.lowerTick, tc.initialLiquidity, tc.lowerTickFeeGrowthOutside, false)

			s.initializeTick(ctx, tc.upperTick, tc.initialLiquidity, tc.upperTickFeeGrowthOutside, true)

			validPool.SetCurrentTick(sdk.NewInt(tc.currentTick))
			clKeeper.SetPool(ctx, validPool)

			err := clKeeper.ChargeFee(ctx, validPoolId, tc.globalFeeGrowth[0])
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

			poolBalanceAfterCollect := s.App.BankKeeper.GetBalance(ctx, validPool.GetAddress(), ETH)
			ownerBalancerAfterCollect := s.App.BankKeeper.GetBalance(ctx, tc.owner, ETH)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Equal(sdk.Coins{}, actualFeesClaimed)

				// balances are unchanged
				s.Require().Equal(poolBalanceBeforeCollect, poolBalanceAfterCollect)
				s.Require().Equal(ownerBalancerAfterCollect, ownerBalancerBeforeCollect)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedFeesClaimed.String(), actualFeesClaimed.String())

			expectedETHAmount := tc.expectedFeesClaimed.AmountOf(ETH)
			s.Require().Equal(expectedETHAmount, poolBalanceBeforeCollect.Sub(poolBalanceAfterCollect).Amount)
			s.Require().Equal(expectedETHAmount, ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect).Amount)
		})
	}
}

func (s *KeeperTestSuite) TestUpdateFeeAccumulatorPosition() {
	var (
		ownerOne = s.TestAccs[0]
		ownerTwo = s.TestAccs[1]
	)
	defaultFrozenUntil := s.Ctx.BlockTime().Add(DefaultFreezeDuration)

	type updateFeeAccumPositionTest struct {
		poolId        uint64
		owner         sdk.AccAddress
		liquidity     sdk.Dec
		lowerTick     int64
		upperTick     int64
		frozenUntil   time.Time
		expectedError error
	}

	positions := map[int]updateFeeAccumPositionTest{
		1: {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			frozenUntil: defaultFrozenUntil,
		},
		2: {
			lowerTick:   DefaultLowerTick + 1,
			upperTick:   DefaultUpperTick,
			frozenUntil: defaultFrozenUntil,
		},
		3: {
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick + 1,
			frozenUntil: defaultFrozenUntil,
		},
	}

	tests := map[string]updateFeeAccumPositionTest{
		"update position with a different pool ID": {
			poolId:      2,
			owner:       ownerOne,
			liquidity:   DefaultLiquidityAmt,
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			frozenUntil: defaultFrozenUntil,
		},
		"update position with a different owner": {
			poolId:      1,
			owner:       ownerTwo,
			liquidity:   DefaultLiquidityAmt,
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick,
			frozenUntil: defaultFrozenUntil,
		},
		"update position with a different lower tick": {
			poolId:      1,
			owner:       ownerOne,
			liquidity:   DefaultLiquidityAmt,
			lowerTick:   DefaultLowerTick + 1,
			upperTick:   DefaultUpperTick,
			frozenUntil: defaultFrozenUntil,
		},
		"update position with a different upper tick": {
			poolId:      1,
			owner:       ownerOne,
			liquidity:   DefaultLiquidityAmt,
			lowerTick:   DefaultLowerTick,
			upperTick:   DefaultUpperTick + 1,
			frozenUntil: defaultFrozenUntil,
		},
		"err: pool does not exist": {
			poolId:        3,
			expectedError: cltypes.PoolNotFoundError{PoolId: 3},
		},
		"err: position does not exist": {
			poolId:        1,
			owner:         ownerOne,
			liquidity:     DefaultLiquidityAmt,
			lowerTick:     DefaultLowerTick - 1,
			upperTick:     DefaultUpperTick,
			frozenUntil:   defaultFrozenUntil,
			expectedError: accum.NoPositionError{Name: cl.FormatPositionAccumulatorKey(1, ownerOne, DefaultLowerTick-1, DefaultUpperTick)},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// Setup two cl pools
			poolOne := s.PrepareConcentratedPool()
			poolTwo := s.PrepareConcentratedPool()

			pools := []cltypes.ConcentratedPoolExtension{poolOne, poolTwo}
			owners := []sdk.AccAddress{ownerOne, ownerTwo}

			// Initialize three base positions in each pool for each owner (total of 12 positions)
			for _, pos := range positions {
				for _, pool := range pools {
					for _, owner := range owners {
						s.initializeFeeAccumulatorPositionWithLiquidity(s.Ctx, pool.GetId(), owner, pos.lowerTick, pos.upperTick, DefaultLiquidityAmt)
					}
				}
			}

			// System under test
			// Update one of the positions as per the test case
			err := s.App.ConcentratedLiquidityKeeper.UpdateFeeAccumulatorPosition(s.Ctx, tc.poolId, tc.owner, tc.liquidity, tc.lowerTick, tc.upperTick)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectedError)
				return
			}

			s.Require().NoError(err)

			// Validate the test case position was updated and all other positions did not change
			for _, pos := range positions {
				for _, pool := range pools {
					for _, owner := range owners {
						liq := DefaultLiquidityAmt
						if pool.GetId() == tc.poolId && owner.Equals(tc.owner) && pos.lowerTick == tc.lowerTick && pos.upperTick == tc.upperTick {
							liq = DefaultLiquidityAmt.Mul(sdk.NewDec(2))
						}
						s.validatePositionFeeAccUpdate(s.Ctx, pool.GetId(), owner, pos.lowerTick, pos.upperTick, pos.frozenUntil, liq)
					}
				}
			}
		})
	}
}
