package v12_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v28/app/apptesting"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
	preModule appmodule.HasPreBlocker
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
	s.preModule = upgrade.NewAppModule(s.App.UpgradeKeeper, addresscodec.NewBech32Codec("osmo"))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const dummyUpgradeHeight = 5

// note that this test does not perfectly test state migration, as pre_upgrade function
// which includes setting up tests by creating new pools would also create twap records
// automatically with this binary. The goal of this test is to test that the upgrade handler
// does not panic upon upgrade.
// Detailed state migration tests are placed within the twap keeper.
func (s *UpgradeTestSuite) TestPoolMigration() {
	testCases := []struct {
		name         string
		pre_upgrade  func() uint64
		upgrade      func()
		post_upgrade func(uint64)
	}{
		{
			"Test that the upgrade succeeds",
			func() uint64 {
				poolId := s.PrepareBalancerPool()
				poolDenoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, poolId)
				s.Require().NoError(err)

				_, err = s.App.TwapKeeper.GetBeginBlockAccumulatorRecord(s.Ctx, poolId, poolDenoms[0], poolDenoms[1])
				s.Require().NoError(err)
				return poolId
			},
			func() {
				s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v12", Height: dummyUpgradeHeight}
				err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
				s.Require().NoError(err)
				_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
				s.Require().NoError(err)

				s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: dummyUpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(dummyUpgradeHeight)
				s.Require().NotPanics(func() {
					_, err := s.preModule.PreBlock(s.Ctx)
					s.Require().NoError(err)
				})
			},
			func(poolId uint64) {
				poolDenoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, poolId)
				s.Require().NoError(err)

				_, err = s.App.TwapKeeper.GetBeginBlockAccumulatorRecord(s.Ctx, poolId, poolDenoms[0], poolDenoms[1])
				s.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			s.SetupTest() // reset

			// creating pools before upgrade
			poolId := tc.pre_upgrade()

			// run upgrade
			tc.upgrade()

			// check that pool migration has been successfully done, did not break state
			tc.post_upgrade(poolId)
		})
	}
}
