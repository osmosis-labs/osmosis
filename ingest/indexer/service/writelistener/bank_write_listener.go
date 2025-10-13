package writelistener

import (
	"bytes"
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	commondomain "github.com/osmosis-labs/osmosis/v31/ingest/common/domain"
	indexerdomain "github.com/osmosis-labs/osmosis/v31/ingest/indexer/domain"
)

var _ commondomain.WriteListener = (*bankWriteListener)(nil)

type bankWriteListener struct {

	// shared context to handle graceful shutdown in case of node stop.
	ctx context.Context

	client indexerdomain.TokenSupplyPublisher

	blockProcessStrategyManager commondomain.BlockProcessStrategyManager
}

func NewBank(ctx context.Context, client indexerdomain.TokenSupplyPublisher, blockProcessStrategyManager commondomain.BlockProcessStrategyManager) commondomain.WriteListener {
	return &bankWriteListener{
		ctx:    ctx,
		client: client,

		blockProcessStrategyManager: blockProcessStrategyManager,
	}
}

// OnWrite implements types.WriteListener.
// OnWrite is called when a write operation is executed on the KVStore.
// If the modified key is a supply key, the updated supply is published to the pubsub client.
// If the modified key is a supply offset key, the updated supply offset is published to the pubsub client.
// If the cold start manager has not ingested initial data, an error is returned.
// For any other key, no action is taken.
// delete parameter is ignored.
func (s *bankWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	if s.blockProcessStrategyManager.ShouldPushAllData() {
		return indexerdomain.ErrDidNotIngestAllData
	}

	// If the cold start manager has ingested initial data and the key is not empty and the key is a supply key.
	if len(key) > 0 && bytes.Equal(banktypes.SupplyKey, key[:1]) {
		denom := string(key[len(banktypes.SupplyKey):])
		if !indexerdomain.ShouldFilterDenom(denom) {
			if err := s.publishSupply(denom, value); err != nil {
				return err
			}
		}
	}

	// If the cold start manager has ingested initial data and the key is not empty and the key is a supply offset key.
	if len(key) > 0 && bytes.Equal(banktypes.SupplyOffsetKey, key[:1]) {
		denom := string(key[len(banktypes.SupplyOffsetKey):])
		if !indexerdomain.ShouldFilterDenom(denom) {
			if err := s.publishSupplyOffset(denom, value); err != nil {
				return err
			}
		}
	}

	return nil
}

// publishSupply publishes the updated supply to the pubsub client.
func (s *bankWriteListener) publishSupply(denom string, updatedSupplyBytes []byte) error {
	// Track updated supplies.
	var updatedSupply osmomath.Int
	err := updatedSupply.Unmarshal(updatedSupplyBytes)
	if err != nil {
		return fmt.Errorf("unable to unmarshal supply value %v", err)
	}

	tokenSupply := indexerdomain.TokenSupply{
		// Denom is the key without the supply prefix.
		Denom:  denom,
		Supply: updatedSupply,
	}

	err = s.client.PublishTokenSupply(s.ctx, tokenSupply)
	if err != nil {
		return fmt.Errorf("unable to publish token supply %v", err)
	}

	return nil
}

// publishSupplyOffset publishes the updated supply offset to the pubsub client.
func (s *bankWriteListener) publishSupplyOffset(denom string, updatedSupplyOffsetBytes []byte) error {
	// Track updated supplies.
	var updatedSupplyOffset osmomath.Int
	err := updatedSupplyOffset.Unmarshal(updatedSupplyOffsetBytes)
	if err != nil {
		return fmt.Errorf("unable to unmarshal supply value %v", err)
	}

	tokenSupplyOffset := indexerdomain.TokenSupplyOffset{
		// Denom is the key without the supply offset prefix.
		Denom:        denom,
		SupplyOffset: updatedSupplyOffset,
	}

	err = s.client.PublishTokenSupplyOffset(s.ctx, tokenSupplyOffset)
	if err != nil {
		return fmt.Errorf("unable to publish token supply %v", err)
	}

	return nil
}
