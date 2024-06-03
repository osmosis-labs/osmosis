package domain

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Ingester interface {
	ProcessBlock(ctx sdk.Context) error
}

type IndexerPubSubClient interface {
	Publish(ctx context.Context, height uint64, block Block) error
}

type Block struct {
	ChainId     string    `json:"chain_id"`
	Height      uint64    `json:"height"`
	BlockTime   time.Time `json:"timestamp"`
	GasConsumed uint64    `json:"gas_consumed"`
}
