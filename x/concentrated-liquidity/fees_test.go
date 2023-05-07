package concentrated_liquidity_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

const (
	NoUSDCExpected = ""
	NoETHExpected  = ""
)

// fields used to identify a fee position.
type positionFields struct {
	poolId     uint64
	owner      sdk.AccAddress
	lowerTick  int64
	upperTick  int64
	positionId uint64
	liquidity  sdk.Dec
}

var (
	oneEth      = sdk.NewDecCoin(ETH, sdk.OneInt())
	oneEthCoins = sdk.NewDecCoins(oneEth)
	onlyUSDC    = [][]string{{USDC}, {USDC}, {USDC}, {USDC}}
	onlyETH     = [][]string{{ETH}, {ETH}, {ETH}, {ETH}}
)

func (s *KeeperTestSuite) TestCreateAndGetFeeAccumulator() {
	type initFeeAccumTest struct {
		poolId              uint64
		initializePoolAccum bool

		expectError bool
	}
	tests := map[string]initFeeAccumTest{
		"default pool setup": {
			poolId:              defaultPoolId,
			initializePoolAccum: true,
		},
		"setup with different poolId": {
			poolId:              defaultPoolId + 1,
			initializePoolAccum: true,
		},
		"pool not initialized": {
			initializePoolAccum: false,
			poolId:              defaultPoolId,
			expectError:         true,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			// system under test
			if tc.initializePoolAccum {
				err := clKeeper.CreateFeeAccumulator(s.Ctx, tc.poolId)
				s.Require().NoError(err)
			}
			poolFeeAccumulator, err := clKeeper.GetFeeAccumulator(s.Ctx, tc.poolId)

			if !tc.expectError {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(accum.AccumulatorObject{}, poolFeeAccumulator)
			}
		})
	}
}

