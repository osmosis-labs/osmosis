package service

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/osmosis-labs/osmosis/v23/x/concentrated-liquidity/model"
	concentratedtypes "github.com/osmosis-labs/osmosis/v23/x/concentrated-liquidity/types"
)

var _ sdk.WriteListener = (*concentratedListener)(nil)

type concentratedListener struct {
	poolTracker PoolTracker
}

func NewConcentratedWriteListerner(poolTracker PoolTracker) sdk.WriteListener {
	return &concentratedListener{
		poolTracker: poolTracker,
	}
}

// OnWrite implements types.WriteListener.
func (s *concentratedListener) OnWrite(storeKey sdk.StoreKey, key []byte, value []byte, delete bool) error {
	storeKeyName := storeKey.Name()

	if storeKeyName == concentratedtypes.ModuleName {

		if bytes.Equal(concentratedtypes.PoolPrefix, key) {
			pool := model.Pool{}

			if err := pool.Unmarshal(value); err != nil {
				return err
			}

			if err := s.poolTracker.TrackConcentrated(pool); err != nil {
				return err
			}
		}

	}

	return nil
}
