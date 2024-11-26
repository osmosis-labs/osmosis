package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/coinutil"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
)

var (
	USDC  = apptesting.USDC
	ETH   = apptesting.ETH
	BAR   = apptesting.BAR
	FOO   = apptesting.FOO
	UOSMO = appparams.BaseCoinUnit

	defaultAmount = osmomath.NewInt(100_000_000_000)

	// Test volume values
	oneMillionVolumeAmt = osmomath.NewInt(1_000_000_000_000)
	sub10KVolumeAmount  = osmomath.NewInt(9_876_543_21)
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
// Call AfterEpochEnd for multiple epochs.
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

	// Define test volume amounts
	volumeA := oneMillionVolumeAmt
	volumeB := sub10KVolumeAmount

	// Create a perpetual set of pools that only perpetual group gauge incentivizes
	perpetualPoolAndGaugeInfo := s.PrepareAllSupportedPools()

	// Create a non-perpetual set of pools that only non-perpetual group gauge incentivizes
	nonPerpetualPoolAndGaugeInfo := s.PrepareAllSupportedPools()

	perpetualGroupPoolIDs := []uint64{
		// perpetual pools
		perpetualPoolAndGaugeInfo.BalancerPoolID, perpetualPoolAndGaugeInfo.ConcentratedPoolID, perpetualPoolAndGaugeInfo.StableSwapPoolID,
	}
	// Compute uneven volumes
	unevenPoolVolumes := setupUnequalVolumeWeights(len(perpetualGroupPoolIDs), volumeA)
	// Setup volumes to let group creation pass.
	s.SetupVolumeForPools(perpetualGroupPoolIDs, unevenPoolVolumes, map[uint64]osmomath.Int{})

	perpetualGroupGaugeID, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins, types.PerpetualNumEpochsPaidOver, s.TestAccs[0], perpetualGroupPoolIDs)
	s.Require().NoError(err)

	// Update volumes post-group creation
	perpetualPoolIDToVolumeMap := map[uint64]osmomath.Int{}
	s.SetupVolumeForPools(perpetualGroupPoolIDs, unevenPoolVolumes, perpetualPoolIDToVolumeMap)

	nonPerpetualGroupPoolIDs := []uint64{
		// non-perpetual pools
		nonPerpetualPoolAndGaugeInfo.ConcentratedPoolID, nonPerpetualPoolAndGaugeInfo.StableSwapPoolID, nonPerpetualPoolAndGaugeInfo.BalancerPoolID,
	}

	// Compute even volumes and update volumeB if rounded
	equalPoolVolumes, volumeB := setupEqualVolumeWeights(len(nonPerpetualGroupPoolIDs), volumeB)
	// Setup volumes to let group creation pass.
	s.SetupVolumeForPools(nonPerpetualGroupPoolIDs, equalPoolVolumes, map[uint64]osmomath.Int{})

	nonPerpetualGroupGaugeID, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins.Add(defaultCoins...).Add(defaultCoins...), types.PerpetualNumEpochsPaidOver+3, s.TestAccs[0], nonPerpetualGroupPoolIDs)
	s.Require().NoError(err)

	// Update volumes post-group creation
	nonPerpetualPoolIDToVolumeMap := map[uint64]osmomath.Int{}
	s.SetupVolumeForPools(nonPerpetualGroupPoolIDs, equalPoolVolumes, nonPerpetualPoolIDToVolumeMap)

	// Calculate the expected distribution
	perpetualPoolIDToExpectedDistributionMap := s.computeExpectedDistributonAmountsFromVolume(perpetualPoolIDToVolumeMap, volumeA)

	// Calculate the expected distribution
	nonPerpetualPoolIDToExpectedDistributionMap := s.computeExpectedDistributonAmountsFromVolume(nonPerpetualPoolIDToVolumeMap, volumeB)

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
	s.SetupVolumeForPools(perpetualGroupPoolIDs, unevenPoolVolumes, map[uint64]osmomath.Int{})
	s.SetupVolumeForPools(nonPerpetualGroupPoolIDs, equalPoolVolumes, nonPerpetualPoolIDToVolumeMap)

	// Only non-perpetual distributes
	nonPerpetualPoolIDToExpectedDistributionMap = s.computeExpectedDistributonAmountsFromVolume(nonPerpetualPoolIDToVolumeMap, volumeB)

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
	s.SetupVolumeForPools(perpetualGroupPoolIDs, equalPoolVolumes, currentEpochPerpetualPoolVolumeMap)
	currentEpochNonPerpetualPoolVolumeMap := map[uint64]osmomath.Int{}
	s.SetupVolumeForPools(nonPerpetualGroupPoolIDs, unevenPoolVolumes, currentEpochNonPerpetualPoolVolumeMap)

	// Both groups distribute
	currentEpochExpectedDistributionsOne := s.computeExpectedDistributonAmountsFromVolume(currentEpochPerpetualPoolVolumeMap, volumeB)

	// Merge previous and current
	perpetualPoolIDToExpectedDistributionMap = osmoutils.MergeCoinMaps(currentEpochExpectedDistributionsOne, perpetualPoolIDToExpectedDistributionMap)

	currentEpochExpectedDistributionsTwo := s.computeExpectedDistributonAmountsFromVolume(currentEpochNonPerpetualPoolVolumeMap, volumeA)

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
	s.SetupVolumeForPools(perpetualGroupPoolIDs, equalPoolVolumes, currentEpochPerpetualPoolVolumeMap)
	currentEpochNonPerpetualPoolVolumeMap = map[uint64]osmomath.Int{}
	s.SetupVolumeForPools(nonPerpetualGroupPoolIDs, unevenPoolVolumes, currentEpochNonPerpetualPoolVolumeMap)

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 4)
	s.Require().NoError(err)

	// Validate distribution - expected distributions stay the same relative to previous epoch

	// Note that the perpetual gauge was not refunded. As a result, it is not distributing anymore.
	s.validateDistributionForGroup(perpetualGroupPoolIDs, perpetualPoolIDToExpectedDistributionMap)
	// Note that this was the last distribution for non-perpetual gauges. As a result they do not distribute anymore.
	s.validateDistributionForGroup(nonPerpetualGroupPoolIDs, nonPerpetualPoolIDToExpectedDistributionMap)

	// Validate that perpetual gauge is still present
	s.ValidateGroupExists(perpetualGroupGaugeID)
}

