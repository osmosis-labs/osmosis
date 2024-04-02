package v25_test

import (
	abci "github.com/cometbft/cometbft/abci/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/osmosis-labs/osmosis/v24/app/apptesting"
	"github.com/osmosis-labs/osmosis/v24/app/upgrades/v25"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	v25UpgradeHeight = 100
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
	})

	// check auction params
	params, err := s.App.AuctionKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)

	// check auction params
	s.Require().Equal(params.MaxBundleSize, v25.AuctionParams.MaxBundleSize)
	s.Require().Equal(params.ReserveFee.Denom, v25.AuctionParams.ReserveFee.Denom)
	s.Require().Equal(params.ReserveFee.Amount.Int64(), v25.AuctionParams.ReserveFee.Amount.Int64())
	s.Require().Equal(params.MinBidIncrement.Denom, v25.AuctionParams.MinBidIncrement.Denom)
	s.Require().Equal(params.MinBidIncrement.Amount.Int64(), v25.AuctionParams.MinBidIncrement.Amount.Int64())
	s.Require().Equal(params.EscrowAccountAddress, v25.AuctionParams.EscrowAccountAddress)
	s.Require().Equal(params.FrontRunningProtection, v25.AuctionParams.FrontRunningProtection)
	s.Require().Equal(params.ProposerFee, v25.AuctionParams.ProposerFee)
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v25UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: v25.Upgrade.UpgradeName, Height: v25UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, exists := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(exists)

	s.Ctx = s.Ctx.WithBlockHeight(v25UpgradeHeight)
}
