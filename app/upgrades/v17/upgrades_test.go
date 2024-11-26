package v17_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/suite"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/app/keepers"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	v17 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v17"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))
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

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v17", Height: dummyUpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: dummyUpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(dummyUpgradeHeight)
}

func dummyTwapRecord(poolId uint64, t time.Time, asset0 string, asset1 string, sp0, accum0, accum1, geomAccum osmomath.Dec) types.TwapRecord { //nolint:unparam // asset1 always receives "usomo"
	return types.TwapRecord{
		PoolId:      poolId,
		Time:        t,
		Asset0Denom: asset0,
		Asset1Denom: asset1,

		P0LastSpotPrice:             sp0,
		P1LastSpotPrice:             osmomath.OneDec().Quo(sp0),
		P0ArithmeticTwapAccumulator: accum0,
		P1ArithmeticTwapAccumulator: accum1,
		GeometricTwapAccumulator:    geomAccum,
	}
}

func assertTwapFlipped(s *UpgradeTestSuite, pre, post types.TwapRecord) {
	s.Require().Equal(pre.Asset0Denom, post.Asset0Denom)
	s.Require().Equal(pre.Asset1Denom, post.Asset1Denom)
	s.Require().Equal(pre.P0LastSpotPrice, post.P1LastSpotPrice)
	s.Require().Equal(pre.P1LastSpotPrice, post.P0LastSpotPrice)
}

func assertEqual(s *UpgradeTestSuite, pre, post interface{}) {
	s.Require().Equal(pre, post)
}

