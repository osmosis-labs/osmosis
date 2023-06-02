package v16_test

import (
	"fmt"
	"testing"

	cosmwasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
	v16 "github.com/osmosis-labs/osmosis/v16/app/upgrades/v16"
	cltypes "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v16/x/protorev/types"
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

		// Ensure proper setup for ProtoRev upgrade testing
		err := upgradeProtorevSetup(suite)
		suite.Require().NoError(err)
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
				suite.PrepareBalancerPoolWithCoins(daiCoin, desiredDenom0Coin)
			},
			func() {
				stakingParams := suite.App.StakingKeeper.GetParams(suite.Ctx)
				stakingParams.BondDenom = "uosmo"
				suite.App.StakingKeeper.SetParams(suite.Ctx, stakingParams)
				dummyUpgrade(suite)
				suite.Require().NotPanics(func() {
					suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
				})

				// Get balancer pool's spot price.
				balancerSpotPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, v16.DaiOsmoPoolId, v16.DAIIBCDenom, v16.DesiredDenom0)
				suite.Require().NoError(err)

				// Validate CL pool was created.
				concentratedPool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, v16.DaiOsmoPoolId+1)
				suite.Require().NoError(err)
				suite.Require().Equal(poolmanagertypes.Concentrated, concentratedPool.GetType())

				// Validate that denom0 and denom1 were set correctly
				concentratedTypePool, ok := concentratedPool.(cltypes.ConcentratedPoolExtension)
				suite.Require().True(ok)
				suite.Require().Equal(v16.DesiredDenom0, concentratedTypePool.GetToken0())
				suite.Require().Equal(v16.DAIIBCDenom, concentratedTypePool.GetToken1())

				// Validate that the spot price of the CL pool is what we expect
				osmoassert.DecApproxEq(suite.T(), concentratedTypePool.GetCurrentSqrtPrice().Power(2), balancerSpotPrice, sdk.NewDec(4))

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

				// Ensure that the protorev upgrade was successful
				verifyProtorevUpdateSuccess(suite)

				// Validate MsgExecuteContract and MsgInstantiateContract were added to the whitelist
				icaHostAllowList := suite.App.ICAHostKeeper.GetParams(suite.Ctx)
				suite.Require().Contains(icaHostAllowList.AllowMessages, sdk.MsgTypeURL(&cosmwasmtypes.MsgExecuteContract{}))
				suite.Require().Contains(icaHostAllowList.AllowMessages, sdk.MsgTypeURL(&cosmwasmtypes.MsgInstantiateContract{}))
			},
			func() {
				// Validate that tokenfactory params have been updated
				params := suite.App.TokenFactoryKeeper.GetParams(suite.Ctx)
				suite.Require().Nil(params.DenomCreationFee)
				suite.Require().Equal(v16.NewDenomCreationGasConsume, params.DenomCreationGasConsume)
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

func upgradeProtorevSetup(suite *UpgradeTestSuite) error {
	account := apptesting.CreateRandomAccounts(1)[0]
	suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

	devFee := sdk.NewCoin("uosmo", sdk.NewInt(1000000))
	if err := suite.App.ProtoRevKeeper.SetDeveloperFees(suite.Ctx, devFee); err != nil {
		return err
	}

	fundCoin := sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000)))

	if err := suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, protorevtypes.ModuleName, fundCoin); err != nil {
		return err
	}

	return nil
}

func verifyProtorevUpdateSuccess(suite *UpgradeTestSuite) {
	// Ensure balance was transferred to the developer account
	devAcc, err := suite.App.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.App.BankKeeper.GetBalance(suite.Ctx, devAcc, "uosmo"), sdk.NewCoin("uosmo", sdk.NewInt(1000000)))

	// Ensure developer fees are empty
	coins, err := suite.App.ProtoRevKeeper.GetAllDeveloperFees(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(coins, []sdk.Coin{})
}
