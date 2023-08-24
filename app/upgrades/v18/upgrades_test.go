package v18_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v17/app/apptesting"
	v17 "github.com/osmosis-labs/osmosis/v17/app/upgrades/v17"

	lockuptypes "github.com/osmosis-labs/osmosis/v17/x/lockup/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v17/x/protorev/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v17/x/superfluid/types"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const (
	dummyUpgradeHeight = 5
	// this would be the amount in the lock that would stay locked during upgrades
	shareStaysLocked = 10000
)

func assertEqual(suite *UpgradeTestSuite, pre, post interface{}) {
	suite.Require().Equal(pre, post)
}

func (suite *UpgradeTestSuite) TestUpgrade() {
	// set up pools first to match v17 state(including linked cl pools)
	suite.setupPoolsToMainnetState()

	// corrupt state to match mainnet state
	suite.setupCorruptedState()

	// upgrade software
	suite.imitateUpgrade()
	suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(time.Hour * 24))

	// after the accum values have been resetted correctly after upgrade, we expect the accumulator store to be initialized with the correct value,
	// which in our test case would be 10000(the amount that was locked)
	valueAfterClear := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "gamm/pool/3",
		Duration:      time.Hour * 24 * 14,
	})
	valueAfterClear.Equal(sdk.NewInt(shareStaysLocked))
}

func (suite *UpgradeTestSuite) imitateUpgrade() {
	suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v18", Height: dummyUpgradeHeight}
	err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
	suite.Require().NoError(err)
	_, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
	suite.Require().True(exists)

	suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight)
}

// first set up pool state to mainnet state
func (suite *UpgradeTestSuite) setupPoolsToMainnetState() {
	var lastPoolID uint64 // To keep track of the last assigned pool ID

	// Sort AssetPairs based on LinkedClassicPool values.
	// We sort both pairs because we use the test asset pairs to create initial state,
	// then use the actual asset pairs to verify the result is correct.
	sort.Sort(ByLinkedClassicPool(v17.AssetPairsForTestsOnly))
	sort.Sort(ByLinkedClassicPool(v17.AssetPairs))

	// Create earlier pools or dummy pools if needed
	for _, assetPair := range v17.AssetPairsForTestsOnly {
		poolID := assetPair.LinkedClassicPool

		// If LinkedClassicPool is specified, but it's smaller than the current pool ID,
		// create dummy pools to fill the gap.
		for lastPoolID+1 < poolID {
			poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, sdk.NewInt(100000000000)), sdk.NewCoin(assetPair.QuoteAsset, sdk.NewInt(100000000000)))
			suite.PrepareBalancerPoolWithCoins(poolCoins...)
			lastPoolID++
		}

		// Now create the pool with the correct pool ID.
		poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, sdk.NewInt(100000000000)), sdk.NewCoin(assetPair.QuoteAsset, sdk.NewInt(100000000000)))
		suite.PrepareBalancerPoolWithCoins(poolCoins...)

		// Enable the GAMM pool for superfluid if the record says so.
		if assetPair.Superfluid {
			poolShareDenom := fmt.Sprintf("gamm/pool/%d", assetPair.LinkedClassicPool)
			superfluidAsset := superfluidtypes.SuperfluidAsset{
				Denom:     poolShareDenom,
				AssetType: superfluidtypes.SuperfluidAssetTypeLPShare,
			}
			suite.App.SuperfluidKeeper.SetSuperfluidAsset(suite.Ctx, superfluidAsset)
		}

		// Update the lastPoolID to the current pool ID.
		lastPoolID = poolID
	}
}

// setupCorruptedState aligns the testing environment with the mainnet state.
// By running this method, it will modify the lockup accumulator to be deleted which has happended in v4.0.0 upgrade.
// In this method, we join pool 3, then delete denom accum store in the lockup module to have the testing environment
// in the correct state.
func (s *UpgradeTestSuite) setupCorruptedState() {
	pool3Denom := "gamm/pool/3"

	// join pool, create lock
	addr, err := sdk.AccAddressFromBech32("osmo1urn0pnx8fl5kt89r5nzqd8htruq7skadc2xdk3")
	s.Require().NoError(err)
	keepers := &s.App.AppKeepers
	err = keepers.BankKeeper.MintCoins(s.Ctx, protorevtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(v17.OSMO, sdk.NewInt(50000000000))))
	s.Require().NoError(err)
	err = keepers.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, protorevtypes.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(v17.OSMO, sdk.NewInt(50000000000))))
	s.Require().NoError(err)
	aktGAMMPool, err := keepers.GAMMKeeper.GetPool(s.Ctx, 3)
	s.Require().NoError(err)
	sharesOut, err := keepers.GAMMKeeper.JoinSwapExactAmountIn(s.Ctx, addr, aktGAMMPool.GetId(), sdk.NewCoins(sdk.NewCoin(v17.OSMO, sdk.NewInt(50000000000))), sdk.ZeroInt())
	s.Require().NoError(err)
	aktSharesDenom := fmt.Sprintf("gamm/pool/%d", aktGAMMPool.GetId())
	shareCoins := sdk.NewCoins(sdk.NewCoin(aktSharesDenom, sharesOut))
	lock, err := keepers.LockupKeeper.CreateLock(s.Ctx, addr, shareCoins, time.Hour*24*14)
	s.Require().NoError(err)

	// also create a lock with the shares that would stay locked during the upgrade.
	// doing this would help us assert if the accumulator has been resetted to the correct value.
	shareCoinsStaysLocked := sdk.NewCoins(sdk.NewCoin(aktSharesDenom, sdk.NewInt(shareStaysLocked)))
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

type ByLinkedClassicPool []v17.AssetPair

func (a ByLinkedClassicPool) Len() int      { return len(a) }
func (a ByLinkedClassicPool) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLinkedClassicPool) Less(i, j int) bool {
	return a[i].LinkedClassicPool < a[j].LinkedClassicPool
}
