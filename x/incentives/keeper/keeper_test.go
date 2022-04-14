package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/incentives/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     *app.OsmosisApp
	querier keeper.Querier
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
	lockableDurations := suite.app.IncentivesKeeper.GetLockableDurations(suite.ctx)
	lockableDurations = append(lockableDurations, 2*time.Second)
	suite.app.IncentivesKeeper.SetLockableDurations(suite.ctx, lockableDurations)

	suite.querier = keeper.NewQuerier(*suite.app.IncentivesKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
