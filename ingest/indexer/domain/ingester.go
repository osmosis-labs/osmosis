package domain

import (
	"context"
)

// TokenSupplyPublisher is an interface for publishing token supply data.
type TokenSupplyPublisher interface {
	PublishTokenSupply(ctx context.Context, tokenSupply TokenSupply) error
	PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset TokenSupplyOffset) error
}

// Ingester is an interface for ingesting & publishing various types of data.
type Ingester interface {
	TokenSupplyPublisher

	PublishBlock(ctx context.Context, block Block) error
	PublishTransaction(ctx context.Context, txn Transaction) error
	PublishPool(ctx context.Context, pool Pool) error
}
