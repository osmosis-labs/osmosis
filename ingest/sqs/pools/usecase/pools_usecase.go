package usecase

import (
	"context"
	"time"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type poolsUseCase struct {
	contextTimeout         time.Duration
	poolsRepository        domain.PoolsRepository
	redisRepositoryManager domain.TxManager
}

// NewPoolsUsecase will create a new pools use case object
func NewPoolsUsecase(timeout time.Duration, poolsRepository domain.PoolsRepository, redisRepositoryManager domain.TxManager) domain.PoolsUsecase {
	return &poolsUseCase{
		contextTimeout:         timeout,
		poolsRepository:        poolsRepository,
		redisRepositoryManager: redisRepositoryManager,
	}
}

// GetAllPools returns all pools from the repository.
func (a *poolsUseCase) GetAllPools(ctx context.Context) ([]domain.PoolI, error) {
	ctx, cancel := context.WithTimeout(ctx, a.contextTimeout)
	defer cancel()

	pools, err := a.poolsRepository.GetAllPools(ctx)
	if err != nil {
		return nil, err
	}

	return pools, nil
}
