package writelistener_test

import (
	"github.com/osmosis-labs/osmosis/v30/ingest/common/pooltracker"
	"github.com/osmosis-labs/osmosis/v30/ingest/common/writelistener"
	gammtypes "github.com/osmosis-labs/osmosis/v30/x/gamm/types"
)

// Tests that the concentrated write listener correctly tracks pool and tick updates.
// It ignores error cases as they are simple and unlikely giving app wiring.
func (s *WriteListenerTestSuite) TestWriteListener_CFMM() {

	// Set up chain state once per test
	s.Setup()

	allPools := s.PrepareAllSupportedPools()

	// Get balancer pool
	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, allPools.BalancerPoolID)
	s.Require().NoError(err)

	balancerModelBz, err := s.App.AppCodec().MarshalInterface(balancerPool)
	s.Require().NoError(err)

	// Get stableswap pool
	stableswapPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, allPools.StableSwapPoolID)
	s.Require().NoError(err)

	stableswapModelBz, err := s.App.AppCodec().MarshalInterface(stableswapPool)
	s.Require().NoError(err)

	testCases := []struct {
		name string

		key      []byte
		value    []byte
		isDelete bool

		expectedPoolID uint64
	}{
		{
			name: "gamm write unrelated to pool state, no-op",

			key:   gammtypes.KeyNextGlobalPoolId,
			value: someValue,
		},
		{
			name: "write pool index balancer",

			key:   gammtypes.GetKeyPrefixPools(allPools.BalancerPoolID),
			value: balancerModelBz,

			expectedPoolID: allPools.BalancerPoolID,
		},
		{
			name: "delete pool index balancer",

			key:      gammtypes.GetKeyPrefixPools(allPools.BalancerPoolID),
			value:    balancerModelBz,
			isDelete: true,

			expectedPoolID: allPools.BalancerPoolID,
		},
		{
			name: "write pool index stableswap",

			key:   gammtypes.GetKeyPrefixPools(allPools.StableSwapPoolID),
			value: stableswapModelBz,

			expectedPoolID: allPools.StableSwapPoolID,
		},
		{
			name: "delete pool index stableswap",

			key:      gammtypes.GetKeyPrefixPools(allPools.StableSwapPoolID),
			value:    stableswapModelBz,
			isDelete: true,

			expectedPoolID: allPools.StableSwapPoolID,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {

			poolTracker := pooltracker.NewMemory()

			gammWriteListener := writelistener.NewGAMM(poolTracker, s.App.AppCodec())

			gammKVStore := s.App.GetKey(gammtypes.ModuleName)

			err := gammWriteListener.OnWrite(gammKVStore, tc.key, tc.value, tc.isDelete)
			s.Require().NoError(err)

			// All non-cfmm pool updates should not be tracked.
			s.Require().Empty(poolTracker.GetConcentratedPools())
			s.Require().Empty(poolTracker.GetCosmWasmPools())
			s.Require().Empty(poolTracker.GetConcentratedPoolIDTickChange())

			if tc.expectedPoolID != 0 {
				cfmmPools := poolTracker.GetCFMMPools()
				s.Require().Len(cfmmPools, 1)

				s.Require().Equal(tc.expectedPoolID, cfmmPools[0].GetId())
			} else {
				cfmmPools := poolTracker.GetCFMMPools()
				s.Require().Len(cfmmPools, 0)
			}
		})
	}
}
