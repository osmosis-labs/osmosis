package balancer_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v10/app/apptesting"
	v10 "github.com/osmosis-labs/osmosis/v10/app/upgrades/v10"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
	// be post-bug
	suite.Ctx = suite.Ctx.WithBlockHeight(v10.ForkHeight)
}
