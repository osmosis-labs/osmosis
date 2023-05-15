package concentrated_liquidity_test

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/client/queryproto"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
)

const validPoolId = 1

func withTickIndex(tick genesis.FullTick, tickIndex int64) genesis.FullTick {
	tick.TickIndex = tickIndex
	return tick
}

func withPoolId(tick genesis.FullTick, poolId uint64) genesis.FullTick {
	tick.PoolId = poolId
	return tick
}

func withLiquidityNetandTickIndex(tick genesis.FullTick, tickIndex int64, liquidityNet sdk.Dec) genesis.FullTick {
	tick.TickIndex = tickIndex
	tick.Info.LiquidityNet = liquidityNet

	return tick
}

func (s *KeeperTestSuite) TestTickOrdering() {
	s.SetupTest()

	storeKey := sdk.NewKVStoreKey("concentrated_liquidity")
	tKey := sdk.NewTransientStoreKey("transient_test")
	s.Ctx = testutil.DefaultContext(storeKey, tKey)
	s.App.ConcentratedLiquidityKeeper = cl.NewKeeper(s.App.AppCodec(), storeKey, s.App.AccountKeeper, s.App.BankKeeper, s.App.GAMMKeeper, s.App.PoolIncentivesKeeper, s.App.IncentivesKeeper, s.App.LockupKeeper, s.App.DistrKeeper, s.App.GetSubspace(types.ModuleName))

	liquidityTicks := []int64{-200, -55, -4, 70, 78, 84, 139, 240, 535}
	for _, t := range liquidityTicks {
		s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, 1, t, model.TickInfo{})
	}

	store := s.Ctx.KVStore(storeKey)
	prefixBz := types.KeyTickPrefixByPoolId(1)
	prefixStore := prefix.NewStore(store, prefixBz)

	// Pick a value and ensure ordering is correct for lte=false, i.e. increasing
	// ticks.
	startKey := types.TickIndexToBytes(-4)
	iter := prefixStore.Iterator(startKey, nil)
	defer iter.Close()

	var vals []int64
	for ; iter.Valid(); iter.Next() {
		tick, err := types.TickIndexFromBytes(iter.Key())
		s.Require().NoError(err)

		vals = append(vals, tick)
	}

	s.Require().Equal([]int64{-4, 70, 78, 84, 139, 240, 535}, vals)

	// Pick a value and ensure ordering is correct for lte=true, i.e. decreasing
	// ticks.
	startKey = types.TickIndexToBytes(84)
	revIter := prefixStore.ReverseIterator(nil, startKey)
	defer revIter.Close()

	vals = nil
	for ; revIter.Valid(); revIter.Next() {
		tick, err := types.TickIndexFromBytes(revIter.Key())
		s.Require().NoError(err)

		vals = append(vals, tick)
	}

	s.Require().Equal([]int64{78, 70, -4, -55, -200}, vals)
}

