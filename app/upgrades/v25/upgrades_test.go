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
	oldMigrationList, lastPoolPositionID, migratedPoolBeforeUpgradeSpreadRewards, nonMigratedPoolBeforeUpgradeSpreadRewards := s.PrepareSpreadRewardsMigrationTestEnv()

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
	})

	s.ExecuteSpreadRewardsMigrationTest(oldMigrationList, lastPoolPositionID, migratedPoolBeforeUpgradeSpreadRewards, nonMigratedPoolBeforeUpgradeSpreadRewards)
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

	// Create two sets of all pools
	s.PrepareAllSupportedPools()
	s.PrepareAllSupportedPools()

	// Update authorized quote denoms
	concentratedParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	concentratedParams.AuthorizedQuoteDenoms = append(concentratedParams.AuthorizedQuoteDenoms, apptesting.USDC)
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, concentratedParams)

	// Create two more concentrated pools with positions
	nonMigratedPoolID := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx)
	migratedPoolID := nonMigratedPoolID + 1
	s.CreateConcentratedPoolsAndFullRangePosition([][]string{
		{"uion", "uosmo"},
		{apptesting.ETH, apptesting.USDC},
	})

	// Extract the position IDs for the last two pools
	migratedPoolPositionID := s.App.ConcentratedLiquidityKeeper.GetNextPositionId(s.Ctx) - 1
	nonMigratedPoolPositionID := migratedPoolPositionID - 1

	// Manually add some spread rewards to the migrated and non-migrated pools
	feeAccumulatorMigratedPool, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, migratedPoolID)
	s.Require().NoError(err)
	feeAccumulatorMigratedPool.AddToAccumulator(sdk.NewDecCoins(sdk.NewDecCoinFromDec("uosmo", sdk.MustNewDecFromStr("276701288297"))))

	feeAccumulatorNonMigratedPool, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, nonMigratedPoolID)
	s.Require().NoError(err)
	feeAccumulatorNonMigratedPool.AddToAccumulator(sdk.NewDecCoins(sdk.NewDecCoinFromDec("uosmo", sdk.MustNewDecFromStr("276701288297"))))

	// Migrated pool claim
	migratedPoolBeforeUpgradeSpreadRewards, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, migratedPoolPositionID)
	s.Require().NoError(err)
	s.Require().NotEmpty(migratedPoolBeforeUpgradeSpreadRewards)

	// Non-migrated pool claim
	nonMigratedPoolBeforeUpgradeSpreadRewards, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, nonMigratedPoolPositionID)
	s.Require().NoError(err)
	s.Require().NotEmpty(nonMigratedPoolBeforeUpgradeSpreadRewards)

	// Overwrite the migration list with the desired pool ID.
	oldMigrationList := concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDsV25
	concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDsV25 = map[uint64]struct{}{}
	concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDsV25[migratedPoolID] = struct{}{}

	return oldMigrationList, migratedPoolPositionID, migratedPoolBeforeUpgradeSpreadRewards, nonMigratedPoolBeforeUpgradeSpreadRewards
}

func (s *UpgradeTestSuite) ExecuteSpreadRewardsMigrationTest(oldMigrationList map[uint64]struct{}, lastPoolPositionID uint64, migratedPoolBeforeUpgradeSpreadRewards, nonMigratedPoolBeforeUpgradeSpreadRewards sdk.Coins) {
	// Migrated pool: ensure that the claimable spread rewards are the same before and after migration
	migratedPoolAfterUpgradeSpreadRewards, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, lastPoolPositionID)
	s.Require().NoError(err)
	s.Require().Equal(migratedPoolBeforeUpgradeSpreadRewards.String(), migratedPoolAfterUpgradeSpreadRewards.String())

	// Non-migrated pool: ensure that the claimable spread rewards are the same before and after migration
	nonMigratedPoolAfterUpgradeSpreadRewards, err := s.App.ConcentratedLiquidityKeeper.GetClaimableSpreadRewards(s.Ctx, lastPoolPositionID-1)
	s.Require().NoError(err)
	s.Require().Equal(nonMigratedPoolBeforeUpgradeSpreadRewards.String(), nonMigratedPoolAfterUpgradeSpreadRewards.String())
}
