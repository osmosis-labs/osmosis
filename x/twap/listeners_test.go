package twap_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

// TestAfterPoolCreatedHook tests if internal tracking logic has been triggered correctly,
// and the correct state entries have been created upon pool creation.
// This test includes test cases for swapping on the same block with pool creation.
func (s *TestSuite) TestAfterPoolCreatedHook() {
	tests := []struct {
		name      string
		poolCoins sdk.Coins
		// if this field is set true, we swap in the same block with pool creation
		runSwap bool
	}{
		{
			"two asset Pool, no swap on pool creation block",
			defaultUniV2Coins,
			false,
		},
		{
			"two asset Pool, swap on pool creation block",
			defaultUniV2Coins,
			true,
		},
		{
			"Three asset balancer pool, no swap on pool creation block",
			defaultThreeAssetCoins,
			false,
		},
		{
			"Three asset balancer pool, no swap on pool creation block",
			defaultThreeAssetCoins,
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

// TestEndBlock tests if records are correctly updated upon endblock.
func (s *TestSuite) TestEndBlock() {
	tests := []struct {
		name       string
		poolCoins  sdk.Coins
		block1Swap bool
		block2Swap bool
	}{
		{
			"no swap after pool creation",
			defaultUniV2Coins,
			false,
			false,
		},
		{
			"swap in the same block with pool creation",
			defaultUniV2Coins,
			true,
			false,
		},
		{
			"swap after a block has passed by after pool creation",
			defaultUniV2Coins,
			false,
			true,
		},
		{
			"swap in both first and second block",
			defaultUniV2Coins,
			true,
			true,
		},
		{
			"three asset pool",
			defaultThreeAssetCoins,
			true,
			true,
		},
	}

	for _, tc := range tests {
		s.SetupTest()
		s.Run(tc.name, func() {
			// first block
			s.Ctx = s.Ctx.WithBlockTime(baseTime)
			poolId := s.PrepareBalancerPoolWithCoins(tc.poolCoins...)

			twapAfterPoolCreation, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denom0, denom1)
			s.Require().NoError(err)

			// run basic swap on the first block if set true
			if tc.block1Swap {
				s.RunBasicSwap(poolId)
			}

			// check that we have correctly stored changed pools
			changedPools := s.twapkeeper.GetChangedPools(s.Ctx)
			s.Require().Equal(1, len(changedPools))
			s.Require().Equal(poolId, changedPools[0])

			s.EndBlock()
			s.Commit()

			// Second block
			secondBlockTime := s.Ctx.BlockTime()

			// get updated twap record after end block
			twapAfterBlock1, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denom0, denom1)
			s.Require().NoError(err)

			// if no swap happened in block1, there should be no change
			// in the most recent twap record after epoch
			if !tc.block1Swap {
				s.Require().Equal(twapAfterPoolCreation, twapAfterBlock1)
			} else {
				// height should not have changed
				s.Require().Equal(twapAfterPoolCreation.Height, twapAfterBlock1.Height)
				// twap time should be same as previous blocktime
				s.Require().Equal(twapAfterPoolCreation.Time, baseTime)

				// accumulators should not have increased, as they are going through the first epoch
				s.Require().Equal(sdk.ZeroDec(), twapAfterBlock1.P0ArithmeticTwapAccumulator)
				s.Require().Equal(sdk.ZeroDec(), twapAfterBlock1.P1ArithmeticTwapAccumulator)
			}

			// check if spot price has been correctly updated in twap record
			asset0sp, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, twapAfterBlock1.Asset0Denom, twapAfterBlock1.Asset1Denom)
			s.Require().NoError(err)
			s.Require().Equal(asset0sp, twapAfterBlock1.P0LastSpotPrice)

			// run basic swap on block two for price change
			if tc.block2Swap {
				s.RunBasicSwap(poolId)
			}

			s.EndBlock()
			s.Commit()

			// Third block
			twapAfterBlock2, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denom0, denom1)
			s.Require().NoError(err)

			// if no swap happened in block 3, twap record should be same with block 2
			if !tc.block2Swap {
				s.Require().Equal(twapAfterBlock1, twapAfterBlock2)
			} else {
				s.Require().Equal(secondBlockTime, twapAfterBlock2.Time)

				// check accumulators incremented - we test details of correct increment in logic
				s.Require().True(twapAfterBlock2.P0ArithmeticTwapAccumulator.GT(twapAfterBlock1.P0ArithmeticTwapAccumulator))
				s.Require().True(twapAfterBlock2.P1ArithmeticTwapAccumulator.GT(twapAfterBlock1.P1ArithmeticTwapAccumulator))
			}

			// check if spot price has been correctly updated in twap record
			asset0sp, err = s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, twapAfterBlock1.Asset0Denom, twapAfterBlock2.Asset1Denom)
			s.Require().NoError(err)
			s.Require().Equal(asset0sp, twapAfterBlock2.P0LastSpotPrice)
		})
	}
}

