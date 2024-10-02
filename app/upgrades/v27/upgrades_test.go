package v27_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v26/app/apptesting"
	v27 "github.com/osmosis-labs/osmosis/v26/app/upgrades/v27"

	"cosmossdk.io/x/upgrade"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	"github.com/osmosis-labs/osmosis/osmomath"
)

const (
	v27UpgradeHeight = int64(10)
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

	s.PrepareSupplyOffsetTest()

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	s.ExecuteSupplyOffsetTest()
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v27UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: v27.Upgrade.UpgradeName, Height: v27UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v27UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v27UpgradeHeight)
}

func (s *UpgradeTestSuite) PrepareSupplyOffsetTest() {
	// Set some supply offsets
	s.App.BankKeeper.AddSupplyOffset(s.Ctx, v27.OsmoToken, osmomath.NewInt(1000))
	s.App.BankKeeper.AddSupplyOffsetOld(s.Ctx, v27.OsmoToken, osmomath.NewInt(-500))
}

func (s *UpgradeTestSuite) ExecuteSupplyOffsetTest() {
	coin := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, v27.OsmoToken)
	offset := s.App.BankKeeper.GetSupplyOffset(s.Ctx, v27.OsmoToken)
	oldOffset := s.App.BankKeeper.GetSupplyOffsetOld(s.Ctx, v27.OsmoToken)

	s.Require().Equal("500uosmo", coin.String())
	s.Require().Equal("500", offset.String())
	s.Require().Equal("0", oldOffset.String())
}
