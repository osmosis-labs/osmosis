package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

// TestGetParams verifies that the params getter works correctly even after
// params modification.
func (s *KeeperTestSuite) TestGetParams() {
	// Query the initial params. They should be equal to the default
	resp, err := s.queryClient.Params(s.Ctx, new(types.QueryParamsRequest))
	s.Require().NoError(err)
	s.Require().Equal(types.DefaultParams(), resp.GetParams())

	// Append new asset to the assets list
	s.AppendNewAsset(asset1)

	// Fill expected values. Append new asset to the default assets list
	newParams := types.DefaultParams()
	newParams.Assets = append(newParams.Assets, asset1)

	// Query the result
	resp, err = s.queryClient.Params(s.Ctx, new(types.QueryParamsRequest))
	s.Require().NoError(err)
	s.Require().Equal(newParams, resp.GetParams())
}
