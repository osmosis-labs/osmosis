package keeper_test

import (
	"context"

	"github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives/types"
)

func (s *KeeperTestSuite) TestParams() {
	s.SetupTest()

	queryClient := s.queryClient

	res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)

	s.Require().Empty(res.Params.DistributionContractAddress)

}
