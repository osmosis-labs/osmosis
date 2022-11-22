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
	s.App.ConcentratedLiquidityKeeper = cl.NewKeeper(s.App.AppCodec(), storeKey)

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
