package writelistener

import (
	"bytes"
	"context"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

var _ storetypes.WriteListener = (*bankWriteListener)(nil)

type bankWriteListener struct {

	// shared context to handle graceful shutdown in case of node stop.
	ctx context.Context

	client indexerdomain.Ingester

	coldStartManager indexerdomain.ColdStartManager
}

func NewBank(ctx context.Context, client indexerdomain.Ingester, coldStartManager indexerdomain.ColdStartManager) storetypes.WriteListener {
	return &bankWriteListener{
		ctx:    ctx,
		client: client,

		coldStartManager: coldStartManager,
	}
}

// OnWrite implements types.WriteListener.
func (s *bankWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	if s.coldStartManager.HasIngestedInitialData() {
		return fmt.Errorf("cold start manager has already ingested initial data")
	}

	// If the cold start manager has ingested initial data and the key is not empty and the key is a supply key.
	if len(key) > 0 && bytes.Equal(banktypes.SupplyKey, key[:1]) {
		denom := key[len(banktypes.SupplyKey):]
		if err := s.publishSupply(denom, value); err != nil {
			return err
		}
	}

	// If the cold start manager has ingested initial data and the key is not empty and the key is a supply key.
	if len(key) > 0 && bytes.Equal(banktypes.SupplyOffsetKey, key[:1]) {
		denom := key[len(banktypes.SupplyOffsetKey):]
		if err := s.publishSupplyOffset(denom, value); err != nil {
			return err
		}
	}
	return nil
}

// publishSupply publishes the updated supply to the pubsub client.
func (s *bankWriteListener) publishSupply(denom, updatedSupplyBytes []byte) error {
	// Track updated supplies.
	var updatedSupply osmomath.Int
	err := updatedSupply.Unmarshal(updatedSupplyBytes)
	if err != nil {
		return fmt.Errorf("unable to unmarshal supply value %v", err)
	}

	tokenSupply := indexerdomain.TokenSupply{
		// Denom is the key without the supply prefix.
		Denom:  string(denom[len(banktypes.SupplyKey):]),
		Supply: updatedSupply,
	}

	err = s.client.PublishTokenSupply(s.ctx, tokenSupply)
	if err != nil {
		return fmt.Errorf("unable to publish token supply %v", err)
	}

	return nil
}

// publishSupplyOffset publishes the updated supply offset to the pubsub client.
func (s *bankWriteListener) publishSupplyOffset(denom, updatedSupplyOffsetBytes []byte) error {
	// Track updated supplies.
	var updatedSupplyOffset osmomath.Int
	err := updatedSupplyOffset.Unmarshal(updatedSupplyOffsetBytes)
	if err != nil {
		return fmt.Errorf("unable to unmarshal supply value %v", err)
	}

	tokenSupplyOffset := indexerdomain.TokenSupplyOffset{
		// Denom is the key without the supply offset prefix.
		Denom:        string(denom[len(banktypes.SupplyOffsetKey):]),
		SupplyOffset: updatedSupplyOffset,
	}

	err = s.client.PublishTokenSupplyOffset(s.ctx, tokenSupplyOffset)
	if err != nil {
		return fmt.Errorf("unable to publish token supply %v", err)
	}

	return nil
}
