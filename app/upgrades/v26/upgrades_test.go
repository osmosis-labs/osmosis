package v26_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
	v26 "github.com/osmosis-labs/osmosis/v25/app/upgrades/v26"

	"cosmossdk.io/x/upgrade"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

const (
	v26UpgradeHeight = int64(10)
)

var (
	consAddr = sdk.ConsAddress(sdk.AccAddress([]byte("addr1_______________")))
	denomA   = "denomA"
	denomB   = "denomB"
	denomC   = "denomC"
	denomD   = "denomD"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))

	s.PrepareTradingPairTakerFeeTest()

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	s.ExecuteTradingPairTakerFeeTest()
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v26UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: v26.Upgrade.UpgradeName, Height: v26UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v26UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v26UpgradeHeight)
}

func (s *UpgradeTestSuite) PrepareTradingPairTakerFeeTest() {
	// Set some trading pair taker fee entries
	s.App.PoolManagerKeeper.SetDenomPairTakerFee(s.Ctx, denomA, denomB, osmomath.MustNewDecFromStr("0.005"))
	s.App.PoolManagerKeeper.SetDenomPairTakerFee(s.Ctx, denomC, denomD, osmomath.MustNewDecFromStr("0.006"))

	expectedTradingPairTakerFees := []poolmanagertypes.DenomPairTakerFee{
		{TokenInDenom: denomC, TokenOutDenom: denomD, TakerFee: osmomath.MustNewDecFromStr("0.006")},
		{TokenInDenom: denomA, TokenOutDenom: denomB, TakerFee: osmomath.MustNewDecFromStr("0.005")},
	}

	// Retrieve all trading pair taker fees, and check if they are as expected
	allTradingPairTakerFees, err := s.App.PoolManagerKeeper.GetAllTradingPairTakerFees(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(allTradingPairTakerFees, 2)
	s.Require().Equal(expectedTradingPairTakerFees, allTradingPairTakerFees)
}

func (s *UpgradeTestSuite) ExecuteTradingPairTakerFeeTest() {
	expectedTradingPairTakerFees := []poolmanagertypes.DenomPairTakerFee{
		{TokenInDenom: denomD, TokenOutDenom: denomC, TakerFee: osmomath.MustNewDecFromStr("0.006")},
		{TokenInDenom: denomC, TokenOutDenom: denomD, TakerFee: osmomath.MustNewDecFromStr("0.006")},
		{TokenInDenom: denomB, TokenOutDenom: denomA, TakerFee: osmomath.MustNewDecFromStr("0.005")},
		{TokenInDenom: denomA, TokenOutDenom: denomB, TakerFee: osmomath.MustNewDecFromStr("0.005")},
	}

	// Retrieve all trading pair taker fees, and check if they are modified as expected
	allTradingPairTakerFees, err := s.App.PoolManagerKeeper.GetAllTradingPairTakerFees(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(allTradingPairTakerFees, 4)
	s.Require().Equal(expectedTradingPairTakerFees, allTradingPairTakerFees)
}