func (s *KeeperTestSuite) TestInitOrUpdateFeeAccumulatorPosition() {
	// Setup is done once so that we test
	// the relationship between test cases.
	// For example, that positions with non-zero liquidity
	// cannot be overriden.
	s.SetupTest()
	s.PrepareConcentratedPool()

	defaultAccount := s.TestAccs[0]

	var (
		defaultPoolId         = uint64(1)
		defaultPositionFields = positionFields{
			defaultPoolId,
			defaultAccount,
			DefaultLowerTick,
			DefaultUpperTick,
			DefaultPositionId,
			DefaultLiquidityAmt,
		}
	)

	withOwner := func(posFields positionFields, owner sdk.AccAddress) positionFields {
		posFields.owner = owner
		return posFields
	}

	withUpperTick := func(posFields positionFields, upperTick int64) positionFields {
		posFields.upperTick = upperTick
		return posFields
	}

	withLowerTick := func(posFields positionFields, lowerTick int64) positionFields {
		posFields.lowerTick = lowerTick
		return posFields
	}

	withPositionId := func(posFields positionFields, positionId uint64) positionFields {
		posFields.positionId = positionId
		return posFields
	}

	withLiquidity := func(posFields positionFields, liquidity sdk.Dec) positionFields {
		posFields.liquidity = liquidity
		return posFields
	}

	clKeeper := s.App.ConcentratedLiquidityKeeper

	type initFeeAccumTest struct {
		name           string
		positionFields positionFields

		expectedLiquidity sdk.Dec
		expectedPass      bool
	}
	tests := []initFeeAccumTest{
		{
			name:           "error: negative liquidity for the first position",
			positionFields: withLiquidity(defaultPositionFields, DefaultLiquidityAmt.Neg()),
			expectedPass:   false,
		},
		{
			name:              "first position",
			positionFields:    defaultPositionFields,
			expectedLiquidity: defaultPositionFields.liquidity,
			expectedPass:      true,
		},
		{
			name:              "second position",
			positionFields:    withPositionId(withLowerTick(defaultPositionFields, DefaultLowerTick+1), DefaultPositionId+1),
			expectedLiquidity: defaultPositionFields.liquidity,
			expectedPass:      true,
		},
		{
			name:              "adding to first position",
			positionFields:    defaultPositionFields,
			expectedPass:      true,
			expectedLiquidity: defaultPositionFields.liquidity.MulInt64(2),
		},
		{
			name:              "removing from first position",
			positionFields:    withLiquidity(defaultPositionFields, defaultPositionFields.liquidity.Neg()),
			expectedPass:      true,
			expectedLiquidity: defaultPositionFields.liquidity,
		},
		{
			name:              "adding to second position",
			positionFields:    withPositionId(withLowerTick(defaultPositionFields, DefaultLowerTick+1), DefaultPositionId+1),
			expectedPass:      true,
			expectedLiquidity: defaultPositionFields.liquidity.MulInt64(2),
		},
		{
			name: "error: non-existing accumulator (wrong pool)",
			positionFields: positionFields{
				defaultPoolId + 1, // non-existing pool
				defaultAccount,
				DefaultLowerTick,
				DefaultUpperTick,
				DefaultPositionId,
				DefaultLiquidityAmt,
			},
			expectedPass: false,
		},
		{
			name:              "existing accumulator, different owner - different position",
			positionFields:    withPositionId(withOwner(defaultPositionFields, s.TestAccs[1]), DefaultPositionId+2),
			expectedLiquidity: defaultPositionFields.liquidity,
			expectedPass:      true,
		},
		{
			name:              "existing accumulator, different upper tick - different position",
			positionFields:    withPositionId(withUpperTick(defaultPositionFields, DefaultUpperTick+1), DefaultPositionId+3),
			expectedLiquidity: defaultPositionFields.liquidity,
			expectedPass:      true,
		},
		{
			name:              "existing accumulator, different lower tick - different position",
			positionFields:    withPositionId(withLowerTick(defaultPositionFields, DefaultUpperTick+1), DefaultPositionId+4),
			expectedLiquidity: defaultPositionFields.liquidity,
			expectedPass:      true,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {

			// System under test
			err := clKeeper.InitOrUpdateFeeAccumulatorPosition(s.Ctx, tc.positionFields.poolId, tc.positionFields.lowerTick, tc.positionFields.upperTick, tc.positionFields.positionId, tc.positionFields.liquidity)
			if tc.expectedPass {
				s.Require().NoError(err)

				// get fee accum and see if position size has been properly initialized
				poolFeeAccumulator, err := clKeeper.GetFeeAccumulator(s.Ctx, defaultPoolId)
				s.Require().NoError(err)

				positionKey := cltypes.KeyFeePositionAccumulator(tc.positionFields.positionId)

				positionSize, err := poolFeeAccumulator.GetPositionSize(positionKey)
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedLiquidity, positionSize)

				positionRecord, err := poolFeeAccumulator.GetPosition(positionKey)
				s.Require().NoError(err)

				feeGrowthOutside, err := clKeeper.GetFeeGrowthOutside(s.Ctx, tc.positionFields.poolId, tc.positionFields.lowerTick, tc.positionFields.upperTick)
				s.Require().NoError(err)

				feeGrowthInside := poolFeeAccumulator.GetValue().Sub(feeGrowthOutside)

				// Position's accumulator must always equal to the fee growth inside the position.
				s.Require().Equal(feeGrowthInside, positionRecord.InitAccumValue)

				// Position's fee growth must be zero. Note, that on position update,
				// the unclaimed rewards are updated if there was fee growth. However,
				// this test case does not set up this condition.
				// It is tested in TestInitOrUpdateFeeAccumulatorPosition_UpdatingPosition.
				s.Require().Equal(cl.EmptyCoins, positionRecord.UnclaimedRewards)
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

	emptyUptimeTrackers := wrapUptimeTrackers(getExpectedUptimes().emptyExpectedAccumValues)

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
				currentTick := pool.GetCurrentTick().Int64()

				s.initializeTick(s.Ctx, currentTick, tc.lowerTick, defaultInitialLiquidity, tc.lowerTickFeeGrowthOutside, emptyUptimeTrackers, false)
				s.initializeTick(s.Ctx, currentTick, tc.upperTick, defaultInitialLiquidity, tc.upperTickFeeGrowthOutside, emptyUptimeTrackers, true)
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

	initialPoolTickInt, err := math.PriceToTickRoundDown(DefaultAmt1.ToDec().Quo(DefaultAmt0.ToDec()), DefaultTickSpacing)
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

			expectError: accum.AccumDoesNotExistError{AccumName: types.KeyFeePoolAccumulator(validPoolId)},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			ctx := suite.Ctx
			clKeeper := suite.App.ConcentratedLiquidityKeeper

			pool, err := clmodel.NewConcentratedLiquidityPool(validPoolId, ETH, USDC, DefaultTickSpacing, DefaultZeroSwapFee)
			suite.Require().NoError(err)

			// N.B.: we set the listener mock because we would like to avoid
			// utilizing the production listeners. The production listeners
			// are irrelevant in the context of the system under test. However,
			// setting them up would require compromising being able to set up other
			// edge case tests. For example, the test case where fee accumulator
			// is not initialized.
			suite.setListenerMockOnConcentratedLiquidityKeeper()

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

			expectError: accum.AccumDoesNotExistError{AccumName: types.KeyFeePoolAccumulator(3)},
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

func (s *KeeperTestSuite) TestQueryAndCollectFees() {
	ownerWithValidPosition := s.TestAccs[0]
	emptyUptimeTrackers := wrapUptimeTrackers(getExpectedUptimes().emptyExpectedAccumValues)

	tests := map[string]struct {
		// setup parameters.
		initialLiquidity          sdk.Dec
		lowerTickFeeGrowthOutside sdk.DecCoins
		upperTickFeeGrowthOutside sdk.DecCoins
		globalFeeGrowth           sdk.DecCoins
		currentTick               int64
		isInvalidPoolIdGiven      bool

		// inputs parameters.
		owner                       sdk.AccAddress
		lowerTick                   int64
		upperTick                   int64
		positionIdToCollectAndQuery uint64

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

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 2,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick > upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   2,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 3,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(20))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick is in between lower and upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   2,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 1,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(20))),
		},
		"single swap left -> right: 2 ticks, one share, current tick == upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

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

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 0,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator updated when crossed.
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: -1,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, two shares, current tick is in between lower and upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   2,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 1,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(20))),
		},

		// imagine swap occurring outside of the position
		// As a result, lower and upper ticks are not updated.
		"swap occurs above the position, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			// none are updated.
			lowerTickFeeGrowthOutside: cl.EmptyCoins,
			upperTickFeeGrowthOutside: cl.EmptyCoins,

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 5,

			expectedFeesClaimed: sdk.NewCoins(),
		},
		"swap occurs below the position, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			// none are updated.
			lowerTickFeeGrowthOutside: cl.EmptyCoins,
			upperTickFeeGrowthOutside: cl.EmptyCoins,

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   -10,
			upperTick:                   -4,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: -13,

			expectedFeesClaimed: sdk.NewCoins(),
		},

		// error cases.

		"position does not exist": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId + 1, // position id does not exist.

			currentTick: 2,

			expectedError: cltypes.PositionIdNotFoundError{PositionId: 2},
		},
		"non owner attempts to collect": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       s.TestAccs[1], // different owner from the one who initialized the position.
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 2,

			expectedError: cltypes.NotPositionOwnerError{Address: s.TestAccs[1].String(), PositionId: DefaultPositionId},
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

			// Set the position in store, otherwise querying via position id will fail.
			err := clKeeper.SetPosition(ctx, validPoolId, ownerWithValidPosition, tc.lowerTick, tc.upperTick, time.Now().UTC(), tc.initialLiquidity, DefaultPositionId, DefaultUnderlyingLockId)
			s.Require().NoError(err)

			s.initializeFeeAccumulatorPositionWithLiquidity(ctx, validPoolId, tc.lowerTick, tc.upperTick, DefaultPositionId, tc.initialLiquidity)

			s.initializeTick(ctx, tc.currentTick, tc.lowerTick, tc.initialLiquidity, tc.lowerTickFeeGrowthOutside, emptyUptimeTrackers, false)

			s.initializeTick(ctx, tc.currentTick, tc.upperTick, tc.initialLiquidity, tc.upperTickFeeGrowthOutside, emptyUptimeTrackers, true)

			validPool.SetCurrentTick(sdk.NewInt(tc.currentTick))
			clKeeper.SetPool(ctx, validPool)

			err = clKeeper.ChargeFee(ctx, validPoolId, tc.globalFeeGrowth[0])
			s.Require().NoError(err)

			poolBalanceBeforeCollect := s.App.BankKeeper.GetBalance(ctx, validPool.GetAddress(), ETH)
			ownerBalancerBeforeCollect := s.App.BankKeeper.GetBalance(ctx, tc.owner, ETH)

			var preQueryPosition accum.Record
			positionKey := cltypes.KeyFeePositionAccumulator(DefaultPositionId)

			// Note the position accumulator before the query to ensure the query in non-mutating.
			accum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(ctx, validPoolId)
			s.Require().NoError(err)
			preQueryPosition, _ = accum.GetPosition(positionKey)

			// System under test
			feeQueryAmount, queryErr := clKeeper.GetClaimableFees(ctx, tc.positionIdToCollectAndQuery)

			// If the query succeeds, the position should not be updated.
			if queryErr == nil {
				accum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(ctx, validPoolId)
				s.Require().NoError(err)
				postQueryPosition, _ := accum.GetPosition(positionKey)
				s.Require().Equal(preQueryPosition, postQueryPosition)
			}

			actualFeesClaimed, err := clKeeper.CollectFees(ctx, tc.owner, tc.positionIdToCollectAndQuery)

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
			s.Require().NoError(queryErr)
			s.Require().Equal(tc.expectedFeesClaimed.String(), actualFeesClaimed.String())
			s.Require().Equal(feeQueryAmount.String(), actualFeesClaimed.String())

			expectedETHAmount := tc.expectedFeesClaimed.AmountOf(ETH)
			s.Require().Equal(expectedETHAmount.String(), poolBalanceBeforeCollect.Sub(poolBalanceAfterCollect).Amount.String())
			s.Require().Equal(expectedETHAmount.String(), ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect).Amount.String())
		})
	}
}