func (s *KeeperTestSuite) TestInitOrUpdateTick() {
	type param struct {
		poolId      uint64
		tickIndex   int64
		liquidityIn sdk.Dec
		upper       bool
	}

	tests := []struct {
		name                   string
		param                  param
		tickExists             bool
		expectedLiquidityNet   sdk.Dec
		expectedLiquidityGross sdk.Dec
		minimumGasConsumed     uint64
		expectedErr            error
	}{
		{
			name: "Init tick 50 with DefaultLiquidityAmt liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       true,
			},
			tickExists:             false,
			expectedLiquidityNet:   DefaultLiquidityAmt.Neg(),
			expectedLiquidityGross: DefaultLiquidityAmt,
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Init tick 50 with DefaultLiquidityAmt liquidity, lower",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       false,
			},
			tickExists:             false,
			expectedLiquidityNet:   DefaultLiquidityAmt,
			expectedLiquidityGross: DefaultLiquidityAmt,
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Update tick 50 that already contains DefaultLiquidityAmt liquidity with DefaultLiquidityAmt more liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       true,
			},
			tickExists:             true,
			expectedLiquidityNet:   DefaultLiquidityAmt.Mul(sdk.NewDec(2)).Neg(),
			expectedLiquidityGross: DefaultLiquidityAmt.Mul(sdk.NewDec(2)),
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Update tick 50 that already contains DefaultLiquidityAmt liquidity with DefaultLiquidityAmt more liquidity, lower",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       false,
			},
			tickExists:             true,
			expectedLiquidityNet:   DefaultLiquidityAmt.Mul(sdk.NewDec(2)),
			expectedLiquidityGross: DefaultLiquidityAmt.Mul(sdk.NewDec(2)),
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Init tick -50 with DefaultLiquidityAmt liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   -50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       true,
			},
			tickExists:             false,
			expectedLiquidityNet:   DefaultLiquidityAmt.Neg(),
			expectedLiquidityGross: DefaultLiquidityAmt,
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Init tick -50 with DefaultLiquidityAmt liquidity, lower",
			param: param{
				poolId:      validPoolId,
				tickIndex:   -50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       false,
			},
			tickExists:             false,
			expectedLiquidityNet:   DefaultLiquidityAmt,
			expectedLiquidityGross: DefaultLiquidityAmt,
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Update tick -50 that already contains DefaultLiquidityAmt liquidity with DefaultLiquidityAmt more liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   -50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       true,
			},
			tickExists:             true,
			expectedLiquidityNet:   DefaultLiquidityAmt.Mul(sdk.NewDec(2)).Neg(),
			expectedLiquidityGross: DefaultLiquidityAmt.Mul(sdk.NewDec(2)),
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Update tick -50 that already contains DefaultLiquidityAmt liquidity with DefaultLiquidityAmt more liquidity, lower",
			param: param{
				poolId:      validPoolId,
				tickIndex:   -50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       false,
			},
			tickExists:             true,
			expectedLiquidityNet:   DefaultLiquidityAmt.Mul(sdk.NewDec(2)),
			expectedLiquidityGross: DefaultLiquidityAmt.Mul(sdk.NewDec(2)),
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Init tick 50 with Negative DefaultLiquidityAmt liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt.Neg(),
				upper:       true,
			},
			tickExists:             false,
			expectedLiquidityNet:   DefaultLiquidityAmt,
			expectedLiquidityGross: DefaultLiquidityAmt.Neg(),
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Update tick 50 that already contains DefaultLiquidityAmt liquidity with -DefaultLiquidityAmt liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt.Neg(),
				upper:       true,
			},
			tickExists:             true,
			expectedLiquidityNet:   sdk.ZeroDec(),
			expectedLiquidityGross: sdk.ZeroDec(),
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Update tick -50 that already contains DefaultLiquidityAmt liquidity with negative DefaultLiquidityAmt liquidity, lower",
			param: param{
				poolId:      validPoolId,
				tickIndex:   -50,
				liquidityIn: DefaultLiquidityAmt.Neg(),
				upper:       false,
			},
			tickExists:             true,
			expectedLiquidityNet:   sdk.ZeroDec(),
			expectedLiquidityGross: sdk.ZeroDec(),
			minimumGasConsumed:     uint64(types.BaseGasFeeForInitializingTick),
		},
		{
			name: "Init tick for non-existing pool",
			param: param{
				poolId:      2,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       true,
			},
			tickExists:  false,
			expectedErr: types.PoolNotFoundError{PoolId: 2},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()

			// Create a default CL pool
			pool := s.PrepareConcentratedPool()
			currentTick := pool.GetCurrentTick().Int64()

			_, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(s.Ctx, 1)
			s.Require().NoError(err)
			feeAccum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(s.Ctx, 1)
			s.Require().NoError(err)

			// manually update accumulator for testing
			defaultAccumCoins := sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(50)))
			feeAccum.AddToAccumulator(defaultAccumCoins)

			existingGasConsumed := s.Ctx.GasMeter().GasConsumed()

			// If tickExists set, initialize the specified tick with defaultLiquidityAmt
			preexistingLiquidity := sdk.ZeroDec()
			if test.tickExists {
				tickInfoBefore, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
				s.Require().NoError(err)
				err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, test.param.poolId, currentTick, test.param.tickIndex, DefaultLiquidityAmt, test.param.upper)
				s.Require().NoError(err)
				if tickInfoBefore.LiquidityGross.IsZero() && test.param.tickIndex <= pool.GetCurrentTick().Int64() {
					tickInfoAfter, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
					s.Require().NoError(err)
					s.Require().Equal(tickInfoAfter.FeeGrowthOppositeDirectionOfLastTraversal, feeAccum.GetValue())
				}
				preexistingLiquidity = DefaultLiquidityAmt
			}

			// Get the tick info for poolId 1
			tickInfoAfter, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
			s.Require().NoError(err)

			// Ensure tick state contains any preexistingLiquidity (zero otherwise)
			s.Require().Equal(preexistingLiquidity, tickInfoAfter.LiquidityGross)

			// Initialize or update the tick according to the test case
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, test.param.poolId, currentTick, test.param.tickIndex, test.param.liquidityIn, test.param.upper)
			if tickInfoAfter.LiquidityGross.IsZero() && test.param.tickIndex <= pool.GetCurrentTick().Int64() {
				tickInfoAfter, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
				s.Require().NoError(err)
				s.Require().Equal(tickInfoAfter.FeeGrowthOppositeDirectionOfLastTraversal, feeAccum.GetValue())
			}
			if test.expectedErr != nil {
				s.Require().ErrorIs(err, test.expectedErr)
				return
			}
			s.Require().NoError(err)

			// Get the tick info for poolId 1 again
			tickInfoAfter, err = s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
			s.Require().NoError(err)

			// Check that the initialized or updated tick matches our expectation
			s.Require().Equal(test.expectedLiquidityNet, tickInfoAfter.LiquidityNet)
			s.Require().Equal(test.expectedLiquidityGross, tickInfoAfter.LiquidityGross)

			if test.param.tickIndex <= 0 {
				s.Require().Equal(defaultAccumCoins, tickInfoAfter.FeeGrowthOppositeDirectionOfLastTraversal)
			} else {
				s.Require().Equal(sdk.DecCoins(nil), tickInfoAfter.FeeGrowthOppositeDirectionOfLastTraversal)
			}

			// Ensure that at least the minimum amount of gas was charged
			gasConsumed := s.Ctx.GasMeter().GasConsumed() - existingGasConsumed
			s.Require().True(gasConsumed >= test.minimumGasConsumed)
		})
	}
}

