package v21_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	v21 "github.com/osmosis-labs/osmosis/v31/app/upgrades/v21"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	"github.com/osmosis-labs/osmosis/v31/x/protorev/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
)

const (
	v21UpgradeHeight = int64(10)
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.SetupWithCustomChainId(v21.TestingChainId)
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	// Pseudo collect cyclic arb profits
	cyclicArbProfits := sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(9000)), sdk.NewCoin("Atom", osmomath.NewInt(3000)))
	err := s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, cyclicArbProfits[0].Denom, cyclicArbProfits[0].Amount)
	s.Require().NoError(err)
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, cyclicArbProfits[1].Denom, cyclicArbProfits[1].Amount)
	s.Require().NoError(err)

	allProtocolRevenue := s.App.ProtoRevKeeper.GetAllProtocolRevenue(s.Ctx)
	// Check all accounting start heights should be the same height as the upgrade
	s.Require().Equal(v21UpgradeHeight, allProtocolRevenue.CyclicArbTracker.HeightAccountingStartsFrom)
	s.Require().Equal(v21UpgradeHeight, allProtocolRevenue.TakerFeesTracker.HeightAccountingStartsFrom)
	// s.Require().Equal(v21UpgradeHeight, allProtocolRevenue.TxFeesTracker.HeightAccountingStartsFrom)
	// All values should be nill except for the cyclic arb profits, which should start at the value it was at time of upgrade
	s.Require().Equal([]sdk.Coin{}, allProtocolRevenue.TakerFeesTracker.TakerFeesToCommunityPool)
	s.Require().Equal([]sdk.Coin{}, allProtocolRevenue.TakerFeesTracker.TakerFeesToStakers)
	// s.Require().Equal(sdk.Coins(nil), allProtocolRevenue.TxFeesTracker.TxFees)
	s.Require().Equal([]sdk.Coin(cyclicArbProfits), allProtocolRevenue.CyclicArbTracker.CyclicArb)

}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v21UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v21", Height: v21UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v21UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v21UpgradeHeight)
}
