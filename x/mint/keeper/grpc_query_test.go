package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v10/x/mint/types"
)

func TestMintGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestGRPCParams() {
	_, _, queryClient := suite.App, suite.Ctx, suite.queryClient

	_, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	_, err = queryClient.EpochProvisions(gocontext.Background(), &types.QueryEpochProvisionsRequest{})
	suite.Require().NoError(err)
}
