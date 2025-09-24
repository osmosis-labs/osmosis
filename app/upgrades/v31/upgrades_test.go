package v31_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/app/apptesting"
	v31 "github.com/osmosis-labs/osmosis/v30/app/upgrades/v31"
	txfeestypes "github.com/osmosis-labs/osmosis/v30/x/txfees/types"
)

const (
	v31UpgradeHeight = int64(10)
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestTakerFeeDistributionSwap() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))

	// Prepare test by setting initial taker fee distribution
	s.PrepareTakerFeeDistributionTest()

	// Run the upgrade
	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		_, err := s.preModule.PreBlock(s.Ctx)
		s.Require().NoError(err)
	})

	// Verify the distribution was swapped correctly
	s.ExecuteTakerFeeDistributionTest()
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v31UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: v31.UpgradeName, Height: v31UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: v31UpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(v31UpgradeHeight)
}

// PrepareTakerFeeDistributionTest sets up the initial state with taker fees going to community pool
func (s *UpgradeTestSuite) PrepareTakerFeeDistributionTest() {
	// Get current poolmanager parameters
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)

	osmoTakerFeeDistribution := poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution
	osmoTakerFeeDistribution.StakingRewards = osmomath.MustNewDecFromStr("0.3")
	osmoTakerFeeDistribution.CommunityPool = osmomath.MustNewDecFromStr("0.7")
	osmoTakerFeeDistribution.Burn = osmomath.MustNewDecFromStr("0.0")

	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution = osmoTakerFeeDistribution
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)
}

// ExecuteTakerFeeDistributionTest verifies that the community_pool and burn values were swapped
func (s *UpgradeTestSuite) ExecuteTakerFeeDistributionTest() {
	// Get poolmanager parameters after upgrade
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)

	// Verify OSMO taker fee distribution
	s.Require().Equal(osmomath.MustNewDecFromStr("0.3"), poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.StakingRewards)
	s.Require().Equal(osmomath.MustNewDecFromStr("0.0"), poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.CommunityPool)
	s.Require().Equal(osmomath.MustNewDecFromStr("0.7"), poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.Burn)

	// Verify that the OSMO total still sums to 1.0
	osmoTotal := poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.CommunityPool.
		Add(poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.Burn).
		Add(poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.StakingRewards)
	s.Require().Equal(osmomath.OneDec(), osmoTotal)

	// Verify non-OSMO taker fee distribution
	s.Require().Equal(osmomath.MustNewDecFromStr("0.225"), poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards)
	s.Require().Equal(osmomath.MustNewDecFromStr("0.525"), poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.Burn)
	s.Require().Equal(osmomath.MustNewDecFromStr("0.25"), poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool)

	// Verify that the non-OSMO total sums to 1.0
	nonOsmoTotal := poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool.
		Add(poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.Burn).
		Add(poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards)
	s.Require().Equal(osmomath.OneDec(), nonOsmoTotal)

	// Verify the module account is set correctly
	takerFeeBurnModuleAccount := s.App.AccountKeeper.GetModuleAccount(s.Ctx, txfeestypes.TakerFeeBurnName)
	s.Require().Equal(txfeestypes.TakerFeeBurnName, takerFeeBurnModuleAccount.GetName())
	s.Require().Equal([]string{}, takerFeeBurnModuleAccount.GetPermissions())
}
