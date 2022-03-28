package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     *app.OsmosisApp
	querier keeper.Querier
	cleanup func()
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
	suite.querier = keeper.NewQuerier(*suite.app.LockupKeeper)
}

func (suite *KeeperTestSuite) SetupTestWithLevelDb() {
	suite.app, suite.cleanup = app.SetupTestingAppWithLevelDb(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
}

func (suite *KeeperTestSuite) Cleanup() {
	suite.cleanup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
