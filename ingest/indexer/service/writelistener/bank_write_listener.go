package writelistener

import (
	"bytes"
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
)

var _ domain.WriteListener = (*bankWriteListener)(nil)

type bankWriteListener struct {

	// shared context to handle graceful shutdown in case of node stop.
	ctx context.Context

	client indexerdomain.PubSubClient

	coldStartManager indexerdomain.ColdStartManager
}

func NewBank(ctx context.Context, client indexerdomain.PubSubClient, coldStartManager indexerdomain.ColdStartManager) domain.WriteListener {
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
		if err := s.publishSupply(denom, value); err != nil {
			return err
		}
	}
	return nil
}

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
