package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
	"github.com/osmosis-labs/osmosis/v11/x/twap"
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
			defaultTwoAssetCoins,
			false,
		},
		"Uni2 Pool, swap on pool creation block": {
			defaultTwoAssetCoins,
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
				expectedRecord, err := twap.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, poolId, denomPairs0[i], denomPairs1[i])
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
			defaultTwoAssetCoins,
			false,
			false,
		},
		{
			"swap in the same block with pool creation",
			defaultTwoAssetCoins,
			true,
			false,
		},
		{
			"swap after a block has passed by after pool creation",
			defaultTwoAssetCoins,
			false,
			true,
		},
		{
			"swap in both first and second block",
			defaultTwoAssetCoins,
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
			twapAfterBlock1, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, secondBlockTime, denom0, denom1)
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
func (s *TestSuite) TestAfterEpochEnd() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockTime(baseTime)
	poolId := s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins...)
	twapBeforeEpoch, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denom0, denom1)
	s.Require().NoError(err)
	pruneEpochIdentifier := s.App.TwapKeeper.PruneEpochIdentifier(s.Ctx)
	recordHistoryKeepPeriod := s.App.TwapKeeper.RecordHistoryKeepPeriod(s.Ctx)

	// make prune record time pass by, running prune epoch after this should prune old record
	s.Ctx = s.Ctx.WithBlockTime(baseTime.Add(recordHistoryKeepPeriod).Add(time.Second))

	allEpochs := s.App.EpochsKeeper.AllEpochInfos(s.Ctx)

	// iterate through all epoch, ensure that epoch only gets pruned in prune epoch identifier
	// we reverse iterate here to test epochs that are not prune epoch
	for i := len(allEpochs) - 1; i >= 0; i-- {
		s.App.TwapKeeper.EpochHooks().AfterEpochEnd(s.Ctx, allEpochs[i].Identifier, int64(1))

		recentTwapRecords, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, baseTime.Add(allEpochs[i].Duration), denom0, denom1)

		// old record should have been pruned here
		if allEpochs[i].Identifier == pruneEpochIdentifier {
			s.Require().Error(err)
			s.Require().NotEqual(twapBeforeEpoch, recentTwapRecords)

			// quit test once the record has been pruned
			return
		} else { // pruning should not be triggered at first, not pruning epoch
			s.Require().NoError(err)
			s.Require().Equal(twapBeforeEpoch, recentTwapRecords)
		}
	}
}

// TestAfterSwap_JoinPool tests hooks for `AfterSwap`, `AfterJoinPool`, and `AfterExitPool`.
// The purpose of this test is to test whether we correctly store the state of the
// pools that has changed with price impact.
func (s *TestSuite) TestPoolStateChange() {
	tests := map[string]struct {
		poolCoins sdk.Coins
		swap      bool
		joinPool  bool
		exitPool  bool
	}{
		"swap triggers track changed pools": {
			poolCoins: defaultTwoAssetCoins,
			swap:      true,
			joinPool:  false,
			exitPool:  false,
		},
		"join pool triggers track changed pools": {
			poolCoins: defaultTwoAssetCoins,
			swap:      false,
			joinPool:  true,
			exitPool:  false,
		},
		"swap and join pool in same block triggers track changed pools": {
			poolCoins: defaultTwoAssetCoins,
			swap:      true,
			joinPool:  true,
			exitPool:  false,
		},
		"three asset pool: swap and join pool in same block triggers track changed pools": {
			poolCoins: defaultThreeAssetCoins,
			swap:      true,
			joinPool:  true,
			exitPool:  false,
		},
		"exit pool triggers track changed pools in two-asset pool": {
			poolCoins: defaultTwoAssetCoins,
			swap:      false,
			joinPool:  false,
			exitPool:  true,
		},
		"exit pool triggers track changed pools in three-asset pool": {
			poolCoins: defaultThreeAssetCoins,
			swap:      false,
			joinPool:  false,
			exitPool:  true,
		},
	}

	for name, tc := range tests {
		s.SetupTest()
		s.Run(name, func() {
			poolId := s.PrepareBalancerPoolWithCoins(tc.poolCoins...)

			s.EndBlock()
			s.Commit()

			if tc.swap {
				s.RunBasicSwap(poolId)
			}

			if tc.joinPool {
				s.RunBasicJoin(poolId)
			}

			if tc.exitPool {
				s.RunBasicExit(poolId)
			}

			// test that either of swapping in a pool, joining a pool, or exiting a pool
			// has triggered `trackChangedPool`, and that we have the state of price
			// impacted pools.
			changedPools := s.twapkeeper.GetChangedPools(s.Ctx)
			s.Require().Equal(1, len(changedPools))
			s.Require().Equal(poolId, changedPools[0])
		})
	}
}
