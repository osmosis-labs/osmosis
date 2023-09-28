package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/coins"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
)

// This is a general test covering distribution to group gauges.
// This test ensures that the expected happy path functions as expected across all possible
// pool and gauge types.
//
// Create a perpetual set of pools that only perpetual group gauge incentivizes
// Create a non-perpetual set of pools that only non-perpetual group gauge incentivizes
//
// Create 2 groups and associate pools with them per description above.
//
// For the first group set the volume so that it is not equal across pools.
//
// For the second group set the volume so that the spread is even.
//
// Call AfterEpochEnd for muliple epochs.
//
// Ensure that the correct amount of rewards are distributed to the correct pool gauges. No panics occur.
//
// This test covers:
// - Changing volume on pools in-between distributions
// - perpetual distribution behavior
// - non-perpetual distribution behavior
// - non-perpetual gauge pruning
func (s *KeeperTestSuite) TestAfterEpochEnd_Group_General() {
	s.SetupTest()

	// Create a perpetual set of pools that only perpetual group gauge incentivizes
	perpetualPoolAndGaugeInfo := s.PrepareAllSupportedPools()

	// Create a non-perpetual set of pools that only non-perpetual group gauge incentivizes
	nonPerpetualPoolAndGaugeInfo := s.PrepareAllSupportedPools()

	perpetualGroupPoolIDs := []uint64{
		// perpetual pools
		perpetualPoolAndGaugeInfo.BalancerPoolID, perpetualPoolAndGaugeInfo.ConcentratedPoolID, perpetualPoolAndGaugeInfo.StableSwapPoolID,
	}

	perpetualGroupGaugeID, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins, types.PerpetualNumEpochsPaidOver, s.TestAccs[0], perpetualGroupPoolIDs)
	s.Require().NoError(err)

	nonPerpetualGroupPoolIDs := []uint64{
		// non-perpetual pools
		nonPerpetualPoolAndGaugeInfo.ConcentratedPoolID, nonPerpetualPoolAndGaugeInfo.StableSwapPoolID, nonPerpetualPoolAndGaugeInfo.BalancerPoolID,
	}

	nonPerpetualGroupGaugeID, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins.Add(defaultCoins...).Add(defaultCoins...), types.PerpetualNumEpochsPaidOver+3, s.TestAccs[0], nonPerpetualGroupPoolIDs)
	s.Require().NoError(err)

	// Define test volume amounts
	oneMillionVolumeAmt := osmomath.NewDec(1_000_000_000_000)
	sub10KVolumeAmount := osmomath.NewDec(9_876_543_21)

	// Setup uneven volumes
	unevenPoolVolumes := setupUnequalVolumeWeights(len(perpetualGroupPoolIDs), oneMillionVolumeAmt)

	perpetualPoolIDToVolumeMap := map[uint64]osmomath.Int{}
	s.setupVolumeForPools(perpetualGroupPoolIDs, unevenPoolVolumes, perpetualPoolIDToVolumeMap)

	// Calculate the expected distribution
	perpetualPoolIDToExpectedDistributionMap := s.computeExpectedDistributonAmountsFromVolume(perpetualPoolIDToVolumeMap, oneMillionVolumeAmt)

	// Setup even volumes
	equalPoolVolumes := setupEqualVolumeWeights(len(nonPerpetualGroupPoolIDs), sub10KVolumeAmount)

	nonPerpetualPoolIDToVolumeMap := map[uint64]osmomath.Int{}

	s.setupVolumeForPools(nonPerpetualGroupPoolIDs, equalPoolVolumes, nonPerpetualPoolIDToVolumeMap)

	// Calculate the expected distribution
	nonPerpetualPoolIDToExpectedDistributionMap := s.computeExpectedDistributonAmountsFromVolume(nonPerpetualPoolIDToVolumeMap, sub10KVolumeAmount)

	distrEpochIdentifier := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier

	////////////////////////////////////////////////////////////////////////////////////////////////////
	// Epoch 1 - Both groups distribute

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 1)
	s.Require().NoError(err)

	// Validate distribution
	s.validateDistributionForGroup(perpetualGroupPoolIDs, perpetualPoolIDToExpectedDistributionMap)
	s.validateDistributionForGroup(nonPerpetualGroupPoolIDs, nonPerpetualPoolIDToExpectedDistributionMap)

	///////////////////////////////////////////////////////////////////////////////////////////////////
	// Epoch 2 - Perpetual was not refunded, only non-perpetual distributes

	// Note that we provide a dummy poolIdToVolumeMap since we do not expect any distribution.
	s.setupVolumeForPools(perpetualGroupPoolIDs, unevenPoolVolumes, map[uint64]osmomath.Int{})
	s.setupVolumeForPools(nonPerpetualGroupPoolIDs, equalPoolVolumes, nonPerpetualPoolIDToVolumeMap)

	// Only non-perpetual distributes
	nonPerpetualPoolIDToExpectedDistributionMap = s.computeExpectedDistributonAmountsFromVolume(nonPerpetualPoolIDToVolumeMap, sub10KVolumeAmount)

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 2)
	s.Require().NoError(err)

	// Validate distribution

	// Note that the perpetual gauge was not refunded. As a result, it is not distributing anymore.
	// We provide poolIDToExpectedDistributionMapOne which is unchanged from the previous epoch.
	s.validateDistributionForGroup(perpetualGroupPoolIDs, perpetualPoolIDToExpectedDistributionMap)

	// Non-perpetual gauge did distributed, so we validate an updated value.
	s.validateDistributionForGroup(nonPerpetualGroupPoolIDs, nonPerpetualPoolIDToExpectedDistributionMap)

	///////////////////////////////////////////////////////////////////////////////////////////////////
	// System under test - Epoch 3 - Perpetual was refunded - both groups distribute - volume switched between the pools
	// Non-perpetual is pruned at the end.

	// Note that, compared to previous epochs, even and uneven are switched
	// As a result, volumes and weights are different so we need to recalculate the expected distriution values per-epoch
	// and then merge with the previous epoch distribution values.
	currentEpochPerpetualPoolVolumeMap := map[uint64]osmomath.Int{}
	s.setupVolumeForPools(perpetualGroupPoolIDs, equalPoolVolumes, currentEpochPerpetualPoolVolumeMap)
	currentEpochNonPerpetualPoolVolumeMap := map[uint64]osmomath.Int{}
	s.setupVolumeForPools(nonPerpetualGroupPoolIDs, unevenPoolVolumes, currentEpochNonPerpetualPoolVolumeMap)

	// Both groups distribute
	currentEpochExpectedDistributionsOne := s.computeExpectedDistributonAmountsFromVolume(currentEpochPerpetualPoolVolumeMap, sub10KVolumeAmount)

	// Merge previous and current
	perpetualPoolIDToExpectedDistributionMap = osmoutils.MergeCoinMaps(currentEpochExpectedDistributionsOne, perpetualPoolIDToExpectedDistributionMap)

	currentEpochExpectedDistributionsTwo := s.computeExpectedDistributonAmountsFromVolume(currentEpochNonPerpetualPoolVolumeMap, oneMillionVolumeAmt)

	// Merge previous and current
	nonPerpetualPoolIDToExpectedDistributionMap = osmoutils.MergeCoinMaps(currentEpochExpectedDistributionsTwo, nonPerpetualPoolIDToExpectedDistributionMap)

	// Refund the perpetual gauge
	err = s.App.IncentivesKeeper.AddToGaugeRewardsInternal(s.Ctx, defaultCoins, perpetualGroupGaugeID)
	s.Require().NoError(err)

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 3)
	s.Require().NoError(err)

	// Validate distribution

	// Perpetual gauge was refunded, so we validate an updated value.
	s.validateDistributionForGroup(perpetualGroupPoolIDs, perpetualPoolIDToExpectedDistributionMap)

	s.validateDistributionForGroup(nonPerpetualGroupPoolIDs, nonPerpetualPoolIDToExpectedDistributionMap)

	// Validate that non-perpetual gauge and group were pruned.
	s.validateGroupNotExists(nonPerpetualGroupGaugeID)

	///////////////////////////////////////////////////////////////////////////////////////////////////
	// Epoch 4 - Perpetual was not refunded, non-perpetual finished - none distribute.

	// We set up the new volumes for pools so that they do not fail silently due to lack of volume.
	currentEpochPerpetualPoolVolumeMap = map[uint64]osmomath.Int{}
	s.setupVolumeForPools(perpetualGroupPoolIDs, equalPoolVolumes, currentEpochPerpetualPoolVolumeMap)
	currentEpochNonPerpetualPoolVolumeMap = map[uint64]osmomath.Int{}
	s.setupVolumeForPools(nonPerpetualGroupPoolIDs, unevenPoolVolumes, currentEpochNonPerpetualPoolVolumeMap)

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 4)
	s.Require().NoError(err)

	// Validate distribution - expected distributions stay the same relative to previous epoch

	// Note that the perpetual gauge was not refunded. As a result, it is not distributing anymore.
	s.validateDistributionForGroup(perpetualGroupPoolIDs, perpetualPoolIDToExpectedDistributionMap)
	// Note that this was the last distribution for non-perpetual gauges. As a result they do not distribute anymore.
	s.validateDistributionForGroup(nonPerpetualGroupPoolIDs, nonPerpetualPoolIDToExpectedDistributionMap)

	// Validate that perpetual gauge is still present
	s.validateGroupExists(perpetualGroupGaugeID)
}

