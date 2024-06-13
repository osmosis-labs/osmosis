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

	client indexerdomain.PubSubClient
}

func NewBank(ctx context.Context, client indexerdomain.PubSubClient) storetypes.WriteListener {
	return &bankWriteListener{
		ctx:    ctx,
		client: client,
	}
}

// OnWrite implements types.WriteListener.
func (s *bankWriteListener) OnWrite(storeKey storetypes.StoreKey, key []byte, value []byte, delete bool) error {
	// Track updated supplies.
	if len(key) > 0 && bytes.Equal(banktypes.SupplyKey, key[:1]) {
		// TODO: deal with supply updates.

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
