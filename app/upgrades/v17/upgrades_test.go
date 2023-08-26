package v17_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	"github.com/osmosis-labs/osmosis/v19/app/keepers"
	v17 "github.com/osmosis-labs/osmosis/v19/app/upgrades/v17"
	cltypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	poolManagerTypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v19/x/superfluid/types"
	"github.com/osmosis-labs/osmosis/v19/x/twap/types"
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

func dummyTwapRecord(poolId uint64, t time.Time, asset0 string, asset1 string, sp0, accum0, accum1, geomAccum sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		PoolId:      poolId,
		Time:        t,
		Asset0Denom: asset0,
		Asset1Denom: asset1,

		P0LastSpotPrice:             sp0,
		P1LastSpotPrice:             sdk.OneDec().Quo(sp0),
		P0ArithmeticTwapAccumulator: accum0,
		P1ArithmeticTwapAccumulator: accum1,
		GeometricTwapAccumulator:    geomAccum,
	}
}

func assertTwapFlipped(suite *UpgradeTestSuite, pre, post types.TwapRecord) {
	suite.Require().Equal(pre.Asset0Denom, post.Asset0Denom)
	suite.Require().Equal(pre.Asset1Denom, post.Asset1Denom)
	suite.Require().Equal(pre.P0LastSpotPrice, post.P1LastSpotPrice)
	suite.Require().Equal(pre.P1LastSpotPrice, post.P0LastSpotPrice)
}

func assertEqual(suite *UpgradeTestSuite, pre, post interface{}) {
	suite.Require().Equal(pre, post)
}

