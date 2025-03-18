package v18_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v29/app/apptesting"
	v17 "github.com/osmosis-labs/osmosis/v29/app/upgrades/v17"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	appparams "github.com/osmosis-labs/osmosis/v29/app/params"
	gammmigration "github.com/osmosis-labs/osmosis/v29/x/gamm/types/migration"
	lockuptypes "github.com/osmosis-labs/osmosis/v29/x/lockup/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v29/x/protorev/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v29/x/superfluid/types"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const (
	dummyUpgradeHeight = 5
	// this would be the amount in the lock that would stay locked during upgrades
	shareStaysLocked = 10000
)

func (s *UpgradeTestSuite) TestUpgrade() {
	s.T().Skip("This test is skipped because it is not relevant anymore. It is kept here for reference.")

	// set up pools first to match v17 state(including linked cl pools)
	s.setupPoolsToMainnetState()

	// corrupt state to match mainnet state
	s.setupCorruptedState()

	// with the corrupted state, distribution used to panic in the `AfterEpochEnd` hook,
	// specifically from the one from incentives keeper.
	// This method ensures that with the corrupted state, we have the same state where
	// distribution would fail.
	s.ensurePreUpgradeDistributionPanics()

	migrationInfo, err := s.App.GAMMKeeper.GetAllMigrationInfo(s.Ctx)
	s.Require().NoError(err)

	link := migrationInfo.BalancerToConcentratedPoolLinks[0]
	s.Require().Equal(uint64(3), link.BalancerPoolId)

	clPoolId := link.ClPoolId

	pool, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, clPoolId)
	s.Require().NoError(err)

	// LP Fails before the upgrade
	lpTokens := sdk.NewCoins(sdk.NewCoin(pool.GetToken0(), osmomath.NewInt(1_000_000)), sdk.NewCoin(pool.GetToken1(), osmomath.NewInt(1_000_000)))
	s.FundAcc(s.TestAccs[0], lpTokens)
	// require a panic
	s.Require().Panics(func() {
		_, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPoolId, s.TestAccs[0], lpTokens)
	})

	// upgrade software
	s.imitateUpgrade()
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24))

	// after the accum values have been reset correctly after upgrade, we expect the accumulator store to be initialized with the correct value,
	// which in our test case would be 10000(the amount that was locked)
	valueAfterClear := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "gamm/pool/3",
		Duration:      time.Hour * 24 * 14,
	})
	valueAfterClear.Equal(osmomath.NewInt(shareStaysLocked))

	s.ensurePostUpgradeDistributionWorks()

	// Elapse time so that incentive distribution is triggered.
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour))

	// Check that can LP and swap into pool 3 with no usses
	// LP
	_, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPoolId, s.TestAccs[0], lpTokens)
	s.Require().NoError(err)

	// Refetch CL Pool
	updatedCLPool, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, clPoolId)
	s.Require().NoError(err)

	// Swap
	toSwap := sdk.NewCoin(pool.GetToken0(), osmomath.NewInt(100))
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(toSwap))
	_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], updatedCLPool, toSwap, pool.GetToken1(), osmomath.NewInt(1), osmomath.ZeroDec())
	s.Require().NoError(err)
}

func (s *UpgradeTestSuite) imitateUpgrade() {
	s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v18", Height: dummyUpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: dummyUpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(dummyUpgradeHeight)
}

