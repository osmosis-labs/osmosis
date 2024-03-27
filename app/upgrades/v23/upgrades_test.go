package v23_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v24/app/apptesting"

	concentratedtypes "github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/types"
)

const (
	v23UpgradeHeight = int64(10)
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	// Set the migration pool ID threshold to far away to simulate pre-migration state.
	s.App.ConcentratedLiquidityKeeper.SetIncentivePoolIDMigrationThreshold(s.Ctx, 1000)

	concentratedPoolIDs := []uint64{}

	// Create two sets of all pools
	allPoolsOne := s.PrepareAllSupportedPools()
	allPoolsTwo := s.PrepareAllSupportedPools()

	concentratedPoolIDs = append(concentratedPoolIDs, allPoolsOne.ConcentratedPoolID)
	concentratedPoolIDs = append(concentratedPoolIDs, allPoolsTwo.ConcentratedPoolID)

	// Update authorized quote denoms
	concentratedParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	concentratedParams.AuthorizedQuoteDenoms = append(concentratedParams.AuthorizedQuoteDenoms, apptesting.USDC)
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, concentratedParams)

	// Create two more concentrated pools with positions
	secondLastPoolID := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx)
	lastPoolID := secondLastPoolID + 1
	concentratedPoolIDs = append(concentratedPoolIDs, secondLastPoolID)
	concentratedPoolIDs = append(concentratedPoolIDs, lastPoolID)
	s.CreateConcentratedPoolsAndFullRangePosition([][]string{
		{"uion", "uosmo"},
		{apptesting.ETH, apptesting.USDC},
	})

	lastPoolPositionID := s.App.ConcentratedLiquidityKeeper.GetNextPositionId(s.Ctx) - 1

	// Create incentive record for last pool
	incentiveCoin := sdk.NewCoin("uosmo", sdk.NewInt(1000000))
	_, err := s.App.ConcentratedLiquidityKeeper.CreateIncentive(s.Ctx, lastPoolID, s.TestAccs[0], incentiveCoin, osmomath.OneDec(), s.Ctx.BlockTime(), concentratedtypes.DefaultAuthorizedUptimes[0])
	s.Require().NoError(err)

	// Create incentive record for second last pool
	_, err = s.App.ConcentratedLiquidityKeeper.CreateIncentive(s.Ctx, secondLastPoolID, s.TestAccs[0], incentiveCoin, osmomath.OneDec(), s.Ctx.BlockTime(), concentratedtypes.DefaultAuthorizedUptimes[0])
	s.Require().NoError(err)

	// Make 60 seconds pass
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Minute))

	err = s.App.ConcentratedLiquidityKeeper.UpdatePoolUptimeAccumulatorsToNow(s.Ctx, lastPoolID)
	s.Require().NoError(err)

	// Migrated pool claim
	migratedPoolBeforeUpgradeIncentives, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, lastPoolPositionID)
	s.Require().NoError(err)
	s.Require().NotEmpty(migratedPoolBeforeUpgradeIncentives)

	// Non-migrated pool claim
	nonMigratedPoolBeforeUpgradeIncentives, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, lastPoolPositionID-1)
	s.Require().NoError(err)
	s.Require().NotEmpty(nonMigratedPoolBeforeUpgradeIncentives)

	// Overwrite the migration list with the desired pool ID.
	oldMigrationList := concentratedtypes.MigratedIncentiveAccumulatorPoolIDs
	concentratedtypes.MigratedIncentiveAccumulatorPoolIDs = map[uint64]struct{}{}
	concentratedtypes.MigratedIncentiveAccumulatorPoolIDs[lastPoolID] = struct{}{}

	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
	})

	// Migrated pool: ensure that the claimable incentives are the same before and after migration
	migratedPoolAfterUpgradeIncentives, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, lastPoolPositionID)
	s.Require().NoError(err)
	s.Require().Equal(migratedPoolBeforeUpgradeIncentives.String(), migratedPoolAfterUpgradeIncentives.String())

	// Non-migrated pool: ensure that the claimable incentives are the same before and after migration
	nonMigratedPoolAfterUpgradeIncentives, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, lastPoolPositionID-1)
	s.Require().NoError(err)
	s.Require().Equal(nonMigratedPoolBeforeUpgradeIncentives.String(), nonMigratedPoolAfterUpgradeIncentives.String())

	// Restore the migration list for use by other tests
	concentratedtypes.MigratedIncentiveAccumulatorPoolIDs = oldMigrationList
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v23UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v23", Height: v23UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, exists := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(exists)

	s.Ctx = s.Ctx.WithBlockHeight(v23UpgradeHeight)
}
