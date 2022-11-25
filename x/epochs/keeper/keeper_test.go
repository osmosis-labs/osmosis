package keeper_test

import (
	"testing"

	"github.com/osmosis-labs/osmosis/x/epochs/simapp/apptesting"
	"github.com/osmosis-labs/osmosis/x/epochs/types"

	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
