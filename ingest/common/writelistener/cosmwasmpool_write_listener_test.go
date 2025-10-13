package writelistener_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankmigv2 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v2"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v30/ingest/common/pooltracker"
	"github.com/osmosis-labs/osmosis/v30/ingest/common/writelistener"
	"github.com/osmosis-labs/osmosis/v30/x/cosmwasmpool/model"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v30/x/cosmwasmpool/types"
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
			name: "cosmwasm pool write unrelated to pool state, no-op",

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

			poolTracker := pooltracker.NewMemory()

			cosmWasmPoolWriteListener := writelistener.NewCosmwasmPool(poolTracker, s.App.WasmKeeper)

			cosmwasmPoolKVStore := s.App.GetKey(cosmwasmpooltypes.ModuleName)

			err := cosmWasmPoolWriteListener.OnWrite(cosmwasmPoolKVStore, tc.key, tc.value, tc.isDelete)
			s.Require().NoError(err)

			// All non-cosmwasm pool updates should not be tracked.
			s.Require().Empty(poolTracker.GetCFMMPools())
			s.Require().Empty(poolTracker.GetConcentratedPools())

			if tc.expectedPoolUpdate {
				cosmWasmPools := poolTracker.GetCosmWasmPools()
				s.Require().Len(cosmWasmPools, 1)

				s.Require().Equal(&cosmWasmPoolModel.CosmWasmPool, cosmWasmPools[0].AsSerializablePool())
			} else {
				cosmWasmPools := poolTracker.GetCosmWasmPools()
				s.Require().Len(cosmWasmPools, 0)
			}
		})
	}
}

func (s *WriteListenerTestSuite) TestWriteListener_CosmWasmBalance() {
	// Set up chain state once per test
	s.Setup()

	cosmWasmPool := s.PrepareCosmWasmPool()

	cosmWasmPoolModel, ok := cosmWasmPool.(*model.Pool)
	s.Require().True(ok)

	concentratedPoolModelBz, err := cosmWasmPoolModel.Marshal()
	s.Require().NoError(err)

	// Trigger cwPool write listener actions at pool creation
	poolTracker := pooltracker.NewMemory()
	cosmWasmPoolWriteListener := writelistener.NewCosmwasmPool(poolTracker, s.App.WasmKeeper)
	cosmwasmPoolKVStore := s.App.GetKey(cosmwasmpooltypes.ModuleName)
	err = cosmWasmPoolWriteListener.OnWrite(cosmwasmPoolKVStore, cosmwasmpooltypes.FormatPoolsPrefix(cosmWasmPoolModel.PoolId), concentratedPoolModelBz, false)
	s.Require().NoError(err)

	// Reset pool tracker, so it will flush the pool changes tracker but keep the address to pool mapping
	poolTracker.Reset()

	testCases := []struct {
		name string

		key      []byte
		value    []byte
		isDelete bool

		expectedPoolUpdate bool
	}{
		{
			name: "balance write unrelated to cosmwasm pool, no-op",

			key:   bankmigv2.CreateAccountBalancesPrefix(s.TestAccs[0]),
			value: someValue, // value is not used for balance changes
		},
		{
			name: "balance write to cosmwasm pool",

			key:   bankmigv2.CreateAccountBalancesPrefix(sdk.MustAccAddressFromBech32(cosmWasmPoolModel.ContractAddress)),
			value: someValue, // value is not used for balance changes

			expectedPoolUpdate: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cosmWasmPoolBalanceWriteListener := writelistener.NewCosmwasmPoolBalance(poolTracker)

			bankKVStore := s.App.GetKey(banktypes.ModuleName)

			err := cosmWasmPoolBalanceWriteListener.OnWrite(bankKVStore, tc.key, tc.value, tc.isDelete)
			s.Require().NoError(err)

			// All non-cosmwasm pool updates should not be tracked.
			s.Require().Empty(poolTracker.GetCFMMPools())
			s.Require().Empty(poolTracker.GetConcentratedPools())

			if tc.expectedPoolUpdate {
				cosmWasmPools := poolTracker.GetCosmWasmPools()
				s.Require().Len(cosmWasmPools, 1)

				s.Require().Equal(&cosmWasmPoolModel.CosmWasmPool, cosmWasmPools[0].AsSerializablePool())
			} else {
				cosmWasmPools := poolTracker.GetCosmWasmPools()
				s.Require().Len(cosmWasmPools, 0)
			}
		})
	}
}
