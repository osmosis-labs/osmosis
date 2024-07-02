package writelistener

import (
	"bytes"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v25/x/concentrated-liquidity/model"
	concentratedtypes "github.com/osmosis-labs/osmosis/v25/x/concentrated-liquidity/types"
)

<<<<<<< HEAD:ingest/sqs/service/writelistener/concentrated_write_listener.go
var _ storetypes.WriteListener = (*concentratedPoolWriteListener)(nil)
=======
var _ commondomain.WriteListener = (*concentratedPoolWriteListener)(nil)
>>>>>>> 415f64ab (refactor(indexer): create ingest/common package (#8471)):ingest/common/writelistener/concentrated_write_listener.go

type concentratedPoolWriteListener struct {
	poolTracker domain.BlockPoolUpdateTracker
}

func NewConcentrated(poolTracker domain.BlockPoolUpdateTracker) storetypes.WriteListener {
	return &concentratedPoolWriteListener{
		poolTracker: poolTracker,
	}
}

// OnWrite implements types.WriteListener.
func (s *concentratedPoolWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	if len(key) == 0 {
		return nil
	}

	// Process pool write
	if bytes.Equal(concentratedtypes.PoolPrefix, key[:1]) {
		pool := model.Pool{}

		if err := pool.Unmarshal(value); err != nil {
			return err
		}

		// Track the changed pool.
		s.poolTracker.TrackConcentrated(&pool)
	}

	// Process pool tick write
	if bytes.Equal(concentratedtypes.TickPrefix, key[:1]) {
		poolIDPrefixBz := key[len(concentratedtypes.TickPrefix) : concentratedtypes.KeyTickPrefixByPoolIdLengthBytes+1]

		poolID := sdk.BigEndianToUint64(poolIDPrefixBz)

		// We simply track the pool ID so that we can read the pool and all its ticks
		// from the store at the end of the block.
		s.poolTracker.TrackConcentratedPoolIDTickChange(poolID)
	}

	return nil
}