// TestAfterEpochEnd tests if records get succesfully deleted via `AfterEpochEnd` hook.
// We test details of correct implementation of pruning method in store test.
// TODO: improve test using `GetAllHistoricalTimeIndexedTWAPs` and `GetAllHistoricalPoolIndexedTWAPs`
func (s *TestSuite) TestAfterEpochEnd() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockTime(baseTime)
	poolId := s.PrepareBalancerPoolWithCoins(defaultUniV2Coins...)
	twapBeforeEpoch, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denom0, denom1)
	s.Require().NoError(err)

	// epoch has passed
	pruneEpochIdentifier := s.App.TwapKeeper.PruneEpochIdentifier(s.Ctx)

	allEpochs := s.App.EpochsKeeper.AllEpochInfos(s.Ctx)
	// iterate through all epoch, ensure that epoch only gets pruned in prune epoch identifier
	for _, epoch := range allEpochs {
		s.Ctx = s.Ctx.WithBlockTime(baseTime.Add(epoch.Duration))
		s.App.TwapKeeper.EpochHooks().AfterEpochEnd(s.Ctx, epoch.Identifier, int64(1))

		recentTwapRecords, err := s.twapkeeper.GetAllMostRecentRecordsForPool(s.Ctx, poolId)
		s.Require().NoError(err)
		if epoch.Identifier == pruneEpochIdentifier {
			s.Require().Equal(0, len(recentTwapRecords))
		} else {
			s.Require().Equal(twapBeforeEpoch, recentTwapRecords[0])
		}
	}
}

// TestAfterSwap_JoinPool tests hooks for `AfterSwap` and `AfterJoinPool`.
// The purpose of this test is to test whether we correctly store the state of the
// pools that has changed with price impact.
func (s *TestSuite) TestAfterSwap_JoinPool() {
	two_asset_coins := defaultUniV2Coins
	three_asset_coins := defaultThreeAssetCoins
	tests := []struct {
		name      string
		poolCoins sdk.Coins
		swap      bool
		joinPool  bool
	}{
		{
			"swap triggers track changed pools",
			two_asset_coins,
			true,
			false,
		},
		{
			"join pool triggers track changed pools",
			two_asset_coins,
			false,
			true,
		},
		{
			"swap and join pool in same block triggers track changed pools",
			two_asset_coins,
			true,
			true,
		},
		{
			"three asset pool: swap and join pool in same block triggers track changed pools",
			three_asset_coins,
			true,
			true,
		},
	}

	for _, tc := range tests {
		s.SetupTest()
		s.Run(tc.name, func() {
			poolId := s.PrepareBalancerPoolWithCoins(tc.poolCoins...)

			if tc.swap {
				s.RunBasicSwap(poolId)
			}

			if tc.joinPool {
				s.RunBasicJoinPool(poolId)
			}

			// test that either of swapping in a pool or joining a pool
			// has triggered `trackChangedPool`, and that we have the state of price
			// impacted pools.
			changedPools := s.twapkeeper.GetChangedPools(s.Ctx)
			s.Require().Equal(1, len(changedPools))
			s.Require().Equal(poolId, changedPools[0])
		})
	}
}
