package v12_test

import (
	"fmt"
	"testing"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
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
func (suite *UpgradeTestSuite) TestPoolMigration() {
	testCases := []struct {
		name         string
		pre_upgrade  func() uint64
		upgrade      func()
		post_upgrade func(uint64)
	}{
		{
			"Test that the upgrade succeeds",
			func() uint64 {
				poolId := suite.PrepareBalancerPool()
				poolDenoms, err := suite.App.GAMMKeeper.GetPoolDenoms(suite.Ctx, poolId)
				suite.Require().NoError(err)

				_, err = suite.App.TwapKeeper.GetBeginBlockAccumulatorRecord(suite.Ctx, poolId, poolDenoms[0], poolDenoms[1])
				suite.Require().NoError(err)
				return poolId
			},
			func() {
				suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v12", Height: dummyUpgradeHeight}
				err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
				suite.Require().NoError(err)
				_, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
				suite.Require().True(exists)

				suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight)
				suite.Require().NotPanics(func() {
					beginBlockRequest := abci.RequestBeginBlock{}
					suite.App.BeginBlocker(suite.Ctx, beginBlockRequest)
				})
			},
			func(poolId uint64) {
				poolDenoms, err := suite.App.GAMMKeeper.GetPoolDenoms(suite.Ctx, poolId)
				suite.Require().NoError(err)

				_, err = suite.App.TwapKeeper.GetBeginBlockAccumulatorRecord(suite.Ctx, poolId, poolDenoms[0], poolDenoms[1])
				suite.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// creating pools before upgrade
			poolId := tc.pre_upgrade()

			// run upgrade
			tc.upgrade()

			// check that pool migration has been successfully done, did not break state
			tc.post_upgrade(poolId)
		})
	}
}
