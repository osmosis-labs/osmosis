package concentrated_liquidity_test

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/model"
	types "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

func (s *KeeperTestSuite) TestTickOrdering() {
	s.SetupTest()

	storeKey := sdk.NewKVStoreKey("concentrated_liquidity")
	tKey := sdk.NewTransientStoreKey("transient_test")
	s.Ctx = testutil.DefaultContext(storeKey, tKey)
	s.App.ConcentratedLiquidityKeeper = cl.NewKeeper(s.App.AppCodec(), storeKey, s.App.BankKeeper)

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

func (s *KeeperTestSuite) TestNextInitializedTick() {
	s.SetupTest()

	liquidityTicks := []int64{-200, -55, -4, 70, 78, 84, 139, 240, 535}
	for _, t := range liquidityTicks {
		s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, 1, t, model.TickInfo{})
	}

	s.Run("lte=true", func() {
		s.Run("returns tick to right if at initialized tick", func() {
			n, initd := s.App.ConcentratedLiquidityKeeper.NextInitializedTick(s.Ctx, 1, 78, false)
			s.Require().Equal(int64(84), n)
			s.Require().True(initd)
		})
		s.Run("returns tick to right if at initialized tick", func() {
			n, initd := s.App.ConcentratedLiquidityKeeper.NextInitializedTick(s.Ctx, 1, -55, false)
			s.Require().Equal(int64(-4), n)
			s.Require().True(initd)
		})
		s.Run("returns the tick directly to the right", func() {
			n, initd := s.App.ConcentratedLiquidityKeeper.NextInitializedTick(s.Ctx, 1, 77, false)
			s.Require().Equal(int64(78), n)
			s.Require().True(initd)
		})
		s.Run("returns the tick directly to the right", func() {
			n, initd := s.App.ConcentratedLiquidityKeeper.NextInitializedTick(s.Ctx, 1, -56, false)
			s.Require().Equal(int64(-55), n)
			s.Require().True(initd)
		})
		s.Run("returns the next words initialized tick if on the right boundary", func() {
			n, initd := s.App.ConcentratedLiquidityKeeper.NextInitializedTick(s.Ctx, 1, -257, false)
			s.Require().Equal(int64(-200), n)
			s.Require().True(initd)
		})
		s.Run("returns the next initialized tick from the next word", func() {
			s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, 1, 340, model.TickInfo{})

			n, initd := s.App.ConcentratedLiquidityKeeper.NextInitializedTick(s.Ctx, 1, 328, false)
			s.Require().Equal(int64(340), n)
			s.Require().True(initd)
		})
	})

	s.Run("lte=false", func() {
		s.Run("returns tick directly to the left of input tick if not initialized", func() {
			n, initd := s.App.ConcentratedLiquidityKeeper.NextInitializedTick(s.Ctx, 1, 79, true)
			s.Require().Equal(int64(78), n)
			s.Require().True(initd)
		})
		s.Run("returns same tick if initialized", func() {
			n, initd := s.App.ConcentratedLiquidityKeeper.NextInitializedTick(s.Ctx, 1, 78, true)
			s.Require().Equal(int64(78), n)
			s.Require().True(initd)
		})
		s.Run("returns next initialized tick far away", func() {
			n, initd := s.App.ConcentratedLiquidityKeeper.NextInitializedTick(s.Ctx, 1, 100, true)
			s.Require().Equal(int64(84), n)
			s.Require().True(initd)
		})
	})
}

func (s *KeeperTestSuite) TestInitOrUpdateTick() {
	const validPoolId = 1
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
			name: "Init tick 50 with 50000000000 liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       true,
			},
			tickExists:             false,
			expectedLiquidityNet:   sdk.NewDec(-50000000000),
			expectedLiquidityGross: sdk.NewDec(50000000000),
		},
		{
			name: "Init tick 50 with 50000000000 liquidity, lower",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       false,
			},
			tickExists:             false,
			expectedLiquidityNet:   sdk.NewDec(50000000000),
			expectedLiquidityGross: sdk.NewDec(50000000000),
		},
		{
			name: "Update tick 50 that already contains 50000000000 liquidity with 50000000000 more liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       true,
			},
			tickExists:             true,
			expectedLiquidityNet:   sdk.NewDec(-100000000000),
			expectedLiquidityGross: sdk.NewDec(100000000000),
		},
		{
			name: "Update tick 50 that already contains 50000000000 liquidity with 50000000000 more liquidity, lower",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       false,
			},
			tickExists:             true,
			expectedLiquidityNet:   sdk.NewDec(100000000000),
			expectedLiquidityGross: sdk.NewDec(100000000000),
		},
		{
			name: "Init tick -50 with 50000000000 liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   -50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       true,
			},
			tickExists:             false,
			expectedLiquidityNet:   sdk.NewDec(-50000000000),
			expectedLiquidityGross: sdk.NewDec(50000000000),
		},
		{
			name: "Init tick -50 with 50000000000 liquidity, lower",
			param: param{
				poolId:      validPoolId,
				tickIndex:   -50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       false,
			},
			tickExists:             false,
			expectedLiquidityNet:   sdk.NewDec(50000000000),
			expectedLiquidityGross: sdk.NewDec(50000000000),
		},
		{
			name: "Update tick -50 that already contains 50000000000 liquidity with 50000000000 more liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   -50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       true,
			},
			tickExists:             true,
			expectedLiquidityNet:   sdk.NewDec(-100000000000),
			expectedLiquidityGross: sdk.NewDec(100000000000),
		},
		{
			name: "Update tick -50 that already contains 50000000000 liquidity with 50000000000 more liquidity, lower",
			param: param{
				poolId:      validPoolId,
				tickIndex:   -50,
				liquidityIn: DefaultLiquidityAmt,
				upper:       false,
			},
			tickExists:             true,
			expectedLiquidityNet:   sdk.NewDec(100000000000),
			expectedLiquidityGross: sdk.NewDec(100000000000),
		},
		{
			name: "Init tick 50 with -50000000000 liquidity, upper",
			param: param{
				poolId:      validPoolId,
				tickIndex:   50,
				liquidityIn: DefaultLiquidityAmt.Neg(),
				upper:       true,
			},
			tickExists:             false,
			expectedLiquidityNet:   sdk.NewDec(50000000000),
			expectedLiquidityGross: sdk.NewDec(-50000000000),
		},
		{
			name: "Update tick 50 that already contains 50000000000 liquidity with -50000000000 liquidity, upper",
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
			name: "Update tick -50 that already contains 50000000000 liquidity with -50000000000 liquidity, lower",
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

			// Create a CL pool with poolId 1
			_, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, ETH, USDC, DefaultCurrSqrtPrice, sdk.NewInt(DefaultCurrTick))
			s.Require().NoError(err)

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
		})
	}
}