// TODO: create the following tests:
// https://github.com/osmosis-labs/osmosis/issues/6559
//
// Test_AfterEpochEnd_Group_OverlappingPoolsInGroups
// Test_AfterEpochEnd_Group_NoVolumeOnePool_SkipSilent
// Test_AfterEpochEnd_Group_ChangeVolumeBetween
// Test_AfterEpochEnd_Group_CreateGroupsBetween
// Test_AfterEpochEnd_Group_SwapAndDistribute

// for each pool ID, retrieves its internal gauge and asserts that the gauge has coins according to the
// poolIDToExpectedDistributionMapOne.
func (s *KeeperTestSuite) validateDistributionForGroup(groupPoolIDs []uint64, poolIDToExpectedDistributionMapOne map[uint64]sdk.Coins) {
	s.Require().NotEmpty(poolIDToExpectedDistributionMapOne)

	for i, poolID := range groupPoolIDs {
		gaugeID, err := s.App.PoolIncentivesKeeper.GetInternalGaugeIDForPool(s.Ctx, poolID)
		s.Require().NoError(err)

		gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
		s.Require().NoError(err)

		fmt.Printf("poolID %d gauge %d %s\n", poolID, gaugeID, gauge.Coins)

		// Note that to avoid leaving dust in the gauge, we distribute
		// all remaning coins to the last gauge.
		// As a result, we allow error tolerance of 1.
		if i == len(groupPoolIDs)-1 {
			// 10 because it accumulates for multi-epoch tests.
			tolerance := osmomath.ErrTolerance{AdditiveTolerance: osmomath.NewDec(10), RoundingDir: osmomath.RoundUp}
			osmoassert.Equal(s.T(), tolerance, poolIDToExpectedDistributionMapOne[poolID], gauge.Coins)
			break
		}

		s.Require().Equal(poolIDToExpectedDistributionMapOne[poolID].String(), gauge.Coins.String())
	}
}

