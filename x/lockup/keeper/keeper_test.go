package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v19/app"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	"github.com/osmosis-labs/osmosis/v19/x/lockup/keeper"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	querier keeper.Querier
	cleanup func()
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.querier = keeper.NewQuerier(*s.App.LockupKeeper)
	unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime
	s.App.IncentivesKeeper.SetLockableDurations(s.Ctx, []time.Duration{
		time.Hour * 24 * 14,
		time.Hour,
		time.Hour * 3,
		time.Hour * 7,
		unbondingDuration,
	})
}

func (s *KeeperTestSuite) SetupTestWithLevelDb() {
	s.App, s.cleanup = app.SetupTestingAppWithLevelDb(false)
	s.Ctx = s.App.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
}

func (s *KeeperTestSuite) Cleanup() {
	s.cleanup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
