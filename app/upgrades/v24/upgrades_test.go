package v24_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	protorevtypes "github.com/osmosis-labs/osmosis/v23/x/protorev/types"
)

const (
	v24UpgradeHeight              = int64(10)
	HistoricalTWAPTimeIndexPrefix = "historical_time_index"
	KeySeparator                  = "|"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	// Set the old KVStore base denoms
	s.App.ProtoRevKeeper.DeprecatedSetBaseDenoms(s.Ctx, []protorevtypes.BaseDenom{
		{Denom: protorevtypes.OsmosisDenomination, StepSize: osmomath.NewInt(1_000_000)},
		{Denom: "atom", StepSize: osmomath.NewInt(1_000_000)},
		{Denom: "weth", StepSize: osmomath.NewInt(1_000_000)}})
	oldBaseDenoms, err := s.App.ProtoRevKeeper.DeprecatedGetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(3, len(oldBaseDenoms))
	s.Require().Equal(oldBaseDenoms[0].Denom, protorevtypes.OsmosisDenomination)
	s.Require().Equal(oldBaseDenoms[1].Denom, "atom")
	s.Require().Equal(oldBaseDenoms[2].Denom, "weth")

	// The new param store should return the default value
	newBaseDenoms := s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().Equal(protorevtypes.DefaultBaseDenoms, newBaseDenoms)

	dummyUpgrade(s)
	s.Require().NotPanics(func() {
		s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
	})

	// The new param store should return the old KVStore values
	newBaseDenoms = s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().Equal(oldBaseDenoms, newBaseDenoms)

	// The old KVStore base denoms should be deleted
	oldBaseDenoms, err = s.App.ProtoRevKeeper.DeprecatedGetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().Empty(oldBaseDenoms)
}

func dummyUpgrade(s *UpgradeTestSuite) {
	s.Ctx = s.Ctx.WithBlockHeight(v24UpgradeHeight - 1)
	plan := upgradetypes.Plan{Name: "v24", Height: v24UpgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, exists := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(exists)

	s.Ctx = s.Ctx.WithBlockHeight(v24UpgradeHeight)
}
