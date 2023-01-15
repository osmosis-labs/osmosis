package concentrated_liquidity_test

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
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
			s.PrepareConcentratedPool()
			_, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(s.Ctx, 1)
			feeAccum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(s.Ctx, 1)
			s.Require().NoError(err)

			// manually update accumulator for testing
			defaultAccumCoins := sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(50)))
			feeAccum.AddToAccumulator(defaultAccumCoins)

			// If tickExists set, initialize the specified tick with defaultLiquidityAmt
			preexistingLiquidity := sdk.ZeroDec()
			if test.tickExists {
				err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, test.param.poolId, test.param.tickIndex, DefaultLiquidityAmt, test.param.upper)
				s.Require().NoError(err)
				preexistingLiquidity = DefaultLiquidityAmt
			}

			// Get the tick info for poolId 1
			tickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)

			// Ensure tick state contains any preexistingLiquidity (zero otherwise)
			s.Require().Equal(preexistingLiquidity, tickInfo.LiquidityGross)

			// Initialize or update the tick according to the test case
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, test.param.poolId, test.param.tickIndex, test.param.liquidityIn, test.param.upper)
			if test.expectedErr != nil {
				s.Require().ErrorIs(err, test.expectedErr)
				return
			}
			s.Require().NoError(err)

			// Get the tick info for poolId 1 again
			tickInfo, err = s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, 1, test.param.tickIndex)
			s.Require().NoError(err)

			// Check that the initialized or updated tick matches our expectation
			s.Require().Equal(test.expectedLiquidityNet, tickInfo.LiquidityNet)
			s.Require().Equal(test.expectedLiquidityGross, tickInfo.LiquidityGross)

			if test.param.tickIndex <= 0 {
				s.Require().Equal(defaultAccumCoins, tickInfo.FeeGrowthOutside)
			} else {
				s.Require().Equal(sdk.DecCoins(nil), tickInfo.FeeGrowthOutside)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetTickInfo() {
	var (
		preInitializedTickIndex = DefaultCurrTick.Int64() + 2
	)

	tests := []struct {
		name             string
		poolToGet        uint64
		tickToGet        int64
		expectedTickInfo model.TickInfo
		expectedErr      error
	}{
		{
			name:      "Get tick info on existing pool and existing tick",
			poolToGet: validPoolId,
			tickToGet: preInitializedTickIndex,
			// Note that FeeGrowthOutside is not updated.
			expectedTickInfo: model.TickInfo{LiquidityGross: DefaultLiquidityAmt, LiquidityNet: DefaultLiquidityAmt.Neg()},
		},
		{
			name:      "Get tick info on existing pool with no existing tick (cur pool tick > tick)",
			poolToGet: validPoolId,
			tickToGet: DefaultCurrTick.Int64() + 1,
			// Note that FeeGrowthOutside is not initialized.
			expectedTickInfo: model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec()},
		},
		{
			name:      "Get tick info on existing pool with no existing tick (cur pool tick == tick), initialized fee growth outside",
			poolToGet: validPoolId,
			tickToGet: DefaultCurrTick.Int64(),
			// Note that FeeGrowthOutside is initialized.
			expectedTickInfo: model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), FeeGrowthOutside: sdk.NewDecCoins(oneEth)},
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
			s.PrepareConcentratedPool()

			// Set up an initialized tick
			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, validPoolId, preInitializedTickIndex, DefaultLiquidityAmt, true)
			s.Require().NoError(err)

			// Charge fee to make sure that the global fee accumulator is always updates.
			// This is to test that the per-tick fee growth accumulator gets initialized.
			if test.poolToGet == validPoolId {
				s.SetupDefaultPosition(test.poolToGet)
			}
			s.App.ConcentratedLiquidityKeeper.ChargeFee(s.Ctx, test.poolToGet, oneEth)

			// System under test
			tickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, test.poolToGet, test.tickToGet)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &test.expectedErr)
				s.Require().Equal(model.TickInfo{}, tickInfo)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(test.expectedTickInfo, tickInfo)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCrossTick() {
	var (
		preInitializedTickIndex = DefaultCurrTick.Int64() - 2
	)

	tests := []struct {
		name      string
		poolToGet uint64
		tickToGet int64
		// upon setting this to true, we manipulate tick's fee growth
		// to a value that is greater than that of accumulator
		manipulateTickFeeGrowth bool
		expectedLiquidityDelta  sdk.Dec
		expectedErr             bool
	}{
		{
			name:                   "Get tick info on existing pool and existing tick",
			poolToGet:              validPoolId,
			tickToGet:              preInitializedTickIndex,
			expectedLiquidityDelta: DefaultLiquidityAmt.Neg(),
		},
		{
			name:                    "Error when trying to cross tick with more than what's in fee accumulator",
			poolToGet:               validPoolId,
			tickToGet:               preInitializedTickIndex,
			manipulateTickFeeGrowth: true,
			expectedErr:             true,
		},
		{
			name:        "Try invalid tick",
			poolToGet:   2,
			tickToGet:   preInitializedTickIndex,
			expectedErr: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a default CL pool
			s.PrepareConcentratedPool()

			// Charge fee to make sure that the global fee accumulator is always updates.
			// This is to test that the per-tick fee growth accumulator gets initialized.
			if test.poolToGet == validPoolId {
				s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000000000)), sdk.NewCoin("usdc", sdk.NewInt(1000000000000))))
				// _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.SetupPosition(test.poolToGet, s.TestAccs[0], DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick)
			}

			// manually update accumulator for testing before initializing ticks
			defaultAccumCoins := sdk.NewDecCoin("foo", sdk.NewInt(50))
			err := s.App.ConcentratedLiquidityKeeper.ChargeFee(s.Ctx, validPoolId, defaultAccumCoins)
			s.Require().NoError(err)

			// Set up an initialized tick
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, validPoolId, preInitializedTickIndex, DefaultLiquidityAmt, true)
			s.Require().NoError(err)

			// manipulate tick fee growth
			if test.manipulateTickFeeGrowth {
				err = s.App.ConcentratedLiquidityKeeper.ChargeFee(s.Ctx, validPoolId, defaultAccumCoins)
				s.Require().NoError(err)

				err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, validPoolId, preInitializedTickIndex, DefaultLiquidityAmt, true)
				s.Require().NoError(err)

				// now reduce fee accumulator manually to create erroring env
				s.App.ConcentratedLiquidityKeeper.SubFromFeeAccumulator(s.Ctx, validPoolId, DefaultFeeAccumCoins)
			}

			// System under test
			liquidityDelta, err := s.App.ConcentratedLiquidityKeeper.CrossTick(s.Ctx, test.poolToGet, test.tickToGet)
			if test.expectedErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().Equal(test.expectedLiquidityDelta, liquidityDelta)

				// now check if fee accumulator has been properly updated
				accum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(s.Ctx, test.poolToGet)
				s.Require().NoError(err)

				// accum value should have gone back to 0 (50 - 50)
				s.Require().Equal(accum.GetValue(), sdk.DecCoins(nil))
			}
		})
	}
}
