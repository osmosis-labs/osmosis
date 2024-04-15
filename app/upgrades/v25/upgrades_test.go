package v25_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/osmosis-labs/osmosis/v24/app/apptesting"

	concentratedtypes "github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/types"
)

const (
	v25UpgradeHeight = int64(10)
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	// Setup spread factor migration test environment
	oldMigrationList, lastPoolPositionID, migratedPoolBeforeUpgradeIncentives, nonMigratedPoolBeforeUpgradeIncentives := s.PrepareSpreadRewardsMigrationTestEnv()

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
	})

	s.ExecuteSpreadRewardsMigrationTest(oldMigrationList, lastPoolPositionID, migratedPoolBeforeUpgradeIncentives, nonMigratedPoolBeforeUpgradeIncentives)
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v25UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v25", Height: v25UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, exists := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(exists)

	s.Ctx = s.Ctx.WithBlockHeight(v25UpgradeHeight)
}

func (s *UpgradeTestSuite) PrepareSpreadRewardsMigrationTestEnv() (map[uint64]struct{}, uint64, sdk.Coins, sdk.Coins) {
	// Set the migration pool ID threshold to far away to simulate pre-migration state.
	s.App.ConcentratedLiquidityKeeper.SetSpreadFactorPoolIDMigrationThreshold(s.Ctx, 1000)

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

	feeAccumulator, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, lastPoolID)
	s.Require().NoError(err)
	feeAccumulator.AddToAccumulator(sdk.NewDecCoins(sdk.NewDecCoinFromDec("uosmo", sdk.MustNewDecFromStr("276701288297452775148000"))))

	feeAccumulator, err = s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, secondLastPoolID)
	s.Require().NoError(err)
	feeAccumulator.AddToAccumulator(sdk.NewDecCoins(sdk.NewDecCoinFromDec("uosmo", sdk.MustNewDecFromStr("276701288297452775148000"))))

	// Migrated pool claim
	migratedPoolBeforeUpgradeIncentives, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, lastPoolPositionID)
	s.Require().NoError(err)
	s.Require().NotEmpty(migratedPoolBeforeUpgradeIncentives)

	// Non-migrated pool claim
	nonMigratedPoolBeforeUpgradeIncentives, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, lastPoolPositionID-1)
	s.Require().NoError(err)
	s.Require().NotEmpty(nonMigratedPoolBeforeUpgradeIncentives)

	// Overwrite the migration list with the desired pool ID.
	oldMigrationList := concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDs
	concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDs = map[uint64]struct{}{}
	concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDs[lastPoolID] = struct{}{}

	return oldMigrationList, lastPoolPositionID, migratedPoolBeforeUpgradeIncentives, nonMigratedPoolBeforeUpgradeIncentives
}

func (s *UpgradeTestSuite) ExecuteSpreadRewardsMigrationTest(oldMigrationList map[uint64]struct{}, lastPoolPositionID uint64, migratedPoolBeforeUpgradeIncentives, nonMigratedPoolBeforeUpgradeIncentives sdk.Coins) {
	// Migrated pool: ensure that the claimable incentives are the same before and after migration
	migratedPoolAfterUpgradeIncentives, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, lastPoolPositionID)
	s.Require().NoError(err)
	s.Require().Equal(migratedPoolBeforeUpgradeIncentives.String(), migratedPoolAfterUpgradeIncentives.String())

	// Non-migrated pool: ensure that the claimable incentives are the same before and after migration
	nonMigratedPoolAfterUpgradeIncentives, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, lastPoolPositionID-1)
	s.Require().NoError(err)
	s.Require().Equal(nonMigratedPoolBeforeUpgradeIncentives.String(), nonMigratedPoolAfterUpgradeIncentives.String())

	// Restore the migration list for use by other tests
	concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDs = oldMigrationList
}