func (s *KeeperTestSuite) TestGetTickInfo() {
	var (
		preInitializedTickIndex = DefaultCurrTick.Int64() + 2
		expectedUptimes         = getExpectedUptimes()
		emptyUptimeTrackers     = wrapUptimeTrackers(expectedUptimes.emptyExpectedAccumValues)
		varyingTokensAndDenoms  = wrapUptimeTrackers(expectedUptimes.varyingTokensMultiDenom)
	)

	tests := []struct {
		name                     string
		poolToGet                uint64
		tickToGet                int64
		preInitUptimeAccumValues []sdk.DecCoins
		expectedTickInfo         model.TickInfo
		expectedErr              error
	}{
		{
			name:      "Get tick info on existing pool and existing tick",
			poolToGet: validPoolId,
			tickToGet: preInitializedTickIndex,
			// Note that FeeGrowthOutside and UptimeGrowthOutside(s) are not updated.
			expectedTickInfo: model.TickInfo{LiquidityGross: DefaultLiquidityAmt, LiquidityNet: DefaultLiquidityAmt.Neg(), UptimeTrackers: emptyUptimeTrackers},
		},
		{
			name:                     "Get tick info on existing pool and existing tick with init but zero global uptime accums",
			poolToGet:                validPoolId,
			tickToGet:                preInitializedTickIndex,
			preInitUptimeAccumValues: expectedUptimes.varyingTokensMultiDenom,
			// Note that neither FeeGrowthOutside nor UptimeGrowthOutsides are updated.
			// We expect uptime trackers to be initialized to zero since tick > active tick
			expectedTickInfo: model.TickInfo{LiquidityGross: DefaultLiquidityAmt, LiquidityNet: DefaultLiquidityAmt.Neg(), UptimeTrackers: emptyUptimeTrackers},
		},
		{
			name:                     "Get tick info on existing pool and existing tick with nonzero global uptime accums",
			poolToGet:                validPoolId,
			tickToGet:                preInitializedTickIndex - 3,
			preInitUptimeAccumValues: expectedUptimes.varyingTokensMultiDenom,
			// Note that both FeeGrowthOutside and UptimeGrowthOutsides are updated.
			// We expect uptime trackers to be initialized to global accums since tick <= active tick
			expectedTickInfo: model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), FeeGrowthOppositeDirectionOfLastTraversal: sdk.NewDecCoins(oneEth), UptimeTrackers: varyingTokensAndDenoms},
		},
		{
			name:                     "Get tick info for active tick on existing pool with existing tick",
			poolToGet:                validPoolId,
			tickToGet:                DefaultCurrTick.Int64(),
			preInitUptimeAccumValues: expectedUptimes.varyingTokensMultiDenom,
			// Both fee growth and uptime trackers are set to global since tickToGet <= current tick
			expectedTickInfo: model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), FeeGrowthOppositeDirectionOfLastTraversal: sdk.NewDecCoins(oneEth), UptimeTrackers: varyingTokensAndDenoms},
		},
		{
			name:      "Get tick info on existing pool with no existing tick (cur pool tick > tick)",
			poolToGet: validPoolId,
			tickToGet: DefaultCurrTick.Int64() + 1,
			// Note that FeeGrowthOutside and UptimeGrowthOutside(s) are not initialized.
			expectedTickInfo: model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), UptimeTrackers: emptyUptimeTrackers},
		},
		{
			name:      "Get tick info on existing pool with no existing tick (cur pool tick == tick), initialized fee growth outside",
			poolToGet: validPoolId,
			tickToGet: DefaultCurrTick.Int64(),
			// Note that FeeGrowthOutside and UptimeGrowthOutside(s) are initialized.
			expectedTickInfo: model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), FeeGrowthOppositeDirectionOfLastTraversal: sdk.NewDecCoins(oneEth), UptimeTrackers: emptyUptimeTrackers},
		},
		{
			name:        "Get tick info on a non-existing pool with no existing tick",
			poolToGet:   2,
			tickToGet:   DefaultCurrTick.Int64() + 1,
			expectedErr: types.PoolNotFoundError{PoolId: 2},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()

			// Create a default CL pool
			clPool := s.PrepareConcentratedPool()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			if test.preInitUptimeAccumValues != nil {
				err := addToUptimeAccums(s.Ctx, clPool.GetId(), clKeeper, test.preInitUptimeAccumValues)
				s.Require().NoError(err)
			}

			// Set up an initialized tick
			err := clKeeper.InitOrUpdateTick(s.Ctx, validPoolId, DefaultCurrTick.Int64(), preInitializedTickIndex, DefaultLiquidityAmt, true)
			s.Require().NoError(err)

			// Charge fee to make sure that the global fee accumulator is always updated.
			// This is to test that the per-tick fee growth accumulator gets initialized.
			if test.poolToGet == validPoolId {
				s.SetupDefaultPosition(test.poolToGet)
			}
			err = clKeeper.ChargeFee(s.Ctx, validPoolId, oneEth)
			s.Require().NoError(err)

			// System under test
			tickInfo, err := clKeeper.GetTickInfo(s.Ctx, test.poolToGet, test.tickToGet)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &test.expectedErr)
				s.Require().Equal(model.TickInfo{}, tickInfo)
			} else {
				s.Require().NoError(err)
				clPool, err = clKeeper.GetPoolById(s.Ctx, validPoolId)
				s.Require().NoError(err)
				s.Require().Equal(test.expectedTickInfo, tickInfo)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCrossTick() {
	var (
		preInitializedTickIndex = DefaultCurrTick.Int64() - 2
		expectedUptimes         = getExpectedUptimes()
		emptyUptimeTrackers     = wrapUptimeTrackers(expectedUptimes.emptyExpectedAccumValues)
		defaultAdditiveFee      = sdk.NewDecCoinFromDec(USDC, sdk.NewDec(1000))
	)

	tests := []struct {
		name                                                  string
		poolToGet                                             uint64
		preInitializedTickIndex                               int64
		tickToGet                                             int64
		initGlobalUptimeAccumValues                           []sdk.DecCoins
		globalUptimeAccumDelta                                []sdk.DecCoins
		expectedUptimeTrackers                                []model.UptimeTracker
		additiveFee                                           sdk.DecCoin
		expectedLiquidityDelta                                sdk.Dec
		expectedTickFeeGrowthOppositeDirectionOfLastTraversal sdk.DecCoins
		expectedErr                                           bool
	}{
		{
			name:                    "Get tick info of existing tick below current tick (nonzero uptime trackers)",
			poolToGet:               validPoolId,
			preInitializedTickIndex: preInitializedTickIndex,
			tickToGet:               preInitializedTickIndex,
			additiveFee:             defaultAdditiveFee,
			// Global uptime accums remain unchanged after tick init
			initGlobalUptimeAccumValues: expectedUptimes.twoHundredTokensMultiDenom,
			globalUptimeAccumDelta:      expectedUptimes.hundredTokensMultiDenom,
			// We expect new uptime trackers to be new global - init global
			// This is because we init them to twoHundredTokensMultiDenom and then add hundredTokensMultiDenom,
			// so when we cross the tick and "flip" it, we expect threeHundredTokensMultiDenom - twoHundredTokensMultiDenom
			expectedUptimeTrackers:                                wrapUptimeTrackers(expectedUptimes.hundredTokensMultiDenom),
			expectedLiquidityDelta:                                DefaultLiquidityAmt.Neg(),
			expectedTickFeeGrowthOppositeDirectionOfLastTraversal: DefaultFeeAccumCoins.Add(defaultAdditiveFee),
		},
		{
			name:                    "Get tick info of existing tick below current tick (nil uptime trackers)",
			poolToGet:               validPoolId,
			preInitializedTickIndex: preInitializedTickIndex,
			tickToGet:               preInitializedTickIndex,
			additiveFee:             defaultAdditiveFee,
			expectedUptimeTrackers:  emptyUptimeTrackers,
			expectedLiquidityDelta:  DefaultLiquidityAmt.Neg(),
			expectedTickFeeGrowthOppositeDirectionOfLastTraversal: DefaultFeeAccumCoins.Add(defaultAdditiveFee),
		},
		{
			name:                    "Get tick info of an existing tick above current tick (nonzero uptime trackers)",
			poolToGet:               validPoolId,
			preInitializedTickIndex: DefaultCurrTick.Int64() + 1,
			tickToGet:               DefaultCurrTick.Int64() + 1,
			additiveFee:             defaultAdditiveFee,
			// Global uptime accums remain unchanged after tick init
			initGlobalUptimeAccumValues: expectedUptimes.twoHundredTokensMultiDenom,
			globalUptimeAccumDelta:      expectedUptimes.hundredTokensMultiDenom,
			// We expect new uptime trackers to be equal to new global
			// This is because we init them to zero (since target tick is above current tick),
			// so when we cross the tick and "flip" it, we expect it to be the global value - 0 = global value.
			expectedUptimeTrackers:                                wrapUptimeTrackers(expectedUptimes.threeHundredTokensMultiDenom),
			expectedLiquidityDelta:                                DefaultLiquidityAmt.Neg(),
			expectedTickFeeGrowthOppositeDirectionOfLastTraversal: DefaultFeeAccumCoins.Add(defaultAdditiveFee).Add(DefaultFeeAccumCoins...),
		},
		{
			name:                    "Get tick info of new tick with a separate existing tick below current tick (nonzero uptime trackers)",
			poolToGet:               validPoolId,
			preInitializedTickIndex: preInitializedTickIndex,
			tickToGet:               DefaultCurrTick.Int64() + 1,
			additiveFee:             defaultAdditiveFee,
			// Global uptime accums remain unchanged after tick init
			initGlobalUptimeAccumValues: expectedUptimes.twoHundredTokensMultiDenom,
			globalUptimeAccumDelta:      expectedUptimes.hundredTokensMultiDenom,
			// We expect new uptime trackers to be equal to new global
			// This is because we init them to zero (since target tick is above current tick),
			// so when we cross the tick and "flip" it, we expect it to be the global value - 0 = global value.
			expectedUptimeTrackers:                                wrapUptimeTrackers(expectedUptimes.threeHundredTokensMultiDenom),
			expectedLiquidityDelta:                                sdk.ZeroDec(),
			expectedTickFeeGrowthOppositeDirectionOfLastTraversal: DefaultFeeAccumCoins.Add(defaultAdditiveFee).Add(DefaultFeeAccumCoins...),
		},
		{
			// Note that this test case covers technically undefined behavior (crossing into the current tick).
			name:                    "Get tick info of existing tick at current tick (nonzero uptime trackers)",
			poolToGet:               validPoolId,
			preInitializedTickIndex: DefaultCurrTick.Int64(),
			tickToGet:               DefaultCurrTick.Int64(),
			additiveFee:             defaultAdditiveFee,
			// Global uptime accums remain unchanged after tick init
			initGlobalUptimeAccumValues: expectedUptimes.twoHundredTokensMultiDenom,
			globalUptimeAccumDelta:      expectedUptimes.hundredTokensMultiDenom,
			// We expect new uptime trackers to be new global - init global
			// This is because we init them to twoHundredTokensMultiDenom and then add hundredTokensMultiDenom,
			// so when we cross the tick and "flip" it, we expect threeHundredTokensMultiDenom - twoHundredTokensMultiDenom
			expectedUptimeTrackers:                                wrapUptimeTrackers(expectedUptimes.hundredTokensMultiDenom),
			expectedLiquidityDelta:                                DefaultLiquidityAmt.Neg(),
			expectedTickFeeGrowthOppositeDirectionOfLastTraversal: DefaultFeeAccumCoins.Add(defaultAdditiveFee),
		},
		{
			name:                    "Twice the default additive fee",
			poolToGet:               validPoolId,
			preInitializedTickIndex: preInitializedTickIndex,
			tickToGet:               preInitializedTickIndex,
			additiveFee:             defaultAdditiveFee.Add(defaultAdditiveFee),
			expectedUptimeTrackers:  emptyUptimeTrackers,
			expectedLiquidityDelta:  DefaultLiquidityAmt.Neg(),
			expectedTickFeeGrowthOppositeDirectionOfLastTraversal: DefaultFeeAccumCoins.Add(defaultAdditiveFee.Add(defaultAdditiveFee)),
		},
		{
			name:                    "Try invalid tick",
			poolToGet:               2,
			preInitializedTickIndex: preInitializedTickIndex,
			tickToGet:               preInitializedTickIndex,
			additiveFee:             defaultAdditiveFee,
			expectedErr:             true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()

			// Create a default CL pool
			clPool := s.PrepareConcentratedPool()
			clPool.SetCurrentTick(DefaultCurrTick)

			if test.poolToGet == validPoolId {
				s.FundAcc(s.TestAccs[0], sdk.NewCoins(DefaultCoin0, DefaultCoin1))
				_, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, test.poolToGet, s.TestAccs[0], DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}

			// Charge fee to make sure that the global fee accumulator is always updated.
			// This is to test that the per-tick fee growth accumulator gets initialized.
			defaultAccumCoins := sdk.NewDecCoin("foo", sdk.NewInt(50))
			err := s.App.ConcentratedLiquidityKeeper.ChargeFee(s.Ctx, validPoolId, defaultAccumCoins)
			s.Require().NoError(err)

			// Initialize global uptime accums
			if test.initGlobalUptimeAccumValues != nil {
				err = addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, test.initGlobalUptimeAccumValues)
				s.Require().NoError(err)
			}

			// Set up an initialized tick
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, validPoolId, DefaultCurrTick.Int64(), test.preInitializedTickIndex, DefaultLiquidityAmt, true)
			s.Require().NoError(err)

			// Update global uptime accums for edge case testing
			if test.globalUptimeAccumDelta != nil {
				err = addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, test.globalUptimeAccumDelta)
				s.Require().NoError(err)
			}

			// update the fee accumulator so that we have accum value > tick fee growth value
			// now we have 100 foo coins inside the pool accumulator
			err = s.App.ConcentratedLiquidityKeeper.ChargeFee(s.Ctx, validPoolId, defaultAccumCoins)
			s.Require().NoError(err)

			// System under test
			liquidityDelta, err := s.App.ConcentratedLiquidityKeeper.CrossTick(s.Ctx, test.poolToGet, test.tickToGet, test.additiveFee)
			if test.expectedErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(test.expectedLiquidityDelta, liquidityDelta)

				// now check if fee accumulator has been properly updated
				accum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(s.Ctx, test.poolToGet)
				s.Require().NoError(err)

				// accum value should not have changed
				s.Require().Equal(accum.GetValue(), sdk.NewDecCoins(defaultAccumCoins).MulDec(sdk.NewDec(2)))

				// check if the tick fee growth outside has been correctly subtracted
				tickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, test.poolToGet, test.tickToGet)
				s.Require().NoError(err)
				s.Require().Equal(test.expectedTickFeeGrowthOppositeDirectionOfLastTraversal, tickInfo.FeeGrowthOppositeDirectionOfLastTraversal)

				// ensure tick being entered has properly updated uptime trackers
				s.Require().Equal(test.expectedUptimeTrackers, tickInfo.UptimeTrackers)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetTickLiquidityForFullRange() {
	defaultTick := withPoolId(defaultTick, defaultPoolId)

	tests := []struct {
		name        string
		presetTicks []genesis.FullTick

		expectedLiquidityDepthForRange []queryproto.LiquidityDepthWithRange
	}{
		{
			name: "one full range position, testing range in between",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},
			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       sdk.NewInt(DefaultMinTick),
					UpperTick:       sdk.NewInt(DefaultMaxTick),
				},
			},
		},
		{
			name: "one ranged position, testing range with greater range than initialized ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, 5, sdk.NewDec(-10)),
			},
			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       sdk.NewInt(DefaultMinTick),
					UpperTick:       sdk.NewInt(5),
				},
			},
		},
		//  	   	10 ----------------- 30
		//  -20 ------------- 20
		{
			name: "two ranged positions, testing overlapping positions",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -20, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, 20, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(50)),
				withLiquidityNetandTickIndex(defaultTick, 30, sdk.NewDec(-50)),
			},
			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       sdk.NewInt(-20),
					UpperTick:       sdk.NewInt(10),
				},
				{
					LiquidityAmount: sdk.NewDec(60),
					LowerTick:       sdk.NewInt(10),
					UpperTick:       sdk.NewInt(20),
				},
				{
					LiquidityAmount: sdk.NewDec(50),
					LowerTick:       sdk.NewInt(20),
					UpperTick:       sdk.NewInt(30),
				},
			},
		},
		//  	   	       10 ----------------- 30
		//  min tick --------------------------------------max tick
		{
			name: "one full ranged position, one narrow position",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(50)),
				withLiquidityNetandTickIndex(defaultTick, 30, sdk.NewDec(-50)),
			},
			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       sdk.NewInt(DefaultMinTick),
					UpperTick:       sdk.NewInt(10),
				},
				{
					LiquidityAmount: sdk.NewDec(60),
					LowerTick:       sdk.NewInt(10),
					UpperTick:       sdk.NewInt(30),
				},
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       sdk.NewInt(30),
					UpperTick:       sdk.NewInt(DefaultMaxTick),
				},
			},
		},
		//              11--13
		//         10 ----------------- 30
		//  -20 ------------- 20
		{
			name: "three ranged positions, testing overlapping positions",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -20, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, 20, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(50)),
				withLiquidityNetandTickIndex(defaultTick, 30, sdk.NewDec(-50)),
				withLiquidityNetandTickIndex(defaultTick, 11, sdk.NewDec(100)),
				withLiquidityNetandTickIndex(defaultTick, 13, sdk.NewDec(-100)),
			},
			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       sdk.NewInt(-20),
					UpperTick:       sdk.NewInt(10),
				},
				{
					LiquidityAmount: sdk.NewDec(60),
					LowerTick:       sdk.NewInt(10),
					UpperTick:       sdk.NewInt(11),
				},
				{
					LiquidityAmount: sdk.NewDec(160),
					LowerTick:       sdk.NewInt(11),
					UpperTick:       sdk.NewInt(13),
				},
				{
					LiquidityAmount: sdk.NewDec(60),
					LowerTick:       sdk.NewInt(13),
					UpperTick:       sdk.NewInt(20),
				},
				{
					LiquidityAmount: sdk.NewDec(50),
					LowerTick:       sdk.NewInt(20),
					UpperTick:       sdk.NewInt(30),
				},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()

			// Create a default CL pool
			s.PrepareConcentratedPool()
			for _, tick := range test.presetTicks {
				s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, tick.PoolId, tick.TickIndex, tick.Info)
			}

			liquidityForRange, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityForFullRange(s.Ctx, defaultPoolId)
			s.Require().NoError(err)
			s.Require().Equal(liquidityForRange, test.expectedLiquidityDepthForRange)
		})
	}
}

