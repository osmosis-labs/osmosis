package gov_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/app"
	"github.com/osmosis-labs/osmosis/v10/x/superfluid/keeper"
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

func assertEventEmitted(suite *KeeperTestSuite, ctx sdk.Context, eventTypeExpected string, numEventsExpected int) {
	allEvents := ctx.EventManager().Events()
	// filter out other events
	actualEvents := make([]sdk.Event, 0, 1)
	for _, event := range allEvents {
		if event.Type == eventTypeExpected {
			actualEvents = append(actualEvents, event)
		}
	}
	suite.Require().Equal(numEventsExpected, len(actualEvents))
}
