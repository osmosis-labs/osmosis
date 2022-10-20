package concentrated_liquidity_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	concentrated_liquidity "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	clkeeper *concentrated_liquidity.Keeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}