// This test focuses on validating groups distributing to pools that are in both groups.
// The structure is:
// Set up 2 groups that have the same pools in them.
// Call AfterEpochEnd hook.
// Validate that the distribution is correct.
func (s *KeeperTestSuite) TestAfterEpochEnd_Group_OverlappingPoolsInGroups() {
	s.SetupTest()

	// Create a set of pools with their internal gauges.
	poolAndGaugeInfo := s.PrepareAllSupportedPools()

	overlappingPoolIDs := []uint64{poolAndGaugeInfo.ConcentratedPoolID, poolAndGaugeInfo.StableSwapPoolID, poolAndGaugeInfo.BalancerPoolID}

	// Setup uneven volumes
	unevenPoolVolumes := setupUnequalVolumeWeights(len(overlappingPoolIDs), oneMillionVolumeAmt)
	// Configure the volumes
	poolIDToVolumeMap := map[uint64]osmomath.Int{}
	s.SetupVolumeForPools(overlappingPoolIDs, unevenPoolVolumes, poolIDToVolumeMap)

	// Create first group
	_, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins, types.PerpetualNumEpochsPaidOver, s.TestAccs[0], overlappingPoolIDs)
	s.Require().NoError(err)

	// Create second group
	_, err = s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins.Add(defaultCoins...).Add(defaultCoins...), types.PerpetualNumEpochsPaidOver+3, s.TestAccs[0], overlappingPoolIDs)
	s.Require().NoError(err)

	// Calculate the expected distribution
	poolIDToExpectedDistributionMap := s.computeExpectedDistributonAmountsFromVolume(poolIDToVolumeMap, oneMillionVolumeAmt)

	// Double the expected amounts by merging because we have two groups distributing the same amount to the same pools.
	poolIDToExpectedDistributionMap = osmoutils.MergeCoinMaps(poolIDToExpectedDistributionMap, poolIDToExpectedDistributionMap)

	distrEpochIdentifier := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 1)
	s.Require().NoError(err)

	// validate distribution
	s.validateDistributionForGroup(overlappingPoolIDs, poolIDToExpectedDistributionMap)
}

