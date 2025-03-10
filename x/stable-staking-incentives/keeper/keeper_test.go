package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v26/app/apptesting"
	"github.com/osmosis-labs/osmosis/v26/x/stable-staking-incentives/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	s.queryClient = types.NewQueryClient(s.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