func (s *KeeperTestSuite) TestPrepareClaimableFees() {
	emptyUptimeTrackers := wrapUptimeTrackers(getExpectedUptimes().emptyExpectedAccumValues)

	tests := map[string]struct {
		// setup parameters.
		initialLiquidity          sdk.Dec
		lowerTickFeeGrowthOutside sdk.DecCoins
		upperTickFeeGrowthOutside sdk.DecCoins
		globalFeeGrowth           sdk.DecCoins
		currentTick               int64
		isInvalidPoolIdGiven      bool

		// inputs parameters.
		lowerTick           int64
		upperTick           int64
		positionIdToPrepare uint64

		// expectations.
		expectedInitAccumValue sdk.DecCoins
		expectedError          error
	}{
		"single swap left -> right: 2 ticks, one share, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 2,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick > upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           2,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 3,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick in between lower and upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           2,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 2 ticks, one share, current tick == upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick == lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator is not updated yet.
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 0,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: -1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, two shares, current tick in between lower and upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           2,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"swap occurs above the position, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: cl.EmptyCoins,
			upperTickFeeGrowthOutside: cl.EmptyCoins,

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 5,

			expectedInitAccumValue: sdk.DecCoins(nil),
		},
		"swap occurs below the position, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: cl.EmptyCoins,
			upperTickFeeGrowthOutside: cl.EmptyCoins,

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           -10,
			upperTick:           -4,
			positionIdToPrepare: DefaultPositionId,

			currentTick: -13,

			expectedInitAccumValue: sdk.DecCoins(nil),
		},

		// error cases.
		"position does not exist": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId + 1, // position id does not exist.

			currentTick: 2,

			expectedError: cltypes.PositionIdNotFoundError{PositionId: 2},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			validPool := s.PrepareConcentratedPool()
			validPoolId := validPool.GetId()

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx

			// Set the position in store.
			err := clKeeper.SetPosition(ctx, validPoolId, s.TestAccs[0], tc.lowerTick, tc.upperTick, time.Now().UTC(), tc.initialLiquidity, DefaultPositionId, DefaultUnderlyingLockId)
			s.Require().NoError(err)

			s.initializeFeeAccumulatorPositionWithLiquidity(ctx, validPoolId, tc.lowerTick, tc.upperTick, DefaultPositionId, tc.initialLiquidity)
			s.initializeTick(ctx, tc.currentTick, tc.lowerTick, tc.initialLiquidity, tc.lowerTickFeeGrowthOutside, emptyUptimeTrackers, false)
			s.initializeTick(ctx, tc.currentTick, tc.upperTick, tc.initialLiquidity, tc.upperTickFeeGrowthOutside, emptyUptimeTrackers, true)
			validPool.SetCurrentTick(sdk.NewInt(tc.currentTick))
			clKeeper.SetPool(ctx, validPool)

			err = clKeeper.ChargeFee(ctx, validPoolId, tc.globalFeeGrowth[0])
			s.Require().NoError(err)

			positionKey := cltypes.KeyFeePositionAccumulator(DefaultPositionId)

			// Note the position accumulator before calling prepare
			accum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(ctx, validPoolId)
			s.Require().NoError(err)

			// System under test
			actualFeesClaimed, err := clKeeper.PrepareClaimableFees(ctx, tc.positionIdToPrepare)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Equal(sdk.Coins(nil), actualFeesClaimed)
				return
			}
			s.Require().NoError(err)

			accum, err = s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(ctx, validPoolId)
			s.Require().NoError(err)

			postPreparePosition, err := accum.GetPosition(positionKey)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedInitAccumValue, postPreparePosition.InitAccumValue)
			s.Require().Equal(tc.initialLiquidity, postPreparePosition.NumShares)

			expectedFeeClaimAmount := tc.expectedInitAccumValue.AmountOf(ETH).Mul(tc.initialLiquidity).TruncateInt()
			s.Require().Equal(expectedFeeClaimAmount, actualFeesClaimed.AmountOf(ETH))
		})
	}
}

