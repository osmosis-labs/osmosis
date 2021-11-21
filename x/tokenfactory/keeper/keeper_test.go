package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/app"
	"github.com/osmosis-labs/osmosis/x/tokenfactory/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	querier     sdk.Querier
	app         *app.OsmosisApp
	queryClient types.QueryClient
	cleanup     func()
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: time.Now().UTC()})

	suite.app.TokenFactoryKeeper.CreateModuleAccount(suite.ctx)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.TokenFactoryKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
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
