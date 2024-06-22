package writelistener

import (
	"bytes"
	"context"
	"fmt"

<<<<<<< HEAD
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

=======
	storetypes "cosmossdk.io/store/types"
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
<<<<<<< HEAD
)

var _ storetypes.WriteListener = (*bankWriteListener)(nil)
=======
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
)

var _ domain.WriteListener = (*bankWriteListener)(nil)
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))

type bankWriteListener struct {

	// shared context to handle graceful shutdown in case of node stop.
	ctx context.Context

	client indexerdomain.TokenSupplyPublisher

	coldStartManager indexerdomain.ColdStartManager
}

<<<<<<< HEAD
func NewBank(ctx context.Context, client indexerdomain.TokenSupplyPublisher, coldStartManager indexerdomain.ColdStartManager) storetypes.WriteListener {
=======
func NewBank(ctx context.Context, client indexerdomain.TokenSupplyPublisher, coldStartManager indexerdomain.ColdStartManager) domain.WriteListener {
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
	return &bankWriteListener{
		ctx:    ctx,
		client: client,

		coldStartManager: coldStartManager,
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
	if !s.coldStartManager.HasIngestedInitialData() {
		return indexerdomain.ErrColdStartManagerDidNotIngest
	}

	// If the cold start manager has ingested initial data and the key is not empty and the key is a supply key.
	if len(key) > 0 && bytes.Equal(banktypes.SupplyKey, key[:1]) {
		denom := string(key[len(banktypes.SupplyKey):])
		if err := s.publishSupply(denom, value); err != nil {
			return err
		}
	}

	// If the cold start manager has ingested initial data and the key is not empty and the key is a supply key.
	if len(key) > 0 && bytes.Equal(banktypes.SupplyOffsetKey, key[:1]) {
		denom := string(key[len(banktypes.SupplyOffsetKey):])
		if err := s.publishSupplyOffset(denom, value); err != nil {
			return err
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
