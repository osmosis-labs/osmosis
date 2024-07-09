package domain

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

// TokenSupplyPublisher is an interface for publishing token supply data.
type TokenSupplyPublisher interface {
	PublishTokenSupply(ctx context.Context, tokenSupply TokenSupply) error
	PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset TokenSupplyOffset) error
}

// Publisher is an interface for publishing various types of data.
type Publisher interface {
	TokenSupplyPublisher

	PublishBlock(ctx context.Context, block Block) error
	PublishTransaction(ctx context.Context, txn Transaction) error
	PublishPool(ctx context.Context, pool Pool) error
	PublishPair(ctx context.Context, pair Pair) error
}

// PairPublisher is an interface for publishing pair data.
type PairPublisher interface {
	// PublishPoolPairs publishes the given pools as pairs.
	PublishPoolPairs(ctx sdk.Context, pools []poolmanagertypes.PoolI) error
}
