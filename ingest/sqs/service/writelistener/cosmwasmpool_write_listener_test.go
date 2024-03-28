package writelistener_test

import (
	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/service"
	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/service/writelistener"
	"github.com/osmosis-labs/osmosis/v23/x/cosmwasmpool/model"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v23/x/cosmwasmpool/types"
)

// Tests that the cosmwasm write listener correctly tracks pool updates
// It ignores error cases as they are simple and unlikely giving app wiring.
func (s *WriteListenerTestSuite) TestWriteListener_CosmWasm() {

	// Set up chain state once per test
	s.Setup()

	cosmWasmPool := s.PrepareCosmWasmPool()

	cosmWasmPoolModel, ok := cosmWasmPool.(*model.Pool)
	s.Require().True(ok)

	concentratedPoolModelBz, err := cosmWasmPoolModel.Marshal()
	s.Require().NoError(err)

	testCases := []struct {
		name string

		key      []byte
		value    []byte
		isDelete bool

		expectedPoolUpdate bool
	}{
		{
			name: "random cosmwasm pool write, no-op",

			key:   cosmwasmpooltypes.KeyCodeIdWhitelist,
			value: someValue,
		},
		{
			name: "write cosmwasm pool pool index",

			key:   cosmwasmpooltypes.FormatPoolsPrefix(cosmWasmPoolModel.PoolId),
			value: concentratedPoolModelBz,

			expectedPoolUpdate: true,
		},
		{
			name: "delete cosmwasm pool index",

			key:   cosmwasmpooltypes.FormatPoolsPrefix(cosmWasmPoolModel.PoolId),
			value: concentratedPoolModelBz,

			expectedPoolUpdate: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {

			poolTracker := service.NewPoolTracker()

			cosmWasmPoolWriteListener := writelistener.NewCosmwasmPool(poolTracker)

			cosmwasmPoolKVStore := s.App.GetKey(cosmwasmpooltypes.ModuleName)

			err := cosmWasmPoolWriteListener.OnWrite(cosmwasmPoolKVStore, tc.key, tc.value, tc.isDelete)
			s.Require().NoError(err)

			// All non-cosmwasm pool updates should not be tracked.
			s.Require().Empty(poolTracker.GetCFMMPools())
			s.Require().Empty(poolTracker.GetConcentratedPools())

			if tc.expectedPoolUpdate {
				cosmWasmPools := poolTracker.GetCosmWasmPools()
				s.Require().Len(cosmWasmPools, 1)

				s.Require().Equal(&cosmWasmPoolModel.CosmWasmPool, cosmWasmPools[0])
			} else {
				cosmWasmPools := poolTracker.GetCosmWasmPools()
				s.Require().Len(cosmWasmPools, 0)
			}
		})
	}
}
