package v30_test

import (
	"testing"
	"time"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	auctiontypes "github.com/skip-mev/block-sdk/v2/x/auction/types"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	v30 "github.com/osmosis-labs/osmosis/v31/app/upgrades/v30"
)

const (
	v30UpgradeHeight = int64(10)
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule                          appmodule.HasPreBlocker
	authorizedQuoteDenomsBeforeUpgrade []string
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestCommunityPoolDenomWhitelistUpgrade() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))

	s.PrepareCommunityPoolDenomWhitelistTest()

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	s.ExecuteCommunityPoolDenomWhitelistTest()
}

func (s *UpgradeTestSuite) TestTopOfBlockAuctionFundTransferUpgrade() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))

	usdcDenom := "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"
	auctionRev := sdk.NewCoins(sdk.NewCoin(usdcDenom, osmomath.NewInt(999999999999999999)))
	s.FundModuleAcc(auctiontypes.ModuleName, auctionRev)

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	auctionModuleAccountAddr := s.App.AccountKeeper.GetModuleAccount(s.Ctx, auctiontypes.ModuleName).GetAddress()
	distributionModuleAccountAddr := s.App.AccountKeeper.GetModuleAccount(s.Ctx, distrtypes.ModuleName).GetAddress()

	auctionModuleBalance := s.App.BankKeeper.GetBalance(s.Ctx, auctionModuleAccountAddr, usdcDenom)
	distributionModuleBalance := s.App.BankKeeper.GetBalance(s.Ctx, distributionModuleAccountAddr, usdcDenom)

	s.Require().Equal(sdk.NewCoin(usdcDenom, osmomath.NewInt(0)), auctionModuleBalance)
	s.Require().Equal(sdk.NewCoin(usdcDenom, osmomath.NewInt(999999999999999999)), distributionModuleBalance)
}

func (s *UpgradeTestSuite) TestUpgradeWithCustomAuthorizedQuoteDenoms() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))

	// Set custom authorized quote denoms
	customAuthorizedQuoteDenoms := []string{
		"uosmo",
		"uatom",
		"uusdc",
		"udai",
		"custom_denom_1",
		"custom_denom_2",
	}

	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.AuthorizedQuoteDenoms = customAuthorizedQuoteDenoms
	// Clear the CommunityPoolDenomWhitelist to ensure it starts empty for the test
	poolManagerParams.TakerFeeParams.CommunityPoolDenomWhitelist = []string{}
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)

	// Verify the custom denoms are set
	updatedParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	s.Require().Equal(customAuthorizedQuoteDenoms, updatedParams.AuthorizedQuoteDenoms)

	// Verify that CommunityPoolDenomWhitelist is still empty before upgrade
	s.Require().Empty(updatedParams.TakerFeeParams.CommunityPoolDenomWhitelist)

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	// Verify that CommunityPoolDenomWhitelist now contains the custom denoms
	finalParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	s.Require().Equal(customAuthorizedQuoteDenoms, finalParams.TakerFeeParams.CommunityPoolDenomWhitelist)
	s.Require().Equal(customAuthorizedQuoteDenoms, finalParams.AuthorizedQuoteDenoms)
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v30UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: v30.UpgradeName, Height: v30UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v30UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v30UpgradeHeight)
}

// PrepareCommunityPoolDenomWhitelistTest prepares the community pool denom whitelist migration test
func (s *UpgradeTestSuite) PrepareCommunityPoolDenomWhitelistTest() {
	// Get current poolmanager parameters
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)

	// Clear the CommunityPoolDenomWhitelist to ensure it starts empty for the test
	poolManagerParams.TakerFeeParams.CommunityPoolDenomWhitelist = []string{}
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)

	// Get the updated parameters to verify the whitelist is now empty
	poolManagerParams = s.App.PoolManagerKeeper.GetParams(s.Ctx)

	// Verify that CommunityPoolDenomWhitelist is empty before upgrade
	s.Require().Empty(poolManagerParams.TakerFeeParams.CommunityPoolDenomWhitelist)

	// Verify that AuthorizedQuoteDenoms has some values
	s.Require().NotEmpty(poolManagerParams.AuthorizedQuoteDenoms)

	// Store the authorized quote denoms for comparison
	s.authorizedQuoteDenomsBeforeUpgrade = poolManagerParams.AuthorizedQuoteDenoms
}

// ExecuteCommunityPoolDenomWhitelistTest executes the community pool denom whitelist migration test
func (s *UpgradeTestSuite) ExecuteCommunityPoolDenomWhitelistTest() {
	// Get poolmanager parameters after upgrade
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)

	// Verify that CommunityPoolDenomWhitelist now contains the same values as AuthorizedQuoteDenoms
	s.Require().Equal(s.authorizedQuoteDenomsBeforeUpgrade, poolManagerParams.TakerFeeParams.CommunityPoolDenomWhitelist)

	// Verify that AuthorizedQuoteDenoms remains unchanged
	s.Require().Equal(s.authorizedQuoteDenomsBeforeUpgrade, poolManagerParams.AuthorizedQuoteDenoms)

	// Verify that the CommunityPoolDenomWhitelist is not empty after upgrade
	s.Require().NotEmpty(poolManagerParams.TakerFeeParams.CommunityPoolDenomWhitelist)

	// Verify that both lists have the same length
	s.Require().Len(poolManagerParams.TakerFeeParams.CommunityPoolDenomWhitelist, len(s.authorizedQuoteDenomsBeforeUpgrade))

	// Verify that each denom in the whitelist is valid
	for _, denom := range poolManagerParams.TakerFeeParams.CommunityPoolDenomWhitelist {
		s.Require().NoError(sdk.ValidateDenom(denom))
	}
}