// This test ensures that the position's fee accumulator is updated correctly when the fee grows.
// It validates that another position within the same tick does not affect the current position.
// It also validates that the position's changes are applied at the right time relative to position's
// fee accumulator creation or update.
func (s *KeeperTestSuite) TestInitOrUpdateFeeAccumulatorPosition_UpdatingPosition() {
	type updateFeeAccumPositionTest struct {
		doesFeeGrowBeforeFirstCall           bool
		doesFeeGrowBetweenFirstAndSecondCall bool
		doesFeeGrowBetweenSecondAndThirdCall bool
		doesFeeGrowAfterThirdCall            bool

		expectedUnclaimedRewardsPositionOne sdk.DecCoins
		expectedUnclaimedRewardsPositionTwo sdk.DecCoins
	}

	tests := map[string]updateFeeAccumPositionTest{
		"1: fee charged prior to first call to InitOrUpdateFeeAccumulatorPosition with position one": {
			doesFeeGrowBeforeFirstCall: true,

			// Growing fee before first position has no effect on the unclaimed rewards
			// of either position because they are not initialized at that point.
			expectedUnclaimedRewardsPositionOne: cl.EmptyCoins,
			// For position two, growing fee has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
		"2: fee charged between first and second call to InitOrUpdateFeeAccumulatorPosition, after position one is created and before position two is created": {
			doesFeeGrowBetweenFirstAndSecondCall: true,

			// Position one's unclaimed rewards increase.
			expectedUnclaimedRewardsPositionOne: DefaultFeeAccumCoins,
			// For position two, growing fee has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
		"3: fee charged between second and third call to InitOrUpdateFeeAccumulatorPosition, after position two is created and before position 1 is updated": {
			doesFeeGrowBetweenSecondAndThirdCall: true,

			// fee charged because it grows between the second and third position being created.
			// when third position is created, the rewards are moved to unclaimed.
			expectedUnclaimedRewardsPositionOne: DefaultFeeAccumCoins,
			// For position two, growing fee has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
		"4: fee charged after third call to InitOrUpdateFeeAccumulatorPosition, after position 1 is updated": {
			doesFeeGrowAfterThirdCall: true,

			// no fee charged because it grows after the position is updated and the rewards are moved to unclaimed.
			expectedUnclaimedRewardsPositionOne: cl.EmptyCoins,
			// For position two, growing fee has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			pool := s.PrepareConcentratedPool()
			poolId := pool.GetId()

			pool.SetCurrentTick(DefaultCurrTick)

			// Imaginary fee charge #1.
			if tc.doesFeeGrowBeforeFirstCall {
				s.crossTickAndChargeFee(poolId, DefaultLowerTick)
			}

			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, poolId, pool.GetCurrentTick().Int64(), DefaultLowerTick, DefaultLiquidityAmt, false)
			s.Require().NoError(err)

			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, poolId, pool.GetCurrentTick().Int64(), DefaultUpperTick, DefaultLiquidityAmt, true)
			s.Require().NoError(err)

			// InitOrUpdateFeeAccumulatorPosition #1 lower tick to upper tick
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateFeeAccumulatorPosition(s.Ctx, poolId, DefaultLowerTick, DefaultUpperTick, DefaultPositionId, DefaultLiquidityAmt)
			s.Require().NoError(err)

			// Imaginary fee charge #2.
			if tc.doesFeeGrowBetweenFirstAndSecondCall {
				s.crossTickAndChargeFee(poolId, DefaultLowerTick)
			}

			// InitOrUpdateFeeAccumulatorPosition # 2 lower tick to upper tick with a different position id.
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateFeeAccumulatorPosition(s.Ctx, poolId, DefaultLowerTick, DefaultUpperTick, DefaultPositionId+1, DefaultLiquidityAmt)
			s.Require().NoError(err)

			// Imaginary fee charge #3.
			if tc.doesFeeGrowBetweenSecondAndThirdCall {
				s.crossTickAndChargeFee(poolId, DefaultLowerTick)
			}

			// InitOrUpdateFeeAccumulatorPosition # 3 lower tick to upper tick with the original position id.
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateFeeAccumulatorPosition(s.Ctx, poolId, DefaultLowerTick, DefaultUpperTick, DefaultPositionId, DefaultLiquidityAmt)
			s.Require().NoError(err)

			// Imaginary fee charge #4.
			if tc.doesFeeGrowAfterThirdCall {
				s.crossTickAndChargeFee(poolId, DefaultLowerTick)
			}

			// Validate original position's fee growth.
			s.validatePositionFeeGrowth(poolId, DefaultPositionId, tc.expectedUnclaimedRewardsPositionOne)

			// Validate second position's fee growth.
			s.validatePositionFeeGrowth(poolId, DefaultPositionId+1, tc.expectedUnclaimedRewardsPositionTwo)

			// Validate position one was updated with default liquidity twice.
			s.validatePositionFeeAccUpdate(s.Ctx, poolId, DefaultPositionId, DefaultLiquidityAmt.MulInt64(2))

			// Validate position two was updated with default liquidity once.
			s.validatePositionFeeAccUpdate(s.Ctx, poolId, DefaultPositionId+1, DefaultLiquidityAmt)
		})
	}
}

