package mocks

import (
	"context"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mvc"
)

type PoolsUsecaseMock struct {
	Pools        []domain.PoolI
	TickModelMap map[uint64]domain.TickModel
}

// GetAllPools implements domain.PoolsUsecase.
func (pm *PoolsUsecaseMock) GetAllPools(ctx context.Context) ([]domain.PoolI, error) {
	return pm.Pools, nil
}

// GetTickModelMap implements mvc.PoolsUsecase.
func (pm *PoolsUsecaseMock) GetTickModelMap(ctx context.Context, poolIDs []uint64) (map[uint64]domain.TickModel, error) {
	return pm.TickModelMap, nil
}

var _ mvc.PoolsUsecase = &PoolsUsecaseMock{}
