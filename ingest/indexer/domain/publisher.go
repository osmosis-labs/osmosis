package domain

import (
	"context"
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
}
