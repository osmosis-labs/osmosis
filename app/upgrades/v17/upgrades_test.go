package v17_test

import (
	"fmt"
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v17/app/apptesting"
	"github.com/osmosis-labs/osmosis/v17/app/keepers"
	v17 "github.com/osmosis-labs/osmosis/v17/app/upgrades/v17"
	cltypes "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v17/x/poolmanager/types"
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
	plan := upgradetypes.Plan{Name: "v17", Height: dummyUpgradeHeight}
	err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
	suite.Require().NoError(err)
	_, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
	suite.Require().True(exists)

	suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight)
}

func (suite *UpgradeTestSuite) TestUpgrade() {
	upgradeSetup := func() {
		// This is done to ensure that we run the InitGenesis() logic for the new modules
		upgradeStoreKey := suite.App.AppKeepers.GetKey(upgradetypes.StoreKey)
		store := suite.Ctx.KVStore(upgradeStoreKey)
		versionStore := prefix.NewStore(store, []byte{upgradetypes.VersionMapByte})
		versionStore.Delete([]byte(cltypes.ModuleName))
	}

	// Allow 0.1% margin of error.
	multiplicativeTolerance := osmomath.ErrTolerance{
		MultiplicativeTolerance: sdk.MustNewDecFromStr("0.001"),
	}

	testCases := []struct {
		name        string
		pre_upgrade func(sdk.Context, *keepers.AppKeepers) (sdk.Coins, uint64)
		upgrade     func(sdk.Context, *keepers.AppKeepers, sdk.Coins, uint64)
	}{
		{
			"Test that the upgrade succeeds",
			func(ctx sdk.Context, keepers *keepers.AppKeepers) (sdk.Coins, uint64) {
				upgradeSetup()

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
						poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, sdk.NewInt(10000000000)), sdk.NewCoin(v17.QuoteAsset, sdk.NewInt(10000000000)))
						suite.PrepareBalancerPoolWithCoins(poolCoins...)
						lastPoolID++
					}

					// Now create the pool with the correct pool ID.
					poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, sdk.NewInt(10000000000)), sdk.NewCoin(v17.QuoteAsset, sdk.NewInt(10000000000)))
					poolId := suite.PrepareBalancerPoolWithCoins(poolCoins...)

					// Send two of the base asset to the community pool.
					twoBaseAsset := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, sdk.NewInt(2000000)))
					suite.FundAcc(suite.TestAccs[0], twoBaseAsset)

					err := suite.App.DistrKeeper.FundCommunityPool(suite.Ctx, twoBaseAsset, suite.TestAccs[0])
					suite.Require().NoError(err)

					// Determine approx how much baseAsset will be used from community pool when 1 OSMO used.
					oneOsmo := sdk.NewCoin(v17.QuoteAsset, sdk.NewInt(1000000))
					pool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, poolId)
					suite.Require().NoError(err)
					respectiveBaseAsset, err := suite.App.GAMMKeeper.CalcOutAmtGivenIn(suite.Ctx, pool, oneOsmo, assetPair.BaseAsset, sdk.ZeroDec())
					suite.Require().NoError(err)

					// Add the amount of baseAsset that will be used to the expectedCoinsUsedInUpgradeHandler.
					expectedCoinsUsedInUpgradeHandler = expectedCoinsUsedInUpgradeHandler.Add(respectiveBaseAsset)

					// Update the lastPoolID to the current pool ID.
					lastPoolID = poolID
				}

				return expectedCoinsUsedInUpgradeHandler, lastPoolID

			},
			func(ctx sdk.Context, keepers *keepers.AppKeepers, expectedCoinsUsedInUpgradeHandler sdk.Coins, lastPoolID uint64) {
				stakingParams := suite.App.StakingKeeper.GetParams(suite.Ctx)
				stakingParams.BondDenom = "uosmo"
				suite.App.StakingKeeper.SetParams(suite.Ctx, stakingParams)

				// Retrieve the community pool balance before the upgrade
				communityPoolAddress := suite.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
				communityPoolBalancePre := suite.App.BankKeeper.GetAllBalances(suite.Ctx, communityPoolAddress)

				// Run upgrade handler.
				dummyUpgrade(suite)
				suite.Require().NotPanics(func() {
					suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
				})

				// Retrieve the community pool balance (and the feePool balance) after the upgrade
				communityPoolBalancePost := suite.App.BankKeeper.GetAllBalances(suite.Ctx, communityPoolAddress)
				feePoolCommunityPoolPost := suite.App.DistrKeeper.GetFeePool(suite.Ctx).CommunityPool

				assetPairs := v17.InitializeAssetPairs(ctx, keepers)

				for i, assetPair := range assetPairs {
					// Validate that the community pool balance has been reduced by the amount of baseAsset that was used to create the pool.
					suite.Require().Equal(communityPoolBalancePre.AmountOf(assetPair.BaseAsset).Sub(expectedCoinsUsedInUpgradeHandler.AmountOf(assetPair.BaseAsset)).String(), communityPoolBalancePost.AmountOf(assetPair.BaseAsset).String())

					// Validate that the fee pool community pool balance has been decreased by the amount of baseAsset that was used to create the pool.
					suite.Require().Equal(communityPoolBalancePost.AmountOf(assetPair.BaseAsset).String(), feePoolCommunityPoolPost.AmountOf(assetPair.BaseAsset).TruncateInt().String())

					// Get balancer pool's spot price.
					balancerSpotPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, assetPair.LinkedClassicPool, v17.QuoteAsset, assetPair.BaseAsset)
					suite.Require().NoError(err)

					// Validate CL pool was created.
					concentratedPool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, lastPoolID+1)
					suite.Require().NoError(err)
					suite.Require().Equal(poolmanagertypes.Concentrated, concentratedPool.GetType())

					// Validate that denom0 and denom1 were set correctly
					concentratedTypePool, ok := concentratedPool.(cltypes.ConcentratedPoolExtension)
					suite.Require().True(ok)
					suite.Require().Equal(assetPair.BaseAsset, concentratedTypePool.GetToken0())
					suite.Require().Equal(v17.QuoteAsset, concentratedTypePool.GetToken1())

					// Validate that the spot price of the CL pool is what we expect
					suite.Require().Equal(0, multiplicativeTolerance.CompareBigDec(concentratedTypePool.GetCurrentSqrtPrice().PowerInteger(2), osmomath.BigDecFromSDKDec(balancerSpotPrice)))

					// Validate that the link is correct.
					migrationInfo, err := suite.App.GAMMKeeper.GetAllMigrationInfo(suite.Ctx)
					link := migrationInfo.BalancerToConcentratedPoolLinks[i]
					suite.Require().Equal(assetPair.LinkedClassicPool, link.BalancerPoolId)
					suite.Require().Equal(concentratedPool.GetId(), link.ClPoolId)

					// Validate the sfs status
					clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(concentratedPool.GetId())
					_, err = suite.App.SuperfluidKeeper.GetSuperfluidAsset(suite.Ctx, clPoolDenom)
					if assetPair.Superfluid {
						suite.Require().NoError(err)
					} else {
						suite.Require().Error(err)
					}

					lastPoolID++
				}

				// Check osmo balance (was used in every pool creation)
				suite.Require().Equal(0, multiplicativeTolerance.Compare(communityPoolBalancePre.AmountOf(v17.QuoteAsset), communityPoolBalancePost.AmountOf(v17.QuoteAsset).Sub(expectedCoinsUsedInUpgradeHandler.AmountOf(v17.QuoteAsset))))
				suite.Require().Equal(communityPoolBalancePost.AmountOf(v17.QuoteAsset).String(), feePoolCommunityPoolPost.AmountOf(v17.QuoteAsset).TruncateInt().String())

				// Validate that all links were created.
				migrationInfo, err := suite.App.GAMMKeeper.GetAllMigrationInfo(suite.Ctx)
				suite.Require().Equal(len(assetPairs), len(migrationInfo.BalancerToConcentratedPoolLinks))
				suite.Require().NoError(err)

			},
		},
		{
			"Fails because CFMM pool is not found",
			func(ctx sdk.Context, keepers *keepers.AppKeepers) (sdk.Coins, uint64) {
				upgradeSetup()
				return sdk.NewCoins(), 0
			},
			func(ctx sdk.Context, keepers *keepers.AppKeepers, expectedCoinsUsedInUpgradeHandler sdk.Coins, lastPoolID uint64) {
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

			expectedCoinsUsedInUpgradeHandler, lastPoolID := tc.pre_upgrade(suite.Ctx, &suite.App.AppKeepers)
			tc.upgrade(suite.Ctx, &suite.App.AppKeepers, expectedCoinsUsedInUpgradeHandler, lastPoolID)
		})
	}
}
