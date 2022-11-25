package apptesting

import (
	"time"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/epochs/simapp"
	"github.com/stretchr/testify/suite"
)

type KeeperTestHelper struct {
	suite.Suite

	App         *simapp.SimApp
	Ctx         sdk.Context
	QueryHelper *baseapp.QueryServiceTestHelper
}

func (suite *KeeperTestHelper) SetupTest() {
	suite.Setup()
}

func (suite *KeeperTestHelper) Setup() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	suite.App = app
	suite.Ctx = ctx
	suite.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: suite.App.GRPCQueryRouter(),
		Ctx:             suite.Ctx,
	}

	suite.SetEpochStartTime()

}

func (suite *KeeperTestHelper) SetEpochStartTime() {
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
