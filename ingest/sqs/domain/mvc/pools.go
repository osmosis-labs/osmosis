package mvc

import (
	"context"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

// PoolsRepository represent the pool's repository contract
type PoolsRepository interface {
	// GetAllPools atomically reads and returns all on-chain pools sorted by ID.
	GetAllPools(context.Context) ([]domain.PoolI, error)
	// GetAllConcentrated atomically reads and returns concentrated pools sorted by ID.
	GetAllConcentrated(context.Context) ([]domain.PoolI, error)
	// GetAllCFMM atomically reads and  returns CFMM pools sorted by ID.
	GetAllCFMM(context.Context) ([]domain.PoolI, error)
	// GetAllCosmWasm atomically reads and returns CosmWasm pools sorted by ID.
	GetAllCosmWasm(context.Context) ([]domain.PoolI, error)
	// StorePools atomically stores the given pools.
	StorePools(ctx context.Context, tx Tx, cfmmPools []domain.PoolI, concentratedPools []domain.PoolI, cosmwasmPools []domain.PoolI) error
	// ClearAllPools atomically clears all pools.
	ClearAllPools(ctx context.Context, tx Tx) error
}

// PoolsUsecase represent the pool's usecases
type PoolsUsecase interface {
	GetAllPools(ctx context.Context) ([]domain.PoolI, error)
}
