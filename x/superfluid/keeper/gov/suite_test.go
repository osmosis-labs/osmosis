package gov_test

import (
	"testing"

	"github.com/osmosis-labs/osmosis/v17/app/apptesting"
	"github.com/osmosis-labs/osmosis/v17/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v17/x/superfluid/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	querier types.QueryServer
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.querier = keeper.NewQuerier(*s.App.SuperfluidKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
