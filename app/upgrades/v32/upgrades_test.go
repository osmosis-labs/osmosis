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
	appparams "github.com/osmosis-labs/osmosis/v31/app/params"
	v32 "github.com/osmosis-labs/osmosis/v31/app/upgrades/v32"
)

const (
	v32UpgradeHeight = int64(10)
	// exploiterUosmo is the frozen account's only other mainnet balance: 703.055696 OSMO.
	exploiterUosmo = int64(703055696)
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestSeizeFrozenAllBTC() {
	s.setupUpgradeTest()

	exploiter := sdk.MustAccAddressFromBech32(v32.ExploiterAddress)
	subDAO := sdk.MustAccAddressFromBech32(v32.LiquiditySubDAOAddress)
	expected := expectedAllBTC(s)

	s.FundAcc(exploiter, sdk.NewCoins(
		sdk.NewCoin(v32.AllBTCDenom, expected),
		sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)),
	))

	s.Require().NoError(s.runUpgrade())

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, exploiter, v32.AllBTCDenom).IsZero())
	s.Require().Equal(
		sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)),
		s.App.BankKeeper.GetBalance(s.Ctx, exploiter, appparams.BaseCoinUnit),
	)
	s.Require().Equal(sdk.NewCoin(v32.AllBTCDenom, expected), s.App.BankKeeper.GetBalance(s.Ctx, subDAO, v32.AllBTCDenom))
}

func (s *UpgradeTestSuite) TestSeizeFrozenAllBTCRejectsUnexpectedBalance() {
	exploiter := sdk.MustAccAddressFromBech32(v32.ExploiterAddress)
	subDAO := sdk.MustAccAddressFromBech32(v32.LiquiditySubDAOAddress)

	for _, amount := range []osmomath.Int{osmomath.ZeroInt(), osmomath.NewInt(1)} {
		s.setupUpgradeTest()

		coins := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)))
		if !amount.IsZero() {
			coins = coins.Add(sdk.NewCoin(v32.AllBTCDenom, amount))
		}
		s.FundAcc(exploiter, coins)

		err := s.runUpgrade()
		s.Require().Error(err)
		s.Require().ErrorContains(err, "seizing frozen allBTC")

		s.Require().Equal(sdk.NewCoin(v32.AllBTCDenom, amount), s.App.BankKeeper.GetBalance(s.Ctx, exploiter, v32.AllBTCDenom))
		s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, subDAO, v32.AllBTCDenom).IsZero())
		s.Require().Equal(
			sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)),
			s.App.BankKeeper.GetBalance(s.Ctx, exploiter, appparams.BaseCoinUnit),
		)
	}
}

func (s *UpgradeTestSuite) TestSeizeFrozenAllBTCLeavesSurplus() {
	s.setupUpgradeTest()

	exploiter := sdk.MustAccAddressFromBech32(v32.ExploiterAddress)
	subDAO := sdk.MustAccAddressFromBech32(v32.LiquiditySubDAOAddress)
	authorized := expectedAllBTC(s)
	surplus := osmomath.NewInt(1)

	s.FundAcc(exploiter, sdk.NewCoins(
		sdk.NewCoin(v32.AllBTCDenom, authorized.Add(surplus)),
		sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)),
	))

	s.Require().NoError(s.runUpgrade())

	s.Require().Equal(sdk.NewCoin(v32.AllBTCDenom, surplus), s.App.BankKeeper.GetBalance(s.Ctx, exploiter, v32.AllBTCDenom))
	s.Require().Equal(
		sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(exploiterUosmo)),
		s.App.BankKeeper.GetBalance(s.Ctx, exploiter, appparams.BaseCoinUnit),
	)
	s.Require().Equal(sdk.NewCoin(v32.AllBTCDenom, authorized), s.App.BankKeeper.GetBalance(s.Ctx, subDAO, v32.AllBTCDenom))
}

func (s *UpgradeTestSuite) TestSeizeFrozenAllBTCAddsToExistingSubDAOBalance() {
	s.setupUpgradeTest()

	exploiter := sdk.MustAccAddressFromBech32(v32.ExploiterAddress)
	subDAO := sdk.MustAccAddressFromBech32(v32.LiquiditySubDAOAddress)
	expected := expectedAllBTC(s)
	existing := osmomath.NewInt(100)

	s.FundAcc(subDAO, sdk.NewCoins(sdk.NewCoin(v32.AllBTCDenom, existing)))
	s.FundAcc(exploiter, sdk.NewCoins(sdk.NewCoin(v32.AllBTCDenom, expected)))

	s.Require().NoError(s.runUpgrade())

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, exploiter, v32.AllBTCDenom).IsZero())
	s.Require().Equal(sdk.NewCoin(v32.AllBTCDenom, existing.Add(expected)), s.App.BankKeeper.GetBalance(s.Ctx, subDAO, v32.AllBTCDenom))
}

func (s *UpgradeTestSuite) setupUpgradeTest() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))
}

func (s *UpgradeTestSuite) runUpgrade() error {
	s.Ctx = s.Ctx.WithBlockHeight(v32UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: v32.UpgradeName, Height: v32UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v32UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v32UpgradeHeight)
	_, err = s.preModule.PreBlock(s.Ctx)
	return err
}

func expectedAllBTC(s *UpgradeTestSuite) osmomath.Int {
	amount, ok := osmomath.NewIntFromString(v32.ExpectedExploiterAllBTCAmount)
	s.Require().True(ok)
	return amount
}
