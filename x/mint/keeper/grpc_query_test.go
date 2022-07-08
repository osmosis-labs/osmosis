package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v7/x/mint/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	_, _, queryClient := suite.App, suite.Ctx, suite.queryClient

	_, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	_, err = queryClient.EpochProvisions(gocontext.Background(), &types.QueryEpochProvisionsRequest{})
	suite.Require().NoError(err)
}

func TestMintTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