// This test focuses on validating group distributing when another existing group
// fails to sync due to distributing to a pool with no volume updated. Such group is expected
// to be skipped silently.
// The group with pools that had volume updated should still distribute.
// The structure is:
// Set up two groups. One distributes to pools that have no volume set.
// Set up volume for appropriate pools.
// Call AfterEpochEnd hook.
// Validate that the distribution is correct to only the pools that had volume updated.
func (s *KeeperTestSuite) TestAfterEpochEnd_Group_NoVolumeOnePool_SkipSilent() {
	s.SetupTest()

	// Create the first set of pools with internal gauges and a group for them.
	poolAndGaugeInfoOne := s.PrepareAllSupportedPools()
	poolIDsGroupOne := []uint64{poolAndGaugeInfoOne.ConcentratedPoolID, poolAndGaugeInfoOne.StableSwapPoolID}
	// setup initial volumes so that a Group can be created.
	s.overwriteVolumes(poolIDsGroupOne, []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount})
	_, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins, types.PerpetualNumEpochsPaidOver, s.TestAccs[0], poolIDsGroupOne)
	s.Require().NoError(err)

	// Create the second set of pools with internal gauges and a group for them.
	poolAndGaugeInfoTwo := s.PrepareAllSupportedPools()
	poolIDsGroupTwo := []uint64{poolAndGaugeInfoTwo.ConcentratedPoolID, poolAndGaugeInfoTwo.StableSwapPoolID}
	// setup initial volumes so that a Group can be created.
	s.overwriteVolumes(poolIDsGroupTwo, []osmomath.Int{defaultVolumeAmount, defaultVolumeAmount})
	_, err = s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins, types.PerpetualNumEpochsPaidOver, s.TestAccs[0], poolIDsGroupTwo)
	s.Require().NoError(err)

	// Overwrite the volumes with zero amounts to trigger an error.
	s.overwriteVolumes(poolIDsGroupTwo, []osmomath.Int{osmomath.ZeroInt(), osmomath.ZeroInt()})

	// Configure the volume only for the pools in the first group.
	// Setup uneven volumes
	unevenPoolVolumes := setupUnequalVolumeWeights(len(poolIDsGroupOne), oneMillionVolumeAmt)
	poolIDToVolumeMap := map[uint64]osmomath.Int{}
	s.SetupVolumeForPools(poolIDsGroupOne, unevenPoolVolumes, poolIDToVolumeMap)

	distrEpochIdentifier := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 1)
	s.Require().NoError(err)

	// First group should distribute because it has volume.
	poolIDToExpectedDistributionMap := s.computeExpectedDistributonAmountsFromVolume(poolIDToVolumeMap, oneMillionVolumeAmt)
	s.validateDistributionForGroup(poolIDsGroupOne, poolIDToExpectedDistributionMap)

	// Second group should not distribute because it has no volume.
	poolIDToExpectedDistributionMap = s.computeExpectedDistributonAmountsFromVolume(map[uint64]osmomath.Int{
		poolAndGaugeInfoTwo.ConcentratedPoolID: osmomath.ZeroInt(),
		poolAndGaugeInfoTwo.StableSwapPoolID:   osmomath.ZeroInt(),
	}, oneMillionVolumeAmt)
	s.validateDistributionForGroup(poolIDsGroupTwo, poolIDToExpectedDistributionMap)
}

