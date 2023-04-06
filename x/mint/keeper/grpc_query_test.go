package keeper_test

import (
	"context"

	"github.com/osmosis-labs/osmosis/v15/x/mint/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	_, _, queryClient := suite.App, suite.Ctx, suite.queryClient

	_, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	_, err = queryClient.EpochProvisions(context.Background(), &types.QueryEpochProvisionsRequest{})
	suite.Require().NoError(err)
}
