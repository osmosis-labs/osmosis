package mocks

import (
	"context"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mvc"
)

type RedisPoolsRepositoryMock struct {
	pools []domain.PoolI
}

// GetTickModelForPools implements mvc.PoolsRepository.
func (*RedisPoolsRepositoryMock) GetTickModelForPools(ctx context.Context, pools []uint64) (map[uint64]domain.TickModel, error) {
	panic("unimplemented")
}

// ClearAllPools implements domain.PoolsRepository.
func (*RedisPoolsRepositoryMock) ClearAllPools(ctx context.Context, tx mvc.Tx) error {
	panic("unimplemented")
}

var _ mvc.PoolsRepository = &RedisPoolsRepositoryMock{}

// GetAllPools implements domain.PoolsRepository.
func (r *RedisPoolsRepositoryMock) GetAllPools(context.Context) ([]domain.PoolI, error) {
	allPools := make([]domain.PoolI, len(r.pools))
	copy(allPools, r.pools)
	return allPools, nil
}

// StorePools implements domain.PoolsRepository.
func (r *RedisPoolsRepositoryMock) StorePools(ctx context.Context, tx mvc.Tx, allPools []domain.PoolI) error {
	r.pools = allPools
	return nil
}
