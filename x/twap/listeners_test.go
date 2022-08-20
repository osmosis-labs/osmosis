package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
	keeper "github.com/osmosis-labs/osmosis/v11/x/twap"
	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

// TestAfterPoolCreatedHook tests if internal tracking logic has been triggered correctly,
// and the correct state entries have been created upon pool creation.
// This test includes test cases for swapping on the same block with pool creation.
func (s *TestSuite) TestAfterPoolCreatedHook() {
	tests := map[string]struct {
		poolCoins sdk.Coins
		// if this field is set true, we swap in the same block with pool creation
		runSwap bool
	}{
		"Uni2 Pool, no swap on pool creation block": {
			defaultUniV2Coins,
			false,
		},
		"Uni2 Pool, swap on pool creation block": {
			defaultUniV2Coins,
			true,
		},
		"Three asset balancer pool, no swap on pool creation block": {
			defaultThreeAssetCoins,
			false,
		},
		"Three asset balancer pool, swap on pool creation block": {
			defaultThreeAssetCoins,
			true,
		},
	}

	for name, tc := range tests {
		s.SetupTest()
		s.Run(name, func() {
			poolId := s.PrepareBalancerPoolWithCoins(tc.poolCoins...)

			if tc.runSwap {
				s.RunBasicSwap(poolId)
			}

			denoms := osmoutils.CoinsDenoms(tc.poolCoins)
			denomPairs0, denomPairs1 := types.GetAllUniqueDenomPairs(denoms)
			expectedRecords := []types.TwapRecord{}
			for i := 0; i < len(denomPairs0); i++ {
				expectedRecord, err := keeper.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denomPairs0[i], denomPairs1[i])
				s.Require().NoError(err)
				expectedRecords = append(expectedRecords, expectedRecord)
			}

			// check internal property, that the pool will go through EndBlock flow.
			s.Require().Equal([]uint64{poolId}, s.twapkeeper.GetChangedPools(s.Ctx))
			s.twapkeeper.EndBlock(s.Ctx)
			s.Commit()

			// check on the correctness of all individual twap records
			for i := 0; i < len(denomPairs0); i++ {
				actualRecord, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, poolId, denomPairs0[i], denomPairs1[i])
				s.Require().NoError(err)
				s.Require().Equal(expectedRecords[i], actualRecord)
				actualRecord, err = s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denomPairs0[i], denomPairs1[i])
				s.Require().NoError(err)
				s.Require().Equal(expectedRecords[i], actualRecord)
			}

			// consistency check that the number of records is exactly equal to the number of denompairs
			allRecords, err := s.twapkeeper.GetAllMostRecentRecordsForPool(s.Ctx, poolId)
			s.Require().NoError(err)
			s.Require().Equal(len(denomPairs0), len(allRecords))
		})
	}
}

// Tests that after a swap, we are triggering internal tracking logic for a pool.
func (s *TestSuite) TestSwapTriggeringTrackPoolId() {
	poolId := s.PrepareBalancerPoolWithCoins(defaultUniV2Coins...)
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

	poolId := s.PrepareBalancerPoolWithCoins(defaultUniV2Coins...)
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
	poolId := s.PrepareBalancerPoolWithCoins(defaultUniV2Coins[0], defaultUniV2Coins[1])
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