func (s *KeeperTestSuite) TestGetTickLiquidityNetInDirection() {
	defaultTick := withPoolId(defaultTick, defaultPoolId)

	tests := []struct {
		name        string
		presetTicks []genesis.FullTick

		// testing params
		poolId          uint64
		tokenIn         string
		currentPoolTick sdk.Int
		startTick       sdk.Int
		boundTick       sdk.Int

		// expected values
		expectedLiquidityDepths []queryproto.TickLiquidityNet
		expectedError           bool
	}{
		{
			name: "one full range position, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(10),
					TickIndex:    sdk.NewInt(DefaultMinTick),
				},
			},
		},
		{
			name: "one full range position, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(-10),
					TickIndex:    sdk.NewInt(DefaultMaxTick),
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, 5, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(10),
					TickIndex:    sdk.NewInt(DefaultMinTick),
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, 5, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(20),
					TickIndex:    sdk.NewInt(5),
				},
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    sdk.NewInt(10),
				},
				{
					LiquidityNet: sdk.NewDec(-10),
					TickIndex:    sdk.NewInt(DefaultMaxTick),
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false, bound tick below with non-empty ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: sdk.NewInt(-15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(20),
					TickIndex:    sdk.NewInt(-10),
				},
			},
		},
		{
			name: "one ranged position, returned empty array",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:                  defaultPoolId,
			tokenIn:                 ETH,
			boundTick:               sdk.NewInt(-5),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false, bound tick below with non-empty ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: sdk.NewInt(10),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    sdk.NewInt(10),
				},
			},
		},
		{
			name: "one full range position, two ranged positions, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, -5, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 5, sdk.NewDec(-20)),
				withLiquidityNetandTickIndex(defaultTick, 2, sdk.NewDec(40)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-40)),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(20),
					TickIndex:    sdk.NewInt(-5),
				},
				{
					LiquidityNet: sdk.NewDec(10),
					TickIndex:    sdk.NewInt(DefaultMinTick),
				},
			},
		},
		{
			name: "one full range position, two ranged positions, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, -5, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 5, sdk.NewDec(-20)),
				withLiquidityNetandTickIndex(defaultTick, 2, sdk.NewDec(40)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-40)),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(40),
					TickIndex:    sdk.NewInt(2),
				},
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    sdk.NewInt(5),
				},
				{
					LiquidityNet: sdk.NewDec(-40),
					TickIndex:    sdk.NewInt(10),
				},
				{
					LiquidityNet: sdk.NewDec(-10),
					TickIndex:    sdk.NewInt(DefaultMaxTick),
				},
			},
		},
		{
			name: "current pool tick == start tick, zero for one",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:          defaultPoolId,
			tokenIn:         ETH,
			currentPoolTick: sdk.NewInt(10),
			startTick:       sdk.NewInt(10),
			boundTick:       sdk.NewInt(-15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(20),
					TickIndex:    sdk.NewInt(-10),
				},
			},
		},
		{
			name: "current pool tick != start tick, zero for one",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:          defaultPoolId,
			tokenIn:         ETH,
			currentPoolTick: sdk.NewInt(21),
			startTick:       sdk.NewInt(10),
			boundTick:       sdk.NewInt(-15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(20),
					TickIndex:    sdk.NewInt(-10),
				},
			},
		},
		{
			name: "11: current pool tick == start tick, one for zero",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:          defaultPoolId,
			tokenIn:         USDC,
			currentPoolTick: sdk.NewInt(5),
			startTick:       sdk.NewInt(5),
			boundTick:       sdk.NewInt(15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    sdk.NewInt(10),
				},
			},
		},
		{
			name: "current pool tick != start tick, one for zero",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:          defaultPoolId,
			tokenIn:         USDC,
			currentPoolTick: sdk.NewInt(-50),
			startTick:       sdk.NewInt(5),
			boundTick:       sdk.NewInt(15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    sdk.NewInt(10),
				},
			},
		},

		// error cases
		{
			name: "error: invalid pool id",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:        5,
			tokenIn:       "invalid_token",
			boundTick:     sdk.NewInt(-5),
			expectedError: true,
		},
		{
			name: "error: invalid token in",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:        defaultPoolId,
			tokenIn:       "invalid_token",
			boundTick:     sdk.NewInt(-5),
			expectedError: true,
		},
		{
			name: "error: wrong direction of bound ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:        defaultPoolId,
			tokenIn:       USDC,
			boundTick:     sdk.NewInt(-5),
			expectedError: true,
		},
		{
			name: "error: bound tick is greater than max tick",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:        defaultPoolId,
			tokenIn:       USDC,
			boundTick:     sdk.NewInt(DefaultMaxTick + 1),
			expectedError: true,
		},
		{
			name: "error: bound tick is greater than min tick",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:        defaultPoolId,
			tokenIn:       ETH,
			boundTick:     sdk.NewInt(DefaultMinTick - 1),
			expectedError: true,
		},
		{
			name: "start tick is in invalid range relative to current pool tick, zero for one",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:          defaultPoolId,
			tokenIn:         ETH,
			currentPoolTick: sdk.NewInt(10),
			startTick:       sdk.NewInt(21),
			boundTick:       sdk.NewInt(-15),
			expectedError:   true,
		},
		{
			name: "start tick is in invalid range relative to current pool tick, one for zero",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:          defaultPoolId,
			tokenIn:         USDC,
			currentPoolTick: sdk.NewInt(5),
			startTick:       sdk.NewInt(-50),
			boundTick:       sdk.NewInt(15),
			expectedError:   true,
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()

			// Create a default CL pool
			pool := s.PrepareConcentratedPool()
			for _, tick := range test.presetTicks {
				s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, tick.PoolId, tick.TickIndex, tick.Info)
			}

			// Force initialize current sqrt price to 1.
			// Normally, initialized during position creation.
			// We only initialize ticks in this test for simplicity.
			curPrice := sdk.OneDec()
			// TODO: consider adding tests for GetTickLiquidityNetInDirection
			// with tick spacing > 1, requiring price to tick conversion with rounding.
			curTick, err := math.PriceToTick(curPrice)
			s.Require().NoError(err)
			if !test.currentPoolTick.IsNil() {
				sqrtPrice, err := math.TickToSqrtPrice(test.currentPoolTick)
				s.Require().NoError(err)

				curTick = test.currentPoolTick
				curPrice = sqrtPrice
			}
			pool.SetCurrentSqrtPrice(curPrice)
			pool.SetCurrentTick(curTick)

			err = s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
			s.Require().NoError(err)

			// system under test
			liquidityForRange, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityNetInDirection(s.Ctx, test.poolId, test.tokenIn, test.startTick, test.boundTick)
			if test.expectedError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(liquidityForRange, test.expectedLiquidityDepths)
		})
	}
}

