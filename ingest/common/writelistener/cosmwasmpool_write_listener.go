package writelistener

import (
	"bytes"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	commondomain "github.com/osmosis-labs/osmosis/v29/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v29/ingest/sqs/domain"
	cosmwasmpoolmodel "github.com/osmosis-labs/osmosis/v29/x/cosmwasmpool/model"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v29/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v29/x/poolmanager/types"
)

var _ commondomain.WriteListener = (*cosmwasmPoolWriteListener)(nil)

type cosmwasmPoolWriteListener struct {
	poolTracker domain.BlockPoolUpdateTracker
	wasmkeeper  *wasmkeeper.Keeper
}

func NewCosmwasmPool(poolTracker domain.BlockPoolUpdateTracker, wasmkeeper *wasmkeeper.Keeper) *cosmwasmPoolWriteListener {
	return &cosmwasmPoolWriteListener{
		poolTracker: poolTracker,
		wasmkeeper:  wasmkeeper,
	}
}

// OnWrite implements types.WriteListener
//
// NOTE: This only detects cwPools that have been created or migrated. It does not detect changes in balances (i.e. swaps / position creation / withdraws)
func (s *cosmwasmPoolWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	// Track the cwPool that was just created/migrated
	if len(key) > 0 && bytes.Equal(cosmwasmpooltypes.PoolsKey, key[:1]) {
		var cosmWasmPool cosmwasmpoolmodel.CosmWasmPool
		if err := cosmWasmPool.Unmarshal(value); err != nil {
			return err
		}

		pool := cosmwasmpoolmodel.Pool{
			CosmWasmPool: cosmWasmPool,
			WasmKeeper:   s.wasmkeeper,
		}

		s.poolTracker.TrackCosmWasm(&pool)

		// Create/modify the cwPool address to pool mapping
		// This is used to check if a balance change is for a cwPool address, and if so, we can retrieve the pool from this mapping
		var poolI poolmanagertypes.PoolI = &pool
		s.poolTracker.TrackCosmWasmPoolsAddressToPoolMap(poolI)
	}
	return nil
}

type cosmwasmPoolBalanceWriteListener struct {
	poolTracker domain.BlockPoolUpdateTracker
}

func NewCosmwasmPoolBalance(poolTracker domain.BlockPoolUpdateTracker) *cosmwasmPoolBalanceWriteListener {
	return &cosmwasmPoolBalanceWriteListener{
		poolTracker: poolTracker,
	}
}

// OnWrite implements types.WriteListener
// Tracks balance changes for cwPools (i.e. swaps / position creation / withdraws)
func (s *cosmwasmPoolBalanceWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	// Check if the key is a balance change for any address
	if len(key) > 0 && key[0] == banktypes.BalancesPrefix[0] {
		// The key is a balance change. Check if the address in question is a cwPool address

		// We expect the key to be of the form:
		// <prefix> (length 1)
		// <address_length> (length 1)
		// <address> (length address_length)
		addressLength := key[1]
		addressBytes := key[1+1 : 1+addressLength+1]
		address := sdk.AccAddress(addressBytes)
		addressStr := address.String()

		cwPoolMap := s.poolTracker.GetCosmWasmPoolsAddressToIDMap()
		if pool, ok := cwPoolMap[addressStr]; ok {
			// The address is a cwPool address. Add the cwPool to the cwPool tracker
			s.poolTracker.TrackCosmWasm(pool)
		}
	}
	return nil
}
