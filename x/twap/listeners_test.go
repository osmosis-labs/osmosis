package twap_test

import (
	"time"

	keeper "github.com/osmosis-labs/osmosis/v11/x/twap"
)

// TestCreatePoolFlow tests that upon a pool being created,
// we have made the correct store entries.
func (s *TestSuite) TestCreateTwoAssetPoolFlow() {
	poolId := s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])

	expectedTwap, err := keeper.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)

	twap, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(expectedTwap, twap)

	twap, err = s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(expectedTwap, twap)
}

// TODO: Write test for create pool, swap, and EndBlock in same block, we use post-swap spot price in record.

// Tests that after a swap, we are triggering internal tracking logic for a pool.
func (s *TestSuite) TestSwapTriggeringTrackPoolId() {
	poolId := s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])
	s.BeginNewBlock(false)
	s.RunBasicSwap(poolId)

	s.Require().Equal([]uint64{poolId}, s.twapkeeper.GetChangedPools(s.Ctx))
}

// TestSwapAndEndBlockTriggeringSave tests that if we:
// * create a pool in block 1
// * swap in block 2
// then after block 2 end block, we have saved records for the pool,
// for both block 1 & 2, with distinct spot prices in their records, and accumulators incremented.
// TODO: Abstract this to be more table driven, and test more pool / block setups.
func (s *TestSuite) TestSwapAndEndBlockTriggeringSave() {
	s.Ctx = s.Ctx.WithBlockTime(baseTime)
	poolId := s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])
	expectedHistoricalTwap, err := keeper.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)

	s.EndBlock()
	s.Commit() // clear transient store
	// Now on a clean state after a create pool
	s.Require().Equal(baseTime.Add(time.Second), s.Ctx.BlockTime())
	s.RunBasicSwap(poolId)

	// accumulators are default right here
	expectedLatestTwapUpToAccum, err := keeper.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)
	// ensure different spot prices
	s.Require().NotEqual(expectedHistoricalTwap.P0LastSpotPrice, expectedLatestTwapUpToAccum.P0LastSpotPrice)
	s.Require().NotEqual(expectedHistoricalTwap.P1LastSpotPrice, expectedLatestTwapUpToAccum.P1LastSpotPrice)

	s.EndBlock()

	// check records
	historicalOldTwap, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, baseTime, denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(expectedHistoricalTwap, historicalOldTwap)

	latestTwap, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(latestTwap.P0LastSpotPrice, expectedLatestTwapUpToAccum.P0LastSpotPrice)
	s.Require().Equal(latestTwap.P1LastSpotPrice, expectedLatestTwapUpToAccum.P1LastSpotPrice)

	latestTwap2, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(latestTwap, latestTwap2)

	// check accumulators incremented - we test details of correct increment in logic
	s.Require().True(latestTwap.P0ArithmeticTwapAccumulator.GT(historicalOldTwap.P0ArithmeticTwapAccumulator))
	s.Require().True(latestTwap.P1ArithmeticTwapAccumulator.GT(historicalOldTwap.P1ArithmeticTwapAccumulator))
}

// TestJoinSwapAndEndBlockTriggeringSave tests that if we:
// * create a pool in block 1
// * join that pool in block 2
// then after block 2 end block, we have saved records for the pool,
// for both block 1 & 2, with distinct spot prices in their records, and accumulators incremented.
// TODO: Abstract this to be more table driven, and test more pool / block setups.
func (s *TestSuite) TestJoinAndEndBlockTriggeringSave() {
	s.Ctx = s.Ctx.WithBlockTime(baseTime)
	poolId := s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])
	expectedHistoricalTwap, err := keeper.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)

	s.EndBlock()
	s.Commit() // clear transient store
	// Now on a clean state after a create pool
	s.Require().Equal(baseTime.Add(time.Second), s.Ctx.BlockTime())
	s.RunBasicJoin(poolId)

	// accumulators are default right here
	expectedLatestTwapUpToAccum, err := keeper.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)

	s.EndBlock()

	// check records
	historicalOldTwap, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, baseTime, denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(expectedHistoricalTwap, historicalOldTwap)

	latestTwap, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(latestTwap.P0LastSpotPrice, expectedLatestTwapUpToAccum.P0LastSpotPrice)
	s.Require().Equal(latestTwap.P1LastSpotPrice, expectedLatestTwapUpToAccum.P1LastSpotPrice)

	latestTwap2, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(latestTwap, latestTwap2)

	// check accumulators incremented - we test details of correct increment in logic
	s.Require().True(latestTwap.P0ArithmeticTwapAccumulator.GT(historicalOldTwap.P0ArithmeticTwapAccumulator))
	s.Require().True(latestTwap.P1ArithmeticTwapAccumulator.GT(historicalOldTwap.P1ArithmeticTwapAccumulator))
}
