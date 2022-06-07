package v9_test

import (
	"fmt"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const dummyUpgradeHeight = 5

func (suite *UpgradeTestSuite) TestUpgradePayments() {
	testCases := []struct {
		msg         string
		pre_update  func()
		update      func()
		post_update func()
		expPass     bool
	}{
		{
			"Test that upgrade does not panic",
			func() {
				// Create pool 1
				suite.PrepareBalancerPool()
			},
			func() {
				// run upgrade
				// TODO: Refactor this all into a helper fn
				suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v9", Height: dummyUpgradeHeight}
				err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx, plan)
				suite.Require().NoError(err)
				plan, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx)
				suite.Require().True(exists)

				suite.Ctx = suite.Ctx.WithBlockHeight(dummyUpgradeHeight)
				suite.Require().NotPanics(func() {
					beginBlockRequest := abci.RequestBeginBlock{}
					suite.App.BeginBlocker(suite.Ctx, beginBlockRequest)
				})
			},
			func() {
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.pre_update()
			tc.update()
			tc.post_update()
		})
	}
}
