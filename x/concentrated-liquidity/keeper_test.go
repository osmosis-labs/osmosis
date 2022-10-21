package concentrated_liquidity_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}
