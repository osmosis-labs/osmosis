package twap_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/twap"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var defaultPoolId uint64 = 1

// TestAfterPoolCreatedHook tests if internal tracking logic has been triggered correctly,
// and the correct state entries have been created upon pool creation.
// This test includes test cases for swapping on the same block with pool creation.
func (s *TestSuite) TestAfterPoolCreatedHook() {
	tests := map[string]struct {
		poolType  []poolmanagertypes.PoolType
		poolCoins sdk.Coins
		// if this field is set true, we swap in the same block with pool creation
		runSwap bool
	}{
		"no swap on pool creation block": {
			[]poolmanagertypes.PoolType{poolmanagertypes.Balancer, poolmanagertypes.Concentrated},
			defaultTwoAssetCoins,
			false,
		},
		"swap on pool creation block": {
			[]poolmanagertypes.PoolType{poolmanagertypes.Balancer, poolmanagertypes.Concentrated},
			defaultTwoAssetCoins,
			true,
		},
		"Three asset balancer pool, no swap on pool creation block": {
			[]poolmanagertypes.PoolType{poolmanagertypes.Balancer},
			defaultThreeAssetCoins,
			false,
		},
		"Three asset balancer pool, swap on pool creation block": {
			[]poolmanagertypes.PoolType{poolmanagertypes.Balancer},
			defaultThreeAssetCoins,
			true,
		},
	}

	for name, tc := range tests {
		for _, poolType := range tc.poolType {
			s.SetupTest()
			s.Run(fmt.Sprintf("%s : ", poolmanagertypes.PoolType_name[int32(poolType)])+name, func() {
				poolId := s.CreatePoolFromTypeWithCoins(poolType, tc.poolCoins)

				if tc.runSwap {
					s.RunBasicSwap(poolId)
				}

				denoms := tc.poolCoins.Denoms()
				denomPairs := types.GetAllUniqueDenomPairs(denoms)
				expectedRecords := []types.TwapRecord{}
				for _, denomPair := range denomPairs {
					expectedRecord, err := twap.NewTwapRecord(s.App.PoolManagerKeeper, s.Ctx, poolId, denomPair.Denom0, denomPair.Denom1)
					s.Require().NoError(err)

					// N.B. The twap records at pool creation are invalid for concentrated liquidity pools
					// due to lacking liquidity.
					if poolType == poolmanagertypes.Concentrated {
						expectedRecord.LastErrorTime = s.Ctx.BlockTime()
					}
					expectedRecords = append(expectedRecords, expectedRecord)
				}

				// check internal property, that the pool will go through EndBlock flow.
				s.Require().Equal([]uint64{poolId}, s.twapkeeper.GetChangedPools(s.Ctx))
				s.twapkeeper.EndBlock(s.Ctx)
				s.Commit()

				// check on the correctness of all individual twap records
				for i, denomPair := range denomPairs {
					actualRecord, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, poolId, denomPair.Denom0, denomPair.Denom1)
					s.Require().NoError(err)
					s.Require().Equal(expectedRecords[i], actualRecord)
					actualRecord, err = s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denomPair.Denom0, denomPair.Denom1)
					s.Require().NoError(err)
					s.Require().Equal(expectedRecords[i], actualRecord)
				}

				// consistency check that the number of records is exactly equal to the number of denompairs
				allRecords, err := s.twapkeeper.GetAllMostRecentRecordsForPool(s.Ctx, poolId)
				s.Require().NoError(err)
				s.Require().Equal(len(denomPairs), len(allRecords))
			})
		}
	}
}

