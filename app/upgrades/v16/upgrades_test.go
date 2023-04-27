package v16_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	v16 "github.com/osmosis-labs/osmosis/v15/app/upgrades/v16"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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

const dummyUpgradeHeight = 5

func dummyUpgrade(suite *UpgradeTestSuite) {
	suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v16", Height: dummyUpgradeHeight}
	err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
	suite.Require().NoError(err)
	plan, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
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

	testCases := []struct {
		name         string
		pre_upgrade  func()
		upgrade      func()
		post_upgrade func()
	}{
		{
			"Test that the upgrade succeeds",
			func() {
				upgradeSetup()

				// Create earlier pools
				for i := uint64(1); i < v16.DaiOsmoPoolId; i++ {
					suite.PrepareBalancerPoolWithCoins(desiredDenom0Coin, daiCoin)
				}

				// Create DAI / OSMO pool
				suite.PrepareBalancerPoolWithCoins(sdk.NewCoin(v16.DAIIBCDenom, desiredDenom0Coin.Amount), desiredDenom0Coin)
			},
			func() {
				dummyUpgrade(suite)
				suite.Require().NotPanics(func() {
					suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
				})

				// Validate CL pool was created.
				concentratedPool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, v16.DaiOsmoPoolId+1)
				suite.Require().NoError(err)
				suite.Require().Equal(poolmanagertypes.Concentrated, concentratedPool.GetType())

				// Validate that denom0 and denom1 were set correctly
				concentratedTypePool, ok := concentratedPool.(cltypes.ConcentratedPoolExtension)
				suite.Require().True(ok)
				suite.Require().Equal(v16.DesiredDenom0, concentratedTypePool.GetToken0())
				suite.Require().Equal(v16.DAIIBCDenom, concentratedTypePool.GetToken1())

				// Validate that link was created.
				migrationInfo, err := suite.App.GAMMKeeper.GetAllMigrationInfo(suite.Ctx)
				suite.Require().Equal(1, len(migrationInfo.BalancerToConcentratedPoolLinks))
				suite.Require().NoError(err)

				// Validate that the link is correct.
				link := migrationInfo.BalancerToConcentratedPoolLinks[0]
				suite.Require().Equal(v16.DaiOsmoPoolId, link.BalancerPoolId)
				suite.Require().Equal(concentratedPool.GetId(), link.ClPoolId)

				// Check authorized denoms are set correctly.
				params := suite.App.ConcentratedLiquidityKeeper.GetParams(suite.Ctx)
				suite.Require().EqualValues(params.AuthorizedQuoteDenoms, v16.AuthorizedQuoteDenoms)
				suite.Require().EqualValues(params.AuthorizedUptimes, v16.AuthorizedUptimes)

				// Permissionless pool creation is disabled.
				suite.Require().False(params.IsPermissionlessPoolCreationEnabled)
			},
			func() {
			},
		},
		{
			"Fails because CFMM pool is not found",
			func() {
				upgradeSetup()
			},
			func() {
				dummyUpgrade(suite)
				suite.Require().Panics(func() {
					suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
				})
			},
			func() {
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.pre_upgrade()
			tc.upgrade()
			tc.post_upgrade()
		})
	}
}
