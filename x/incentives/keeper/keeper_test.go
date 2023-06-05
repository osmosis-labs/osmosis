package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
	"github.com/osmosis-labs/osmosis/v16/x/incentives/keeper"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	querier keeper.Querier
}

// SetupTest sets incentives parameters from the suite's context
func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.querier = keeper.NewQuerier(*s.App.IncentivesKeeper)
	lockableDurations := s.App.IncentivesKeeper.GetLockableDurations(s.Ctx)
	lockableDurations = append(lockableDurations, 2*time.Second)
	s.App.IncentivesKeeper.SetLockableDurations(s.Ctx, lockableDurations)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
