package v16_test

import (
	"testing"

	cosmwasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/store/prefix"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v31/app/params"
	v16 "github.com/osmosis-labs/osmosis/v31/app/upgrades/v16"
	cltypes "github.com/osmosis-labs/osmosis/v31/x/concentrated-liquidity/types"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v31/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
	protorevtypes "github.com/osmosis-labs/osmosis/v31/x/protorev/types"
)

var (
	DAIIBCDenom = "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const dummyUpgradeHeight = 5

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v16", Height: dummyUpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight)
}

func (s *UpgradeTestSuite) TestUpgrade() {
	upgradeSetup := func() {
		// This is done to ensure that we run the InitGenesis() logic for the new modules
		upgradeStoreKey := s.App.AppKeepers.GetKey(upgradetypes.StoreKey)
		store := s.Ctx.KVStore(upgradeStoreKey)
		versionStore := prefix.NewStore(store, []byte{upgradetypes.VersionMapByte})
		versionStore.Delete([]byte(cltypes.ModuleName))

		// Ensure proper setup for ProtoRev upgrade testing
		err := upgradeProtorevSetup(s)
		s.Require().NoError(err)
	}

	// Allow 0.01% margin of error.
	multiplicativeTolerance := osmomath.ErrTolerance{
		MultiplicativeTolerance: osmomath.MustNewDecFromStr("0.0001"),
	}
	defaultDaiAmount, _ := osmomath.NewIntFromString("73000000000000000000000")
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
					s.PrepareBalancerPoolWithCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10000000000)), sdk.NewCoin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", defaultDaiAmount))
				}

				// Create DAI / OSMO pool
				s.PrepareBalancerPoolWithCoins(sdk.NewCoin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", defaultDaiAmount), sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10000000000)))
			},
			func() {
				stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
				s.Require().NoError(err)
				stakingParams.BondDenom = appparams.BaseCoinUnit
				s.App.StakingKeeper.SetParams(s.Ctx, stakingParams)

				oneDai := sdk.NewCoins(sdk.NewCoin(v16.DAIIBCDenom, osmomath.NewInt(1000000000000000000)))

				// Send one dai to the community pool (this is true in current mainnet)
				s.FundAcc(s.TestAccs[0], oneDai)

				err = s.App.DistrKeeper.FundCommunityPool(s.Ctx, oneDai, s.TestAccs[0])
				s.Require().NoError(err)

				// Determine approx how much OSMO will be used from community pool when 1 DAI used.
				daiOsmoGammPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, v16.DaiOsmoPoolId)
				s.Require().NoError(err)
				respectiveOsmo, err := s.App.GAMMKeeper.CalcOutAmtGivenIn(s.Ctx, daiOsmoGammPool, oneDai[0], v16.DesiredDenom0, osmomath.ZeroDec())
				s.Require().NoError(err)

				// Retrieve the community pool balance before the upgrade
				communityPoolAddress := s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
				communityPoolBalancePre := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)

				dummyUpgrade(s)
				s.Require().NotPanics(func() {
					_, err := s.App.BeginBlocker(s.Ctx)
					s.Require().NoError(err)
				})

				// Retrieve the community pool balance (and the feePool balance) after the upgrade
				communityPoolBalancePost := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)
				feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
				s.Require().NoError(err)
				feePoolCommunityPoolPost := feePool.CommunityPool

				// Validate that the community pool balance has been reduced by the amount of OSMO that was used to create the pool
				// Note we use all the osmo, but a small amount of DAI is left over due to rounding when creating the first position.
				s.Require().Equal(communityPoolBalancePre.AmountOf(appparams.BaseCoinUnit).Sub(respectiveOsmo.Amount).String(), communityPoolBalancePost.AmountOf(appparams.BaseCoinUnit).String())
				osmoassert.Equal(s.T(), multiplicativeTolerance, communityPoolBalancePre.AmountOf(v16.DAIIBCDenom), oneDai[0].Amount.Sub(communityPoolBalancePost.AmountOf(v16.DAIIBCDenom)))

				// Validate that the fee pool community pool balance has been decreased by the amount of OSMO/DAI that was used to create the pool
				s.Require().Equal(communityPoolBalancePost.AmountOf(appparams.BaseCoinUnit).String(), feePoolCommunityPoolPost.AmountOf(appparams.BaseCoinUnit).TruncateInt().String())
				s.Require().Equal(communityPoolBalancePost.AmountOf(v16.DAIIBCDenom).String(), feePoolCommunityPoolPost.AmountOf(v16.DAIIBCDenom).TruncateInt().String())

				// Get balancer pool's spot price.
				balancerSpotPrice, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, v16.DaiOsmoPoolId, v16.DAIIBCDenom, v16.DesiredDenom0)
				s.Require().NoError(err)

				// Validate CL pool was created.
				concentratedPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, v16.DaiOsmoPoolId+1)
				s.Require().NoError(err)
				s.Require().Equal(poolmanagertypes.Concentrated, concentratedPool.GetType())

				// Validate that denom0 and denom1 were set correctly
				concentratedTypePool, ok := concentratedPool.(cltypes.ConcentratedPoolExtension)
				s.Require().True(ok)
				s.Require().Equal(v16.DesiredDenom0, concentratedTypePool.GetToken0())
				s.Require().Equal(v16.DAIIBCDenom, concentratedTypePool.GetToken1())

				// Validate that the spot price of the CL pool is what we expect
				osmoassert.Equal(s.T(), multiplicativeTolerance, concentratedTypePool.GetCurrentSqrtPrice().PowerInteger(2), balancerSpotPrice)

				// Validate that link was created.
				migrationInfo, err := s.App.GAMMKeeper.GetAllMigrationInfo(s.Ctx)
				s.Require().Equal(1, len(migrationInfo.BalancerToConcentratedPoolLinks))
				s.Require().NoError(err)

				// Validate that the link is correct.
				link := migrationInfo.BalancerToConcentratedPoolLinks[0]
				s.Require().Equal(v16.DaiOsmoPoolId, link.BalancerPoolId)
				s.Require().Equal(concentratedPool.GetId(), link.ClPoolId)

				// Check authorized denoms are set correctly.
				params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
				// s.Require().EqualValues(params.AuthorizedQuoteDenoms, v16.AuthorizedQuoteDenoms)
				s.Require().EqualValues(params.AuthorizedUptimes, v16.AuthorizedUptimes)

				// Permissionless pool creation is disabled.
				s.Require().False(params.IsPermissionlessPoolCreationEnabled)

				// Ensure that the protorev upgrade was successful
				// verifyProtorevUpdateSuccess(s)

				// Validate MsgExecuteContract and MsgInstantiateContract were added to the whitelist
				icaHostAllowList := s.App.ICAHostKeeper.GetParams(s.Ctx)
				s.Require().Contains(icaHostAllowList.AllowMessages, sdk.MsgTypeURL(&cosmwasmtypes.MsgExecuteContract{}))
				s.Require().Contains(icaHostAllowList.AllowMessages, sdk.MsgTypeURL(&cosmwasmtypes.MsgInstantiateContract{}))

				// Validate that expedited quorum was set to 2/3

				// GetTallyParams no longer exists, keeping commented for historical purposes
				// expQuorum := s.App.GovKeeper.GetTallyParams(s.Ctx).ExpeditedQuorum
				// s.Require().Equal(osmomath.NewDec(2).Quo(osmomath.NewDec(3)), expQuorum)

				// Validate that cw pool module address is allowed to upload contract code
				allowedAddresses := s.App.WasmKeeper.GetParams(s.Ctx).CodeUploadAccess.Addresses
				isCwPoolModuleAddressAllowedUpload := osmoutils.Contains(allowedAddresses, s.App.AccountKeeper.GetModuleAddress(cosmwasmpooltypes.ModuleName).String())
				s.Require().True(isCwPoolModuleAddressAllowedUpload)
			},
			func() {
				// Validate that tokenfactory params have been updated
				params := s.App.TokenFactoryKeeper.GetParams(s.Ctx)
				s.Require().Nil(params.DenomCreationFee)
				s.Require().Equal(v16.NewDenomCreationGasConsume, params.DenomCreationGasConsume)
			},
		},
		{
			"Fails because CFMM pool is not found",
			func() {
				upgradeSetup()
			},
			func() {
				dummyUpgrade(s)
				s.Require().NotPanics(func() {
					_, err := s.App.BeginBlocker(s.Ctx)
					s.Require().NoError(err)
				})
			},
			func() {
			},
		},
	}

	_ = testCases
	// for _, tc := range testCases {
	// 	s.Run(fmt.Sprintf("Case %s", tc.name), func() {
	// 		s.SetupTest() // reset

	// 		tc.pre_upgrade()
	// 		tc.upgrade()
	// 		tc.post_upgrade()
	// 	})
	// }
}