func (s *KeeperTestSuite) TestValidateTickRangeIsValid() {
	// use 2 as default tick spacing
	defaultTickSpacing := uint64(2)

	tests := []struct {
		name          string
		lowerTick     int64
		upperTick     int64
		tickSpacing   uint64
		expectedError error
	}{
		{
			name:      "happy path with default tick spacing",
			lowerTick: 2,
			upperTick: 4,
		},
		{
			name:        "happy path with non default tick spacing",
			tickSpacing: 3,
			lowerTick:   3,
			upperTick:   6,
		},
		{
			name:          "lower tick is not divisible by default tick spacing",
			lowerTick:     3,
			upperTick:     2,
			expectedError: types.TickSpacingError{LowerTick: 3, UpperTick: 2, TickSpacing: defaultTickSpacing},
		},
		{
			name:          "upper tick is not divisible by default tick spacing",
			lowerTick:     2,
			upperTick:     3,
			expectedError: types.TickSpacingError{LowerTick: 2, UpperTick: 3, TickSpacing: defaultTickSpacing},
		},
		{
			name:          "lower tick is not divisible by tick spacing",
			lowerTick:     4,
			upperTick:     3,
			tickSpacing:   3,
			expectedError: types.TickSpacingError{LowerTick: 4, UpperTick: 3, TickSpacing: 3},
		},
		{
			name:          "upper tick is not divisible by tick spacing",
			lowerTick:     3,
			upperTick:     4,
			tickSpacing:   3,
			expectedError: types.TickSpacingError{LowerTick: 3, UpperTick: 4, TickSpacing: 3},
		},
		{
			name:          "lower tick is smaller than min tick",
			lowerTick:     DefaultMinTick - 2,
			upperTick:     2,
			expectedError: types.InvalidTickError{Tick: DefaultMinTick - 2, IsLower: true, MinTick: DefaultMinTick, MaxTick: DefaultMaxTick},
		},
		{
			name:          "lower tick is greater than max tick",
			lowerTick:     DefaultMaxTick + 2,
			upperTick:     DefaultMaxTick + 4,
			expectedError: types.InvalidTickError{Tick: DefaultMaxTick + 2, IsLower: true, MinTick: DefaultMinTick, MaxTick: DefaultMaxTick},
		},
		{
			name:          "upper tick is smaller than min tick",
			lowerTick:     2,
			upperTick:     DefaultMinTick - 2,
			expectedError: types.InvalidTickError{Tick: DefaultMinTick - 2, IsLower: false, MinTick: DefaultMinTick, MaxTick: DefaultMaxTick},
		},
		{
			name:          "upper tick is greater than max tick",
			lowerTick:     2,
			upperTick:     DefaultMaxTick + 2,
			expectedError: types.InvalidTickError{Tick: DefaultMaxTick + 2, IsLower: false, MinTick: DefaultMinTick, MaxTick: DefaultMaxTick},
		},
		{
			name:      "lower tick is greater than upper tick",
			lowerTick: 2,
			upperTick: 0,

			expectedError: types.InvalidLowerUpperTickError{LowerTick: 2, UpperTick: 0},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			tickSpacing := defaultTickSpacing
			if test.tickSpacing != uint64(0) {
				tickSpacing = test.tickSpacing
			}

			// System Under Test
			err := cl.ValidateTickRangeIsValid(tickSpacing, test.lowerTick, test.upperTick)

			if test.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedError.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetAllInitializedTicksForPool() {
	const (
		// chosen randomly
		defaultPoolId = 676
	)

	defaultTick := withPoolId(defaultTick, defaultPoolId)

	tests := []struct {
		name                   string
		preSetTicks            []genesis.FullTick
		expectedTicksOverwrite []genesis.FullTick
		expectedError          error
	}{
		{
			name:        "one positive tick per pool",
			preSetTicks: []genesis.FullTick{defaultTick},
		},
		{
			name:        "one negative tick per pool",
			preSetTicks: []genesis.FullTick{withTickIndex(defaultTick, -1)},
		},
		{
			name:        "one zero tick per pool",
			preSetTicks: []genesis.FullTick{withTickIndex(defaultTick, 0)},
		},
		{
			name: "multiple ticks per pool",
			preSetTicks: []genesis.FullTick{
				defaultTick,
				withTickIndex(defaultTick, -1),
				withTickIndex(defaultTick, 0),
				withTickIndex(defaultTick, -200),
				withTickIndex(defaultTick, 1000),
				withTickIndex(defaultTick, -999),
			},
			expectedTicksOverwrite: []genesis.FullTick{
				withTickIndex(defaultTick, -999),
				withTickIndex(defaultTick, -200),
				withTickIndex(defaultTick, -1),
				withTickIndex(defaultTick, 0),
				defaultTick,
				withTickIndex(defaultTick, 1000),
			},
		},
		{
			name: "multiple ticks per multiple pools",
			preSetTicks: []genesis.FullTick{
				defaultTick,
				withTickIndex(defaultTick, -1),
				withPoolId(withTickIndex(defaultTick, 0), 3),
				withTickIndex(defaultTick, -200),
				withTickIndex(defaultTick, 1000),
				withTickIndex(defaultTick, -999),
				withPoolId(withTickIndex(defaultTick, -4), 90),
				withTickIndex(defaultTick, 33),
				withPoolId(withTickIndex(defaultTick, 44), 1200),
				withPoolId(withTickIndex(defaultTick, -1000), 3),
				withTickIndex(defaultTick, -1234),
				withPoolId(withTickIndex(defaultTick, 1000), 3), // duplicate for another pool.
			},
			expectedTicksOverwrite: []genesis.FullTick{
				withTickIndex(defaultTick, -1234),
				withTickIndex(defaultTick, -999),
				withTickIndex(defaultTick, -200),
				withTickIndex(defaultTick, -1),
				defaultTick,
				withTickIndex(defaultTick, 33),
				withTickIndex(defaultTick, 1000),
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			for _, tick := range test.preSetTicks {
				s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, tick.PoolId, tick.TickIndex, tick.Info)
			}

			// If overwrite is not specified, we expect the pre-set ticks to be returned.
			expectedTicks := test.preSetTicks
			if len(test.expectedTicksOverwrite) > 0 {
				expectedTicks = test.expectedTicksOverwrite
			}

			// System Under Test
			ticks, err := s.App.ConcentratedLiquidityKeeper.GetAllInitializedTicksForPool(s.Ctx, defaultPoolId)

			s.Require().NoError(err)

			s.Require().Equal(len(expectedTicks), len(ticks))
			for i, expectedTick := range expectedTicks {
				s.Require().Equal(expectedTick, ticks[i], "expected tick %d to be %v, got %v", i, expectedTick, ticks[i])
			}
		})
	}
}
