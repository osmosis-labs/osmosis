package mvc

import (
	"context"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

// PoolsRepository represent the pool's repository contract
type PoolsRepository interface {
	// GetAllPools atomically reads and returns all on-chain pools sorted by ID.
	GetAllPools(context.Context) ([]domain.PoolI, error)
	// StorePools atomically stores the given pools.
	StorePools(ctx context.Context, tx Tx, pools []domain.PoolI) error
	// ClearAllPools atomically clears all pools.
	ClearAllPools(ctx context.Context, tx Tx) error
}

// PoolsUsecase represent the pool's usecases
type PoolsUsecase interface {
	GetAllPools(ctx context.Context) ([]domain.PoolI, error)
}
