package v31_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	v31 "github.com/osmosis-labs/osmosis/v30/app/upgrades/v31"
	gammtypes "github.com/osmosis-labs/osmosis/v30/x/gamm/types"
	superfuidtypes "github.com/osmosis-labs/osmosis/v30/x/superfluid/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v30/x/txfees/types"
)

const (
	v31UpgradeHeight = int64(10)
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpdateTakerFeeDistribution() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))

	// Prepare test by setting initial taker fee distribution
	s.PrepareTakerFeeDistributionTest()

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	// Verify the distribution was swapped correctly
	s.ExecuteTakerFeeDistributionTest()
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v31UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: v31.UpgradeName, Height: v31UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v31UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v31UpgradeHeight)
}

// PrepareTakerFeeDistributionTest sets up the initial state with taker fees going to community pool
func (s *UpgradeTestSuite) PrepareTakerFeeDistributionTest() {
	// Get current poolmanager parameters
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)

	osmoTakerFeeDistribution := poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution
	osmoTakerFeeDistribution.StakingRewards = osmomath.MustNewDecFromStr("0.3")
	osmoTakerFeeDistribution.CommunityPool = osmomath.MustNewDecFromStr("0.7")
	osmoTakerFeeDistribution.Burn = osmomath.MustNewDecFromStr("0.0")

	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution = osmoTakerFeeDistribution
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)
}

// ExecuteTakerFeeDistributionTest verifies that the community_pool and burn values were swapped
func (s *UpgradeTestSuite) ExecuteTakerFeeDistributionTest() {
	// Get poolmanager parameters after upgrade
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)

	// Verify OSMO taker fee distribution
	s.Require().Equal(osmomath.MustNewDecFromStr("0.3"), poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.StakingRewards)
	s.Require().Equal(osmomath.MustNewDecFromStr("0.0"), poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.CommunityPool)
	s.Require().Equal(osmomath.MustNewDecFromStr("0.7"), poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.Burn)

	// Verify that the OSMO total still sums to 1.0
	osmoTotal := poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.CommunityPool.
		Add(poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.Burn).
		Add(poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.StakingRewards)
	s.Require().Equal(osmomath.OneDec(), osmoTotal)

	// Verify non-OSMO taker fee distribution
	s.Require().Equal(osmomath.MustNewDecFromStr("0.225"), poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards)
	s.Require().Equal(osmomath.MustNewDecFromStr("0.525"), poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.Burn)
	s.Require().Equal(osmomath.MustNewDecFromStr("0.25"), poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool)

	// Verify that the non-OSMO total sums to 1.0
	nonOsmoTotal := poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool.
		Add(poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.Burn).
		Add(poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards)
	s.Require().Equal(osmomath.OneDec(), nonOsmoTotal)

	// Verify the module account is set correctly
	takerFeeBurnModuleAccount := s.App.AccountKeeper.GetModuleAccount(s.Ctx, txfeestypes.TakerFeeBurnName)
	s.Require().Equal(txfeestypes.TakerFeeBurnName, takerFeeBurnModuleAccount.GetName())
	s.Require().Nil(takerFeeBurnModuleAccount.GetPermissions())
}

func (s *UpgradeTestSuite) TestSuperfluidCleanup() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))

	// Prepare test by setting up some superfluid state
	s.PrepareSuperfluidCleanupTest()

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	// Verify all superfluid state was cleaned up
	s.ExecuteSuperfluidCleanupTest()
}

