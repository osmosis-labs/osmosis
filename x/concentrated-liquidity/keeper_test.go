package concentrated_liquidity_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	concentrated_liquidity "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
)

type TestSuite struct {
	apptesting.KeeperTestHelper
	concentratedLiquidityKeeper *concentrated_liquidity.Keeper
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupTest() {
	s.Setup()
}