func (s *KeeperTestSuite) TestPreparePositionAccumulator() {
	validPositionKey := cltypes.KeyFeePositionAccumulator(1)
	invalidPositionKey := cltypes.KeyFeePositionAccumulator(2)
	tests := []struct {
		name               string
		poolId             uint64
		feeGrowthOutside   sdk.DecCoins
		invalidPositionKey bool
		expectError        error
	}{
		{
			name:             "happy path",
			feeGrowthOutside: oneEthCoins,
		},
		{
			name:               "error: non existent accumulator",
			feeGrowthOutside:   oneEthCoins,
			invalidPositionKey: true,
			expectError:        accum.NoPositionError{Name: invalidPositionKey},
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// Setup test env.
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper
			s.PrepareConcentratedPool()
			poolFeeAccumulator, err := clKeeper.GetFeeAccumulator(s.Ctx, defaultPoolId)
			s.Require().NoError(err)
			positionKey := validPositionKey

			// Initialize position accumulator.
			err = poolFeeAccumulator.NewPositionCustomAcc(positionKey, sdk.OneDec(), sdk.DecCoins{}, nil)
			s.Require().NoError(err)

			// Record the initial position accumulator value.
			positionPre, err := accum.GetPosition(poolFeeAccumulator, positionKey)
			s.Require().NoError(err)

			// If the test case requires an invalid position key, set it.
			if tc.invalidPositionKey {
				positionKey = invalidPositionKey
			}

			// System under test.
			err = cl.PreparePositionAccumulator(poolFeeAccumulator, positionKey, tc.feeGrowthOutside)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			// Record the final position accumulator value.
			positionPost, err := accum.GetPosition(poolFeeAccumulator, positionKey)
			s.Require().NoError(err)

			// Check that the difference between the new and old position accumulator values is equal to the fee growth outside.
			positionAccumDelta := positionPost.InitAccumValue.Sub(positionPre.InitAccumValue)
			s.Require().Equal(tc.feeGrowthOutside, positionAccumDelta)
		})
	}
}

type Positions struct {
	numSwaps       int
	numAccounts    int
	numFullRange   int
	numNarrowRange int
	numConsecutive int
	numOverlapping int
}

