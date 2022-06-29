package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v7/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochLifeCycle() {
	suite.SetupTest()

	epochInfo := types.NewGenesisEpochInfo("monthly", time.Hour*24*30)
	suite.App.EpochsKeeper.AddEpochInfo(suite.Ctx, epochInfo)
	epochInfoSaved := suite.App.EpochsKeeper.GetEpochInfo(suite.Ctx, "monthly")
	// setup expected epoch info
	expectedEpochInfo := epochInfo
	expectedEpochInfo.StartTime = suite.Ctx.BlockTime()
	expectedEpochInfo.CurrentEpochStartHeight = suite.Ctx.BlockHeight()
	suite.Require().Equal(expectedEpochInfo, epochInfoSaved)

	allEpochs := suite.App.EpochsKeeper.AllEpochInfos(suite.Ctx)
	suite.Require().Len(allEpochs, 4)
	suite.Require().Equal(allEpochs[0].Identifier, "day") // alphabetical order
	suite.Require().Equal(allEpochs[1].Identifier, "hour")
	suite.Require().Equal(allEpochs[2].Identifier, "monthly")
	suite.Require().Equal(allEpochs[3].Identifier, "week")
}