// PrepareSuperfluidCleanupTest sets up some superfluid state to be cleaned up
func (s *UpgradeTestSuite) PrepareSuperfluidCleanupTest() {
	// Create a superfluid asset
	denom := "gamm/pool/1"
	asset := superfuidtypes.SuperfluidAsset{
		Denom:     denom,
		AssetType: superfuidtypes.SuperfluidAssetTypeLPShare,
	}
	s.App.SuperfluidKeeper.SetSuperfluidAsset(s.Ctx, asset)

	// Create an intermediary account with a test validator address
	valAddr := sdk.ValAddress([]byte("testvaraddr1"))
	intermediaryAcc := superfuidtypes.NewSuperfluidIntermediaryAccount(denom, valAddr.String(), 1)
	s.App.SuperfluidKeeper.SetIntermediaryAccount(s.Ctx, intermediaryAcc)

	// Create a lock for testing
	lockOwner := s.TestAccs[0]
	coins := sdk.NewCoins(sdk.NewCoin(denom, osmomath.NewInt(1000000)))

	// Fund the account with the required coins
	err := s.App.BankKeeper.MintCoins(s.Ctx, gammtypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, gammtypes.ModuleName, lockOwner, coins)
	s.Require().NoError(err)

	// Create the lock (longer than unbonding period to allow synthetic locks)
	lock, err := s.App.LockupKeeper.CreateLock(s.Ctx, lockOwner, coins, time.Hour*24*21)
	s.Require().NoError(err)

	// Set lock-intermediary connection
	s.App.SuperfluidKeeper.SetLockIdIntermediaryAccountConnection(s.Ctx, lock.ID, intermediaryAcc)

	// Create synthetic lock for superbonding (active superfluid staking)
	stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	unbondingDuration := stakingParams.UnbondingTime

	stakingSynthDenom := denom + "/superbonding/" + valAddr.String()
	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, lock.ID, stakingSynthDenom, unbondingDuration, false)
	s.Require().NoError(err)

	// Create a second lock with superunbonding (undelegating superfluid stake)
	coins2 := sdk.NewCoins(sdk.NewCoin(denom, osmomath.NewInt(500000)))

	// Fund the account for the second lock
	err = s.App.BankKeeper.MintCoins(s.Ctx, gammtypes.ModuleName, coins2)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, gammtypes.ModuleName, lockOwner, coins2)
	s.Require().NoError(err)

	// Create the second lock (longer than unbonding period to allow synthetic locks)
	lock2, err := s.App.LockupKeeper.CreateLock(s.Ctx, lockOwner, coins2, time.Hour*24*21)
	s.Require().NoError(err)

	// Set lock-intermediary connection for second lock
	s.App.SuperfluidKeeper.SetLockIdIntermediaryAccountConnection(s.Ctx, lock2.ID, intermediaryAcc)

	// Create synthetic lock for superunbonding (undelegating superfluid stake)
	unstakingSynthDenom := denom + "/superunbonding/" + valAddr.String()
	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, lock2.ID, unstakingSynthDenom, unbondingDuration, true)
	s.Require().NoError(err)

	// Set an osmo equivalent multiplier
	s.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(s.Ctx, 1, denom, osmomath.NewDec(2))

	// Set unpool allowed pools
	s.App.SuperfluidKeeper.SetUnpoolAllowedPools(s.Ctx, []uint64{1, 2, 3})

	// Verify all state was set up correctly

	// Verify superfluid assets
	assets := s.App.SuperfluidKeeper.GetAllSuperfluidAssets(s.Ctx)
	s.Require().Equal(1, len(assets))
	s.Require().Equal(denom, assets[0].Denom)

	// Verify intermediary accounts
	intermediaryAccounts := s.App.SuperfluidKeeper.GetAllIntermediaryAccounts(s.Ctx)
	s.Require().Equal(1, len(intermediaryAccounts))
	s.Require().Equal(denom, intermediaryAccounts[0].Denom)
	s.Require().Equal(valAddr.String(), intermediaryAccounts[0].ValAddr)

	// Verify lock-intermediary connections
	connections := s.App.SuperfluidKeeper.GetAllLockIdIntermediaryAccountConnections(s.Ctx)
	s.Require().Equal(2, len(connections))
	lockIds := []uint64{connections[0].LockId, connections[1].LockId}
	s.Require().Contains(lockIds, lock.ID)
	s.Require().Contains(lockIds, lock2.ID)

	// Verify osmo equivalent multipliers
	multipliers := s.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(s.Ctx)
	s.Require().Equal(1, len(multipliers))
	s.Require().Equal(denom, multipliers[0].Denom)
	s.Require().Equal(osmomath.NewDec(2), multipliers[0].Multiplier)

	// Verify unpool allowed pools
	unpoolPools := s.App.SuperfluidKeeper.GetUnpoolAllowedPools(s.Ctx)
	s.Require().Equal(3, len(unpoolPools))
	s.Require().Equal([]uint64{1, 2, 3}, unpoolPools)

	// Verify the lock exists
	retrievedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
	s.Require().NoError(err)
	s.Require().Equal(lockOwner.String(), retrievedLock.Owner)
	s.Require().Equal(coins, retrievedLock.Coins)

	// Verify synthetic locks were created (both superbonding and superunbonding)
	allSynthLocks := s.App.LockupKeeper.GetAllSyntheticLockups(s.Ctx)
	s.Require().Equal(2, len(allSynthLocks))

	// Verify we have one superbonding and one superunbonding
	superbondingCount := 0
	superunbondingCount := 0
	for _, synthLock := range allSynthLocks {
		switch synthLock.SynthDenom {
		case stakingSynthDenom:
			superbondingCount++
			s.Require().Equal(lock.ID, synthLock.UnderlyingLockId)
			s.Require().False(synthLock.EndTime.After(s.Ctx.BlockTime()))
		case unstakingSynthDenom:
			superunbondingCount++
			s.Require().Equal(lock2.ID, synthLock.UnderlyingLockId)
			s.Require().True(synthLock.EndTime.After(s.Ctx.BlockTime()))
		}
	}
	s.Require().Equal(1, superbondingCount)
	s.Require().Equal(1, superunbondingCount)
}