// TestEndBlock tests if records are correctly updated upon endblock.
func (s *TestSuite) TestEndBlock() {
	tests := []struct {
		name string
		// Run this test case for every pool type specified here.
		poolTypes  []poolmanagertypes.PoolType
		poolCoins  sdk.Coins
		block1Swap bool
		block2Swap bool
	}{
		{
			"no swap after pool creation",
			[]poolmanagertypes.PoolType{poolmanagertypes.Balancer, poolmanagertypes.Concentrated},
			defaultTwoAssetCoins,
			false,
			false,
		},
		{
			"swap in the same block with pool creation",
			[]poolmanagertypes.PoolType{poolmanagertypes.Balancer, poolmanagertypes.Concentrated},
			defaultTwoAssetCoins,
			true,
			false,
		},
		{
			"swap after a block has passed by after pool creation",
			[]poolmanagertypes.PoolType{poolmanagertypes.Balancer, poolmanagertypes.Concentrated},
			defaultTwoAssetCoins,
			false,
			true,
		},
		{
			"swap in both first and second block",
			[]poolmanagertypes.PoolType{poolmanagertypes.Balancer, poolmanagertypes.Concentrated},
			defaultTwoAssetCoins,
			true,
			true,
		},
		{
			"three asset pool",
			[]poolmanagertypes.PoolType{poolmanagertypes.Balancer},
			defaultThreeAssetCoins,
			true,
			true,
		},
	}

	for _, tc := range tests {
		for _, poolType := range tc.poolTypes {
			tc := tc
			s.Run(fmt.Sprintf("%s : ", poolmanagertypes.PoolType_name[int32(poolType)])+tc.name, func() {
				s.SetupTest()
				// first block
				s.Ctx = s.Ctx.WithBlockTime(baseTime)

				poolId := s.CreatePoolFromTypeWithCoins(poolType, tc.poolCoins)

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
					if poolType == poolmanagertypes.Concentrated {
						// For concentrated liquidity pools, the twap records created after pool creation
						// are initialized with the the last error time as the current block time and invalid spot price.
						// This is because the spot price is not available until we add liquidity.
						// In this test, we create a full range position in the same block as pool creation. This full range
						// position correctly sets the spot price of the twap record in the end block. However, we get
						// twapAfterPoolCreation after the end block. As a result, these records differ.
						s.Require().NotEqual(twapAfterPoolCreation, twapAfterBlock1)
					} else {
						s.Require().Equal(twapAfterPoolCreation, twapAfterBlock1)
					}
				} else {
					// height should not have changed
					s.Require().Equal(twapAfterPoolCreation.Height, twapAfterBlock1.Height)
					// twap time should be same as previous blocktime
					s.Require().Equal(twapAfterPoolCreation.Time, baseTime)

					// accumulators should not have increased, as they are going through the first epoch
					s.Require().Equal(osmomath.ZeroDec(), twapAfterBlock1.P0ArithmeticTwapAccumulator)
					s.Require().Equal(osmomath.ZeroDec(), twapAfterBlock1.P1ArithmeticTwapAccumulator)
				}

				// check if spot price has been correctly updated in twap record
				asset0sp, err := s.App.PoolManagerKeeper.RouteCalculateSpotPrice(s.Ctx, poolId, twapAfterBlock1.Asset0Denom, twapAfterBlock1.Asset1Denom)
				s.Require().NoError(err)
				// Note: twap only supports decimal precision of 18. Thus, truncation.
				s.Require().Equal(asset0sp.Dec(), twapAfterBlock1.P0LastSpotPrice)

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
				asset0sp, err = s.App.PoolManagerKeeper.RouteCalculateSpotPrice(s.Ctx, poolId, twapAfterBlock1.Asset0Denom, twapAfterBlock2.Asset1Denom)
				s.Require().NoError(err)
				// Note: twap only supports decimal precision of 18. Thus, truncation.
				s.Require().Equal(asset0sp.Dec(), twapAfterBlock2.P0LastSpotPrice)
			})
		}
	}
}