// This test focuses on volume being changed for the group across epochs.
// It sets up 1 Group
// For the first epoch, it sets up uneven volumes with volumeA total
// For the second epoch, it sets up even volumes with volumeB total
// After each epoch, it validates the distribution amounts.
func (s *KeeperTestSuite) Test_AfterEpochEnd_Group_ChangeVolumeBetween() {
	s.SetupTest()

	var (
		volumeA = oneMillionVolumeAmt
		volumeB = sub10KVolumeAmount
	)

	// Create the first set of pools with internal gauges and a group for them.
	poolAndGaugeInfo := s.PrepareAllSupportedPools()
	poolIDsGroup := []uint64{poolAndGaugeInfo.ConcentratedPoolID, poolAndGaugeInfo.StableSwapPoolID}

	// Setup uneven volumes with volumeA total amount
	unevenPoolVolumes := setupUnequalVolumeWeights(len(poolIDsGroup), volumeA)
	poolIDToVolumeMap := map[uint64]osmomath.Int{}
	s.SetupVolumeForPools(poolIDsGroup, unevenPoolVolumes, poolIDToVolumeMap)

	// Create non-perpetual group distribution over 2 epochs.
	_, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins.Add(defaultCoins...), types.PerpetualNumEpochsPaidOver+2, s.TestAccs[0], poolIDsGroup)
	s.Require().NoError(err)

	distrEpochIdentifier := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier

	// Estimate the expected distribution
	poolIDToExpectedDistributionMap := s.computeExpectedDistributonAmountsFromVolume(poolIDToVolumeMap, volumeA)

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 1)
	s.Require().NoError(err)

	s.validateDistributionForGroup(poolIDsGroup, poolIDToExpectedDistributionMap)

	// Now, configure even volume amounts from a different total volume amount
	// Update volumeB if rounded
	equalPoolVolumes, volumeB := setupEqualVolumeWeights(len(poolIDsGroup), volumeB)
	currentEpochPerpetualPoolVolumeMap := map[uint64]osmomath.Int{}
	s.SetupVolumeForPools(poolIDsGroup, equalPoolVolumes, currentEpochPerpetualPoolVolumeMap)

	// Estimate the expected distribution
	currentEpochExpectedDistributionsOne := s.computeExpectedDistributonAmountsFromVolume(currentEpochPerpetualPoolVolumeMap, volumeB)

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 2)
	s.Require().NoError(err)

	// Merge previous and current
	// We must do this because the volume configuration changed across epochs so we used a separate map for storing
	// volumes and expected distributions across.
	poolIDToExpectedDistributionMap = osmoutils.MergeCoinMaps(currentEpochExpectedDistributionsOne, poolIDToExpectedDistributionMap)

	// Group should distribute expected amounts.
	s.validateDistributionForGroup(poolIDsGroup, poolIDToExpectedDistributionMap)
}

// This test focuses on a new group being added in-between epochs
// The structure is:
// Setup even volume across pools
// Create Group that is non-perpetual over 2 epochs with 2x defaultCoins
// Call after epoch hook
// Update volume to increase but keep being even
// Create Group that is perpetual and has 1x defaultCoins
// Call after epoch hook
// Validate that 3x defaultCoins distributed evenly across pools
func (s *KeeperTestSuite) Test_AfterEpochEnd_Group_CreateGroupsBetween() {
	s.SetupTest()

	var volumeA = oneMillionVolumeAmt

	// Create set of pools with internal gauges and a group for them.
	poolAndGaugeInfo := s.PrepareAllSupportedPools()
	poolIDsGroup := []uint64{poolAndGaugeInfo.ConcentratedPoolID, poolAndGaugeInfo.StableSwapPoolID}

	// Setup even volumes with volumeA total amount
	equalPoolVolumes, volumeA := setupEqualVolumeWeights(len(poolIDsGroup), volumeA)
	poolIDToVolumeMap := map[uint64]osmomath.Int{}
	s.SetupVolumeForPools(poolIDsGroup, equalPoolVolumes, poolIDToVolumeMap)

	// Create non-perpetual group distribution over 2 epochs.
	_, err := s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins.Add(defaultCoins...), types.PerpetualNumEpochsPaidOver+2, s.TestAccs[0], poolIDsGroup)
	s.Require().NoError(err)

	distrEpochIdentifier := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier

	// Estimate the expected distribution
	poolIDToExpectedDistributionMap := s.computeExpectedDistributonAmountsFromVolume(poolIDToVolumeMap, volumeA)

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 1)
	s.Require().NoError(err)

	s.validateDistributionForGroup(poolIDsGroup, poolIDToExpectedDistributionMap)

	// Create perpetual group distributing to the same pool (for ease of setup)
	_, err = s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins, types.PerpetualNumEpochsPaidOver, s.TestAccs[0], poolIDsGroup)
	s.Require().NoError(err)

	s.IncreaseVolumeForPools(poolIDsGroup, equalPoolVolumes)

	// Compute the expected
	// Since we have a non-perpetual gauge with 2x defaultCoins and a perpetual gauge with 1x defaultCoins,
	// The final total is 3x the expected for the non-perpetual gauge in the first epoch.
	// Merge operations below essentially do the 3x multiplication.
	expectedFinalCoinsDistributed := osmoutils.MergeCoinMaps(poolIDToExpectedDistributionMap, poolIDToExpectedDistributionMap)
	expectedFinalCoinsDistributed = osmoutils.MergeCoinMaps(expectedFinalCoinsDistributed, poolIDToExpectedDistributionMap)

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 2)
	s.Require().NoError(err)

	// Group should distribute expected amounts.
	s.validateDistributionForGroup(poolIDsGroup, expectedFinalCoinsDistributed)
}

