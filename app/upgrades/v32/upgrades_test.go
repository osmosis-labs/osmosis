package v32_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	v32 "github.com/osmosis-labs/osmosis/v31/app/upgrades/v32"
	protorevtypes "github.com/osmosis-labs/osmosis/v31/x/protorev/types"
)

const (
	v32UpgradeHeight = int64(10)
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestMoveProtorevNonBaseDenomsToSubDAO() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))

	recipient, err := sdk.AccAddressFromBech32(v32.LiquiditySubDAOAddress)
	s.Require().NoError(err)
	protorevAddr := s.App.AccountKeeper.GetModuleAddress(protorevtypes.ModuleName)

	baseDenoms, err := s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(baseDenoms)
	baseDenom := baseDenoms[0].Denom

	baseCoin := sdk.NewCoin(baseDenom, osmomath.NewInt(1_000))
	stuckCoins := sdk.NewCoins(
		sdk.NewCoin("ibc/STUCKA", osmomath.NewInt(5_000)),
		sdk.NewCoin("ibc/STUCKB", osmomath.NewInt(7_000)),
	)
	s.FundModuleAcc(protorevtypes.ModuleName, stuckCoins.Add(baseCoin))

	recipientBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, recipient)

	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	// Base denom stays in protorev for the regular epoch distribution; everything else is gone.
	s.Require().Equal(sdk.NewCoins(baseCoin), s.App.BankKeeper.GetAllBalances(s.Ctx, protorevAddr))
	s.Require().Equal(recipientBefore.Add(stuckCoins...), s.App.BankKeeper.GetAllBalances(s.Ctx, recipient))
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v32UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: v32.UpgradeName, Height: v32UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v32UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v32UpgradeHeight)
}
