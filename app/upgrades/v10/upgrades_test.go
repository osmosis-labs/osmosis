package v10_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	v10 "github.com/osmosis-labs/osmosis/v15/app/upgrades/v10"
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

func (suite *UpgradeTestSuite) TestUpgradePayments() {
	testCases := []struct {
		msg     string
		upgrade func()
	}{
		{
			"Test that upgrade succeeds",
			func() {
				// run upgrade
				// First run block N-1, begin new block takes ctx height + 1
				suite.Ctx = suite.Ctx.WithBlockHeight(v10.ForkHeight - 2)
				suite.BeginNewBlock(false)

				// run upgrade height
				suite.Require().NotPanics(func() {
					suite.BeginNewBlock(false)
				})
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			tc.upgrade()
		})
	}
}