// This test focuses on configuring volume by swapping instead of using
// a direct volume setter helper in poolmanager contrary to all other tests.
// Since we track volume in bond denom (OSMO), we first setup 2 pools that are paired with the bond denom.
// Next, we setup two pools that are to be packaged in group. One of the tokens in the pool is a token that is also
// paired with bond denom in one of the first 2 pools.
// Increase volume by swapping in the second pair of pools.
// Create a group with the second pair of pools.
// Increase volume by swapping in the second pair of pools.
// Call AfterEpochEnd hook.
// Validate that the distribution is correct.
func (s *KeeperTestSuite) Test_AfterEpochEnd_Group_SwapAndDistribute() {
	// Setup UOSMO as bond denom
	stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	stakingParams.BondDenom = UOSMO
	s.App.StakingKeeper.SetParams(s.Ctx, stakingParams)

	// Create UOSMO / USDC pool
	s.PrepareCustomBalancerPool([]balancer.PoolAsset{
		{Token: sdk.NewCoin(UOSMO, defaultAmount), Weight: osmomath.OneInt()},
		{Token: sdk.NewCoin(USDC, defaultAmount), Weight: osmomath.OneInt()},
	}, balancer.PoolParams{
		SwapFee: osmomath.ZeroDec(),
		ExitFee: osmomath.ZeroDec(),
	})

	// Create UOSMO / BAR pool
	s.PrepareCustomBalancerPool([]balancer.PoolAsset{
		{Token: sdk.NewCoin(UOSMO, defaultAmount), Weight: osmomath.OneInt()},
		{Token: sdk.NewCoin(BAR, defaultAmount), Weight: osmomath.OneInt()},
	}, balancer.PoolParams{
		SwapFee: osmomath.ZeroDec(),
		ExitFee: osmomath.ZeroDec(),
	})

	// Create two pools for group
	// ETH / USDC
	ethUSDCPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(ETH, USDC)
	ethUSDCPoolID := ethUSDCPool.GetId()
	// FOO / BAR
	fooBARPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(FOO, BAR)
	fooBARPoolID := fooBARPool.GetId()

	// Swap USDC in to create volume in concentrated pool
	// Since we set up 1:1 ratio in all pools, we expect volume to be equal to amount in (defaultAmount)
	usdcCoinIn := sdk.NewCoin(USDC, defaultAmount)
	s.increaseVolumeBySwap(ethUSDCPoolID, usdcCoinIn, defaultAmount, ETH)

	// Swap BAR in to create volume in concentrated pool
	barCoinIn := sdk.NewCoin(BAR, defaultAmount)
	s.increaseVolumeBySwap(fooBARPoolID, barCoinIn, defaultAmount, FOO)

	// Create a perpetual group.
	_, err = s.App.IncentivesKeeper.CreateGroup(s.Ctx, defaultCoins, types.PerpetualNumEpochsPaidOver, s.TestAccs[0], []uint64{ethUSDCPoolID, fooBARPoolID})
	s.Require().NoError(err)

	distrEpochIdentifier := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier

	// Increase volume since group creation
	s.increaseVolumeBySwap(ethUSDCPoolID, usdcCoinIn, defaultAmount, ETH)
	s.increaseVolumeBySwap(fooBARPoolID, barCoinIn, defaultAmount, FOO)

	// System under test
	err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 1)
	s.Require().NoError(err)

	// Since volume was equal, we expect equal split of incentives
	// across pools.
	halfDefaultCoins := coinutil.QuoRaw(defaultCoins, 2)
	s.validateDistributionForGroup([]uint64{ethUSDCPoolID, fooBARPoolID}, map[uint64]sdk.Coins{
		ethUSDCPoolID: halfDefaultCoins,
		fooBARPoolID:  halfDefaultCoins,
	})
}

