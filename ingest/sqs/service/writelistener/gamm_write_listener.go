package writelistener

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
	gammtypes "github.com/osmosis-labs/osmosis/v23/x/gamm/types"
)

var _ storetypes.WriteListener = (*concentratedPoolWriteListener)(nil)

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
