package v10_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	v10 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v10"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgradePayments() {
	testCases := []struct {
		msg     string
		upgrade func()
	}{
		{
			"Test that upgrade succeeds",
			func() {
				// run upgrade
				// First run block N-1, begin new block takes ctx height + 1
				s.Ctx = s.Ctx.WithBlockHeight(v10.ForkHeight - 2)
				s.BeginNewBlock(false)

				// run upgrade height
				s.Require().NotPanics(func() {
					s.BeginNewBlock(false)
				})
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			s.SetupTest() // reset
			tc.upgrade()
		})
	}
}
