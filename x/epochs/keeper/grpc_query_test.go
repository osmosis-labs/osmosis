package keeper_test

import (
	gocontext "context"

	"github.com/osmosis-labs/osmosis/v10/x/epochs/types"
)

func (suite *KeeperTestSuite) TestQueryEpochInfos() {
	suite.SetupTest()
	queryClient := suite.queryClient

	// Check that querying epoch infos on default genesis returns the default genesis epoch infos
	epochInfosResponse, err := queryClient.EpochInfos(gocontext.Background(), &types.QueryEpochsInfoRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(epochInfosResponse.Epochs, 3)
	expectedEpochs := types.DefaultGenesis().Epochs
	for id := range expectedEpochs {
		expectedEpochs[id].StartTime = suite.Ctx.BlockTime()
		expectedEpochs[id].CurrentEpochStartHeight = suite.Ctx.BlockHeight()
	}

	suite.Require().Equal(expectedEpochs, epochInfosResponse.Epochs)
}
