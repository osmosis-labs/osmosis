package v22_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v31/app/params"
	"github.com/osmosis-labs/osmosis/v31/x/protorev/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
)

const (
	v22UpgradeHeight = int64(10)
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

	expectedTakerFeeForStakers := []sdk.Coin{sdk.NewCoin("uakt", osmomath.NewInt(3000)), sdk.NewCoin("uatom", osmomath.NewInt(1000)), sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(2000))}
	expectedTakerFeeForCommunityPool := []sdk.Coin{sdk.NewCoin("uakt", osmomath.NewInt(2000)), sdk.NewCoin("uatom", osmomath.NewInt(3000)), sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000))}
	expectedTrackerStartHeight := int64(3)

	// Set up old protorev tracker prior to upgrade
	s.App.PoolManagerKeeper.SetTakerFeeTrackerStartHeight(s.Ctx, expectedTrackerStartHeight)
	newTakerFeeForStakers := poolmanagertypes.TrackedVolume{
		Amount: expectedTakerFeeForStakers,
	}
	osmoutils.MustSet(s.Ctx.KVStore(s.App.GetKey(poolmanagertypes.StoreKey)), poolmanagertypes.KeyTakerFeeStakersProtoRev, &newTakerFeeForStakers)

	newTakerFeeForCommunityPool := poolmanagertypes.TrackedVolume{
		Amount: expectedTakerFeeForCommunityPool,
	}
	osmoutils.MustSet(s.Ctx.KVStore(s.App.GetKey(poolmanagertypes.StoreKey)), poolmanagertypes.KeyTakerFeeCommunityPoolProtoRev, &newTakerFeeForCommunityPool)

	// Set up cyclic arb tracker just to double check that it is not affected by the upgrade
	s.App.ProtoRevKeeper.SetCyclicArbProfitTrackerStartHeight(s.Ctx, expectedTrackerStartHeight)
	cyclicArbProfits := sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(9000)), sdk.NewCoin("Atom", osmomath.NewInt(3000)))
	err := s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, cyclicArbProfits[0].Denom, cyclicArbProfits[0].Amount)
	s.Require().NoError(err)
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, cyclicArbProfits[1].Denom, cyclicArbProfits[1].Amount)
	s.Require().NoError(err)

	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	allProtocolRevenue := s.App.ProtoRevKeeper.GetAllProtocolRevenue(s.Ctx)

	// Check that the taker fee tracker for stakers has been migrated correctly
	s.Require().Equal(expectedTakerFeeForStakers, allProtocolRevenue.TakerFeesTracker.TakerFeesToStakers)
	s.Require().Equal(expectedTakerFeeForCommunityPool, allProtocolRevenue.TakerFeesTracker.TakerFeesToCommunityPool)
	s.Require().Equal(expectedTrackerStartHeight, allProtocolRevenue.TakerFeesTracker.HeightAccountingStartsFrom)
	s.Require().Equal(expectedTrackerStartHeight, allProtocolRevenue.CyclicArbTracker.HeightAccountingStartsFrom)
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v22UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v22", Height: v22UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v22UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v22UpgradeHeight)
}
