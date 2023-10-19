package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type poolsUseCase struct {
	contextTimeout  time.Duration
	poolsRepository domain.PoolsRepository
}

// NewPoolsUsecase will create a new pools use case object
func NewPoolsUsecase(timeout time.Duration, poolsRepository domain.PoolsRepository) domain.PoolsUsecase {
	return &poolsUseCase{
		contextTimeout:  timeout,
		poolsRepository: poolsRepository,
	}
}

// GetAllPools returns all pools from the repository.
func (a *poolsUseCase) GetAllPools(ctx context.Context) ([]domain.PoolI, error) {
	ctx, cancel := context.WithTimeout(ctx, a.contextTimeout)
	defer cancel()

	cfmmPools, err := a.poolsRepository.GetAllCFMM(ctx)
	if err != nil {
		return nil, err
	}

	concentratedPools, err := a.poolsRepository.GetAllConcentrated(ctx)
	if err != nil {
		return nil, err
	}

	cosmWasmPools, err := a.poolsRepository.GetAllCosmWasm(ctx)
	if err != nil {
		return nil, err
	}

	allPools := make([]domain.PoolI, 0, len(cfmmPools)+len(concentratedPools)+len(cosmWasmPools))
	for _, pool := range cfmmPools {
		allPools = append(allPools, pool)
	}

	for _, pool := range concentratedPools {
		allPools = append(allPools, pool)
	}

	for _, pool := range cosmWasmPools {
		allPools = append(allPools, pool)
	}

	// Sort by ID
	sort.Slice(allPools, func(i, j int) bool {
		return allPools[i].GetId() < allPools[j].GetId()
	})

	return allPools, nil
}
