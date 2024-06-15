package writelistener

import (
	"bytes"
	"context"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

var _ storetypes.WriteListener = (*bankWriteListener)(nil)

type bankWriteListener struct {

	// shared context to handle graceful shutdown in case of node stop.
	ctx context.Context

	client indexerdomain.PubSubClient

	coldStartManager domain.ColdStartManager
}

func NewBank(ctx context.Context, client indexerdomain.PubSubClient, coldStartManager domain.ColdStartManager) storetypes.WriteListener {
	return &bankWriteListener{
		ctx:    ctx,
		client: client,

		coldStartManager: coldStartManager,
	}
}

// OnWrite implements types.WriteListener.
func (s *bankWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	// If the cold start manager has ingested initial data and the key is not empty and the key is a supply key.
	if s.coldStartManager.HasIngestedInitialData() && len(key) > 0 && bytes.Equal(banktypes.SupplyKey, key[:1]) {
		// Track updated supplies.
		var updatedSupply osmomath.Int
		err := updatedSupply.Unmarshal(value)
		if err != nil {
			return fmt.Errorf("unable to unmarshal supply value %v", err)
		}

		tokenSupply := indexerdomain.TokenSupply{
			// Denom is the key without the supply prefix.
			Denom:  string(key[len(banktypes.SupplyKey):]),
			Supply: updatedSupply,
		}

		err = s.client.PublishTokenSupply(s.ctx, tokenSupply)
		if err != nil {
			return fmt.Errorf("unable to publish token supply %v", err)
		}
	}

	return nil
}
