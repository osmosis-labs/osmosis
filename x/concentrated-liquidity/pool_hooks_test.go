package concentrated_liquidity_test

func (s *KeeperTestSuite) TestSetAndGetPoolHookContract() {
	validCosmwasmAddress := "osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9"
	invalidCosmwasmAddress := "osmo1{}{}4hj2tfpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9"
	validActionPrefix := "beforeSwap"
	validPoolId := uint64(1)

	tests := map[string]struct {
		cosmwasmAddress string
		actionPrefix    string
		poolId          uint64

		// We do boolean checks instead of exact error checks because any
		// expected errors would come from lower level calls that don't
		// conform to our error types.
		expectErrOnSet bool
	}{
		"basic valid test": {
			// Random correctly constructed address (we do not check contract existence at the layer)
			cosmwasmAddress: validCosmwasmAddress,
			actionPrefix:    validActionPrefix,
			poolId:          validPoolId,
		},
		"error: incorrectly constructed address": {

			cosmwasmAddress: invalidCosmwasmAddress,
			actionPrefix:    validActionPrefix,
			poolId:          validPoolId,

			expectErrOnSet: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// Set contract address using SetPoolHookContract
			err := s.clk.SetPoolHookContract(s.Ctx, 1, tc.actionPrefix, tc.cosmwasmAddress)

			// If expect error on set, check here
			if tc.expectErrOnSet {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// Get contract address
			contractAddress := s.clk.GetPoolHookContract(s.Ctx, 1, tc.actionPrefix)

			// Assertions
			s.Require().Equal(tc.cosmwasmAddress, contractAddress)
		})
	}
}

// TestCallPoolActionListener should be lightweight, as most of the testing logic will be in the higher level functions (e.g. shouldnt do a swap here)
// Errors:
// * gas limit exceeded (can just set hook to something that does consume gas > limit)
