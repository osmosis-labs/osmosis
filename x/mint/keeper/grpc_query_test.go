package keeper_test

import (
	gocontext "context"

	"github.com/osmosis-labs/osmosis/v11/x/mint/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	_, _, queryClient := suite.App, suite.Ctx, suite.queryClient

	_, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	_, err = queryClient.EpochProvisions(gocontext.Background(), &types.QueryEpochProvisionsRequest{})
	suite.Require().NoError(err)
}
