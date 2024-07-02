package writelistener

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	gammtypes "github.com/osmosis-labs/osmosis/v25/x/gamm/types"
)

<<<<<<< HEAD:ingest/sqs/service/writelistener/cfmm_write_listener.go
var _ storetypes.WriteListener = (*cfmmPoolWriteListener)(nil)
=======
var _ commondomain.WriteListener = (*cfmmPoolWriteListener)(nil)
>>>>>>> 415f64ab (refactor(indexer): create ingest/common package (#8471)):ingest/common/writelistener/cfmm_write_listener.go

type cfmmPoolWriteListener struct {
	poolTracker domain.BlockPoolUpdateTracker
	codec       codec.Codec
}

func NewGAMM(poolTracker domain.BlockPoolUpdateTracker, appCodec codec.Codec) storetypes.WriteListener {
	return &cfmmPoolWriteListener{
		poolTracker: poolTracker,
		codec:       appCodec,
	}
}

// OnWrite implements types.WriteListener.
func (s *cfmmPoolWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	// Track the changed pool.
	if len(key) > 0 && bytes.Equal(gammtypes.KeyPrefixPools, key[:1]) {
		var pool gammtypes.CFMMPoolI
		if err := s.codec.UnmarshalInterface(value, &pool); err != nil {
			return err
		}

		s.poolTracker.TrackCFMM(pool)
	}

	return nil
}
