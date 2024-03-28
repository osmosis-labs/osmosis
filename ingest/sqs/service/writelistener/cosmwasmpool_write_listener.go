package writelistener

import (
	"bytes"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

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