func verifyProtorevUpdateSuccess(s *UpgradeTestSuite) {
	panic("unimplemented")
}

func upgradeProtorevSetup(s *UpgradeTestSuite) error {
	account := apptesting.CreateRandomAccounts(1)[0]
	s.App.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, account)

	devFee := sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000))
	if err := s.App.ProtoRevKeeper.SetDeveloperFees(s.Ctx, devFee); err != nil {
		return err
	}

	fundCoin := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000)))

	if err := s.App.AppKeepers.BankKeeper.MintCoins(s.Ctx, protorevtypes.ModuleName, fundCoin); err != nil {
		return err
	}

	return nil
}

// func verifyProtorevUpdateSuccess(s *UpgradeTestSuite) {
// 	// Ensure balance was transferred to the developer account
// 	devAcc, err := s.App.ProtoRevKeeper.GetDeveloperAccount(s.Ctx)
// 	s.Require().NoError(err)
// 	s.Require().Equal(s.App.BankKeeper.GetBalance(s.Ctx, devAcc, appparams.BaseCoinUnit), sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000)))

// 	// Ensure developer fees are empty
// 	coins, err := s.App.ProtoRevKeeper.GetAllDeveloperFees(s.Ctx)
// 	s.Require().NoError(err)
// 	s.Require().Equal(coins, []sdk.Coin{})
// }
