package writelistener

import (
	"bytes"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/osmosis-labs/osmosis/v24/ingest/sqs/domain"
	cosmwasmpoolmodel "github.com/osmosis-labs/osmosis/v24/x/cosmwasmpool/model"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v24/x/cosmwasmpool/types"
)

var _ storetypes.WriteListener = (*cosmwasmPoolWriteListener)(nil)

type cosmwasmPoolWriteListener struct {
	poolTracker domain.BlockPoolUpdateTracker
}

func NewCosmwasmPool(poolTracker domain.BlockPoolUpdateTracker) storetypes.WriteListener {
	return &cosmwasmPoolWriteListener{
		poolTracker: poolTracker,
	}
}

// OnWrite implements types.WriteListener.
func (s *cosmwasmPoolWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	if storeKey.Name() == wasmtypes.StoreKey {
		fmt.Println("AAA key", key)
		fmt.Println("AAA key string", string(key))
		fmt.Println("AAA value", value)
		fmt.Println("AAA value string", string(value))
		fmt.Println()
	}
	// Track the changed pool.
	if len(key) > 0 && bytes.Equal(cosmwasmpooltypes.PoolsKey, key[:1]) {
		var pool cosmwasmpoolmodel.CosmWasmPool
		if err := pool.Unmarshal(value); err != nil {
			return err
		}

		s.poolTracker.TrackCosmWasm(&pool)
	}
	return nil
}