func (s *KeeperTestSuite) TestFunctional_Fees_Swaps() {
	positions := Positions{
		numSwaps:       7,
		numAccounts:    5,
		numFullRange:   4,
		numNarrowRange: 3,
		numConsecutive: 2,
		numOverlapping: 1,
	}
	// Init suite.
	s.SetupTest()

	// Default setup only creates 3 accounts, but we need 5 for this test.
	s.TestAccs = apptesting.CreateRandomAccounts(positions.numAccounts)

	// Create a default CL pool, but with a 0.3 percent swap fee.
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, sdk.MustNewDecFromStr("0.002"))

	positionIds := make([][]uint64, 4)
	// Setup full range position across all four accounts
	for i := 0; i < positions.numFullRange; i++ {
		positionId := s.SetupFullRangePositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[0] = append(positionIds[0], positionId)
	}

	// Setup narrow range position across three of four accounts
	for i := 0; i < positions.numNarrowRange; i++ {
		positionId := s.SetupDefaultPositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[1] = append(positionIds[1], positionId)
	}

	// Setup consecutive range position (in relation to narrow range position) across two of four accounts
	for i := 0; i < positions.numConsecutive; i++ {
		positionId := s.SetupConsecutiveRangePositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[2] = append(positionIds[2], positionId)
	}

	// Setup overlapping range position (in relation to narrow range position) on one of four accounts
	for i := 0; i < positions.numOverlapping; i++ {
		positionId := s.SetupOverlappingRangePositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[3] = append(positionIds[3], positionId)
	}

	// Swap multiple times USDC for ETH, therefore increasing the spot price
	ticksActivatedAfterEachSwap, totalFeesExpected, _, _ := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin1, ETH, cltypes.MaxSpotPrice, positions.numSwaps)
	s.CollectAndAssertFees(s.Ctx, clPool.GetId(), totalFeesExpected, positionIds, [][]sdk.Int{ticksActivatedAfterEachSwap}, onlyUSDC, positions)

	// Swap multiple times ETH for USDC, therefore decreasing the spot price
	ticksActivatedAfterEachSwap, totalFeesExpected, _, _ = s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin0, USDC, cltypes.MinSpotPrice, positions.numSwaps)
	s.CollectAndAssertFees(s.Ctx, clPool.GetId(), totalFeesExpected, positionIds, [][]sdk.Int{ticksActivatedAfterEachSwap}, onlyETH, positions)

	// Do the same swaps as before, however this time we collect fees after both swap directions are complete.
	ticksActivatedAfterEachSwapUp, totalFeesExpectedUp, _, _ := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin1, ETH, cltypes.MaxSpotPrice, positions.numSwaps)
	ticksActivatedAfterEachSwapDown, totalFeesExpectedDown, _, _ := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin0, USDC, cltypes.MinSpotPrice, positions.numSwaps)
	totalFeesExpected = totalFeesExpectedUp.Add(totalFeesExpectedDown...)

	// We expect all positions to have both denoms in their fee accumulators except USDC for the overlapping range position since
	// it was not activated during the USDC -> ETH swap direction but was activated during the ETH -> USDC swap direction.
	ticksActivatedAfterEachSwapTest := [][]sdk.Int{ticksActivatedAfterEachSwapUp, ticksActivatedAfterEachSwapDown}
	denomsExpected := [][]string{{USDC, ETH}, {USDC, ETH}, {USDC, ETH}, {NoUSDCExpected, ETH}}

	s.CollectAndAssertFees(s.Ctx, clPool.GetId(), totalFeesExpected, positionIds, ticksActivatedAfterEachSwapTest, denomsExpected, positions)
}

