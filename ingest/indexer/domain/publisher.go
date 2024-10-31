package domain

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v27/ingest/common/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
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
	PublishPair(ctx context.Context, pair Pair) error
}

// PairPublisher is an interface for publishing pair data.
type PairPublisher interface {
	// PublishPoolPairs publishes the given pools as pairs.
	// The difference between this function and PublishPair is:
	// - PublishPair operates on the pair level and publishes a single pair.
	// - PublishPoolPairs operates on the pool level and publishes all the pair combo in the pool
	//   with the taker fee and spread factor, as well as the newly created pool metadata, if any.
	PublishPoolPairs(ctx sdk.Context, pools []poolmanagertypes.PoolI, createdPoolIDs map[uint64]commondomain.PoolCreation) error
}
