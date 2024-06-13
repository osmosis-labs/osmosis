package domain

import (
	"context"
)

type PubSubClient interface {
	PublishBlock(ctx context.Context, block Block) error
	PublishTransaction(ctx context.Context, txn Transaction) error
	PublishAsset(ctx context.Context, asset Asset) error
	PublishPool(ctx context.Context, pool Pool) error
}