// This test focuses on various functional testing around fees and LP logic.
// It tests invariants such as the following:
// - can create positions in the same range, swap between them and yet collect the correct fees.
// - correct proportions of fees for overlapping positions are withdrawn.
// - withdrawing full liquidity claims correctly under the hood.
// - withdrawing partial liquidity does not withdraw but still lets fee claim as desired.
func (s *KeeperTestSuite) TestFunctional_Fees_LP() {
	// Setup.
	s.SetupTest()
	s.TestAccs = apptesting.CreateRandomAccounts(5)

	var (
		ctx                         = s.Ctx
		concentratedLiquidityKeeper = s.App.ConcentratedLiquidityKeeper
		owner                       = s.TestAccs[0]
	)

	// Create pool with 0.2% swap fee.
	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, sdk.MustNewDecFromStr("0.002"))
	fundCoins := sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0.MulRaw(2)), sdk.NewCoin(USDC, DefaultAmt1.MulRaw(2)))
	s.FundAcc(owner, fundCoins)

	// Errors since no position.
	_, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(s.Ctx, owner, pool, sdk.NewCoin(ETH, sdk.OneInt()), USDC, pool.GetSwapFee(s.Ctx), types.MaxSpotPrice)
	s.Require().Error(err)

	// Create position in the default range 1.
	positionIdOne, _, _, liquidity, _, err := concentratedLiquidityKeeper.CreatePosition(ctx, pool.GetId(), owner, DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
	s.Require().NoError(err)

	// Swap once.
	ticksActivatedAfterEachSwap, totalFeesExpected, _, _ := s.swapAndTrackXTimesInARow(pool.GetId(), DefaultCoin1, ETH, cltypes.MaxSpotPrice, 1)

	// Withdraw half.
	halfLiquidity := liquidity.Mul(sdk.NewDecWithPrec(5, 1))
	_, _, err = concentratedLiquidityKeeper.WithdrawPosition(ctx, owner, positionIdOne, halfLiquidity)
	s.Require().NoError(err)

	// Collect fees.
	feesCollected := s.collectFeesAndCheckInvariance(ctx, 0, DefaultMinTick, DefaultMaxTick, positionIdOne, sdk.NewCoins(), []string{USDC}, [][]sdk.Int{ticksActivatedAfterEachSwap})
	s.Require().Equal(totalFeesExpected, feesCollected)

	// Unclaimed rewards should be emptied since fees were collected.
	s.validatePositionFeeGrowth(pool.GetId(), positionIdOne, cl.EmptyCoins)

	// Create position in the default range 2.
	positionIdTwo, _, _, fullLiquidity, _, err := concentratedLiquidityKeeper.CreatePosition(ctx, pool.GetId(), owner, DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
	s.Require().NoError(err)

	// Swap once in the other direction.
	ticksActivatedAfterEachSwap, totalFeesExpected, _, _ = s.swapAndTrackXTimesInARow(pool.GetId(), DefaultCoin0, USDC, cltypes.MinSpotPrice, 1)

	// This should claim under the hood for position 2 since full liquidity is removed.
	balanceBeforeWithdraw := s.App.BankKeeper.GetBalance(ctx, owner, ETH)
	amtDenom0, _, err := concentratedLiquidityKeeper.WithdrawPosition(ctx, owner, positionIdTwo, fullLiquidity)
	s.Require().NoError(err)
	balanceAfterWithdraw := s.App.BankKeeper.GetBalance(ctx, owner, ETH)

	// Validate that the correct amount of ETH was collected in withdraw for position two.
	// total fees * full liquidity / (full liquidity + half liquidity)
	expectedPositionToWithdraw := totalFeesExpected.AmountOf(ETH).ToDec().Mul(fullLiquidity.Quo(fullLiquidity.Add(halfLiquidity))).TruncateInt()
	s.Require().Equal(expectedPositionToWithdraw.String(), balanceAfterWithdraw.Sub(balanceBeforeWithdraw).Amount.Sub(amtDenom0).String())

	// Validate cannot claim for withdrawn position.
	_, err = s.App.ConcentratedLiquidityKeeper.CollectFees(ctx, owner, positionIdTwo)
	s.Require().Error(err)

	feesCollected = s.collectFeesAndCheckInvariance(ctx, 0, DefaultMinTick, DefaultMaxTick, positionIdOne, sdk.NewCoins(), []string{ETH}, [][]sdk.Int{ticksActivatedAfterEachSwap})

	// total fees * half liquidity / (full liquidity + half liquidity)
	expectesFeesCollected := totalFeesExpected.AmountOf(ETH).ToDec().Mul(halfLiquidity.Quo(fullLiquidity.Add(halfLiquidity))).TruncateInt()
	s.Require().Equal(expectesFeesCollected.String(), feesCollected.AmountOf(ETH).String())

	// Create position in the default range 3.
	positionIdThree, _, _, fullLiquidity, _, err := concentratedLiquidityKeeper.CreatePosition(ctx, pool.GetId(), owner, DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
	s.Require().NoError(err)

	collectedThree, err := s.App.ConcentratedLiquidityKeeper.CollectFees(ctx, owner, positionIdThree)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins(nil), collectedThree)
}

// CollectAndAssertFees collects fees from a given pool for all positions and verifies that the total fees collected match the expected total fees.
// The method also checks that if the ticks that were active during the swap lie within the range of a position, then the position's fee accumulators
// are not empty. The total fees collected are compared to the expected total fees within an additive tolerance defined by an error tolerance struct.
func (s *KeeperTestSuite) CollectAndAssertFees(ctx sdk.Context, poolId uint64, totalFees sdk.Coins, positionIds [][]uint64, activeTicks [][]sdk.Int, expectedFeeDenoms [][]string, positions Positions) {
	var totalFeesCollected sdk.Coins
	// Claim full range position fees across all four accounts
	for i := 0; i < positions.numFullRange; i++ {
		totalFeesCollected = s.collectFeesAndCheckInvariance(ctx, i, DefaultMinTick, DefaultMaxTick, positionIds[0][i], totalFeesCollected, expectedFeeDenoms[0], activeTicks)
	}

	// Claim narrow range position fees across three of four accounts
	for i := 0; i < positions.numNarrowRange; i++ {
		totalFeesCollected = s.collectFeesAndCheckInvariance(ctx, i, DefaultLowerTick, DefaultUpperTick, positionIds[1][i], totalFeesCollected, expectedFeeDenoms[1], activeTicks)
	}

	// Claim consecutive range position fees across two of four accounts
	for i := 0; i < positions.numConsecutive; i++ {
		totalFeesCollected = s.collectFeesAndCheckInvariance(ctx, i, DefaultExponentConsecutivePositionLowerTick.Int64(), DefaultExponentConsecutivePositionUpperTick.Int64(), positionIds[2][i], totalFeesCollected, expectedFeeDenoms[2], activeTicks)
	}

	// Claim overlapping range position fees on one of four accounts
	for i := 0; i < positions.numOverlapping; i++ {
		totalFeesCollected = s.collectFeesAndCheckInvariance(ctx, i, DefaultExponentOverlappingPositionLowerTick.Int64(), DefaultExponentOverlappingPositionUpperTick.Int64(), positionIds[3][i], totalFeesCollected, expectedFeeDenoms[3], activeTicks)
	}

	// Define error tolerance
	var errTolerance osmomath.ErrTolerance
	errTolerance.AdditiveTolerance = sdk.NewDec(10)
	errTolerance.RoundingDir = osmomath.RoundDown

	// Check that the total fees collected is equal to the total fees (within a tolerance)
	for _, coin := range totalFeesCollected {
		expected := totalFees.AmountOf(coin.Denom)
		actual := coin.Amount
		s.Require().Equal(0, errTolerance.Compare(expected, actual), fmt.Sprintf("expected (%s), actual (%s)", expected, actual))
	}
}

// tickStatusInvariance tests if the swap position was active during the given tick range and
// checks that the fees collected are non-zero if the position was active, or zero otherwise.
func (s *KeeperTestSuite) tickStatusInvariance(ticksActivatedAfterEachSwap [][]sdk.Int, lowerTick, upperTick int64, coins sdk.Coins, expectedFeeDenoms []string) {
	var positionWasActive bool
	// Check if the position was active during the swap
	for i, ticks := range ticksActivatedAfterEachSwap {
		for _, tick := range ticks {
			if tick.GTE(sdk.NewInt(lowerTick)) && tick.LTE(sdk.NewInt(upperTick)) {
				positionWasActive = true
				break
			}
		}
		if positionWasActive {
			// If the position was active, check that the fees collected are non-zero
			if expectedFeeDenoms[i] != NoUSDCExpected && expectedFeeDenoms[i] != NoETHExpected {
				s.Require().True(coins.AmountOf(expectedFeeDenoms[i]).GT(sdk.ZeroInt()), "denom: %s", expectedFeeDenoms[i])
			}
		} else {
			// If the position was not active, check that the fees collected are zero
			s.Require().Nil(coins)
		}
	}
}

// swapAndTrackXTimesInARow performs `numSwaps` swaps and tracks the tick activated after each swap.
// It also returns the total fees collected, the total token in, and the total token out.
func (s *KeeperTestSuite) swapAndTrackXTimesInARow(poolId uint64, coinIn sdk.Coin, coinOutDenom string, priceLimit sdk.Dec, numSwaps int) (ticksActivatedAfterEachSwap []sdk.Int, totalFees sdk.Coins, totalTokenIn sdk.Coin, totalTokenOut sdk.Coin) {
	// Retrieve pool
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	// Determine amount needed to fulfill swap numSwaps times and fund account that much
	amountNeededForSwap := coinIn.Amount.Mul(sdk.NewInt(int64(numSwaps)))
	swapCoin := sdk.NewCoin(coinIn.Denom, amountNeededForSwap)
	s.FundAcc(s.TestAccs[4], sdk.NewCoins(swapCoin))

	ticksActivatedAfterEachSwap = append(ticksActivatedAfterEachSwap, clPool.GetCurrentTick())
	totalFees = sdk.NewCoins(sdk.NewCoin(USDC, sdk.ZeroInt()), sdk.NewCoin(ETH, sdk.ZeroInt()))
	// Swap numSwaps times, recording the tick activated after and swap and fees we expect to collect based on the amount in
	totalTokenIn = sdk.NewCoin(coinIn.Denom, sdk.ZeroInt())
	totalTokenOut = sdk.NewCoin(coinOutDenom, sdk.ZeroInt())
	for i := 0; i < numSwaps; i++ {
		coinIn, coinOut, _, _, _, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(s.Ctx, s.TestAccs[4], clPool, coinIn, coinOutDenom, clPool.GetSwapFee(s.Ctx), priceLimit)
		s.Require().NoError(err)
		fee := coinIn.Amount.ToDec().Mul(clPool.GetSwapFee(s.Ctx))
		totalFees = totalFees.Add(sdk.NewCoin(coinIn.Denom, fee.TruncateInt()))
		clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
		s.Require().NoError(err)
		ticksActivatedAfterEachSwap = append(ticksActivatedAfterEachSwap, clPool.GetCurrentTick())
		totalTokenIn = totalTokenIn.Add(coinIn)
		totalTokenOut = totalTokenOut.Add(coinOut)
	}
	return ticksActivatedAfterEachSwap, totalFees, totalTokenIn, totalTokenOut
}

// collectFeesAndCheckInvariance collects fees from the concentrated liquidity pool and checks the resulting tick status invariance.
func (s *KeeperTestSuite) collectFeesAndCheckInvariance(ctx sdk.Context, accountIndex int, minTick, maxTick int64, positionId uint64, feesCollected sdk.Coins, expectedFeeDenoms []string, activeTicks [][]sdk.Int) (totalFeesCollected sdk.Coins) {
	coins, err := s.App.ConcentratedLiquidityKeeper.CollectFees(ctx, s.TestAccs[accountIndex], positionId)
	s.Require().NoError(err)
	totalFeesCollected = feesCollected.Add(coins...)
	s.tickStatusInvariance(activeTicks, minTick, maxTick, coins, expectedFeeDenoms)
	return totalFeesCollected
}