func (s *UpgradeTestSuite) TestUpgrade() {
	// Allow 0.1% margin of error.
	multiplicativeTolerance := osmomath.ErrTolerance{
		MultiplicativeTolerance: osmomath.MustNewDecFromStr("0.001"),
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
						poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, osmomath.NewInt(10000000000)), sdk.NewCoin(assetPair.QuoteAsset, osmomath.NewInt(10000000000)))
						s.PrepareBalancerPoolWithCoins(poolCoins...)
						lastPoolID++
					}

					// Now create the pool with the correct pool ID.
					poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, osmomath.NewInt(10000000000)), sdk.NewCoin(assetPair.QuoteAsset, osmomath.NewInt(10000000000)))
					s.PrepareBalancerPoolWithCoins(poolCoins...)

					// 0.1 OSMO used to get the respective base asset amount, 0.1 OSMO used to create the position
					osmoIn := sdk.NewCoin(v17.OSMO, osmomath.NewInt(100000).MulRaw(2))

					// Add the amount of osmo that will be used to the expectedCoinsUsedInUpgradeHandler.
					expectedCoinsUsedInUpgradeHandler = expectedCoinsUsedInUpgradeHandler.Add(osmoIn)

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

				existingPool := s.PrepareConcentratedPoolWithCoins("ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4", appparams.BaseCoinUnit)
				existingPool2 := s.PrepareConcentratedPoolWithCoins("akash", appparams.BaseCoinUnit)
				existingBalancerPoolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin("atom", osmomath.NewInt(10000000000)), sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10000000000)))

				// create few TWAP records for the pools
				t1 := dummyTwapRecord(existingPool.GetId(), time.Now().Add(-time.Hour*24), "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4", appparams.BaseCoinUnit, osmomath.NewDec(10),
					osmomath.OneDec().MulInt64(10*10),
					osmomath.OneDec().MulInt64(3),
					osmomath.ZeroDec())

				t2 := dummyTwapRecord(existingPool.GetId(), time.Now().Add(-time.Hour*10), "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4", appparams.BaseCoinUnit, osmomath.NewDec(30),
					osmomath.OneDec().MulInt64(10*10+10),
					osmomath.OneDec().MulInt64(5),
					osmomath.ZeroDec())

				t3 := dummyTwapRecord(existingPool.GetId(), time.Now().Add(-time.Hour), "ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4", appparams.BaseCoinUnit, osmomath.NewDec(20),
					osmomath.OneDec().MulInt64(10*10+10*5),
					osmomath.OneDec().MulInt64(10),
					osmomath.ZeroDec())

				t4 := dummyTwapRecord(existingPool2.GetId(), time.Now().Add(-time.Hour*24), "akash", appparams.BaseCoinUnit, osmomath.NewDec(10),
					osmomath.OneDec().MulInt64(10*10*10),
					osmomath.OneDec().MulInt64(5),
					osmomath.ZeroDec())

				t5 := dummyTwapRecord(existingPool2.GetId(), time.Now().Add(-time.Hour), "akash", appparams.BaseCoinUnit, osmomath.NewDec(20),
					osmomath.OneDec().MulInt64(10),
					osmomath.OneDec().MulInt64(2),
					osmomath.ZeroDec())

				t6 := dummyTwapRecord(existingBalancerPoolId, time.Now().Add(-time.Hour), "atom", appparams.BaseCoinUnit, osmomath.NewDec(10),
					osmomath.OneDec().MulInt64(10),
					osmomath.OneDec().MulInt64(10),
					osmomath.ZeroDec())

				t7 := dummyTwapRecord(existingBalancerPoolId, time.Now().Add(-time.Minute*20), "atom", appparams.BaseCoinUnit, osmomath.NewDec(50),
					osmomath.OneDec().MulInt64(10*5),
					osmomath.OneDec().MulInt64(5),
					osmomath.ZeroDec())

				// store TWAP records
				s.App.TwapKeeper.StoreNewRecord(s.Ctx, t1)
				s.App.TwapKeeper.StoreNewRecord(s.Ctx, t2)
				s.App.TwapKeeper.StoreNewRecord(s.Ctx, t3)
				s.App.TwapKeeper.StoreNewRecord(s.Ctx, t4)
				s.App.TwapKeeper.StoreNewRecord(s.Ctx, t5)
				s.App.TwapKeeper.StoreNewRecord(s.Ctx, t6)
				s.App.TwapKeeper.StoreNewRecord(s.Ctx, t7)

				return expectedCoinsUsedInUpgradeHandler, existingBalancerPoolId
			},
			func(keepers *keepers.AppKeepers, expectedCoinsUsedInUpgradeHandler sdk.Coins, lastPoolID uint64) {
				lastPoolIdMinusOne := lastPoolID - 1
				lastPoolIdMinusTwo := lastPoolID - 2
				stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
				stakingParams.BondDenom = appparams.BaseCoinUnit
				s.App.StakingKeeper.SetParams(s.Ctx, stakingParams)

				// Retrieve the community pool balance before the upgrade
				communityPoolAddress := s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
				communityPoolBalancePre := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)

				numPoolPreUpgrade := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) - 1
				clPool1TwapRecordPreUpgrade, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(s.Ctx, lastPoolIdMinusTwo)
				s.Require().NoError(err)

				clPool1TwapRecordHistoricalPoolIndexPreUpgrade, err := keepers.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(s.Ctx, lastPoolIdMinusTwo)
				s.Require().NoError(err)

				clPool2TwapRecordPreUpgrade, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(s.Ctx, lastPoolIdMinusOne)
				s.Require().NoError(err)

				clPool2TwapRecordHistoricalPoolIndexPreUpgrade, err := keepers.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(s.Ctx, lastPoolIdMinusOne)
				s.Require().NoError(err)

				// Run upgrade handler.
				dummyUpgrade(s)
				s.Require().NotPanics(func() {
					_, err := s.preModule.PreBlock(s.Ctx)
					s.Require().NoError(err)
				})

				clPool1TwapRecordPostUpgrade, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(s.Ctx, lastPoolIdMinusTwo)
				s.Require().NoError(err)

				clPool1TwapRecordHistoricalPoolIndexPostUpgrade, err := keepers.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(s.Ctx, lastPoolIdMinusTwo)
				s.Require().NoError(err)

				clPool2TwapRecordPostUpgrade, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(s.Ctx, lastPoolIdMinusOne)
				s.Require().NoError(err)

				clPool2TwapRecordHistoricalPoolIndexPostUpgrade, err := keepers.TwapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(s.Ctx, lastPoolIdMinusOne)
				s.Require().NoError(err)

				// check that all TWAP records aren't empty
				s.Require().NotEmpty(clPool1TwapRecordPostUpgrade)
				s.Require().NotEmpty(clPool1TwapRecordHistoricalPoolIndexPostUpgrade)
				s.Require().NotEmpty(clPool2TwapRecordPostUpgrade)
				s.Require().NotEmpty(clPool2TwapRecordHistoricalPoolIndexPostUpgrade)

				for _, data := range []struct {
					pre, post []types.TwapRecord
				}{
					{clPool1TwapRecordPreUpgrade, clPool1TwapRecordPostUpgrade},
					{clPool1TwapRecordHistoricalPoolIndexPreUpgrade, clPool1TwapRecordHistoricalPoolIndexPostUpgrade},
					{clPool2TwapRecordPreUpgrade, clPool2TwapRecordPostUpgrade},
					{clPool2TwapRecordHistoricalPoolIndexPreUpgrade, clPool2TwapRecordHistoricalPoolIndexPostUpgrade},
				} {
					for i := range data.post {
						assertTwapFlipped(s, data.pre[i], data.post[i])
					}
				}

				// Retrieve the community pool balance (and the feePool balance) after the upgrade
				communityPoolBalancePost := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)
				feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
				s.Require().NoError(err)
				feePoolCommunityPoolPost := feePool.CommunityPool

				assetPairs, err := v17.InitializeAssetPairs(s.Ctx, keepers)
				s.Require().NoError(err)

				for i, assetPair := range assetPairs {
					// Get balancer pool's spot price.
					balancerSpotPrice, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, assetPair.LinkedClassicPool, assetPair.QuoteAsset, assetPair.BaseAsset)
					s.Require().NoError(err)

					// Validate CL pool was created.
					concentratedPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, lastPoolID+1)
					s.Require().NoError(err)
					s.Require().Equal(poolmanagertypes.Concentrated, concentratedPool.GetType())

					// Validate that denom0 and denom1 were set correctly
					concentratedTypePool, ok := concentratedPool.(cltypes.ConcentratedPoolExtension)
					s.Require().True(ok)
					s.Require().Equal(assetPair.BaseAsset, concentratedTypePool.GetToken0())
					s.Require().Equal(assetPair.QuoteAsset, concentratedTypePool.GetToken1())

					// Validate that the spot price of the CL pool is what we expect
					osmoassert.Equal(s.T(), multiplicativeTolerance, concentratedTypePool.GetCurrentSqrtPrice().PowerInteger(2), balancerSpotPrice)

					// Validate that the link is correct.
					migrationInfo, err := s.App.GAMMKeeper.GetAllMigrationInfo(s.Ctx)
					s.Require().NoError(err)
					link := migrationInfo.BalancerToConcentratedPoolLinks[i]
					s.Require().Equal(assetPair.LinkedClassicPool, link.BalancerPoolId)
					s.Require().Equal(concentratedPool.GetId(), link.ClPoolId)

					// Validate the sfs status
					clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(concentratedPool.GetId())
					_, err = s.App.SuperfluidKeeper.GetSuperfluidAsset(s.Ctx, clPoolDenom)
					if assetPair.Superfluid {
						s.Require().NoError(err)
					} else {
						s.Require().Error(err)
					}

					lastPoolID++
				}

				// Validate that the community pool balance has been reduced by the amount of osmo that was used to create the pool.
				s.Require().Equal(communityPoolBalancePre.Sub(expectedCoinsUsedInUpgradeHandler...).String(), communityPoolBalancePost.String())

				// Validate that the fee pool community pool balance has been decreased by the amount of osmo that was used to create the pool.
				s.Require().Equal(sdk.NewDecCoinsFromCoins(communityPoolBalancePost...).String(), feePoolCommunityPoolPost.String())

				numPoolPostUpgrade := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) - 1

				// Number of pools created should be equal to the number of records in the asset pairs.
				s.Require().Equal(len(assetPairs), int(numPoolPostUpgrade-numPoolPreUpgrade))

				// Validate that all links were created.
				migrationInfo, err := s.App.GAMMKeeper.GetAllMigrationInfo(s.Ctx)
				s.Require().Equal(len(assetPairs), len(migrationInfo.BalancerToConcentratedPoolLinks))
				s.Require().NoError(err)
			},
		},
		{
			"Test that the upgrade succeeds: testnet",
			func(keepers *keepers.AppKeepers) (sdk.Coins, uint64) {
				s.Ctx = s.Ctx.WithChainID("osmo-test-5")

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
						poolCoins := sdk.NewCoins(sdk.NewCoin(assetPair.BaseAsset, osmomath.NewInt(10000000000)), sdk.NewCoin(assetPair.QuoteAsset, osmomath.NewInt(10000000000)))
						s.PrepareBalancerPoolWithCoins(poolCoins...)

						// 0.1 OSMO used to get the respective base asset amount, 0.1 OSMO used to create the position
						osmoIn := sdk.NewCoin(v17.OSMO, osmomath.NewInt(100000).MulRaw(2))

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
						s.App.SuperfluidKeeper.SetSuperfluidAsset(s.Ctx, superfluidAsset)
					}
				}

				// We now create various pools that are not balancer pools.
				// This is to test if the testnet upgrade handler properly handles pools that are not of type balancer (i.e. should ignore them and move on).

				// Stableswap pool
				s.CreatePoolFromType(poolmanagertypes.Stableswap)
				// Cosmwasm pool
				s.CreatePoolFromType(poolmanagertypes.CosmWasm)
				// CL pool
				s.CreatePoolFromType(poolmanagertypes.Concentrated)

				lastPoolID += 3

				return expectedCoinsUsedInUpgradeHandler, lastPoolID
			},
			func(keepers *keepers.AppKeepers, expectedCoinsUsedInUpgradeHandler sdk.Coins, lastPoolID uint64) {
				// Set the bond denom to uosmo
				stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
				s.Require().NoError(err)
				stakingParams.BondDenom = appparams.BaseCoinUnit
				s.App.StakingKeeper.SetParams(s.Ctx, stakingParams)

				// Retrieve the community pool balance before the upgrade
				communityPoolAddress := s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
				communityPoolBalancePre := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)

				numPoolPreUpgrade := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) - 1

				gammPoolsPreUpgrade, err := s.App.GAMMKeeper.GetPools(s.Ctx)
				s.Require().NoError(err)

				// Run upgrade handler.
				dummyUpgrade(s)
				s.Require().NotPanics(func() {
					_, err := s.preModule.PreBlock(s.Ctx)
					s.Require().NoError(err)
				})

				// Retrieve the community pool balance (and the feePool balance) after the upgrade
				communityPoolBalancePost := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)
				feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
				s.Require().NoError(err)
				feePoolCommunityPoolPost := feePool.CommunityPool

				indexOffset := int(0)
				assetListIndex := int(0)

				// For testnet, we run through all gamm pools (not just the asset list)
				for i, pool := range gammPoolsPreUpgrade {
					// Skip pools that are not balancer pools
					if pool.GetType() != poolmanagertypes.Balancer {
						indexOffset++
						continue
					}

					gammPoolId := pool.GetId()
					cfmmPool, err := keepers.GAMMKeeper.GetCFMMPool(s.Ctx, gammPoolId)
					s.Require().NoError(err)

					poolCoins := cfmmPool.GetTotalPoolLiquidity(s.Ctx)

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
					balancerSpotPrice, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, gammPoolId, quoteAsset, baseAsset)
					s.Require().NoError(err)

					// Validate CL pool was created.
					concentratedPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, lastPoolID+1)
					s.Require().NoError(err)
					s.Require().Equal(poolmanagertypes.Concentrated, concentratedPool.GetType())

					// Validate that denom0 and denom1 were set correctly
					concentratedTypePool, ok := concentratedPool.(cltypes.ConcentratedPoolExtension)
					s.Require().True(ok)
					s.Require().Equal(baseAsset, concentratedTypePool.GetToken0())
					s.Require().Equal(quoteAsset, concentratedTypePool.GetToken1())

					// Validate that the spot price of the CL pool is what we expect
					osmoassert.Equal(s.T(), multiplicativeTolerance, concentratedTypePool.GetCurrentSqrtPrice().PowerInteger(2), balancerSpotPrice)

					// Validate that the link is correct.
					migrationInfo, err := s.App.GAMMKeeper.GetAllMigrationInfo(s.Ctx)
					s.Require().NoError(err)
					link := migrationInfo.BalancerToConcentratedPoolLinks[i-indexOffset]
					s.Require().Equal(gammPoolId, link.BalancerPoolId)
					s.Require().Equal(concentratedPool.GetId(), link.ClPoolId)

					// Validate the sfs status.
					// If the poolId matches a poolId on that asset list that had superfluid enabled, this pool should also be superfluid enabled.
					// Otherwise, it should not be superfluid enabled.
					assetListPoolId := v17.AssetPairsForTestsOnly[assetListIndex].LinkedClassicPool
					clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(concentratedPool.GetId())
					_, err = s.App.SuperfluidKeeper.GetSuperfluidAsset(s.Ctx, clPoolDenom)
					if assetListPoolId == gammPoolId {
						s.Require().NoError(err)
						assetListIndex++
						for assetListIndex < len(v17.AssetPairsForTestsOnly)-1 && v17.AssetPairsForTestsOnly[assetListIndex].Superfluid == false {
							assetListIndex++
						}
					} else {
						s.Require().Error(err)
					}

					lastPoolID++
				}

				// Validate that the community pool balance has been reduced by the amount of osmo that was used to create the pool.
				s.Require().Equal(communityPoolBalancePre.Sub(expectedCoinsUsedInUpgradeHandler...).String(), communityPoolBalancePost.String())

				// Validate that the fee pool community pool balance has been decreased by the amount of osmo that was used to create the pool.
				s.Require().Equal(sdk.NewDecCoinsFromCoins(communityPoolBalancePost...).String(), feePoolCommunityPoolPost.String())

				numPoolPostUpgrade := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) - 1
				numPoolsCreated := numPoolPostUpgrade - numPoolPreUpgrade

				// Number of pools created should be equal to the number of pools preUpgrade minus the number of pools that were not eligible for migration.
				numPoolsEligibleForMigration := numPoolPreUpgrade - 3
				s.Require().Equal(int(numPoolsEligibleForMigration), int(numPoolsCreated))

				// Validate that all links were created.
				migrationInfo, err := s.App.GAMMKeeper.GetAllMigrationInfo(s.Ctx)
				s.Require().Equal(int(numPoolsEligibleForMigration), len(migrationInfo.BalancerToConcentratedPoolLinks))
				s.Require().NoError(err)
			},
		},
		{
			"Fails because CFMM pool is not found",
			func(keepers *keepers.AppKeepers) (sdk.Coins, uint64) {
				return sdk.NewCoins(), 0
			},
			func(keepers *keepers.AppKeepers, expectedCoinsUsedInUpgradeHandler sdk.Coins, lastPoolID uint64) {
				dummyUpgrade(s)
				s.Require().NotPanics(func() {
					_, err := s.preModule.PreBlock(s.Ctx)
					s.Require().NoError(err)
				})
			},
		},
	}
	_ = testCases
	// for _, tc := range testCases {
	// 	s.Run(fmt.Sprintf("Case %s", tc.name), func() {
	// 		s.SetupTest() // reset

	// 		expectedCoinsUsedInUpgradeHandler, lastPoolID := tc.pre_upgrade(&s.App.AppKeepers)
	// 		tc.upgrade(&s.App.AppKeepers, expectedCoinsUsedInUpgradeHandler, lastPoolID)
	// 	})
	// }
}
