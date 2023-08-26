package v16_test

import (
	"fmt"
	"testing"

	cosmwasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	v16 "github.com/osmosis-labs/osmosis/v19/app/upgrades/v16"
	cltypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v19/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v19/x/protorev/types"
)

var (
	DAIIBCDenom         = "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7"
	defaultDaiAmount, _ = sdk.NewIntFromString("73000000000000000000000")
	defaultDenom0mount  = sdk.NewInt(10000000000)
	desiredDenom0       = "uosmo"
	desiredDenom0Coin   = sdk.NewCoin(desiredDenom0, defaultDenom0mount)
	daiCoin             = sdk.NewCoin(DAIIBCDenom, defaultDaiAmount)
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

	// Allow 0.01% margin of error.
	multiplicativeTolerance := osmomath.ErrTolerance{
		MultiplicativeTolerance: sdk.MustNewDecFromStr("0.0001"),
	}
	defaultDaiAmount, _ := sdk.NewIntFromString("73000000000000000000000")
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
					suite.PrepareBalancerPoolWithCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", defaultDaiAmount))
				}

				// Create DAI / OSMO pool
				suite.PrepareBalancerPoolWithCoins(sdk.NewCoin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", defaultDaiAmount), sdk.NewCoin("uosmo", sdk.NewInt(10000000000)))

			},
			func() {
				stakingParams := suite.App.StakingKeeper.GetParams(suite.Ctx)
				stakingParams.BondDenom = "uosmo"
				suite.App.StakingKeeper.SetParams(suite.Ctx, stakingParams)

				oneDai := sdk.NewCoins(sdk.NewCoin(v16.DAIIBCDenom, sdk.NewInt(1000000000000000000)))

				// Send one dai to the community pool (this is true in current mainnet)
				suite.FundAcc(suite.TestAccs[0], oneDai)

				err := suite.App.DistrKeeper.FundCommunityPool(suite.Ctx, oneDai, suite.TestAccs[0])
				suite.Require().NoError(err)

				// Determine approx how much OSMO will be used from community pool when 1 DAI used.
				daiOsmoGammPool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, v16.DaiOsmoPoolId)
				suite.Require().NoError(err)
				respectiveOsmo, err := suite.App.GAMMKeeper.CalcOutAmtGivenIn(suite.Ctx, daiOsmoGammPool, oneDai[0], v16.DesiredDenom0, sdk.ZeroDec())
				suite.Require().NoError(err)

				// Retrieve the community pool balance before the upgrade
				communityPoolAddress := suite.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
				communityPoolBalancePre := suite.App.BankKeeper.GetAllBalances(suite.Ctx, communityPoolAddress)

				dummyUpgrade(suite)
				suite.Require().NotPanics(func() {
					suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
				})

				// Retrieve the community pool balance (and the feePool balance) after the upgrade
				communityPoolBalancePost := suite.App.BankKeeper.GetAllBalances(suite.Ctx, communityPoolAddress)
				feePoolCommunityPoolPost := suite.App.DistrKeeper.GetFeePool(suite.Ctx).CommunityPool

				// Validate that the community pool balance has been reduced by the amount of OSMO that was used to create the pool
				// Note we use all the osmo, but a small amount of DAI is left over due to rounding when creating the first position.
				suite.Require().Equal(communityPoolBalancePre.AmountOf("uosmo").Sub(respectiveOsmo.Amount).String(), communityPoolBalancePost.AmountOf("uosmo").String())
				suite.Require().Equal(0, multiplicativeTolerance.Compare(communityPoolBalancePre.AmountOf(v16.DAIIBCDenom), oneDai[0].Amount.Sub(communityPoolBalancePost.AmountOf(v16.DAIIBCDenom))))

				// Validate that the fee pool community pool balance has been decreased by the amount of OSMO/DAI that was used to create the pool
				suite.Require().Equal(communityPoolBalancePost.AmountOf("uosmo").String(), feePoolCommunityPoolPost.AmountOf("uosmo").TruncateInt().String())
				suite.Require().Equal(communityPoolBalancePost.AmountOf(v16.DAIIBCDenom).String(), feePoolCommunityPoolPost.AmountOf(v16.DAIIBCDenom).TruncateInt().String())

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
				suite.Require().Equal(0, multiplicativeTolerance.CompareBigDec(concentratedTypePool.GetCurrentSqrtPrice().PowerInteger(2), osmomath.BigDecFromSDKDec(balancerSpotPrice)))

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

				// Validate that expedited quorum was set to 2/3
				expQuorum := suite.App.GovKeeper.GetTallyParams(suite.Ctx).ExpeditedQuorum
				suite.Require().Equal(sdk.NewDec(2).Quo(sdk.NewDec(3)), expQuorum)

				// Validate that cw pool module address is allowed to upload contract code
				allowedAddresses := suite.App.WasmKeeper.GetParams(suite.Ctx).CodeUploadAccess.Addresses
				isCwPoolModuleAddressAllowedUpload := osmoutils.Contains(allowedAddresses, suite.App.AccountKeeper.GetModuleAddress(cosmwasmpooltypes.ModuleName).String())
				suite.Require().True(isCwPoolModuleAddressAllowedUpload)
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
