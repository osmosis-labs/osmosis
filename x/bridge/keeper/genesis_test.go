package keeper_test

// TestGenesis tests that the default genesis state is valid and vdlidates
// that all genesis denoms appear in tokenfactory
func (s *KeeperTestSuite) TestGenesis() {
	// Get all tf denoms created by the bridge module
	tfDenoms := s.GetBridgeTFDenoms()

	// Get all denoms based on the assets stored in the module params
	bridgeDenoms := s.GetBridgeDenoms()

	// Compare two denom lists
	s.Require().ElementsMatch(tfDenoms, bridgeDenoms)
}
