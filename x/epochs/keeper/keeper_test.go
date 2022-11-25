package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/epochs/simapp"
	"github.com/osmosis-labs/osmosis/x/epochs/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	App         *simapp.SimApp
	Ctx         sdk.Context
	QueryHelper *baseapp.QueryServiceTestHelper
	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) Setup() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	suite.App = app
	suite.Ctx = ctx
	suite.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: suite.App.GRPCQueryRouter(),
		Ctx:             suite.Ctx,
	}
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)

	suite.SetEpochStartTime()

}

func (suite *KeeperTestSuite) SetEpochStartTime() {
	epochsKeeper := suite.App.EpochsKeeper

	for _, epoch := range epochsKeeper.AllEpochInfos(suite.Ctx) {
		epoch.StartTime = suite.Ctx.BlockTime()
		epochsKeeper.DeleteEpochInfo(suite.Ctx, epoch.Identifier)
		err := epochsKeeper.AddEpochInfo(suite.Ctx, epoch)
		if err != nil {
			panic(err)
		}
	}
}
