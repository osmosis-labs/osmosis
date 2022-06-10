package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v7/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochLifeCycle() {
	suite.SetupTest()

	epochInfo := types.NewGenesisEpochInfo("monthly", time.Hour*24*30)
	suite.App.EpochsKeeper.SetEpochInfo(suite.Ctx, epochInfo)
	epochInfoSaved := suite.App.EpochsKeeper.GetEpochInfo(suite.Ctx, "monthly")
	suite.Require().Equal(epochInfo, epochInfoSaved)

	allEpochs := suite.App.EpochsKeeper.AllEpochInfos(suite.Ctx)
	suite.Require().Len(allEpochs, 4)
	suite.Require().Equal(allEpochs[0].Identifier, "day") // alphabetical order
	suite.Require().Equal(allEpochs[1].Identifier, "hour")
	suite.Require().Equal(allEpochs[2].Identifier, "monthly")
	suite.Require().Equal(allEpochs[3].Identifier, "week")
}