// computes the expected distribution values for each pool in the map based on the volume each one has and the total volume.
// The expected distribution is calculated pro-rata based on the volume of each pool.
func (*KeeperTestSuite) computeExpectedDistributonAmountsFromVolume(poolIDToVolumeMap map[uint64]math.Int, totalVolume math.LegacyDec) map[uint64]sdk.Coins {
	poolIDToExpectedDistributionMapOne := map[uint64]sdk.Coins{}
	for poolID, volume := range poolIDToVolumeMap {
		currentDistribution := coins.MulDec(defaultCoins, volume.ToLegacyDec().Quo(totalVolume))

		fmt.Printf("poolId %d, currentDistribution %s\n", poolID, currentDistribution)

		poolIDToExpectedDistributionMapOne[poolID] = currentDistribution
	}
	return poolIDToExpectedDistributionMapOne
}

// sets up the volume for the pools in the group
// mutates poolIDToVolumeMap
func (s *KeeperTestSuite) setupVolumeForPools(poolIDs []uint64, volumesForEachPool []osmomath.Dec, poolIDToVolumeMap map[uint64]math.Int) {
	bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)

	s.Require().Equal(len(poolIDs), len(volumesForEachPool))
	for i := 0; i < len(poolIDs); i++ {
		currentPoolID := poolIDs[i]

		currentVolume := volumesForEachPool[i]
		currentVolumeInt := currentVolume.TruncateInt()

		fmt.Printf("currentVolume %d %s\n", i, currentVolume)

		// Retrieve the existing volume to add to it.
		existingVolume := s.App.PoolManagerKeeper.GetOsmoVolumeForPool(s.Ctx, currentPoolID)

		s.App.PoolManagerKeeper.SetVolume(s.Ctx, currentPoolID, sdk.NewCoins(sdk.NewCoin(bondDenom, existingVolume.Add(currentVolumeInt))))

		if existingVolume, ok := poolIDToVolumeMap[currentPoolID]; ok {
			poolIDToVolumeMap[currentPoolID] = existingVolume.Add(currentVolumeInt)
		} else {
			poolIDToVolumeMap[currentPoolID] = currentVolumeInt
		}
	}
}

// sets up the volume weights that add app to totalVolumeAmount
// and are unequal.
//
// The formula to determine the weight ratio is:
// a_i = i / (n(n+1)/2)
// It is chosen so that the sum of all weights is 1 and each weight is unique.
func setupUnequalVolumeWeights(numVolumeWeightsToCreate int, totalVolumeAmount math.LegacyDec) []math.LegacyDec {
	unequalVolumeRatios := make([]osmomath.Dec, 0, numVolumeWeightsToCreate)
	n := osmomath.NewDec(int64(numVolumeWeightsToCreate))
	for i := 0; i < numVolumeWeightsToCreate; i++ {
		denominator := n.Mul(n.Add(osmomath.OneDec())).Quo(osmomath.NewDec(2))
		unequalVolumeRatios = append(unequalVolumeRatios, osmomath.NewDec(int64(i+1)).Quo(denominator).Mul(totalVolumeAmount))
	}
	return unequalVolumeRatios
}

// sets up the volume weights that add app to totalVolumeAmount
// and are equal.
//
// The distribution of weights is chosen so that the sum of all them is 1 and each weight is equal.
func setupEqualVolumeWeights(numVolumeWeightsToCreate int, totalVolumeAmount math.LegacyDec) []math.LegacyDec {
	equalVolumeRatios := make([]osmomath.Dec, 0, numVolumeWeightsToCreate)
	for i := 0; i < numVolumeWeightsToCreate; i++ {
		currentVolume := osmomath.OneDec().Quo(osmomath.NewDec(int64(numVolumeWeightsToCreate))).Mul(totalVolumeAmount)
		equalVolumeRatios = append(equalVolumeRatios, currentVolume)
	}
	return equalVolumeRatios
}
