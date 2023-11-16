package mocks

import (
	"context"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type PoolsUsecaseMock struct {
	Pools []domain.PoolI
}

// GetAllPools implements domain.PoolsUsecase.
func (r *PoolsUsecaseMock) GetAllPools(ctx context.Context) ([]domain.PoolI, error) {
	return r.Pools, nil
}

var _ domain.PoolsUsecase = &PoolsUsecaseMock{}
