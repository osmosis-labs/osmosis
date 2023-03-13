package concentrated_liquidity_test

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

const validPoolId = 1

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
	prefixBz := types.KeyTickPrefix(1)
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
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, test.poolToGet, s.TestAccs[0], DefaultCoin0.Amount, DefaultCoin1.Amount, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultFreezeDuration)
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

func (s *KeeperTestSuite) TestGetLiquidityDepthFromIterator() {
	firstTickLiquidityDepth := types.LiquidityDepth{
		TickIndex:    sdk.NewInt(-3),
		LiquidityNet: sdk.NewDec(-30),
	}
	secondTickLiquidityDepth := types.LiquidityDepth{
		TickIndex:    sdk.NewInt(1),
		LiquidityNet: sdk.NewDec(10),
	}
	thirdTickLiquidityDepth := types.LiquidityDepth{
		TickIndex:    sdk.NewInt(2),
		LiquidityNet: sdk.NewDec(20),
	}
	fourthTickLiquidityDepth := types.LiquidityDepth{
		TickIndex:    sdk.NewInt(4),
		LiquidityNet: sdk.NewDec(40),
	}
	tests := []struct {
		name                    string
		invalidPool             bool
		expectedErr             bool
		lowerTick               int64
		upperTick               int64
		expectedLiquidityDepths []types.LiquidityDepth
	}{
		{
			name:      "Entire range of user position",
			lowerTick: firstTickLiquidityDepth.TickIndex.Int64(),
			upperTick: fourthTickLiquidityDepth.TickIndex.Int64(),
			expectedLiquidityDepths: []types.LiquidityDepth{
				firstTickLiquidityDepth,
				secondTickLiquidityDepth,
				thirdTickLiquidityDepth,
				fourthTickLiquidityDepth,
			},
		},
		{
			name:      "Half range of user position",
			lowerTick: thirdTickLiquidityDepth.TickIndex.Int64(),
			upperTick: fourthTickLiquidityDepth.TickIndex.Int64(),
			expectedLiquidityDepths: []types.LiquidityDepth{
				thirdTickLiquidityDepth,
				fourthTickLiquidityDepth,
			},
		},
		{
			name:      "single range",
			lowerTick: thirdTickLiquidityDepth.TickIndex.Int64(),
			upperTick: thirdTickLiquidityDepth.TickIndex.Int64(),
			expectedLiquidityDepths: []types.LiquidityDepth{
				thirdTickLiquidityDepth,
			},
		},
		{
			name:                    "tick that does not exist",
			lowerTick:               10,
			upperTick:               10,
			expectedLiquidityDepths: []types.LiquidityDepth{},
		},
		{
			name:        "invalid pool id",
			invalidPool: true,
			lowerTick:   thirdTickLiquidityDepth.TickIndex.Int64(),
			upperTick:   fourthTickLiquidityDepth.TickIndex.Int64(),
			expectedErr: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a default CL pool
			pool := s.PrepareConcentratedPool()

			// Create ticks
			// Initialized tickIndex -> liquidity net gross as following:
			// 1 -> 10, 2 -> 20, 3 -> 30, 4 -> 40
			s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, pool.GetId(), firstTickLiquidityDepth.TickIndex.Int64(), model.TickInfo{
				LiquidityNet: firstTickLiquidityDepth.LiquidityNet,
			})
			s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, pool.GetId(), secondTickLiquidityDepth.TickIndex.Int64(), model.TickInfo{
				LiquidityNet: secondTickLiquidityDepth.LiquidityNet,
			})
			s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, pool.GetId(), thirdTickLiquidityDepth.TickIndex.Int64(), model.TickInfo{
				LiquidityNet: thirdTickLiquidityDepth.LiquidityNet,
			})
			s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, pool.GetId(), fourthTickLiquidityDepth.TickIndex.Int64(), model.TickInfo{
				LiquidityNet: fourthTickLiquidityDepth.LiquidityNet,
			})

			paramPoolId := pool.GetId()
			if test.invalidPool {
				paramPoolId = pool.GetId() + 1
			}

			// System Under Test
			liquidityDepths, err := s.App.ConcentratedLiquidityKeeper.GetPerTickLiquidityDepthFromRange(
				s.Ctx,
				paramPoolId,
				test.lowerTick,
				test.upperTick,
			)

			if test.expectedErr {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().True(reflect.DeepEqual(liquidityDepths, test.expectedLiquidityDepths))
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