// increase volume in the given pool by swapping in the given amount of coins.
// validates that the final volume is increased by the expected amount.
func (s *KeeperTestSuite) increaseVolumeBySwap(poolID uint64, tokeInCoin sdk.Coin, expectedVolumeAmtIncrease osmomath.Int, denomOut string) {
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(tokeInCoin))

	originalVoume := s.App.PoolManagerKeeper.GetOsmoVolumeForPool(s.Ctx, poolID)

	_, _, err := s.App.PoolManagerKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], poolID, tokeInCoin, denomOut, osmomath.ZeroInt())
	s.Require().NoError(err)

	finalVolume := s.App.PoolManagerKeeper.GetOsmoVolumeForPool(s.Ctx, poolID)
	s.Require().NotEqual(osmomath.ZeroInt().String(), finalVolume.String())
	s.Require().Equal(expectedVolumeAmtIncrease.String(), finalVolume.Sub(originalVoume).String())
}

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
		// all remaining coins to the last gauge.
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
func (*KeeperTestSuite) computeExpectedDistributonAmountsFromVolume(poolIDToVolumeMap map[uint64]math.Int, totalVolume osmomath.Int) map[uint64]sdk.Coins {
	totalVolumeDec := totalVolume.ToLegacyDec()
	poolIDToExpectedDistributionMapOne := map[uint64]sdk.Coins{}
	for poolID, volume := range poolIDToVolumeMap {
		currentDistribution := coinutil.MulDec(defaultCoins, volume.ToLegacyDec().Quo(totalVolumeDec))

		// Note, the reason we do this is because otherwise
		// the validation fails with 0uosmo expected vs "" actual
		// Since these are the same things, we equate the expected to an empty coins.
		if currentDistribution.IsZero() {
			currentDistribution = sdk.NewCoins()
		}

		fmt.Printf("poolId %d, currentDistribution %s\n", poolID, currentDistribution)

		poolIDToExpectedDistributionMapOne[poolID] = currentDistribution
	}
	return poolIDToExpectedDistributionMapOne
}

// sets up the volume weights that add app to totalVolumeAmount
// and are unequal.
//
// The formula to determine the weight ratio is:
// a_i = i / (n(n+1)/2)
// It is chosen so that the sum of all weights is 1 and each weight is unique.
func setupUnequalVolumeWeights(numVolumeWeightsToCreate int, totalVolumeAmount math.Int) []osmomath.Int {
	totalVolumeDec := totalVolumeAmount.ToLegacyDec()
	unequalVolumeRatios := make([]osmomath.Int, 0, numVolumeWeightsToCreate)
	n := osmomath.NewDec(int64(numVolumeWeightsToCreate))
	for i := 0; i < numVolumeWeightsToCreate; i++ {
		denominator := n.Mul(n.Add(osmomath.OneDec())).Quo(osmomath.NewDec(2))
		unequalVolumeRatios = append(unequalVolumeRatios, osmomath.NewDec(int64(i+1)).Quo(denominator).Mul(totalVolumeDec).TruncateInt())
	}
	return unequalVolumeRatios
}

// sets up the volume weights that add app to totalVolumeAmount
// and are equal.
//
// The distribution of weights is chosen so that the sum of all them is 1 and each weight is equal.
// due to rounding, the total volume amount may not be exactly equal to the sum of all weights.
// this causes rounding issues in validating final distribution amounts.
// As a result, we return the updated total volume amount that may have been rounded.
func setupEqualVolumeWeights(numVolumeWeightsToCreate int, totalVolumeAmount math.Int) ([]osmomath.Int, osmomath.Int) {
	totalVolumeDec := totalVolumeAmount.ToLegacyDec()
	equalVolumeRatios := make([]osmomath.Int, 0, numVolumeWeightsToCreate)

	updatedTotalVolume := osmomath.ZeroInt()
	for i := 0; i < numVolumeWeightsToCreate; i++ {
		currentVolume := osmomath.OneDec().Quo(osmomath.NewDec(int64(numVolumeWeightsToCreate))).Mul(totalVolumeDec)

		currentVolumeInt := currentVolume.TruncateInt()
		equalVolumeRatios = append(equalVolumeRatios, currentVolumeInt)

		updatedTotalVolume = updatedTotalVolume.Add(currentVolumeInt)
	}

	return equalVolumeRatios, updatedTotalVolume
}
