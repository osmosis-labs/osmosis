package concentrated_liquidity_test

import (
	"errors"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types/genesis"
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

func withLiquidityNetandTickIndex(tick genesis.FullTick, tickIndex int64, liquidityNet osmomath.Dec) genesis.FullTick {
	tick.TickIndex = tickIndex
	tick.Info.LiquidityNet = liquidityNet

	return tick
}

func (s *KeeperTestSuite) TestTickOrdering() {
	s.SetupTest()

	storeKey := storetypes.NewKVStoreKey("concentrated_liquidity")
	tKey := storetypes.NewTransientStoreKey("transient_test")
	s.Ctx = testutil.DefaultContext(storeKey, tKey)
	s.App.ConcentratedLiquidityKeeper = cl.NewKeeper(s.App.AppCodec(), storeKey, s.App.AccountKeeper, s.App.BankKeeper, s.App.GAMMKeeper, s.App.PoolIncentivesKeeper, s.App.IncentivesKeeper, s.App.LockupKeeper, s.App.DistrKeeper, s.App.ContractKeeper, s.App.GetSubspace(types.ModuleName))

	liquidityTicks := []int64{-200, -55, -4, 70, 78, 84, 139, 240, 535}
	for _, t := range liquidityTicks {
		s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, 1, t, &model.TickInfo{})
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
		poolId           uint64
		tickIndex        int64
		liquidityIn      osmomath.Dec
		initLiquidityNet bool
		upper            bool
	}

	tests := []struct {
		name                   string
		param                  param
		tickExists             bool
		expectedLiquidityNet   osmomath.Dec
		expectedLiquidityGross osmomath.Dec
		minimumGasConsumed     bool
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
			minimumGasConsumed:     true,
		},
		{
			name: "Init tick 50 with DefaultLiquidityAmt liquidity, upper, only initialize liquidity net",
			param: param{
				poolId:           validPoolId,
				tickIndex:        50,
				liquidityIn:      DefaultLiquidityAmt,
				upper:            true,
				initLiquidityNet: true,
			},
			tickExists:             false,
			expectedLiquidityNet:   osmomath.ZeroDec(),
			expectedLiquidityGross: DefaultLiquidityAmt,
			minimumGasConsumed:     false,
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
			minimumGasConsumed:     true,
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
			expectedLiquidityNet:   DefaultLiquidityAmt.Mul(osmomath.NewDec(2)).Neg(),
			expectedLiquidityGross: DefaultLiquidityAmt.Mul(osmomath.NewDec(2)),
			minimumGasConsumed:     false,
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
			expectedLiquidityNet:   DefaultLiquidityAmt.Mul(osmomath.NewDec(2)),
			expectedLiquidityGross: DefaultLiquidityAmt.Mul(osmomath.NewDec(2)),
			minimumGasConsumed:     false,
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
			minimumGasConsumed:     true,
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
			minimumGasConsumed:     true,
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
			expectedLiquidityNet:   DefaultLiquidityAmt.Mul(osmomath.NewDec(2)).Neg(),
			expectedLiquidityGross: DefaultLiquidityAmt.Mul(osmomath.NewDec(2)),
			minimumGasConsumed:     false,
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
			expectedLiquidityNet:   DefaultLiquidityAmt.Mul(osmomath.NewDec(2)),
			expectedLiquidityGross: DefaultLiquidityAmt.Mul(osmomath.NewDec(2)),
			minimumGasConsumed:     false,
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
			minimumGasConsumed:     true,
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
			expectedLiquidityNet:   osmomath.ZeroDec(),
			expectedLiquidityGross: osmomath.ZeroDec(),
			minimumGasConsumed:     false,
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
			expectedLiquidityNet:   osmomath.ZeroDec(),
			expectedLiquidityGross: osmomath.ZeroDec(),
			minimumGasConsumed:     false,
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

			_, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, 1)
			s.Require().NoError(err)
			spreadFactorAccum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, 1)
			s.Require().NoError(err)

			// manually update accumulator for testing
			defaultAccumCoins := sdk.NewDecCoins(sdk.NewDecCoin("foo", osmomath.NewInt(50)))
			spreadFactorAccum.AddToAccumulator(defaultAccumCoins)

			// If tickExists set, initialize the specified tick with defaultLiquidityAmt
			preexistingLiquidity := osmomath.ZeroDec()
			if test.tickExists {
				tickInfoBefore, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
				s.Require().NoError(err)
				tickIsEmpty, err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, test.param.poolId, test.param.tickIndex, DefaultLiquidityAmt, test.param.upper)
				s.Require().False(tickIsEmpty)
				s.Require().NoError(err)
				if tickInfoBefore.LiquidityGross.IsZero() && test.param.tickIndex <= pool.GetCurrentTick() {
					tickInfoAfter, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
					s.Require().NoError(err)
					s.Require().Equal(tickInfoAfter.SpreadRewardGrowthOppositeDirectionOfLastTraversal, spreadFactorAccum.GetValue())
				}
				preexistingLiquidity = DefaultLiquidityAmt
			}

			// if this param is set to true, we manually set the tick liquidity net value to default liquidity amount
			// for testing purpose.
			if test.param.initLiquidityNet {
				tickInfoBefore, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
				s.Require().NoError(err)

				tickInfoBefore.LiquidityNet = DefaultLiquidityAmt
				s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, 1, test.param.tickIndex, &tickInfoBefore)
			}

			// Get the tick info for poolId 1
			tickInfoAfter, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
			s.Require().NoError(err)

			// Ensure tick state contains any preexistingLiquidity (zero otherwise)
			s.Require().Equal(preexistingLiquidity, tickInfoAfter.LiquidityGross)

			existingGasConsumed := s.Ctx.GasMeter().GasConsumed()

			// System under test.
			// Initialize or update the tick according to the test case
			tickIsEmpty, err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, test.param.poolId, test.param.tickIndex, test.param.liquidityIn, test.param.upper)
			if tickInfoAfter.LiquidityGross.IsZero() && test.param.tickIndex <= pool.GetCurrentTick() {
				tickInfoAfter, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
				s.Require().NoError(err)
				s.Require().Equal(tickInfoAfter.SpreadRewardGrowthOppositeDirectionOfLastTraversal, spreadFactorAccum.GetValue())
			}
			if test.expectedErr != nil {
				s.Require().ErrorIs(err, test.expectedErr)
				return
			}
			s.Require().NoError(err)

			if test.expectedLiquidityGross.IsZero() && test.expectedLiquidityNet.IsZero() {
				s.Require().True(tickIsEmpty)
			} else {
				s.Require().False(tickIsEmpty)
			}

			// Get the tick info for poolId 1 again
			tickInfoAfter, err = s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
			s.Require().NoError(err)

			// Check that the initialized or updated tick matches our expectation
			s.Require().Equal(test.expectedLiquidityNet, tickInfoAfter.LiquidityNet)
			s.Require().Equal(test.expectedLiquidityGross, tickInfoAfter.LiquidityGross)

			if test.param.tickIndex <= 0 {
				s.Require().Equal(defaultAccumCoins, tickInfoAfter.SpreadRewardGrowthOppositeDirectionOfLastTraversal)
			} else {
				s.Require().Equal(sdk.DecCoins(nil), tickInfoAfter.SpreadRewardGrowthOppositeDirectionOfLastTraversal)
			}

			// Ensure that at least the minimum amount of gas was charged
			gasConsumed := s.Ctx.GasMeter().GasConsumed() - existingGasConsumed
			if test.minimumGasConsumed {
				s.Require().True(gasConsumed >= uint64(types.BaseGasFeeForInitializingTick))
			} else {
				s.Require().True(gasConsumed < uint64(types.BaseGasFeeForInitializingTick))
			}

		})
	}
}