// first set up pool state to mainnet state
func (s *UpgradeTestSuite) setupPoolsToMainnetState() {
	var lastPoolID uint64 // To keep track of the last assigned pool ID

	// Sort AssetPairs based on LinkedClassicPool values.
	// We sort both pairs because we use the test asset pairs to create initial state,
	// then use the actual asset pairs to verify the result is correct.
	sort.Sort(ByLinkedClassicPool(v17.AssetPairsForTestsOnly))

	// Create earlier pools or dummy pools if needed
	for _, assetPair := range v17.AssetPairsForTestsOnly {
		poolID := assetPair.LinkedClassicPool

		// If LinkedClassicPool is specified, but it's smaller than the current pool ID,
		// create dummy pools to fill the gap.
		for lastPoolID+1 < poolID {
			poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, osmomath.NewInt(100000000000)), sdk.NewCoin(assetPair.QuoteAsset, osmomath.NewInt(100000000000)))
			s.PrepareBalancerPoolWithCoins(poolCoins...)
			lastPoolID++
		}

		// Now create the pool with the correct pool ID.
		poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, osmomath.NewInt(100000000000)), sdk.NewCoin(assetPair.QuoteAsset, osmomath.NewInt(100000000000)))
		s.PrepareBalancerPoolWithCoins(poolCoins...)

		// Enable the GAMM pool for superfluid if the record says so.
		if assetPair.Superfluid {
			poolShareDenom := fmt.Sprintf("gamm/pool/%d", assetPair.LinkedClassicPool)
			superfluidAsset := superfluidtypes.SuperfluidAsset{
				Denom:     poolShareDenom,
				AssetType: superfluidtypes.SuperfluidAssetTypeLPShare,
			}
			s.App.SuperfluidKeeper.SetSuperfluidAsset(s.Ctx, superfluidAsset)
		}

		// Update the lastPoolID to the current pool ID.
		lastPoolID = poolID
	}
}

// setupCorruptedState aligns the testing environment with the mainnet state.
// By running this method, it will modify the lockup accumulator to be deleted which has happened in v4.0.0 upgrade.
// In this method, we join pool 3, then delete denom accum store in the lockup module to have the testing environment
// in the correct state.
func (s *UpgradeTestSuite) setupCorruptedState() {
	pool3Denom := "gamm/pool/3"

	// join pool, create lock
	addr, err := sdk.AccAddressFromBech32("osmo1urn0pnx8fl5kt89r5nzqd8htruq7skadc2xdk3")
	s.Require().NoError(err)
	keepers := &s.App.AppKeepers
	err = keepers.BankKeeper.MintCoins(s.Ctx, protorevtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(v17.OSMO, osmomath.NewInt(50000000000))))
	s.Require().NoError(err)
	err = keepers.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, protorevtypes.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(v17.OSMO, osmomath.NewInt(50000000000))))
	s.Require().NoError(err)
	aktGAMMPool, err := keepers.GAMMKeeper.GetPool(s.Ctx, 3)
	s.Require().NoError(err)
	sharesOut, err := keepers.GAMMKeeper.JoinSwapExactAmountIn(s.Ctx, addr, aktGAMMPool.GetId(), sdk.NewCoins(sdk.NewCoin(v17.OSMO, osmomath.NewInt(50000000000))), osmomath.ZeroInt())
	s.Require().NoError(err)
	aktSharesDenom := fmt.Sprintf("gamm/pool/%d", aktGAMMPool.GetId())
	shareCoins := sdk.NewCoins(sdk.NewCoin(aktSharesDenom, sharesOut))
	lock, err := keepers.LockupKeeper.CreateLock(s.Ctx, addr, shareCoins, time.Hour*24*14)
	s.Require().NoError(err)

	// also create a lock with the shares that would stay locked during the upgrade.
	// doing this would help us assert if the accumulator has been reset to the correct value.
	shareCoinsStaysLocked := sdk.NewCoins(sdk.NewCoin(aktSharesDenom, osmomath.NewInt(shareStaysLocked)))
	s.FundAcc(addr, shareCoinsStaysLocked)
	_, err = keepers.LockupKeeper.CreateLock(s.Ctx, addr, shareCoinsStaysLocked, time.Hour*24*14)
	s.Require().NoError(err)

	// get value before clearing denom accum store, this should be in positive value
	valueBeforeClear := keepers.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "gamm/pool/3",
		Duration:      time.Hour * 24 * 14,
	})

	// this should be a positive value
	s.Require().True(!valueBeforeClear.IsNegative())

	// Clear gamm/pool/3 denom accumulation store
	s.clearDenomAccumulationStore(pool3Denom)
	// Remove the lockup created for pool 3 above to get negative amount of accum value
	err = keepers.LockupKeeper.ForceUnlock(s.Ctx, lock)
	s.Require().NoError(err)

	valueAfterClear := keepers.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "gamm/pool/3",
		Duration:      time.Hour * 24 * 14,
	})

	s.Require().True(valueAfterClear.IsNegative())
	s.Require().True(shareCoins[0].Amount.Neg().Equal(valueAfterClear))
}