// ExecuteSuperfluidCleanupTest verifies all superfluid state was cleaned up
func (s *UpgradeTestSuite) ExecuteSuperfluidCleanupTest() {
	// Verify all superfluid assets were deleted
	assets := s.App.SuperfluidKeeper.GetAllSuperfluidAssets(s.Ctx)
	s.Require().Equal(0, len(assets))

	// Verify all intermediary accounts were deleted
	intermediaryAccounts := s.App.SuperfluidKeeper.GetAllIntermediaryAccounts(s.Ctx)
	s.Require().Equal(0, len(intermediaryAccounts))

	// Verify all lock-intermediary connections were deleted
	connections := s.App.SuperfluidKeeper.GetAllLockIdIntermediaryAccountConnections(s.Ctx)
	s.Require().Equal(0, len(connections))

	// Verify all osmo equivalent multipliers were deleted
	multipliers := s.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(s.Ctx)
	s.Require().Equal(0, len(multipliers))

	// Verify unpool allowed pools were deleted
	unpoolPools := s.App.SuperfluidKeeper.GetUnpoolAllowedPools(s.Ctx)
	s.Require().Equal(0, len(unpoolPools))

	// Verify the underlying locks were unlocked and gamm tokens returned to delegator
	lockOwner := s.TestAccs[0]
	denom := "gamm/pool/1"

	// Check that account has no locked coins at all
	lockedCoins := s.App.LockupKeeper.GetAccountLockedCoins(s.Ctx, lockOwner)
	s.Require().True(lockedCoins.IsZero())

	// Verify gamm tokens were returned to delegator's wallet
	delegatorBalance := s.App.BankKeeper.GetBalance(s.Ctx, lockOwner, denom)
	expectedAmount := osmomath.NewInt(1000000).Add(osmomath.NewInt(500000)) // lock + lock2
	s.Require().Equal(expectedAmount, delegatorBalance.Amount)

	// Check that no synthetic locks with superfluid denoms exist
	allSyntheticLocks := s.App.LockupKeeper.GetAllSyntheticLockups(s.Ctx)
	for _, synthLock := range allSyntheticLocks {
		// Verify no superbonding or superunbonding synthetic locks exist
		s.Require().NotContains(synthLock.SynthDenom, "superbonding")
		s.Require().NotContains(synthLock.SynthDenom, "superunbonding")
	}

	// Verify superfluid module has no unexpected balance
	// (some dust might remain due to rounding, but should be minimal)
	moduleAddr := s.App.AccountKeeper.GetModuleAddress(superfuidtypes.ModuleName)
	moduleBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, moduleAddr)

	// Module should have zero balance
	s.Require().Len(moduleBalance, 0)

	// Verify no delegations exist from any address that looks like an intermediary account
	// Intermediary accounts follow a specific pattern, we check there are no delegations from them
	allDelegations, err := s.App.StakingKeeper.GetAllDelegations(s.Ctx)
	s.Require().NoError(err)

	for _, delegation := range allDelegations {
		delegatorAddr, err := sdk.AccAddressFromBech32(delegation.DelegatorAddress)
		s.Require().NoError(err)

		// Check if this looks like an intermediary account by trying to get it from the keeper
		// Since we deleted all intermediary accounts, GetIntermediaryAccount should return empty
		intermediaryAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, delegatorAddr)

		// If we get a non-empty intermediary account, it means there's orphaned state
		s.Require().Empty(intermediaryAcc.Denom, "Found orphaned delegation from intermediary account: %s", delegatorAddr.String())
	}
}
