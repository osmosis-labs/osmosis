package v25_test

import (
	"testing"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	auctiontypes "github.com/skip-mev/block-sdk/v2/x/auction/types"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	v4 "github.com/cosmos/cosmos-sdk/x/slashing/migrations/v4"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v28/app/apptesting"
	v25 "github.com/osmosis-labs/osmosis/v28/app/upgrades/v25"
	concentratedtypes "github.com/osmosis-labs/osmosis/v28/x/concentrated-liquidity/types"
)

const (
	v25UpgradeHeight = int64(10)
)

var (
	consAddr = sdk.ConsAddress(sdk.AccAddress([]byte("addr1_______________")))
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))

	// Setup spread factor migration test environment
	oldMigrationList, lastPoolPositionID, migratedPoolBeforeUpgradeSpreadRewards, nonMigratedPoolBeforeUpgradeSpreadRewards := s.PrepareSpreadRewardsMigrationTestEnv()
	preMigrationSigningInfo := s.prepareMissedBlocksCounterTest()

	// Check consensus params before upgrade
	consParamsPre, err := s.App.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEqual(consParamsPre.Evidence.MaxAgeDuration, v25.NewMaxAgeDuration)
	s.Require().NotEqual(consParamsPre.Evidence.MaxAgeNumBlocks, v25.NewMaxAgeNumBlocks)

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	// check auction module account
	acc := s.App.AccountKeeper.GetModuleAccount(s.Ctx, auctiontypes.ModuleName)
	s.Require().NotNil(acc)
	s.Require().Equal(auctiontypes.ModuleName, acc.GetName())
	s.Require().Equal(authtypes.NewModuleAddress(auctiontypes.ModuleName), acc.GetAddress())

	// check auction params
	params, err := s.App.AuctionKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)

	// Check auction params
	s.Require().Equal(params.MaxBundleSize, v25.AuctionParams.MaxBundleSize)
	s.Require().Equal(params.ReserveFee.Denom, v25.AuctionParams.ReserveFee.Denom)
	s.Require().Equal(params.ReserveFee.Amount.Int64(), v25.AuctionParams.ReserveFee.Amount.Int64())
	s.Require().Equal(params.MinBidIncrement.Denom, v25.AuctionParams.MinBidIncrement.Denom)
	s.Require().Equal(params.MinBidIncrement.Amount.Int64(), v25.AuctionParams.MinBidIncrement.Amount.Int64())
	s.Require().Equal(params.EscrowAccountAddress, v25.AuctionParams.EscrowAccountAddress)
	s.Require().Equal(params.FrontRunningProtection, v25.AuctionParams.FrontRunningProtection)
	s.Require().Equal(params.ProposerFee, v25.AuctionParams.ProposerFee)

	// Get smartaccount params
	smartAccountParams := s.App.SmartAccountKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)

	// Check smartaccount params
	s.Require().Equal(smartAccountParams.IsSmartAccountActive, v25.IsSmartAccountActive)
	s.Require().Equal(smartAccountParams.MaximumUnauthenticatedGas, v25.MaximumUnauthenticatedGas)
	s.Require().Equal(smartAccountParams.CircuitBreakerControllers[0], v25.CircuitBreakerController)

	// Check consensus params after upgrade
	consParamsPost, err := s.App.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(consParamsPost.Evidence.MaxAgeDuration, v25.NewMaxAgeDuration)
	s.Require().Equal(consParamsPost.Evidence.MaxAgeNumBlocks, v25.NewMaxAgeNumBlocks)

	s.ExecuteSpreadRewardsMigrationTest(oldMigrationList, lastPoolPositionID, migratedPoolBeforeUpgradeSpreadRewards, nonMigratedPoolBeforeUpgradeSpreadRewards)
	s.executeMissedBlocksCounterTest(preMigrationSigningInfo)
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v25UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: v25.Upgrade.UpgradeName, Height: v25UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v25UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v25UpgradeHeight)
}

