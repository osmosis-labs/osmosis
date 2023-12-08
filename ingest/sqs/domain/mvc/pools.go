package mvc

import (
	"context"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
)

// PoolsRepository represent the pool's repository contract
type PoolsRepository interface {
	// GetAllPools atomically reads and returns all on-chain pools sorted by ID.
	// Note that this does NOT return tick models for the concentrated pools
	GetAllPools(context.Context) ([]domain.PoolI, error)

	// GetPools atomically reads and returns the pools with the given IDs.
	// Note that this does NOT return tick models for the concentrated pools
	GetPools(ctx context.Context, poolIDs map[uint64]struct{}) (map[uint64]domain.PoolI, error)

	GetTickModelForPools(ctx context.Context, pools []uint64) (map[uint64]domain.TickModel, error)

	// StorePools atomically stores the given pools.
	StorePools(ctx context.Context, tx Tx, pools []domain.PoolI) error
	// ClearAllPools atomically clears all pools.
	ClearAllPools(ctx context.Context, tx Tx) error
}

// PoolsUsecase represent the pool's usecases
type PoolsUsecase interface {
	GetAllPools(ctx context.Context) ([]domain.PoolI, error)

	// GetRoutesFromCandidates converts candidate routes to routes intrusmented with all the data necessary for estimating
	// a swap. This data entails the pool data, the taker fee.
	GetRoutesFromCandidates(ctx context.Context, candidateRoutes route.CandidateRoutes, takerFeeMap domain.TakerFeeMap, tokenInDenom, tokenOutDenom string) ([]route.RouteImpl, error)

	GetTickModelMap(ctx context.Context, poolIDs []uint64) (map[uint64]domain.TickModel, error)
	// GetPool returns the pool with the given ID.
	GetPool(ctx context.Context, poolID uint64) (domain.PoolI, error)
}