func (suite *UpgradeTestSuite) TestUpgrade() {
	// Allow 0.1% margin of error.
	multiplicativeTolerance := osmomath.ErrTolerance{
		MultiplicativeTolerance: sdk.MustNewDecFromStr("0.001"),
	}

	testCases := []struct {
		name        string
		pre_upgrade func(*keepers.AppKeepers) (sdk.Coins, uint64)
		upgrade     func(*keepers.AppKeepers, sdk.Coins, uint64)
	}{
		{
			"Test that the upgrade succeeds: mainnet",
			func(keepers *keepers.AppKeepers) (sdk.Coins, uint64) {
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
						poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, sdk.NewInt(10000000000)), sdk.NewCoin(assetPair.QuoteAsset, sdk.NewInt(10000000000)))
						suite.PrepareBalancerPoolWithCoins(poolCoins...)
						lastPoolID++
					}

					// Now create the pool with the correct pool ID.
					poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, sdk.NewInt(10000000000)), sdk.NewCoin(assetPair.QuoteAsset, sdk.NewInt(10000000000)))
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

				existingPool := suite.PrepareConcentratedPoolWithCoins("ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4", "uosmo")
				existingPool2 := suite.PrepareConcentratedPoolWithCoins("akash", "uosmo")
				existingBalancerPoolId := suite.PrepareBalancerPoolWithCoins(sdk.NewCoin("atom", sdk.NewInt(10000000000)), sdk.NewCoin("uosmo", sdk.NewInt(10000000000)))

				// create few TWAP records for the pools
				t1 := dummyTwapRecord(existingPool.GetId(), time.Now().Add(-time.Hour*24), "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4", "uosmo", sdk.NewDec(10),
					sdk.OneDec().MulInt64(10*10),
					sdk.OneDec().MulInt64(3),
					sdk.ZeroDec())

				t2 := dummyTwapRecord(existingPool.GetId(), time.Now().Add(-time.Hour*10), "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4", "uosmo", sdk.NewDec(30),
					sdk.OneDec().MulInt64(10*10+10),
					sdk.OneDec().MulInt64(5),
					sdk.ZeroDec())

				t3 := dummyTwapRecord(existingPool.GetId(), time.Now().Add(-time.Hour), "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4", "uosmo", sdk.NewDec(20),
					sdk.OneDec().MulInt64(10*10+10*5),
					sdk.OneDec().MulInt64(10),
					sdk.ZeroDec())

				t4 := dummyTwapRecord(existingPool2.GetId(), time.Now().Add(-time.Hour*24), "akash", "uosmo", sdk.NewDec(10),
					sdk.OneDec().MulInt64(10*10*10),
					sdk.OneDec().MulInt64(5),
					sdk.ZeroDec())

				t5 := dummyTwapRecord(existingPool2.GetId(), time.Now().Add(-time.Hour), "akash", "uosmo", sdk.NewDec(20),
					sdk.OneDec().MulInt64(10),
					sdk.OneDec().MulInt64(2),
					sdk.ZeroDec())

				t6 := dummyTwapRecord(existingBalancerPoolId, time.Now().Add(-time.Hour), "atom", "uosmo", sdk.NewDec(10),
					sdk.OneDec().MulInt64(10),
					sdk.OneDec().MulInt64(10),
					sdk.ZeroDec())

				t7 := dummyTwapRecord(existingBalancerPoolId, time.Now().Add(-time.Minute*20), "atom", "uosmo", sdk.NewDec(50),
					sdk.OneDec().MulInt64(10*5),
					sdk.OneDec().MulInt64(5),
					sdk.ZeroDec())

				// store TWAP records
				suite.App.TwapKeeper.StoreNewRecord(suite.Ctx, t1)
				suite.App.TwapKeeper.StoreNewRecord(suite.Ctx, t2)
				suite.App.TwapKeeper.StoreNewRecord(suite.Ctx, t3)
				suite.App.TwapKeeper.StoreNewRecord(suite.Ctx, t4)
				suite.App.TwapKeeper.StoreNewRecord(suite.Ctx, t5)
				suite.App.TwapKeeper.StoreNewRecord(suite.Ctx, t6)
				suite.App.TwapKeeper.StoreNewRecord(suite.Ctx, t7)

				return expectedCoinsUsedInUpgradeHandler, existingBalancerPoolId

			},
			func(keepers *keepers.AppKeepers, expectedCoinsUsedInUpgradeHandler sdk.Coins, lastPoolID uint64) {
				lastPoolIdMinusOne := lastPoolID - 1
				lastPoolIdMinusTwo := lastPoolID - 2
				stakingParams := suite.App.StakingKeeper.GetParams(suite.Ctx)
				stakingParams.BondDenom = "uosmo"
				suite.App.StakingKeeper.SetParams(suite.Ctx, stakingParams)

				// Retrieve the community pool balance before the upgrade
				communityPoolAddress := suite.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
				communityPoolBalancePre := suite.App.BankKeeper.GetAllBalances(suite.Ctx, communityPoolAddress)

				numPoolPreUpgrade := suite.App.PoolManagerKeeper.GetNextPoolId(suite.Ctx) - 1
				clPool1TwapRecordPreUpgrade, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(suite.Ctx, lastPoolIdMinusTwo)
				suite.Require().NoError(err)

				clPool1TwapRecordHistoricalPoolIndexPreUpgrade, err := keepers.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(suite.Ctx, lastPoolIdMinusTwo)
				suite.Require().NoError(err)

				clPool2TwapRecordPreUpgrade, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(suite.Ctx, lastPoolIdMinusOne)
				suite.Require().NoError(err)

				clPool2TwapRecordHistoricalPoolIndexPreUpgrade, err := keepers.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(suite.Ctx, lastPoolIdMinusOne)
				suite.Require().NoError(err)

				clPoolsTwapRecordHistoricalTimeIndexPreUpgrade, err := keepers.TwapKeeper.GetAllHistoricalTimeIndexedTWAPs(suite.Ctx)
				suite.Require().NoError(err)

				// Run upgrade handler.
				dummyUpgrade(suite)
				suite.Require().NotPanics(func() {
					suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
				})

				clPool1TwapRecordPostUpgrade, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(suite.Ctx, lastPoolIdMinusTwo)
				suite.Require().NoError(err)

				clPool1TwapRecordHistoricalPoolIndexPostUpgrade, err := keepers.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(suite.Ctx, lastPoolIdMinusTwo)
				suite.Require().NoError(err)

				clPool2TwapRecordPostUpgrade, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(suite.Ctx, lastPoolIdMinusOne)
				suite.Require().NoError(err)

				clPool2TwapRecordHistoricalPoolIndexPostUpgrade, err := keepers.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(suite.Ctx, lastPoolIdMinusOne)
				suite.Require().NoError(err)

				clPoolsTwapRecordHistoricalTimeIndexPostUpgrade, err := keepers.TwapKeeper.GetAllHistoricalTimeIndexedTWAPs(suite.Ctx)
				suite.Require().NoError(err)

				// check that all TWAP records aren't empty
				suite.Require().NotEmpty(clPool1TwapRecordPostUpgrade)
				suite.Require().NotEmpty(clPool1TwapRecordHistoricalPoolIndexPostUpgrade)
				suite.Require().NotEmpty(clPool2TwapRecordPostUpgrade)
				suite.Require().NotEmpty(clPool2TwapRecordHistoricalPoolIndexPostUpgrade)
				suite.Require().NotEmpty(clPoolsTwapRecordHistoricalTimeIndexPostUpgrade)

				for _, data := range []struct {
					pre, post []types.TwapRecord
				}{
					{clPool1TwapRecordPreUpgrade, clPool1TwapRecordPostUpgrade},
					{clPool1TwapRecordHistoricalPoolIndexPreUpgrade, clPool1TwapRecordHistoricalPoolIndexPostUpgrade},
					{clPool2TwapRecordPreUpgrade, clPool2TwapRecordPostUpgrade},
					{clPool2TwapRecordHistoricalPoolIndexPreUpgrade, clPool2TwapRecordHistoricalPoolIndexPostUpgrade},
				} {
					for i := range data.post {
						assertTwapFlipped(suite, data.pre[i], data.post[i])
					}
				}

				for i := range clPoolsTwapRecordHistoricalTimeIndexPostUpgrade {
					record := clPoolsTwapRecordHistoricalTimeIndexPostUpgrade[i]
					if record.PoolId == lastPoolIdMinusOne || record.PoolId == lastPoolIdMinusTwo {
						assertTwapFlipped(suite, clPoolsTwapRecordHistoricalTimeIndexPreUpgrade[i], record)
					} else if record.PoolId == lastPoolID {
						assertEqual(suite, clPoolsTwapRecordHistoricalTimeIndexPreUpgrade[i], record)
					}
				}

				// Retrieve the community pool balance (and the feePool balance) after the upgrade
				communityPoolBalancePost := suite.App.BankKeeper.GetAllBalances(suite.Ctx, communityPoolAddress)
				feePoolCommunityPoolPost := suite.App.DistrKeeper.GetFeePool(suite.Ctx).CommunityPool

				assetPairs, err := v17.InitializeAssetPairs(suite.Ctx, keepers)
				suite.Require().NoError(err)

				for i, assetPair := range assetPairs {
					// Get balancer pool's spot price.
					balancerSpotPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, assetPair.LinkedClassicPool, assetPair.QuoteAsset, assetPair.BaseAsset)
					suite.Require().NoError(err)

					// Validate CL pool was created.
					concentratedPool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, lastPoolID+1)
					suite.Require().NoError(err)
					suite.Require().Equal(poolmanagertypes.Concentrated, concentratedPool.GetType())

					// Validate that denom0 and denom1 were set correctly
					concentratedTypePool, ok := concentratedPool.(cltypes.ConcentratedPoolExtension)
					suite.Require().True(ok)
					suite.Require().Equal(assetPair.BaseAsset, concentratedTypePool.GetToken0())
					suite.Require().Equal(assetPair.QuoteAsset, concentratedTypePool.GetToken1())

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

				// Validate that the community pool balance has been reduced by the amount of osmo that was used to create the pool.
				suite.Require().Equal(communityPoolBalancePre.Sub(expectedCoinsUsedInUpgradeHandler).String(), communityPoolBalancePost.String())

				// Validate that the fee pool community pool balance has been decreased by the amount of osmo that was used to create the pool.
				suite.Require().Equal(sdk.NewDecCoinsFromCoins(communityPoolBalancePost...).String(), feePoolCommunityPoolPost.String())

				numPoolPostUpgrade := suite.App.PoolManagerKeeper.GetNextPoolId(suite.Ctx) - 1

				// Number of pools created should be equal to the number of records in the asset pairs.
				suite.Require().Equal(len(assetPairs), int(numPoolPostUpgrade-numPoolPreUpgrade))

				// Validate that all links were created.
				migrationInfo, err := suite.App.GAMMKeeper.GetAllMigrationInfo(suite.Ctx)
				suite.Require().Equal(len(assetPairs), len(migrationInfo.BalancerToConcentratedPoolLinks))
				suite.Require().NoError(err)

			},
		},
		{
			"Test that the upgrade succeeds: testnet",
			func(keepers *keepers.AppKeepers) (sdk.Coins, uint64) {
				suite.Ctx = suite.Ctx.WithChainID("osmo-test-5")

				var lastPoolID uint64 // To keep track of the last assigned pool ID

				sort.Sort(ByLinkedClassicPool(v17.AssetPairsForTestsOnly))
				sort.Sort(ByLinkedClassicPool(v17.AssetPairs))

				expectedCoinsUsedInUpgradeHandler := sdk.NewCoins()

				// Create earlier pools or dummy pools if needed
				for _, assetPair := range v17.AssetPairsForTestsOnly {
					poolID := assetPair.LinkedClassicPool

					// For testnet, we create a CL pool for ANY balancer pool.
					// The only thing we use the assetPair list here for to select some pools to enable superfluid for.
					for lastPoolID+1 < poolID {
						poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, sdk.NewInt(10000000000)), sdk.NewCoin(assetPair.QuoteAsset, sdk.NewInt(10000000000)))
						suite.PrepareBalancerPoolWithCoins(poolCoins...)

						// 0.1 OSMO used to get the respective base asset amount, 0.1 OSMO used to create the position
						osmoIn := sdk.NewCoin(v17.OSMO, sdk.NewInt(100000).MulRaw(2))

						// Add the amount of osmo that will be used to the expectedCoinsUsedInUpgradeHandler.
						expectedCoinsUsedInUpgradeHandler = expectedCoinsUsedInUpgradeHandler.Add(osmoIn)

						lastPoolID++
					}

					// Enable the GAMM pool for superfluid if the asset pair is marked as superfluid.
					if assetPair.Superfluid {
						poolShareDenom := fmt.Sprintf("gamm/pool/%d", assetPair.LinkedClassicPool)
						superfluidAsset := superfluidtypes.SuperfluidAsset{
							Denom:     poolShareDenom,
							AssetType: superfluidtypes.SuperfluidAssetTypeLPShare,
						}
						suite.App.SuperfluidKeeper.SetSuperfluidAsset(suite.Ctx, superfluidAsset)
					}
				}

				// We now create various pools that are not balancer pools.
				// This is to test if the testnet upgrade handler properly handles pools that are not of type balancer (i.e. should ignore them and move on).

				// Stableswap pool
				suite.CreatePoolFromType(poolmanagertypes.Stableswap)
				// Cosmwasm pool
				suite.CreatePoolFromType(poolmanagertypes.CosmWasm)
				// CL pool
				suite.CreatePoolFromType(poolmanagertypes.Concentrated)

				lastPoolID += 3

				return expectedCoinsUsedInUpgradeHandler, lastPoolID

			},
			func(keepers *keepers.AppKeepers, expectedCoinsUsedInUpgradeHandler sdk.Coins, lastPoolID uint64) {
				// Set the bond denom to uosmo
				stakingParams := suite.App.StakingKeeper.GetParams(suite.Ctx)
				stakingParams.BondDenom = "uosmo"
				suite.App.StakingKeeper.SetParams(suite.Ctx, stakingParams)

				// Retrieve the community pool balance before the upgrade
				communityPoolAddress := suite.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
				communityPoolBalancePre := suite.App.BankKeeper.GetAllBalances(suite.Ctx, communityPoolAddress)

				numPoolPreUpgrade := suite.App.PoolManagerKeeper.GetNextPoolId(suite.Ctx) - 1

				gammPoolsPreUpgrade, err := suite.App.GAMMKeeper.GetPools(suite.Ctx)
				suite.Require().NoError(err)

				// Run upgrade handler.
				dummyUpgrade(suite)
				suite.Require().NotPanics(func() {
					suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
				})

				// Retrieve the community pool balance (and the feePool balance) after the upgrade
				communityPoolBalancePost := suite.App.BankKeeper.GetAllBalances(suite.Ctx, communityPoolAddress)
				feePoolCommunityPoolPost := suite.App.DistrKeeper.GetFeePool(suite.Ctx).CommunityPool

				indexOffset := int(0)
				assetListIndex := int(0)

				// For testnet, we run through all gamm pools (not just the asset list)
				for i, pool := range gammPoolsPreUpgrade {
					// Skip pools that are not balancer pools
					if pool.GetType() != poolManagerTypes.Balancer {
						indexOffset++
						continue
					}

					gammPoolId := pool.GetId()
					cfmmPool, err := keepers.GAMMKeeper.GetCFMMPool(suite.Ctx, gammPoolId)
					suite.Require().NoError(err)

					poolCoins := cfmmPool.GetTotalPoolLiquidity(suite.Ctx)

					// Retrieve quoteAsset and baseAsset from the poolCoins
					quoteAsset, baseAsset := "", ""
					for _, coin := range poolCoins {
						if coin.Denom == v17.OSMO {
							quoteAsset = coin.Denom
						} else {
							baseAsset = coin.Denom
						}
					}
					if quoteAsset == "" || baseAsset == "" {
						indexOffset++
						continue
					}

					// Get balancer pool's spot price.
					balancerSpotPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, gammPoolId, quoteAsset, baseAsset)
					suite.Require().NoError(err)

					// Validate CL pool was created.
					concentratedPool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, lastPoolID+1)
					suite.Require().NoError(err)
					suite.Require().Equal(poolmanagertypes.Concentrated, concentratedPool.GetType())

					// Validate that denom0 and denom1 were set correctly
					concentratedTypePool, ok := concentratedPool.(cltypes.ConcentratedPoolExtension)
					suite.Require().True(ok)
					suite.Require().Equal(baseAsset, concentratedTypePool.GetToken0())
					suite.Require().Equal(quoteAsset, concentratedTypePool.GetToken1())

					// Validate that the spot price of the CL pool is what we expect
					suite.Require().Equal(0, multiplicativeTolerance.CompareBigDec(concentratedTypePool.GetCurrentSqrtPrice().PowerInteger(2), osmomath.BigDecFromSDKDec(balancerSpotPrice)))

					// Validate that the link is correct.
					migrationInfo, err := suite.App.GAMMKeeper.GetAllMigrationInfo(suite.Ctx)
					link := migrationInfo.BalancerToConcentratedPoolLinks[i-indexOffset]
					suite.Require().Equal(gammPoolId, link.BalancerPoolId)
					suite.Require().Equal(concentratedPool.GetId(), link.ClPoolId)

					// Validate the sfs status.
					// If the poolId matches a poolId on that asset list that had superfluid enabled, this pool should also be superfluid enabled.
					// Otherwise, it should not be superfluid enabled.
					assetListPoolId := v17.AssetPairsForTestsOnly[assetListIndex].LinkedClassicPool
					clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(concentratedPool.GetId())
					_, err = suite.App.SuperfluidKeeper.GetSuperfluidAsset(suite.Ctx, clPoolDenom)
					if assetListPoolId == gammPoolId {
						suite.Require().NoError(err)
						assetListIndex++
						for assetListIndex < len(v17.AssetPairsForTestsOnly)-1 && v17.AssetPairsForTestsOnly[assetListIndex].Superfluid == false {
							assetListIndex++
						}
					} else {
						suite.Require().Error(err)
					}

					lastPoolID++
				}

				// Validate that the community pool balance has been reduced by the amount of osmo that was used to create the pool.
				suite.Require().Equal(communityPoolBalancePre.Sub(expectedCoinsUsedInUpgradeHandler).String(), communityPoolBalancePost.String())

				// Validate that the fee pool community pool balance has been decreased by the amount of osmo that was used to create the pool.
				suite.Require().Equal(sdk.NewDecCoinsFromCoins(communityPoolBalancePost...).String(), feePoolCommunityPoolPost.String())

				numPoolPostUpgrade := suite.App.PoolManagerKeeper.GetNextPoolId(suite.Ctx) - 1
				numPoolsCreated := numPoolPostUpgrade - numPoolPreUpgrade

				// Number of pools created should be equal to the number of pools preUpgrade minus the number of pools that were not eligible for migration.
				numPoolsEligibleForMigration := numPoolPreUpgrade - 3
				suite.Require().Equal(int(numPoolsEligibleForMigration), int(numPoolsCreated))

				// Validate that all links were created.
				migrationInfo, err := suite.App.GAMMKeeper.GetAllMigrationInfo(suite.Ctx)
				suite.Require().Equal(int(numPoolsEligibleForMigration), len(migrationInfo.BalancerToConcentratedPoolLinks))
				suite.Require().NoError(err)

			},
		},
		{
			"Fails because CFMM pool is not found",
			func(keepers *keepers.AppKeepers) (sdk.Coins, uint64) {
				return sdk.NewCoins(), 0
			},
			func(keepers *keepers.AppKeepers, expectedCoinsUsedInUpgradeHandler sdk.Coins, lastPoolID uint64) {
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

			expectedCoinsUsedInUpgradeHandler, lastPoolID := tc.pre_upgrade(&suite.App.AppKeepers)
			tc.upgrade(&suite.App.AppKeepers, expectedCoinsUsedInUpgradeHandler, lastPoolID)
		})
	}
}
