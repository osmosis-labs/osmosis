package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

// TestAfterPoolCreatedHook tests if internal tracking logic has been triggered correctly,
// and the corerct state entries have been created upon pool creation.
// This test includes test cases for swapping on the same block with pool creation.
func (s *TestSuite) TestAfterPoolCreatedHook() {
	two_asset_coin := defaultUniV2Coins
	three_asset_coins := defaultThreeAssetCoins

	tests := []struct {
		name      string
		poolCoins sdk.Coins
		// if this field is set true, we swap in the same block with pool creation
		runSwap bool
	}{
		{
			"Uni2 Pool, no swap on pool creation block",
			two_asset_coin,
			false,
		},
		{
			"Uni2 Pool, swap on pool creation block",
			two_asset_coin,
			true,
		},
		{
			"Three asset balancer pool, no swap on pool creation block",
			three_asset_coins,
			false,
		},
		{
			"Three asset balancer pool, no swap on pool creation block",
			three_asset_coins,
			true,
		},
	}

	for _, tc := range tests {
		s.SetupTest()
		s.Run(tc.name, func() {
			var poolId uint64
			// prepare pool according to test case
			poolId = s.PrepareBalancerPoolWithCoins(tc.poolCoins...)

			if tc.runSwap {
				s.RunBasicSwap(poolId)
			}

			timeBeforeBeginBlock := s.Ctx.BlockTime()
			s.twapkeeper.EndBlock(s.Ctx)
			s.BeginNewBlock(false)

			// check that creating a pool saved a new state of changed pools
			s.Require().Equal([]uint64{poolId}, s.twapkeeper.GetChangedPools(s.Ctx))

			// check that all twap records have been created for all denom pairs in pool
			twapRecords, err := s.twapkeeper.GetAllMostRecentRecordsForPool(s.Ctx, poolId)
			s.Require().NoError(err)
			denoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, poolId)
			s.Require().NoError(err)
			denomPairs0, _ := types.GetAllUniqueDenomPairs(denoms)
			s.Require().Equal(len(denomPairs0), len(twapRecords))

			// check on the individual twap records
			for _, twapRecord := range twapRecords {
				s.Require().Equal(poolId, twapRecord.PoolId)
				s.Require().Equal(sdk.ZeroDec(), twapRecord.P0ArithmeticTwapAccumulator)
				s.Require().Equal(sdk.ZeroDec(), twapRecord.P1ArithmeticTwapAccumulator)

				asset0sp, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, twapRecord.Asset0Denom, twapRecord.Asset1Denom)
				s.Require().NoError(err)
				s.Require().Equal(asset0sp, twapRecord.P0LastSpotPrice)
				s.Require().Equal(timeBeforeBeginBlock, twapRecord.Time)

				// check that we have same state entry
				storeRepresentationRecord, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, poolId, twapRecord.Asset0Denom, twapRecord.Asset1Denom)
				s.Require().Equal(twapRecord, storeRepresentationRecord)

				twapRecordBeforeTime, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), twapRecord.Asset0Denom, twapRecord.Asset1Denom)
				s.Require().NoError(err)
				s.Require().Equal(twapRecord, twapRecordBeforeTime)
			}
		})
	}
}

// TestSwapAndEndBlockTriggeringSave tests that if we:
// * create a pool in block 1
// * swap in block 2
// then after block 2 end block, we have saved records for the pool,
// for both block 1 & 2, with distinct spot prices in their records, and accumulators incremented.
// TODO: Abstract this to be more table driven, and test more pool / block setups.
func (s *TestSuite) TestSwapAndEndBlockTriggeringSave() {
	s.Ctx = s.Ctx.WithBlockTime(baseTime)

	poolId := s.PrepareBalancerPoolWithCoins(defaultUniV2Coins...)
	expectedHistoricalTwap, err := types.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denom0, denom1)
	s.Require().NoError(err)

	s.EndBlock()
	s.Commit() // clear transient store
	// Now on a clean state after a create pool
	s.Require().Equal(baseTime.Add(time.Second), s.Ctx.BlockTime())
	s.RunBasicSwap(poolId)

	// accumulators are default right here
	expectedLatestTwapUpToAccum, err := types.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denom0, denom1)
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
