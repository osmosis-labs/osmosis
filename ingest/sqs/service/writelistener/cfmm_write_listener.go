package writelistener

import (
	"bytes"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	gammtypes "github.com/osmosis-labs/osmosis/v25/x/gamm/types"
)

var _ domain.WriteListener = (*cfmmPoolWriteListener)(nil)

type cfmmPoolWriteListener struct {
	poolTracker domain.BlockPoolUpdateTracker
	codec       codec.Codec
}

func NewGAMM(poolTracker domain.BlockPoolUpdateTracker, appCodec codec.Codec) *cfmmPoolWriteListener {
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
