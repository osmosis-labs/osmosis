package v9_test

import (
	"fmt"
	"time"

	"cosmossdk.io/core/header"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

const dummyUpgradeHeight = 5

func (s *UpgradeTestSuite) TestUpgradePayments() {
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
				s.PrepareBalancerPool()
			},
			func() {
				// run upgrade
				// TODO: Refactor this all into a helper fn
				s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v9", Height: dummyUpgradeHeight}
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
			func() {
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			s.SetupTest() // reset

			tc.pre_update()
			tc.update()
			tc.post_update()
		})
	}
}
