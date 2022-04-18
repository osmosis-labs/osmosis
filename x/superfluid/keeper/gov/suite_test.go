package gov_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	querier keeper.Querier
	app     *app.OsmosisApp
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})
	suite.querier = keeper.NewQuerier(*suite.app.SuperfluidKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