// TestAfterEpochEnd tests if records get successfully deleted via `AfterEpochEnd` hook.
// We test details of correct implementation of pruning method in store test.
// Specifically, the newest record that is younger than the (current block time - record keep period)
// is kept, and the rest are deleted.
func (s *TestSuite) TestAfterEpochEnd() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockTime(baseTime)

	// Create TWAP record from pool creation.
	s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins...)

	// Assume some time has passed and new record created.
	s.Ctx = s.Ctx.WithBlockTime(tPlus10sp5Record.Time)
	newestRecord := tPlus10sp5Record

	s.twapkeeper.StoreNewRecord(s.Ctx, newestRecord)

	twapsBeforeEpoch, err := s.twapkeeper.GetAllHistoricalPoolIndexedTWAPs(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(2, len(twapsBeforeEpoch))

	pruneEpochIdentifier := s.App.TwapKeeper.PruneEpochIdentifier(s.Ctx)
	recordHistoryKeepPeriod := s.App.TwapKeeper.RecordHistoryKeepPeriod(s.Ctx)

	// make prune record time pass by, running prune epoch after this should prune old record
	s.Ctx = s.Ctx.WithBlockTime(newestRecord.Time.Add(recordHistoryKeepPeriod).Add(time.Second))

	allEpochs := s.App.EpochsKeeper.AllEpochInfos(s.Ctx)

	// iterate through all epoch, ensure that epoch only gets pruned in prune epoch identifier
	// we reverse iterate here to test epochs that are not prune epoch
	for i := len(allEpochs) - 1; i >= 0; i-- {
		err = s.App.TwapKeeper.EpochHooks().AfterEpochEnd(s.Ctx, allEpochs[i].Identifier, int64(1))
		s.Require().NoError(err)

		lastKeptTime := s.Ctx.BlockTime().Add(-s.twapkeeper.RecordHistoryKeepPeriod(s.Ctx))
		pruneState := s.twapkeeper.GetPruningState(s.Ctx)

		// state entry should be set for pruning state
		if allEpochs[i].Identifier == pruneEpochIdentifier {
			s.Require().Equal(true, pruneState.IsPruning)
			s.Require().Equal(lastKeptTime, pruneState.LastKeptTime)

			// reset pruning state to make sure other epochs do not modify it
			s.twapkeeper.SetPruningState(s.Ctx, types.PruningState{})
		} else { // pruning should not be triggered at first, not pruning epoch
			s.Require().NoError(err)
			s.Require().Equal(false, pruneState.IsPruning)
			s.Require().Equal(time.Time{}, pruneState.LastKeptTime)
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

// This test validates that all twap record mutators (listeners) run as expected
// and update twap + last spot price error at the desired points in the execution flow.
// It assumed that every state change message occurs in a separate block.
// This is in contrast to TestPoolStateChange_Concentrated_SameBlock
// that runs the same tests but within the same block.
func (s *TestSuite) TestPoolStateChange_Concentrated_SeparateBlocks() {
	s.SetupTest()

	var (
		validateUpdatedRecordEqualsPrevious = func(previousRecord types.TwapRecord) (updatedRecord types.TwapRecord) {
			updatedRecord = s.validateRecordUpdated(previousRecord.LastErrorTime)

			s.Require().Equal(updatedRecord, previousRecord)
			return updatedRecord
		}

		validateUpdatedRecordDoesNotEqualPrevious = func(previousRecord types.TwapRecord, expectedLastErrTime time.Time) (updatedRecord types.TwapRecord) {
			updatedRecord = s.validateRecordUpdated(expectedLastErrTime)

			s.Require().NotEqual(updatedRecord, previousRecord)
			return updatedRecord
		}

		////////////////////////////////////////////////////////////
		// First Block: create pool
		clPool = s.PrepareConcentratedPoolWithCoins(denom0, denom1)
		poolId = clPool.GetId()
	)

	// Concentrated liqidity pool creation triggers twap record creation. However, it contains twap error as there is no liquidity
	// a pool creation.
	s.validateChangedPool()

	firstBlockTime := s.Ctx.BlockTime()

	s.EndBlock()
	s.Commit()

	twapAfterPoolCreation := s.validateRecordUpdated(firstBlockTime)

	////////////////////////////////////////////////////////////
	// Second block: create a first position that initializes twap record.

	positionId, liquidityCreated := s.CreateFullRangePosition(clPool, defaultTwoAssetCoins)

	s.validateChangedPool()

	s.EndBlock()
	s.Commit()

	twapAfterFirstPositionCreation := validateUpdatedRecordDoesNotEqualPrevious(twapAfterPoolCreation, firstBlockTime)

	////////////////////////////////////////////////////////////
	// Third block: create a second position that does not change twap record.

	positionId2, liquidityCreated2 := s.CreateFullRangePosition(clPool, defaultTwoAssetCoins)

	s.validateUnchangedPool()

	s.EndBlock()
	s.Commit()

	// Note that creating second position has no effect.
	twapAfterSecondPositionCreation := validateUpdatedRecordEqualsPrevious(twapAfterFirstPositionCreation)

	////////////////////////////////////////////////////////////
	// Fourth block: perform swap and update twap.

	s.RunBasicSwap(poolId)

	s.validateChangedPool()

	s.EndBlock()
	s.Commit()

	// Note that performing a swap changes twap
	twapAfterSwap1 := validateUpdatedRecordDoesNotEqualPrevious(twapAfterSecondPositionCreation, firstBlockTime)

	////////////////////////////////////////////////////////////
	// Fifth block: partial withdraw -> does not update twap.

	s.WithdrawFullRangePosition(clPool, positionId2, liquidityCreated2)

	s.validateUnchangedPool()

	s.EndBlock()
	s.Commit()

	// Note that withdrwaing positions while there is liquidity remaining in pool
	// has no effect on twap.
	twapAfterWithdawPosition2 := validateUpdatedRecordEqualsPrevious(twapAfterSwap1)

	////////////////////////////////////////////////////////////
	// Sixth Block: withdraw in-full -> twap changes with error.

	s.WithdrawFullRangePosition(clPool, positionId, liquidityCreated)

	s.validateChangedPool()

	blockTimeSix := s.Ctx.BlockTime()

	s.EndBlock()
	s.Commit()

	// Note that removing all liquidity, created a twap record with last
	// error time equal to the block time when all liquidity was removed.
	twapAfterWithdawPosition1 := validateUpdatedRecordDoesNotEqualPrevious(twapAfterWithdawPosition2, blockTimeSix)

	////////////////////////////////////////////////////////////
	// Seventh block: create new position, updating twap with valid spot price

	s.CreateFullRangePosition(clPool, defaultTwoAssetCoins)

	s.validateChangedPool()

	s.EndBlock()
	s.Commit()

	// Note that when we re-add liqudity after fully removing it, for the context of twap,
	// we assume that it is new. That is, we drop the state and knowledge of the
	// last error time prior to that as well as any twap history.
	twapAfterCreatePosition3 := validateUpdatedRecordDoesNotEqualPrevious(twapAfterWithdawPosition1, blockTimeSix)

	////////////////////////////////////////////////////////////
	// Eight block: swap after re-adding liquidity

	s.RunBasicSwap(poolId)

	s.validateChangedPool()

	s.EndBlock()
	s.Commit()

	// Note that performing a swap changes twap
	validateUpdatedRecordDoesNotEqualPrevious(twapAfterCreatePosition3, blockTimeSix)
}

// This test validates that all twap record mutators (listeners) run as expected
// and update twap + last spot price error at the desired points in the execution flow.
// It assumed that every state change message occurs within the same block.
// This is in contrast to TestPoolStateChange_Concentrated_SeparateBlocks
// that runs the same tests but within a separate block.
func (s *TestSuite) TestPoolStateChange_Concentrated_SameBlock() {
	s.SetupTest()

	var (
		////////////////////////////////////////////////////////////
		// 1: create pool
		clPool = s.PrepareConcentratedPoolWithCoins(denom0, denom1)
		poolId = clPool.GetId()
	)

	firstBlockTime := s.Ctx.BlockTime()

	////////////////////////////////////////////////////////////
	// 2: create a first position that initializes twap record.

	positionId, liquidityCreated := s.CreateFullRangePosition(clPool, defaultTwoAssetCoins)

	////////////////////////////////////////////////////////////
	// 3: create a second position that does not change twap record.

	positionId2, liquidityCreated2 := s.CreateFullRangePosition(clPool, defaultTwoAssetCoins)

	// Note that creating second position has no effect.

	////////////////////////////////////////////////////////////
	// 4: perform swap and update twap.

	s.RunBasicSwap(poolId)

	////////////////////////////////////////////////////////////
	// 5: partial withdraw -> does not update twap.

	s.WithdrawFullRangePosition(clPool, positionId2, liquidityCreated2)

	////////////////////////////////////////////////////////////
	// 6: withdraw in-full -> twap changes with error.

	s.WithdrawFullRangePosition(clPool, positionId, liquidityCreated)

	////////////////////////////////////////////////////////////
	// 7: create new position, updating twap with valid spot price

	s.CreateFullRangePosition(clPool, defaultTwoAssetCoins)

	////////////////////////////////////////////////////////////
	// 8: swap after re-adding liquidity

	s.RunBasicSwap(poolId)

	s.validateChangedPool()

	s.EndBlock()
	s.Commit()

	// Note that when we create a concentrated liquidity pool in the same block with
	// create initial position that correctly initializes spot price and twap,
	// we still end up having the error time set to the block time when the pool was created.
	s.validateRecordUpdated(firstBlockTime)
}

func (s *TestSuite) validateRecordUpdated(expectedLastErrTime time.Time) (updatedRecord types.TwapRecord) {
	updatedRecord, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, defaultPoolId, s.Ctx.BlockTime(), denom0, denom1)
	s.Require().NoError(err)
	s.Require().Equal(expectedLastErrTime, updatedRecord.LastErrorTime)

	return updatedRecord
}

func (s *TestSuite) validateChangedPool() {
	changedPools := s.twapkeeper.GetChangedPools(s.Ctx)
	s.Require().Equal(1, len(changedPools))
	s.Require().Equal(defaultPoolId, changedPools[0])
}

func (s *TestSuite) validateUnchangedPool() {
	changedPools := s.twapkeeper.GetChangedPools(s.Ctx)
	s.Require().Equal(0, len(changedPools))
}

// This test should create multiple mock pools, test one pool's spot price returning an error,
// and ensure end blocks still work safely.
// func (s *TestSuite) TestSafetyWithPoolThatHasSpotPriceError() {
// 	s.Require().Fail("Need to implement")
// }