func (s *UpgradeTestSuite) PrepareSpreadRewardsMigrationTestEnv() (map[uint64]struct{}, uint64, sdk.Coins, sdk.Coins) {
	// Set the migration pool ID threshold to far away to simulate pre-migration state.
	s.App.ConcentratedLiquidityKeeper.SetSpreadFactorPoolIDMigrationThreshold(s.Ctx, 1000)

	// Create two sets of all pools
	s.PrepareAllSupportedPools()
	s.PrepareAllSupportedPools()

	// Update authorized quote denoms
	concentratedParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	// concentratedParams.AuthorizedQuoteDenoms = append(concentratedParams.AuthorizedQuoteDenoms, apptesting.USDC)
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
	feeAccumulatorMigratedPool.AddToAccumulator(sdk.NewDecCoins(sdk.NewDecCoinFromDec("uosmo", osmomath.MustNewDecFromStr("276701288297"))))

	feeAccumulatorNonMigratedPool, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, nonMigratedPoolID)
	s.Require().NoError(err)
	feeAccumulatorNonMigratedPool.AddToAccumulator(sdk.NewDecCoins(sdk.NewDecCoinFromDec("uosmo", osmomath.MustNewDecFromStr("276701288297"))))

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

func (s *UpgradeTestSuite) prepareMissedBlocksCounterTest() slashingtypes.ValidatorSigningInfo {
	cdc := moduletestutil.MakeTestEncodingConfig(slashing.AppModuleBasic{}).Codec
	slashingStoreKey := s.App.AppKeepers.GetKey(slashingtypes.StoreKey)
	store := s.Ctx.KVStore(slashingStoreKey)

	// Replicate current mainnet state where, someones missed block counter is greater than their actual missed blocks
	preMigrationSigningInfo := slashingtypes.ValidatorSigningInfo{
		Address:             consAddr.String(),
		StartHeight:         10,
		IndexOffset:         100,
		JailedUntil:         time.Time{},
		Tombstoned:          false,
		MissedBlocksCounter: 1000,
	}

	// Set the missed blocks for the validator
	for i := 0; i < 10; i++ {
		err := s.App.SlashingKeeper.SetMissedBlockBitmapValue(s.Ctx, consAddr, int64(i), true)
		s.Require().NoError(err)
	}

	// Validate that the missed block bitmap value is of length 10 (the real missed blocks value), differing from the missed blocks counter of 1000 (the incorrect value)
	missedBlocks, err := s.App.SlashingKeeper.GetValidatorMissedBlocks(s.Ctx, consAddr)
	s.Require().NoError(err)
	s.Require().Len(missedBlocks, 10)

	// store old signing info and bitmap entries
	bz := cdc.MustMarshal(&preMigrationSigningInfo)
	store.Set(v4.ValidatorSigningInfoKey(consAddr), bz)

	return preMigrationSigningInfo
}

func (s *UpgradeTestSuite) executeMissedBlocksCounterTest(preMigrationSigningInfo slashingtypes.ValidatorSigningInfo) {
	postMigrationSigningInfo, err := s.App.SlashingKeeper.GetValidatorSigningInfo(s.Ctx, consAddr)
	s.Require().NoError(err)

	// Check that the missed blocks counter was set to the correct value
	s.Require().Equal(int64(10), postMigrationSigningInfo.MissedBlocksCounter)

	// Check that all other fields are the same
	s.Require().Equal(preMigrationSigningInfo.Address, postMigrationSigningInfo.Address)
	s.Require().Equal(preMigrationSigningInfo.StartHeight, postMigrationSigningInfo.StartHeight)
	s.Require().Equal(preMigrationSigningInfo.IndexOffset, postMigrationSigningInfo.IndexOffset)
	s.Require().Equal(preMigrationSigningInfo.JailedUntil, postMigrationSigningInfo.JailedUntil)
	s.Require().Equal(preMigrationSigningInfo.Tombstoned, postMigrationSigningInfo.Tombstoned)
}
