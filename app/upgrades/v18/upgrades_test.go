package v18_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v17/app/apptesting"
	"github.com/osmosis-labs/osmosis/v17/app/keepers"
	v17 "github.com/osmosis-labs/osmosis/v17/app/upgrades/v17"
	superfluidtypes "github.com/osmosis-labs/osmosis/v17/x/superfluid/types"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
}

type ByLinkedClassicPool []v17.AssetPair

func (a ByLinkedClassicPool) Len() int      { return len(a) }
func (a ByLinkedClassicPool) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLinkedClassicPool) Less(i, j int) bool {
	return a[i].LinkedClassicPool < a[j].LinkedClassicPool
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const dummyUpgradeHeight = 5

func dummyUpgrade(suite *UpgradeTestSuite) {
	suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v18", Height: dummyUpgradeHeight}
	err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
	suite.Require().NoError(err)
	_, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
	suite.Require().True(exists)

	suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight)
}

func assertEqual(suite *UpgradeTestSuite, pre, post interface{}) {
	suite.Require().Equal(pre, post)
}

func (suite *UpgradeTestSuite) TestUpgrade() {

	testCases := []struct {
		name        string
		pre_upgrade func(*keepers.AppKeepers)
		upgrade     func(*keepers.AppKeepers)
	}{
		{
			"Test that the upgrade succeeds: mainnet",
			func(keepers *keepers.AppKeepers) {
				var lastPoolID uint64 // To keep track of the last assigned pool ID

				// Sort AssetPairs based on LinkedClassicPool values.
				// We sort both pairs because we use the test asset pairs to create initial state,
				// then use the actual asset pairs to verify the result is correct.
				sort.Sort(ByLinkedClassicPool(v17.AssetPairsForTestsOnly))
				sort.Sort(ByLinkedClassicPool(v17.AssetPairs))

				expectedCoinsUsedInUpgradeHandler := sdk.NewCoins()

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

					// 0.1 OSMO used to get the respective base asset amount, 0.1 OSMO used to create the position
					osmoIn := sdk.NewCoin(v17.OSMO, sdk.NewInt(100000).MulRaw(2))

					// Add the amount of osmo that will be used to the expectedCoinsUsedInUpgradeHandler.
					expectedCoinsUsedInUpgradeHandler = expectedCoinsUsedInUpgradeHandler.Add(osmoIn)

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

			},
			func(keepers *keepers.AppKeepers) {
				// Run upgrade handler.
				dummyUpgrade(suite)
				suite.Require().NotPanics(func() {
					suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
				})

				suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(time.Hour * 24))
			},
		},
		{
			"Fails because CFMM pool is not found",
			func(keepers *keepers.AppKeepers) {
			},
			func(keepers *keepers.AppKeepers) {
				dummyUpgrade(suite)
				suite.Require().Panics(func() {
					suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
				})
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.pre_upgrade(&suite.App.AppKeepers)
			tc.upgrade(&suite.App.AppKeepers)
		})
	}
}