// We want to ensure that with the corrupted state of the lockup accumulator,
// `AfterEpochEnd` was panicking.
// We can do this check by creating a CL pool, then trying to distribute using that specific
// CL pool gauge. If our test setup was correct, this should panic.
func (s *UpgradeTestSuite) ensurePreUpgradeDistributionPanics() {
	epochInfo := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx)

	// add pool 3 denom (AKT) ti authorized quote denom param.
	clParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	// authorizedQuoteDenom := append(clParams.AuthorizedQuoteDenoms, v17.AKTIBCDenom)
	// clParams.AuthorizedQuoteDenoms = authorizedQuoteDenom
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, clParams)

	// prepare CL pool with the same denom as pool 3, which is the pool we are testing with
	clPool := s.PrepareConcentratedPoolWithCoins(v17.AKTIBCDenom, v17.OSMO)
	balancerToCLPoolLink := []gammmigration.BalancerToConcentratedPoolLink{
		{
			BalancerPoolId: 3,
			ClPoolId:       clPool.GetId(),
		},
	}

	// set migration record between the new CL pool and the old pool(pool number 3)
	migrationInfo := gammmigration.MigrationRecords{
		BalancerToConcentratedPoolLinks: balancerToCLPoolLink,
	}
	s.App.GAMMKeeper.SetMigrationRecords(s.Ctx, migrationInfo)

	// add new coins to the CL pool gauge so that it would be distributed after epoch ends then trigger panic
	coinsToAdd := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000)))
	gagueId, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, clPool.GetId(), epochInfo.Duration)
	s.Require().NoError(err)
	gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gagueId)
	s.Require().NoError(err)

	addr := sdk.AccAddress([]byte("addrx---------------"))
	s.FundAcc(addr, coinsToAdd)
	err = s.App.IncentivesKeeper.AddToGaugeRewards(s.Ctx, addr, coinsToAdd, gauge.Id)
	s.Require().NoError(err)

	// add block time so that rewards get distributed
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 25))
	s.BeginNewBlock(false)
	s.App.EpochsKeeper.BeforeEpochStart(s.Ctx, epochInfo.GetIdentifier(), 1)

	s.Require().Panics(func() {
		err = s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, epochInfo.GetIdentifier(), 1)
		s.Require().NoError(err)
	})
}

// clearDenomAccumulationStore clears denom accumulation store in the lockup keeper,
// this was cleared in v4.0.0 upgrade.
// Creating raw pools would re-initialize these pools, thus to properly imitate mainnet state,
// we need to manually delete this again.
func (s *UpgradeTestSuite) clearDenomAccumulationStore(denom string) {
	// Get Prefix
	capacity := len(lockuptypes.KeyPrefixLockAccumulation) + len(denom) + 1
	res := make([]byte, len(lockuptypes.KeyPrefixLockAccumulation), capacity)
	copy(res, lockuptypes.KeyPrefixLockAccumulation)
	res = append(res, []byte(denom+"/")...)

	lockupTypesStoreKeys := s.App.AppKeepers.GetKey(lockuptypes.StoreKey)
	store := prefix.NewStore(s.Ctx.KVStore(lockupTypesStoreKeys), res)
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

func (s *UpgradeTestSuite) ensurePostUpgradeDistributionWorks() {
	epochInfo := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx)

	// add block time so that rewards get distributed
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 25))
	s.BeginNewBlock(false)
	s.App.EpochsKeeper.BeforeEpochStart(s.Ctx, epochInfo.GetIdentifier(), 1)

	s.Require().NotPanics(func() {
		err := s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, epochInfo.GetIdentifier(), 1)
		s.Require().NoError(err)
	})
}

type ByLinkedClassicPool []v17.AssetPair

func (a ByLinkedClassicPool) Len() int      { return len(a) }
func (a ByLinkedClassicPool) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLinkedClassicPool) Less(i, j int) bool {
	return a[i].LinkedClassicPool < a[j].LinkedClassicPool
}
