package concentrated_liquidity_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cl "github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/types/genesis"
)

func (s *KeeperTestSuite) TestBeginBlockLazyMigration() {
	const incentiveDenom = "uosmo"
	s.SetupTest()

	// Set the migration pool ID threshold to far away to simulate pre-migration state.
	s.App.ConcentratedLiquidityKeeper.SetIncentivePoolIDMigrationThreshold(s.Ctx, 1100)

	// Reset the migration map
	types.FinalIncentiveAccumulatorPoolIDsToMigrate = map[uint64]struct{}{}

	// Create multiple pools to be migrated over the next 10 blocks
	for i := 0; i < 5; i++ {
		poolID, _, _, _, _, _ := s.CreateTestPoolWithIncentivisedPositions(false)
		types.FinalIncentiveAccumulatorPoolIDsToMigrate[poolID] = struct{}{}
	}

	// Create incentivized pool flow
	//
	lastPoolID,
		positionOneID,
		positionTwoID,
		ticksBeforeMigration,
		claimableIncentivesOneBeforeMigration,
		expectedInitialAccumulatorGrowth := s.CreateTestPoolWithIncentivisedPositions(false)

	// Overwrite the migration list with the desired pool ID.
	types.FinalIncentiveAccumulatorPoolIDsToMigrate[lastPoolID] = struct{}{}

	// System under test, lazy migrate pools
	err := cl.MigrateMainnetPools(s.Ctx, *s.App.ConcentratedLiquidityKeeper)
	s.Require().NoError(err)

	// Ensure that the pool accumulator has been properly migrated
	expectedMigratedAccumulatorGrowth := expectedInitialAccumulatorGrowth.MulDecTruncate(cl.PerUnitLiqScalingFactor)
	updatedUptimeAcc, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, lastPoolID)
	s.Require().NoError(err)
	s.Require().Equal(len(types.SupportedUptimes), len(updatedUptimeAcc))
	incentivizedUpdatedAccumulator := updatedUptimeAcc[0]
	s.Require().Equal(expectedMigratedAccumulatorGrowth.String(), incentivizedUpdatedAccumulator.GetValue().String())

	// Ensure that the ticks have been migrated
	ticksAfterMigration, err := s.App.ConcentratedLiquidityKeeper.GetAllInitializedTicksForPool(s.Ctx, lastPoolID)
	s.Require().NoError(err)

	s.Require().NotEmpty(ticksBeforeMigration)
	s.Require().Equal(len(ticksBeforeMigration), len(ticksAfterMigration))
	for i := range ticksBeforeMigration {
		// Validate that the tick uptime accumulator has been properly migrated
		s.Require().Equal(
			ticksBeforeMigration[i].Info.UptimeTrackers.List[0].UptimeGrowthOutside.MulDecTruncate(cl.PerUnitLiqScalingFactor),
			ticksAfterMigration[i].Info.UptimeTrackers.List[0].UptimeGrowthOutside,
		)
	}

	// Ensure that position 1 accumulator is not updated (zero)
	s.validateUptimePositionAccumulator(incentivizedUpdatedAccumulator, positionOneID, cl.EmptyCoins)

	// Rerun the same flow to get the same result for the incentive
	//
	_,
		_,
		positionTwoIDAfterMigration,
		ticksAfterMigration,
		claimableIncentivesCompareOneAfterMigration,
		expectedInitialAccumulatorGrowth := s.CreateTestPoolWithIncentivisedPositions(true)

	// Do the same swap as before the migration to get the same result
	s.Require().Equal(claimableIncentivesCompareOneAfterMigration.String(), claimableIncentivesOneBeforeMigration.String())

	// Ensure that position 2 cannot claim any incentives
	s.validateClaimableIncentives(positionTwoID, sdk.NewCoins())
	s.validateClaimableIncentives(positionTwoIDAfterMigration, sdk.NewCoins())

	// Migrate pools over the next 10 blocks
	//
	for i := 6; i >= 0; i-- {
		// Set the migration pool ID threshold to far away to simulate pre-migration state.
		s.App.ConcentratedLiquidityKeeper.SetIncentivePoolIDMigrationThreshold(s.Ctx, 1100)

		// System under test, lazy migrate pools
		err := cl.MigrateMainnetPools(s.Ctx, *s.App.ConcentratedLiquidityKeeper)
		s.Require().NoError(err)

		// We delete keys from the map here, as we're simulating mainnet state, and this is the easiest way to test
		delete(types.FinalIncentiveAccumulatorPoolIDsToMigrate, uint64(i))
	}
}