func (s *KeeperTestSuite) TestGetTickInfo() {
	var (
		preInitializedTickIndex           = DefaultCurrTick + 2
		expectedUptimes                   = getExpectedUptimes()
		emptyUptimeTrackers               = wrapUptimeTrackers(expectedUptimes.emptyExpectedAccumValues)
		emptyUptimeTrackersModel          = model.UptimeTrackers{List: emptyUptimeTrackers}
		scaledVaryingTokensAndDenomsModel = model.UptimeTrackers{List: wrapUptimeTrackers(s.scaleUptimeAccumulators(expectedUptimes.varyingTokensMultiDenom))}
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
			// Note that SpreadRewardGrowthOutside and UptimeGrowthOutside(s) are not updated.
			expectedTickInfo: model.TickInfo{LiquidityGross: DefaultLiquidityAmt, LiquidityNet: DefaultLiquidityAmt.Neg(), UptimeTrackers: emptyUptimeTrackersModel},
		},
		{
			name:                     "Get tick info on existing pool and existing tick with init but zero global uptime accums",
			poolToGet:                validPoolId,
			tickToGet:                preInitializedTickIndex,
			preInitUptimeAccumValues: expectedUptimes.varyingTokensMultiDenom,
			// Note that neither SpreadRewardGrowthOutside nor UptimeGrowthOutsides are updated.
			// We expect uptime trackers to be initialized to zero since tick > active tick
			expectedTickInfo: model.TickInfo{LiquidityGross: DefaultLiquidityAmt, LiquidityNet: DefaultLiquidityAmt.Neg(), UptimeTrackers: emptyUptimeTrackersModel},
		},
		{
			name:                     "Get tick info on existing pool and existing tick with nonzero global uptime accums",
			poolToGet:                validPoolId,
			tickToGet:                preInitializedTickIndex - 3,
			preInitUptimeAccumValues: expectedUptimes.varyingTokensMultiDenom,
			// Note that both SpreadRewardGrowthOutside and UptimeGrowthOutsides are updated.
			// We expect uptime trackers to be initialized to global accums since tick <= active tick
			expectedTickInfo: model.TickInfo{LiquidityGross: osmomath.ZeroDec(), LiquidityNet: osmomath.ZeroDec(), SpreadRewardGrowthOppositeDirectionOfLastTraversal: sdk.NewDecCoins(oneEth), UptimeTrackers: scaledVaryingTokensAndDenomsModel},
		},
		{
			name:                     "Get tick info for active tick on existing pool with existing tick",
			poolToGet:                validPoolId,
			tickToGet:                DefaultCurrTick,
			preInitUptimeAccumValues: expectedUptimes.varyingTokensMultiDenom,
			// Both spread reward growth and uptime trackers are set to global since tickToGet <= current tick
			expectedTickInfo: model.TickInfo{LiquidityGross: osmomath.ZeroDec(), LiquidityNet: osmomath.ZeroDec(), SpreadRewardGrowthOppositeDirectionOfLastTraversal: sdk.NewDecCoins(oneEth), UptimeTrackers: scaledVaryingTokensAndDenomsModel},
		},
		{
			name:      "Get tick info on existing pool with no existing tick (cur pool tick > tick)",
			poolToGet: validPoolId,
			tickToGet: DefaultCurrTick + 1,
			// Note that SpreadRewardGrowthOutside and UptimeGrowthOutside(s) are not initialized.
			expectedTickInfo: model.TickInfo{LiquidityGross: osmomath.ZeroDec(), LiquidityNet: osmomath.ZeroDec(), UptimeTrackers: emptyUptimeTrackersModel},
		},
		{
			name:      "Get tick info on existing pool with no existing tick (cur pool tick == tick), initialized spread reward growth outside",
			poolToGet: validPoolId,
			tickToGet: DefaultCurrTick,
			// Note that SpreadRewardGrowthOutside and UptimeGrowthOutside(s) are initialized.
			expectedTickInfo: model.TickInfo{LiquidityGross: osmomath.ZeroDec(), LiquidityNet: osmomath.ZeroDec(), SpreadRewardGrowthOppositeDirectionOfLastTraversal: sdk.NewDecCoins(oneEth), UptimeTrackers: emptyUptimeTrackersModel},
		},
		{
			name:        "Get tick info on a non-existing pool with no existing tick",
			poolToGet:   2,
			tickToGet:   DefaultCurrTick + 1,
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

			// Upscale accum value
			test.expectedTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal = test.expectedTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal.MulDecTruncate(cl.PerUnitLiqScalingFactor)

			if test.preInitUptimeAccumValues != nil {
				err := addToUptimeAccums(s.Ctx, clPool.GetId(), clKeeper, test.preInitUptimeAccumValues)
				s.Require().NoError(err)
			}

			// Set up an initialized tick
			_, err := clKeeper.InitOrUpdateTick(s.Ctx, validPoolId, preInitializedTickIndex, DefaultLiquidityAmt, true)
			s.Require().NoError(err)

			// Charge spread factor to make sure that the global spread factor accumulator is always updated.
			// This is to test that the per-tick spread reward growth accumulator gets initialized.
			if test.poolToGet == validPoolId {
				s.SetupDefaultPosition(test.poolToGet)
			}
			s.AddToSpreadRewardAccumulator(validPoolId, oneEth)

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
		preInitializedTickIndex     = DefaultCurrTick - 2
		expectedUptimes             = getExpectedUptimes()
		emptyUptimeTrackers         = wrapUptimeTrackers(expectedUptimes.emptyExpectedAccumValues)
		defaultAdditiveSpreadFactor = sdk.NewDecCoinFromDec(USDC, osmomath.NewDec(1000).MulTruncate(cl.PerUnitLiqScalingFactor))
	)

	tests := []struct {
		name                                                           string
		poolToGet                                                      uint64
		preInitializedTickIndex                                        int64
		tickToGet                                                      int64
		initGlobalUptimeAccumValues                                    []sdk.DecCoins
		globalUptimeAccumDelta                                         []sdk.DecCoins
		expectedUptimeTrackers                                         []model.UptimeTracker
		additiveSpreadFactor                                           sdk.DecCoin
		expectedLiquidityDelta                                         osmomath.Dec
		expectedTickSpreadRewardGrowthOppositeDirectionOfLastTraversal sdk.DecCoins
		expectedErr                                                    error
	}{
		{
			name:                    "Get tick info of existing tick below current tick (nonzero uptime trackers)",
			poolToGet:               validPoolId,
			preInitializedTickIndex: preInitializedTickIndex,
			tickToGet:               preInitializedTickIndex,
			additiveSpreadFactor:    defaultAdditiveSpreadFactor,
			// Global uptime accums remain unchanged after tick init
			initGlobalUptimeAccumValues: expectedUptimes.twoHundredTokensMultiDenom,
			globalUptimeAccumDelta:      expectedUptimes.hundredTokensMultiDenom,
			// We expect new uptime trackers to be new global - init global
			// This is because we init them to twoHundredTokensMultiDenom and then add hundredTokensMultiDenom,
			// so when we cross the tick and "flip" it, we expect threeHundredTokensMultiDenom - twoHundredTokensMultiDenom
			// Note that initGlobalUptimeAccumValues and globalUptimeAccumDelta get scaled by the addToUptimeAccums(...) helper
			// As a result, we also need to scale the expectedUptimeTrackers
			expectedUptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(expectedUptimes.hundredTokensMultiDenom)),
			expectedLiquidityDelta: DefaultLiquidityAmt.Neg(),
			expectedTickSpreadRewardGrowthOppositeDirectionOfLastTraversal: DefaultSpreadRewardAccumCoins.Add(defaultAdditiveSpreadFactor),
		},
		{
			name:                    "Get tick info of existing tick below current tick (nil uptime trackers)",
			poolToGet:               validPoolId,
			preInitializedTickIndex: preInitializedTickIndex,
			tickToGet:               preInitializedTickIndex,
			additiveSpreadFactor:    defaultAdditiveSpreadFactor,
			expectedUptimeTrackers:  emptyUptimeTrackers,
			expectedLiquidityDelta:  DefaultLiquidityAmt.Neg(),
			expectedTickSpreadRewardGrowthOppositeDirectionOfLastTraversal: DefaultSpreadRewardAccumCoins.Add(defaultAdditiveSpreadFactor),
		},
		{
			name:                    "Get tick info of an existing tick above current tick (nonzero uptime trackers)",
			poolToGet:               validPoolId,
			preInitializedTickIndex: DefaultCurrTick + 1,
			tickToGet:               DefaultCurrTick + 1,
			additiveSpreadFactor:    defaultAdditiveSpreadFactor,
			// Global uptime accums remain unchanged after tick init
			initGlobalUptimeAccumValues: expectedUptimes.twoHundredTokensMultiDenom,
			globalUptimeAccumDelta:      expectedUptimes.hundredTokensMultiDenom,
			// We expect new uptime trackers to be equal to new global
			// This is because we init them to zero (since target tick is above current tick),
			// so when we cross the tick and "flip" it, we expect it to be the global value - 0 = global value.
			// Note that initGlobalUptimeAccumValues and globalUptimeAccumDelta get scaled by the addToUptimeAccums(...) helper
			// As a result, we also need to scale the expectedUptimeTrackers
			expectedUptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(expectedUptimes.threeHundredTokensMultiDenom)),
			expectedLiquidityDelta: DefaultLiquidityAmt.Neg(),
			expectedTickSpreadRewardGrowthOppositeDirectionOfLastTraversal: DefaultSpreadRewardAccumCoins.Add(defaultAdditiveSpreadFactor).Add(DefaultSpreadRewardAccumCoins...),
		},
		{
			name:                    "Get tick info of new tick with a separate existing tick below current tick (nonzero uptime trackers)",
			poolToGet:               validPoolId,
			preInitializedTickIndex: preInitializedTickIndex,
			tickToGet:               DefaultCurrTick + 1,
			additiveSpreadFactor:    defaultAdditiveSpreadFactor,
			// Global uptime accums remain unchanged after tick init
			initGlobalUptimeAccumValues: expectedUptimes.twoHundredTokensMultiDenom,
			globalUptimeAccumDelta:      expectedUptimes.hundredTokensMultiDenom,
			// We expect new uptime trackers to be equal to new global
			// This is because we init them to zero (since target tick is above current tick),
			// so when we cross the tick and "flip" it, we expect it to be the global value - 0 = global value.
			// Note that initGlobalUptimeAccumValues and globalUptimeAccumDelta get scaled by the addToUptimeAccums(...) helper
			// As a result, we also need to scale the expectedUptimeTrackers
			expectedUptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(expectedUptimes.threeHundredTokensMultiDenom)),
			expectedLiquidityDelta: osmomath.ZeroDec(),
			expectedTickSpreadRewardGrowthOppositeDirectionOfLastTraversal: DefaultSpreadRewardAccumCoins.Add(defaultAdditiveSpreadFactor).Add(DefaultSpreadRewardAccumCoins...),
		},
		{
			// Note that this test case covers technically undefined behavior (crossing into the current tick).
			name:                    "Get tick info of existing tick at current tick (nonzero uptime trackers)",
			poolToGet:               validPoolId,
			preInitializedTickIndex: DefaultCurrTick,
			tickToGet:               DefaultCurrTick,
			additiveSpreadFactor:    defaultAdditiveSpreadFactor,
			// Global uptime accums remain unchanged after tick init
			initGlobalUptimeAccumValues: expectedUptimes.twoHundredTokensMultiDenom,
			globalUptimeAccumDelta:      expectedUptimes.hundredTokensMultiDenom,
			// We expect new uptime trackers to be new global - init global
			// This is because we init them to twoHundredTokensMultiDenom and then add hundredTokensMultiDenom,
			// so when we cross the tick and "flip" it, we expect threeHundredTokensMultiDenom - twoHundredTokensMultiDenom
			// Note that initGlobalUptimeAccumValues and globalUptimeAccumDelta get scaled by the addToUptimeAccums(...) helper
			// As a result, we also need to scale the expectedUptimeTrackers
			expectedUptimeTrackers: wrapUptimeTrackers(s.scaleUptimeAccumulators(expectedUptimes.hundredTokensMultiDenom)),
			expectedLiquidityDelta: DefaultLiquidityAmt.Neg(),
			expectedTickSpreadRewardGrowthOppositeDirectionOfLastTraversal: DefaultSpreadRewardAccumCoins.Add(defaultAdditiveSpreadFactor),
		},
		{
			name:                    "Twice the default additive spread factor",
			poolToGet:               validPoolId,
			preInitializedTickIndex: preInitializedTickIndex,
			tickToGet:               preInitializedTickIndex,
			additiveSpreadFactor:    defaultAdditiveSpreadFactor.Add(defaultAdditiveSpreadFactor),
			expectedUptimeTrackers:  emptyUptimeTrackers,
			expectedLiquidityDelta:  DefaultLiquidityAmt.Neg(),
			expectedTickSpreadRewardGrowthOppositeDirectionOfLastTraversal: DefaultSpreadRewardAccumCoins.Add(defaultAdditiveSpreadFactor.Add(defaultAdditiveSpreadFactor)),
		},
		{
			name:                    "error: Nil tick",
			poolToGet:               validPoolId,
			preInitializedTickIndex: preInitializedTickIndex,
			tickToGet:               preInitializedTickIndex,
			additiveSpreadFactor:    defaultAdditiveSpreadFactor,
			expectedErr:             types.ErrNextTickInfoNil,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			// Create a default CL pool
			clPool := s.PrepareConcentratedPool()
			clPool.SetCurrentTick(DefaultCurrTick)

			if test.poolToGet == validPoolId {
				s.FundAcc(s.TestAccs[0], sdk.NewCoins(DefaultCoin0, DefaultCoin1))
				_, err := s.Clk.CreatePosition(s.Ctx, test.poolToGet, s.TestAccs[0], DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}

			// Charge spread factor to make sure that the global spread factor accumulator is always updated.
			// This is to test that the per-tick spread reward growth accumulator gets initialized.
			defaultAccumCoins := sdk.NewDecCoin("foo", osmomath.NewInt(50))
			s.AddToSpreadRewardAccumulator(validPoolId, defaultAccumCoins)

			// Initialize global uptime accums
			if test.initGlobalUptimeAccumValues != nil {
				err := addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, test.initGlobalUptimeAccumValues)
				s.Require().NoError(err)
			}

			// Set up an initialized tick
			_, err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, validPoolId, test.preInitializedTickIndex, DefaultLiquidityAmt, true)
			s.Require().NoError(err)

			// Update global uptime accums for edge case testing
			if test.globalUptimeAccumDelta != nil {
				err = addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, test.globalUptimeAccumDelta)
				s.Require().NoError(err)
			}

			// update the spread factor accumulator so that we have accum value > tick spread reward growth value
			// now we have 100 foo coins inside the pool accumulator
			s.AddToSpreadRewardAccumulator(validPoolId, defaultAccumCoins)

			var nextTickInfo *model.TickInfo

			// Initialize next tick info based on test case
			if test.expectedErr == nil {
				// If no error expected, pre-fetch from state.
				tickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, test.poolToGet, test.tickToGet)
				s.Require().NoError(err)
				nextTickInfo = &tickInfo
			} else if errors.Is(test.expectedErr, types.ErrNextTickInfoNil) {
				// If expecting nil tick error, set to nil
				nextTickInfo = nil
			} else {
				// If expecting other error, set to empty tick info
				nextTickInfo = &model.TickInfo{}
			}

			var uptimeAccums []*accum.AccumulatorObject
			var spreadRewardAccum *accum.AccumulatorObject
			if test.poolToGet == validPoolId {
				uptimeAccums, err = s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, test.poolToGet)
				s.Require().NoError(err)

				spreadRewardAccum, err = s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, test.poolToGet)
				s.Require().NoError(err)
			}

			// System under test
			err = s.App.ConcentratedLiquidityKeeper.CrossTick(s.Ctx, test.poolToGet, test.tickToGet, nextTickInfo, test.additiveSpreadFactor, spreadRewardAccum.GetValue(), uptimeAccums)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &test.expectedErr)
			} else {
				s.Require().NoError(err)
				liquidityDelta := nextTickInfo.LiquidityNet
				s.Require().Equal(test.expectedLiquidityDelta, liquidityDelta)

				// now check if spread factor accumulator has been properly updated
				accum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, test.poolToGet)
				s.Require().NoError(err)

				// accum value should not have changed
				s.Require().Equal(accum.GetValue(), sdk.NewDecCoins(defaultAccumCoins).MulDec(osmomath.NewDec(2)).MulDecTruncate(cl.PerUnitLiqScalingFactor))

				// check if the tick spread reward growth outside has been correctly subtracted
				tickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, test.poolToGet, test.tickToGet)
				s.Require().NoError(err)
				s.Require().Equal(test.expectedTickSpreadRewardGrowthOppositeDirectionOfLastTraversal, tickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal)

				// ensure tick being entered has properly updated uptime trackers
				s.Require().Equal(test.expectedUptimeTrackers, tickInfo.UptimeTrackers.List)

				// ensure the event is emitted with updated tick accumulators.
				s.AssertEventEmitted(s.Ctx, types.TypeEvtCrossTick, 1)
			}
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
		{
			name:      "lower tick is equal to max tick.",
			lowerTick: types.MaxTick,
			upperTick: types.MaxTick,

			expectedError: types.InvalidTickError{Tick: types.MaxTick, IsLower: true, MinTick: types.MinInitializedTick, MaxTick: types.MaxTick},
		},
		{
			name:      "upper tick is equal to min tick.",
			lowerTick: types.MinInitializedTick,
			upperTick: types.MinInitializedTick,

			expectedError: types.InvalidTickError{Tick: types.MinInitializedTick, IsLower: false, MinTick: types.MinInitializedTick, MaxTick: types.MaxTick},
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
