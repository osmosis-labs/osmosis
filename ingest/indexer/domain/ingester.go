package domain

import (
	"context"
)

// Ingester is an interface for ingesting & publishing various types of data.
type Ingester interface {
	PublishBlock(ctx context.Context, block Block) error
	PublishTransaction(ctx context.Context, txn Transaction) error
	PublishPool(ctx context.Context, pool Pool) error
	PublishTokenSupply(ctx context.Context, tokenSupply TokenSupply) error
	PublishTokenSupplyOffset(ctx context.Context, tokenSupply TokenSupplyOffset) error
}