func (s *KeeperTestSuite) CreateTestPoolWithIncentivisedPositions(isMigrated bool) (uint64, uint64, uint64, []genesis.FullTick, sdk.Coins, sdk.DecCoins) {
	const incentiveDenom = "uosmo"
	var emissionRatePerSecDec = osmomath.OneDec()
	// Create default CL pool
	concentratedPool := s.PrepareConcentratedPool()
	poolID := concentratedPool.GetId()

	// Create position one
	// It has position accumulator snapshot of zero
	positionOneID, positionOneLiquidity := s.CreateFullRangePosition(concentratedPool, DefaultCoins)

	// Create incentive
	totalIncentiveAmount := sdk.NewCoin(incentiveDenom, osmomath.NewInt(1000000))
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(totalIncentiveAmount))
	_, err := s.App.ConcentratedLiquidityKeeper.CreateIncentive(
		s.Ctx,
		poolID,
		s.TestAccs[0],
		totalIncentiveAmount,
		emissionRatePerSecDec,
		s.Ctx.BlockTime(),
		types.DefaultAuthorizedUptimes[0],
	)
	s.Require().NoError(err)

	// Increate block time
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Minute))

	// Refetch pool
	concentratedPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolID)
	s.Require().NoError(err)
	currentTick := concentratedPool.GetCurrentTick()

	// Create position two (narrow)
	// It has non-zero position accumulator snapshot
	s.FundAcc(s.TestAccs[0], DefaultCoins)
	positionDataTwo, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(
		s.Ctx,
		poolID,
		s.TestAccs[0],
		DefaultCoins,
		osmomath.ZeroInt(),
		osmomath.ZeroInt(),
		currentTick-100,
		currentTick+100,
	)
	s.Require().NoError(err)
	positionTwoID := positionDataTwo.ID

	// Refetch pool
	concentratedPool, err = s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, poolID)
	s.Require().NoError(err)

	// Cross next right tick to update the tick accumulator by swapping
	amtIn, _, _ := s.computeSwapAmounts(poolID, concentratedPool.GetCurrentSqrtPrice(), currentTick+100, false, false)
	s.swapOneForZeroRight(poolID, sdk.NewCoin(USDC, amtIn.Ceil().TruncateInt()))

	// Sync acccumulator
	err = s.App.ConcentratedLiquidityKeeper.UpdatePoolUptimeAccumulatorsToNow(s.Ctx, poolID)
	s.Require().NoError(err)

	// Retrieve pool uptime accumulator
	uptimeAcc, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, poolID)
	s.Require().NoError(err)

	// Ensure that the accumulator has been properly initialized
	var expectedInitialAccumulatorGrowth sdk.DecCoins
	if isMigrated {
		expectedInitialAccumulatorGrowth = sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(
				incentiveDenom,
				osmomath.NewDec(60).MulMut(cl.PerUnitLiqScalingFactor).QuoTruncate(positionOneLiquidity)),
		)
	} else {
		expectedInitialAccumulatorGrowth = sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(
				incentiveDenom,
				osmomath.NewDec(60).QuoTruncate(positionOneLiquidity)),
		)
	}
	s.Require().Equal(len(types.SupportedUptimes), len(uptimeAcc))
	s.Require().Equal(expectedInitialAccumulatorGrowth.String(), uptimeAcc[0].GetValue().String())

	// Get ticks before migration
	ticksBeforeMigration, err := s.App.ConcentratedLiquidityKeeper.GetAllInitializedTicksForPool(s.Ctx, poolID)
	s.Require().NoError(err)

	// Get claimable amount for position one before the migration
	claimableIncentives, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, positionOneID)
	s.Require().NoError(err)

	return poolID, positionOneID, positionTwoID, ticksBeforeMigration, claimableIncentives, expectedInitialAccumulatorGrowth
}
