package concentrated_liquidity_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/query"
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
	s.App.ConcentratedLiquidityKeeper = cl.NewKeeper(s.App.AppCodec(), storeKey, s.App.BankKeeper, s.App.GetSubspace(types.ModuleName))

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
			s.Setup()

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
					s.Require().Equal(tickInfoAfter.FeeGrowthOutside, feeAccum.GetValue())
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
				s.Require().Equal(tickInfoAfter.FeeGrowthOutside, feeAccum.GetValue())
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
				s.Require().Equal(defaultAccumCoins, tickInfoAfter.FeeGrowthOutside)
			} else {
				s.Require().Equal(sdk.DecCoins(nil), tickInfoAfter.FeeGrowthOutside)
			}
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
			expectedTickInfo: model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), FeeGrowthOutside: sdk.NewDecCoins(oneEth), UptimeTrackers: varyingTokensAndDenoms},
		},
		{
			name:                     "Get tick info for active tick on existing pool with existing tick",
			poolToGet:                validPoolId,
			tickToGet:                DefaultCurrTick.Int64(),
			preInitUptimeAccumValues: expectedUptimes.varyingTokensMultiDenom,
			// Both fee growth and uptime trackers are set to global since tickToGet <= current tick
			expectedTickInfo: model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), FeeGrowthOutside: sdk.NewDecCoins(oneEth), UptimeTrackers: varyingTokensAndDenoms},
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
			expectedTickInfo: model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), FeeGrowthOutside: sdk.NewDecCoins(oneEth), UptimeTrackers: emptyUptimeTrackers},
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
			s.Setup()

			// Create a default CL pool
			clPool := s.PrepareConcentratedPool()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			if test.preInitUptimeAccumValues != nil {
				addToUptimeAccums(s.Ctx, clPool.GetId(), clKeeper, test.preInitUptimeAccumValues)
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
		name                         string
		poolToGet                    uint64
		preInitializedTickIndex      int64
		tickToGet                    int64
		initGlobalUptimeAccumValues  []sdk.DecCoins
		globalUptimeAccumDelta       []sdk.DecCoins
		expectedUptimeTrackers       []model.UptimeTracker
		additiveFee                  sdk.DecCoin
		expectedLiquidityDelta       sdk.Dec
		expectedTickFeeGrowthOutside sdk.DecCoins
		expectedErr                  bool
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
			expectedUptimeTrackers:       wrapUptimeTrackers(expectedUptimes.hundredTokensMultiDenom),
			expectedLiquidityDelta:       DefaultLiquidityAmt.Neg(),
			expectedTickFeeGrowthOutside: DefaultFeeAccumCoins.Add(defaultAdditiveFee),
		},
		{
			name:                         "Get tick info of existing tick below current tick (nil uptime trackers)",
			poolToGet:                    validPoolId,
			preInitializedTickIndex:      preInitializedTickIndex,
			tickToGet:                    preInitializedTickIndex,
			additiveFee:                  defaultAdditiveFee,
			expectedUptimeTrackers:       emptyUptimeTrackers,
			expectedLiquidityDelta:       DefaultLiquidityAmt.Neg(),
			expectedTickFeeGrowthOutside: DefaultFeeAccumCoins.Add(defaultAdditiveFee),
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
			expectedUptimeTrackers:       wrapUptimeTrackers(expectedUptimes.threeHundredTokensMultiDenom),
			expectedLiquidityDelta:       DefaultLiquidityAmt.Neg(),
			expectedTickFeeGrowthOutside: DefaultFeeAccumCoins.Add(defaultAdditiveFee).Add(DefaultFeeAccumCoins...),
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
			expectedUptimeTrackers:       wrapUptimeTrackers(expectedUptimes.threeHundredTokensMultiDenom),
			expectedLiquidityDelta:       sdk.ZeroDec(),
			expectedTickFeeGrowthOutside: DefaultFeeAccumCoins.Add(defaultAdditiveFee).Add(DefaultFeeAccumCoins...),
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
			expectedUptimeTrackers:       wrapUptimeTrackers(expectedUptimes.hundredTokensMultiDenom),
			expectedLiquidityDelta:       DefaultLiquidityAmt.Neg(),
			expectedTickFeeGrowthOutside: DefaultFeeAccumCoins.Add(defaultAdditiveFee),
		},
		{
			name:                         "Twice the default additive fee",
			poolToGet:                    validPoolId,
			preInitializedTickIndex:      preInitializedTickIndex,
			tickToGet:                    preInitializedTickIndex,
			additiveFee:                  defaultAdditiveFee.Add(defaultAdditiveFee),
			expectedUptimeTrackers:       emptyUptimeTrackers,
			expectedLiquidityDelta:       DefaultLiquidityAmt.Neg(),
			expectedTickFeeGrowthOutside: DefaultFeeAccumCoins.Add(defaultAdditiveFee.Add(defaultAdditiveFee)),
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
			s.Setup()

			// Create a default CL pool
			clPool := s.PrepareConcentratedPool()
			clPool.SetCurrentTick(DefaultCurrTick)

			if test.poolToGet == validPoolId {
				s.FundAcc(s.TestAccs[0], sdk.NewCoins(DefaultCoin0, DefaultCoin1))
				_, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, test.poolToGet, s.TestAccs[0], DefaultCoin0.Amount, DefaultCoin1.Amount, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}

			// Charge fee to make sure that the global fee accumulator is always updated.
			// This is to test that the per-tick fee growth accumulator gets initialized.
			defaultAccumCoins := sdk.NewDecCoin("foo", sdk.NewInt(50))
			err := s.App.ConcentratedLiquidityKeeper.ChargeFee(s.Ctx, validPoolId, defaultAccumCoins)
			s.Require().NoError(err)

			// Initialize global uptime accums
			if test.initGlobalUptimeAccumValues != nil {
				addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, test.initGlobalUptimeAccumValues)
			}

			// Set up an initialized tick
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, validPoolId, DefaultCurrTick.Int64(), test.preInitializedTickIndex, DefaultLiquidityAmt, true)
			s.Require().NoError(err)

			// Update global uptime accums for edge case testing
			if test.globalUptimeAccumDelta != nil {
				addToUptimeAccums(s.Ctx, clPool.GetId(), s.App.ConcentratedLiquidityKeeper, test.globalUptimeAccumDelta)
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
				s.Require().Equal(test.expectedTickFeeGrowthOutside, tickInfo.FeeGrowthOutside)

				// ensure tick being entered has properly updated uptime trackers
				s.Require().Equal(test.expectedUptimeTrackers, tickInfo.UptimeTrackers)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetTickLiquidityForRange() {
	defaultTick := withPoolId(defaultTick, defaultPoolId)

	tests := []struct {
		name        string
		presetTicks []genesis.FullTick

		expectedLiquidityDepthForRange []query.LiquidityDepthWithRange
	}{
		{
			name: "one full range position, testing range in between",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},
			expectedLiquidityDepthForRange: []query.LiquidityDepthWithRange{
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
			expectedLiquidityDepthForRange: []query.LiquidityDepthWithRange{
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
			expectedLiquidityDepthForRange: []query.LiquidityDepthWithRange{
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
			expectedLiquidityDepthForRange: []query.LiquidityDepthWithRange{
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
			expectedLiquidityDepthForRange: []query.LiquidityDepthWithRange{
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
			s.Setup()

			// Create a default CL pool
			s.PrepareConcentratedPool()
			for _, tick := range test.presetTicks {
				s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, tick.PoolId, tick.TickIndex, tick.Info)
			}

			liquidityForRange, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityForRange(s.Ctx, defaultPoolId)
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
		expectedLiquidityDepths []query.TickLiquidityNet
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{},
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			s.Setup()

			// Create a default CL pool
			pool := s.PrepareConcentratedPool()
			for _, tick := range test.presetTicks {
				s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, tick.PoolId, tick.TickIndex, tick.Info)
			}

			// Force initialize current sqrt price to 1.
			// Normally, initialized during position creation.
			// We only initialize ticks in this test for simplicity.
			curPrice := sdk.OneDec()
			curTick, err := math.PriceToTick(curPrice, pool.GetExponentAtPriceOne())
			s.Require().NoError(err)
			if !test.currentPoolTick.IsNil() {
				sqrtPrice, err := math.TickToSqrtPrice(test.currentPoolTick, pool.GetExponentAtPriceOne())
				s.Require().NoError(err)

				curTick = test.currentPoolTick
				curPrice = sqrtPrice
			}
			pool.SetCurrentSqrtPrice(curPrice)
			pool.SetCurrentTick(curTick)

			s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)

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
			name:          "lower tick is not divisible by deafult tick spacing",
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
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Setup()

			// use default exponent at price one
			exponentAtPriceOne := DefaultExponentAtPriceOne

			tickSpacing := defaultTickSpacing
			if test.tickSpacing != uint64(0) {
				tickSpacing = test.tickSpacing
			}

			// System Under Test
			err := cl.ValidateTickInRangeIsValid(tickSpacing, exponentAtPriceOne, test.lowerTick, test.upperTick)

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
			s.Setup()

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

func (s *KeeperTestSuite) TestGetTickLiquidityNetInDirection_BoundTick() {
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
		expectedLiquidityDepths []query.TickLiquidityNet
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{},
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			expectedLiquidityDepths: []query.TickLiquidityNet{
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
			s.Setup()

			// Create a default CL pool
			pool := s.PrepareConcentratedPool()
			for _, tick := range test.presetTicks {
				s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, tick.PoolId, tick.TickIndex, tick.Info)
			}

			// Force initialize current sqrt price to 1.
			// Normally, initialized during position creation.
			// We only initialize ticks in this test for simplicity.
			curPrice := sdk.OneDec()
			curTick, err := math.PriceToTick(curPrice, pool.GetExponentAtPriceOne())
			s.Require().NoError(err)
			if !test.currentPoolTick.IsNil() {
				sqrtPrice, err := math.TickToSqrtPrice(test.currentPoolTick, pool.GetExponentAtPriceOne())
				s.Require().NoError(err)

				curTick = test.currentPoolTick
				curPrice = sqrtPrice
			}
			pool.SetCurrentSqrtPrice(curPrice)
			pool.SetCurrentTick(curTick)

			s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)

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

// This test estimates the calculation of the tick bound for estimate swap query on
// frontend. We prototype and test it directly in Go code for the ease of setup.
// It does not have to be exact but should be close enough to the actual calculation.
// The goal is to estimate the bound tick to achive two goals:
// - minimize the number of redundant ticks the query has to fetch on top of what's required
// by the swap estimate stemming from over estimating the bound tick.
// - minimize the number of round trips (redundant queries) stemming from underestimating
// the bound tick.
//
// For context, the e2e swap estimate is as follows:
// 1. Assumme frontend has knowledge of the pool, its current tick, current sqrt price and active liquidity
// It can be fetched from Pools or Pool query
//
// 2. Swap amount in is provided by the user
// For estimating tick bound in swap in given out, assume that amount out is the amount in.
//
// 3. Determine swap direction.
//
// a) swap out given in: token 0 in -> zeroForOne, token 1 in -> oneForZero
//
// b) swap in given out: token 0 out -> zeroFoOne, token 1 out -> oneForZero
// (note that in the actual swap estimate of swap in given out, the swap direction is the opposite of what it is
// in the bound estimate. The reason is that for bound estimate, we need to know how far ahead to swap by only knowing
// the token in (so we kind of assume that we are swapping token out in))
//
// 3. Estimate the bound tick
// Calculations are taken from: https://uniswap.org/whitepaper-v3.pdf
//
// a) zeroForOne from step 2
// Swapping token 0 in for token 1 out.
// sqrt P_t = sqrt P_c + L / token_0
// Higher L -> higher tick estimate. This is good because we want to overestimate
// to grab enough ticks in one round trip query.
// Fee charge makes the target final tick greater so do charge it on toke_0.
//
// b) oneForZero from step 2
// Swapping token 1 in for token 0 out
// sqrt P_t = sqrt P_c + token_1 / L
// Higher L -> Smaller target estimate. We want higher to have
// a buffer and get all ticks in 1 query. Therefore, take 50% of current
// To gurantee we get all data in single query. The value of 50% is chosen randomly and
// can be adjusted for better performance via config.
// Fee charge makes the target smaller. We want buffer to get all ticks
// Therefore, drop fee.
//
// 4. Query GetTickLiquidityNetInDirection with the bound tick estimate from step 3.
// This query should return active tick liquidity and all liquidity net amounts from current tick to bound tick.
// If it happens so that the bound tick is estimated incorrectly, then can either do:
// a) query full range in direction of the swap
// b) use a variation of a binary search by doubling the earlier bound tick estimate until you query enoudh ticks
// For re-querying, it is possible to start from the previous bound tick by setting start tick to the old bound tick value.
//
// 5. Estimate swap amount for price impact protection on frontend.
// Having current active tick liquidity, amount in/out and liquidity net amounts from step 4, give enough
// information to calculate the price impact protection.
func (s *KeeperTestSuite) TestFunctional_EstimateTickBound_OutGivenIn_Frontend() {
	tests := map[string]SwapTest{

		//          		   5000
		//  		   4545 -----|----- 5500
		//  4000 ----------- 4545
		"copy of fee 3 swap out given in- two positions with consecutive price ranges: eth -> usdc (5% fee) (one for zero)": {
			tokenIn:                  sdk.NewCoin("eth", sdk.NewInt(2000000)),
			tokenOutDenom:            "usdc",
			priceLimit:               sdk.NewDec(4094),
			swapFee:                  sdk.MustNewDecFromStr("0.05"),
			secondPositionLowerPrice: sdk.NewDec(4000),
			secondPositionUpperPrice: sdk.NewDec(4545),

			expectedTokenIn:                   sdk.NewCoin("eth", sdk.NewInt(2000000)),
			expectedTokenOut:                  sdk.NewCoin("usdc", sdk.NewInt(8691708221)),
			expectedFeeGrowthAccumulatorValue: sdk.MustNewDecFromStr("0.000073738597832046"),
			expectedTick:                      sdk.NewInt(301393),
			expectedSqrtPrice:                 sdk.MustNewDecFromStr("64.336946417392457832"),
			newLowerPrice:                     sdk.NewDec(4000),
			newUpperPrice:                     sdk.NewDec(4545),
			expectedLiquidityNet: []query.TickLiquidityNet{
				{
					LiquidityNet: sdk.MustNewDecFromStr("319146854.154260122418390252"),
					TickIndex:    sdk.NewInt(305450),
				},
				{
					LiquidityNet: sdk.MustNewDecFromStr("1198735489.597250295669959397"),
					TickIndex:    sdk.NewInt(300000),
				},
			},
		},
		//          5000
		//  4545 -----|----- 5500
		// 			   5501 ----------- 6250
		"copy of fee 6 swap out given in - two sequential positions with a gap usdc -> eth (3% fee) (zero for one)": {
			tokenIn:                  sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenOutDenom:            "eth",
			priceLimit:               sdk.NewDec(6106),
			secondPositionLowerPrice: sdk.NewDec(5501),
			secondPositionUpperPrice: sdk.NewDec(6250),
			swapFee:                  sdk.MustNewDecFromStr("0.03"),

			expectedLiquidityNet: []query.TickLiquidityNet{
				{
					LiquidityNet: sdk.MustNewDecFromStr("-1517882343.751510418088349649"),
					TickIndex:    sdk.NewInt(315000),
				},
				{
					LiquidityNet: sdk.MustNewDecFromStr("1199528406.187413669220031452"),
					TickIndex:    sdk.NewInt(315010),
				},
				// It should be getting a third one - currently underestimates
			},
		},
	}

	for name, test := range tests {
		test := test
		s.Run(name, func() {
			s.Setup()
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000000000)), sdk.NewCoin("usdc", sdk.NewInt(1000000000000))))
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000000000)), sdk.NewCoin("usdc", sdk.NewInt(1000000000000))))

			// Create default CL pool
			s.PrepareConcentratedPool()

			// add default position
			s.SetupDefaultPosition(1)

			// add second position depending on the test
			if !test.secondPositionLowerPrice.IsNil() {
				newLowerTick, err := math.PriceToTick(test.secondPositionLowerPrice, DefaultExponentAtPriceOne)
				s.Require().NoError(err)
				newUpperTick, err := math.PriceToTick(test.secondPositionUpperPrice, DefaultExponentAtPriceOne)
				s.Require().NoError(err)

				_, _, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, 1, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64())
				s.Require().NoError(err)
			}

			// Get pool
			pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, 1)

			var (
				maxTicksPerQuery = sdk.NewInt(5000)
				sqrtPriceTarget  sdk.Dec
			)

			isZeroForOne := test.tokenIn.Denom == pool.GetToken0()

			if isZeroForOne {
				// a) zeroForOne from step 2
				// Swapping token 0 in for token 1 out.
				// sqrt P_t = sqrt P_c + L / token_0
				// Higher L -> higher tick estimate. This is good because we want to overestimate
				// to grab enough ticks in one round trip query.
				// Fee charge makes the target final tick smaller so drop it.

				estimate := pool.GetCurrentSqrtPrice().Sub(pool.GetLiquidity().Quo(test.tokenIn.Amount.ToDec()))
				fmt.Println("estimate sqrt price", estimate)

				// Note, that if we only have a few positions in the pool, the estimate will be quite off
				// as current tick liquidity will vary from active range to the next range.
				// Therefore, we take the max of the estimate and the minimum sqrt price.
				// We expect the estimate to work much better assumming that the pool has a lot of positions.
				// where there is little variation in liquidity between tick ranges.
				sqrtPriceTarget = sdk.MaxDec(estimate, types.MinSqrtPrice)

			} else {
				// b) oneForZero from step 2
				// Swapping token 1 in for token 0 out
				// sqrt P_t = sqrt P_c + token_1 / L
				// Higher L -> Smaller target estimate. We want higher to have
				// a buffer and get all ticks in 1 query. Therefore, take 50% of current
				// To gurantee we get all data in single query. The value of 50% is chosen randomly and
				// can be adjusted for better performance via config.
				// Fee charge makes the target smaller. We want buffer to get all ticks
				// Therefore, drop fee.

				estimate := pool.GetCurrentSqrtPrice().Add(test.tokenIn.Amount.ToDec().Quo(pool.GetLiquidity()))
				fmt.Println("estimate sqrt price", estimate)

				// Similarly to swapping to the left of the current sqrt price,
				// estimating tick bound in the other direction, we take the max of the estimate and the maximum sqrt price.
				// We expect the estimate to work much better assumming that the pool has a lot of positions.
				sqrtPriceTarget = sdk.MinDec(estimate, types.MaxSqrtPrice)
			}

			computedBoundTick, err := math.PriceToTick(sqrtPriceTarget.PowerMut(2), pool.GetExponentAtPriceOne())
			s.Require().NoError(err)

			// On top of the above algorithm, we mey want to bound the tick by some hardcoded value (e.g. 5000 ticks per query)
			// Also chosen randomly, we should test it in a more realistic environment and adjust the value accordingly.
			boundTick := sdk.MaxInt(computedBoundTick, maxTicksPerQuery)

			liquidityNet, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityNetInDirection(s.Ctx, pool.GetId(), test.tokenIn.Denom, sdk.Int{}, boundTick)
			s.Require().NoError(err)

			fmt.Println("liquidityNet", liquidityNet)

			s.Require().Equal(test.expectedLiquidityNet, liquidityNet)

			// perform calc
			_, _, _, _, _, sqrtPrice, err := s.App.ConcentratedLiquidityKeeper.CalcOutAmtGivenInInternal(
				s.Ctx,
				test.tokenIn, test.tokenOutDenom,
				test.swapFee, test.priceLimit, pool.GetId())

			s.Require().NoError(err)
			// This print helps to see by how much the estimation algorithm was off.
			fmt.Println("actual sqrtPrice", sqrtPrice)
		})
	}
}
