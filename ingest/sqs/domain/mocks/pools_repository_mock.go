package mocks

import (
	"context"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
)

type RedisPoolsRepositoryMock struct {
	Pools     []domain.PoolI
	TickModel map[uint64]domain.TickModel
}

// GetPools implements mvc.PoolsRepository.
func (r *RedisPoolsRepositoryMock) GetPools(ctx context.Context, poolIDs map[uint64]struct{}) (map[uint64]domain.PoolI, error) {
	result := map[uint64]domain.PoolI{}
	for _, pool := range r.Pools {
		result[pool.GetId()] = pool
	}
	return result, nil
}

// GetTickModelForPools implements mvc.PoolsRepository.
func (r *RedisPoolsRepositoryMock) GetTickModelForPools(ctx context.Context, pools []uint64) (map[uint64]domain.TickModel, error) {
	return r.TickModel, nil
}

// ClearAllPools implements domain.PoolsRepository.
func (*RedisPoolsRepositoryMock) ClearAllPools(ctx context.Context, tx mvc.Tx) error {
	panic("unimplemented")
}

var _ mvc.PoolsRepository = &RedisPoolsRepositoryMock{}

// GetAllPools implements domain.PoolsRepository.
func (r *RedisPoolsRepositoryMock) GetAllPools(context.Context) ([]domain.PoolI, error) {
	allPools := make([]domain.PoolI, len(r.Pools))
	copy(allPools, r.Pools)
	return allPools, nil
}

// StorePools implements domain.PoolsRepository.
func (r *RedisPoolsRepositoryMock) StorePools(ctx context.Context, tx mvc.Tx, allPools []domain.PoolI) error {
	r.Pools = allPools
	return nil
}
